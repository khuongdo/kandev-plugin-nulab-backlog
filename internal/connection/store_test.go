package connection

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/backlog"
	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/redact"
	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/testutil"
)

const ws = "ws-1"

var testUser = backlog.User{ID: 1234, UserID: "test.user", Name: "Test User"}

func secretOf(t *testing.T, secrets *fakeSecrets, workspaceID string) map[string]any {
	t.Helper()
	raw, ok := secrets.snapshot()["backlog.connection."+workspaceID]
	require.True(t, ok, "secret missing")
	var v map[string]any
	require.NoError(t, json.Unmarshal([]byte(raw), &v))
	return v
}

func recordOf(t *testing.T, state *fakeState, workspaceID string) map[string]any {
	t.Helper()
	v, ok := state.snapshot()[stateKey("workspace", workspaceID, "connection")]
	require.True(t, ok, "record missing")
	return v
}

func TestStoreFirstSaveWritesSecretThenRecordWithEpochOne(t *testing.T) {
	secrets, state := newFakeSecrets(), newFakeState()
	store := newTestStore(secrets, state)
	log, _ := testLogger(t)
	key := testutil.APIKey(t)

	view, err := store.Save(redact.WithLogger(context.Background(), log), ws, "example-space.backlog.com", key, testUser)
	require.NoError(t, err)

	sec := secretOf(t, secrets, ws)
	require.Equal(t, key, sec["apiKey"])
	require.Equal(t, "example-space.backlog.com", sec["spaceHost"])
	require.EqualValues(t, 1, sec["connectionEpoch"])

	rec := recordOf(t, state, ws)
	require.EqualValues(t, 1, rec["connectionEpoch"])
	require.EqualValues(t, 1, rec["schemaVersion"])
	require.Equal(t, "Test User", rec["connectedUserName"])
	require.EqualValues(t, 1234, rec["connectedUserId"])
	require.Equal(t, "api_key", rec["authMethod"])
	require.Equal(t, fixedNow.Format(time.RFC3339), rec["connectedAt"])
	recJSON, _ := json.Marshal(rec)
	require.NotContains(t, string(recJSON), key, "the key must never be in plugin state (BR3.1)")

	require.Equal(t, View{
		Connected: true, State: StateConnected, SpaceHost: "example-space.backlog.com",
		AuthMethod: "api_key", ConnectedUserName: "Test User", HasAPIKey: true, ConnectionEpoch: 1,
	}, view)
}

func TestStoreReplaceIncrementsTheEpoch(t *testing.T) {
	secrets, state := newFakeSecrets(), newFakeState()
	store := newTestStore(secrets, state)
	log, _ := testLogger(t)

	_, err := store.Save(redact.WithLogger(context.Background(), log), ws, "a.backlog.com", testutil.APIKey(t), testUser)
	require.NoError(t, err)
	view, err := store.Save(redact.WithLogger(context.Background(), log), ws, "b.backlog.jp", testutil.APIKey(t), testUser)
	require.NoError(t, err)

	require.Equal(t, 2, view.ConnectionEpoch)
	require.Equal(t, "b.backlog.jp", view.SpaceHost)
	require.EqualValues(t, 2, secretOf(t, secrets, ws)["connectionEpoch"])
	require.EqualValues(t, 2, recordOf(t, state, ws)["connectionEpoch"])
}

func TestStoreRecordFailureRestoresThePreviousSecret(t *testing.T) {
	secrets, state := newFakeSecrets(), newFakeState()
	store := newTestStore(secrets, state)
	log, _ := testLogger(t)
	oldKey := testutil.APIKey(t)
	_, err := store.Save(redact.WithLogger(context.Background(), log), ws, "a.backlog.com", oldKey, testUser)
	require.NoError(t, err)
	before := secrets.snapshot()

	state.failSet = true
	_, err = store.Save(redact.WithLogger(context.Background(), log), ws, "b.backlog.com", testutil.APIKey(t), testUser)
	require.ErrorIs(t, err, ErrStore)
	require.Equal(t, before, secrets.snapshot())

	view, err := store.Load(context.Background(), ws)
	require.NoError(t, err)
	require.Equal(t, StateConnected, view.State)
	require.Equal(t, "a.backlog.com", view.SpaceHost)
}

func TestStoreRecordFailureOnFirstConnectDeletesTheNewSecret(t *testing.T) {
	secrets, state := newFakeSecrets(), newFakeState()
	state.failSet = true
	store := newTestStore(secrets, state)
	log, _ := testLogger(t)

	_, err := store.Save(redact.WithLogger(context.Background(), log), ws, "a.backlog.com", testutil.APIKey(t), testUser)
	require.ErrorIs(t, err, ErrStore)
	require.Empty(t, secrets.snapshot())
}

func TestStoreRollbackFailureLeavesAnErrorViewAndLogsInconsistency(t *testing.T) {
	secrets, state := newFakeSecrets(), newFakeState()
	store := newTestStore(secrets, state)
	log, buf := testLogger(t)
	_, err := store.Save(redact.WithLogger(context.Background(), log), ws, "a.backlog.com", testutil.APIKey(t), testUser)
	require.NoError(t, err)

	state.failSet = true
	secrets.failSet = secrets.setCalls + 2 // the new secret write succeeds, the restore fails
	newKey := testutil.APIKey(t)
	_, err = store.Save(redact.WithLogger(context.Background(), log), ws, "a.backlog.com", newKey, testUser)
	require.ErrorIs(t, err, ErrStore)

	view, err := store.Load(context.Background(), ws)
	require.NoError(t, err)
	require.Equal(t, StateError, view.State)
	require.False(t, view.Connected)
	require.False(t, view.HasAPIKey)

	var event map[string]any
	for _, line := range strings.Split(strings.TrimSpace(buf.String()), "\n") {
		var m map[string]any
		require.NoError(t, json.Unmarshal([]byte(line), &m))
		if m["event"] == "connection_inconsistent" {
			event = m
		}
	}
	require.NotNil(t, event, "connection_inconsistent must be logged")
	require.Equal(t, "ERROR", event["level"])
	require.EqualValues(t, 1, event["previousEpoch"])
	require.EqualValues(t, 2, event["newEpoch"])
	testutil.AssertNoLeak(t, buf.String(), newKey, 8)
}

func TestStoreLoadStates(t *testing.T) {
	key := testutil.APIKey(t)
	record := map[string]any{
		"schemaVersion": 1, "spaceHost": "a.backlog.com", "authMethod": "api_key",
		"connectedUserName": "Test User", "connectedUserId": 1234, "connectionEpoch": 3,
		"connectedAt": fixedNow.Format(time.RFC3339),
	}
	secret := func(host string, epoch int) string {
		b, _ := json.Marshal(map[string]any{"apiKey": key, "spaceHost": host, "connectionEpoch": epoch})
		return string(b)
	}
	cases := []struct {
		name      string
		record    map[string]any
		secret    string
		wantState string
		wantKey   bool
	}{
		{"no record is not_connected", nil, "", StateNotConnected, false},
		{"a leftover secret without a record is not_connected", nil, secret("a.backlog.com", 3), StateNotConnected, false},
		{"a matching secret is connected", record, secret("a.backlog.com", 3), StateConnected, true},
		{"a missing secret is error", record, "", StateError, false},
		{"an epoch mismatch is error", record, secret("a.backlog.com", 4), StateError, false},
		{"a host mismatch is error", record, secret("b.backlog.com", 3), StateError, false},
		{"an unreadable secret is error", record, "not-json", StateError, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			secrets, state := newFakeSecrets(), newFakeState()
			if tc.record != nil {
				state.data[stateKey("workspace", ws, "connection")] = tc.record
			}
			if tc.secret != "" {
				secrets.data["backlog.connection."+ws] = tc.secret
			}
			view, err := newTestStore(secrets, state).Load(context.Background(), ws)
			require.NoError(t, err)
			require.Equal(t, tc.wantState, view.State)
			require.Equal(t, tc.wantState == StateConnected, view.Connected)
			require.Equal(t, tc.wantKey, view.HasAPIKey)
			out, _ := json.Marshal(view)
			require.NotContains(t, string(out), key)
		})
	}
}

func TestStoreLoadFailuresAreStoreErrors(t *testing.T) {
	for _, tc := range []struct {
		name  string
		setup func(*fakeSecrets, *fakeState)
	}{
		{"state read fails", func(_ *fakeSecrets, st *fakeState) { st.failGet = true }},
		{"secret read fails", func(s *fakeSecrets, st *fakeState) {
			st.data[stateKey("workspace", ws, "connection")] = map[string]any{"schemaVersion": 1, "spaceHost": "a.backlog.com", "connectionEpoch": 1}
			s.failGet = true
		}},
		{"an unknown schema version is never overwritten", func(_ *fakeSecrets, st *fakeState) {
			st.data[stateKey("workspace", ws, "connection")] = map[string]any{"schemaVersion": 2}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			secrets, state := newFakeSecrets(), newFakeState()
			tc.setup(secrets, state)
			_, err := newTestStore(secrets, state).Load(context.Background(), ws)
			require.ErrorIs(t, err, ErrStore)
		})
	}
}

func TestStoreRollbackRunsOnAFreshContextWhenTheActionIsCancelled(t *testing.T) {
	secrets, state := newFakeSecrets(), newFakeState()
	store := newTestStore(secrets, state)
	log, _ := testLogger(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	state.cancelSet = cancel

	_, err := store.Save(redact.WithLogger(ctx, log), ws, "a.backlog.com", testutil.APIKey(t), testUser)
	require.True(t, errors.Is(err, context.Canceled), "cancellation is returned unchanged")
	require.Empty(t, secrets.snapshot(), "the rollback deleted the new secret despite the cancelled action")
	require.NoError(t, secrets.ctxErrSet[len(secrets.ctxErrSet)-1], "rollback context must not be the cancelled one")
}

func TestStoreNextEpochNeverDecreasesWhenTheSecretIsMissing(t *testing.T) {
	secrets, state := newFakeSecrets(), newFakeState()
	state.data[stateKey("workspace", ws, "connection")] = map[string]any{
		"schemaVersion": 1, "spaceHost": "a.backlog.com", "connectionEpoch": 7,
	}
	view, err := newTestStore(secrets, state).Save(context.Background(), ws, "a.backlog.com", testutil.APIKey(t), testUser)
	require.NoError(t, err)
	require.Equal(t, 8, view.ConnectionEpoch)
}

func TestStoreSecretWriteFailureWritesNoRecord(t *testing.T) {
	secrets, state := newFakeSecrets(), newFakeState()
	secrets.failSet = 1
	_, err := newTestStore(secrets, state).Save(context.Background(), ws, "a.backlog.com", testutil.APIKey(t), testUser)
	require.ErrorIs(t, err, ErrStore)
	require.Empty(t, state.snapshot())
	require.Empty(t, secrets.snapshot())
}

func TestStorePreviousSecretReadFailureWritesNothing(t *testing.T) {
	secrets, state := newFakeSecrets(), newFakeState()
	secrets.failGet = true
	_, err := newTestStore(secrets, state).Save(context.Background(), ws, "a.backlog.com", testutil.APIKey(t), testUser)
	require.ErrorIs(t, err, ErrStore)
	require.Empty(t, state.snapshot())
	require.Equal(t, 0, secrets.setCalls)
}

func TestStoreCallsAreBoundedByTheCallTimeout(t *testing.T) {
	secrets, state := newFakeSecrets(), newFakeState()
	secrets.blockSet = true
	store := newTestStore(secrets, state)
	store.CallTimeout = 20 * time.Millisecond
	store.RollbackTimeout = 20 * time.Millisecond
	start := time.Now()
	_, err := store.Save(context.Background(), ws, "a.backlog.com", testutil.APIKey(t), testUser)
	require.ErrorIs(t, err, ErrStore)
	require.Less(t, time.Since(start), time.Second)
}

func TestStoreWorkspacesAreIndependent(t *testing.T) {
	secrets, state := newFakeSecrets(), newFakeState()
	store := newTestStore(secrets, state)
	_, err := store.Save(context.Background(), "ws-a", "a.backlog.com", testutil.APIKey(t), testUser)
	require.NoError(t, err)
	_, err = store.Save(context.Background(), "ws-b", "b.backlog.jp", testutil.APIKey(t), testUser)
	require.NoError(t, err)

	// A new Store over the same stores models a plugin restart (NFR5.7).
	restarted := newTestStore(secrets, state)
	a, err := restarted.Load(context.Background(), "ws-a")
	require.NoError(t, err)
	b, err := restarted.Load(context.Background(), "ws-b")
	require.NoError(t, err)
	require.Equal(t, "a.backlog.com", a.SpaceHost)
	require.Equal(t, "b.backlog.jp", b.SpaceHost)
}

func switchOf(state *fakeState, workspaceID string) (map[string]any, bool) {
	v, ok := state.snapshot()[stateKey("workspace", workspaceID, "integration")]
	return v, ok
}

func TestSwitchWithNoRecordIsEnabled(t *testing.T) {
	enabled, err := newTestStore(newFakeSecrets(), newFakeState()).LoadSwitch(context.Background(), ws)
	require.NoError(t, err)
	require.True(t, enabled, "an absent IntegrationSwitch means on (BR7.1)")
}

func TestSwitchSaveRoundTripsWithChangedAtFromTheClock(t *testing.T) {
	secrets, state := newFakeSecrets(), newFakeState()
	store := newTestStore(secrets, state)
	for _, want := range []bool{false, true} {
		t.Run(fmt.Sprintf("enabled=%v", want), func(t *testing.T) {
			require.NoError(t, store.SaveSwitch(context.Background(), ws, want))
			got, err := store.LoadSwitch(context.Background(), ws)
			require.NoError(t, err)
			require.Equal(t, want, got)
			rec, ok := switchOf(state, ws)
			require.True(t, ok)
			require.Equal(t, map[string]any{
				"schemaVersion": float64(1), "enabled": want, "changedAt": fixedNow.Format(time.RFC3339),
			}, rec)
		})
	}
}

func TestSwitchSaveNeverTouchesTheConnectionOrTheSecret(t *testing.T) {
	secrets, state := newFakeSecrets(), newFakeState()
	store := newTestStore(secrets, state)
	_, err := store.Save(context.Background(), ws, "a.backlog.com", testutil.APIKey(t), testUser)
	require.NoError(t, err)
	beforeSecrets := secrets.snapshot()
	beforeRecord := recordOf(t, state, ws)
	setCalls := secrets.setCalls

	require.NoError(t, store.SaveSwitch(context.Background(), ws, false))
	require.NoError(t, store.SaveSwitch(context.Background(), ws, true))

	require.Equal(t, beforeSecrets, secrets.snapshot(), "BR7.4: the secret is untouched")
	require.Equal(t, setCalls, secrets.setCalls, "no secret write at all")
	require.Equal(t, beforeRecord, recordOf(t, state, ws), "BR7.4: the connection record is untouched")
	view, err := store.Load(context.Background(), ws)
	require.NoError(t, err)
	require.Equal(t, StateConnected, view.State, "turning back on restores the connection without a reconnect")
}

func TestSwitchUndecodableValueFailsClosed(t *testing.T) {
	for _, tc := range []struct {
		name  string
		value map[string]any
	}{
		{"enabled is not a boolean", map[string]any{"schemaVersion": 1, "enabled": "no"}},
		{"enabled is missing", map[string]any{"schemaVersion": 1}},
		{"an unknown schema version", map[string]any{"schemaVersion": 2, "enabled": true}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			state := newFakeState()
			state.data[stateKey("workspace", ws, "integration")] = tc.value
			enabled, err := newTestStore(newFakeSecrets(), state).LoadSwitch(context.Background(), ws)
			require.ErrorIs(t, err, ErrStore, "NFR3.9: never read as enabled")
			require.False(t, enabled)
		})
	}
}

func TestSwitchStoreFailuresAreStoreErrors(t *testing.T) {
	state := newFakeState()
	state.failGet = true
	_, err := newTestStore(newFakeSecrets(), state).LoadSwitch(context.Background(), ws)
	require.ErrorIs(t, err, ErrStore)

	state = newFakeState()
	state.failSet = true
	require.ErrorIs(t, newTestStore(newFakeSecrets(), state).SaveSwitch(context.Background(), ws, false), ErrStore)
}

func TestViewCarriesEnabledInEveryState(t *testing.T) {
	setups := map[string]func(*testing.T, *Store, *fakeSecrets){
		StateNotConnected: func(*testing.T, *Store, *fakeSecrets) {},
		StateConnected: func(t *testing.T, s *Store, _ *fakeSecrets) {
			_, err := s.Save(context.Background(), ws, "a.backlog.com", testutil.APIKey(t), testUser)
			require.NoError(t, err)
		},
		StateError: func(t *testing.T, s *Store, sec *fakeSecrets) {
			_, err := s.Save(context.Background(), ws, "a.backlog.com", testutil.APIKey(t), testUser)
			require.NoError(t, err)
			delete(sec.data, "backlog.connection."+ws)
		},
	}
	for state, setup := range setups {
		for _, enabled := range []bool{true, false} {
			t.Run(fmt.Sprintf("%s enabled=%v", state, enabled), func(t *testing.T) {
				secrets, st := newFakeSecrets(), newFakeState()
				store := newTestStore(secrets, st)
				setup(t, store, secrets)
				require.NoError(t, store.SaveSwitch(context.Background(), ws, enabled))
				view, err := store.Load(context.Background(), ws)
				require.NoError(t, err)
				require.Equal(t, state, view.State)
				require.Equal(t, enabled, view.Enabled)
				out, _ := json.Marshal(view)
				require.Contains(t, string(out), fmt.Sprintf(`"enabled":%v`, enabled))
			})
		}
	}
}

func TestLoadFailsWhenTheSwitchCannotBeRead(t *testing.T) {
	state := newFakeState()
	state.data[stateKey("workspace", ws, "integration")] = map[string]any{"schemaVersion": 1, "enabled": 3}
	_, err := newTestStore(newFakeSecrets(), state).Load(context.Background(), ws)
	require.ErrorIs(t, err, ErrStore)
}
