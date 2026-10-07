package connection

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/backlog"
	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/redact"
)

// Action error codes (contract C5 ActionError.code) used in U1.
const (
	CodeValidation  = "validation"
	CodeRateLimited = "rate_limited"
	CodeUnreachable = "unreachable"
	CodeConflict    = "conflict"
	CodeInternal    = "internal"
	// CodeReconnectRequired means the stored credentials no longer work (U2).
	CodeReconnectRequired = "reconnect_required"
	// CodeIntegrationDisabled is returned while Backlog is off for the workspace (BR7.3).
	CodeIntegrationDisabled = "integration_disabled"
)

// Input field names reported with CodeValidation.
const (
	FieldSpaceURL = "spaceUrl"
	FieldAPIKey   = "apiKey"
	FieldEnabled  = "enabled"
	FieldOAuth    = "oauth"
)

// ErrConflict is returned when a Connect is already running in the workspace (BR2.6).
var ErrConflict = errors.New("another connect is running in this workspace")

// ErrIntegrationDisabled is returned while Backlog is off for the workspace (BR7.3).
var ErrIntegrationDisabled = errors.New("backlog is turned off for this workspace")

// ErrReconnectRequired is returned when Backlog refused the stored
// credentials, or the OAuth sign-in must be done again.
var ErrReconnectRequired = errors.New("the Backlog credentials no longer work; connect again")

// FieldError is a validation failure on one input field.
type FieldError struct {
	Field string
	Err   error
}

func (e *FieldError) Error() string { return fmt.Sprintf("invalid %s: %v", e.Field, e.Err) }
func (e *FieldError) Unwrap() error { return e.Err }

// Gateway is the Backlog calls the connection needs. backlog.Client
// implements it; tests swap in a fake.
type Gateway interface {
	Myself(ctx context.Context, creds backlog.Credentials) (backlog.User, error)
	Projects(ctx context.Context, creds backlog.Credentials) ([]backlog.Project, error)
	ExchangeOAuthCode(ctx context.Context, spaceHost string, client backlog.OAuthClient, code, redirectURI string) (backlog.TokenSet, error)
	RefreshToken(ctx context.Context, spaceHost string, client backlog.OAuthClient, refreshToken string) (backlog.TokenSet, error)
	// U4: the Git access check of connection.test (AC5.5.2).
	Repositories(ctx context.Context, creds backlog.Credentials, projectKey string) ([]backlog.Repository, error)
	CheckGitAccess(ctx context.Context, spaceHost, username, password, projectKey, repo string) error
}

// ConfigReader reads the plugin's operator config. pluginsdk.Host satisfies it.
type ConfigReader interface {
	GetConfig(ctx context.Context) (map[string]any, error)
}

// ConnectInput is the connection.connect_api_key request body.
type ConnectInput struct {
	SpaceURL string `json:"spaceUrl"`
	APIKey   string `json:"apiKey"`
}

// Service runs the connection workflows (U1 WF2, WF3, WF6 and U2).
type Service struct {
	gateway Gateway
	store   *Store
	// Config holds the OAuth app; nil means OAuth is not set up.
	Config ConfigReader

	// Connect budget (performance-design, NFR1.4). The rollback has its own
	// Store.RollbackTimeout on a fresh context, so the worst case is
	// Deadline + RollbackTimeout = 14 s.
	Deadline       time.Duration // everything except the rollback (12 s)
	PreCallTimeout time.Duration // switch read, validation and lock (1 s)
	BacklogTimeout time.Duration // the Myself call, also capped at Deadline - StoreTimeout (10 s)
	StoreTimeout   time.Duration // second switch read and the writes together (2 s)

	mu       sync.Mutex
	inflight map[string]struct{}      // ConnectAttempt per workspace
	locks    map[string]chan struct{} // one write lock per workspace
	hub      hub
}

// NewService returns a Service with the design time budget.
func NewService(gateway Gateway, store *Store) *Service {
	return &Service{
		gateway: gateway, store: store,
		Deadline: 12 * time.Second, PreCallTimeout: time.Second,
		BacklogTimeout: 10 * time.Second, StoreTimeout: 2 * time.Second,
		inflight: map[string]struct{}{}, locks: map[string]chan struct{}{},
	}
}

// Get returns the workspace's connection view (WF2).
func (s *Service) Get(ctx context.Context, workspaceID string) (View, error) {
	return s.store.Load(ctx, workspaceID)
}

// SwitchView is the connection.set_enabled reply. It carries only the switch
// value: building a full View needs the record and the secret, which NFR1.3
// rules out, and the UI reloads the full view with connection.get.
type SwitchView struct {
	Enabled bool `json:"enabled"`
}

// SetEnabled turns Backlog on or off for the workspace (WF6). It does one
// state read (the previous value, for the log), one state write and no secret
// access (NFR1.3, BR7.4). Once the write succeeds it never reports an error.
func (s *Service) SetEnabled(ctx context.Context, workspaceID string, enabled bool) (SwitchView, error) {
	start := time.Now()
	// An unreadable previous value must not block the write, so an admin can
	// always repair a broken switch; it is logged as "unknown".
	var previous any = "unknown"
	if prev, err := s.store.LoadSwitch(ctx, workspaceID); err == nil {
		previous = prev
	}
	if err := s.store.SaveSwitch(ctx, workspaceID, enabled); err != nil {
		return SwitchView{}, err
	}
	redact.Logger(ctx).InfoContext(ctx, "integration switch changed", "event", "integration_switch_changed",
		"previousEnabled", previous, "enabled", enabled, "durationMs", time.Since(start).Milliseconds())
	return SwitchView{Enabled: enabled}, nil
}

// Connect verifies the key with Backlog and stores the connection (WF3).
// Nothing is written unless Backlog confirms the key.
func (s *Service) Connect(ctx context.Context, workspaceID string, in ConnectInput) (View, error) {
	start := time.Now()
	ctx = redact.WithSecrets(ctx, in.APIKey)
	var info connectInfo
	view, err := s.connect(ctx, workspaceID, in, &info)
	logOutcome(ctx, info, view, err, time.Since(start))
	return view, err
}

// connectInfo carries what the outcome log needs but the view must not hold.
type connectInfo struct {
	host   string // set only once the address is valid
	userID int64
}

func (s *Service) connect(ctx context.Context, workspaceID string, in ConnectInput, info *connectInfo) (View, error) {
	ctx, cancel := context.WithTimeout(ctx, s.Deadline)
	defer cancel()
	addr, key, err := s.preCall(ctx, workspaceID, in, info)
	if err != nil {
		return View{}, err
	}
	defer s.unlock(workspaceID)

	user, err := s.verify(ctx, addr, key)
	if err != nil {
		return View{}, err
	}
	info.userID = user.ID

	storeCtx, cancelStore := context.WithTimeout(ctx, s.StoreTimeout)
	defer cancelStore()
	if err := s.RequireEnabled(storeCtx, workspaceID); err != nil {
		return View{}, err // turned off while Myself ran: write nothing (BR7.3)
	}
	unlock, err := s.lockWS(storeCtx, workspaceID)
	if err != nil {
		return View{}, err
	}
	defer unlock()
	prev, next, view, err := s.store.saveConnection(storeCtx, workspaceID, addr.Host, secret{APIKey: key}, user)
	if err != nil {
		return View{}, err
	}
	view.Enabled, view.Restored = true, s.emit(ctx, workspaceID, prev, next)
	return view, nil
}

// lockWS takes the workspace's write lock, so writes and the events they send
// keep their order. The returned func releases it.
func (s *Service) lockWS(ctx context.Context, workspaceID string) (func(), error) {
	s.mu.Lock()
	l, ok := s.locks[workspaceID]
	if !ok {
		l = make(chan struct{}, 1)
		s.locks[workspaceID] = l
	}
	s.mu.Unlock()
	select {
	case l <- struct{}{}:
		return func() { <-l }, nil
	case <-ctx.Done():
		return nil, s.store.storeErr(ctx, "wait for the workspace lock", ctx.Err())
	}
}

// preCall runs the steps before the Backlog call under PreCallTimeout: the
// switch, the input checks (no network, BR1.3) and the ConnectAttempt lock.
// On success the caller owns the lock.
func (s *Service) preCall(ctx context.Context, workspaceID string, in ConnectInput, info *connectInfo) (SpaceAddress, string, error) {
	pctx, cancel := context.WithTimeout(ctx, s.PreCallTimeout)
	defer cancel()
	if err := s.RequireEnabled(pctx, workspaceID); err != nil {
		return SpaceAddress{}, "", err
	}
	addr, err := ParseSpaceAddress(in.SpaceURL)
	if err != nil {
		return SpaceAddress{}, "", &FieldError{Field: FieldSpaceURL, Err: err}
	}
	info.host = addr.Host
	key, err := ValidateAPIKey(in.APIKey)
	if err != nil {
		return SpaceAddress{}, "", &FieldError{Field: FieldAPIKey, Err: err}
	}
	if !s.tryLock(workspaceID) {
		return SpaceAddress{}, "", ErrConflict
	}
	return addr, key, nil
}

// verify calls Myself for at most BacklogTimeout, and never past the
// deadline minus StoreTimeout, so the store steps keep their time.
func (s *Service) verify(ctx context.Context, addr SpaceAddress, key string) (backlog.User, error) {
	limit := s.BacklogTimeout
	if d, ok := ctx.Deadline(); ok {
		limit = min(limit, time.Until(d)-s.StoreTimeout)
	}
	callCtx, cancel := context.WithTimeout(ctx, limit)
	defer cancel()
	// A 429 is answered at once with its wait, never retried (NFR2.1).
	user, err := s.gateway.Myself(backlog.NoRetry(callCtx), backlog.Credentials{SpaceHost: addr.Host, APIKey: key})
	if err != nil {
		return backlog.User{}, verifyError(err)
	}
	return user, nil
}

// RequireEnabled returns ErrIntegrationDisabled while Backlog is off, and a
// store error when the switch cannot be read (fail closed, NFR3.9). The plugin
// guard calls it before every guarded action.
func (s *Service) RequireEnabled(ctx context.Context, workspaceID string) error {
	enabled, err := s.store.LoadSwitch(ctx, workspaceID)
	if err != nil {
		return err
	}
	if !enabled {
		return ErrIntegrationDisabled
	}
	return nil
}

// verifyError maps a rejected key or a missing space onto the input field
// (BR2.3, BR2.9); every other failure passes through unchanged.
func verifyError(err error) error {
	var be *backlog.Error
	if errors.As(err, &be) {
		switch be.Kind {
		case backlog.KindUnauthorized, backlog.KindForbidden:
			return &FieldError{Field: FieldAPIKey, Err: err}
		case backlog.KindNotFound:
			return &FieldError{Field: FieldSpaceURL, Err: err}
		}
	}
	return err
}

func (s *Service) tryLock(workspaceID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, busy := s.inflight[workspaceID]; busy {
		return false
	}
	s.inflight[workspaceID] = struct{}{}
	return true
}

func (s *Service) unlock(workspaceID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.inflight, workspaceID)
}

// Outcome is the action error an error maps to.
type Outcome struct {
	Code              string
	Field             string
	RetryAfterSeconds int
}

// Classify maps a Connect or Get error to its action error code (WF3 outcome
// table, functional-spec Error Outcomes). A nil error gives the zero Outcome.
func Classify(err error) Outcome {
	var fe *FieldError
	var be *backlog.Error
	switch {
	case err == nil:
		return Outcome{}
	case errors.As(err, &fe):
		return Outcome{Code: CodeValidation, Field: fe.Field}
	case errors.Is(err, ErrInvalidProjects):
		return Outcome{Code: CodeValidation, Field: FieldProjectKeys}
	case errors.Is(err, ErrOAuthNotConfigured):
		return Outcome{Code: CodeValidation, Field: FieldOAuth}
	case errors.Is(err, ErrNoGitCredential):
		return Outcome{Code: CodeValidation, Field: FieldGitCredential}
	case errors.Is(err, ErrConflict):
		return Outcome{Code: CodeConflict}
	case errors.Is(err, ErrIntegrationDisabled):
		return Outcome{Code: CodeIntegrationDisabled}
	case errors.Is(err, ErrReconnectRequired), errors.Is(err, ErrNotConnected):
		return Outcome{Code: CodeReconnectRequired}
	case errors.Is(err, ErrStore):
		return Outcome{Code: CodeInternal} // a store step that ran out of time is internal, not unreachable
	case errors.As(err, &be) && be.Kind == backlog.KindRateLimited:
		secs := int(math.Ceil(be.RetryAfter.Seconds()))
		return Outcome{Code: CodeRateLimited, RetryAfterSeconds: max(secs, 1)}
	case errors.As(err, &be) && (be.Kind == backlog.KindUnauthorized || be.Kind == backlog.KindForbidden):
		return Outcome{Code: CodeReconnectRequired} // a non-Connect call (R-06)
	case errors.As(err, &be), errors.Is(err, context.DeadlineExceeded):
		return Outcome{Code: CodeUnreachable}
	default:
		return Outcome{Code: CodeInternal}
	}
}

// logOutcome writes exactly one connect_succeeded or connect_failed event
// (NFR11.2). It never logs the key, a URL or the display name.
func logOutcome(ctx context.Context, info connectInfo, view View, err error, d time.Duration) {
	log := redact.Logger(ctx)
	if err == nil {
		log.InfoContext(ctx, "connect succeeded", "event", "connect_succeeded", "spaceHost", info.host,
			"backlogUserId", info.userID, "connectionEpoch", view.ConnectionEpoch, "durationMs", d.Milliseconds())
		return
	}
	attrs := []any{"event", "connect_failed", "errorCode", Classify(err).Code, "durationMs", d.Milliseconds()}
	if info.host != "" {
		attrs = append(attrs, "spaceHost", info.host)
	}
	var be *backlog.Error
	if errors.As(err, &be) && be.Status != 0 {
		attrs = append(attrs, "backlogStatus", be.Status)
	}
	log.WarnContext(ctx, "connect failed", attrs...)
}
