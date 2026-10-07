package connection

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"testing/synctest"
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
	testutil.AssertNoLeak(t, buf.String(), newKey)
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

// ---- U2 store cases ----

func oauthTokens(t *testing.T, expiresAt time.Time) backlog.TokenSet {
	t.Helper()
	return backlog.TokenSet{AccessToken: testutil.Token(t), RefreshToken: testutil.Token(t), ExpiresAt: expiresAt}
}

func TestOAuthSaveWritesTheTokenSecretAndRecord(t *testing.T) {
	secrets, state := newFakeSecrets(), newFakeState()
	store := newTestStore(secrets, state)
	tokens := oauthTokens(t, fixedNow.Add(time.Hour))

	view, err := store.SaveOAuth(context.Background(), ws, "a.backlog.com", tokens, testUser)
	require.NoError(t, err)

	sec := secretOf(t, secrets, ws)
	require.Equal(t, "oauth", sec["authMethod"])
	require.Equal(t, tokens.AccessToken, sec["accessToken"])
	require.Equal(t, tokens.RefreshToken, sec["refreshToken"])
	require.Equal(t, fixedNow.Add(time.Hour).Format(time.RFC3339), sec["expiresAt"])
	require.NotContains(t, sec, "apiKey")
	require.EqualValues(t, 1, sec["connectionEpoch"])
	rec := recordOf(t, state, ws)
	require.Equal(t, "oauth", rec["authMethod"])
	recJSON, _ := json.Marshal(rec)
	require.NotContains(t, string(recJSON), tokens.AccessToken)
	require.Equal(t, StateConnected, view.State)
	require.Equal(t, "oauth", view.AuthMethod)
	require.True(t, view.HasOAuthToken)
	require.False(t, view.HasAPIKey)
	require.Equal(t, 1, view.ConnectionEpoch)
}

func TestOAuthSaveRecordFailureRollsBackTheNewSecret(t *testing.T) {
	secrets, state := newFakeSecrets(), newFakeState()
	state.failSet = true
	_, err := newTestStore(secrets, state).SaveOAuth(context.Background(), ws, "a.backlog.com", oauthTokens(t, fixedNow), testUser)
	require.ErrorIs(t, err, ErrStore)
	require.Empty(t, secrets.snapshot(), "secret written first, then deleted by the U1 rollback")
}

func TestTokenUpdateTokensRewritesOnlyTheSecretAndKeepsTheEpoch(t *testing.T) {
	secrets, state := newFakeSecrets(), newFakeState()
	store := newTestStore(secrets, state)
	_, err := store.SaveOAuth(context.Background(), ws, "a.backlog.com", oauthTokens(t, fixedNow), testUser)
	require.NoError(t, err)
	recBefore, stateWrites := recordOf(t, state, ws), state.setCalls
	next := oauthTokens(t, fixedNow.Add(time.Hour))

	require.NoError(t, store.UpdateTokens(context.Background(), ws, 1, next))
	sec := secretOf(t, secrets, ws)
	require.Equal(t, next.AccessToken, sec["accessToken"])
	require.Equal(t, next.RefreshToken, sec["refreshToken"], "the rotated refresh token is stored (AC1.4.2)")
	require.EqualValues(t, 1, sec["connectionEpoch"])
	require.Equal(t, stateWrites, state.setCalls, "the record is never touched")
	require.Equal(t, recBefore, recordOf(t, state, ws))
	view, err := store.Load(context.Background(), ws)
	require.NoError(t, err)
	require.Equal(t, StateConnected, view.State)
}

func TestTokenUpdateTokensWithAStaleEpochWritesNothing(t *testing.T) {
	secrets, state := newFakeSecrets(), newFakeState()
	store := newTestStore(secrets, state)
	_, err := store.SaveOAuth(context.Background(), ws, "a.backlog.com", oauthTokens(t, fixedNow), testUser)
	require.NoError(t, err)
	before, setCalls := secrets.snapshot(), secrets.setCalls
	err = store.UpdateTokens(context.Background(), ws, 7, oauthTokens(t, fixedNow))
	require.ErrorIs(t, err, ErrStale)
	require.Equal(t, before, secrets.snapshot())
	require.Equal(t, setCalls, secrets.setCalls)
}

func TestTokenMarkSignInAgainChangesTheViewState(t *testing.T) {
	secrets, state := newFakeSecrets(), newFakeState()
	store := newTestStore(secrets, state)
	_, err := store.SaveOAuth(context.Background(), ws, "a.backlog.com", oauthTokens(t, fixedNow), testUser)
	require.NoError(t, err)
	require.ErrorIs(t, store.MarkSignInAgain(context.Background(), ws, 9), ErrStale)
	require.NoError(t, store.MarkSignInAgain(context.Background(), ws, 1))
	require.Equal(t, true, recordOf(t, state, ws)["signInAgain"])
	view, err := store.Load(context.Background(), ws)
	require.NoError(t, err)
	require.Equal(t, StateSignInAgain, view.State)
	require.False(t, view.Connected)
	require.Equal(t, "a.backlog.com", view.SpaceHost)
}

func TestOAuthPendingStateIsSingleUse(t *testing.T) {
	secrets, state := newFakeSecrets(), newFakeState()
	store := newTestStore(secrets, state)
	nonce, err := newNonce()
	require.NoError(t, err)
	verifier := testutil.Token(t)
	require.NoError(t, store.SavePending(context.Background(), ws, nonce, hashHex(verifier), "a.backlog.com", fixedNow.Add(10*time.Minute)))

	rec, ok := state.snapshot()[stateKey("workspace", ws, "oauth_pending")]
	require.True(t, ok)
	require.EqualValues(t, 1, rec["schemaVersion"])
	require.Equal(t, "a.backlog.com", rec["spaceHost"])
	require.Equal(t, fixedNow.Add(10*time.Minute).Format(time.RFC3339), rec["expiresAt"])
	require.Len(t, rec["nonceHash"], 64, "hex SHA-256, never the nonce")
	require.Equal(t, hashHex(verifier), rec["verifierHash"], "only the hash of the browser verifier")

	_, found, err := store.TakePending(context.Background(), ws, nonce, testutil.Token(t))
	require.NoError(t, err)
	require.False(t, found, "another browser's verifier never matches")
	_, kept := state.snapshot()[stateKey("workspace", ws, "oauth_pending")]
	require.True(t, kept, "and does not use up the sign-in")
	require.NotContains(t, fmt.Sprint(rec), fmt.Sprintf("%x", nonce))

	host, found, err := store.TakePending(context.Background(), ws, nonce, verifier)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, "a.backlog.com", host)
	_, still := state.snapshot()[stateKey("workspace", ws, "oauth_pending")]
	require.False(t, still, "deleted before it is returned")

	_, found, err = store.TakePending(context.Background(), ws, nonce, verifier)
	require.NoError(t, err)
	require.False(t, found, "a used state is refused")
}

func TestOAuthPendingWrongNonceAndExpiryAreRefused(t *testing.T) {
	secrets, state := newFakeSecrets(), newFakeState()
	store := newTestStore(secrets, state)
	nonce, _ := newNonce()
	other, _ := newNonce()
	verifier := testutil.Token(t)
	require.NoError(t, store.SavePending(context.Background(), ws, nonce, hashHex(verifier), "a.backlog.com", fixedNow.Add(10*time.Minute)))

	_, found, err := store.TakePending(context.Background(), ws, other, verifier)
	require.NoError(t, err)
	require.False(t, found, "a different nonce never matches")

	store.Now = func() time.Time { return fixedNow.Add(10*time.Minute + time.Second) }
	_, found, err = store.TakePending(context.Background(), ws, nonce, verifier)
	require.NoError(t, err)
	require.False(t, found, "an expired state is refused")
	_, still := state.snapshot()[stateKey("workspace", ws, "oauth_pending")]
	require.False(t, still, "an expired state is deleted")
}

func TestOAuthPendingDeleteAndUndecodable(t *testing.T) {
	secrets, state := newFakeSecrets(), newFakeState()
	store := newTestStore(secrets, state)
	nonce, _ := newNonce()
	verifier := testutil.Token(t)
	require.NoError(t, store.SavePending(context.Background(), ws, nonce, hashHex(verifier), "a.backlog.com", fixedNow.Add(time.Minute)))
	require.NoError(t, store.DeletePending(context.Background(), ws))
	require.Empty(t, state.snapshot())

	state.data[stateKey("workspace", ws, "oauth_pending")] = map[string]any{"schemaVersion": 1, "nonceHash": 3}
	_, found, err := store.TakePending(context.Background(), ws, nonce, verifier)
	require.NoError(t, err)
	require.False(t, found)

	state.failDel = true
	require.ErrorIs(t, store.DeletePending(context.Background(), ws), ErrStore)
}

func TestDisconnectDeletesTheSecretThenWritesADisconnectRecord(t *testing.T) {
	secrets, state := newFakeSecrets(), newFakeState()
	store := newTestStore(secrets, state)
	_, err := store.Save(context.Background(), ws, "a.backlog.com", testutil.APIKey(t), testUser)
	require.NoError(t, err)
	_, err = store.SaveProjects(context.Background(), ws, []string{"PROJ"})
	require.NoError(t, err)

	view, changed, err := store.Disconnect(context.Background(), ws)
	require.NoError(t, err)
	require.True(t, changed)
	require.Empty(t, secrets.snapshot(), "no API key, token or refresh token is left")
	rec := recordOf(t, state, ws)
	require.Equal(t, true, rec["disconnected"])
	require.EqualValues(t, 3, rec["connectionEpoch"], "epoch + 1")
	require.Equal(t, "a.backlog.com", rec["previousSpaceHost"])
	require.Equal(t, []any{"PROJ"}, rec["previousProjects"])
	require.Empty(t, rec["connectedUserName"])
	require.Empty(t, rec["spaceHost"])
	require.Equal(t, StateNotConnected, view.State)
	require.Equal(t, 3, view.ConnectionEpoch)

	loaded, err := store.Load(context.Background(), ws)
	require.NoError(t, err)
	require.Equal(t, StateNotConnected, loaded.State)

	again, err := store.Save(context.Background(), ws, "b.backlog.com", testutil.APIKey(t), testUser)
	require.NoError(t, err)
	require.Equal(t, 4, again.ConnectionEpoch, "the epoch never goes back")
}

func TestDisconnectWhenNotConnectedChangesNothing(t *testing.T) {
	secrets, state := newFakeSecrets(), newFakeState()
	store := newTestStore(secrets, state)
	_, changed, err := store.Disconnect(context.Background(), ws)
	require.NoError(t, err)
	require.False(t, changed)
	require.Empty(t, state.snapshot())

	_, err = store.Save(context.Background(), ws, "a.backlog.com", testutil.APIKey(t), testUser)
	require.NoError(t, err)
	_, _, err = store.Disconnect(context.Background(), ws)
	require.NoError(t, err)
	writes := state.setCalls
	_, changed, err = store.Disconnect(context.Background(), ws)
	require.NoError(t, err)
	require.False(t, changed)
	require.Equal(t, writes, state.setCalls)
}

func TestDisconnectSecretDeleteFailureWritesNoRecord(t *testing.T) {
	secrets, state := newFakeSecrets(), newFakeState()
	store := newTestStore(secrets, state)
	_, err := store.Save(context.Background(), ws, "a.backlog.com", testutil.APIKey(t), testUser)
	require.NoError(t, err)
	before := recordOf(t, state, ws)
	secrets.failDel = true
	_, _, err = store.Disconnect(context.Background(), ws)
	require.ErrorIs(t, err, ErrStore)
	require.Equal(t, before, recordOf(t, state, ws))
}

func TestProjectsSaveProjectsRewritesTheSecretThenTheRecord(t *testing.T) {
	secrets, state := newFakeSecrets(), newFakeState()
	store := newTestStore(secrets, state)
	key := testutil.APIKey(t)
	_, err := store.Save(context.Background(), ws, "a.backlog.com", key, testUser)
	require.NoError(t, err)

	view, err := store.SaveProjects(context.Background(), ws, []string{"PROJ", "DEMO"})
	require.NoError(t, err)
	require.Equal(t, []string{"PROJ", "DEMO"}, view.SelectedProjects)
	require.Equal(t, 2, view.ConnectionEpoch)
	sec := secretOf(t, secrets, ws)
	require.Equal(t, key, sec["apiKey"], "same values")
	require.EqualValues(t, 2, sec["connectionEpoch"])
	rec := recordOf(t, state, ws)
	require.EqualValues(t, 2, rec["connectionEpoch"])
	require.Equal(t, []any{"PROJ", "DEMO"}, rec["selectedProjects"])
	require.Equal(t, "Test User", rec["connectedUserName"])
}

func TestProjectsSaveProjectsRecordFailureRestoresThePreviousSecret(t *testing.T) {
	secrets, state := newFakeSecrets(), newFakeState()
	store := newTestStore(secrets, state)
	_, err := store.Save(context.Background(), ws, "a.backlog.com", testutil.APIKey(t), testUser)
	require.NoError(t, err)
	before := secrets.snapshot()
	state.failSet = true
	_, err = store.SaveProjects(context.Background(), ws, []string{"PROJ"})
	require.ErrorIs(t, err, ErrStore)
	require.Equal(t, before, secrets.snapshot())
}

func TestProjectsSaveProjectsNeedsAConnection(t *testing.T) {
	_, err := newTestStore(newFakeSecrets(), newFakeState()).SaveProjects(context.Background(), ws, []string{"PROJ"})
	require.ErrorIs(t, err, ErrNotConnected)
}

func TestSpaceChangeClearsTheSelectionAndRemembersThePreviousHost(t *testing.T) {
	secrets, state := newFakeSecrets(), newFakeState()
	store := newTestStore(secrets, state)
	_, err := store.Save(context.Background(), ws, "a.backlog.com", testutil.APIKey(t), testUser)
	require.NoError(t, err)
	_, err = store.SaveProjects(context.Background(), ws, []string{"PROJ"})
	require.NoError(t, err)

	view, err := store.Save(context.Background(), ws, "b.backlog.com", testutil.APIKey(t), testUser)
	require.NoError(t, err)
	require.Empty(t, view.SelectedProjects)
	rec := recordOf(t, state, ws)
	require.Equal(t, "a.backlog.com", rec["previousSpaceHost"])
	require.Equal(t, []any{"PROJ"}, rec["previousProjects"])
	require.NotContains(t, rec, "selectedProjects")

	same, err := store.Save(context.Background(), ws, "b.backlog.com", testutil.APIKey(t), testUser)
	require.NoError(t, err)
	require.Equal(t, "b.backlog.com", same.SpaceHost, "same host keeps the selection and the memory")
	require.Equal(t, "a.backlog.com", recordOf(t, state, ws)["previousSpaceHost"])
}

func TestRestorePutsBackThePreviousProjects(t *testing.T) {
	t.Run("after a space change", func(t *testing.T) {
		secrets, state := newFakeSecrets(), newFakeState()
		store := newTestStore(secrets, state)
		_, _ = store.Save(context.Background(), ws, "a.backlog.com", testutil.APIKey(t), testUser)
		_, _ = store.SaveProjects(context.Background(), ws, []string{"PROJ"})
		_, _ = store.Save(context.Background(), ws, "b.backlog.com", testutil.APIKey(t), testUser)
		_, _ = store.SaveProjects(context.Background(), ws, []string{"DEMO"})
		view, err := store.Save(context.Background(), ws, "a.backlog.com", testutil.APIKey(t), testUser)
		require.NoError(t, err)
		require.Equal(t, []string{"PROJ"}, view.SelectedProjects)
		require.Equal(t, "b.backlog.com", recordOf(t, state, ws)["previousSpaceHost"])
	})
	t.Run("after a disconnect", func(t *testing.T) {
		secrets, state := newFakeSecrets(), newFakeState()
		store := newTestStore(secrets, state)
		_, _ = store.Save(context.Background(), ws, "a.backlog.com", testutil.APIKey(t), testUser)
		_, _ = store.SaveProjects(context.Background(), ws, []string{"PROJ"})
		_, _, _ = store.Disconnect(context.Background(), ws)
		view, err := store.SaveOAuth(context.Background(), ws, "a.backlog.com", oauthTokens(t, fixedNow), testUser)
		require.NoError(t, err)
		require.Equal(t, []string{"PROJ"}, view.SelectedProjects)
		require.NotContains(t, recordOf(t, state, ws), "previousSpaceHost", "nothing left to restore")
	})
}

func TestU2_StoreCallsKeepTheOneSecondLimit(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		state := newFakeState()
		state.blockGet = true
		store := newTestStore(newFakeSecrets(), state)
		nonce, verifier := []byte("n"), "v"
		for name, call := range map[string]func() error{
			"TakePending":  func() error { _, _, err := store.TakePending(context.Background(), ws, nonce, verifier); return err },
			"SaveProjects": func() error { _, err := store.SaveProjects(context.Background(), ws, nil); return err },
			"Disconnect":   func() error { _, _, err := store.Disconnect(context.Background(), ws); return err },
		} {
			start := time.Now()
			require.ErrorIs(t, call(), ErrStore, name)
			require.Equal(t, time.Second, time.Since(start), name)
		}
	})
}

func TestU2_WorkspacesStayIndependent(t *testing.T) {
	secrets, state := newFakeSecrets(), newFakeState()
	store := newTestStore(secrets, state)
	_, _ = store.Save(context.Background(), "ws-a", "a.backlog.com", testutil.APIKey(t), testUser)
	_, _ = store.Save(context.Background(), "ws-b", "b.backlog.com", testutil.APIKey(t), testUser)
	_, _, err := store.Disconnect(context.Background(), "ws-a")
	require.NoError(t, err)
	b, err := store.Load(context.Background(), "ws-b")
	require.NoError(t, err)
	require.Equal(t, StateConnected, b.State)
}

// U4: the Git secret backlog.git.<ws> (US5.5, AC1.5.4, AC1.8.2).

const gitKey = "backlog.git." + ws

func gitSecretOf(t *testing.T, secrets *fakeSecrets) (map[string]any, bool) {
	t.Helper()
	raw, ok := secrets.snapshot()[gitKey]
	if !ok {
		return nil, false
	}
	var v map[string]any
	require.NoError(t, json.Unmarshal([]byte(raw), &v))
	return v, true
}

func connectedStore(t *testing.T, host string) (*Store, *fakeSecrets, *fakeState, context.Context) {
	t.Helper()
	secrets, state := newFakeSecrets(), newFakeState()
	store := newTestStore(secrets, state)
	log, _ := testLogger(t)
	ctx := redact.WithLogger(context.Background(), log)
	_, err := store.Save(ctx, ws, host, testutil.APIKey(t), testUser)
	require.NoError(t, err)
	return store, secrets, state, ctx
}

func TestU4_GitStore_SaveBindsToTheConnectedHost(t *testing.T) {
	store, secrets, _, ctx := connectedStore(t, "a.backlog.com")
	pw := testutil.Token(t)
	require.NoError(t, store.SaveGit(ctx, ws, "lan", pw))
	sec, ok := gitSecretOf(t, secrets)
	require.True(t, ok)
	require.Equal(t, map[string]any{"username": "lan", "password": pw, "spaceHost": "a.backlog.com", "revision": float64(1)}, sec)

	require.NoError(t, store.SaveGit(ctx, ws, "lan2", testutil.Token(t)))
	sec, _ = gitSecretOf(t, secrets)
	require.EqualValues(t, 2, sec["revision"], "every save bumps the revision")

	view, err := store.Load(ctx, ws)
	require.NoError(t, err)
	require.True(t, view.HasGitCredential)
	b, _ := json.Marshal(view)
	testutil.AssertNoLeak(t, string(b), pw)
}

func TestU4_GitStore_SaveNeedsAConnection(t *testing.T) {
	secrets, state := newFakeSecrets(), newFakeState()
	store := newTestStore(secrets, state)
	require.ErrorIs(t, store.SaveGit(context.Background(), ws, "lan", testutil.Token(t)), ErrNotConnected)
	require.Empty(t, secrets.snapshot())
}

func TestU4_GitStore_LoadIgnoresAnotherHost(t *testing.T) {
	store, secrets, _, ctx := connectedStore(t, "a.backlog.com")
	secrets.data[gitKey] = `{"username":"lan","password":"TESTSECRET-x","spaceHost":"b.backlog.com","revision":1}`
	view, err := store.Load(ctx, ws)
	require.NoError(t, err)
	require.False(t, view.HasGitCredential)
	secrets.data[gitKey] = `not json`
	view, err = store.Load(ctx, ws)
	require.NoError(t, err)
	require.False(t, view.HasGitCredential)
}

func TestU4_GitStore_DisconnectDeletesTheGitSecretFirst(t *testing.T) {
	store, secrets, state, ctx := connectedStore(t, "a.backlog.com")
	require.NoError(t, store.SaveGit(ctx, ws, "lan", testutil.Token(t)))
	secrets.ops = nil
	_, changed, err := store.Disconnect(ctx, ws)
	require.NoError(t, err)
	require.True(t, changed)
	require.Equal(t, []string{"del:backlog.connection." + ws, "del:" + gitKey}, secrets.ops)
	require.Empty(t, secrets.snapshot())
	require.Equal(t, true, recordOf(t, state, ws)["disconnected"])
}

func TestU4_GitStore_FailedGitDeleteWritesNoRecord(t *testing.T) {
	store, secrets, state, ctx := connectedStore(t, "a.backlog.com")
	require.NoError(t, store.SaveGit(ctx, ws, "lan", testutil.Token(t)))
	secrets.failDelKey = gitKey
	_, _, err := store.Disconnect(ctx, ws)
	require.ErrorIs(t, err, ErrStore)
	require.NotContains(t, recordOf(t, state, ws), "disconnected", "no disconnect record")
}

func TestU4_GitStore_SpaceChangeDeletesItFirst(t *testing.T) {
	store, secrets, _, ctx := connectedStore(t, "a.backlog.com")
	require.NoError(t, store.SaveGit(ctx, ws, "lan", testutil.Token(t)))
	secrets.ops = nil
	_, err := store.Save(ctx, ws, "b.backlog.jp", testutil.APIKey(t), testUser)
	require.NoError(t, err)
	require.Equal(t, []string{"del:" + gitKey, "set:backlog.connection." + ws}, secrets.ops)
	_, ok := gitSecretOf(t, secrets)
	require.False(t, ok)
}

func TestU4_GitStore_SameHostChangesKeepIt(t *testing.T) {
	store, secrets, _, ctx := connectedStore(t, "a.backlog.com")
	require.NoError(t, store.SaveGit(ctx, ws, "lan", testutil.Token(t)))
	_, err := store.Save(ctx, ws, "a.backlog.com", testutil.APIKey(t), testUser)
	require.NoError(t, err)
	_, err = store.SaveProjects(ctx, ws, []string{"PROJ"})
	require.NoError(t, err)
	tokens := oauthTokens(t, fixedNow.Add(time.Hour))
	view, err := store.SaveOAuth(ctx, ws, "a.backlog.com", tokens, testUser)
	require.NoError(t, err)
	require.NoError(t, store.UpdateTokens(ctx, ws, view.ConnectionEpoch, oauthTokens(t, fixedNow.Add(2*time.Hour))))
	_, ok := gitSecretOf(t, secrets)
	require.True(t, ok, "same-host replacement, project saves and token refresh keep it")
	view, err = store.Load(ctx, ws)
	require.NoError(t, err)
	require.True(t, view.HasGitCredential)
}

func TestOAuthPendingRefusesAnEmptyVerifier(t *testing.T) {
	secrets, state := newFakeSecrets(), newFakeState()
	store := newTestStore(secrets, state)
	nonce, _ := newNonce()
	require.NoError(t, store.SavePending(context.Background(), ws, nonce, hashHex(""), "a.backlog.com", fixedNow.Add(10*time.Minute)))

	_, found, err := store.TakePending(context.Background(), ws, nonce, "")
	require.NoError(t, err)
	require.False(t, found, "an empty verifier never matches, even against sha256(\"\")")
	require.Contains(t, state.snapshot(), stateKey("workspace", ws, "oauth_pending"), "and does not use up the sign-in")
}
