package plugin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"strings"
	"sync"
	"testing"
	"testing/synctest"
	"time"

	"github.com/kandev/kandev/pkg/pluginsdk"
	"github.com/stretchr/testify/require"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/backlog"
	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/connection"
	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/testutil"
)

// fakeHost implements the five store methods the plugin uses. The embedded
// nil Host makes any other call panic, which proves U1 uses nothing else.
type fakeHost struct {
	pluginsdk.Host
	mu       sync.Mutex
	secrets  map[string]string
	state    map[string]map[string]any
	failRead bool
	// failSecretRead makes GetSecret fail, to prove an action never needs it.
	failSecretRead bool
	reads          []string      // secret and state keys read, in order
	writes         int           // secret and state writes
	getDelay       time.Duration // each GetState takes this long (virtual time under synctest)
	config         map[string]any
	failConfig     bool
	configCalls    int
}

func (h *fakeHost) GetConfig(context.Context) (map[string]any, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.configCalls++
	if h.failConfig {
		return nil, errors.New("config unavailable")
	}
	return h.config, nil
}

func newFakeHost() *fakeHost {
	return &fakeHost{secrets: map[string]string{}, state: map[string]map[string]any{}}
}

func (h *fakeHost) GetState(ctx context.Context, scope, scopeID, key string) (map[string]any, bool, error) {
	if h.getDelay > 0 {
		select {
		case <-time.After(h.getDelay):
		case <-ctx.Done():
			return nil, false, ctx.Err()
		}
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	h.reads = append(h.reads, "state:"+key)
	if h.failRead {
		return nil, false, errors.New("state unavailable")
	}
	v, ok := h.state[scope+"/"+scopeID+"/"+key]
	return v, ok, nil
}

func (h *fakeHost) SetState(_ context.Context, scope, scopeID, key string, value map[string]any) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.writes++
	h.state[scope+"/"+scopeID+"/"+key] = value
	return nil
}

func (h *fakeHost) DeleteState(_ context.Context, scope, scopeID, key string) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.writes++
	delete(h.state, scope+"/"+scopeID+"/"+key)
	return nil
}

func (h *fakeHost) GetSecret(_ context.Context, key string) (string, bool, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.reads = append(h.reads, "secret:"+key)
	if h.failSecretRead {
		return "", false, errors.New("secret store unavailable")
	}
	v, ok := h.secrets[key]
	return v, ok, nil
}

func (h *fakeHost) SetSecret(_ context.Context, key, value string) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.writes++
	h.secrets[key] = value
	return nil
}

func (h *fakeHost) DeleteSecret(_ context.Context, key string) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.secrets, key)
	return nil
}

type fakeGateway struct {
	err   error
	panic bool
	block bool // never answer: wait for the call's deadline
	calls int

	projects      []backlog.Project
	projectsErr   error
	tokens        backlog.TokenSet
	exchangeErr   error
	exchangePanic bool
	exchanges     int
}

func (g *fakeGateway) Projects(context.Context, backlog.Credentials) ([]backlog.Project, error) {
	return g.projects, g.projectsErr
}

func (g *fakeGateway) ExchangeOAuthCode(context.Context, string, backlog.OAuthClient, string, string) (backlog.TokenSet, error) {
	g.exchanges++
	if g.exchangePanic {
		panic("exchange exploded")
	}
	return g.tokens, g.exchangeErr
}

func (g *fakeGateway) RefreshToken(context.Context, string, backlog.OAuthClient, string) (backlog.TokenSet, error) {
	return g.tokens, nil
}

func (g *fakeGateway) Myself(ctx context.Context, _ backlog.Credentials) (backlog.User, error) {
	g.calls++
	if g.block {
		<-ctx.Done()
		return backlog.User{}, &backlog.Error{Kind: backlog.KindUnreachable, Class: "timeout"}
	}
	if g.panic {
		panic("gateway exploded")
	}
	return backlog.User{ID: 1234, UserID: "test.user", Name: "Test User"}, g.err
}

type syncBuffer struct {
	mu sync.Mutex
	sb strings.Builder
}

func (b *syncBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.sb.Write(p)
}

func (b *syncBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.sb.String()
}

type rig struct {
	rt   *Runtime
	host *fakeHost
	gw   *fakeGateway
	logs *syncBuffer

	verifier string // the browser cookie of the last startState
}

func newRig(t *testing.T) *rig {
	t.Helper()
	gw, logs, host := &fakeGateway{}, &syncBuffer{}, newFakeHost()
	rt := newRuntime(gw, logs, "debug")
	rt.SetHost(host)
	return &rig{rt: rt, host: host, gw: gw, logs: logs}
}

func (r *rig) call(t *testing.T, key string, body any) (*pluginsdk.PluginActionResponse, map[string]any) {
	t.Helper()
	raw, ok := body.([]byte)
	if !ok && body != nil {
		var err error
		raw, err = json.Marshal(body)
		require.NoError(t, err)
	}
	resp, err := r.rt.HandleAction(context.Background(), &pluginsdk.PluginActionRequest{
		ActionKey: key,
		Context:   pluginsdk.VerifiedActionContext{WorkspaceID: "ws-1", ActorID: "user-1"},
		Body:      raw,
	})
	require.NoError(t, err, "HandleAction never returns a Go error")
	require.NotNil(t, resp)
	require.Equal(t, "application/json", resp.Headers["Content-Type"])
	var out map[string]any
	require.NoError(t, json.Unmarshal(resp.Body, &out))
	return resp, out
}

func connectBody(key string) map[string]string {
	return map[string]string{"spaceUrl": "example-space.backlog.com", "apiKey": key}
}

func TestConnectionGetWhenNotConnected(t *testing.T) {
	r := newRig(t)
	resp, out := r.call(t, actionGet, nil)
	require.Equal(t, 200, resp.Status)
	require.Equal(t, "not_connected", out["state"])
	require.Equal(t, false, out["connected"])
	require.Equal(t, false, out["hasApiKey"])
}

func TestConnectApiKeySuccessThenGetIsConnected(t *testing.T) {
	r := newRig(t)
	key := testutil.APIKey(t)
	resp, out := r.call(t, actionConnectAPIKey, connectBody(key))
	require.Equal(t, 200, resp.Status)
	require.Equal(t, "connected", out["state"])
	require.Equal(t, "Test User", out["connectedUserName"])
	require.Equal(t, "example-space.backlog.com", out["spaceHost"])
	require.Equal(t, "api_key", out["authMethod"])
	require.Equal(t, true, out["hasApiKey"])
	require.EqualValues(t, 1, out["connectionEpoch"])

	_, got := r.call(t, actionGet, nil)
	require.Equal(t, "connected", got["state"])
	require.Contains(t, r.host.secrets, "backlog.connection.ws-1")
}

func TestConnectionGetErrorStateWhenSecretIsMissing(t *testing.T) {
	r := newRig(t)
	r.call(t, actionConnectAPIKey, connectBody(testutil.APIKey(t)))
	delete(r.host.secrets, "backlog.connection.ws-1")
	_, out := r.call(t, actionGet, nil)
	require.Equal(t, "error", out["state"])
	require.Equal(t, false, out["connected"])
}

func TestErrorCodesMapToStatusAndActionError(t *testing.T) {
	cases := []struct {
		name       string
		body       any
		gwErr      error
		failRead   bool
		status     int
		code       string
		field      string
		retryAfter string
	}{
		{name: "invalid address", body: map[string]string{"spaceUrl": "evil.example.com", "apiKey": "k-123456"}, status: 400, code: "validation", field: "spaceUrl"},
		{name: "rejected key", body: "key", gwErr: &backlog.Error{Kind: backlog.KindUnauthorized, Status: 401}, status: 400, code: "validation", field: "apiKey"},
		{name: "space not found", body: "key", gwErr: &backlog.Error{Kind: backlog.KindNotFound, Status: 404}, status: 400, code: "validation", field: "spaceUrl"},
		{name: "rate limited", body: "key", gwErr: &backlog.Error{Kind: backlog.KindRateLimited, Status: 429, RetryAfter: 42 * time.Second}, status: 429, code: "rate_limited", retryAfter: "42"},
		{name: "unreachable", body: "key", gwErr: &backlog.Error{Kind: backlog.KindUnreachable, Status: 503}, status: 503, code: "unreachable"},
		{name: "store failure", body: "key", failRead: true, status: 500, code: "internal"},
		{name: "malformed body", body: []byte("{not json"), status: 400, code: "validation"},
		{name: "empty body", body: []byte(""), status: 400, code: "validation"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := newRig(t)
			r.gw.err = tc.gwErr
			r.host.failRead = tc.failRead
			body := tc.body
			if body == "key" {
				body = connectBody(testutil.APIKey(t))
			}
			resp, out := r.call(t, actionConnectAPIKey, body)
			require.Equal(t, tc.status, resp.Status)
			e, ok := out["error"].(map[string]any)
			require.True(t, ok, "ActionError shape")
			require.Equal(t, tc.code, e["code"])
			if tc.field != "" {
				require.Equal(t, tc.field, e["field"])
			} else {
				require.NotContains(t, e, "field")
			}
			if tc.retryAfter != "" {
				require.Equal(t, tc.retryAfter, resp.Headers["Retry-After"])
				require.EqualValues(t, 42, e["retryAfterSeconds"])
			} else {
				require.NotContains(t, e, "retryAfterSeconds")
			}
			require.NotEmpty(t, e["requestId"])
			require.Contains(t, r.logs.String(), fmt.Sprintf(`"requestId":"%s"`, e["requestId"]))
		})
	}
}

func TestConflictMapsTo409(t *testing.T) {
	require.Equal(t, 409, statusFor(connection.CodeConflict))
	require.Equal(t, 500, statusFor("something-new"))
}

func TestGetStoreFailureIsInternal(t *testing.T) {
	r := newRig(t)
	r.host.failRead = true
	resp, out := r.call(t, actionGet, nil)
	require.Equal(t, 500, resp.Status)
	require.Equal(t, "internal", out["error"].(map[string]any)["code"])
}

func TestUnknownActionIs404JSON(t *testing.T) {
	r := newRig(t)
	resp, out := r.call(t, "connection.nope", nil)
	require.Equal(t, 404, resp.Status)
	require.Equal(t, "not_found", out["error"].(map[string]any)["code"])
}

func TestPanicInHandlerIsRecoveredAsInternal(t *testing.T) {
	r := newRig(t)
	r.gw.panic = true
	resp, out := r.call(t, actionConnectAPIKey, connectBody(testutil.APIKey(t)))
	require.Equal(t, 500, resp.Status)
	require.Equal(t, "internal", out["error"].(map[string]any)["code"])
	require.Contains(t, r.logs.String(), `"event":"action_panic"`)
}

func TestMissingHostIsInternal(t *testing.T) {
	rt := newRuntime(&fakeGateway{}, &syncBuffer{}, "")
	resp, err := rt.HandleAction(context.Background(), &pluginsdk.PluginActionRequest{
		ActionKey: actionGet, Context: pluginsdk.VerifiedActionContext{WorkspaceID: "ws-1"},
	})
	require.NoError(t, err)
	require.Equal(t, 500, resp.Status)
}

func TestLogLinesCarryTheRequiredFields(t *testing.T) {
	r := newRig(t)
	r.call(t, actionConnectAPIKey, connectBody(testutil.APIKey(t)))
	lines := strings.Split(strings.TrimSpace(r.logs.String()), "\n")
	require.NotEmpty(t, lines)
	for _, line := range lines {
		var m map[string]any
		require.NoError(t, json.Unmarshal([]byte(line), &m), line)
		for _, f := range []string{"time", "level", "event"} {
			require.Contains(t, m, f, line)
		}
		if m["event"] != "plugin_started" {
			require.Equal(t, "ws-1", m["workspaceId"], line)
			require.NotEmpty(t, m["requestId"], line)
		}
	}
	require.Contains(t, r.logs.String(), `"event":"plugin_started"`)
}

func TestResponsesAndLogsNeverContainTheKey(t *testing.T) {
	r := newRig(t)
	key := testutil.APIKey(t)
	var bodies []string
	for _, gwErr := range []error{nil, &backlog.Error{Kind: backlog.KindUnauthorized, Status: 401}, &backlog.Error{Kind: backlog.KindUnreachable}} {
		r.gw.err = gwErr
		resp, _ := r.call(t, actionConnectAPIKey, connectBody(key))
		bodies = append(bodies, string(resp.Body))
		resp, _ = r.call(t, actionGet, nil)
		bodies = append(bodies, string(resp.Body))
	}
	testutil.AssertNoLeak(t, strings.Join(bodies, "\n"), key)
	testutil.AssertNoLeak(t, r.logs.String(), key, "Test User")
}

func TestNewRuntimeIsAKandevPlugin(t *testing.T) {
	var p pluginsdk.Plugin = NewRuntime()
	_, isAction := p.(pluginsdk.ActionHandler)
	require.True(t, isAction)
}

func errorOf(t *testing.T, out map[string]any) map[string]any {
	t.Helper()
	e, ok := out["error"].(map[string]any)
	require.True(t, ok, "ActionError shape")
	return e
}

func TestSetEnabledDoesOneStateReadOneStateWriteAndNoSecretAccess(t *testing.T) {
	r := newRig(t)
	r.call(t, actionConnectAPIKey, connectBody(testutil.APIKey(t)))
	secretsBefore := maps.Clone(r.host.secrets)
	for _, enabled := range []bool{false, true} {
		r.host.reads, r.host.writes = nil, 0
		resp, out := r.call(t, actionSetEnabled, map[string]any{"enabled": enabled})
		require.Equal(t, 200, resp.Status)
		require.Equal(t, []string{"state:integration"}, r.host.reads, "NFR1.3: one state read (the switch), no secret read")
		require.Equal(t, 1, r.host.writes, "NFR1.3: one write")
		require.Equal(t, secretsBefore, r.host.secrets, "the write is the switch, never the secret (BR7.4)")
		require.Equal(t, map[string]any{"enabled": enabled}, out, "the reply carries only what the switch write knows")
	}
	require.Contains(t, r.logs.String(), `"event":"integration_switch_changed"`)
}

func TestSetEnabledSucceedsWhenTheWriteSucceeds(t *testing.T) {
	r := newRig(t)
	r.call(t, actionConnectAPIKey, connectBody(testutil.APIKey(t)))
	r.host.failSecretRead = true // anything after the write must not turn a saved switch into an error
	resp, out := r.call(t, actionSetEnabled, map[string]any{"enabled": false})
	require.Equal(t, 200, resp.Status)
	require.Equal(t, false, out["enabled"])
	_, got := r.call(t, actionGet, nil) // get still reports the failure honestly
	require.Equal(t, "internal", errorOf(t, got)["code"])
}

func TestSetEnabledRejectsANonBooleanEnabled(t *testing.T) {
	for name, body := range map[string]any{
		"missing":   map[string]any{},
		"a string":  map[string]any{"enabled": "true"},
		"a number":  map[string]any{"enabled": 1},
		"null":      map[string]any{"enabled": nil},
		"malformed": []byte("{not json"),
		"empty":     []byte(""),
	} {
		t.Run(name, func(t *testing.T) {
			r := newRig(t)
			resp, out := r.call(t, actionSetEnabled, body)
			require.Equal(t, 400, resp.Status)
			e := errorOf(t, out)
			require.Equal(t, "validation", e["code"])
			require.Equal(t, "enabled", e["field"])
			require.Zero(t, r.host.writes, "nothing written")
		})
	}
}

func TestSetEnabledUsesOnlyTheVerifiedWorkspace(t *testing.T) {
	r := newRig(t)
	resp, _ := r.call(t, actionSetEnabled, map[string]any{"enabled": false, "workspaceId": "ws-other"})
	require.Equal(t, 200, resp.Status)
	require.Contains(t, r.host.state, "workspace/ws-1/integration")
	require.NotContains(t, r.host.state, "workspace/ws-other/integration")
}

func TestWhileOffOnlyGetAndSetEnabledRun(t *testing.T) {
	r := newRig(t)
	r.call(t, actionSetEnabled, map[string]any{"enabled": false})
	r.host.reads, r.host.writes, r.gw.calls = nil, 0, 0

	resp, out := r.call(t, actionConnectAPIKey, connectBody(testutil.APIKey(t)))
	require.Equal(t, 409, resp.Status)
	require.Equal(t, "integration_disabled", errorOf(t, out)["code"])
	require.Equal(t, []string{"state:integration"}, r.host.reads, "the guard reads only the switch")
	require.Zero(t, r.host.writes)
	require.Zero(t, r.gw.calls)

	resp, out = r.call(t, actionGet, nil)
	require.Equal(t, 200, resp.Status, "connection.get always works")
	require.Equal(t, false, out["enabled"])

	resp, out = r.call(t, actionSetEnabled, map[string]any{"enabled": true})
	require.Equal(t, 200, resp.Status, "connection.set_enabled always works")
	require.Equal(t, true, out["enabled"])

	resp, _ = r.call(t, actionConnectAPIKey, connectBody(testutil.APIKey(t)))
	require.Equal(t, 200, resp.Status)
}

func TestGuardFailsClosed(t *testing.T) {
	for name, setup := range map[string]func(*fakeHost){
		"the switch read fails": func(h *fakeHost) { h.failRead = true },
		"the switch value is undecodable": func(h *fakeHost) {
			h.state["workspace/ws-1/integration"] = map[string]any{"schemaVersion": 1, "enabled": "yes"}
		},
	} {
		t.Run(name, func(t *testing.T) {
			r := newRig(t)
			setup(r.host)
			resp, out := r.call(t, actionConnectAPIKey, connectBody(testutil.APIKey(t)))
			require.Equal(t, 500, resp.Status)
			require.Equal(t, "internal", errorOf(t, out)["code"])
			require.Zero(t, r.gw.calls, "the action never runs")
			require.Zero(t, r.host.writes)
		})
	}
}

func TestGuardedActionsAreEveryActionExceptGetAndSetEnabled(t *testing.T) {
	require.False(t, guarded(actionGet))
	require.False(t, guarded(actionSetEnabled))
	require.True(t, guarded(actionConnectAPIKey))
	require.True(t, guarded("issues.list"), "actions added by later units are guarded by default")
}

func TestIntegrationDisabledMapsTo409(t *testing.T) {
	require.Equal(t, 409, statusFor(connection.CodeIntegrationDisabled))
}

func TestGuardRunsInsideTheConnectDeadline(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		r := newRig(t)
		r.host.getDelay = 900 * time.Millisecond // the guard read and Connect's own switch read
		r.gw.block = true
		start := time.Now()
		resp, out := r.call(t, actionConnectAPIKey, connectBody(testutil.APIKey(t)))
		require.Equal(t, 503, resp.Status)
		require.Equal(t, "unreachable", errorOf(t, out)["code"])
		// The 12 s deadline starts before the guard, so the Backlog call still
		// ends at deadline minus 2 s and the action stays within 14 s (NFR1.4).
		require.Equal(t, 10*time.Second, time.Since(start))
	})
}
