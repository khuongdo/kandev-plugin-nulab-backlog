package issues

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/backlog"
	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/connection"
)

// syncRig links task-17 to PROJ-118 and task-18 to PROJ-120, and starts a
// syncer with the 1-minute tick. Call inside a synctest bubble.
func syncRig(t *testing.T) (*rig, *Syncer) {
	t.Helper()
	r := newRig(t)
	r.link(t, Link{IssueKey: "PROJ-118", TaskID: "task-17", TaskKey: "T-17", LastKnownStatus: "In Progress"})
	r.link(t, Link{IssueKey: "PROJ-120", TaskID: "task-18", TaskKey: "T-18", LastKnownStatus: "Resolved"})
	y := NewSyncer(r.svc, r.log)
	y.Start()
	t.Cleanup(y.Stop)
	return r, y
}

// after advances virtual time by d and lets the worker finish.
func after(d time.Duration) {
	time.Sleep(d)
	synctest.Wait()
}

// cycleLines returns the issue_sync_cycle log records.
func cycleLines(t *testing.T, logs string) []map[string]any {
	t.Helper()
	var out []map[string]any
	for _, line := range strings.Split(strings.TrimSpace(logs), "\n") {
		var m map[string]any
		if json.Unmarshal([]byte(line), &m) == nil && m["event"] == "issue_sync_cycle" {
			out = append(out, m)
		}
	}
	return out
}

func TestU3_Sync_RunsOnTheInterval(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		r, _ := syncRig(t)
		after(time.Minute)
		require.Equal(t, 2, r.gw.count("issue"), "the first cycle runs at the first tick")
		after(4 * time.Minute)
		require.Equal(t, 2, r.gw.count("issue"), "not before 5 minutes have passed")
		after(time.Minute)
		require.Equal(t, 4, r.gw.count("issue"), "AC4.2.1: every 5 minutes by default")

		_, err := r.svc.SetPollInterval(r.ctx, "ws-1", 2.0)
		require.NoError(t, err)
		after(time.Minute)
		require.Equal(t, 4, r.gw.count("issue"))
		after(time.Minute)
		require.Equal(t, 6, r.gw.count("issue"), "then every 2 minutes")
		r.gw.set(func(g *fakeGateway) {
			for _, c := range g.classes["issue"] {
				require.Equal(t, backlog.Background, c)
			}
			for op := range g.calls {
				require.Equal(t, "issue", op, "AC4.1.3: only GET reads")
			}
		})
	})
}

func TestU3_Sync_StatusShowsAfterOneCycle(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		r, _ := syncRig(t)
		r.gw.set(func(g *fakeGateway) { g.issues[0].StatusName = "Closed" })
		after(time.Minute)
		l := byTask(r.links(t), "task-17")
		require.Equal(t, "Closed", l.LastKnownStatus, "AC4.1.1")
		require.Equal(t, time.Now().UTC().Format(time.RFC3339), l.StatusUpdatedAt)
		lines := cycleLines(t, r.logs.String())
		require.Len(t, lines, 1, "AC8.3.2: one line per cycle")
		require.EqualValues(t, 1, lines[0]["workspaceCount"])
		require.EqualValues(t, 1, lines[0]["updated"])
		require.EqualValues(t, 0, lines[0]["errors"])
		require.Contains(t, lines[0], "durationMs")
		require.Contains(t, lines[0], "rateLimitWaits")
	})
}

func TestU3_Sync_UnavailableKeepsTheLink(t *testing.T) {
	for name, err := range map[string]error{"404": notFound(), "403": forbidden()} {
		t.Run(name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				r, _ := syncRig(t)
				r.gw.set(func(g *fakeGateway) { g.errs["issue:PROJ-118"] = err })
				r.gw.set(func(g *fakeGateway) { g.issues[1].StatusName = "Closed" })
				after(time.Minute)
				ls := r.links(t)
				require.Len(t, ls, 2, "AC4.1.2: the link is kept")
				require.True(t, byTask(ls, "task-17").Unavailable)
				require.Equal(t, "In Progress", byTask(ls, "task-17").LastKnownStatus)
				require.Equal(t, "Closed", byTask(ls, "task-18").LastKnownStatus, "other links still update")
				r.gw.set(func(g *fakeGateway) { delete(g.errs, "issue:PROJ-118") })
				after(5 * time.Minute)
				require.False(t, byTask(r.links(t), "task-17").Unavailable, "it comes back")
			})
		})
	}
}

func TestU3_Sync_StaleAfterThreeFailures(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		r, _ := syncRig(t)
		r.gw.set(func(g *fakeGateway) { g.errs["issue"] = &backlog.Error{Kind: backlog.KindUnreachable, Status: 500} })
		after(time.Minute)
		after(5 * time.Minute)
		require.False(t, byTask(r.links(t), "task-17").Stale())
		after(5 * time.Minute)
		l := byTask(r.links(t), "task-17")
		require.True(t, l.Stale(), "AC4.1.4")
		require.Equal(t, "In Progress", l.LastKnownStatus, "the last known status is kept")
		lines := cycleLines(t, r.logs.String())
		require.EqualValues(t, 2, lines[2]["errors"])
	})
}

func TestU3_Sync_RestartKeepsTheSchedule(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		r, y := syncRig(t)
		after(time.Minute)
		require.Equal(t, 2, r.gw.count("issue"))
		y.Stop()
		svc := NewService(r.gw, r.conn, r.host, NewStore(r.state))
		y2 := NewSyncer(svc, r.log)
		y2.Start()
		defer y2.Stop()
		after(time.Minute)
		require.Equal(t, 2, r.gw.count("issue"), "AC4.1.5: the new worker continues from lastCycleAt")
		after(3 * time.Minute)
		require.Equal(t, 2, r.gw.count("issue"))
		after(time.Minute)
		require.Equal(t, 4, r.gw.count("issue"))
	})
}

func TestU3_Sync_RefreshNeverOverlaps(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		r, y := syncRig(t)
		block := make(chan struct{})
		r.gw.set(func(g *fakeGateway) { g.block = block })
		time.Sleep(time.Minute) // the tick's cycle now waits on Backlog
		synctest.Wait()
		done := make(chan RefreshResult)
		go func() {
			res, err := y.Refresh(r.ctx, "ws-1")
			require.NoError(t, err)
			done <- res
		}()
		synctest.Wait()
		r.gw.set(func(g *fakeGateway) { g.block = nil })
		close(block)
		res := <-done
		require.LessOrEqual(t, r.gw.peak(), 1, "AC4.2.4: never two polls at once")
		require.Equal(t, time.Now().UTC().Format(time.RFC3339), res.RefreshedAt)
		require.Equal(t, 4, r.gw.count("issue"), "the tick's cycle, then the refresh")
	})
}

func TestU3_Sync_RefreshReturnsTheCount(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		r, y := syncRig(t)
		r.gw.set(func(g *fakeGateway) { g.issues[0].StatusName = "Closed" })
		res, err := y.Refresh(r.ctx, "ws-1")
		require.NoError(t, err)
		require.Equal(t, RefreshResult{UpdatedCount: 1, RefreshedAt: time.Now().UTC().Format(time.RFC3339)}, res, "AC4.2.3")
		require.Len(t, cycleLines(t, r.logs.String()), 1)
		after(time.Minute)
		require.Equal(t, 2, r.gw.count("issue"), "a refresh counts as the workspace's cycle")
	})
}

func TestU3_Sync_SkipsWhenOffOrDisconnected(t *testing.T) {
	cases := map[string]func(c *fakeConn){
		"switch off":    func(c *fakeConn) { c.disabled = true },
		"switch broken": func(c *fakeConn) { c.enabledErr = connection.ErrStore },
		"not connected": func(c *fakeConn) { c.currentErr = connection.ErrNotConnected },
	}
	for name, setup := range cases {
		t.Run(name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				r, _ := syncRig(t)
				r.conn.set(setup)
				after(time.Minute)
				require.Zero(t, r.gw.total(), "AC1.5.4")
				require.Zero(t, r.host.lists())
			})
		})
	}
}

func TestU3_Sync_DropsResultsOfAnOldEpoch(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		r, _ := syncRig(t)
		block := make(chan struct{})
		r.gw.set(func(g *fakeGateway) {
			g.block = block
			g.issues[0].StatusName = "Closed"
		})
		time.Sleep(time.Minute)
		synctest.Wait()
		r.conn.set(func(c *fakeConn) { c.snap.ConnectionEpoch = 2 })
		close(block)
		synctest.Wait()
		require.Equal(t, "In Progress", byTask(r.links(t), "task-17").LastKnownStatus, "AC1.8.3")
	})
}

// R-02 (review 1): the task list is not a reliable inventory (config tasks,
// offset paging) and Kandev only reports a definite not-found as a gRPC
// NotFound, which needs a go.mod change to read. So a task missing from the
// list keeps its link; task.deleted is the only removal. The missing link is
// not polled and is counted once in the cycle line.
func TestU3_Sync_KeepsLinksOfMissingTasks(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		r, _ := syncRig(t)
		r.host.remove("task-18")
		r.host.mu.Lock()
		r.host.tasks[0].Key = "WEB-17"
		r.host.mu.Unlock()
		after(time.Minute)
		ls := r.links(t)
		require.Len(t, ls, 2, "R-02: absence from the task list never deletes a link")
		require.Equal(t, "WEB-17", byTask(ls, "task-17").TaskKey, "the task key is refreshed")
		require.Equal(t, "T-18", byTask(ls, "task-18").TaskKey)
		require.Equal(t, 1, r.host.lists(), "one task list per workspace per cycle")
		require.Equal(t, 1, r.gw.count("issue"), "a missing task's issue is not read")
		after(5 * time.Minute)
		lines := cycleLines(t, r.logs.String())
		require.Len(t, lines, 2)
		for _, l := range lines {
			require.EqualValues(t, 1, l["missingTasks"], "logged once per cycle")
		}
		require.Len(t, r.links(t), 2)
	})
}

// R-01 (review 1): a link written while a cycle waits on Backlog refers to a
// task the cycle's task list has never seen; the cycle must keep it.
func TestU3_Sync_KeepsLinksMadeMidCycle(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		r, _ := syncRig(t)
		block := make(chan struct{})
		r.gw.set(func(g *fakeGateway) { g.block = block })
		after(time.Minute) // the task list is read; the cycle waits on Backlog
		require.Equal(t, 1, r.host.lists())
		ref, err := r.host.CreateTask(r.ctx, NewTask{WorkspaceID: "ws-1", Title: "New"})
		require.NoError(t, err)
		r.link(t, Link{IssueKey: "PROJ-121", TaskID: ref.ID, TaskKey: ref.Key})
		r.gw.set(func(g *fakeGateway) { g.block = nil })
		close(block)
		synctest.Wait()
		ls := r.links(t)
		require.Len(t, ls, 3, "R-01: the new link survives the cycle")
		require.Equal(t, ref.Key, byTask(ls, ref.ID).TaskKey)
	})
}

func TestU3_Sync_ListIsNotBlockedByACycle(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		r, _ := syncRig(t)
		block := make(chan struct{})
		r.gw.set(func(g *fakeGateway) { g.block = block })
		after(time.Minute) // the cycle waits on Backlog
		page, err := r.svc.List(context.Background(), "ws-1", Query{})
		require.NoError(t, err)
		require.NotEmpty(t, page.Items)
		close(block)
	})
}
