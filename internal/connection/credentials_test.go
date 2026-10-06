package connection

import (
	"encoding/json"
	"sync"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/backlog"
	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/testutil"
)

func TestConnectionReaderCurrentReturnsTheSnapshot(t *testing.T) {
	u := newU2(t)
	_, err := u.svc.Current(u.ctx, ws)
	require.ErrorIs(t, err, ErrNotConnected)

	u.connectKey(t, "a.backlog.com")
	_, err = u.svc.store.SaveProjects(u.ctx, ws, []string{"PROJ"})
	require.NoError(t, err)
	snap, err := u.svc.Current(u.ctx, ws)
	require.NoError(t, err)
	require.Equal(t, Snapshot{SpaceHost: "a.backlog.com", AuthMethod: "api_key", ConnectionEpoch: 2, SelectedProjects: []string{"PROJ"}}, snap)
}

func TestConnectionReaderCredentialsReturnsTheAPIKey(t *testing.T) {
	u := newU2(t)
	key := u.connectKey(t, "a.backlog.com")
	creds, epoch, err := u.svc.Credentials(u.ctx, ws)
	require.NoError(t, err)
	require.Equal(t, backlog.Credentials{SpaceHost: "a.backlog.com", APIKey: key}, creds)
	require.Equal(t, 1, epoch)
}

func TestConnectionReaderMismatchedPairIsNotConnected(t *testing.T) {
	u := newU2(t)
	u.connectKey(t, "a.backlog.com")
	raw := u.secrets.snapshot()["backlog.connection."+ws]
	var sec map[string]any
	require.NoError(t, json.Unmarshal([]byte(raw), &sec))
	sec["connectionEpoch"] = 9
	b, _ := json.Marshal(sec)
	u.secrets.data["backlog.connection."+ws] = string(b)

	_, _, err := u.svc.Credentials(u.ctx, ws)
	require.ErrorIs(t, err, ErrNotConnected, "BR2.11 before any secret is handed out")
	_, err = u.svc.Current(u.ctx, ws)
	require.ErrorIs(t, err, ErrNotConnected)
	require.Equal(t, CodeReconnectRequired, Classify(err).Code)
}

func TestTokenFreshTokenIsReturnedWithoutRefresh(t *testing.T) {
	u := newU2(t)
	tokens := u.connectOAuth(t, 10*time.Minute)
	creds, _, err := u.svc.Credentials(u.ctx, ws)
	require.NoError(t, err)
	require.Equal(t, backlog.Credentials{SpaceHost: "a.backlog.com", AccessToken: tokens.AccessToken}, creds)
	_, _, _, refresh := u.gw.counts()
	require.Zero(t, refresh)
}

func TestTokenNearExpiryIsRefreshedBeforeItIsReturned(t *testing.T) {
	u := newU2(t)
	u.connectOAuth(t, 4*time.Minute) // under the 5-minute threshold (AC1.4.1)
	u.gw.refreshed = backlog.TokenSet{AccessToken: testutil.Token(t), RefreshToken: testutil.Token(t), ExpiresAt: u.clock().Add(time.Hour)}

	creds, epoch, err := u.svc.Credentials(u.ctx, ws)
	require.NoError(t, err)
	require.Equal(t, u.gw.refreshed.AccessToken, creds.AccessToken)
	require.Equal(t, 1, epoch, "a refresh never changes the epoch")
	sec := secretOf(t, u.secrets, ws)
	require.Equal(t, u.gw.refreshed.RefreshToken, sec["refreshToken"], "the rotated refresh token replaces the old one")
	require.EqualValues(t, 1, sec["connectionEpoch"])
	require.Len(t, eventsNamed(u.events(t), "token_refreshed"), 1)
}

func TestTokenFiveConcurrentCallersCauseExactlyOneRefresh(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		u := newU2(t)
		u.connectOAuth(t, -time.Minute) // already expired
		u.gw.refreshed = backlog.TokenSet{AccessToken: testutil.Token(t), RefreshToken: testutil.Token(t), ExpiresAt: u.clock().Add(time.Hour)}
		u.gw.refreshDelay = 200 * time.Millisecond

		var wg sync.WaitGroup
		results := make([]string, 5)
		for i := range 5 {
			wg.Go(func() {
				creds, _, err := u.svc.Credentials(u.ctx, ws)
				require.NoError(t, err)
				results[i] = creds.AccessToken
			})
		}
		wg.Wait()
		_, _, _, refresh := u.gw.counts()
		require.Equal(t, 1, refresh, "AC1.4.2")
		for _, r := range results {
			require.Equal(t, u.gw.refreshed.AccessToken, r)
		}
		require.Equal(t, u.gw.refreshed.RefreshToken, secretOf(t, u.secrets, ws)["refreshToken"])
	})
}

func TestTokenRefusedRefreshAsksToSignInAgain(t *testing.T) {
	for name, refusal := range map[string]error{
		"400 invalid_grant": &backlog.Error{Kind: backlog.KindInvalid, Status: 400},
		"401":               &backlog.Error{Kind: backlog.KindUnauthorized, Status: 401},
	} {
		t.Run(name, func(t *testing.T) {
			u := newU2(t)
			u.connectOAuth(t, -time.Minute)
			u.gw.refreshErr = refusal

			_, _, err := u.svc.Credentials(u.ctx, ws)
			require.ErrorIs(t, err, ErrReconnectRequired)
			require.Equal(t, CodeReconnectRequired, Classify(err).Code)
			view, err := u.svc.Get(u.ctx, ws)
			require.NoError(t, err)
			require.Equal(t, StateSignInAgain, view.State)

			for range 3 { // later cycles make no Backlog request until a new sign-in
				_, _, err = u.svc.Credentials(u.ctx, ws)
				require.ErrorIs(t, err, ErrReconnectRequired)
			}
			myself, projects, exchange, refresh := u.gw.counts()
			require.Equal(t, 1, refresh, "exactly one attempt (AC1.4.3)")
			require.Zero(t, myself+projects+exchange)
			require.Len(t, eventsNamed(u.events(t), "token_refresh_failed"), 1)
		})
	}
}

func TestTokenUnreachableRefreshKeepsTheState(t *testing.T) {
	u := newU2(t)
	u.connectOAuth(t, -time.Minute)
	u.gw.refreshErr = &backlog.Error{Kind: backlog.KindUnreachable, Class: "timeout"}
	_, _, err := u.svc.Credentials(u.ctx, ws)
	require.Equal(t, CodeUnreachable, Classify(err).Code)
	view, err := u.svc.Get(u.ctx, ws)
	require.NoError(t, err)
	require.Equal(t, StateConnected, view.State, "a temporary failure never asks to sign in again")
}

func TestTokenRefreshFinishingAfterAnEpochChangeWritesNothing(t *testing.T) {
	u := newU2(t)
	u.connectOAuth(t, -time.Minute)
	u.gw.refreshed = backlog.TokenSet{AccessToken: testutil.Token(t), RefreshToken: testutil.Token(t), ExpiresAt: u.clock().Add(time.Hour)}
	var newer map[string]string
	u.gw.onRefresh = func() {
		// Another writer replaced the connection while the refresh ran.
		_, err := u.svc.store.Save(u.ctx, ws, "a.backlog.com", testutil.APIKey(t), testUser)
		require.NoError(t, err)
		newer = u.secrets.snapshot()
	}
	_, _, err := u.svc.Credentials(u.ctx, ws)
	require.ErrorIs(t, err, ErrReconnectRequired)
	require.Equal(t, newer, u.secrets.snapshot(), "the late refresh wrote nothing")
}

func TestTokenRefreshWithoutOAuthConfigIsNotConfigured(t *testing.T) {
	u := newU2(t)
	u.connectOAuth(t, -time.Minute)
	u.cfg.m = nil
	_, _, err := u.svc.Credentials(u.ctx, ws)
	require.ErrorIs(t, err, ErrOAuthNotConfigured)
	_, _, _, refresh := u.gw.counts()
	require.Zero(t, refresh)
}
