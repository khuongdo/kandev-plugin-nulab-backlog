package connection

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/testutil"
)

const testBaseURL = "https://kandev.example.test"

func oauthConfigMap(t *testing.T) (map[string]any, string) {
	t.Helper()
	secret := testutil.Token(t)
	return map[string]any{
		"oauth_client_id":     "client-id-1",
		"oauth_client_secret": secret,
		"public_base_url":     testBaseURL,
	}, secret
}

func TestOAuthConfigParsesTheThreeFields(t *testing.T) {
	m, secret := oauthConfigMap(t)
	cfg, err := ParseOAuthConfig(m)
	require.NoError(t, err)
	require.Equal(t, "client-id-1", cfg.ClientID)
	require.Equal(t, secret, cfg.ClientSecret)
	require.Equal(t, testBaseURL+"/api/plugins/nulab-backlog/webhooks/oauth-callback", cfg.RedirectURI())
	require.Equal(t, "client-id-1", cfg.Client().ClientID)
	require.Equal(t, secret, cfg.Client().ClientSecret)
}

func TestOAuthConfigMissingOrBlankFieldIsNotConfigured(t *testing.T) {
	for _, field := range []string{"oauth_client_id", "oauth_client_secret", "public_base_url"} {
		for name, value := range map[string]any{"missing": nil, "blank": "  ", "not a string": 42} {
			t.Run(field+" "+name, func(t *testing.T) {
				m, secret := oauthConfigMap(t)
				if value == nil {
					delete(m, field)
				} else {
					m[field] = value
				}
				_, err := ParseOAuthConfig(m)
				require.ErrorIs(t, err, ErrOAuthNotConfigured)
				testutil.AssertNoLeak(t, err.Error(), secret, 8)
			})
		}
	}
	_, err := ParseOAuthConfig(nil)
	require.ErrorIs(t, err, ErrOAuthNotConfigured)
}

func TestOAuthConfigPublicBaseURLRules(t *testing.T) {
	cases := []struct {
		raw      string
		redirect string // empty means rejected
	}{
		{"https://kandev.example.test", testBaseURL + "/api/plugins/nulab-backlog/webhooks/oauth-callback"},
		{"https://kandev.example.test/", testBaseURL + "/api/plugins/nulab-backlog/webhooks/oauth-callback"},
		{"https://kandev.example.test:8443/kandev", "https://kandev.example.test:8443/kandev/api/plugins/nulab-backlog/webhooks/oauth-callback"},
		{"http://localhost:38429", "http://localhost:38429/api/plugins/nulab-backlog/webhooks/oauth-callback"},
		{"http://127.0.0.1:38429/", "http://127.0.0.1:38429/api/plugins/nulab-backlog/webhooks/oauth-callback"},
		{"http://kandev.example.test", ""},
		{"ftp://kandev.example.test", ""},
		{"kandev.example.test", ""},
		{"/relative", ""},
		{"https://kandev.example.test/?x=1", ""},
		{"https://kandev.example.test/#frag", ""},
		{"https://user:pw@kandev.example.test", ""},
		{"https://", ""},
	}
	for _, tc := range cases {
		t.Run(tc.raw, func(t *testing.T) {
			m, _ := oauthConfigMap(t)
			m["public_base_url"] = tc.raw
			cfg, err := ParseOAuthConfig(m)
			if tc.redirect == "" {
				require.ErrorIs(t, err, ErrOAuthNotConfigured)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.redirect, cfg.RedirectURI())
		})
	}
}

func TestOAuthConfigHidesTheClientSecretWhenFormatted(t *testing.T) {
	m, secret := oauthConfigMap(t)
	cfg, err := ParseOAuthConfig(m)
	require.NoError(t, err)
	testutil.AssertNoLeak(t, fmt.Sprintf("%v %+v %#v %s", cfg, cfg, cfg, cfg), secret, 8)
	require.True(t, errors.Is(fmt.Errorf("wrap: %w", ErrOAuthNotConfigured), ErrOAuthNotConfigured))
}
