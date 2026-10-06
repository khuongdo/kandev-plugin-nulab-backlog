package git

import (
	"context"
	"sync"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/backlog"
	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/redact"
)

// cycle waits one watch interval and lets the worker finish.
func cycle(w *Watcher) {
	time.Sleep(w.Every)
	synctest.Wait()
}

func TestU4_Watcher_CreatesOneTaskPerMatchingPR(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		r := newRig(t)
		r.gw.setPRs("PROJ/web-app", prs(3))
		w := r.saveWatch(t, nil)
		watcher := r.startWatcher(t)
		require.Equal(t, 5*time.Minute, watcher.Every)
		cycle(watcher)
		require.Equal(t, []string{"Review PR #1: Change 1", "Review PR #2: Change 2", "Review PR #3: Change 3"}, r.host.titles())
		task := r.host.tasks[0]
		require.Equal(t, NewTask{WorkspaceID: ws, WorkflowID: "wf-1", WorkflowStepID: "step-1", Title: "Review PR #1: Change 1",
			Description: "https://" + host + "/git/PROJ/web-app/pullRequests/1",
			Metadata:    map[string]any{MetadataKey: LinkKey(host, 11, 1)}}, task)
		require.Len(t, r.links(t), 3, "each task is linked to its PR")
		got := r.watches(t)[0]
		require.Equal(t, w.ID, got.ID)
		require.Equal(t, 3, got.CreatedCount)
		require.Zero(t, got.PendingCount)
		cycle(watcher)
		require.Equal(t, 3, r.host.taskCount(), "never twice")
	})
}

func TestU4_Watch_CreatesAtMostTenPerCycle(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		r := newRig(t)
		r.gw.setPRs("PROJ/web-app", prs(25))
		r.saveWatch(t, nil)
		watcher := r.startWatcher(t)
		for i, want := range []struct{ created, pending int }{{10, 15}, {20, 5}, {25, 0}} {
			cycle(watcher)
			require.Equal(t, want.created, r.host.taskCount(), "cycle %d", i+1)
			got := r.watches(t)[0]
			require.Equal(t, want.created, got.CreatedCount)
			require.Equal(t, want.pending, got.PendingCount)
		}
		require.Equal(t, "Review PR #1: Change 1", r.host.titles()[0], "oldest number first")
		require.Equal(t, "Review PR #11: Change 11", r.host.titles()[10])
	})
}

func TestU4_Watcher_DeletedTaskIsNeverRecreated(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		r := newRig(t)
		r.gw.setPRs("PROJ/web-app", prs(2))
		r.saveWatch(t, nil)
		watcher := r.startWatcher(t)
		cycle(watcher)
		r.host.mu.Lock()
		r.host.deleted["task-1"] = true
		r.host.mu.Unlock()
		cycle(watcher)
		require.Equal(t, 2, r.host.taskCount())
	})
}

func TestU4_Watcher_FailureContinuesNextCycle(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		r := newRig(t)
		r.gw.setPRs("PROJ/web-app", prs(10))
		r.host.failOn = 5
		r.saveWatch(t, nil)
		watcher := r.startWatcher(t)
		cycle(watcher)
		require.Equal(t, 4, r.host.taskCount())
		require.Equal(t, 6, r.watches(t)[0].PendingCount)
		cycle(watcher)
		require.Equal(t, 10, r.host.taskCount())
		require.Equal(t, "Review PR #5: Change 5", r.host.titles()[4], "continues from the 5th")
		cycleLogs := r.events(t, "watch_cycle")
		require.EqualValues(t, 1, cycleLogs[0]["errors"])
		require.EqualValues(t, 4, cycleLogs[0]["created"])
	})
}

func TestU4_Watcher_ReservedEntryIsResolvedByMetadata(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		r := newRig(t)
		r.gw.setPRs("PROJ/web-app", prs(2))
		w := r.saveWatch(t, nil)
		// A crash after Tasks().Create for PR 1 but before the task id was stored,
		// and a crash before Tasks().Create for PR 2.
		_, err := r.host.CreateTask(r.ctx, NewTask{WorkspaceID: ws, Title: "earlier", Metadata: map[string]any{MetadataKey: LinkKey(host, 11, 1)}})
		require.NoError(t, err)
		require.NoError(t, r.store.UpdateLedger(r.ctx, ws, func(l []LedgerEntry) ([]LedgerEntry, error) {
			return append(l, LedgerEntry{Key: LinkKey(host, 11, 1), WatchID: w.ID}, LedgerEntry{Key: LinkKey(host, 11, 2), WatchID: w.ID}), nil
		}))
		watcher := r.startWatcher(t)
		cycle(watcher)
		require.Equal(t, []string{"earlier", "Review PR #2: Change 2"}, r.host.titles())
		ledger, _ := r.store.Ledger(r.ctx, ws)
		require.Equal(t, []LedgerEntry{{Key: LinkKey(host, 11, 1), TaskID: "task-1", WatchID: w.ID}, {Key: LinkKey(host, 11, 2), TaskID: "task-2", WatchID: w.ID}}, ledger)
		require.Len(t, r.links(t), 2)
	})
}

func TestU4_Watcher_RestartCreatesNoDuplicates(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		r := newRig(t)
		r.gw.setPRs("PROJ/web-app", prs(3))
		r.saveWatch(t, nil)
		first := NewWatcher(r.svc, redact.Logger(r.ctx))
		first.Start()
		cycle(first)
		first.Stop()
		svc := NewService(r.gw, r.conn, r.host, NewStore(r.state))
		second := NewWatcher(svc, redact.Logger(r.ctx))
		second.Start()
		defer second.Stop()
		cycle(second)
		require.Equal(t, 3, r.host.taskCount())
	})
}

func TestU4_Watcher_RunRacingATickMakesOneTaskPerPR(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		r := newRig(t)
		r.gw.setPRs("PROJ/web-app", prs(4))
		w := r.saveWatch(t, nil)
		watcher := r.startWatcher(t)
		time.Sleep(watcher.Every - time.Millisecond)
		var wg sync.WaitGroup
		for range 3 {
			wg.Go(func() { require.NoError(t, watcher.Run(context.Background(), ws, w.ID)) })
		}
		wg.Wait()
		time.Sleep(time.Millisecond)
		synctest.Wait()
		require.Equal(t, 4, r.host.taskCount())
		require.ErrorIs(t, watcher.Run(context.Background(), ws, "missing"), ErrNotFound)
	})
}

func TestU4_Watcher_EpochChangeDropsTheCycle(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		r := newRig(t)
		r.gw.setPRs("PROJ/web-app", prs(3))
		r.saveWatch(t, nil)
		r.gw.onPRs = func() { r.conn.update(func(c *fakeConn) { c.snap.ConnectionEpoch++ }) }
		watcher := r.startWatcher(t)
		cycle(watcher)
		require.Zero(t, r.host.taskCount(), "late results are dropped (AC1.8.3)")
		ledger, _ := r.store.Ledger(r.ctx, ws)
		require.Empty(t, ledger)
	})
}

func TestU4_Watcher_BackgroundCallsAndOneLogLine(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		r := newRig(t)
		r.gw.setPRs("PROJ/web-app", prs(2))
		r.saveWatch(t, nil)
		r.saveWatch(t, func(in *WatchInput) { in.RepoName = "api" })
		r.gw.logWait = true // what backlog.Client logs while it waits out a 429
		watcher := r.startWatcher(t)
		cycle(watcher)
		r.gw.mu.Lock()
		classes := r.gw.classes
		r.gw.mu.Unlock()
		require.Equal(t, []backlog.CallClass{backlog.Background, backlog.Background}, classes)
		lines := r.events(t, "watch_cycle")
		require.Len(t, lines, 1)
		require.EqualValues(t, 2, lines[0]["watchCount"])
		require.EqualValues(t, 2, lines[0]["created"])
		require.EqualValues(t, 0, lines[0]["errors"])
		require.EqualValues(t, 2, lines[0]["rateLimitWaits"])
		require.Contains(t, lines[0], "durationMs")
	})
}

func TestU4_Watcher_SwitchOffSkipsWorkspace(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		r := newRig(t)
		r.gw.setPRs("PROJ/web-app", prs(2))
		w := r.saveWatch(t, nil)
		r.conn.update(func(c *fakeConn) { c.disabled = true })
		watcher := r.startWatcher(t)
		calls := r.gw.total()
		for range 3 {
			cycle(watcher)
		}
		require.NoError(t, watcher.Run(context.Background(), ws, w.ID))
		synctest.Wait()
		require.Equal(t, calls, r.gw.total(), "no Backlog call while the switch is off (BR7.3)")
		require.Zero(t, r.host.taskCount())
		for _, line := range r.events(t, "watch_cycle") {
			require.EqualValues(t, 0, line["errors"], "a disabled workspace is skipped, not failed")
		}
		r.conn.update(func(c *fakeConn) { c.disabled = false })
		cycle(watcher)
		require.Equal(t, 2, r.host.taskCount(), "switched back on, it resumes")
	})
}

func TestU4_Watcher_SwitchReadFailsClosed(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		r := newRig(t)
		r.gw.setPRs("PROJ/web-app", prs(2))
		r.saveWatch(t, nil)
		r.conn.update(func(c *fakeConn) { c.switchErr = errInjected })
		watcher := r.startWatcher(t)
		calls := r.gw.total()
		cycle(watcher)
		require.Equal(t, calls, r.gw.total(), "no Backlog call when the switch cannot be read (NFR3.9)")
		require.Zero(t, r.host.taskCount())
		require.EqualValues(t, 1, r.events(t, "watch_cycle")[0]["errors"])
	})
}

func TestU4_Watcher_SwitchOffMidCycleStops(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		r := newRig(t)
		r.gw.setPRs("PROJ/web-app", prs(3))
		r.saveWatch(t, nil)
		r.gw.onPRs = func() { r.conn.update(func(c *fakeConn) { c.disabled = true }) }
		watcher := r.startWatcher(t)
		cycle(watcher)
		require.Zero(t, r.host.taskCount(), "no task is created after the switch turns off")
		ledger, _ := r.store.Ledger(r.ctx, ws)
		require.Empty(t, ledger)
	})
}
