package issues

import (
	"context"
	"errors"
	"log/slog"
	"slices"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/backlog"
	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/connection"
	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/redact"
)

const (
	watchPageSize = 100 // issues per page of a run (BR3.12)
	watchMaxPages = 5   // pages per run at most (BR3.12)
	overshootBack = 5   // a run starts this many issues before the cursor (BR3.12)
)

// Watcher runs the issue watches (WF5): one goroutine ticks every Tick (1
// minute) and runs each active watch whose own interval has passed, and
// issues.watches.run requests, one at a time, so a run never overlaps
// another run of the same watch. Its store calls wait (bounded) for the
// Kandev host, and a failed tick is retried at the next one (NFR6).
// ponytail: one worker for all workspaces, like the PR watcher; a long 429
// wait delays other watches.
type Watcher struct {
	svc  *Service
	Tick time.Duration
	Log  *slog.Logger // set before Start

	runs     chan watchRequest
	stop     chan struct{}
	done     chan struct{}
	stopOnce sync.Once
}

type watchRequest struct{ ws, id string }

// NewWatcher returns a stopped watcher with the 1-minute tick.
func NewWatcher(svc *Service, log *slog.Logger) *Watcher {
	if log == nil {
		log = slog.New(slog.DiscardHandler)
	}
	return &Watcher{svc: svc, Tick: time.Minute, Log: log, runs: make(chan watchRequest, 64)}
}

// Start runs the worker goroutine.
func (y *Watcher) Start() {
	y.stop, y.done = make(chan struct{}), make(chan struct{})
	go y.loop()
}

// Stop ends the worker and waits for it. Safe to call more than once.
func (y *Watcher) Stop() {
	if y.stop == nil {
		return
	}
	y.stopOnce.Do(func() { close(y.stop) })
	<-y.done
}

// Run queues one run of an active watch and returns at once (BR3.9).
func (y *Watcher) Run(ctx context.Context, ws, id string) error {
	w, err := y.svc.findWatch(ctx, ws, id)
	if err != nil {
		return err
	}
	if w.State != StateActive {
		return ErrConflict
	}
	select {
	case y.runs <- watchRequest{ws: ws, id: id}:
	default: // the queue is full of runs already; the next tick covers it
	}
	return nil
}

func (y *Watcher) loop() {
	defer close(y.done)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		<-y.stop
		cancel() // stop a run waiting on Backlog
	}()
	ticker := time.NewTicker(y.Tick)
	defer ticker.Stop()
	for {
		select {
		case <-y.stop:
			return
		case <-ticker.C:
			y.cycle(ctx)
		case r := <-y.runs:
			st, start := &watchStats{watches: 1}, time.Now()
			y.run(y.statsCtx(ctx, st), r.ws, r.id, st)
			y.logCycle(ctx, st, start)
		}
	}
}

// watchStats is what one issue_watch_cycle line reports.
type watchStats struct {
	watches, created, skipped, errors int
	waits                             atomic.Int64
}

// cycle applies the connection to every workspace's watches, then runs
// the active watches that are due.
func (y *Watcher) cycle(ctx context.Context) {
	st, start := &watchStats{}, time.Now()
	ctx = y.statsCtx(ctx, st)
	idx, err := y.svc.store.WatchIndex(ctx)
	if err != nil {
		st.errors++
	}
	for _, ws := range idx {
		if err := y.svc.reconcileWatches(ctx, ws); err != nil {
			st.errors++
			continue
		}
		list, err := y.svc.store.Watches(ctx, ws)
		if err != nil {
			st.errors++
			continue
		}
		for _, w := range list {
			if w.State == StateActive && y.svc.watchDue(w) {
				st.watches++
				y.run(ctx, ws, w.ID, st)
			}
		}
	}
	if st.watches > 0 || st.errors > 0 {
		y.logCycle(ctx, st, start)
	}
}

func (y *Watcher) run(ctx context.Context, ws, id string, st *watchStats) {
	res, err := y.svc.runWatch(ctx, ws, id)
	st.created += res.created
	st.skipped += res.skipped
	if err != nil {
		st.errors++
		redact.Logger(ctx).WarnContext(ctx, "issue watch failed", "event", "issue_watch_failed", "workspaceId", ws,
			"watchId", id, "errorCode", lastErrorCode(err))
	}
}

// statsCtx puts the watcher's logger on ctx, counting backlog_wait lines.
func (y *Watcher) statsCtx(ctx context.Context, st *watchStats) context.Context {
	return redact.WithLogger(ctx, slog.New(waitCounter{Handler: y.Log.Handler(), n: &st.waits}))
}

func (y *Watcher) logCycle(ctx context.Context, st *watchStats, start time.Time) {
	y.Log.InfoContext(ctx, "issue watch cycle", "event", "issue_watch_cycle", "watchCount", st.watches,
		"created", st.created, "skipped", st.skipped, "errors", st.errors,
		"durationMs", time.Since(start).Milliseconds(), "rateLimitWaits", st.waits.Load())
}

// reconcileWatches applies the switch and the connection to a workspace's
// watches: off or not connected turns them not_connected (BR3.9).
func (s *Service) reconcileWatches(ctx context.Context, ws string) error {
	err := s.conn.RequireEnabled(ctx, ws)
	switch {
	case errors.Is(err, connection.ErrIntegrationDisabled):
		return s.applyWatches(ctx, ws, nil, false)
	case err != nil:
		return err
	}
	snap, err := s.conn.Current(ctx, ws)
	switch {
	case errors.Is(err, connection.ErrNotConnected), errors.Is(err, connection.ErrReconnectRequired):
		return s.applyWatches(ctx, ws, nil, false)
	case err != nil:
		return err
	}
	return s.applyWatches(ctx, ws, &target{host: snap.SpaceHost, selected: snap.SelectedProjects}, true)
}

// watchDue reports whether the watch never ran or its interval passed (BR3.3).
func (s *Service) watchDue(w IssueWatch) bool {
	last, err := time.Parse(time.RFC3339, w.LastRunAt)
	if err != nil {
		return true
	}
	return !s.Now().Before(last.Add(time.Duration(w.IntervalMinutes) * time.Minute))
}

// lastErrorCode maps a run failure to the watch's lastError (BR3.11-BR3.14).
func lastErrorCode(err error) string {
	var be *backlog.Error
	switch {
	case errors.Is(err, ErrLedgerFull):
		return ErrorLedgerFull
	case errors.Is(err, ErrWorkflowMissing):
		return ErrorWorkflowMissing
	case errors.Is(err, connection.ErrReconnectRequired), errors.As(err, &be) && be.Kind == backlog.KindUnauthorized:
		return ErrorUnauthorized
	case errors.As(err, &be) && be.Kind == backlog.KindRateLimited:
		return ErrorRateLimited
	}
	return ErrorUnavailable
}

type runResult struct{ created, skipped, pending int }

// runIssueWatch runs one watch once and returns how many tasks it created.
func (s *Service) runIssueWatch(ctx context.Context, ws, id string) (int, error) {
	res, err := s.runWatch(ctx, ws, id)
	return res.created, err
}

// runWatch runs one active watch (WF5 step 2): it walks the matching issues
// after the cursor, oldest created first, and creates at most one task
// (BR3.4). The outcome is stored on the watch; a late result of an old
// connection epoch is dropped (AC1.8.3). With the switch off it does nothing.
func (s *Service) runWatch(ctx context.Context, ws, id string) (runResult, error) {
	if err := s.conn.RequireEnabled(ctx, ws); err != nil {
		if errors.Is(err, connection.ErrIntegrationDisabled) {
			return runResult{}, nil
		}
		return runResult{}, err
	}
	w, err := s.findWatch(ctx, ws, id)
	if errors.Is(err, ErrNotFound) || (err == nil && w.State != StateActive) {
		return runResult{}, nil // deleted or paused meanwhile
	}
	if err != nil {
		return runResult{}, err
	}
	snap, err := s.conn.Current(ctx, ws)
	if err != nil {
		return runResult{}, err
	}
	if w.SpaceHost != snap.SpaceHost || !slices.Contains(snap.SelectedProjects, w.ProjectKey) {
		return runResult{}, nil // reconciliation turns it not_connected
	}
	r := &watchRun{s: s, ws: ws, w: w, snap: snap, cursor: w.Cursor}
	runErr := r.do(ctx)
	if errors.Is(runErr, context.Canceled) {
		return r.res, runErr
	}
	if errors.Is(runErr, ErrStale) {
		runErr = nil
	}
	if err := s.finishRun(ctx, ws, w, r, runErr); err != nil && runErr == nil {
		runErr = err
	}
	return r.res, runErr
}

// finishRun stores the run's cursor, counts and error on the watch, unless
// its filters were edited meanwhile (then only the counts and times).
func (s *Service) finishRun(ctx context.Context, ws string, w IssueWatch, r *watchRun, runErr error) error {
	code := ""
	if runErr != nil {
		code = lastErrorCode(runErr)
	}
	return s.store.UpdateWatches(ctx, ws, func(list []IssueWatch) ([]IssueWatch, error) {
		i := slices.IndexFunc(list, func(x IssueWatch) bool { return x.ID == w.ID })
		if i < 0 {
			return nil, errUnchanged
		}
		if sameFilters(list[i], w) {
			list[i].Cursor = r.cursor
		}
		list[i].CreatedCount += r.res.created
		list[i].PendingCount = r.res.pending
		list[i].LastRunAt = s.now()
		list[i].LastError = code
		return list, nil
	})
}

// watchRun is one run of one watch.
type watchRun struct {
	s      *Service
	ws     string
	w      IssueWatch
	snap   connection.Snapshot
	creds  backlog.Credentials
	cursor *IssueCursor
	links  []Link
	ledger map[string]bool // keys already in the ledger
	res    runResult
}

func (r *watchRun) key(issueID int64) string {
	return r.snap.SpaceHost + "|" + strconv.FormatInt(issueID, 10)
}

func (r *watchRun) do(ctx context.Context) error {
	ctx, creds, err := r.s.credentials(ctx, r.ws)
	if err != nil {
		return err
	}
	r.creds = creds
	if r.links, err = r.s.store.Links(ctx, r.ws); err != nil {
		return err
	}
	if err := r.resolveReserved(ctx); err != nil {
		return err
	}
	if len(r.ledger) >= maxLedgerEntries {
		return ErrLedgerFull
	}
	projects, err := r.s.projectsFor(ctx, r.ws, r.snap.ConnectionEpoch, creds)
	if err != nil {
		return err
	}
	p, ok := projects[r.w.ProjectKey]
	if !ok {
		return nil // the project is gone from the space
	}
	return r.walk(ctx, r.query(p.ID))
}

func (r *watchRun) query(projectID int64) backlog.IssueQuery {
	q := backlog.IssueQuery{ProjectIDs: []int64{projectID}, StatusIDs: r.w.StatusIDs, CreatedSince: r.cursor.CreatedSince(),
		Sort: "created", Order: "asc", Count: watchPageSize}
	if r.w.AssigneeID != 0 {
		q.AssigneeIDs = []int64{r.w.AssigneeID}
	}
	if r.w.CreatedUserID != 0 {
		q.CreatedUserIDs = []int64{r.w.CreatedUserID}
	}
	if r.cursor != nil {
		q.Offset = max(0, r.cursor.DayOffset-overshootBack)
	}
	return q
}

// walk reads at most 5 pages from near the cursor and handles each issue
// after it until one task is created (BR3.12). If the first page starts
// after the cursor, the start may have overshot because earlier issues
// left the list, so it reads once more from 0 (R-09).
func (r *watchRun) walk(ctx context.Context, q backlog.IssueQuery) error {
	var seen []time.Time
	for n := range watchMaxPages {
		page, err := r.s.gateway.Issues(ctx, r.creds, backlog.Background, q)
		if err != nil {
			return err
		}
		if n == 0 && q.Offset > 0 && (len(page) == 0 || r.cursor.After(createdAt(page[0]), page[0].ID)) {
			q.Offset = 0
			continue
		}
		for i, issue := range page {
			created := createdAt(issue)
			seen = append(seen, created)
			if !r.cursor.After(created, issue.ID) {
				continue
			}
			if r.res.created > 0 {
				r.res.pending++
				continue
			}
			if err := r.handle(ctx, issue); err != nil {
				return err
			}
			r.advance(issue.ID, created, q.Offset+i, q.CreatedSince, seen)
		}
		if len(page) < watchPageSize || r.res.created > 0 {
			return nil
		}
		q.Offset += len(page)
	}
	return nil
}

func createdAt(i backlog.Issue) time.Time {
	t, _ := time.Parse(time.RFC3339, i.Created) // Backlog always sends RFC 3339
	return t
}

// advance moves the cursor to the issue at position pos of the list for
// since. When the issue's own createdSince differs, its offset is counted
// over the issues seen this run, which can only undercount: a later run
// then starts a little earlier, never past an issue.
func (r *watchRun) advance(id int64, created time.Time, pos int, since string, seen []time.Time) {
	c := &IssueCursor{Created: created, IssueID: id, DayOffset: pos}
	if own := c.CreatedSince(); own != since {
		c.DayOffset = 0
		for _, t := range seen[:len(seen)-1] {
			if t.UTC().Format(time.DateOnly) >= own {
				c.DayOffset++
			}
		}
	}
	r.cursor = c
}

func (r *watchRun) linked(issue backlog.Issue) (Link, bool) {
	i := slices.IndexFunc(r.links, func(l Link) bool {
		return l.IssueKey == issue.IssueKey && l.SpaceHost == r.snap.SpaceHost && l.State == StateActive
	})
	if i < 0 {
		return Link{}, false
	}
	return r.links[i], true
}

// handle passes one issue after the cursor: an issue in the ledger or with
// a link is skipped (BR3.6); otherwise it is reserved, its task created
// and linked, then marked created (BR3.7).
func (r *watchRun) handle(ctx context.Context, issue backlog.Issue) error {
	key := r.key(issue.ID)
	if r.ledger[key] {
		r.res.skipped++
		return nil
	}
	if _, ok := r.linked(issue); ok {
		r.res.skipped++
		return r.addEntry(ctx, IssueWatchLedgerEntry{Key: key, Outcome: OutcomeSkippedLinked, At: r.s.now()})
	}
	if err := r.s.unchanged(ctx, r.ws, r.snap.ConnectionEpoch); err != nil {
		return err
	}
	if err := r.addEntry(ctx, IssueWatchLedgerEntry{Key: key, Outcome: OutcomeReserved, At: r.s.now()}); err != nil {
		return err
	}
	ref, err := r.createTask(ctx, issue)
	if ref.ID == "" {
		if rmErr := r.setEntry(ctx, key, nil); rmErr != nil {
			redact.Logger(ctx).ErrorContext(ctx, "issue watch reservation not removed", "event", "issue_watch_ledger_failed",
				"watchId", r.w.ID)
		}
		return err
	}
	r.res.created++
	if mErr := r.setEntry(ctx, key, &IssueWatchLedgerEntry{Key: key, Outcome: OutcomeCreated, TaskID: ref.ID, At: r.s.now()}); mErr != nil {
		return mErr
	}
	return err // a link that was not stored; the ledger still stops a second task
}

// createTask creates and links the issue's task in the watch's workflow,
// exactly like issues.create_task.
func (r *watchRun) createTask(ctx context.Context, issue backlog.Issue) (TaskRef, error) {
	return r.s.createLinkedTask(ctx, r.ws, r.snap, issue, r.w.WorkflowID, r.w.WorkflowStepID)
}

func (r *watchRun) addEntry(ctx context.Context, e IssueWatchLedgerEntry) error {
	err := r.s.store.UpdateLedger(ctx, r.ws, r.w.ID, func(l []IssueWatchLedgerEntry) ([]IssueWatchLedgerEntry, error) {
		return append(l, e), nil
	})
	if err == nil {
		r.ledger[e.Key] = true
	}
	return err
}

// setEntry replaces the entry with key by e, or removes it when e is nil.
func (r *watchRun) setEntry(ctx context.Context, key string, e *IssueWatchLedgerEntry) error {
	err := r.s.store.UpdateLedger(ctx, r.ws, r.w.ID, func(l []IssueWatchLedgerEntry) ([]IssueWatchLedgerEntry, error) {
		i := slices.IndexFunc(l, func(x IssueWatchLedgerEntry) bool { return x.Key == key })
		switch {
		case i < 0:
			return nil, errUnchanged
		case e == nil:
			return slices.Delete(l, i, i+1), nil
		}
		l[i] = *e
		return l, nil
	})
	if err == nil && e == nil {
		delete(r.ledger, key)
	}
	return err
}

// resolveReserved finishes reservations left without a task (BR3.7): one
// whose issue now has a link is marked created, the others are removed so
// the issue is tried again.
func (r *watchRun) resolveReserved(ctx context.Context) error {
	var kept []IssueWatchLedgerEntry
	err := r.s.store.UpdateLedger(ctx, r.ws, r.w.ID, func(l []IssueWatchLedgerEntry) ([]IssueWatchLedgerEntry, error) {
		kept = l
		if !slices.ContainsFunc(l, func(e IssueWatchLedgerEntry) bool { return e.Outcome == OutcomeReserved }) {
			return nil, errUnchanged
		}
		next := l[:0]
		for _, e := range l {
			if e.Outcome == OutcomeReserved {
				i := slices.IndexFunc(r.links, func(x Link) bool {
					return x.State == StateActive && r.key(x.IssueID) == e.Key
				})
				if i < 0 {
					continue
				}
				e.Outcome, e.TaskID, e.At = OutcomeCreated, r.links[i].TaskID, r.s.now()
			}
			next = append(next, e)
		}
		kept = next
		return next, nil
	})
	r.ledger = make(map[string]bool, len(kept))
	for _, e := range kept {
		r.ledger[e.Key] = true
	}
	return err
}
