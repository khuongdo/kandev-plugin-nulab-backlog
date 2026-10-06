package connection

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/backlog"
	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/redact"
)

// ErrStore marks a failure of Kandev's state or secret store. The action
// layer maps it to the internal error code.
var ErrStore = errors.New("connection store failure")

// ErrStale is returned when the connection changed after it was read, so a
// late write would overwrite a newer connection (R-08).
var ErrStale = errors.New("the connection changed meanwhile")

// ErrNotConnected is returned when the workspace has no usable connection.
var ErrNotConnected = errors.New("backlog is not connected for this workspace")

// Connection view states (C5 ConnectionView.state).
const (
	StateNotConnected = "not_connected"
	StateConnected    = "connected"
	StateError        = "error"
	// StateSignInAgain means the OAuth refresh was refused (AC1.4.3).
	StateSignInAgain = "sign_in_again"
)

const (
	stateScope    = "workspace"
	stateKeyName  = "connection"
	switchKeyName = "integration"
	secretPrefix  = "backlog.connection."
	gitPrefix     = "backlog.git." // U4: the Git credential (US5.5)
	schemaVersion = 1
	authAPIKey    = "api_key"
	authOAuth     = "oauth"
	pendingKey    = "oauth_pending"
)

// SecretStore is the part of Kandev's secret store the connection needs.
// pluginsdk.Host satisfies it.
type SecretStore interface {
	GetSecret(ctx context.Context, key string) (string, bool, error)
	SetSecret(ctx context.Context, key, value string) error
	DeleteSecret(ctx context.Context, key string) error
}

// StateStore is the part of Kandev's state store the connection needs.
// pluginsdk.Host satisfies it.
type StateStore interface {
	GetState(ctx context.Context, scope, scopeID, key string) (map[string]any, bool, error)
	SetState(ctx context.Context, scope, scopeID, key string, value map[string]any) error
	DeleteState(ctx context.Context, scope, scopeID, key string) error
}

// View is the ConnectionView returned to the UI. It has no field that can
// hold the key (BR3.2).
type View struct {
	Connected         bool     `json:"connected"`
	Enabled           bool     `json:"enabled"` // IntegrationSwitch value (BR7.1)
	State             string   `json:"state"`
	SpaceHost         string   `json:"spaceHost,omitempty"`
	AuthMethod        string   `json:"authMethod,omitempty"`
	ConnectedUserName string   `json:"connectedUserName,omitempty"`
	HasAPIKey         bool     `json:"hasApiKey"`
	HasOAuthToken     bool     `json:"hasOAuthToken"`
	HasGitCredential  bool     `json:"hasGitCredential"`
	ConnectionEpoch   int      `json:"connectionEpoch,omitempty"`
	SelectedProjects  []string `json:"selectedProjects"`
	// Restored is set on the reply of a connect that restored an earlier
	// connection to this space (M12). It is never stored.
	Restored bool `json:"restored,omitempty"`
	// GitCheck is set on the connection.test reply when a Git credential is
	// stored: ok, invalid or untested (AC5.5.2). It is never stored.
	GitCheck string `json:"gitCheck,omitempty"`
}

// record is the SpaceConnection stored in plugin state. It holds no secret.
type record struct {
	SchemaVersion     int    `json:"schemaVersion"`
	SpaceHost         string `json:"spaceHost"`
	AuthMethod        string `json:"authMethod"`
	ConnectedUserName string `json:"connectedUserName"`
	ConnectedUserID   int64  `json:"connectedUserId"`
	ConnectionEpoch   int    `json:"connectionEpoch"`
	ConnectedAt       string `json:"connectedAt"`

	// U2 fields. All are optional, so U1 records still decode (schemaVersion 1).
	SelectedProjects  []string `json:"selectedProjects,omitempty"`
	PreviousSpaceHost string   `json:"previousSpaceHost,omitempty"` // remembered for a restore (M12)
	PreviousProjects  []string `json:"previousProjects,omitempty"`
	Disconnected      bool     `json:"disconnected,omitempty"` // a disconnect record keeps the epoch
	SignInAgain       bool     `json:"signInAgain,omitempty"`  // the OAuth refresh was refused
}

// switchRecord is the IntegrationSwitch stored in plugin state. Enabled is a
// pointer so a missing value is told apart from false (fail closed, NFR3.9).
type switchRecord struct {
	SchemaVersion int    `json:"schemaVersion"`
	Enabled       *bool  `json:"enabled"`
	ChangedAt     string `json:"changedAt,omitempty"`
}

// secret is the connection secret, written in one atomic SetSecret. It holds
// either an API key or an OAuth token pair.
type secret struct {
	APIKey          string `json:"apiKey,omitempty"`
	SpaceHost       string `json:"spaceHost"`
	ConnectionEpoch int    `json:"connectionEpoch"`

	// U2 fields; a U1 secret has none and is an API key secret.
	AuthMethod   string `json:"authMethod,omitempty"`
	AccessToken  string `json:"accessToken,omitempty"`
	RefreshToken string `json:"refreshToken,omitempty"`
	ExpiresAt    string `json:"expiresAt,omitempty"` // RFC 3339
}

// method returns the secret's auth method; a U1 secret is api_key.
func (s secret) method() string {
	if s.AuthMethod == "" {
		return authAPIKey
	}
	return s.AuthMethod
}

// Store writes and reads the epoch-linked record and secret (BR2.8, BR2.11).
type Store struct {
	secrets SecretStore
	state   StateStore

	CallTimeout     time.Duration    // limit per store call (1 s)
	RollbackTimeout time.Duration    // limit for the rollback on a fresh context (2 s)
	Now             func() time.Time // clock for connectedAt
}

// NewStore returns a Store with the production time limits.
func NewStore(secrets SecretStore, state StateStore) *Store {
	return &Store{secrets: secrets, state: state, CallTimeout: time.Second, RollbackTimeout: 2 * time.Second, Now: time.Now}
}

// Load builds the view for a workspace (WF2). It never returns the key.
func (s *Store) Load(ctx context.Context, workspaceID string) (View, error) {
	enabled, err := s.LoadSwitch(ctx, workspaceID)
	if err != nil {
		return View{}, err
	}
	p, err := s.readPair(ctx, workspaceID)
	if err != nil {
		return View{}, err
	}
	v := p.view()
	v.Enabled = enabled
	if p.connected() {
		g, err := s.readGit(ctx, workspaceID)
		if err != nil {
			return View{}, err
		}
		v.HasGitCredential = g.matches(p.rec.SpaceHost)
	}
	return v, nil
}

// pair is the stored record and secret of one workspace.
type pair struct {
	rec   record
	found bool // a record exists
	sec   secret
	raw   string // the raw secret, for a rollback
}

func (p pair) view() View {
	if !p.found {
		return View{State: StateNotConnected}
	}
	return viewOf(p.rec, p.sec)
}

// epoch is the higher of the record and secret epochs, so the next epoch
// never goes back (R-08).
func (p pair) epoch() int { return max(p.rec.ConnectionEpoch, p.sec.ConnectionEpoch) }

// connected applies BR2.11: record and secret agree and hold a credential.
func (p pair) connected() bool {
	return p.found && !p.rec.Disconnected && (p.sec.APIKey != "" || p.sec.AccessToken != "") &&
		p.sec.ConnectionEpoch == p.rec.ConnectionEpoch && p.sec.SpaceHost == p.rec.SpaceHost
}

// readPair reads the record, then the secret. A leftover secret without a
// record is still read, so its epoch and value are kept for a rollback.
func (s *Store) readPair(ctx context.Context, workspaceID string) (pair, error) {
	rec, found, err := s.readRecord(ctx, workspaceID)
	if err != nil {
		return pair{}, err
	}
	sec, raw, err := s.readSecret(ctx, workspaceID)
	if err != nil {
		return pair{}, err
	}
	return pair{rec: rec, found: found, sec: sec, raw: raw}, nil
}

// LoadSwitch reports whether Backlog is on for the workspace. No record means
// on (BR7.1); a record that cannot be decoded is an error, never "on" (NFR3.9).
func (s *Store) LoadSwitch(ctx context.Context, workspaceID string) (bool, error) {
	value, found, err := s.getState(ctx, workspaceID, switchKeyName)
	if err != nil {
		return false, s.storeErr(ctx, "read switch", err)
	}
	if !found {
		return true, nil
	}
	var rec switchRecord
	if decode(value, &rec) != nil || rec.SchemaVersion != schemaVersion || rec.Enabled == nil {
		return false, fmt.Errorf("read switch: undecodable value: %w", ErrStore)
	}
	return *rec.Enabled, nil
}

// SaveSwitch writes only the IntegrationSwitch record; the connection record
// and the secret are never touched (BR7.4).
func (s *Store) SaveSwitch(ctx context.Context, workspaceID string, enabled bool) error {
	return s.setState(ctx, workspaceID, switchKeyName, switchRecord{
		SchemaVersion: schemaVersion, Enabled: &enabled, ChangedAt: s.Now().UTC().Format(time.RFC3339),
	}, "write switch")
}

// Save stores a verified API key connection in the BR2.8 order and returns its view.
func (s *Store) Save(ctx context.Context, workspaceID, host, apiKey string, user backlog.User) (View, error) {
	_, _, v, err := s.saveConnection(ctx, workspaceID, host, secret{APIKey: apiKey}, user)
	return v, err
}

// SaveOAuth stores a verified OAuth connection in the BR2.8 order.
func (s *Store) SaveOAuth(ctx context.Context, workspaceID, host string, tokens backlog.TokenSet, user backlog.User) (View, error) {
	_, _, v, err := s.saveConnection(ctx, workspaceID, host, oauthSecret(tokens), user)
	return v, err
}

func oauthSecret(t backlog.TokenSet) secret {
	return secret{AuthMethod: authOAuth, AccessToken: t.AccessToken, RefreshToken: t.RefreshToken,
		ExpiresAt: t.ExpiresAt.UTC().Format(time.RFC3339)}
}

// saveConnection is the one connect path: it handles a space change and a
// restore, bumps the epoch and writes in the BR2.8 order. It returns the
// record before and after, for the ConnectionChanged event.
func (s *Store) saveConnection(ctx context.Context, workspaceID, host string, sec secret, user backlog.User) (prev, next record, v View, err error) {
	p, err := s.readPair(ctx, workspaceID)
	if err != nil {
		return record{}, record{}, View{}, err
	}
	next = nextConnection(p.rec, host)
	next.SchemaVersion, next.SpaceHost, next.AuthMethod = schemaVersion, host, sec.method()
	next.ConnectedUserName, next.ConnectedUserID = user.Name, user.ID
	next.ConnectionEpoch, next.ConnectedAt = p.epoch()+1, s.Now().UTC().Format(time.RFC3339)
	sec.SpaceHost, sec.ConnectionEpoch = host, next.ConnectionEpoch
	if p.found && !p.rec.Disconnected && p.rec.SpaceHost != host {
		// A space change drops the old space's Git password before the new
		// connection is written (AC1.8.2, ADR-005).
		if err := s.deleteGit(ctx, workspaceID); err != nil {
			return record{}, record{}, View{}, err
		}
	}
	if err := s.write(ctx, workspaceID, next, sec, p); err != nil {
		return record{}, record{}, View{}, err
	}
	return p.rec, next, viewOf(next, sec), nil
}

// nextConnection carries the project selection and the restore memory into
// a new connection. Same host while connected keeps both. Another host
// clears the selection and remembers the old host and its projects; the
// remembered host puts its projects back (M12).
func nextConnection(prev record, host string) record {
	var next record
	connected := prev.SpaceHost != "" && !prev.Disconnected
	if connected && prev.SpaceHost == host {
		next.SelectedProjects, next.PreviousSpaceHost, next.PreviousProjects = prev.SelectedProjects, prev.PreviousSpaceHost, prev.PreviousProjects
		return next
	}
	restored := prev.PreviousSpaceHost == host
	if restored {
		next.SelectedProjects = prev.PreviousProjects
	} else if !connected {
		next.PreviousSpaceHost, next.PreviousProjects = prev.PreviousSpaceHost, prev.PreviousProjects
	}
	if connected {
		next.PreviousSpaceHost, next.PreviousProjects = prev.SpaceHost, prev.SelectedProjects
	}
	return next
}

// write stores secret then record (BR2.8); a failed step restores the
// previous secret from prev.
func (s *Store) write(ctx context.Context, workspaceID string, rec record, sec secret, prev pair) error {
	secJSON, err := json.Marshal(sec) //nolint:gosec // G117: this value goes only to Kandev's encrypted secret store (BR3.1)
	if err != nil {
		return fmt.Errorf("encode secret: %w", ErrStore)
	}
	if err := s.call(ctx, func(c context.Context) error {
		return s.secrets.SetSecret(c, secretPrefix+workspaceID, string(secJSON))
	}); err != nil {
		return s.rollback(ctx, workspaceID, prev.raw, prev.epoch(), rec.ConnectionEpoch, err)
	}
	recMap, err := toMap(rec)
	if err == nil {
		err = s.call(ctx, func(c context.Context) error {
			return s.state.SetState(c, stateScope, workspaceID, stateKeyName, recMap)
		})
	}
	if err != nil {
		return s.rollback(ctx, workspaceID, prev.raw, prev.epoch(), rec.ConnectionEpoch, err)
	}
	return nil
}

// UpdateTokens stores refreshed tokens in the secret only, keeping the epoch.
// If the secret's epoch is no longer epoch, nothing is written (ErrStale).
func (s *Store) UpdateTokens(ctx context.Context, workspaceID string, epoch int, tokens backlog.TokenSet) error {
	sec, _, err := s.readSecret(ctx, workspaceID)
	if err != nil {
		return err
	}
	if sec.ConnectionEpoch != epoch || sec.method() != authOAuth {
		return ErrStale
	}
	next := oauthSecret(tokens)
	next.SpaceHost, next.ConnectionEpoch = sec.SpaceHost, sec.ConnectionEpoch
	b, err := json.Marshal(next) //nolint:gosec // G117: this value goes only to Kandev's encrypted secret store (BR3.1)
	if err == nil {
		err = s.call(ctx, func(c context.Context) error {
			return s.secrets.SetSecret(c, secretPrefix+workspaceID, string(b))
		})
	}
	if err != nil {
		return s.storeErr(ctx, "write tokens", err)
	}
	return nil
}

// MarkSignInAgain records that the OAuth refresh was refused (AC1.4.3).
func (s *Store) MarkSignInAgain(ctx context.Context, workspaceID string, epoch int) error {
	rec, found, err := s.readRecord(ctx, workspaceID)
	if err != nil {
		return err
	}
	if !found || rec.ConnectionEpoch != epoch {
		return ErrStale
	}
	rec.SignInAgain = true
	return s.setState(ctx, workspaceID, stateKeyName, rec, "write record")
}

// Disconnect deletes the secret and the Git secret, then rewrites the record as a disconnect
// record with epoch + 1 that remembers the host and its projects. changed is
// false when there was no connection to remove.
func (s *Store) Disconnect(ctx context.Context, workspaceID string) (v View, changed bool, err error) {
	p, err := s.readPair(ctx, workspaceID)
	if err != nil {
		return View{}, false, err
	}
	if !p.found || p.rec.Disconnected {
		return p.view(), false, nil
	}
	if err := s.call(ctx, func(c context.Context) error {
		return s.secrets.DeleteSecret(c, secretPrefix+workspaceID)
	}); err != nil {
		return View{}, false, s.storeErr(ctx, "delete secret", err)
	}
	if err := s.deleteGit(ctx, workspaceID); err != nil {
		return View{}, false, err // no disconnect record (AC1.5.4)
	}
	next := record{
		SchemaVersion: schemaVersion, ConnectionEpoch: p.epoch() + 1, Disconnected: true,
		PreviousSpaceHost: p.rec.SpaceHost, PreviousProjects: p.rec.SelectedProjects,
	}
	if err := s.setState(ctx, workspaceID, stateKeyName, next, "write record"); err != nil {
		return View{}, false, err
	}
	return viewOf(next, secret{}), true, nil
}

// SaveProjects stores the project selection. The secret is rewritten with
// the same values and the new epoch, then the record, in the BR2.8 order.
func (s *Store) SaveProjects(ctx context.Context, workspaceID string, keys []string) (View, error) {
	_, _, v, err := s.saveProjects(ctx, workspaceID, keys)
	return v, err
}

func (s *Store) saveProjects(ctx context.Context, workspaceID string, keys []string) (prev, next record, v View, err error) {
	p, err := s.readPair(ctx, workspaceID)
	if err != nil {
		return record{}, record{}, View{}, err
	}
	if !p.connected() {
		return record{}, record{}, View{}, ErrNotConnected
	}
	if slices.Equal(p.rec.SelectedProjects, keys) {
		return p.rec, p.rec, viewOf(p.rec, p.sec), nil // unchanged: nothing to write
	}
	next, sec := p.rec, p.sec
	next.SelectedProjects = keys
	next.ConnectionEpoch = p.epoch() + 1
	sec.ConnectionEpoch = next.ConnectionEpoch
	if err := s.write(ctx, workspaceID, next, sec, p); err != nil {
		return record{}, record{}, View{}, err
	}
	return p.rec, next, viewOf(next, sec), nil
}

// pendingRecord is the single-use OAuth sign-in in progress (state key
// oauth_pending). It holds only a hash of the nonce.
type pendingRecord struct {
	SchemaVersion int    `json:"schemaVersion"`
	NonceHash     string `json:"nonceHash"`
	SpaceHost     string `json:"spaceHost"`
	ExpiresAt     string `json:"expiresAt"`
}

// SavePending stores the sign-in in progress; a newer one replaces it.
func (s *Store) SavePending(ctx context.Context, workspaceID string, nonce []byte, host string, expiresAt time.Time) error {
	sum := sha256.Sum256(nonce)
	return s.setState(ctx, workspaceID, pendingKey, pendingRecord{
		SchemaVersion: schemaVersion, NonceHash: hex.EncodeToString(sum[:]),
		SpaceHost: host, ExpiresAt: expiresAt.UTC().Format(time.RFC3339),
	}, "write pending sign-in")
}

// TakePending returns the host of the pending sign-in when nonce matches it
// (constant time) and it has not expired. A matching record is deleted
// before it is returned, so it can be used once; an expired one is deleted.
func (s *Store) TakePending(ctx context.Context, workspaceID string, nonce []byte) (string, bool, error) {
	value, found, err := s.getState(ctx, workspaceID, pendingKey)
	if err != nil {
		return "", false, s.storeErr(ctx, "read pending sign-in", err)
	}
	var p pendingRecord
	if !found || decode(value, &p) != nil || p.SchemaVersion != schemaVersion {
		return "", false, nil
	}
	if exp, err := time.Parse(time.RFC3339, p.ExpiresAt); err != nil || !s.Now().Before(exp) {
		return "", false, s.DeletePending(ctx, workspaceID)
	}
	want, _ := hex.DecodeString(p.NonceHash)
	got := sha256.Sum256(nonce)
	if subtle.ConstantTimeCompare(want, got[:]) != 1 {
		return "", false, nil
	}
	if err := s.DeletePending(ctx, workspaceID); err != nil {
		return "", false, err
	}
	return p.SpaceHost, true, nil
}

// DeletePending removes the sign-in in progress, if any.
func (s *Store) DeletePending(ctx context.Context, workspaceID string) error {
	if err := s.call(ctx, func(c context.Context) error {
		return s.state.DeleteState(c, stateScope, workspaceID, pendingKey)
	}); err != nil {
		return s.storeErr(ctx, "delete pending sign-in", err)
	}
	return nil
}

// setState writes one workspace state key under the per-call limit.
func (s *Store) setState(ctx context.Context, workspaceID, key string, v any, op string) error {
	m, err := toMap(v)
	if err == nil {
		err = s.call(ctx, func(c context.Context) error {
			return s.state.SetState(c, stateScope, workspaceID, key, m)
		})
	}
	if err != nil {
		return s.storeErr(ctx, op, err)
	}
	return nil
}

// rollback restores the previous secret, or deletes the new one, on a fresh
// context so a cancelled action cannot stop it (BR2.8 step 4). It returns the
// error Save reports: the caller's cancellation unchanged, else ErrStore.
func (s *Store) rollback(ctx context.Context, workspaceID, prevRaw string, prevEpoch, newEpoch int, cause error) error {
	rctx, cancel := context.WithTimeout(context.Background(), s.RollbackTimeout)
	defer cancel()
	var err error
	if prevRaw != "" {
		err = s.secrets.SetSecret(rctx, secretPrefix+workspaceID, prevRaw)
	} else {
		err = s.secrets.DeleteSecret(rctx, secretPrefix+workspaceID)
	}
	if err != nil {
		redact.Logger(ctx).ErrorContext(ctx, "connection inconsistent", "event", "connection_inconsistent",
			"previousEpoch", prevEpoch, "newEpoch", newEpoch)
	}
	if errors.Is(ctx.Err(), context.Canceled) {
		return ctx.Err()
	}
	return fmt.Errorf("store connection: %w: %w", ErrStore, cause)
}

func (s *Store) readRecord(ctx context.Context, workspaceID string) (record, bool, error) {
	value, found, err := s.getState(ctx, workspaceID, stateKeyName)
	if err != nil {
		return record{}, false, s.storeErr(ctx, "read record", err)
	}
	if !found {
		return record{}, false, nil
	}
	var rec record
	if err := decode(value, &rec); err != nil || rec.SchemaVersion != schemaVersion {
		// An unknown or broken record is never overwritten (contract: schemaVersion).
		return record{}, false, fmt.Errorf("read record: unsupported schema: %w", ErrStore)
	}
	return rec, true, nil
}

// getState reads one workspace state key under the per-call limit.
func (s *Store) getState(ctx context.Context, workspaceID, key string) (map[string]any, bool, error) {
	var (
		value map[string]any
		found bool
	)
	err := s.call(ctx, func(c context.Context) error {
		var err error
		value, found, err = s.state.GetState(c, stateScope, workspaceID, key)
		return err
	})
	return value, found, err
}

// readSecret returns the parsed secret and its raw value. An unreadable value
// parses as the zero secret, which never matches a record.
func (s *Store) readSecret(ctx context.Context, workspaceID string) (secret, string, error) {
	var raw string
	err := s.call(ctx, func(c context.Context) error {
		var err error
		raw, _, err = s.secrets.GetSecret(c, secretPrefix+workspaceID)
		return err
	})
	if err != nil {
		return secret{}, "", s.storeErr(ctx, "read secret", err)
	}
	var sec secret
	if raw != "" && json.Unmarshal([]byte(raw), &sec) != nil {
		sec = secret{}
	}
	return sec, raw, nil
}

// call runs one store call under the per-call limit.
func (s *Store) call(ctx context.Context, fn func(context.Context) error) error {
	c, cancel := context.WithTimeout(ctx, s.CallTimeout)
	defer cancel()
	return fn(c)
}

// storeErr returns a cancellation unchanged. Anything else, including a
// store step that ran out of time, is a store error (code internal); the
// wrapped error still matches context.DeadlineExceeded with errors.Is.
func (s *Store) storeErr(ctx context.Context, op string, err error) error {
	if errors.Is(ctx.Err(), context.Canceled) {
		return ctx.Err()
	}
	return fmt.Errorf("%s: %w: %w", op, ErrStore, err)
}

// viewOf applies BR2.11: connected only when record and secret agree. A
// disconnect record is not_connected.
func viewOf(rec record, sec secret) View {
	if rec.Disconnected {
		return View{State: StateNotConnected, ConnectionEpoch: rec.ConnectionEpoch}
	}
	v := View{
		State: StateError, SpaceHost: rec.SpaceHost, AuthMethod: rec.AuthMethod,
		ConnectedUserName: rec.ConnectedUserName, ConnectionEpoch: rec.ConnectionEpoch,
		SelectedProjects: rec.SelectedProjects,
	}
	if (pair{rec: rec, found: true, sec: sec}).connected() {
		v.HasAPIKey, v.HasOAuthToken = sec.APIKey != "", sec.AccessToken != ""
		v.State, v.Connected = StateConnected, true
		if rec.SignInAgain {
			v.State, v.Connected = StateSignInAgain, false
		}
	}
	return v
}

// decode converts a state map into a typed record.
func decode(value map[string]any, out any) error {
	b, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, out)
}

func toMap(v any) (map[string]any, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var m map[string]any
	return m, json.Unmarshal(b, &m)
}

// SaveGit stores the Git credential bound to the connected space host, with
// the revision one above the stored one (US5.5).
func (s *Store) SaveGit(ctx context.Context, workspaceID, username, password string) error {
	p, err := s.readPair(ctx, workspaceID)
	if err != nil {
		return err
	}
	if !p.connected() {
		return ErrNotConnected
	}
	prev, err := s.readGit(ctx, workspaceID)
	if err != nil {
		return err
	}
	b, err := json.Marshal(gitSecret{Username: username, Password: password, SpaceHost: p.rec.SpaceHost, Revision: prev.Revision + 1}) //nolint:gosec // G117: this value goes only to Kandev's encrypted secret store (BR3.1)
	if err == nil {
		err = s.call(ctx, func(c context.Context) error { return s.secrets.SetSecret(c, gitPrefix+workspaceID, string(b)) })
	}
	if err != nil {
		return s.storeErr(ctx, "write git secret", err)
	}
	return nil
}

// readGit returns the stored Git secret; a missing or unreadable one is the zero value.
func (s *Store) readGit(ctx context.Context, workspaceID string) (gitSecret, error) {
	var raw string
	err := s.call(ctx, func(c context.Context) error {
		var err error
		raw, _, err = s.secrets.GetSecret(c, gitPrefix+workspaceID)
		return err
	})
	if err != nil {
		return gitSecret{}, s.storeErr(ctx, "read git secret", err)
	}
	var g gitSecret
	if raw != "" && json.Unmarshal([]byte(raw), &g) != nil {
		g = gitSecret{}
	}
	return g, nil
}

func (s *Store) deleteGit(ctx context.Context, workspaceID string) error {
	if err := s.call(ctx, func(c context.Context) error { return s.secrets.DeleteSecret(c, gitPrefix+workspaceID) }); err != nil {
		return s.storeErr(ctx, "delete git secret", err)
	}
	return nil
}

// matches reports whether the secret holds a credential for host.
func (g gitSecret) matches(host string) bool {
	return g.Password != "" && g.Username != "" && g.SpaceHost == host
}
