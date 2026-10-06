package git

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

// Item states of links and watches.
const (
	StatusActive       = "active"
	StatusPaused       = "paused" // watches only
	StatusNotConnected = "not_connected"
)

// FieldLimit is reported when a list is full.
const FieldLimit = "limit"

const (
	schemaVersion = 1
	keyLinks      = "git.links"
	keyWatches    = "git.watches"
	keyLedger     = "git.ledger"
	keyQueries    = "git.queries"
	keyIndex      = "git.watch_index" // instance key: workspaces with links or watches
	maxLinks      = 500
	maxWatches    = 50
	maxQueries    = 50
)

var errLimit = errors.New("the list is full")

// errUnchanged tells update that fn changed nothing, so nothing is written.
var errUnchanged = errors.New("unchanged")

// Link associates a Kandev task with a Backlog pull request.
type Link struct {
	TaskID       string `json:"taskId"`
	SpaceHost    string `json:"spaceHost"`
	ProjectKey   string `json:"projectKey"`
	RepoName     string `json:"repoName"`
	RepositoryID int64  `json:"repositoryId"`
	Number       int    `json:"number"`
	Title        string `json:"title,omitempty"`
	Status       string `json:"status"`
}

// Key is the link's reviewKey.
func (l Link) Key() string { return LinkKey(l.SpaceHost, l.RepositoryID, l.Number) }

// Watch is a saved PR watch (US6.1).
type Watch struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	SpaceHost      string   `json:"spaceHost"`
	ProjectKey     string   `json:"projectKey"`
	RepoName       string   `json:"repoName"`
	Statuses       []string `json:"statuses"`
	Assignee       string   `json:"assignee"`
	Creator        string   `json:"creator"`
	IssueKey       string   `json:"issueKey,omitempty"`
	AssigneeID     int64    `json:"assigneeId,omitempty"`
	CreatedUserID  int64    `json:"createdUserId,omitempty"`
	IssueID        int64    `json:"issueId,omitempty"`
	WorkflowID     string   `json:"workflowId"`
	WorkflowStepID string   `json:"workflowStepId,omitempty"`
	State          string   `json:"state"`
	CreatedCount   int      `json:"createdCount"`
	PendingCount   int      `json:"pendingCount"`
	LastRunAt      string   `json:"lastRunAt,omitempty"`
}

// LedgerEntry records that a watch reserved (and then created) the task of
// one pull request, so it is created at most once (AC6.2.3).
type LedgerEntry struct {
	Key     string `json:"key"`              // LinkKey
	TaskID  string `json:"taskId,omitempty"` // empty while only reserved
	WatchID string `json:"watchId"`
}

// Query is a saved pull request query (US6.3).
type Query = QueryInput

// StateStore is the part of Kandev's state store U4 needs. pluginsdk.Host satisfies it.
type StateStore interface {
	GetState(ctx context.Context, scope, scopeID, key string) (map[string]any, bool, error)
	SetState(ctx context.Context, scope, scopeID, key string, value map[string]any) error
}

// Store keeps U4's workspace documents. Each is one state value
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

type doc[T any] struct {
	SchemaVersion int `json:"schemaVersion"`
	Items         []T `json:"items"`
}

// Links returns the workspace's PR links.
func (s *Store) Links(ctx context.Context, ws string) ([]Link, error) {
	return load[Link](ctx, s, "workspace", ws, keyLinks)
}

// Watches returns the workspace's PR watches.
func (s *Store) Watches(ctx context.Context, ws string) ([]Watch, error) {
	return load[Watch](ctx, s, "workspace", ws, keyWatches)
}

// Ledger returns the workspace's watch ledger.
func (s *Store) Ledger(ctx context.Context, ws string) ([]LedgerEntry, error) {
	return load[LedgerEntry](ctx, s, "workspace", ws, keyLedger)
}

// Queries returns the workspace's saved queries.
func (s *Store) Queries(ctx context.Context, ws string) ([]Query, error) {
	return load[Query](ctx, s, "workspace", ws, keyQueries)
}

// Index lists the workspaces that have links or watches.
func (s *Store) Index(ctx context.Context) ([]string, error) {
	return load[string](ctx, s, "instance", "", keyIndex)
}

// UpdateLinks replaces the links with fn's result.
func (s *Store) UpdateLinks(ctx context.Context, ws string, fn func([]Link) ([]Link, error)) error {
	return update(ctx, s, ws, keyLinks, maxLinks, true, fn)
}

// UpdateWatches replaces the watches with fn's result.
func (s *Store) UpdateWatches(ctx context.Context, ws string, fn func([]Watch) ([]Watch, error)) error {
	return update(ctx, s, ws, keyWatches, maxWatches, true, fn)
}

// UpdateLedger replaces the ledger with fn's result.
// ponytail: the ledger is never pruned; prune entries of deleted watches if it grows large.
func (s *Store) UpdateLedger(ctx context.Context, ws string, fn func([]LedgerEntry) ([]LedgerEntry, error)) error {
	return update(ctx, s, ws, keyLedger, 0, false, fn)
}

// UpdateQueries replaces the saved queries with fn's result.
func (s *Store) UpdateQueries(ctx context.Context, ws string, fn func([]Query) ([]Query, error)) error {
	return update(ctx, s, ws, keyQueries, maxQueries, false, fn)
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
// connection.ErrStore (code internal), still matching DeadlineExceeded.
func storeErr(ctx context.Context, op string, err error) error {
	if errors.Is(ctx.Err(), context.Canceled) {
		return ctx.Err()
	}
	return fmt.Errorf("%s: %w: %w", op, connection.ErrStore, err)
}
