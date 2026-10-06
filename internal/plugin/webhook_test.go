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
func (r *rig) startState(t *testing.T) string {
	t.Helper()
	resp, out := r.call(t, keyStartOAuth, map[string]string{"spaceUrl": "example-space.backlog.com"})
	require.Equal(t, 200, resp.Status)
	au, err := url.Parse(out["authorizeUrl"].(string))
	require.NoError(t, err)
	return au.Query().Get("state")
}

func (r *rig) callback(t *testing.T, method, key string, q url.Values) *pluginsdk.WebhookResponse {
	t.Helper()
	resp, err := r.rt.HandleWebhook(context.Background(), &pluginsdk.WebhookRequest{WebhookKey: key, Method: method, Query: q.Encode()})
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
			testutil.AssertNoLeak(t, loc, code, 8, state, r.gw.tokens.AccessToken, "SECRET-BODY-MARKER")
			require.Contains(t, r.logs.String(), `"event":"oauth_callback"`)
			testutil.AssertNoLeak(t, r.logs.String(), code, 8, state)
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
