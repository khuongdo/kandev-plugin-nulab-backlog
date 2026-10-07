package connection

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/backlog"
	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/testutil"
)

// hashHex is what the settings UI sends as verifierHash: hex SHA-256 of the cookie value.
func hashHex(v string) string {
	sum := sha256.Sum256([]byte(v))
	return hex.EncodeToString(sum[:])
}

// start begins a sign-in as the browser would, keeping its verifier cookie in u.verifier.
func (u *u2) start(t *testing.T) (state string) {
	t.Helper()
	u.verifier = testutil.Token(t)
	res, err := u.svc.StartOAuth(u.ctx, ws, StartInput{SpaceURL: "Example-Space.backlog.com", VerifierHash: hashHex(u.verifier)})
	require.NoError(t, err)
	au, err := url.Parse(res.AuthorizeURL)
	require.NoError(t, err)
	return au.Query().Get("state")
}

func (u *u2) okTokens(t *testing.T) backlog.TokenSet {
	t.Helper()
	u.gw.tokens = backlog.TokenSet{AccessToken: testutil.Token(t), RefreshToken: testutil.Token(t), ExpiresAt: u.clock().Add(time.Hour)}
	return u.gw.tokens
}

func TestOAuthStartBuildsTheAuthorizeURL(t *testing.T) {
	u := newU2(t)
	res, err := u.svc.StartOAuth(u.ctx, ws, StartInput{SpaceURL: " https://Example-Space.backlog.com/ ", VerifierHash: hashHex(testutil.Token(t))})
	require.NoError(t, err)
	au, err := url.Parse(res.AuthorizeURL)
	require.NoError(t, err)
	require.Equal(t, "https", au.Scheme)
	require.Equal(t, "example-space.backlog.com", au.Host)
	require.Equal(t, "/OAuth2AccessRequest.action", au.Path)
	q := au.Query()
	require.Equal(t, "code", q.Get("response_type"))
	require.Equal(t, "client-id-1", q.Get("client_id"))
	require.Equal(t, testBaseURL+"/api/plugins/nulab-backlog/webhooks/oauth-callback", q.Get("redirect_uri"))
	gotWS, _, err := decodeState(q.Get("state"))
	require.NoError(t, err)
	require.Equal(t, ws, gotWS)
	require.Len(t, q, 4)
	pending := u.state.snapshot()[stateKey("workspace", ws, "oauth_pending")]
	require.Equal(t, fixedNow.Add(10*time.Minute).Format(time.RFC3339), pending["expiresAt"], "expires after 10 minutes")
	require.Equal(t, "example-space.backlog.com", pending["spaceHost"])
	require.Equal(t, 1, u.cfg.calls, "config is read on every start")
}

func TestOAuthStartRefusals(t *testing.T) {
	t.Run("an invalid address", func(t *testing.T) {
		u := newU2(t)
		_, err := u.svc.StartOAuth(u.ctx, ws, StartInput{SpaceURL: "http://evil.example.com"})
		out := Classify(err)
		require.Equal(t, CodeValidation, out.Code)
		require.Equal(t, FieldSpaceURL, out.Field)
		require.Empty(t, u.state.snapshot())
	})
	t.Run("no OAuth config", func(t *testing.T) {
		u := newU2(t)
		u.cfg.m = map[string]any{}
		_, err := u.svc.StartOAuth(u.ctx, ws, StartInput{SpaceURL: "a.backlog.com", VerifierHash: hashHex(testutil.Token(t))})
		require.ErrorIs(t, err, ErrOAuthNotConfigured)
		require.Equal(t, Outcome{Code: CodeValidation, Field: FieldOAuth}, Classify(err))
		u.svc.Config = nil
		_, err = u.svc.StartOAuth(u.ctx, ws, StartInput{SpaceURL: "a.backlog.com", VerifierHash: hashHex(testutil.Token(t))})
		require.ErrorIs(t, err, ErrOAuthNotConfigured)
	})
	t.Run("Backlog is off", func(t *testing.T) {
		u := newU2(t)
		require.NoError(t, u.svc.store.SaveSwitch(u.ctx, ws, false))
		_, err := u.svc.StartOAuth(u.ctx, ws, StartInput{SpaceURL: "a.backlog.com"})
		require.Equal(t, CodeIntegrationDisabled, Classify(err).Code)
	})
}

func TestOAuthCompleteConnects(t *testing.T) {
	u := newU2(t)
	events := subscribe(t, u.svc)
	state := u.start(t)
	tokens := u.okTokens(t)
	code := testutil.Token(t)

	res := u.svc.CompleteOAuth(u.ctx, url.Values{"state": {state}, "code": {code}}, u.verifier)
	require.Equal(t, OAuthResult{WorkspaceID: ws, Outcome: OutcomeConnected}, res)
	require.Equal(t, code, u.gw.exchangeCode)
	require.Equal(t, testBaseURL+"/api/plugins/nulab-backlog/webhooks/oauth-callback", u.gw.exchangeURI)
	require.Equal(t, "client-id-1", u.gw.exchangeID)
	require.Equal(t, backlog.Credentials{SpaceHost: "example-space.backlog.com", AccessToken: tokens.AccessToken}, u.gw.creds, "Myself with the Bearer token")
	view, err := u.svc.Get(u.ctx, ws)
	require.NoError(t, err)
	require.Equal(t, StateConnected, view.State)
	require.Equal(t, "oauth", view.AuthMethod)
	e := events.next(t)
	require.Equal(t, ReasonConnected, e.Reason)
	require.Equal(t, 1, e.ConnectionEpoch)
}

func TestOAuthCompleteCancelledStoresNothing(t *testing.T) {
	u := newU2(t)
	state := u.start(t)
	res := u.svc.CompleteOAuth(u.ctx, url.Values{"state": {state}, "error": {"access_denied"}}, u.verifier)
	require.Equal(t, OutcomeCancelled, res.Outcome)
	require.Equal(t, ws, res.WorkspaceID)
	require.Empty(t, u.state.snapshot(), "the pending state is deleted")
	require.Empty(t, u.secrets.snapshot())
	_, _, exchange, _ := u.gw.counts()
	require.Zero(t, exchange)
}

func TestOAuthCompleteBadStateOrCodeMakesNoTokenRequest(t *testing.T) {
	other, _ := newNonce()
	cases := map[string]func(state string) url.Values{
		"missing state":   func(string) url.Values { return url.Values{"code": {"c-1234"}} },
		"malformed state": func(string) url.Values { return url.Values{"state": {"%%%"}, "code": {"c-1234"}} },
		"unknown state":   func(string) url.Values { return url.Values{"state": {encodeState(ws, other)}, "code": {"c-1234"}} },
		"missing code":    func(s string) url.Values { return url.Values{"state": {s}} },
	}
	for name, query := range cases {
		t.Run(name, func(t *testing.T) {
			u := newU2(t)
			state := u.start(t)
			u.okTokens(t)
			res := u.svc.CompleteOAuth(u.ctx, query(state), u.verifier)
			require.Equal(t, OutcomeFailed, res.Outcome)
			_, _, exchange, _ := u.gw.counts()
			require.Zero(t, exchange, "AC1.3.3")
			require.Empty(t, u.secrets.snapshot())
		})
	}
	t.Run("a used state", func(t *testing.T) {
		u := newU2(t)
		state := u.start(t)
		u.okTokens(t)
		require.Equal(t, OutcomeConnected, u.svc.CompleteOAuth(u.ctx, url.Values{"state": {state}, "code": {"c-1234"}}, u.verifier).Outcome)
		require.Equal(t, OutcomeFailed, u.svc.CompleteOAuth(u.ctx, url.Values{"state": {state}, "code": {"c-1234"}}, u.verifier).Outcome)
		_, _, exchange, _ := u.gw.counts()
		require.Equal(t, 1, exchange)
	})
	t.Run("an expired state", func(t *testing.T) {
		u := newU2(t)
		state := u.start(t)
		u.okTokens(t)
		u.advance(10*time.Minute + time.Second)
		require.Equal(t, OutcomeFailed, u.svc.CompleteOAuth(u.ctx, url.Values{"state": {state}, "code": {"c-1234"}}, u.verifier).Outcome)
		_, _, exchange, _ := u.gw.counts()
		require.Zero(t, exchange)
	})
}

func TestOAuthCompleteFailuresStoreNothing(t *testing.T) {
	cases := map[string]func(u *u2){
		"the exchange fails":    func(u *u2) { u.gw.exchangeErr = &backlog.Error{Kind: backlog.KindInvalid, Status: 400} },
		"Myself fails":          func(u *u2) { u.gw.err = &backlog.Error{Kind: backlog.KindUnauthorized, Status: 401} },
		"the config is removed": func(u *u2) { u.cfg.m = nil },
		"Backlog is turned off during the exchange": func(u *u2) {
			u.gw.onExchange = func() { _ = u.svc.store.SaveSwitch(u.ctx, ws, false) }
		},
		"the record write fails": func(u *u2) { u.gw.onExchange = func() { u.state.failSet = true } },
	}
	for name, setup := range cases {
		t.Run(name, func(t *testing.T) {
			u := newU2(t)
			state := u.start(t)
			u.okTokens(t)
			setup(u)
			res := u.svc.CompleteOAuth(u.ctx, url.Values{"state": {state}, "code": {"c-1234"}}, u.verifier)
			require.Equal(t, OutcomeFailed, res.Outcome)
			require.Empty(t, u.secrets.snapshot())
			_, hasRecord := u.state.snapshot()[stateKey("workspace", ws, "connection")]
			require.False(t, hasRecord)
			require.Empty(t, u.changedLogs(t))
		})
	}
}

func TestU2_StartNeedsAVerifierHash(t *testing.T) {
	for name, hash := range map[string]string{
		"missing":        "",
		"too short":      strings.Repeat("a", 63),
		"too long":       strings.Repeat("a", 65),
		"not hex":        strings.Repeat("g", 64),
		"upper-case hex": strings.Repeat("A", 64),
	} {
		t.Run(name, func(t *testing.T) {
			u := newU2(t)
			_, err := u.svc.StartOAuth(u.ctx, ws, StartInput{SpaceURL: "a.backlog.com", VerifierHash: hash})
			require.Equal(t, Outcome{Code: CodeValidation, Field: FieldVerifierHash}, Classify(err))
			require.Empty(t, u.state.snapshot(), "no pending sign-in")
		})
	}
	t.Run("a valid hash is stored with the pending sign-in", func(t *testing.T) {
		u := newU2(t)
		_ = u.start(t)
		pending := u.state.snapshot()[stateKey("workspace", ws, "oauth_pending")]
		require.Equal(t, hashHex(u.verifier), pending["verifierHash"])
		testutil.AssertNoLeak(t, u.logs.String(), u.verifier)
	})
}

func TestU2_CallbackChecksTheVerifierCookie(t *testing.T) {
	cases := map[string]func(u *u2) string{
		"no cookie":              func(*u2) string { return "" },
		"wrong cookie":           func(*u2) string { return testutil.Token(t) },
		"the hash as the cookie": func(u *u2) string { return hashHex(u.verifier) },
	}
	for name, cookie := range cases {
		t.Run(name, func(t *testing.T) {
			u := newU2(t)
			state := u.start(t)
			u.okTokens(t)
			q := url.Values{"state": {state}, "code": {"c-1234"}}
			bad := cookie(u)

			res := u.svc.CompleteOAuth(u.ctx, q, bad)
			require.Equal(t, OAuthResult{WorkspaceID: ws, Outcome: OutcomeFailed}, res)
			_, _, exchange, _ := u.gw.counts()
			require.Zero(t, exchange, "no token request before the verifier matches")
			require.Empty(t, u.secrets.snapshot())
			require.Contains(t, u.state.snapshot(), stateKey("workspace", ws, "oauth_pending"),
				"the pending sign-in is kept, so a stranger cannot cancel it")
			testutil.AssertNoLeak(t, u.logs.String(), u.verifier, bad)

			res = u.svc.CompleteOAuth(u.ctx, q, u.verifier)
			require.Equal(t, OutcomeConnected, res.Outcome, "the starting browser still finishes")
			testutil.AssertNoLeak(t, u.logs.String(), u.verifier)
		})
	}
}

func TestU2_CallbackCancelNeedsTheVerifier(t *testing.T) {
	u := newU2(t)
	state := u.start(t)
	res := u.svc.CompleteOAuth(u.ctx, url.Values{"state": {state}, "error": {"access_denied"}}, "")
	require.Equal(t, OutcomeFailed, res.Outcome)
	require.Contains(t, u.state.snapshot(), stateKey("workspace", ws, "oauth_pending"))
}

// plantEmptyHashPending stores a pending sign-in bound to sha256(""), as if the
// start-time check were bypassed, and returns its state.
func (u *u2) plantEmptyHashPending(t *testing.T) string {
	t.Helper()
	nonce, err := newNonce()
	require.NoError(t, err)
	require.NoError(t, u.svc.store.SavePending(u.ctx, ws, nonce, emptyVerifierHash, "example-space.backlog.com", u.clock().Add(pendingTTL)))
	return encodeState(ws, nonce)
}

func TestU2_EmptyVerifierNeverMatches(t *testing.T) {
	require.Equal(t, emptyVerifierHash, hashHex(""))

	t.Run("start refuses the hash of the empty verifier", func(t *testing.T) {
		u := newU2(t)
		_, err := u.svc.StartOAuth(u.ctx, ws, StartInput{SpaceURL: "a.backlog.com", VerifierHash: emptyVerifierHash})
		require.Equal(t, Outcome{Code: CodeValidation, Field: FieldVerifierHash}, Classify(err))
		require.Empty(t, u.state.snapshot(), "no pending sign-in")
	})

	t.Run("a callback without a cookie or with an empty one fails and keeps the record", func(t *testing.T) {
		u := newU2(t)
		u.okTokens(t)
		state := u.plantEmptyHashPending(t)
		res := u.svc.CompleteOAuth(u.ctx, url.Values{"state": {state}, "code": {"c-1234"}}, "")
		require.Equal(t, OAuthResult{WorkspaceID: ws, Outcome: OutcomeFailed}, res)
		_, _, exchange, _ := u.gw.counts()
		require.Zero(t, exchange, "no token request")
		require.Empty(t, u.secrets.snapshot())
		require.Contains(t, u.state.snapshot(), stateKey("workspace", ws, "oauth_pending"), "the record is kept")
		require.Contains(t, u.logs.String(), `"reason":"bad_state"`)
	})

	t.Run("the callback refuses an empty verifier before taking the lock or comparing", func(t *testing.T) {
		u := newU2(t)
		state := u.plantEmptyHashPending(t)
		unlock, err := u.svc.lockWS(u.ctx, ws) // a busy workspace would make a later check time out
		require.NoError(t, err)
		defer unlock()
		ctx, cancel := context.WithTimeout(u.ctx, 50*time.Millisecond)
		defer cancel()
		res := u.svc.CompleteOAuth(ctx, url.Values{"state": {state}, "code": {"c-1234"}}, "")
		require.Equal(t, OAuthResult{WorkspaceID: ws, Outcome: OutcomeFailed}, res)
		require.Contains(t, u.logs.String(), `"reason":"bad_state"`, "refused up front, not after waiting for the lock")
	})

	t.Run("end to end: attacker starts with the empty hash, victim has no cookie", func(t *testing.T) {
		u := newU2(t)
		u.okTokens(t)
		_, err := u.svc.StartOAuth(u.ctx, ws, StartInput{SpaceURL: "Example-Space.backlog.com", VerifierHash: emptyVerifierHash})
		require.Error(t, err, "the attacker's start is refused")
		state := u.plantEmptyHashPending(t) // even if a record with that hash exists
		res := u.svc.CompleteOAuth(u.ctx, url.Values{"state": {state}, "code": {"c-1234"}}, "")
		require.Equal(t, OutcomeFailed, res.Outcome)
		_, _, exchange, _ := u.gw.counts()
		require.Zero(t, exchange)
		require.Empty(t, u.secrets.snapshot())
		require.Contains(t, u.state.snapshot(), stateKey("workspace", ws, "oauth_pending"))
	})
}
