package authcode

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/int128/kubelogin/mocks/github.com/int128/kubelogin/pkg/infrastructure/browser_mock"
	"github.com/int128/kubelogin/mocks/github.com/int128/kubelogin/pkg/oidc/client_mock"
	"github.com/int128/kubelogin/pkg/oidc"
	"github.com/int128/kubelogin/pkg/oidc/client"
	"github.com/int128/kubelogin/pkg/pkce"
	"github.com/int128/kubelogin/pkg/testing/logger"
	"github.com/stretchr/testify/mock"
)

func TestBrowser_Do(t *testing.T) {
	timeout := 5 * time.Second

	t.Run("Success", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), timeout)
		defer cancel()
		o := &BrowserOption{
			BindAddress:                []string{"127.0.0.1:8000"},
			SkipOpenBrowser:            true,
			AuthenticationTimeout:      10 * time.Second,
			LocalServerCertFile:        "/path/to/local-server-cert",
			LocalServerKeyFile:         "/path/to/local-server-key",
			OpenURLAfterAuthentication: "https://example.com/success.html",
			AuthRequestExtraParams:     map[string]string{"ttl": "86400", "reauth": "true"},
		}
		mockClient := client_mock.NewMockInterface(t)
		mockClient.EXPECT().NegotiatedPKCEMethod().Return(pkce.NoMethod)
		mockClient.EXPECT().
			GetTokenByAuthCode(mock.Anything, mock.Anything, mock.Anything).
			Run(func(_ context.Context, in client.GetTokenByAuthCodeInput, readyChan chan<- string) {
				if diff := cmp.Diff(o.BindAddress, in.BindAddress); diff != "" {
					t.Errorf("BindAddress mismatch (-want +got):\n%s", diff)
				}
				if diff := cmp.Diff(BrowserRedirectHTML("https://example.com/success.html"), in.LocalServerSuccessHTML); diff != "" {
					t.Errorf("LocalServerSuccessHTML mismatch (-want +got):\n%s", diff)
				}
				if diff := cmp.Diff(o.AuthRequestExtraParams, in.AuthRequestExtraParams); diff != "" {
					t.Errorf("AuthRequestExtraParams mismatch (-want +got):\n%s", diff)
				}
				if diff := cmp.Diff(o.LocalServerKeyFile, in.LocalServerKeyFile); diff != "" {
					t.Errorf("LocalServerKeyFile mismatch (-want +got):\n%s", diff)
				}
				if diff := cmp.Diff(o.LocalServerCertFile, in.LocalServerCertFile); diff != "" {
					t.Errorf("LocalServerCertFile mismatch (-want +got):\n%s", diff)
				}
				readyChan <- "LOCAL_SERVER_URL"
			}).
			Return(&oidc.TokenSet{
				IDToken:      "YOUR_ID_TOKEN",
				RefreshToken: "YOUR_REFRESH_TOKEN",
			}, nil)
		u := Browser{
			Logger: logger.New(t),
		}
		got, err := u.Do(ctx, o, mockClient)
		if err != nil {
			t.Errorf("Do returned error: %+v", err)
		}
		want := &oidc.TokenSet{
			IDToken:      "YOUR_ID_TOKEN",
			RefreshToken: "YOUR_REFRESH_TOKEN",
		}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("OpenBrowser", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), timeout)
		defer cancel()
		o := &BrowserOption{
			BindAddress:           []string{"127.0.0.1:8000"},
			AuthenticationTimeout: 10 * time.Second,
		}
		mockClient := client_mock.NewMockInterface(t)
		mockClient.EXPECT().NegotiatedPKCEMethod().Return(pkce.NoMethod)
		mockClient.EXPECT().
			GetTokenByAuthCode(mock.Anything, mock.Anything, mock.Anything).
			Run(func(_ context.Context, _ client.GetTokenByAuthCodeInput, readyChan chan<- string) {
				readyChan <- "LOCAL_SERVER_URL"
			}).
			Return(&oidc.TokenSet{
				IDToken:      "YOUR_ID_TOKEN",
				RefreshToken: "YOUR_REFRESH_TOKEN",
			}, nil)
		mockBrowser := browser_mock.NewMockInterface(t)
		mockBrowser.EXPECT().
			Open("LOCAL_SERVER_URL").
			Return(nil)
		u := Browser{
			Logger:  logger.New(t),
			Browser: mockBrowser,
		}
		got, err := u.Do(ctx, o, mockClient)
		if err != nil {
			t.Errorf("Do returned error: %+v", err)
		}
		want := &oidc.TokenSet{
			IDToken:      "YOUR_ID_TOKEN",
			RefreshToken: "YOUR_REFRESH_TOKEN",
		}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("OpenBrowserCommand", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), timeout)
		defer cancel()
		o := &BrowserOption{
			BindAddress:           []string{"127.0.0.1:8000"},
			BrowserCommand:        "firefox",
			AuthenticationTimeout: 10 * time.Second,
		}
		mockClient := client_mock.NewMockInterface(t)
		mockClient.EXPECT().NegotiatedPKCEMethod().Return(pkce.NoMethod)
		mockClient.EXPECT().
			GetTokenByAuthCode(mock.Anything, mock.Anything, mock.Anything).
			Run(func(_ context.Context, _ client.GetTokenByAuthCodeInput, readyChan chan<- string) {
				readyChan <- "LOCAL_SERVER_URL"
			}).
			Return(&oidc.TokenSet{
				IDToken:      "YOUR_ID_TOKEN",
				RefreshToken: "YOUR_REFRESH_TOKEN",
			}, nil)
		mockBrowser := browser_mock.NewMockInterface(t)
		mockBrowser.EXPECT().
			OpenCommand(mock.Anything, "LOCAL_SERVER_URL", "firefox").
			Return(nil)
		u := Browser{
			Logger:  logger.New(t),
			Browser: mockBrowser,
		}
		got, err := u.Do(ctx, o, mockClient)
		if err != nil {
			t.Errorf("Do returned error: %+v", err)
		}
		want := &oidc.TokenSet{
			IDToken:      "YOUR_ID_TOKEN",
			RefreshToken: "YOUR_REFRESH_TOKEN",
		}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("mismatch (-want +got):\n%s", diff)
		}
	})
}
