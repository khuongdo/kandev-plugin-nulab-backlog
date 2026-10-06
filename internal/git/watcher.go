package git

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/backlog"
	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/redact"
)

const (
	// perCycle is the most tasks one watch creates per cycle (AC6.2.2).
	perCycle = 10
	// watchPRs is how many pull requests one watch reads per cycle.
	watchPRs = 100
)

// Watcher is U4's own timer (ADR-003): one goroutine runs every watch each
// Every, and git.watches.run requests, one at a time, so a Run racing a tick
// never creates a task twice (AC6.2.4).
// ponytail: one worker for all workspaces; a long 429 wait delays other
// workspaces' cycles, shard per workspace if that matters.
type Watcher struct {
	svc   *Service
	Every time.Duration // 5 minutes (FR4.2 default)
	Log   *slog.Logger  // set before Start

	runs     chan runRequest
	stop     chan struct{}
	done     chan struct{}
	stopOnce sync.Once
}

type runRequest struct{ ws, id string }

// NewWatcher returns a stopped watcher with the 5-minute interval.
func NewWatcher(svc *Service, log *slog.Logger) *Watcher {
	if log == nil {
		log = slog.New(slog.DiscardHandler)
	}
	return &Watcher{svc: svc, Every: 5 * time.Minute, Log: log, runs: make(chan runRequest, 64)}
}

// Start runs the worker goroutine.
func (w *Watcher) Start() {
	w.stop, w.done = make(chan struct{}), make(chan struct{})
	go w.loop()
}

// Stop ends the worker and waits for it. Safe to call more than once.
func (w *Watcher) Stop() {
	if w.stop == nil {
		return
	}
	w.stopOnce.Do(func() { close(w.stop) })
	<-w.done
}

func (w *Watcher) loop() {
	defer close(w.done)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		<-w.stop
		cancel() // stop a cycle waiting on Backlog
	}()
	ticker := time.NewTicker(w.Every)
	defer ticker.Stop()
	for {
		select {
		case <-w.stop:
			return
		case <-ticker.C:
			w.cycleAll(ctx)
		case r := <-w.runs:
			w.runOne(ctx, r)
		}
	}
}

// Run queues one cycle of an active watch and returns at once (queued),
// because ten task creations can exceed the 15 s action limit.
func (w *Watcher) Run(ctx context.Context, ws, id string) error {
	watch, err := w.svc.findWatch(ctx, ws, id)
	if err != nil {
		return err
	}
	if watch.State != StatusActive {
		return ErrConflict
	}
	select {
	case w.runs <- runRequest{ws: ws, id: id}:
	default: // the queue is full of runs already; the next tick covers it
	}
	return nil
}

// cycleStats is what one watch_cycle line reports.
type cycleStats struct {
	watches, created, errors int
	waits                    atomic.Int64
}

func (w *Watcher) cycleAll(ctx context.Context) {
	st, start := &cycleStats{}, time.Now()
	ctx = w.cycleCtx(ctx, st)
	// Each cycle first applies the current connection to every item, so
	// events missed while the plugin was stopped are caught up (startup
	// reconciliation; the Host is only injected after NewRuntime returns).
	if err := w.svc.ReconcileAll(ctx); err != nil {
		st.errors++
	}
	idx, err := w.svc.store.Index(ctx)
	if err != nil {
		st.errors++
	}
	for _, ws := range idx {
		watches, err := w.svc.store.Watches(ctx, ws)
		if err != nil {
			st.errors++
			continue
		}
		for _, watch := range watches {
			if watch.State == StatusActive {
				st.watches++
				w.svc.runWatch(ctx, ws, watch.ID, st)
			}
		}
	}
	w.logCycle(ctx, st, start)
}

func (w *Watcher) runOne(ctx context.Context, r runRequest) {
	st, start := &cycleStats{watches: 1}, time.Now()
	ctx = w.cycleCtx(ctx, st)
	w.svc.runWatch(ctx, r.ws, r.id, st)
	w.logCycle(ctx, st, start)
}

// cycleCtx puts the watcher's logger on ctx, wrapped so backlog_wait lines
// (written by backlog.Client while it waits out a 429) are counted.
func (w *Watcher) cycleCtx(ctx context.Context, st *cycleStats) context.Context {
	return redact.WithLogger(ctx, slog.New(waitCounter{Handler: w.Log.Handler(), n: &st.waits}))
}

func (w *Watcher) logCycle(ctx context.Context, st *cycleStats, start time.Time) {
	w.Log.InfoContext(ctx, "watch cycle", "event", "watch_cycle", "watchCount", st.watches, "created", st.created,
		"errors", st.errors, "durationMs", time.Since(start).Milliseconds(), "rateLimitWaits", st.waits.Load())
}

// waitCounter counts event=backlog_wait records and passes every record on.
type waitCounter struct {
	slog.Handler
	n *atomic.Int64
}

func (h waitCounter) Handle(ctx context.Context, r slog.Record) error {
	r.Attrs(func(a slog.Attr) bool {
		if a.Key == "event" && a.Value.String() == "backlog_wait" {
			h.n.Add(1)
		}
		return true
	})
	return h.Handler.Handle(ctx, r)
}

func (h waitCounter) WithAttrs(attrs []slog.Attr) slog.Handler {
	return waitCounter{Handler: h.Handler.WithAttrs(attrs), n: h.n}
}

func (h waitCounter) WithGroup(name string) slog.Handler {
	return waitCounter{Handler: h.Handler.WithGroup(name), n: h.n}
}

// runWatch runs one watch once and counts its outcome.
func (s *Service) runWatch(ctx context.Context, ws, id string, st *cycleStats) {
	created, err := s.cycleWatch(ctx, ws, id)
	st.created += created
	if err != nil {
		st.errors++
		redact.Logger(ctx).WarnContext(ctx, "watch failed", "event", "watch_failed", "workspaceId", ws,
			"watchId", id, "errorCode", errorCode(err))
	}
}

// cycleWatch creates the tasks of one watch, oldest pull request first, at
// most perCycle. Each task is reserved in the ledger, created with its
// LinkKey in the task metadata, then its id is stored, so it is created at
// most once even across a crash (AC6.2.3). A reserved entry without a task
// id is found again by its metadata. If the connection epoch moves, the
// cycle writes nothing more (AC1.8.3).
func (s *Service) cycleWatch(ctx context.Context, ws, id string) (int, error) {
	watch, err := s.findWatch(ctx, ws, id)
	if err != nil || watch.State != StatusActive {
		return 0, nil // deleted or paused meanwhile
	}
	snap, err := s.conn.Current(ctx, ws)
	if err != nil {
		return 0, err
	}
	if watch.SpaceHost != snap.SpaceHost || !slices.Contains(snap.SelectedProjects, watch.ProjectKey) {
		return 0, nil // ConnectionChanged turns it off
	}
	ctx, creds, err := s.credentials(ctx, ws)
	if err != nil {
		return 0, err
	}
	prs, err := s.gateway.PullRequests(ctx, creds, backlog.Background, watch.ProjectKey, watch.RepoName, watchQuery(watch))
	if err != nil {
		return 0, err
	}
	sort.Slice(prs, func(i, j int) bool { return prs[i].Number < prs[j].Number })
	if s.unchanged(ctx, ws, snap.ConnectionEpoch) != nil {
		return 0, nil
	}
	ledger, err := s.store.Ledger(ctx, ws)
	if err != nil {
		return 0, err
	}
	created, todo, err := s.resolveReserved(ctx, ws, snap.SpaceHost, watch, prs, ledger)
	if err != nil {
		return created, err
	}
	var failed error
	n := 0
	for _, pr := range todo {
		if n == perCycle {
			break
		}
		if failed = s.createOne(ctx, ws, snap.SpaceHost, snap.ConnectionEpoch, watch, pr); failed != nil {
			break
		}
		n++
	}
	if errors.Is(failed, ErrStale) {
		failed = nil // a late result is dropped, not an error
	}
	err = s.store.UpdateWatches(ctx, ws, func(list []Watch) ([]Watch, error) {
		i := slices.IndexFunc(list, func(w Watch) bool { return w.ID == watch.ID })
		if i < 0 {
			return nil, errUnchanged
		}
		list[i].CreatedCount += created + n
		list[i].PendingCount = len(todo) - n
		list[i].LastRunAt = s.Now().UTC().Format(time.RFC3339)
		return list, nil
	})
	if failed == nil {
		failed = err
	}
	return created + n, failed
}

// resolveReserved finishes reserved entries whose task already exists and
// returns the pull requests still to create.
func (s *Service) resolveReserved(ctx context.Context, ws, host string, watch Watch, prs []backlog.PullRequest, ledger []LedgerEntry) (int, []backlog.PullRequest, error) {
	var todo []backlog.PullRequest
	resolved := 0
	for _, pr := range prs {
		key := LinkKey(host, pr.RepositoryID, pr.Number)
		i := slices.IndexFunc(ledger, func(e LedgerEntry) bool { return e.Key == key })
		switch {
		case i < 0:
			todo = append(todo, pr)
		case ledger[i].TaskID == "":
			taskID, found, err := s.host.FindTaskByMetadata(ctx, ws, MetadataKey, key)
			if err != nil {
				return resolved, nil, err
			}
			if !found {
				todo = append(todo, pr)
				continue
			}
			if err := s.finish(ctx, ws, host, watch, pr, key, taskID); err != nil {
				return resolved, nil, err
			}
			resolved++
		}
	}
	return resolved, todo, nil
}

func (s *Service) createOne(ctx context.Context, ws, host string, epoch int, watch Watch, pr backlog.PullRequest) error {
	if err := s.unchanged(ctx, ws, epoch); err != nil {
		return err
	}
	key := LinkKey(host, pr.RepositoryID, pr.Number)
	err := s.store.UpdateLedger(ctx, ws, func(l []LedgerEntry) ([]LedgerEntry, error) {
		if slices.ContainsFunc(l, func(e LedgerEntry) bool { return e.Key == key }) {
			return nil, errUnchanged
		}
		return append(l, LedgerEntry{Key: key, WatchID: watch.ID}), nil
	})
	if err != nil {
		return err
	}
	taskID, err := s.host.CreateTask(ctx, NewTask{
		WorkspaceID: ws, WorkflowID: watch.WorkflowID, WorkflowStepID: watch.WorkflowStepID,
		Title:       fmt.Sprintf("Review PR #%d: %s", pr.Number, pr.Summary),
		Description: prURL(host, watch.ProjectKey, watch.RepoName, pr.Number),
		Metadata:    map[string]any{MetadataKey: key},
	})
	if err != nil {
		return err
	}
	return s.finish(ctx, ws, host, watch, pr, key, taskID)
}

// finish stores the task id in the ledger and links the task to the PR.
func (s *Service) finish(ctx context.Context, ws, host string, watch Watch, pr backlog.PullRequest, key, taskID string) error {
	err := s.store.UpdateLedger(ctx, ws, func(l []LedgerEntry) ([]LedgerEntry, error) {
		i := slices.IndexFunc(l, func(e LedgerEntry) bool { return e.Key == key })
		if i < 0 {
			return append(l, LedgerEntry{Key: key, TaskID: taskID, WatchID: watch.ID}), nil
		}
		l[i].TaskID = taskID
		return l, nil
	})
	if err != nil {
		return err
	}
	return s.putLink(ctx, ws, Link{TaskID: taskID, SpaceHost: host, ProjectKey: watch.ProjectKey, RepoName: watch.RepoName,
		RepositoryID: pr.RepositoryID, Number: pr.Number, Title: pr.Summary, Status: StatusActive})
}

func watchQuery(w Watch) backlog.PullRequestQuery {
	q := backlog.PullRequestQuery{StatusIDs: statusIDs(w.Statuses), Count: watchPRs}
	if w.AssigneeID != 0 {
		q.AssigneeIDs = []int64{w.AssigneeID}
	}
	if w.CreatedUserID != 0 {
		q.CreatedUserIDs = []int64{w.CreatedUserID}
	}
	if w.IssueID != 0 {
		q.IssueIDs = []int64{w.IssueID}
	}
	return q
}
