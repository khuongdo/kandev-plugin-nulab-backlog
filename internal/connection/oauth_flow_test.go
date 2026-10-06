package connection

import (
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/backlog"
	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/testutil"
)

func (u *u2) start(t *testing.T) (state string) {
	t.Helper()
	res, err := u.svc.StartOAuth(u.ctx, ws, StartInput{SpaceURL: "Example-Space.backlog.com"})
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
	res, err := u.svc.StartOAuth(u.ctx, ws, StartInput{SpaceURL: " https://Example-Space.backlog.com/ "})
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
		_, err := u.svc.StartOAuth(u.ctx, ws, StartInput{SpaceURL: "a.backlog.com"})
		require.ErrorIs(t, err, ErrOAuthNotConfigured)
		require.Equal(t, Outcome{Code: CodeValidation, Field: FieldOAuth}, Classify(err))
		u.svc.Config = nil
		_, err = u.svc.StartOAuth(u.ctx, ws, StartInput{SpaceURL: "a.backlog.com"})
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

	res := u.svc.CompleteOAuth(u.ctx, url.Values{"state": {state}, "code": {code}})
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
	res := u.svc.CompleteOAuth(u.ctx, url.Values{"state": {state}, "error": {"access_denied"}})
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
			res := u.svc.CompleteOAuth(u.ctx, query(state))
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
		require.Equal(t, OutcomeConnected, u.svc.CompleteOAuth(u.ctx, url.Values{"state": {state}, "code": {"c-1234"}}).Outcome)
		require.Equal(t, OutcomeFailed, u.svc.CompleteOAuth(u.ctx, url.Values{"state": {state}, "code": {"c-1234"}}).Outcome)
		_, _, exchange, _ := u.gw.counts()
		require.Equal(t, 1, exchange)
	})
	t.Run("an expired state", func(t *testing.T) {
		u := newU2(t)
		state := u.start(t)
		u.okTokens(t)
		u.advance(10*time.Minute + time.Second)
		require.Equal(t, OutcomeFailed, u.svc.CompleteOAuth(u.ctx, url.Values{"state": {state}, "code": {"c-1234"}}).Outcome)
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
			res := u.svc.CompleteOAuth(u.ctx, url.Values{"state": {state}, "code": {"c-1234"}})
			require.Equal(t, OutcomeFailed, res.Outcome)
			require.Empty(t, u.secrets.snapshot())
			_, hasRecord := u.state.snapshot()[stateKey("workspace", ws, "connection")]
			require.False(t, hasRecord)
			require.Empty(t, u.changedLogs(t))
		})
	}
}
