package backlog

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/redact"
	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/testutil"
)

const redirectURI = "https://kandev.example.test/api/plugins/nulab-backlog/webhooks/oauth-callback"

// tokenEndpoint serves a token response built at run time and records the form.
func tokenEndpoint(t *testing.T, status int, access, refresh string) (http.HandlerFunc, *url.Values, *http.Request) {
	t.Helper()
	form, req := &url.Values{}, &http.Request{}
	return func(w http.ResponseWriter, r *http.Request) {
		*req = *r.Clone(context.Background())
		require.NoError(t, r.ParseForm())
		*form = r.PostForm
		w.WriteHeader(status)
		if status == 200 {
			_, _ = fmt.Fprintf(w, `{"access_token":%q,"token_type":"Bearer","expires_in":3600,"refresh_token":%q}`, access, refresh)
			return
		}
		_, _ = w.Write(readFixture(t, "token_invalid_grant.json"))
	}, form, req
}

func oauthLogs() (context.Context, *bytes.Buffer) {
	var buf bytes.Buffer
	log := slog.New(redact.NewHandler(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})))
	return redact.WithLogger(context.Background(), log), &buf
}

func TestOAuthExchangePostsTheCodeForm(t *testing.T) {
	access, refresh, secret, code := testutil.Token(t), testutil.Token(t), testutil.Token(t), testutil.Token(t)
	h, form, req := tokenEndpoint(t, 200, access, refresh)
	c, _ := fakeBacklog(t, h)
	c.Now = func() time.Time { return tokenNow }

	got, err := c.ExchangeOAuthCode(context.Background(), spaceHost, OAuthClient{ClientID: "client-id-1", ClientSecret: secret}, code, redirectURI)
	require.NoError(t, err)
	require.Equal(t, TokenSet{AccessToken: access, RefreshToken: refresh, ExpiresAt: tokenNow.Add(time.Hour)}, got)
	require.Equal(t, http.MethodPost, req.Method)
	require.Equal(t, "/api/v2/oauth2/token", req.URL.Path)
	require.NotNil(t, req.TLS)
	require.Equal(t, "application/x-www-form-urlencoded", req.Header.Get("Content-Type"))
	require.Empty(t, req.Header.Get("Authorization"))
	require.Empty(t, req.URL.RawQuery)
	require.Equal(t, url.Values{
		"grant_type": {"authorization_code"}, "code": {code}, "redirect_uri": {redirectURI},
		"client_id": {"client-id-1"}, "client_secret": {secret},
	}, *form)
}

func TestOAuthExchangeErrorsAndLogsHideEverySecret(t *testing.T) {
	cases := []struct {
		status int
		kind   Kind
	}{{400, KindInvalid}, {401, KindUnauthorized}, {302, KindUnreachable}, {500, KindUnreachable}}
	for _, tc := range cases {
		t.Run(http.StatusText(tc.status), func(t *testing.T) {
			secret, code := testutil.Token(t), testutil.Token(t)
			c, _ := fakeBacklog(t, func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Location", "https://evil.example.test/")
				w.WriteHeader(tc.status)
				_, _ = w.Write(readFixture(t, "token_invalid_grant.json"))
			})
			ctx, logs := oauthLogs()
			_, err := c.ExchangeOAuthCode(ctx, spaceHost, OAuthClient{ClientID: "client-id-1", ClientSecret: secret}, code, redirectURI)
			var be *Error
			require.True(t, errors.As(err, &be))
			require.Equal(t, tc.kind, be.Kind)
			text := err.Error() + logs.String()
			testutil.AssertNoLeak(t, text, secret, "SECRET-BODY-MARKER", "invalid_grant")
			testutil.AssertNoLeak(t, text, code)
			require.Contains(t, logs.String(), `"path":"/api/v2/oauth2/token"`)
		})
	}
}

func TestOAuthExchangeBodyLimitAndHTTPSOnly(t *testing.T) {
	c, _ := fakeBacklog(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"access_token":"` + strings.Repeat("x", 2<<20) + `"}`))
	})
	_, err := c.ExchangeOAuthCode(context.Background(), spaceHost, OAuthClient{ClientID: "c", ClientSecret: "s-1234"}, "code-1234", redirectURI)
	var be *Error
	require.True(t, errors.As(err, &be))
	require.Equal(t, "body", be.Class)

	var hits atomic.Int32
	c, _ = fakeBacklog(t, func(http.ResponseWriter, *http.Request) { hits.Add(1) })
	_, err = c.ExchangeOAuthCode(context.Background(), "evil.io/x", OAuthClient{}, "code-1234", redirectURI)
	require.Error(t, err)
	require.Zero(t, hits.Load())
}

func TestTokenRefreshSendsTheRefreshGrantAndReturnsTheRotatedToken(t *testing.T) {
	access, rotated, secret, old := testutil.Token(t), testutil.Token(t), testutil.Token(t), testutil.Token(t)
	h, form, _ := tokenEndpoint(t, 200, access, rotated)
	c, _ := fakeBacklog(t, h)
	got, err := c.RefreshToken(context.Background(), spaceHost, OAuthClient{ClientID: "client-id-1", ClientSecret: secret}, old)
	require.NoError(t, err)
	require.Equal(t, access, got.AccessToken)
	require.Equal(t, rotated, got.RefreshToken)
	require.Equal(t, url.Values{
		"grant_type": {"refresh_token"}, "refresh_token": {old},
		"client_id": {"client-id-1"}, "client_secret": {secret},
	}, *form)
}

func TestTokenRefreshRefusals(t *testing.T) {
	for status, kind := range map[int]Kind{400: KindInvalid, 401: KindUnauthorized} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			old, secret := testutil.Token(t), testutil.Token(t)
			h, _, _ := tokenEndpoint(t, status, "", "")
			c, _ := fakeBacklog(t, h)
			ctx, logs := oauthLogs()
			_, err := c.RefreshToken(ctx, spaceHost, OAuthClient{ClientID: "client-id-1", ClientSecret: secret}, old)
			var be *Error
			require.True(t, errors.As(err, &be))
			require.Equal(t, kind, be.Kind)
			testutil.AssertNoLeak(t, err.Error()+logs.String(), old, "invalid_grant")
			testutil.AssertNoLeak(t, logs.String(), secret)
		})
	}
}
