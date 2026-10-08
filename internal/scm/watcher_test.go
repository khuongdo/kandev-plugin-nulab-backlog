package scm

import (
	"errors"
	"sync"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/require"
)

func watchInput(p Provider) WatchInput {
	return WatchInput{QueryInput: query("Reviews", p), WorkflowID: "wf-1", WorkflowStepID: "step-1"}
}

// FR4.3: save, edit (keeps state and counts), pause, resume, delete.
func TestWatches_CRUD(t *testing.T) {
	h := newHarness(t)
	h.mapRepo(t, GitHub, "PROJ", "acme/web")
	w, err := h.svc.SaveWatch(h.ctx, ws, watchInput(GitHub))
	require.NoError(t, err)
	require.Equal(t, WatchActive, w.State)
	require.Equal(t, 5, w.IntervalMinutes, "5 minutes by default")

	p, err := h.svc.PauseWatch(h.ctx, ws, w.ID)
	require.NoError(t, err)
	require.Equal(t, WatchPaused, p.State)
	edit := watchInput(GitHub)
	edit.ID, edit.Name, edit.IntervalMinutes = w.ID, "Renamed", 30
	e, err := h.svc.SaveWatch(h.ctx, ws, edit)
	require.NoError(t, err)
	require.Equal(t, WatchPaused, e.State, "an edit keeps the state")
	require.Equal(t, 30, e.IntervalMinutes)
	r, err := h.svc.ResumeWatch(h.ctx, ws, w.ID)
	require.NoError(t, err)
	require.Equal(t, WatchActive, r.State)

	list, err := h.svc.ListWatches(h.ctx, ws)
	require.NoError(t, err)
	require.Len(t, list, 1)
	require.NoError(t, h.svc.DeleteWatch(h.ctx, ws, w.ID))
	require.ErrorIs(t, h.svc.DeleteWatch(h.ctx, ws, w.ID), ErrNotFound)
	_, err = h.svc.PauseWatch(h.ctx, ws, w.ID)
	require.ErrorIs(t, err, ErrNotFound)
	edit.ID = "missing"
	_, err = h.svc.SaveWatch(h.ctx, ws, edit)
	require.ErrorIs(t, err, ErrNotFound)

	bad := watchInput(GitHub)
	bad.Repo = "acme/api"
	_, err = h.svc.SaveWatch(h.ctx, ws, bad)
	require.Equal(t, FieldRepository, fieldOf(t, err))
	bad = watchInput(GitHub)
	bad.WorkflowID = ""
	_, err = h.svc.SaveWatch(h.ctx, ws, bad)
	require.Equal(t, FieldWorkflow, fieldOf(t, err))
}

// FR4.3: at most one new task per watch per run, oldest first, never twice.
func TestRunWatch_OneTaskPerRunAndNeverTwice(t *testing.T) {
	h := newHarness(t)
	h.mapRepo(t, GitHub, "PROJ", "acme/web")
	h.clients[GitHub].setPRs("acme/web",
		pr(GitHub, "acme/web", 9, "PROJ-5 Newer", "b", StateOpen),
		pr(GitHub, "acme/web", 8, "Older", "a", StateOpen))
	w, err := h.svc.SaveWatch(h.ctx, ws, watchInput(GitHub))
	require.NoError(t, err)

	for i, want := range []int{1, 1, 0} {
		n, err := h.svc.RunWatch(h.ctx, ws, w.ID)
		require.NoError(t, err)
		require.Equal(t, want, n, "run %d", i)
	}
	tasks := h.tasks.all()
	require.Len(t, tasks, 2)
	require.Equal(t, NewTask{WorkspaceID: ws, WorkflowID: "wf-1", WorkflowStepID: "step-1", Title: "Review PR #8: Older",
		Description: "https://github.com/acme/web/pull/8", Metadata: map[string]any{MetadataKey: "github|acme/web|8"}}, tasks[0])
	links, err := h.svc.Links(h.ctx, ws, "task-1", "")
	require.NoError(t, err)
	require.Len(t, links, 1, "the created task is linked to its PR")
	auto, err := h.svc.Links(h.ctx, ws, "", "PROJ-5")
	require.NoError(t, err)
	require.Len(t, auto, 1, "FR5.3: a watch run also auto-links")
	list, _ := h.svc.ListWatches(h.ctx, ws)
	require.Equal(t, 2, list[0].CreatedCount)
	require.Equal(t, "2026-10-07T09:00:00Z", list[0].LastRunAt)

	_, err = h.svc.RunWatch(h.ctx, ws, "missing")
	require.ErrorIs(t, err, ErrNotFound)
	_, err = h.svc.PauseWatch(h.ctx, ws, w.ID)
	require.NoError(t, err)
	_, err = h.svc.RunWatch(h.ctx, ws, w.ID)
	require.ErrorIs(t, err, ErrConflict, "a paused watch does not run")
}

// FR4.3: concurrent runs reserve in the ledger, so a PR gets one task.
func TestRunWatch_ConcurrentRunsCreateOneTask(t *testing.T) {
	h := newHarness(t)
	h.mapRepo(t, GitLab, "PROJ", "acme/web")
	h.clients[GitLab].setPRs("acme/web", pr(GitLab, "acme/web", 1, "x", "b", StateOpen))
	w, err := h.svc.SaveWatch(h.ctx, ws, watchInput(GitLab))
	require.NoError(t, err)
	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() { _, _ = h.svc.RunWatch(h.ctx, ws, w.ID) })
	}
	wg.Wait()
	require.Len(t, h.tasks.all(), 1)
}

// A failed task creation frees the reservation, so the next run retries.
func TestRunWatch_FailedCreateIsRetried(t *testing.T) {
	h := newHarness(t)
	h.mapRepo(t, GitHub, "PROJ", "acme/web")
	h.clients[GitHub].setPRs("acme/web", pr(GitHub, "acme/web", 1, "x", "b", StateOpen))
	w, err := h.svc.SaveWatch(h.ctx, ws, watchInput(GitHub))
	require.NoError(t, err)
	h.tasks.err = errors.New("workflow gone")
	_, err = h.svc.RunWatch(h.ctx, ws, w.ID)
	require.Error(t, err)
	h.tasks.err = nil
	n, err := h.svc.RunWatch(h.ctx, ws, w.ID)
	require.NoError(t, err)
	require.Equal(t, 1, n)
}

// FR4.3: the watcher runs a watch once its interval has passed (injected clock).
func TestCycle_RespectsThePerWatchInterval(t *testing.T) {
	h := newHarness(t)
	h.mapRepo(t, GitHub, "PROJ", "acme/web")
	h.clients[GitHub].setPRs("acme/web", pr(GitHub, "acme/web", 1, "a", "a", StateOpen),
		pr(GitHub, "acme/web", 2, "b", "b", StateOpen), pr(GitHub, "acme/web", 3, "c", "c", StateOpen))
	in := watchInput(GitHub)
	in.IntervalMinutes = 10
	_, err := h.svc.SaveWatch(h.ctx, ws, in)
	require.NoError(t, err)
	w := NewWatcher(h.svc, nil)

	w.cycle(h.ctx) // never ran: due
	require.Len(t, h.tasks.all(), 1)
	h.now = h.now.Add(9 * time.Minute)
	w.cycle(h.ctx)
	require.Len(t, h.tasks.all(), 1, "not due yet")
	h.now = h.now.Add(time.Minute)
	w.cycle(h.ctx)
	require.Len(t, h.tasks.all(), 2)
}

// FR6.1: a workspace with Backlog off is skipped before any provider call.
func TestCycle_SkipsAWorkspaceWithBacklogOff(t *testing.T) {
	h := newHarness(t)
	h.mapRepo(t, GitHub, "PROJ", "acme/web")
	_, err := h.svc.SaveWatch(h.ctx, ws, watchInput(GitHub))
	require.NoError(t, err)
	h.conn.set(func(c *fakeConn) { c.disabled = true })
	NewWatcher(h.svc, nil).cycle(h.ctx)
	require.Zero(t, h.clients[GitHub].count("ListPRs"))
	_, err = h.svc.RunWatch(h.ctx, ws, "any")
	require.ErrorIs(t, err, ErrNotFound)
}

// FR4.4: linked PR states are refreshed on the watcher schedule.
func TestRefreshLinks_UpdatesStatesAndTitles(t *testing.T) {
	h := newHarness(t)
	h.mapRepo(t, GitHub, "PROJ", "acme/web")
	h.clients[GitHub].setPRs("acme/web", pr(GitHub, "acme/web", 1, "Old title", "b", StateOpen))
	_, err := h.svc.Link(h.ctx, ws, "task-1", "https://github.com/acme/web/pull/1")
	require.NoError(t, err)
	h.clients[GitHub].setPRs("acme/web", pr(GitHub, "acme/web", 1, "New title", "b", StateMerged))
	w := NewWatcher(h.svc, nil)
	w.cycle(h.ctx)
	links, err := h.svc.Links(h.ctx, ws, "task-1", "")
	require.NoError(t, err)
	require.Equal(t, StateMerged, links[0].State)
	require.Equal(t, "New title", links[0].Title)

	h.clients[GitHub].setPRs("acme/web", pr(GitHub, "acme/web", 1, "New title", "b", StateClosed))
	h.now = h.now.Add(4 * time.Minute)
	w.cycle(h.ctx)
	links, _ = h.svc.Links(h.ctx, ws, "task-1", "")
	require.Equal(t, StateMerged, links[0].State, "refreshed every 5 minutes, like Backlog Git watches")
	h.now = h.now.Add(time.Minute)
	w.cycle(h.ctx)
	links, _ = h.svc.Links(h.ctx, ws, "task-1", "")
	require.Equal(t, StateClosed, links[0].State)
}

// NFR3: a 429 from one provider stops only that provider's refresh.
func TestRefreshLinks_ARateLimitedProviderDoesNotBlockOthers(t *testing.T) {
	h := newHarness(t)
	for _, p := range []Provider{GitHub, GitLab} {
		h.mapRepo(t, p, "PROJ", "acme/web")
		h.clients[p].setPRs("acme/web", pr(p, "acme/web", 1, "x", "b", StateOpen), pr(p, "acme/web", 2, "y", "b", StateOpen))
	}
	for _, url := range []string{"https://github.com/acme/web/pull/1", "https://github.com/acme/web/pull/2",
		"https://gitlab.com/acme/web/-/merge_requests/1"} {
		_, err := h.svc.Link(h.ctx, ws, "task-1", url)
		require.NoError(t, err)
	}
	h.clients[GitHub].setErr("GetPR", &HTTPError{Provider: GitHub, Status: 429, RetryAfter: time.Minute})
	h.clients[GitLab].setPRs("acme/web", pr(GitLab, "acme/web", 1, "x", "b", StateMerged))
	before := h.clients[GitHub].count("GetPR")
	h.svc.RefreshLinks(h.ctx, ws)
	require.Equal(t, before+1, h.clients[GitHub].count("GetPR"), "GitHub stops after its first 429")
	links, _ := h.svc.Links(h.ctx, ws, "task-1", "")
	for _, l := range links {
		if l.Provider == GitLab {
			require.Equal(t, StateMerged, l.State)
		}
	}
	require.Contains(t, h.logs.String(), `"event":"scm_refresh_failed"`)
}

// The watcher goroutine runs cycles on its own ticker and stops cleanly.
func TestWatcher_StartRunsCyclesAndStopEnds(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		h := newHarness(t)
		h.mapRepo(t, Bitbucket, "PROJ", "acme/web")
		h.clients[Bitbucket].setPRs("acme/web", pr(Bitbucket, "acme/web", 1, "x", "b", StateOpen))
		_, err := h.svc.SaveWatch(h.ctx, ws, watchInput(Bitbucket))
		require.NoError(t, err)
		w := NewWatcher(h.svc, nil)
		w.Stop() // stopping a watcher that never started is fine
		w.Start()
		time.Sleep(w.Every + time.Second)
		synctest.Wait()
		require.Len(t, h.tasks.all(), 1)
		w.Stop()
		w.Stop()
	})
}

// FR1.5: the watcher runs and refreshes only the active provider's items.
func TestCycle_TouchesOnlyTheActiveProvider(t *testing.T) {
	h := newHarness(t)
	for _, p := range []Provider{GitHub, GitLab} {
		h.mapRepo(t, p, "PROJ", "acme/web")
		h.clients[p].setPRs("acme/web", pr(p, "acme/web", 1, "x", "b", StateOpen))
		_, err := h.svc.SaveWatch(h.ctx, ws, watchInput(p))
		require.NoError(t, err)
	}
	for _, url := range []string{"https://github.com/acme/web/pull/1", "https://gitlab.com/acme/web/-/merge_requests/1"} {
		_, err := h.svc.Link(h.ctx, ws, "task-1", url)
		require.NoError(t, err)
	}
	require.NoError(t, h.svc.SetActive(h.ctx, ws, GitLab))
	before := h.clients[GitHub].count("GetPR") + h.clients[GitHub].count("ListPRs")
	NewWatcher(h.svc, nil).cycle(h.ctx)
	require.Equal(t, before, h.clients[GitHub].count("GetPR")+h.clients[GitHub].count("ListPRs"),
		"no GitHub call while GitLab is active")
	require.Positive(t, h.clients[GitLab].count("GetPR"), "GitLab links are refreshed")
	tasks := h.tasks.all()
	require.Len(t, tasks, 1)
	require.Equal(t, "https://gitlab.com/acme/web/-/merge_requests/1", tasks[0].Description)
}
