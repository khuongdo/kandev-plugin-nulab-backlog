package backlog

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/testutil"
)

var tokenNow = time.Date(2026, 10, 6, 12, 0, 0, 0, time.UTC)

func TestTokenParseDecodesATokenResponse(t *testing.T) {
	got, err := parseTokenSet(readFixture(t, "token_ok.json"), tokenNow)
	require.NoError(t, err)
	require.Equal(t, "TESTSECRET-access-token-ok", got.AccessToken)
	require.Equal(t, "TESTSECRET-refresh-token-ok", got.RefreshToken)
	require.Equal(t, tokenNow.Add(3600*time.Second), got.ExpiresAt, "ExpiresAt is Now plus expires_in")
}

func TestTokenParseAcceptsLowerCaseBearer(t *testing.T) {
	_, err := parseTokenSet([]byte(`{"access_token":"a-1","token_type":"bearer","expires_in":60,"refresh_token":"r-1"}`), tokenNow)
	require.NoError(t, err)
}

func TestTokenParseRejectsUnusableBodies(t *testing.T) {
	cases := map[string]string{ //nolint:gosec // G101: fake token bodies built for the parser test
		"missing access_token":  `{"token_type":"Bearer","expires_in":3600,"refresh_token":"r-1"}`,
		"missing refresh_token": `{"access_token":"a-1","token_type":"Bearer","expires_in":3600}`,
		"zero expires_in":       `{"access_token":"a-1","token_type":"Bearer","expires_in":0,"refresh_token":"r-1"}`,
		"negative expires_in":   `{"access_token":"a-1","token_type":"Bearer","expires_in":-5,"refresh_token":"r-1"}`,
		"missing expires_in":    `{"access_token":"a-1","token_type":"Bearer","refresh_token":"r-1"}`,
		"not a Bearer token":    `{"access_token":"a-1","token_type":"mac","expires_in":3600,"refresh_token":"r-1"}`,
		"invalid JSON":          `<html>`,
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := parseTokenSet([]byte(body), tokenNow)
			var be *Error
			require.True(t, errors.As(err, &be))
			require.Equal(t, KindUnreachable, be.Kind)
			require.Equal(t, "body", be.Class)
			require.NotContains(t, err.Error(), "a-1")
		})
	}
}

func TestTokenSetAndOAuthClientHideSecretsWhenFormatted(t *testing.T) {
	access, refresh, secret := testutil.Token(t), testutil.Token(t), testutil.Token(t)
	ts := TokenSet{AccessToken: access, RefreshToken: refresh, ExpiresAt: tokenNow}
	oc := OAuthClient{ClientID: "client-id-1", ClientSecret: secret}
	out := fmt.Sprintf("%v %+v %#v %s %v %+v %#v %s", ts, ts, ts, ts, oc, oc, oc, oc)
	testutil.AssertNoLeak(t, out, access, 8)
	testutil.AssertNoLeak(t, out, refresh, 8)
	testutil.AssertNoLeak(t, out, secret, 8)
	require.Contains(t, out, "client-id-1", "the client id is not a secret")
}
