package connection

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/url"
	"strings"
	"sync"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/backlog"
	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/redact"
	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/testutil"
)

// fakeGateway is a Gateway whose result and blocking behaviour are scripted.
type fakeGateway struct {
	mu       sync.Mutex
	calls    int
	user     backlog.User
	err      error
	block    bool          // block until ctx is done or release is closed
	entered  chan struct{} // receives once per call when set
	release  chan struct{}
	deadline time.Duration // remaining time seen on the call's context
	creds    backlog.Credentials
	delay    time.Duration // answer after this long (virtual time under synctest)
	onCall   func()        // runs inside the call, e.g. to turn the switch off
	noRetry  bool          // the call's context asked for no 429 retry

	// U2 calls.
	projects      []backlog.Project
	projectsErr   error
	projectsCalls int
	tokens        backlog.TokenSet // returned by ExchangeOAuthCode
	exchangeErr   error
	exchangeCalls int
	exchangeCode  string
	exchangeURI   string
	exchangeID    string
	onExchange    func()
	refreshed     backlog.TokenSet // returned by RefreshToken
	refreshErr    error
	refreshCalls  int
	refreshDelay  time.Duration // virtual time under synctest
	onRefresh     func()
}

func (g *fakeGateway) Myself(ctx context.Context, c backlog.Credentials) (backlog.User, error) {
	g.mu.Lock()
	g.calls++
	g.creds = c
	g.noRetry = backlog.RetryDisabled(ctx)
	if d, ok := ctx.Deadline(); ok {
		g.deadline = time.Until(d)
	}
	block, entered, release, delay, onCall := g.block, g.entered, g.release, g.delay, g.onCall
	g.mu.Unlock()
	if entered != nil {
		entered <- struct{}{}
	}
	if onCall != nil {
		onCall()
	}
	if delay > 0 {
		select {
		case <-ctx.Done():
			return backlog.User{}, &backlog.Error{Kind: backlog.KindUnreachable, Class: "timeout"}
		case <-time.After(delay):
		}
	}
	if block {
		select {
		case <-ctx.Done():
			// Like backlog.Client: cancellation unchanged, a deadline is Unreachable.
			if errors.Is(ctx.Err(), context.Canceled) {
				return backlog.User{}, context.Canceled
			}
			return backlog.User{}, &backlog.Error{Kind: backlog.KindUnreachable, Class: "timeout"}
		case <-release:
		}
	}
	return g.user, g.err
}

func (g *fakeGateway) callCount() int {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.calls
}

func (g *fakeGateway) Projects(_ context.Context, c backlog.Credentials) ([]backlog.Project, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.projectsCalls++
	g.creds = c
	return g.projects, g.projectsErr
}

func (g *fakeGateway) ExchangeOAuthCode(_ context.Context, _ string, client backlog.OAuthClient, code, redirectURI string) (backlog.TokenSet, error) {
	g.mu.Lock()
	g.exchangeCalls++
	g.exchangeCode, g.exchangeURI, g.exchangeID = code, redirectURI, client.ClientID
	hook := g.onExchange
	g.mu.Unlock()
	if hook != nil {
		hook()
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.tokens, g.exchangeErr
}

func (g *fakeGateway) RefreshToken(ctx context.Context, _ string, _ backlog.OAuthClient, _ string) (backlog.TokenSet, error) {
	g.mu.Lock()
	g.refreshCalls++
	delay, hook := g.refreshDelay, g.onRefresh
	g.mu.Unlock()
	if delay > 0 {
		time.Sleep(delay)
	}
	if hook != nil {
		hook()
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.refreshed, g.refreshErr
}

func (g *fakeGateway) counts() (myself, projects, exchange, refresh int) {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.calls, g.projectsCalls, g.exchangeCalls, g.refreshCalls
}

type harness struct {
	svc     *Service
	gw      *fakeGateway
	secrets *fakeSecrets
	state   *fakeState
	logs    *syncBuffer
	ctx     context.Context
}

func newHarness(t *testing.T) *harness {
	t.Helper()
	secrets, state := newFakeSecrets(), newFakeState()
	switchOn(t, state, ws, "ws-a", "ws-b")
	gw := &fakeGateway{user: testUser}
	buf := &syncBuffer{}
	log := slog.New(redact.NewHandler(slog.NewJSONHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug})))
	return &harness{
		svc: NewService(gw, newTestStore(secrets, state)), gw: gw,
		secrets: secrets, state: state, logs: buf,
		ctx: redact.WithLogger(context.Background(), log),
	}
}

func (h *harness) events(t *testing.T) []map[string]any {
	t.Helper()
	var out []map[string]any
	for _, line := range strings.Split(strings.TrimSpace(h.logs.String()), "\n") {
		if line == "" {
			continue
		}
		var m map[string]any
		require.NoError(t, json.Unmarshal([]byte(line), &m))
		out = append(out, m)
	}
	return out
}

func eventsNamed(events []map[string]any, name string) []map[string]any {
	var out []map[string]any
	for _, e := range events {
		if e["event"] == name {
			out = append(out, e)
		}
	}
	return out
}

func be(kind backlog.Kind, status int) error { return &backlog.Error{Kind: kind, Status: status} }

// wf3Rows is every row of the functional-spec WF3 outcome table, plus the
// validation and store-failure exits.
var wf3Rows = []struct {
	name      string
	spaceURL  string
	badKey    bool
	gwErr     error
	failStore bool
	code      string
	field     string
	retry     int
	gwCalls   int
}{
	{name: "200 with a valid user connects", spaceURL: "example-space.backlog.com", gwCalls: 1},
	{name: "200 with an unusable body is unreachable", spaceURL: "example-space.backlog.com", gwErr: &backlog.Error{Kind: backlog.KindUnreachable, Status: 200, Class: "body"}, code: CodeUnreachable, gwCalls: 1},
	{name: "401 is a validation error on apiKey", spaceURL: "example-space.backlog.com", gwErr: be(backlog.KindUnauthorized, 401), code: CodeValidation, field: FieldAPIKey, gwCalls: 1},
	{name: "403 is a validation error on apiKey", spaceURL: "example-space.backlog.com", gwErr: be(backlog.KindForbidden, 403), code: CodeValidation, field: FieldAPIKey, gwCalls: 1},
	{name: "404 is a validation error on spaceUrl", spaceURL: "example-space.backlog.com", gwErr: be(backlog.KindNotFound, 404), code: CodeValidation, field: FieldSpaceURL, gwCalls: 1},
	{name: "429 is rate_limited with the wait", spaceURL: "example-space.backlog.com", gwErr: &backlog.Error{Kind: backlog.KindRateLimited, Status: 429, RetryAfter: 1500 * time.Millisecond}, code: CodeRateLimited, retry: 2, gwCalls: 1},
	{name: "400 is unreachable", spaceURL: "example-space.backlog.com", gwErr: be(backlog.KindInvalid, 400), code: CodeUnreachable, gwCalls: 1},
	{name: "422 is unreachable", spaceURL: "example-space.backlog.com", gwErr: be(backlog.KindInvalid, 422), code: CodeUnreachable, gwCalls: 1},
	{name: "409 is unreachable", spaceURL: "example-space.backlog.com", gwErr: be(backlog.KindConflict, 409), code: CodeUnreachable, gwCalls: 1},
	{name: "another 4xx is unreachable", spaceURL: "example-space.backlog.com", gwErr: be(backlog.KindUnreachable, 418), code: CodeUnreachable, gwCalls: 1},
	{name: "3xx is unreachable", spaceURL: "example-space.backlog.com", gwErr: be(backlog.KindUnreachable, 302), code: CodeUnreachable, gwCalls: 1},
	{name: "5xx is unreachable", spaceURL: "example-space.backlog.com", gwErr: be(backlog.KindUnreachable, 503), code: CodeUnreachable, gwCalls: 1},
	{name: "a timeout is unreachable", spaceURL: "example-space.backlog.com", gwErr: &backlog.Error{Kind: backlog.KindUnreachable, Class: "timeout"}, code: CodeUnreachable, gwCalls: 1},
	{name: "an oversized body is unreachable", spaceURL: "example-space.backlog.com", gwErr: &backlog.Error{Kind: backlog.KindUnreachable, Status: 200, Class: "body"}, code: CodeUnreachable, gwCalls: 1},
	{name: "an invalid address makes no call", spaceURL: "http://example-space.backlog.com", code: CodeValidation, field: FieldSpaceURL},
	{name: "an invalid key makes no call", spaceURL: "example-space.backlog.com", badKey: true, code: CodeValidation, field: FieldAPIKey},
	{name: "a store failure is internal", spaceURL: "example-space.backlog.com", failStore: true, code: CodeInternal, gwCalls: 1},
}

func TestConnectOutcomeTable(t *testing.T) {
	for _, tc := range wf3Rows {
		t.Run(tc.name, func(t *testing.T) {
			h := newHarness(t)
			// Start connected so we can prove a failure keeps the old connection (NFR5.8).
			oldKey := testutil.APIKey(t)
			_, err := h.svc.Connect(h.ctx, ws, ConnectInput{SpaceURL: "old.backlog.com", APIKey: oldKey})
			require.NoError(t, err)
			beforeSecrets, beforeState := h.secrets.snapshot(), h.state.snapshot()
			h.gw.calls = 0
			h.gw.err = tc.gwErr
			h.state.failSet = tc.failStore
			key := testutil.APIKey(t)
			if tc.badKey {
				key = "bad key"
			}

			view, err := h.svc.Connect(h.ctx, ws, ConnectInput{SpaceURL: tc.spaceURL, APIKey: key})
			require.Equal(t, tc.gwCalls, h.gw.callCount())
			if tc.code == "" {
				require.NoError(t, err)
				require.True(t, view.Connected)
				require.Equal(t, "example-space.backlog.com", view.SpaceHost)
				require.Equal(t, "Test User", view.ConnectedUserName)
				require.Equal(t, 2, view.ConnectionEpoch)
				require.Equal(t, key, h.gw.creds.APIKey)
				return
			}
			out := Classify(err)
			require.Equal(t, tc.code, out.Code)
			require.Equal(t, tc.field, out.Field)
			require.Equal(t, tc.retry, out.RetryAfterSeconds)
			require.Equal(t, beforeSecrets, h.secrets.snapshot(), "secret unchanged")
			require.Equal(t, beforeState, h.state.snapshot(), "record unchanged")
		})
	}
}

// Connect must answer a 429 at once (NFR2.1, T-RATE-01): its Myself call asks
// the client for no retry, so one request is made and nothing is stored.
func TestConnectAsksForNoRetryOn429(t *testing.T) {
	h := newHarness(t)
	h.gw.err = &backlog.Error{Kind: backlog.KindRateLimited, Status: 429, RetryAfter: 2 * time.Second}
	_, err := h.svc.Connect(h.ctx, ws, ConnectInput{SpaceURL: "example-space.backlog.com", APIKey: testutil.APIKey(t)})
	require.Equal(t, Outcome{Code: CodeRateLimited, RetryAfterSeconds: 2}, Classify(err))
	require.Equal(t, 1, h.gw.callCount())
	require.True(t, h.gw.noRetry, "the Connect Myself call must not retry")
	require.Empty(t, h.secrets.snapshot())
}

func TestOAuthSignInAsksForNoRetry(t *testing.T) {
	u := newU2(t)
	state := u.start(t)
	u.okTokens(t)
	u.svc.CompleteOAuth(u.ctx, url.Values{"state": {state}, "code": {testutil.Token(t)}}, u.verifier)
	require.True(t, u.gw.noRetry, "the sign-in Myself call must not retry")
}

func TestConnectionTestKeepsTheRetry(t *testing.T) {
	h := newHarness(t)
	_, err := h.svc.Connect(h.ctx, ws, ConnectInput{SpaceURL: "example-space.backlog.com", APIKey: testutil.APIKey(t)})
	require.NoError(t, err)
	_, err = h.svc.Test(h.ctx, ws)
	require.NoError(t, err)
	require.False(t, h.gw.noRetry, "connection.test keeps the shared retry")
}

func TestConnectCancelledWritesNothingAndReturnsCancellation(t *testing.T) {
	h := newHarness(t)
	ctx, cancel := context.WithCancel(h.ctx)
	h.gw.block = true
	h.gw.entered = make(chan struct{}, 1)
	go func() { <-h.gw.entered; cancel() }()
	_, err := h.svc.Connect(ctx, ws, ConnectInput{SpaceURL: "a.backlog.com", APIKey: testutil.APIKey(t)})
	require.ErrorIs(t, err, context.Canceled)
	require.Empty(t, h.secrets.snapshot())
	require.Empty(t, h.state.nonSwitch())
}

func TestSecondConnectInTheSameWorkspaceIsAConflict(t *testing.T) {
	h := newHarness(t)
	h.gw.block = true
	h.gw.entered = make(chan struct{}, 2)
	h.gw.release = make(chan struct{})
	done := make(chan error, 1)
	go func() {
		_, err := h.svc.Connect(h.ctx, ws, ConnectInput{SpaceURL: "a.backlog.com", APIKey: testutil.APIKey(t)})
		done <- err
	}()
	// Fail fast instead of hanging if the first Connect returns before it
	// reaches Backlog (for example when the workspace is off).
	select {
	case <-h.gw.entered:
	case err := <-done:
		require.FailNow(t, "first Connect returned before calling Backlog", "err: %v", err)
	}

	_, err := h.svc.Connect(h.ctx, ws, ConnectInput{SpaceURL: "a.backlog.com", APIKey: testutil.APIKey(t)})
	require.ErrorIs(t, err, ErrConflict)
	require.Equal(t, CodeConflict, Classify(err).Code)

	close(h.gw.release)
	require.NoError(t, <-done)
	// The lock is released on return, so a later Connect proceeds.
	_, err = h.svc.Connect(h.ctx, ws, ConnectInput{SpaceURL: "a.backlog.com", APIKey: testutil.APIKey(t)})
	require.NoError(t, err)
}

func TestConnectsInDifferentWorkspacesRunInParallel(t *testing.T) {
	h := newHarness(t)
	h.gw.block = true
	h.gw.entered = make(chan struct{}, 2)
	h.gw.release = make(chan struct{})
	errs := make(chan error, 2)
	for _, w := range []string{"ws-a", "ws-b"} {
		go func() {
			_, err := h.svc.Connect(h.ctx, w, ConnectInput{SpaceURL: "a.backlog.com", APIKey: testutil.APIKey(t)})
			errs <- err
		}()
	}
	<-h.gw.entered
	<-h.gw.entered // both are inside the Backlog call at the same time
	close(h.gw.release)
	require.NoError(t, <-errs)
	require.NoError(t, <-errs)
}

func TestNewServiceUsesTheDesignBudget(t *testing.T) {
	svc := NewService(&fakeGateway{}, NewStore(newFakeSecrets(), newFakeState()))
	require.Equal(t, 12*time.Second, svc.Deadline)
	require.Equal(t, time.Second, svc.PreCallTimeout)
	require.Equal(t, 10*time.Second, svc.BacklogTimeout)
	require.Equal(t, 2*time.Second, svc.StoreTimeout)
	require.Equal(t, time.Second, svc.store.CallTimeout)
	require.Equal(t, 2*time.Second, svc.store.RollbackTimeout)
	require.LessOrEqual(t, svc.Deadline+svc.store.RollbackTimeout, 14*time.Second, "NFR1.4: 1 s under Kandev's 15 s action limit")
}

// The budget tests run in a synctest bubble: time is virtual, so the real
// production limits are asserted exactly without any real waiting.
func connectTimed(t *testing.T, h *harness) (time.Duration, error) {
	t.Helper()
	start := time.Now()
	_, err := h.svc.Connect(h.ctx, ws, ConnectInput{SpaceURL: "a.backlog.com", APIKey: testutil.APIKey(t)})
	return time.Since(start), err
}

func TestConnectBudget(t *testing.T) {
	t.Run("a stalled first switch read is internal at 1 s", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			h := newHarness(t)
			h.state.blockGet = true
			d, err := connectTimed(t, h)
			require.Equal(t, CodeInternal, Classify(err).Code)
			require.Equal(t, time.Second, d)
			require.Equal(t, 0, h.gw.callCount())
		})
	})
	t.Run("a Backlog that never answers is unreachable at 10 s", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			h := newHarness(t)
			h.gw.block = true
			d, err := connectTimed(t, h)
			require.Equal(t, CodeUnreachable, Classify(err).Code)
			require.Equal(t, 10*time.Second, d)
			require.Empty(t, h.secrets.snapshot())
		})
	})
	t.Run("a Backlog that answers in 2 s connects well under 3 s", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			h := newHarness(t)
			h.gw.delay = 2 * time.Second
			d, err := connectTimed(t, h)
			require.NoError(t, err)
			require.Less(t, d, 3*time.Second, "NFR1.2")
		})
	})
	t.Run("the Backlog call never runs past the deadline minus 2 s", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			h := newHarness(t)
			h.state.getDelay = 900 * time.Millisecond // the first switch read uses most of the pre-call second
			h.gw.block = true
			d, err := connectTimed(t, h)
			require.Equal(t, CodeUnreachable, Classify(err).Code)
			require.Equal(t, 9100*time.Millisecond, h.gw.deadline, "min(10 s, 12 s - 0.9 s - 2 s)")
			require.Equal(t, 10*time.Second, d, "returns at deadline minus 2 s")
		})
	})
	t.Run("the store steps share 2 s and the worst case stays within 14 s", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			h := newHarness(t)
			_, err := h.svc.Connect(h.ctx, ws, ConnectInput{SpaceURL: "old.backlog.com", APIKey: testutil.APIKey(t)})
			require.NoError(t, err)
			before := h.secrets.snapshot()
			h.state.getDelay = 900 * time.Millisecond
			h.gw.delay = 9 * time.Second // answers just inside its 9.1 s limit
			h.secrets.blockSet = true    // the new secret write and its rollback both stall
			d, err := connectTimed(t, h)
			require.Equal(t, CodeInternal, Classify(err).Code)
			// 0.9 s pre-call + 9 s call + 2 s store steps + 2 s rollback.
			require.Equal(t, 13900*time.Millisecond, d)
			require.LessOrEqual(t, d, 14*time.Second, "NFR1.4 total")
			require.Equal(t, before, h.secrets.snapshot())
		})
	})
}

func TestConnectLogsExactlyOneOutcomeEvent(t *testing.T) {
	h := newHarness(t)
	_, err := h.svc.Connect(h.ctx, ws, ConnectInput{SpaceURL: "a.backlog.com", APIKey: testutil.APIKey(t)})
	require.NoError(t, err)
	ok := eventsNamed(h.events(t), "connect_succeeded")
	require.Len(t, ok, 1)
	require.Equal(t, "INFO", ok[0]["level"])
	require.Equal(t, "a.backlog.com", ok[0]["spaceHost"])
	require.EqualValues(t, 1234, ok[0]["backlogUserId"])
	require.EqualValues(t, 1, ok[0]["connectionEpoch"])
	require.Contains(t, ok[0], "durationMs")

	h.gw.err = be(backlog.KindUnauthorized, 401)
	_, _ = h.svc.Connect(h.ctx, ws, ConnectInput{SpaceURL: "a.backlog.com", APIKey: testutil.APIKey(t)})
	_, _ = h.svc.Connect(h.ctx, ws, ConnectInput{SpaceURL: "not valid", APIKey: testutil.APIKey(t)})
	failed := eventsNamed(h.events(t), "connect_failed")
	require.Len(t, failed, 2)
	require.Equal(t, "WARN", failed[0]["level"])
	require.Equal(t, "a.backlog.com", failed[0]["spaceHost"])
	require.Equal(t, CodeValidation, failed[0]["errorCode"])
	require.EqualValues(t, 401, failed[0]["backlogStatus"])
	require.Contains(t, failed[0], "durationMs")
	require.NotContains(t, failed[1], "spaceHost", "an invalid address is never logged")
	require.NotContains(t, failed[1], "backlogStatus")
}

func TestGetReturnsTheStoredView(t *testing.T) {
	h := newHarness(t)
	v, err := h.svc.Get(h.ctx, ws)
	require.NoError(t, err)
	require.Equal(t, StateNotConnected, v.State)
	_, err = h.svc.Connect(h.ctx, ws, ConnectInput{SpaceURL: "a.backlog.com", APIKey: testutil.APIKey(t)})
	require.NoError(t, err)
	v, err = h.svc.Get(h.ctx, ws)
	require.NoError(t, err)
	require.Equal(t, StateConnected, v.State)
	h.state.failGet = true
	_, err = h.svc.Get(h.ctx, ws)
	require.Equal(t, CodeInternal, Classify(err).Code)
}

func TestClassifyUnknownErrorsAreInternal(t *testing.T) {
	require.Equal(t, CodeInternal, Classify(errors.New("boom")).Code)
	require.Equal(t, CodeInternal, Classify(context.Canceled).Code)
	require.Equal(t, CodeUnreachable, Classify(context.DeadlineExceeded).Code)
	require.Equal(t, Outcome{}, Classify(nil))
}

func TestConnectNeverLeaksTheKeyOrDisplayName(t *testing.T) {
	key := testutil.APIKey(t)
	var texts []string
	for _, tc := range wf3Rows {
		h := newHarness(t)
		h.gw.err = tc.gwErr
		h.state.failSet = tc.failStore
		view, err := h.svc.Connect(h.ctx, ws, ConnectInput{SpaceURL: tc.spaceURL, APIKey: " " + key + " "})
		if err != nil {
			texts = append(texts, err.Error())
		}
		b, _ := json.Marshal(view)
		texts = append(texts, string(b))
		testutil.AssertNoLeak(t, h.logs.String(), key, "Test User")
	}
	testutil.AssertNoLeak(t, strings.Join(texts, "\n"), key)
}

func TestSetEnabledWritesOnlyTheSwitchAndReturnsItsValue(t *testing.T) {
	h := newHarness(t)
	_, err := h.svc.Connect(h.ctx, ws, ConnectInput{SpaceURL: "a.backlog.com", APIKey: testutil.APIKey(t)})
	require.NoError(t, err)
	beforeSecrets, setCalls := h.secrets.snapshot(), h.secrets.setCalls
	recordBefore := recordOf(t, h.state, ws)

	view, err := h.svc.SetEnabled(h.ctx, ws, false)
	require.NoError(t, err)
	require.Equal(t, SwitchView{Enabled: false}, view)
	require.Equal(t, beforeSecrets, h.secrets.snapshot())
	require.Equal(t, setCalls, h.secrets.setCalls)
	require.Equal(t, recordBefore, recordOf(t, h.state, ws))

	view, err = h.svc.SetEnabled(h.ctx, ws, true)
	require.NoError(t, err)
	require.Equal(t, SwitchView{Enabled: true}, view)
	full, err := h.svc.Get(h.ctx, ws)
	require.NoError(t, err)
	require.True(t, full.Connected, "the connection is kept, so turning back on needs no reconnect (BR7.4)")
}

func TestSetEnabledLogsOnlyTheSwitchChange(t *testing.T) {
	h := newHarness(t)
	_, err := h.svc.SetEnabled(h.ctx, ws, false)
	require.NoError(t, err)
	_, err = h.svc.SetEnabled(h.ctx, ws, true)
	require.NoError(t, err)
	changed := eventsNamed(h.events(t), "integration_switch_changed")
	require.Len(t, changed, 2)
	require.Equal(t, true, changed[0]["previousEnabled"])
	require.Equal(t, false, changed[0]["enabled"])
	require.Equal(t, false, changed[1]["previousEnabled"])
	require.Equal(t, true, changed[1]["enabled"])
	for _, e := range changed {
		require.Equal(t, "INFO", e["level"])
		keys := make([]string, 0, len(e))
		for k := range e {
			keys = append(keys, k)
		}
		require.ElementsMatch(t, []string{"time", "level", "msg", "event", "previousEnabled", "enabled", "durationMs"}, keys,
			"nothing else is logged; workspaceId and requestId come from the request logger")
	}
}

func TestSetEnabledRepairsAnUndecodableSwitch(t *testing.T) {
	h := newHarness(t)
	h.state.data[stateKey("workspace", ws, "integration")] = map[string]any{"schemaVersion": 1, "enabled": "garbage"}
	view, err := h.svc.SetEnabled(h.ctx, ws, true)
	require.NoError(t, err)
	require.True(t, view.Enabled)
	changed := eventsNamed(h.events(t), "integration_switch_changed")
	require.Len(t, changed, 1)
	require.Equal(t, "unknown", changed[0]["previousEnabled"], "an unreadable previous value does not block the repair")
}

func TestSetEnabledIsSuccessOnceTheSwitchIsWritten(t *testing.T) {
	h := newHarness(t)
	_, err := h.svc.Connect(h.ctx, ws, ConnectInput{SpaceURL: "a.backlog.com", APIKey: testutil.APIKey(t)})
	require.NoError(t, err)
	h.secrets.failGet, h.secrets.getCalls = true, 0
	view, err := h.svc.SetEnabled(h.ctx, ws, false)
	require.NoError(t, err, "a saved switch is never reported as a failure (WF6 step 6)")
	require.False(t, view.Enabled)
	require.Zero(t, h.secrets.getCalls, "no secret access (NFR1.3)")
	enabled, err := h.svc.store.LoadSwitch(h.ctx, ws)
	require.NoError(t, err)
	require.False(t, enabled)
}

func TestSetEnabledWriteFailureIsInternal(t *testing.T) {
	h := newHarness(t)
	h.state.failSet = true
	_, err := h.svc.SetEnabled(h.ctx, ws, false)
	require.Equal(t, CodeInternal, Classify(err).Code)
	require.Empty(t, eventsNamed(h.events(t), "integration_switch_changed"))
}

func TestConnectWhileOffMakesNoCallReadAndWrite(t *testing.T) {
	h := newHarness(t)
	_, err := h.svc.SetEnabled(h.ctx, ws, false)
	require.NoError(t, err)
	stateWrites := h.state.setCalls
	beforeState := h.state.snapshot()

	_, err = h.svc.Connect(h.ctx, ws, ConnectInput{SpaceURL: "a.backlog.com", APIKey: testutil.APIKey(t)})
	require.ErrorIs(t, err, ErrIntegrationDisabled)
	require.Equal(t, CodeIntegrationDisabled, Classify(err).Code)
	require.Equal(t, 0, h.gw.callCount(), "no Backlog call")
	require.Equal(t, 0, h.secrets.getCalls, "no secret read")
	require.Equal(t, 0, h.secrets.setCalls, "no secret write")
	require.Equal(t, stateWrites, h.state.setCalls, "no state write")
	require.Equal(t, beforeState, h.state.snapshot())
}

func TestConnectChecksTheSwitchAgainBeforeStoring(t *testing.T) {
	h := newHarness(t)
	h.gw.onCall = func() { require.NoError(t, h.svc.store.SaveSwitch(context.Background(), ws, false)) }

	view, err := h.svc.Connect(h.ctx, ws, ConnectInput{SpaceURL: "a.backlog.com", APIKey: testutil.APIKey(t)})
	require.ErrorIs(t, err, ErrIntegrationDisabled, "BR7.3")
	require.Equal(t, View{}, view)
	require.Equal(t, 1, h.gw.callCount())
	require.Empty(t, h.secrets.snapshot(), "nothing stored")
	_, hasRecord := h.state.snapshot()[stateKey("workspace", ws, "connection")]
	require.False(t, hasRecord)
}

func TestConnectedViewCarriesEnabled(t *testing.T) {
	h := newHarness(t)
	view, err := h.svc.Connect(h.ctx, ws, ConnectInput{SpaceURL: "a.backlog.com", APIKey: testutil.APIKey(t)})
	require.NoError(t, err)
	require.True(t, view.Enabled)
}
