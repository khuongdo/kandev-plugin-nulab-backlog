package issues

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
	keyLinks      = "issues.links"
	keySettings   = "issues.settings"
	keyIndex      = "issues.index" // instance key: workspaces with links
	maxLinks      = 1000
)

var errLimit = errors.New("the list is full")

// errUnchanged tells an update that fn changed nothing, so nothing is written.
var errUnchanged = errors.New("unchanged")

// Settings is the workspace's sync schedule (US4.2). It lives in
// IssueIntegration's own state: writing the Connection record would raise
// the connection epoch and drop in-flight results.
type Settings struct {
	PollMinutes int    `json:"pollMinutes"`
	LastCycleAt string `json:"lastCycleAt,omitempty"`
}

// StateStore is the part of Kandev's state store U3 needs. pluginsdk.Host satisfies it.
type StateStore interface {
	GetState(ctx context.Context, scope, scopeID, key string) (map[string]any, bool, error)
	SetState(ctx context.Context, scope, scopeID, key string, value map[string]any) error
}

// Store keeps U3's documents: per workspace issues.links {schemaVersion,
// items} and issues.settings {schemaVersion, pollMinutes, lastCycleAt}, and
// the instance index issues.index. A read-modify-write holds the
// workspace's mutex.
// ponytail: the document helpers are copied from internal/git/store.go, not
// shared; extract them when a third package needs them.
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

type listDoc[T any] struct {
	Items []T `json:"items"`
}

// Links returns the workspace's issue links.
func (s *Store) Links(ctx context.Context, ws string) ([]Link, error) {
	var d listDoc[Link]
	err := s.load(ctx, "workspace", ws, keyLinks, &d)
	return d.Items, err
}

// Settings returns the workspace's sync settings, with the 5-minute default.
func (s *Store) Settings(ctx context.Context, ws string) (Settings, error) {
	st := Settings{PollMinutes: DefaultPollMinutes}
	err := s.load(ctx, "workspace", ws, keySettings, &st)
	if err != nil {
		return Settings{}, err
	}
	return st, nil
}

// Index lists the workspaces that have links.
func (s *Store) Index(ctx context.Context) ([]string, error) {
	var d listDoc[string]
	err := s.load(ctx, "instance", "", keyIndex, &d)
	return d.Items, err
}

// UpdateLinks replaces the links with fn's result, at most 1,000.
func (s *Store) UpdateLinks(ctx context.Context, ws string, fn func([]Link) ([]Link, error)) error {
	defer s.lock("workspace/" + ws)()
	links, err := s.Links(ctx, ws)
	if err != nil {
		return err
	}
	next, err := fn(links)
	if errors.Is(err, errUnchanged) {
		return nil
	}
	if err != nil {
		return err
	}
	if len(next) > maxLinks {
		return &connection.FieldError{Field: FieldLimit, Err: errLimit}
	}
	if err := s.put(ctx, "workspace", ws, keyLinks, listDoc[Link]{Items: next}); err != nil {
		return err
	}
	if len(next) > 0 {
		return s.addIndex(ctx, ws)
	}
	return nil
}

// UpdateSettings replaces the settings with fn's result.
func (s *Store) UpdateSettings(ctx context.Context, ws string, fn func(Settings) (Settings, error)) error {
	defer s.lock("settings/" + ws)()
	st, err := s.Settings(ctx, ws)
	if err != nil {
		return err
	}
	next, err := fn(st)
	if errors.Is(err, errUnchanged) {
		return nil
	}
	if err != nil {
		return err
	}
	return s.put(ctx, "workspace", ws, keySettings, next)
}

func (s *Store) addIndex(ctx context.Context, ws string) error {
	defer s.lock("instance")()
	idx, err := s.Index(ctx)
	if err != nil || slices.Contains(idx, ws) {
		return err
	}
	return s.put(ctx, "instance", "", keyIndex, listDoc[string]{Items: append(idx, ws)})
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

// load decodes a document into v; a missing one leaves v unchanged. An
// unknown or broken document is an error and is never overwritten.
func (s *Store) load(ctx context.Context, scope, scopeID, key string, v any) error {
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
		return storeErr(ctx, "read "+key, err)
	}
	if !found {
		return nil
	}
	var version struct {
		SchemaVersion int `json:"schemaVersion"`
	}
	b, err := json.Marshal(value)
	if err == nil {
		err = json.Unmarshal(b, &version)
	}
	if err == nil && version.SchemaVersion == schemaVersion {
		err = json.Unmarshal(b, v)
	} else if err == nil {
		err = errors.New("unsupported schema")
	}
	if err != nil {
		return fmt.Errorf("read %s: %w: %w", key, connection.ErrStore, err)
	}
	return nil
}

func (s *Store) put(ctx context.Context, scope, scopeID, key string, v any) error {
	b, err := json.Marshal(v)
	var m map[string]any
	if err == nil {
		err = json.Unmarshal(b, &m)
	}
	if err == nil {
		m["schemaVersion"] = schemaVersion
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
