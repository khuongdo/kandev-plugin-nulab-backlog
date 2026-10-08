package scm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"sync"
	"time"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/connection"
)

const (
	schemaVersion = 1
	keySettings   = "scm.settings"
	keyLinks      = "scm.links"
	keyDismissed  = "scm.dismissed"
	keyQueries    = "scm.queries"
	keyWatches    = "scm.watches"
	keyLedger     = "scm.ledger"
	keyIndex      = "scm.index" // instance key: workspaces with links or watches
	maxLinks      = 500
	maxQueries    = 50
	maxWatches    = 50
)

var errLimit = errors.New("the list is full")

// errUnchanged tells an update that fn changed nothing, so nothing is written.
var errUnchanged = errors.New("unchanged")

// Mapping is the repositories of one provider mapped to one Backlog project (FR3.1).
type Mapping struct {
	ProjectKey string   `json:"projectKey"`
	Repos      []string `json:"repos"`
}

// Settings is one provider's workspace settings. The token itself lives only
// in Kandev's secret store (NFR1); HasToken says whether one is stored.
type Settings struct {
	Provider  Provider  `json:"provider"`
	Source    string    `json:"source,omitempty"` // "" or MethodToken: a typed token; MethodCLI: the CLI login (FR6.1)
	HasToken  bool      `json:"hasToken"`
	Account   string    `json:"account,omitempty"`   // the token's account name (FR2.4)
	AccountID string    `json:"accountId,omitempty"` // matches PullRequest.AuthorID for "me"
	LastError string    `json:"lastError,omitempty"` // the error code of the last failed test
	Mappings  []Mapping `json:"mappings"`
}

// Link associates a pull request with a Kandev task (manual, FR5.1) or a
// Backlog issue (automatic, FR5.2).
type Link struct {
	PRRef
	TaskID     string `json:"taskId,omitempty"`
	IssueKey   string `json:"issueKey,omitempty"`
	Auto       bool   `json:"auto,omitempty"`
	ProjectKey string `json:"projectKey,omitempty"`
	Title      string `json:"title,omitempty"`
	State      string `json:"state,omitempty"`
	URL        string `json:"url"`
}

// Query is a saved pull request query of one provider (FR4.2).
type Query struct {
	QueryInput
	IsDefault bool `json:"isDefault,omitempty"` // one per provider
	// Unmapped is set on read when the repository is no longer mapped (FR3.4).
	Unmapped bool `json:"unmapped,omitempty"`
}

// Watch states.
const (
	WatchActive = "active"
	WatchPaused = "paused"
)

// Watch is a saved pull request watch of one provider (FR4.3).
type Watch struct {
	WatchInput
	State        string `json:"state"`
	CreatedCount int    `json:"createdCount"`
	LastRunAt    string `json:"lastRunAt,omitempty"`
	Unmapped     bool   `json:"unmapped,omitempty"` // set on read (FR3.4)
}

// LedgerEntry records that a watch created (or reserved) the task of one
// pull request, so the task is created at most once (FR4.3).
type LedgerEntry struct {
	Key     string `json:"key"`              // PRRef.Key
	TaskID  string `json:"taskId,omitempty"` // empty while only reserved
	WatchID string `json:"watchId"`
}

// StateStore is the part of Kandev's state store the package needs.
// pluginsdk.Host satisfies it.
type StateStore interface {
	GetState(ctx context.Context, scope, scopeID, key string) (map[string]any, bool, error)
	SetState(ctx context.Context, scope, scopeID, key string, value map[string]any) error
}

// Store keeps the workspace documents. Each is one state value
// {schemaVersion, items}; a read-modify-write holds the workspace's mutex.
// ponytail: one document per list, capped; split per key if the caps are hit.
type Store struct {
	state       StateStore
	CallTimeout time.Duration // limit per store call (1 s)

	mu    sync.Mutex
	locks map[string]*sync.Mutex
}

// NewStore returns a Store with the 1-second call limit.
func NewStore(state StateStore) *Store {
	return &Store{state: state, CallTimeout: time.Second, locks: map[string]*sync.Mutex{}}
}

// Settings returns the providers' settings.
func (s *Store) Settings(ctx context.Context, ws string) ([]Settings, error) {
	return load[Settings](ctx, s, "workspace", ws, keySettings)
}

// Links returns the pull request links.
func (s *Store) Links(ctx context.Context, ws string) ([]Link, error) {
	return load[Link](ctx, s, "workspace", ws, keyLinks)
}

// Dismissed returns the removed auto-links, as "<PRRef.Key>|<issue key>".
func (s *Store) Dismissed(ctx context.Context, ws string) ([]string, error) {
	return load[string](ctx, s, "workspace", ws, keyDismissed)
}

// Queries returns the saved queries.
func (s *Store) Queries(ctx context.Context, ws string) ([]Query, error) {
	return load[Query](ctx, s, "workspace", ws, keyQueries)
}

// Watches returns the watches.
func (s *Store) Watches(ctx context.Context, ws string) ([]Watch, error) {
	return load[Watch](ctx, s, "workspace", ws, keyWatches)
}

// Ledger returns the watch ledger.
func (s *Store) Ledger(ctx context.Context, ws string) ([]LedgerEntry, error) {
	return load[LedgerEntry](ctx, s, "workspace", ws, keyLedger)
}

// Index lists the workspaces that have links or watches.
func (s *Store) Index(ctx context.Context) ([]string, error) {
	return load[string](ctx, s, "instance", "", keyIndex)
}

// UpdateSettings replaces the settings with fn's result.
func (s *Store) UpdateSettings(ctx context.Context, ws string, fn func([]Settings) ([]Settings, error)) error {
	return update(ctx, s, ws, keySettings, 0, false, fn)
}

// UpdateLinks replaces the links with fn's result.
func (s *Store) UpdateLinks(ctx context.Context, ws string, fn func([]Link) ([]Link, error)) error {
	return update(ctx, s, ws, keyLinks, maxLinks, true, fn)
}

// UpdateDismissed replaces the dismissed auto-links with fn's result.
// ponytail: never pruned; prune entries of deleted links if it grows large.
func (s *Store) UpdateDismissed(ctx context.Context, ws string, fn func([]string) ([]string, error)) error {
	return update(ctx, s, ws, keyDismissed, 0, false, fn)
}

// UpdateQueries replaces the saved queries with fn's result.
func (s *Store) UpdateQueries(ctx context.Context, ws string, fn func([]Query) ([]Query, error)) error {
	return update(ctx, s, ws, keyQueries, maxQueries, false, fn)
}

// UpdateWatches replaces the watches with fn's result.
func (s *Store) UpdateWatches(ctx context.Context, ws string, fn func([]Watch) ([]Watch, error)) error {
	return update(ctx, s, ws, keyWatches, maxWatches, true, fn)
}

// UpdateLedger replaces the ledger with fn's result.
// ponytail: never pruned, like git.ledger; prune entries of deleted watches if it grows large.
func (s *Store) UpdateLedger(ctx context.Context, ws string, fn func([]LedgerEntry) ([]LedgerEntry, error)) error {
	return update(ctx, s, ws, keyLedger, 0, false, fn)
}

func (s *Store) lock(name string) func() {
	s.mu.Lock()
	l, ok := s.locks[name]
	if !ok {
		l = &sync.Mutex{}
		s.locks[name] = l
	}
	s.mu.Unlock()
	l.Lock()
	return l.Unlock
}

type doc[T any] struct {
	SchemaVersion int `json:"schemaVersion"`
	Items         []T `json:"items"`
}

func load[T any](ctx context.Context, s *Store, scope, scopeID, key string) ([]T, error) {
	var (
		value map[string]any
		found bool
	)
	err := s.call(ctx, func(c context.Context) error {
		var err error
		value, found, err = s.state.GetState(c, scope, scopeID, key)
		return err
	})
	if err != nil {
		return nil, storeErr(ctx, "read "+key, err)
	}
	if !found {
		return nil, nil
	}
	var d doc[T]
	b, err := json.Marshal(value)
	if err == nil {
		err = json.Unmarshal(b, &d)
	}
	if err != nil || d.SchemaVersion != schemaVersion {
		// An unknown or broken document is never overwritten.
		return nil, fmt.Errorf("read %s: unsupported schema: %w", key, connection.ErrStore)
	}
	return d.Items, nil
}

// update runs one read-modify-write under the workspace mutex. limit 0 means
// no cap; indexed adds the workspace to the instance index when items exist.
func update[T any](ctx context.Context, s *Store, ws, key string, limit int, indexed bool, fn func([]T) ([]T, error)) error {
	defer s.lock("workspace/" + ws)()
	items, err := load[T](ctx, s, "workspace", ws, key)
	if err != nil {
		return err
	}
	next, err := fn(items)
	if errors.Is(err, errUnchanged) {
		return nil
	}
	if err != nil {
		return err
	}
	if limit > 0 && len(next) > limit {
		return &connection.FieldError{Field: FieldLimit, Err: errLimit}
	}
	if err := s.put(ctx, "workspace", ws, key, next); err != nil {
		return err
	}
	if indexed && len(next) > 0 {
		return s.addIndex(ctx, ws)
	}
	return nil
}

func (s *Store) addIndex(ctx context.Context, ws string) error {
	defer s.lock("instance")()
	idx, err := s.Index(ctx)
	if err != nil || slices.Contains(idx, ws) {
		return err
	}
	return s.put(ctx, "instance", "", keyIndex, append(idx, ws))
}

func (s *Store) put(ctx context.Context, scope, scopeID, key string, items any) error {
	b, err := json.Marshal(map[string]any{"schemaVersion": schemaVersion, "items": items})
	var m map[string]any
	if err == nil {
		err = json.Unmarshal(b, &m)
	}
	if err == nil {
		err = s.call(ctx, func(c context.Context) error { return s.state.SetState(c, scope, scopeID, key, m) })
	}
	if err != nil {
		return storeErr(ctx, "write "+key, err)
	}
	return nil
}

func (s *Store) call(ctx context.Context, fn func(context.Context) error) error {
	c, cancel := context.WithTimeout(ctx, s.CallTimeout)
	defer cancel()
	return fn(c)
}

// storeErr returns the caller's cancellation unchanged; anything else is
// connection.ErrStore (code internal).
func storeErr(ctx context.Context, op string, err error) error {
	if errors.Is(ctx.Err(), context.Canceled) {
		return ctx.Err()
	}
	return fmt.Errorf("%s: %w: %w", op, connection.ErrStore, err)
}
