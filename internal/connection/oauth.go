package connection

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/backlog"
	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/redact"
)

// ErrOAuthNotConfigured is returned when the operator has not set the OAuth
// app in Settings > Plugins. Its text never holds a config value.
var ErrOAuthNotConfigured = errors.New("OAuth is not configured on this Kandev server")

// callbackPath is the webhook route Kandev relays to HandleWebhook.
const callbackPath = "/api/plugins/nulab-backlog/webhooks/oauth-callback"

// OAuthConfig is the operator config read with Host.GetConfig (R-02).
type OAuthConfig struct {
	ClientID      string
	ClientSecret  string
	PublicBaseURL string // no trailing slash
}

// String hides the client secret.
func (c OAuthConfig) String() string {
	return fmt.Sprintf("connection.OAuthConfig{ClientID: %s, PublicBaseURL: %s}", c.ClientID, c.PublicBaseURL)
}

// GoString hides the client secret for %#v as well.
func (c OAuthConfig) GoString() string { return c.String() }

// RedirectURI is the callback address registered with the Nulab OAuth app.
func (c OAuthConfig) RedirectURI() string { return c.PublicBaseURL + callbackPath }

// Client returns the gateway's view of the OAuth app.
func (c OAuthConfig) Client() backlog.OAuthClient {
	return backlog.OAuthClient{ClientID: c.ClientID, ClientSecret: c.ClientSecret}
}

// ParseOAuthConfig reads oauth_client_id, oauth_client_secret and
// public_base_url. Any missing, blank or invalid value is ErrOAuthNotConfigured.
func ParseOAuthConfig(m map[string]any) (OAuthConfig, error) {
	get := func(k string) string { s, _ := m[k].(string); return strings.TrimSpace(s) }
	cfg := OAuthConfig{ClientID: get("oauth_client_id"), ClientSecret: get("oauth_client_secret")}
	base := get("public_base_url")
	if cfg.ClientID == "" || cfg.ClientSecret == "" || base == "" {
		return OAuthConfig{}, fmt.Errorf("missing OAuth config field: %w", ErrOAuthNotConfigured)
	}
	u, err := url.Parse(base)
	if err != nil || u.Host == "" || u.User != nil || u.RawQuery != "" || u.ForceQuery || u.Fragment != "" ||
		strings.Contains(base, "#") || !allowedScheme(u) {
		return OAuthConfig{}, fmt.Errorf("public_base_url must be an absolute https URL: %w", ErrOAuthNotConfigured)
	}
	cfg.PublicBaseURL = strings.TrimSuffix(base, "/")
	return cfg, nil
}

// allowedScheme is https, or http only for a local trial run (assumption A2).
func allowedScheme(u *url.URL) bool {
	switch u.Scheme {
	case "https":
		return true
	case "http":
		h := u.Hostname()
		return h == "localhost" || h == "127.0.0.1"
	default:
		return false
	}
}

// errBadState is returned for any OAuth state that cannot be decoded.
var errBadState = errors.New("malformed OAuth state")

const (
	nonceLen       = 32
	maxWorkspaceID = 128
)

// workspaceIDPattern keeps a decoded workspace id safe to put in a path.
var workspaceIDPattern = regexp.MustCompile(`^[A-Za-z0-9_-][A-Za-z0-9._-]*$`)

// newNonce returns 32 random bytes (256 bits).
func newNonce() ([]byte, error) {
	b := make([]byte, nonceLen)
	if _, err := rand.Read(b); err != nil {
		return nil, fmt.Errorf("generate OAuth nonce: %w", err)
	}
	return b, nil
}

// encodeState is base64url(nonce || workspaceID), unpadded.
func encodeState(workspaceID string, nonce []byte) string {
	return base64.RawURLEncoding.EncodeToString(append(append([]byte(nil), nonce...), workspaceID...))
}

// decodeState returns the workspace and nonce of a state, or errBadState.
func decodeState(s string) (string, []byte, error) {
	if len(s) > base64.RawURLEncoding.EncodedLen(nonceLen+maxWorkspaceID) {
		return "", nil, errBadState
	}
	raw, err := base64.RawURLEncoding.Strict().DecodeString(s)
	if err != nil || len(raw) <= nonceLen {
		return "", nil, errBadState
	}
	ws := string(raw[nonceLen:])
	if !workspaceIDPattern.MatchString(ws) {
		return "", nil, errBadState
	}
	return ws, raw[:nonceLen], nil
}

// pendingTTL is how long a started sign-in may take.
const pendingTTL = 10 * time.Minute

// OAuth callback outcomes, passed to the settings page as ?oauth=.
const (
	OutcomeConnected = "connected"
	OutcomeCancelled = "cancelled"
	OutcomeFailed    = "failed"
)

// FieldVerifierHash is the start_oauth field holding the browser verifier's hash.
const FieldVerifierHash = "verifierHash"

// VerifierCookie is the cookie the settings UI sets before leaving for
// Backlog. Kandev drops Set-Cookie on webhook replies but forwards other
// cookies to a public webhook, so the callback can check it (R-01).
const VerifierCookie = "nulab_backlog_oauth_verifier"

// verifierHashPattern is a lower-case hex SHA-256.
var verifierHashPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)

// emptyVerifierHash is sha256(""). A start bound to it would match any
// browser that sends no cookie, so it is refused (review R-01).
const emptyVerifierHash = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"

// StartInput is the connection.start_oauth body. VerifierHash is the hex
// SHA-256 of the verifier cookie; the verifier itself never reaches the server
// until the callback.
type StartInput struct {
	SpaceURL     string `json:"spaceUrl"`
	VerifierHash string `json:"verifierHash"`
}

// StartResult is the connection.start_oauth reply.
type StartResult struct {
	AuthorizeURL string `json:"authorizeUrl"`
}

// OAuthResult is what the callback redirect needs. WorkspaceID is empty when
// the state could not be decoded.
type OAuthResult struct {
	WorkspaceID string
	Outcome     string
	Restored    bool
}

// StartOAuth stores a single-use pending sign-in and returns Backlog's
// authorization page address (AC1.3.1).
func (s *Service) StartOAuth(ctx context.Context, workspaceID string, in StartInput) (StartResult, error) {
	if err := s.RequireEnabled(ctx, workspaceID); err != nil {
		return StartResult{}, err
	}
	addr, err := ParseSpaceAddress(in.SpaceURL)
	if err != nil {
		return StartResult{}, &FieldError{Field: FieldSpaceURL, Err: err}
	}
	if !verifierHashPattern.MatchString(in.VerifierHash) {
		return StartResult{}, &FieldError{Field: FieldVerifierHash, Err: errors.New("must be a lower-case hex SHA-256")}
	}
	if in.VerifierHash == emptyVerifierHash {
		return StartResult{}, &FieldError{Field: FieldVerifierHash, Err: errors.New("must be the hash of a non-empty verifier")}
	}
	cfg, err := s.oauthConfig(ctx)
	if err != nil {
		return StartResult{}, err
	}
	nonce, err := newNonce()
	if err != nil {
		return StartResult{}, err
	}
	if err := s.store.SavePending(ctx, workspaceID, nonce, in.VerifierHash, addr.Host, s.store.Now().Add(pendingTTL)); err != nil {
		return StartResult{}, err
	}
	u := url.URL{Scheme: "https", Host: addr.Host, Path: "/OAuth2AccessRequest.action", RawQuery: url.Values{
		"response_type": {"code"}, "client_id": {cfg.ClientID},
		"redirect_uri": {cfg.RedirectURI()}, "state": {encodeState(workspaceID, nonce)},
	}.Encode()}
	redact.Logger(ctx).InfoContext(ctx, "oauth started", "event", "oauth_started", "spaceHost", addr.Host)
	return StartResult{AuthorizeURL: u.String()}, nil
}

// CompleteOAuth handles the callback query (AC1.3.2, AC1.3.3). verifier is
// the VerifierCookie value relayed with the callback ("" when absent).
// Nothing is stored and no token request is made unless the state and the
// verifier match a pending sign-in and a code is present.
func (s *Service) CompleteOAuth(ctx context.Context, q url.Values, verifier string) OAuthResult {
	start := time.Now()
	ctx = redact.WithSecrets(ctx, q.Get("state"), q.Get("code"), verifier)
	res, reason := OAuthResult{Outcome: OutcomeFailed}, "bad_state"
	if workspaceID, nonce, err := decodeState(q.Get("state")); err == nil {
		res, reason = s.completeOAuth(ctx, workspaceID, nonce, verifier, q)
	}
	log := redact.Logger(ctx)
	if reason != "" {
		log.WarnContext(ctx, "oauth failed", "event", "oauth_failed", "reason", reason,
			"durationMs", time.Since(start).Milliseconds())
	}
	return res
}

func (s *Service) completeOAuth(ctx context.Context, workspaceID string, nonce []byte, verifier string, q url.Values) (OAuthResult, string) {
	failed := OAuthResult{WorkspaceID: workspaceID, Outcome: OutcomeFailed}
	if verifier == "" { // no cookie: never compared, the record is kept (R-01)
		return failed, "bad_state"
	}
	unlock, err := s.lockWS(ctx, workspaceID)
	if err != nil {
		return failed, "store"
	}
	defer unlock()
	host, found, err := s.store.TakePending(ctx, workspaceID, nonce, verifier)
	switch {
	case err != nil:
		return failed, "store"
	case !found:
		return failed, "bad_state"
	case q.Get("error") == "access_denied":
		return OAuthResult{WorkspaceID: workspaceID, Outcome: OutcomeCancelled}, "cancelled"
	case q.Get("error") != "":
		return failed, "denied"
	case q.Get("code") == "":
		return failed, "missing_code"
	}
	return s.signIn(ctx, workspaceID, host, q.Get("code"))
}

// signIn exchanges the code, verifies the token with Myself and stores the
// connection, checking the switch before the exchange and again before the write.
func (s *Service) signIn(ctx context.Context, workspaceID, host, code string) (OAuthResult, string) {
	failed := OAuthResult{WorkspaceID: workspaceID, Outcome: OutcomeFailed}
	if s.RequireEnabled(ctx, workspaceID) != nil {
		return failed, "disabled"
	}
	cfg, err := s.oauthConfig(ctx)
	if err != nil {
		return failed, "not_configured"
	}
	ctx = redact.WithSecrets(ctx, cfg.ClientSecret)
	tokens, err := s.gateway.ExchangeOAuthCode(ctx, host, cfg.Client(), code, cfg.RedirectURI())
	if err != nil {
		return failed, "exchange"
	}
	ctx = redact.WithSecrets(ctx, tokens.AccessToken, tokens.RefreshToken)
	user, err := s.gateway.Myself(backlog.NoRetry(ctx), backlog.Credentials{SpaceHost: host, AccessToken: tokens.AccessToken})
	if err != nil {
		return failed, "verify"
	}
	if s.RequireEnabled(ctx, workspaceID) != nil {
		return failed, "disabled"
	}
	prev, next, _, err := s.store.saveConnection(ctx, workspaceID, host, oauthSecret(tokens), user)
	if err != nil {
		return failed, "store"
	}
	restored := s.emit(ctx, workspaceID, prev, next)
	redact.Logger(ctx).InfoContext(ctx, "oauth completed", "event", "oauth_completed", "spaceHost", host,
		"connectionEpoch", next.ConnectionEpoch, "backlogUserId", user.ID)
	return OAuthResult{WorkspaceID: workspaceID, Outcome: OutcomeConnected, Restored: restored}, ""
}
