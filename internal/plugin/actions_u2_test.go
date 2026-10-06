package plugin

import (
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/backlog"
	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/testutil"
)

// Action keys added by U2, spelled as in manifest.yaml.
const (
	keyStartOAuth   = "connection.start_oauth"
	keyTest         = "connection.test"
	keyDisconnect   = "connection.disconnect"
	keyListProjects = "connection.list_projects"
	keySetProjects  = "connection.set_projects"
)

func oauthConfig(t *testing.T) (map[string]any, string) {
	t.Helper()
	secret := testutil.Token(t)
	return map[string]any{
		"oauth_client_id": "client-id-1", "oauth_client_secret": secret,
		"public_base_url": "https://kandev.example.test",
	}, secret
}

func (r *rig) connected(t *testing.T) string {
	t.Helper()
	key := testutil.APIKey(t)
	resp, _ := r.call(t, actionConnectAPIKey, connectBody(key))
	require.Equal(t, 200, resp.Status)
	return key
}

func TestRecheckActionSucceedsAndRejectedKeyIsReconnectRequired(t *testing.T) {
	r := newRig(t)
	r.connected(t)
	resp, out := r.call(t, keyTest, nil)
	require.Equal(t, 200, resp.Status)
	require.Equal(t, "Test User", out["connectedUserName"])

	r.gw.err = &backlog.Error{Kind: backlog.KindUnauthorized, Status: 401}
	resp, out = r.call(t, keyTest, nil)
	require.Equal(t, 401, resp.Status)
	require.Equal(t, "reconnect_required", errorOf(t, out)["code"])
}

func TestDisconnectActionRemovesTheSecretAndPendingSignIn(t *testing.T) {
	r := newRig(t)
	r.host.config, _ = oauthConfig(t)
	r.connected(t)
	resp, _ := r.call(t, keyStartOAuth, map[string]string{"spaceUrl": "example-space.backlog.com"})
	require.Equal(t, 200, resp.Status)
	require.Contains(t, r.host.state, "workspace/ws-1/oauth_pending")

	resp, out := r.call(t, keyDisconnect, nil)
	require.Equal(t, 200, resp.Status)
	require.Equal(t, "not_connected", out["state"])
	require.Equal(t, true, out["enabled"])
	require.Empty(t, r.host.secrets)
	require.NotContains(t, r.host.state, "workspace/ws-1/oauth_pending")
}

func TestProjectsActions(t *testing.T) {
	r := newRig(t)
	r.connected(t)
	r.gw.projects = []backlog.Project{{ID: 101, Key: "PROJ", Name: "Test Project"}}

	resp, out := r.call(t, keyListProjects, nil)
	require.Equal(t, 200, resp.Status)
	require.Equal(t, []any{map[string]any{"projectKey": "PROJ", "projectId": float64(101), "projectName": "Test Project", "selected": false}}, out["projects"])

	resp, out = r.call(t, keySetProjects, map[string]any{"projectKeys": []string{"PROJ"}, "workspaceId": "ws-other"})
	require.Equal(t, 200, resp.Status)
	require.Equal(t, []any{"PROJ"}, out["selectedProjects"])
	require.NotContains(t, r.host.state, "workspace/ws-other/connection", "the workspace comes from the verified context")

	for name, body := range map[string]any{
		"not a list":   map[string]any{"projectKeys": "PROJ"},
		"missing":      map[string]any{},
		"malformed":    []byte("{not json"),
		"unknown key":  map[string]any{"projectKeys": []string{"NOPE"}},
		"a number key": map[string]any{"projectKeys": []any{1}},
	} {
		t.Run(name, func(t *testing.T) {
			resp, out := r.call(t, keySetProjects, body)
			require.Equal(t, 400, resp.Status)
			require.Equal(t, "validation", errorOf(t, out)["code"])
			require.Equal(t, "projectKeys", errorOf(t, out)["field"])
		})
	}

	r.gw.projectsErr = &backlog.Error{Kind: backlog.KindRateLimited, Status: 429, RetryAfter: 7 * time.Second}
	resp, out = r.call(t, keyListProjects, nil)
	require.Equal(t, 429, resp.Status)
	require.Equal(t, "7", resp.Headers["Retry-After"])
	require.EqualValues(t, 7, errorOf(t, out)["retryAfterSeconds"])
}

func TestOAuthStartActionReadsTheConfigOnEveryCall(t *testing.T) {
	r := newRig(t)
	cfg, secret := oauthConfig(t)
	r.host.config = cfg
	for range 2 {
		resp, out := r.call(t, keyStartOAuth, map[string]string{"spaceUrl": "example-space.backlog.com"})
		require.Equal(t, 200, resp.Status)
		au, err := url.Parse(out["authorizeUrl"].(string))
		require.NoError(t, err)
		require.Equal(t, "example-space.backlog.com", au.Host)
		require.Equal(t, "/OAuth2AccessRequest.action", au.Path)
		testutil.AssertNoLeak(t, string(resp.Body), secret, 8)
	}
	require.Equal(t, 2, r.host.configCalls)
	testutil.AssertNoLeak(t, r.logs.String(), secret, 8)
}

func TestOAuthStartActionWithoutConfig(t *testing.T) {
	r := newRig(t)
	r.host.config = map[string]any{}
	resp, out := r.call(t, keyStartOAuth, map[string]string{"spaceUrl": "example-space.backlog.com"})
	require.Equal(t, 400, resp.Status)
	require.Equal(t, "validation", errorOf(t, out)["code"])
	require.Equal(t, "oauth", errorOf(t, out)["field"])

	r.host.failConfig = true
	resp, out = r.call(t, keyStartOAuth, map[string]string{"spaceUrl": "example-space.backlog.com"})
	require.Equal(t, 500, resp.Status)
	require.Equal(t, "internal", errorOf(t, out)["code"])

	resp, _ = r.call(t, keyStartOAuth, []byte("{not json"))
	require.Equal(t, 400, resp.Status)
}

func TestU2_GuardRefusesEveryNewActionWhileOff(t *testing.T) {
	r := newRig(t)
	r.connected(t)
	r.call(t, actionSetEnabled, map[string]any{"enabled": false})
	for _, key := range []string{keyStartOAuth, keyTest, keyDisconnect, keyListProjects, keySetProjects} {
		resp, out := r.call(t, key, map[string]any{})
		require.Equal(t, 409, resp.Status, key)
		require.Equal(t, "integration_disabled", errorOf(t, out)["code"], key)
	}
	require.NotEmpty(t, r.host.secrets, "disconnect did not run")
}

func TestActionFailureLogOnABacklog500(t *testing.T) {
	r := newRig(t)
	key := r.connected(t)
	r.gw.err = &backlog.Error{Kind: backlog.KindUnreachable, Status: 500}
	before := len(strings.Split(strings.TrimSpace(r.logs.String()), "\n"))
	resp, out := r.call(t, keyTest, nil)
	require.Equal(t, 503, resp.Status)

	var failed []map[string]any
	for _, line := range strings.Split(strings.TrimSpace(r.logs.String()), "\n")[before:] {
		m := parseLine(t, line)
		if m["event"] == "action_failed" {
			failed = append(failed, m)
		}
	}
	require.Len(t, failed, 1, "AC8.3.1: exactly one record")
	e := failed[0]
	require.Equal(t, "WARN", e["level"])
	require.Equal(t, keyTest, e["action"])
	require.Equal(t, "unreachable", e["errorCode"])
	require.Contains(t, e, "durationMs")
	require.Equal(t, "ws-1", e["workspaceId"])
	require.Equal(t, errorOf(t, out)["requestId"], e["requestId"])
	testutil.AssertNoLeak(t, r.logs.String(), key, 8)
}

func TestU2_ValidationFailuresAreNotLoggedAsActionFailed(t *testing.T) {
	r := newRig(t)
	r.connected(t)
	r.call(t, keySetProjects, map[string]any{"projectKeys": "x"})
	require.NotContains(t, r.logs.String(), `"event":"action_failed"`)
}

func TestU2_ReconnectRequiredMapsTo401(t *testing.T) {
	require.Equal(t, 401, statusFor("reconnect_required"))
}
