package plugin

import (
	"context"
	"encoding/json"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/kandev/kandev/pkg/pluginsdk"
	"github.com/stretchr/testify/require"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/backlog"
	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/testutil"
)

const settingsPath = "/settings/workspaces/ws-1/integrations/nulab-backlog"

func parseLine(t *testing.T, line string) map[string]any {
	t.Helper()
	var m map[string]any
	require.NoError(t, json.Unmarshal([]byte(line), &m), line)
	return m
}

// startState runs connection.start_oauth and returns the state it minted.
// The verifier cookie is kept in r.verifier, as the browser would keep it.
func (r *rig) startState(t *testing.T) string {
	t.Helper()
	body, verifier := startBody(t)
	resp, out := r.call(t, keyStartOAuth, body)
	require.Equal(t, 200, resp.Status)
	r.verifier = verifier
	au, err := url.Parse(out["authorizeUrl"].(string))
	require.NoError(t, err)
	return au.Query().Get("state")
}

// callback relays the callback with the verifier cookie among other cookies,
// as Kandev forwards it to a public webhook.
func (r *rig) callback(t *testing.T, method, key string, q url.Values) *pluginsdk.WebhookResponse {
	t.Helper()
	return r.callbackWith(t, method, key, q, map[string]string{"Cookie": "theme=dark; nulab_backlog_oauth_verifier=" + r.verifier + "; lang=en"})
}

func (r *rig) callbackWith(t *testing.T, method, key string, q url.Values, headers map[string]string) *pluginsdk.WebhookResponse {
	t.Helper()
	resp, err := r.rt.HandleWebhook(context.Background(), &pluginsdk.WebhookRequest{WebhookKey: key, Method: method, Query: q.Encode(), Headers: headers})
	require.NoError(t, err, "HandleWebhook never returns a Go error")
	require.NotNil(t, resp)
	return resp
}

func oauthRig(t *testing.T) *rig {
	t.Helper()
	r := newRig(t)
	r.host.config, _ = oauthConfig(t)
	r.gw.tokens = backlog.TokenSet{AccessToken: testutil.Token(t), RefreshToken: testutil.Token(t), ExpiresAt: time.Now().Add(time.Hour)}
	return r
}

func TestWebhookRedirectsWithTheOutcome(t *testing.T) {
	cases := []struct {
		name  string
		query func(state, code string) url.Values
		want  string
	}{
		{"connected", func(s, c string) url.Values { return url.Values{"state": {s}, "code": {c}} }, settingsPath + "?oauth=connected"},
		{"cancelled", func(s, _ string) url.Values { return url.Values{"state": {s}, "error": {"access_denied"}} }, settingsPath + "?oauth=cancelled"},
		{"failed", func(s, _ string) url.Values { return url.Values{"state": {s}} }, settingsPath + "?oauth=failed"},
		{"undecodable state", func(_, c string) url.Values { return url.Values{"state": {"%%"}, "code": {c}} }, "/settings/integrations?oauth=failed"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := oauthRig(t)
			state, code := r.startState(t), testutil.Token(t)
			resp := r.callback(t, "GET", "oauth-callback", tc.query(state, code))
			require.EqualValues(t, 302, resp.Status)
			loc := resp.Headers["Location"]
			require.Equal(t, tc.want, loc)
			testutil.AssertNoLeak(t, loc, code, state, r.gw.tokens.AccessToken, "SECRET-BODY-MARKER", r.verifier)
			require.Contains(t, r.logs.String(), `"event":"oauth_callback"`)
			testutil.AssertNoLeak(t, r.logs.String(), code, state, r.verifier)
		})
	}
}

func TestWebhookAddsRestoredOnARestore(t *testing.T) {
	r := oauthRig(t)
	r.call(t, actionConnectAPIKey, connectBody(testutil.APIKey(t)))
	r.call(t, keyDisconnect, nil)
	resp := r.callback(t, "GET", "oauth-callback", url.Values{"state": {r.startState(t)}, "code": {"c-1234"}})
	require.Equal(t, settingsPath+"?oauth=connected&restored=1", resp.Headers["Location"])
}

func TestWebhookRefusals(t *testing.T) {
	r := oauthRig(t)
	require.EqualValues(t, 404, r.callback(t, "GET", "nope", nil).Status)
	require.EqualValues(t, 405, r.callback(t, "POST", "oauth-callback", nil).Status)
	resp, err := r.rt.HandleWebhook(context.Background(), &pluginsdk.WebhookRequest{WebhookKey: "oauth-callback", Method: "GET", Query: "%zz"})
	require.NoError(t, err)
	require.Equal(t, "/settings/integrations?oauth=failed", resp.Headers["Location"])
}

func TestWebhookPanicIsRecoveredAsFailed(t *testing.T) {
	r := oauthRig(t)
	r.gw.exchangePanic = true
	resp := r.callback(t, "GET", "oauth-callback", url.Values{"state": {r.startState(t)}, "code": {"c-1234"}})
	require.EqualValues(t, 302, resp.Status)
	require.Equal(t, "/settings/integrations?oauth=failed", resp.Headers["Location"])
	require.True(t, strings.Contains(r.logs.String(), `"event":"webhook_panic"`))
}

func TestU2_WebhookNeedsTheVerifierCookie(t *testing.T) {
	cases := map[string]func(r *rig) map[string]string{
		"no Cookie header":   func(*rig) map[string]string { return nil },
		"no verifier cookie": func(*rig) map[string]string { return map[string]string{"Cookie": "theme=dark"} },
		"wrong verifier": func(*rig) map[string]string {
			return map[string]string{"Cookie": "nulab_backlog_oauth_verifier=" + testutil.Token(t)}
		},
		"empty verifier cookie": func(*rig) map[string]string {
			return map[string]string{"Cookie": "theme=dark; nulab_backlog_oauth_verifier=; lang=en"}
		},
		"malformed Cookie":     func(*rig) map[string]string { return map[string]string{"Cookie": ";;=;"} },
		"verifier in a header": func(r *rig) map[string]string { return map[string]string{"X-Verifier": r.verifier} },
	}
	for name, headers := range cases {
		t.Run(name, func(t *testing.T) {
			r := oauthRig(t)
			state := r.startState(t)
			q := url.Values{"state": {state}, "code": {testutil.Token(t)}}
			resp := r.callbackWith(t, "GET", "oauth-callback", q, headers(r))
			require.Equal(t, settingsPath+"?oauth=failed", resp.Headers["Location"])
			require.Zero(t, r.gw.exchanges, "no token request without the starting browser")
			require.Contains(t, r.host.state, "workspace/ws-1/oauth_pending", "the pending sign-in is kept")
			testutil.AssertNoLeak(t, r.logs.String(), r.verifier)

			resp = r.callback(t, "GET", "oauth-callback", q)
			require.Equal(t, settingsPath+"?oauth=connected", resp.Headers["Location"])
			require.Equal(t, 1, r.gw.exchanges)
			testutil.AssertNoLeak(t, r.logs.String()+resp.Headers["Location"], r.verifier)
		})
	}
}
