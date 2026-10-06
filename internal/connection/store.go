package connection

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/backlog"
	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/redact"
)

// ErrStore marks a failure of Kandev's state or secret store. The action
// layer maps it to the internal error code.
var ErrStore = errors.New("connection store failure")

// Connection view states (C5 ConnectionView.state).
const (
	StateNotConnected = "not_connected"
	StateConnected    = "connected"
	StateError        = "error"
)

const (
	stateScope    = "workspace"
	stateKeyName  = "connection"
	switchKeyName = "integration"
	secretPrefix  = "backlog.connection."
	schemaVersion = 1
	authAPIKey    = "api_key"
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
}

// View is the ConnectionView returned to the UI. It has no field that can
// hold the key (BR3.2).
type View struct {
	Connected         bool   `json:"connected"`
	Enabled           bool   `json:"enabled"` // IntegrationSwitch value (BR7.1)
	State             string `json:"state"`
	SpaceHost         string `json:"spaceHost,omitempty"`
	AuthMethod        string `json:"authMethod,omitempty"`
	ConnectedUserName string `json:"connectedUserName,omitempty"`
	HasAPIKey         bool   `json:"hasApiKey"`
	HasOAuthToken     bool   `json:"hasOAuthToken"`
	HasGitCredential  bool   `json:"hasGitCredential"`
	ConnectionEpoch   int    `json:"connectionEpoch,omitempty"`
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
}

// switchRecord is the IntegrationSwitch stored in plugin state. Enabled is a
// pointer so a missing value is told apart from false (fail closed, NFR3.9).
type switchRecord struct {
	SchemaVersion int    `json:"schemaVersion"`
	Enabled       *bool  `json:"enabled"`
	ChangedAt     string `json:"changedAt,omitempty"`
}

// secret is the ApiKeySecret value, written in one atomic SetSecret.
type secret struct {
	APIKey          string `json:"apiKey"`
	SpaceHost       string `json:"spaceHost"`
	ConnectionEpoch int    `json:"connectionEpoch"`
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
	rec, found, err := s.readRecord(ctx, workspaceID)
	if err != nil {
		return View{}, err
	}
	if !found {
		return View{State: StateNotConnected, Enabled: enabled}, nil
	}
	sec, _, err := s.readSecret(ctx, workspaceID)
	if err != nil {
		return View{}, err
	}
	v := viewOf(rec, sec)
	v.Enabled = enabled
	return v, nil
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
	value, err := toMap(switchRecord{
		SchemaVersion: schemaVersion, Enabled: &enabled, ChangedAt: s.Now().UTC().Format(time.RFC3339),
	})
	if err == nil {
		err = s.call(ctx, func(c context.Context) error {
			return s.state.SetState(c, stateScope, workspaceID, switchKeyName, value)
		})
	}
	if err != nil {
		return s.storeErr(ctx, "write switch", err)
	}
	return nil
}

// Save stores a verified connection in the BR2.8 order and returns its view.
func (s *Store) Save(ctx context.Context, workspaceID, host, apiKey string, user backlog.User) (View, error) {
	rec, _, err := s.readRecord(ctx, workspaceID)
	if err != nil {
		return View{}, err
	}
	prevSec, prevRaw, err := s.readSecret(ctx, workspaceID)
	if err != nil {
		return View{}, err
	}
	prevEpoch := max(rec.ConnectionEpoch, prevSec.ConnectionEpoch)
	newSec := secret{APIKey: apiKey, SpaceHost: host, ConnectionEpoch: prevEpoch + 1}
	newRec := record{
		SchemaVersion: schemaVersion, SpaceHost: host, AuthMethod: authAPIKey,
		ConnectedUserName: user.Name, ConnectedUserID: user.ID,
		ConnectionEpoch: newSec.ConnectionEpoch, ConnectedAt: s.Now().UTC().Format(time.RFC3339),
	}

	secJSON, err := json.Marshal(newSec) //nolint:gosec // G117: this value goes only to Kandev's encrypted secret store (BR3.1)
	if err != nil {
		return View{}, fmt.Errorf("encode secret: %w", ErrStore)
	}
	if err := s.call(ctx, func(c context.Context) error {
		return s.secrets.SetSecret(c, secretPrefix+workspaceID, string(secJSON))
	}); err != nil {
		return View{}, s.rollback(ctx, workspaceID, prevRaw, prevEpoch, newSec.ConnectionEpoch, err)
	}
	recMap, err := toMap(newRec)
	if err == nil {
		err = s.call(ctx, func(c context.Context) error {
			return s.state.SetState(c, stateScope, workspaceID, stateKeyName, recMap)
		})
	}
	if err != nil {
		return View{}, s.rollback(ctx, workspaceID, prevRaw, prevEpoch, newSec.ConnectionEpoch, err)
	}
	return viewOf(newRec, newSec), nil
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

// viewOf applies BR2.11: connected only when record and secret agree.
func viewOf(rec record, sec secret) View {
	v := View{
		State: StateError, SpaceHost: rec.SpaceHost, AuthMethod: rec.AuthMethod,
		ConnectedUserName: rec.ConnectedUserName, ConnectionEpoch: rec.ConnectionEpoch,
	}
	if sec.APIKey != "" && sec.ConnectionEpoch == rec.ConnectionEpoch && sec.SpaceHost == rec.SpaceHost {
		v.State, v.Connected, v.HasAPIKey = StateConnected, true, true
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
