package scm

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"sort"
	"sync"
	"time"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/redact"
)

// refreshEvery is how often linked PR states are refreshed: the Backlog Git
// watch cycle (FR4.4).
const refreshEvery = 5 * time.Minute

// errTaken tells a run that another run reserved the pull request first.
var errTaken = errors.New("already reserved")

// RunWatch runs an active, mapped watch once now (FR4.3) and returns the
// number of tasks it created: at most one.
func (s *Service) RunWatch(ctx context.Context, ws, id string) (int, error) {
	list, err := s.ListWatches(ctx, ws)
	if err != nil {
		return 0, err
	}
	i := slices.IndexFunc(list, func(w Watch) bool { return w.ID == id })
	switch {
	case i < 0:
		return 0, ErrNotFound
	case list[i].State != WatchActive || list[i].Unmapped:
		return 0, ErrConflict
	}
	st, err := s.settings(ctx, ws, list[i].Provider)
	if err != nil {
		return 0, err
	}
	return s.runWatch(ctx, ws, list[i], st)
}

// runWatch creates the task of the oldest matching pull request that has
// none yet. The PR is reserved in the ledger first, so concurrent runs
// create it once; a failed creation frees the reservation for the next run.
// ponytail: a crash between reserving and creating loses that one task
// rather than risking a duplicate; recover by task metadata if that matters.
func (s *Service) runWatch(ctx context.Context, ws string, w Watch, st Settings) (int, error) {
	ctx, cred, err := s.credential(ctx, ws, w.Provider)
	if err != nil {
		return 0, err
	}
	res, err := s.clients[w.Provider].ListPRs(ctx, cred, w.Repo, ListQuery{States: w.Statuses, Page: 1, PerPage: watchPRs})
	if err != nil {
		return 0, err
	}
	prs := byAuthor(res.Items, w.Author, st.AccountID)
	s.autoLink(ctx, ws, st, w.Repo, prs)
	sort.Slice(prs, func(i, j int) bool { return prs[i].Number < prs[j].Number })
	ledger, err := s.store.Ledger(ctx, ws)
	if err != nil {
		return 0, err
	}
	created := 0
	for _, pr := range prs {
		if slices.ContainsFunc(ledger, func(e LedgerEntry) bool { return e.Key == pr.Key() }) {
			continue
		}
		err := s.createOne(ctx, ws, w, pr)
		if errors.Is(err, errTaken) {
			continue
		}
		if err != nil {
			return 0, err
		}
		created = 1
		break
	}
	return created, s.store.UpdateWatches(ctx, ws, func(list []Watch) ([]Watch, error) {
		i := slices.IndexFunc(list, func(x Watch) bool { return x.ID == w.ID })
		if i < 0 {
			return nil, errUnchanged
		}
		list[i].CreatedCount += created
		list[i].LastRunAt = s.Now().UTC().Format(time.RFC3339)
		return list, nil
	})
}

// createOne reserves pr, creates its task and links the task to it.
func (s *Service) createOne(ctx context.Context, ws string, w Watch, pr PullRequest) error {
	key := pr.Key()
	err := s.store.UpdateLedger(ctx, ws, func(l []LedgerEntry) ([]LedgerEntry, error) {
		if slices.ContainsFunc(l, func(e LedgerEntry) bool { return e.Key == key }) {
			return nil, errTaken
		}
		return append(l, LedgerEntry{Key: key, WatchID: w.ID}), nil
	})
	if err != nil {
		return err
	}
	taskID, err := s.tasks.CreateTask(ctx, NewTask{WorkspaceID: ws, WorkflowID: w.WorkflowID, WorkflowStepID: w.WorkflowStepID,
		Title: fmt.Sprintf("Review PR #%d: %s", pr.Number, pr.Title), Description: PRURL(pr.PRRef),
		Metadata: map[string]any{MetadataKey: key}})
	if err != nil {
		release := s.store.UpdateLedger(ctx, ws, func(l []LedgerEntry) ([]LedgerEntry, error) {
			return slices.DeleteFunc(l, func(e LedgerEntry) bool { return e.Key == key && e.TaskID == "" }), nil
		})
		return errors.Join(err, release)
	}
	err = s.store.UpdateLedger(ctx, ws, func(l []LedgerEntry) ([]LedgerEntry, error) {
		for i := range l {
			if l[i].Key == key {
				l[i].TaskID = taskID
			}
		}
		return l, nil
	})
	if err != nil {
		return err
	}
	return s.putLink(ctx, ws, Link{PRRef: pr.PRRef, TaskID: taskID, ProjectKey: w.ProjectKey, Title: pr.Title,
		State: pr.State, URL: PRURL(pr.PRRef)})
}

// RefreshLinks reads every linked pull request once and stores its state and
// title (FR4.4). A 429 stops only that provider for this cycle (NFR3); every
// failure is logged, never returned.
func (s *Service) RefreshLinks(ctx context.Context, ws string) {
	links, err := s.store.Links(ctx, ws)
	if err != nil {
		s.logRefresh(ctx, ws, "", err)
		return
	}
	fresh := map[string]PullRequest{}
	for _, p := range Providers {
		s.refreshProvider(ctx, ws, p, links, fresh)
	}
	if len(fresh) == 0 {
		return
	}
	err = s.store.UpdateLinks(ctx, ws, func(list []Link) ([]Link, error) {
		changed := false
		for i, l := range list {
			if pr, ok := fresh[l.Key()]; ok && (pr.State != l.State || pr.Title != l.Title) {
				list[i].State, list[i].Title, changed = pr.State, pr.Title, true
			}
		}
		if !changed {
			return nil, errUnchanged
		}
		return list, nil
	})
	if err != nil {
		s.logRefresh(ctx, ws, "", err)
	}
}

func (s *Service) refreshProvider(ctx context.Context, ws string, p Provider, links []Link, fresh map[string]PullRequest) {
	var refs []PRRef
	for _, l := range links {
		if l.Provider == p && !slices.Contains(refs, l.PRRef) {
			refs = append(refs, l.PRRef)
		}
	}
	if len(refs) == 0 {
		return
	}
	ctx, cred, err := s.credential(ctx, ws, p)
	if err != nil {
		if !errors.Is(err, ErrNoToken) {
			s.logRefresh(ctx, ws, p, err)
		}
		return
	}
	for _, ref := range refs {
		pr, err := s.clients[p].GetPR(ctx, cred, ref.Repo, ref.Number)
		if err != nil {
			s.logRefresh(ctx, ws, p, err)
			if IsStatus(err, 401) {
				s.forgetCLI(p) // the CLI may hold a new token by the next cycle (FR3.2)
			}
			if IsStatus(err, 429) || ctx.Err() != nil {
				return
			}
			continue
		}
		fresh[ref.Key()] = pr
	}
}

func (s *Service) logRefresh(ctx context.Context, ws string, p Provider, err error) {
	redact.Logger(ctx).WarnContext(ctx, "pull request refresh failed", "event", "scm_refresh_failed",
		"workspaceId", ws, "provider", string(p), "errorCode", errorCode(err))
}

// Watcher runs the watches whose interval has passed and refreshes linked
// PR states, on its own one-minute timer like internal/git's (ADR-003).
// ponytail: one worker for all workspaces; shard per workspace if a slow
// provider delays the others' cycles.
type Watcher struct {
	svc   *Service
	Every time.Duration // 1 minute: the finest watch interval
	Log   *slog.Logger  // set before Start

	lastRefresh map[string]time.Time // per workspace; touched only by the worker
	stop        chan struct{}
	done        chan struct{}
	stopOnce    sync.Once
}

// NewWatcher returns a stopped watcher with the one-minute tick.
func NewWatcher(svc *Service, log *slog.Logger) *Watcher {
	if log == nil {
		log = slog.New(slog.DiscardHandler)
	}
	return &Watcher{svc: svc, Every: time.Minute, Log: log, lastRefresh: map[string]time.Time{}}
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
	ctx, cancel := context.WithCancel(redact.WithLogger(context.Background(), w.Log))
	defer cancel()
	go func() {
		<-w.stop
		cancel() // stop a cycle waiting on a provider
	}()
	ticker := time.NewTicker(w.Every)
	defer ticker.Stop()
	for {
		select {
		case <-w.stop:
			return
		case <-ticker.C:
			w.cycle(ctx)
		}
	}
}

// cycle runs every due watch and, every 5 minutes, refreshes the links of
// each indexed workspace whose Backlog switch is on (FR6.1).
func (w *Watcher) cycle(ctx context.Context) {
	idx, err := w.svc.store.Index(ctx)
	if err != nil {
		w.svc.logRefresh(ctx, "", "", err)
		return
	}
	now := w.svc.Now()
	for _, ws := range idx {
		if w.svc.conn.RequireEnabled(ctx, ws) != nil {
			continue
		}
		w.runDue(ctx, ws, now)
		if now.Sub(w.lastRefresh[ws]) >= refreshEvery {
			w.lastRefresh[ws] = now
			w.svc.RefreshLinks(ctx, ws)
		}
	}
}

func (w *Watcher) runDue(ctx context.Context, ws string, now time.Time) {
	list, err := w.svc.ListWatches(ctx, ws)
	if err != nil {
		w.svc.logRefresh(ctx, ws, "", err)
		return
	}
	for _, watch := range list {
		last, _ := time.Parse(time.RFC3339, watch.LastRunAt)
		due := watch.LastRunAt == "" || !now.Before(last.Add(time.Duration(watch.IntervalMinutes)*time.Minute))
		if watch.State != WatchActive || watch.Unmapped || !due {
			continue
		}
		st, err := w.svc.settings(ctx, ws, watch.Provider)
		if err == nil {
			_, err = w.svc.runWatch(ctx, ws, watch, st)
		}
		if err != nil {
			redact.Logger(ctx).WarnContext(ctx, "watch failed", "event", "scm_watch_failed", "workspaceId", ws,
				"watchId", watch.ID, "provider", string(watch.Provider), "errorCode", errorCode(err))
		}
	}
}
