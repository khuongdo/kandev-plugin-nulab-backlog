package issues

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"sync"
	"sync/atomic"
	"time"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/backlog"
	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/connection"
	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/redact"
)

// refreshTimeout is the most issues.refresh waits for its cycle.
const refreshTimeout = 10 * time.Second

// Syncer is U3's status sync (ADR-003): one goroutine ticks every Tick (the
// 1-minute minimum interval) and runs each workspace whose interval has
// passed since its stored lastCycleAt, so a restart keeps the schedule
// (AC4.1.5). Refresh requests run on the same goroutine, so two polls never
// overlap (AC4.2.4).
// ponytail: one Issue GET per linked issue per cycle; switch to id[] batches
// when workspaces exceed about 200 links.
type Syncer struct {
	svc  *Service
	Tick time.Duration
	Log  *slog.Logger // set before Start

	refresh  chan refreshRequest
	stop     chan struct{}
	done     chan struct{}
	stopOnce sync.Once
}

type refreshRequest struct {
	ws    string
	reply chan refreshReply
}

type refreshReply struct {
	res RefreshResult
	err error
}

// RefreshResult is the issues.refresh reply.
type RefreshResult struct {
	UpdatedCount int    `json:"updatedCount"`
	RefreshedAt  string `json:"refreshedAt"`
}

// NewSyncer returns a stopped syncer with the 1-minute tick.
func NewSyncer(svc *Service, log *slog.Logger) *Syncer {
	if log == nil {
		log = slog.New(slog.DiscardHandler)
	}
	return &Syncer{svc: svc, Tick: time.Minute, Log: log, refresh: make(chan refreshRequest)}
}

// Start runs the worker goroutine.
func (y *Syncer) Start() {
	y.stop, y.done = make(chan struct{}), make(chan struct{})
	go y.loop()
}

// Stop ends the worker and waits for it. Safe to call more than once.
func (y *Syncer) Stop() {
	if y.stop == nil {
		return
	}
	y.stopOnce.Do(func() { close(y.stop) })
	<-y.done
}

var errStopped = errors.New("the issue sync is not running")

// Refresh runs one cycle of the workspace now and waits for it, at most 10 s.
func (y *Syncer) Refresh(ctx context.Context, ws string) (RefreshResult, error) {
	ctx, cancel := context.WithTimeout(ctx, refreshTimeout)
	defer cancel()
	req := refreshRequest{ws: ws, reply: make(chan refreshReply, 1)}
	select {
	case y.refresh <- req:
	case <-y.done:
		return RefreshResult{}, errStopped
	case <-ctx.Done():
		return RefreshResult{}, fmt.Errorf("refresh: %w", ctx.Err())
	}
	select {
	case r := <-req.reply:
		return r.res, r.err
	case <-ctx.Done():
		return RefreshResult{}, fmt.Errorf("refresh: %w", ctx.Err())
	}
}

func (y *Syncer) loop() {
	defer close(y.done)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		<-y.stop
		cancel() // stop a cycle waiting on Backlog
	}()
	ticker := time.NewTicker(y.Tick)
	defer ticker.Stop()
	for {
		select {
		case <-y.stop:
			return
		case <-ticker.C:
			y.cycleDue(ctx)
		case req := <-y.refresh:
			st, start := &cycleStats{workspaces: 1}, time.Now()
			n, err := y.svc.syncWorkspace(y.cycleCtx(ctx, st), req.ws, st)
			st.updated = n
			y.logCycle(ctx, st, start)
			req.reply <- refreshReply{res: RefreshResult{UpdatedCount: n, RefreshedAt: y.svc.now()}, err: err}
		}
	}
}

// cycleStats is what one issue_sync_cycle line reports.
type cycleStats struct {
	workspaces, updated int
	errors              atomic.Int64
	missing             atomic.Int64 // links whose task is not in the task list
	waits               atomic.Int64
}

// cycleDue reconciles every workspace with the connection, then runs each
// workspace that is due.
func (y *Syncer) cycleDue(ctx context.Context) {
	st, start := &cycleStats{}, time.Now()
	ctx = y.cycleCtx(ctx, st)
	// Events missed while the plugin was stopped are caught up here (the
	// Host is only injected after the runtime is built).
	if err := y.svc.ReconcileAll(ctx); err != nil {
		st.errors.Add(1)
	}
	idx, err := y.svc.store.Index(ctx)
	if err != nil {
		st.errors.Add(1)
	}
	for _, ws := range idx {
		set, err := y.svc.store.Settings(ctx, ws)
		if err != nil {
			st.errors.Add(1)
			continue
		}
		if !y.svc.due(set) {
			continue
		}
		st.workspaces++
		n, err := y.svc.syncWorkspace(ctx, ws, st)
		st.updated += n
		if err != nil {
			st.errors.Add(1)
		}
	}
	if st.workspaces > 0 || st.errors.Load() > 0 {
		y.logCycle(ctx, st, start)
	}
}

// due reports whether the workspace's interval has passed since its last cycle.
func (s *Service) due(set Settings) bool {
	last, err := time.Parse(time.RFC3339, set.LastCycleAt)
	if err != nil {
		return true
	}
	return !s.Now().Before(last.Add(time.Duration(set.PollMinutes) * time.Minute))
}

// cycleCtx puts the syncer's logger on ctx, wrapped so backlog_wait lines
// (written by backlog.Client while it waits out a 429) are counted.
func (y *Syncer) cycleCtx(ctx context.Context, st *cycleStats) context.Context {
	return redact.WithLogger(ctx, slog.New(waitCounter{Handler: y.Log.Handler(), n: &st.waits}))
}

func (y *Syncer) logCycle(ctx context.Context, st *cycleStats, start time.Time) {
	y.Log.InfoContext(ctx, "issue sync cycle", "event", "issue_sync_cycle", "workspaceCount", st.workspaces,
		"updated", st.updated, "errors", st.errors.Load(), "durationMs", time.Since(start).Milliseconds(),
		"rateLimitWaits", st.waits.Load(), "missingTasks", st.missing.Load())
}

// waitCounter counts event=backlog_wait records and passes every record on.
// ponytail: copied from internal/git/watcher.go; share it when a third worker needs it.
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

type checked struct {
	issue backlog.Issue
	err   error
}

// syncWorkspace runs one cycle of a workspace and returns how many links
// changed status. With the switch off or no connection it makes no request
// (AC1.5.4). It lists the workspace's tasks once to refresh task keys,
// reads each linked issue once
// (Background), and drops everything if the connection epoch moved
// meanwhile (AC1.8.3). A 404 or 403 marks a link unavailable but keeps it
// (AC4.1.2); other failures count towards "may be out of date" (AC4.1.4).
//
// The cycle never deletes a link (review 1, R-01/R-02): the task list misses
// config tasks and can skip entries across offset pages, and a link may be
// written while the cycle waits on Backlog. Kandev reports a deleted task to
// Tasks().Get only as a gRPC NotFound, and reading that code would add grpc
// as a direct go.mod requirement, so a task missing from the list is only
// skipped and counted (missingTasks); task.deleted removes links (AC2.3.2).
func (s *Service) syncWorkspace(ctx context.Context, ws string, st *cycleStats) (int, error) {
	if err := s.conn.RequireEnabled(ctx, ws); err != nil {
		if errors.Is(err, connection.ErrIntegrationDisabled) {
			return 0, nil
		}
		return 0, err
	}
	snap, err := s.conn.Current(ctx, ws)
	switch {
	case errors.Is(err, connection.ErrNotConnected), errors.Is(err, connection.ErrReconnectRequired):
		return 0, nil
	case err != nil:
		return 0, err
	}
	tasks, taskErr := s.host.ListTasks(ctx, ws)
	if taskErr != nil {
		st.errors.Add(1) // keep every link: the task list is unknown
	}
	exists := map[string]string{} // task id -> key
	for _, t := range tasks {
		exists[t.ID] = t.Key
	}
	missing := func(l Link) bool { _, ok := exists[l.TaskID]; return taskErr == nil && !ok }
	polled := func(l Link) bool { return l.State == StateActive && l.SpaceHost == snap.SpaceHost && !missing(l) }

	links, err := s.store.Links(ctx, ws)
	if err != nil {
		return 0, err
	}
	for _, l := range links {
		if missing(l) {
			st.missing.Add(1)
		}
	}
	results := map[string]checked{}
	if slices.ContainsFunc(links, polled) {
		cctx, creds, err := s.credentials(ctx, ws)
		if err != nil {
			return 0, err
		}
		for _, l := range links {
			if _, done := results[l.IssueKey]; done || !polled(l) {
				continue
			}
			i, err := s.gateway.Issue(cctx, creds, backlog.Background, l.IssueKey)
			if errors.Is(err, context.Canceled) {
				return 0, err
			}
			if err != nil && !unavailable(err) {
				st.errors.Add(1)
			}
			results[l.IssueKey] = checked{issue: i, err: err}
		}
	}
	if s.unchanged(ctx, ws, snap.ConnectionEpoch) != nil {
		return 0, nil // a late result is dropped
	}
	now := s.now()
	updated := 0
	err = s.store.UpdateLinks(ctx, ws, func(links []Link) ([]Link, error) {
		next := links[:0]
		for _, l := range links {
			if key := exists[l.TaskID]; key != "" {
				l.TaskKey = key
			}
			if r, ok := results[l.IssueKey]; ok && polled(l) {
				switch {
				case r.err == nil:
					if r.issue.StatusName != l.LastKnownStatus {
						updated++
						l.LastKnownStatus = r.issue.StatusName
					}
					l.StatusUpdatedAt, l.FailCount, l.Unavailable = now, 0, false
				case unavailable(r.err):
					l.Unavailable = true
				default:
					l.FailCount++
				}
			}
			next = append(next, l)
		}
		return next, nil
	})
	if err != nil {
		return 0, err
	}
	return updated, s.store.UpdateSettings(ctx, ws, func(set Settings) (Settings, error) {
		set.LastCycleAt = now
		return set, nil
	})
}
