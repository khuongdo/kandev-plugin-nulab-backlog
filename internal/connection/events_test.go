package connection

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/backlog"
	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/testutil"
)

func TestConnectionChangedArrivesInEpochOrder(t *testing.T) {
	u := newU2(t)
	u.gw.projects = spaceProjects
	events := subscribe(t, u.svc)
	u.connectKey(t, "a.backlog.com")
	_, err := u.svc.SetProjects(u.ctx, ws, []any{"PROJ"})
	require.NoError(t, err)
	u.connectKey(t, "b.backlog.com")
	_, err = u.svc.Disconnect(u.ctx, ws)
	require.NoError(t, err)

	var got []string
	for epoch := 1; epoch <= 4; epoch++ {
		e := events.next(t)
		require.Equal(t, epoch, e.ConnectionEpoch)
		require.Equal(t, ws, e.WorkspaceID)
		got = append(got, string(e.Reason))
	}
	require.Equal(t, []string{"connected", "projects_changed", "space_changed", "disconnected"}, got)
	logs := u.changedLogs(t)
	require.Len(t, logs, 4)
	require.Equal(t, "projects_changed", logs[1]["reason"])
	require.EqualValues(t, 2, logs[1]["connectionEpoch"])
	require.Equal(t, true, logs[1]["restore"])
	require.EqualValues(t, 1, logs[1]["projectCount"])
}

func TestConnectionChangedSlowSubscriberDoesNotBlockTheAction(t *testing.T) {
	u := newU2(t)
	release := make(chan struct{})
	got := make(chan ConnectionChanged, 4)
	unsub := u.svc.Subscribe(func(e ConnectionChanged) { <-release; got <- e })
	defer unsub()
	u.connectKey(t, "a.backlog.com") // returns while the subscriber is stuck
	u.connectKey(t, "b.backlog.com")
	close(release)
	for _, want := range []Reason{ReasonConnected, ReasonSpaceChanged} {
		select {
		case e := <-got:
			require.Equal(t, want, e.Reason)
		case <-time.After(5 * time.Second):
			t.Fatal("event not delivered")
		}
	}
}

func TestConnectionChangedPanickingSubscriberIsRecovered(t *testing.T) {
	u := newU2(t)
	unsub := u.svc.Subscribe(func(ConnectionChanged) { panic("subscriber exploded") })
	defer unsub()
	events := subscribe(t, u.svc)
	u.connectKey(t, "a.backlog.com")
	require.Equal(t, ReasonConnected, events.next(t).Reason)
	u.connectKey(t, "b.backlog.com")
	require.Equal(t, ReasonSpaceChanged, events.next(t).Reason, "the other subscriber keeps receiving")
	require.Eventually(t, func() bool {
		return len(eventsNamed(u.events(t), "connection_subscriber_panic")) == 2
	}, 5*time.Second, time.Millisecond)
}

func TestConnectionChangedOnlyAfterTheWriteSucceeded(t *testing.T) {
	u := newU2(t)
	u.state.failSet = true
	_, err := u.svc.Connect(u.ctx, ws, ConnectInput{SpaceURL: "a.backlog.com", APIKey: testutil.APIKey(t)})
	require.Error(t, err)
	require.Empty(t, u.changedLogs(t))
}

func TestConnectionChangedUnsubscribeStopsDelivery(t *testing.T) {
	u := newU2(t)
	events := subscribe(t, u.svc)
	events.unsub()
	events.unsub() // twice is harmless
	u.connectKey(t, "a.backlog.com")
	require.Len(t, u.changedLogs(t), 1)
	select {
	case e := <-events.ch:
		t.Fatalf("unexpected event %v", e.Reason)
	default:
	}
}

func TestU2_LogsOneLinePerOutcome(t *testing.T) {
	u := newU2(t)
	state := u.start(t)
	u.okTokens(t)
	require.Equal(t, OutcomeConnected, u.svc.CompleteOAuth(u.ctx, url.Values{"state": {state}, "code": {"c-1234"}}).Outcome)
	require.Equal(t, OutcomeFailed, u.svc.CompleteOAuth(u.ctx, url.Values{"state": {state}, "code": {"c-1234"}}).Outcome)
	_, err := u.svc.Test(u.ctx, ws)
	require.NoError(t, err)
	u.advance(2 * time.Hour)
	u.gw.refreshed = backlog.TokenSet{AccessToken: testutil.Token(t), RefreshToken: testutil.Token(t), ExpiresAt: u.clock().Add(time.Hour)}
	_, _, err = u.svc.Credentials(u.ctx, ws)
	require.NoError(t, err)
	u.advance(2 * time.Hour)
	u.gw.refreshErr = &backlog.Error{Kind: backlog.KindInvalid, Status: 400}
	_, _, _ = u.svc.Credentials(u.ctx, ws)

	all := u.events(t)
	for name, n := range map[string]int{
		"oauth_started": 1, "oauth_completed": 1, "oauth_failed": 1, "token_refreshed": 1,
		"token_refresh_failed": 1, "connection_tested": 1, "connection_changed": 1,
	} {
		require.Len(t, eventsNamed(all, name), n, name)
	}
	failed := eventsNamed(all, "oauth_failed")[0]
	require.Equal(t, "bad_state", failed["reason"])
	require.Equal(t, "WARN", failed["level"])
	changed := eventsNamed(all, "connection_changed")[0]
	for _, k := range []string{"reason", "connectionEpoch", "restore", "projectCount"} {
		require.Contains(t, changed, k)
	}
}

func TestU2_NoSecretInLogsErrorsViewsOrEvents(t *testing.T) {
	u := newU2(t)
	events := subscribe(t, u.svc)
	var texts []string
	record := func(v any, err error) {
		b, _ := json.Marshal(v)
		texts = append(texts, string(b))
		if err != nil {
			texts = append(texts, err.Error())
		}
	}
	key := testutil.APIKey(t)
	record(u.svc.Connect(u.ctx, ws, ConnectInput{SpaceURL: "example-space.backlog.com", APIKey: key}))
	record(u.svc.Test(u.ctx, ws))
	u.gw.err = &backlog.Error{Kind: backlog.KindUnauthorized, Status: 401}
	record(u.svc.Test(u.ctx, ws))
	u.gw.err = nil

	res, err := u.svc.StartOAuth(u.ctx, ws, StartInput{SpaceURL: "example-space.backlog.com"})
	require.NoError(t, err)
	au, _ := url.Parse(res.AuthorizeURL)
	state := au.Query().Get("state")
	tokens := u.okTokens(t)
	code := testutil.Token(t)
	record(u.svc.CompleteOAuth(u.ctx, url.Values{"state": {state}, "code": {code}}), nil)
	u.advance(2 * time.Hour)
	refreshed := backlog.TokenSet{AccessToken: testutil.Token(t), RefreshToken: testutil.Token(t), ExpiresAt: u.clock().Add(time.Hour)}
	u.gw.refreshed = refreshed
	record(u.svc.Get(u.ctx, ws))
	_, _, err = u.svc.Credentials(u.ctx, ws)
	require.NoError(t, err)
	u.advance(2 * time.Hour)
	u.gw.refreshErr = &backlog.Error{Kind: backlog.KindUnauthorized, Status: 401}
	_, _, err = u.svc.Credentials(u.ctx, ws)
	record(nil, err)
	record(u.svc.Disconnect(u.ctx, ws))
	for range 3 {
		texts = append(texts, fmt.Sprintf("%+v", events.next(t)))
	}

	all := strings.Join(texts, "\n")
	logs := u.logs.String()
	for _, secret := range []string{key, tokens.AccessToken, tokens.RefreshToken, refreshed.AccessToken, refreshed.RefreshToken, u.clientSecret, code} {
		testutil.AssertNoLeak(t, all, secret, 8)
		testutil.AssertNoLeak(t, logs, secret, 8)
	}
	testutil.AssertNoLeak(t, all+logs, "", 8, state, state[:16])
	testutil.AssertNoLeak(t, logs, "", 8, "Test User")
}
