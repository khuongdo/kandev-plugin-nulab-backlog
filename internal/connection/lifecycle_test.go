package connection

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/backlog"
	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/testutil"
)

func TestRecheckReturnsTheFreshUserName(t *testing.T) {
	u := newU2(t)
	key := u.connectKey(t, "a.backlog.com")
	u.gw.user = backlog.User{ID: 1234, Name: "Renamed User"}
	view, err := u.svc.Test(u.ctx, ws)
	require.NoError(t, err)
	require.Equal(t, "Renamed User", view.ConnectedUserName)
	require.True(t, view.Connected)
	require.Equal(t, key, u.gw.creds.APIKey)
	require.Len(t, eventsNamed(u.events(t), "connection_tested"), 1)
}

func TestRecheckRejectedKeyIsReconnectRequiredAndChangesNothing(t *testing.T) {
	for _, kind := range []backlog.Kind{backlog.KindUnauthorized, backlog.KindForbidden} {
		t.Run(kind.String(), func(t *testing.T) {
			u := newU2(t)
			u.connectKey(t, "a.backlog.com")
			beforeSecrets, beforeState := u.secrets.snapshot(), u.state.snapshot()
			u.gw.err = &backlog.Error{Kind: kind, Status: 401}
			_, err := u.svc.Test(u.ctx, ws)
			require.ErrorIs(t, err, ErrReconnectRequired)
			require.Equal(t, Outcome{Code: CodeReconnectRequired}, Classify(err))
			require.Equal(t, beforeSecrets, u.secrets.snapshot())
			require.Equal(t, beforeState, u.state.snapshot())
		})
	}
}

func TestRecheckPassesUnreachableAndRateLimitedThrough(t *testing.T) {
	u := newU2(t)
	u.connectKey(t, "a.backlog.com")
	u.gw.err = &backlog.Error{Kind: backlog.KindRateLimited, Status: 429, RetryAfter: 30 * time.Second}
	_, err := u.svc.Test(u.ctx, ws)
	require.Equal(t, Outcome{Code: CodeRateLimited, RetryAfterSeconds: 30}, Classify(err))
	u.gw.err = &backlog.Error{Kind: backlog.KindUnreachable, Status: 503}
	_, err = u.svc.Test(u.ctx, ws)
	require.Equal(t, CodeUnreachable, Classify(err).Code)
	_, err = newU2(t).svc.Test(u.ctx, "ws-none")
	require.Equal(t, CodeReconnectRequired, Classify(err).Code)
}

func TestDisconnectDeletesTheCredentialsAndSendsOneEvent(t *testing.T) {
	u := newU2(t)
	u.connectKey(t, "a.backlog.com")
	_ = u.start(t)
	events := subscribe(t, u.svc)
	myselfBefore, _, _, _ := u.gw.counts()

	view, err := u.svc.Disconnect(u.ctx, ws)
	require.NoError(t, err)
	require.Equal(t, StateNotConnected, view.State)
	require.Empty(t, u.secrets.snapshot())
	_, pending := u.state.snapshot()[stateKey("workspace", ws, "oauth_pending")]
	require.False(t, pending, "the pending OAuth sign-in is deleted")
	e := events.next(t)
	require.Equal(t, ConnectionChanged{WorkspaceID: ws, Reason: ReasonDisconnected, ConnectionEpoch: 2}, e)
	require.Len(t, u.changedLogs(t), 2, "connect + disconnect")

	_, _, err = u.svc.Credentials(u.ctx, ws)
	require.ErrorIs(t, err, ErrNotConnected)
	myself, projects, _, _ := u.gw.counts()
	require.Equal(t, myselfBefore, myself, "the gateway sees 0 calls")
	require.Zero(t, projects)
}

func TestDisconnectWhenNotConnectedIsANoOp(t *testing.T) {
	u := newU2(t)
	view, err := u.svc.Disconnect(u.ctx, ws)
	require.NoError(t, err)
	require.Equal(t, StateNotConnected, view.State)
	require.Empty(t, u.changedLogs(t))
}

func TestReplaceSameHostIsCredentialsReplaced(t *testing.T) {
	u := newU2(t)
	u.connectKey(t, "a.backlog.com")
	_, err := u.svc.store.SaveProjects(u.ctx, ws, []string{"PROJ"})
	require.NoError(t, err)
	events := subscribe(t, u.svc)
	newKey := u.connectKey(t, "A.backlog.com")
	e := events.next(t)
	require.Equal(t, ReasonCredentialsReplaced, e.Reason)
	require.Equal(t, 3, e.ConnectionEpoch)
	require.Equal(t, []string{"PROJ"}, e.SelectedProjects, "selected projects kept")
	creds, _, err := u.svc.Credentials(u.ctx, ws)
	require.NoError(t, err)
	require.Equal(t, newKey, creds.APIKey, "only the new key is used (AC1.6.1)")
}

func TestReplaceRejectedKeyKeepsTheOldConnection(t *testing.T) {
	u := newU2(t)
	u.connectKey(t, "a.backlog.com")
	beforeSecrets, beforeState := u.secrets.snapshot(), u.state.snapshot()
	u.gw.err = &backlog.Error{Kind: backlog.KindUnauthorized, Status: 401}
	_, err := u.svc.Connect(u.ctx, ws, ConnectInput{SpaceURL: "a.backlog.com", APIKey: testutil.APIKey(t)})
	require.Error(t, err)
	require.Equal(t, beforeSecrets, u.secrets.snapshot())
	require.Equal(t, beforeState, u.state.snapshot())
	require.Len(t, u.changedLogs(t), 1, "only the first connect")
}

func TestSpaceChangeClearsProjectsAndLetsLateResultsBeDropped(t *testing.T) {
	u := newU2(t)
	oldKey := u.connectKey(t, "a.backlog.com")
	_, err := u.svc.store.SaveProjects(u.ctx, ws, []string{"PROJ"})
	require.NoError(t, err)
	_, oldEpoch, err := u.svc.Credentials(u.ctx, ws)
	require.NoError(t, err)
	events := subscribe(t, u.svc)

	u.connectKey(t, "b.backlog.com")
	e := events.next(t)
	require.Equal(t, ReasonSpaceChanged, e.Reason)
	require.Equal(t, "b.backlog.com", e.SpaceHost)
	require.Empty(t, e.SelectedProjects)
	require.False(t, e.Restore)
	snap, err := u.svc.Current(u.ctx, ws)
	require.NoError(t, err)
	require.Greater(t, snap.ConnectionEpoch, oldEpoch, "a late result from space A carries the lower epoch (AC1.8.3)")
	creds, _, err := u.svc.Credentials(u.ctx, ws)
	require.NoError(t, err)
	require.NotEqual(t, oldKey, creds.APIKey, "the secret of space A is replaced")
}

func TestRestoreReconnectingToTheRememberedHost(t *testing.T) {
	t.Run("after a disconnect", func(t *testing.T) {
		u := newU2(t)
		u.connectKey(t, "a.backlog.com")
		_, _ = u.svc.store.SaveProjects(u.ctx, ws, []string{"PROJ"})
		_, err := u.svc.Disconnect(u.ctx, ws)
		require.NoError(t, err)
		events := subscribe(t, u.svc)
		view, err := u.svc.Connect(u.ctx, ws, ConnectInput{SpaceURL: "a.backlog.com", APIKey: testutil.APIKey(t)})
		require.NoError(t, err)
		require.True(t, view.Restored)
		require.Equal(t, []string{"PROJ"}, view.SelectedProjects)
		e := events.next(t)
		require.Equal(t, ReasonConnected, e.Reason)
		require.True(t, e.Restore)
	})
	t.Run("after a space change", func(t *testing.T) {
		u := newU2(t)
		u.connectKey(t, "a.backlog.com")
		_, _ = u.svc.store.SaveProjects(u.ctx, ws, []string{"PROJ"})
		u.connectKey(t, "b.backlog.com")
		events := subscribe(t, u.svc)
		view, err := u.svc.Connect(u.ctx, ws, ConnectInput{SpaceURL: "a.backlog.com", APIKey: testutil.APIKey(t)})
		require.NoError(t, err)
		require.True(t, view.Restored)
		e := events.next(t)
		require.Equal(t, ReasonSpaceChanged, e.Reason)
		require.True(t, e.Restore)
		require.Equal(t, []string{"PROJ"}, e.SelectedProjects)
	})
	t.Run("through OAuth", func(t *testing.T) {
		u := newU2(t)
		u.connectKey(t, "example-space.backlog.com")
		_, err := u.svc.Disconnect(u.ctx, ws)
		require.NoError(t, err)
		state := u.start(t)
		u.okTokens(t)
		res := u.svc.CompleteOAuth(u.ctx, map[string][]string{"state": {state}, "code": {"c-1234"}})
		require.Equal(t, OAuthResult{WorkspaceID: ws, Outcome: OutcomeConnected, Restored: true}, res)
	})
}
