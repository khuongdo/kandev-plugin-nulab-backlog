package issues

import (
	"context"
	"fmt"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/backlog"
	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/connection"
)

// watchIssues replaces the fake's issues with n open PROJ issues created one
// minute apart from 2026-10-01T00:00Z, ids 7001..7000+n, keys PROJ-1..PROJ-n.
func watchIssues(g *fakeGateway, n int) {
	g.set(func(g *fakeGateway) {
		g.issues = nil
		base := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
		for i := 1; i <= n; i++ {
			g.issues = append(g.issues, backlog.Issue{ID: int64(7000 + i), ProjectID: 101, IssueKey: fmt.Sprintf("PROJ-%d", i),
				Summary: fmt.Sprintf("Bug %d", i), StatusID: 1, StatusName: "Open", PriorityID: 3,
				Created: base.Add(time.Duration(i) * time.Minute).Format(time.RFC3339)})
		}
	})
}

func (r *rig) saveIssueWatch(t *testing.T, mut func(*IssueWatchInput)) IssueWatch {
	t.Helper()
	in := IssueWatchInput{Name: "Open bugs", ProjectKey: "PROJ", StatusIDs: []int64{1}, Assignee: WhoAnyone,
		Creator: WhoAnyone, WorkflowID: "wf-1", WorkflowStepID: "step-1"}
	if mut != nil {
		mut(&in)
	}
	w, err := r.svc.SaveWatch(r.ctx, "ws-1", in)
	require.NoError(t, err)
	return w
}

func (r *rig) issueWatch(t *testing.T, id string) IssueWatch {
	t.Helper()
	list, err := r.store.Watches(context.Background(), "ws-1")
	require.NoError(t, err)
	for _, w := range list {
		if w.ID == id {
			return w
		}
	}
	t.Fatalf("watch %s not found", id)
	return IssueWatch{}
}

func (r *rig) ledger(t *testing.T, id string) []IssueWatchLedgerEntry {
	t.Helper()
	l, err := r.store.Ledger(context.Background(), "ws-1", id)
	require.NoError(t, err)
	return l
}

func (r *rig) setCursor(t *testing.T, id string, c *IssueCursor) {
	t.Helper()
	require.NoError(t, r.store.UpdateWatches(context.Background(), "ws-1", func(l []IssueWatch) ([]IssueWatch, error) {
		for i := range l {
			if l[i].ID == id {
				l[i].Cursor = c
			}
		}
		return l, nil
	}))
}

func (r *rig) issueQueries() []backlog.IssueQuery {
	r.gw.mu.Lock()
	defer r.gw.mu.Unlock()
	return append([]backlog.IssueQuery(nil), r.gw.queries...)
}

func TestIssueWatch_SaveNewResolvesMeAndDefaults(t *testing.T) {
	r := newRig(t)
	w := r.saveIssueWatch(t, func(in *IssueWatchInput) { in.Creator = WhoMe })
	require.NotEmpty(t, w.ID)
	require.Equal(t, StateActive, w.State)
	require.Equal(t, 5, w.IntervalMinutes, "BR3.2")
	require.Equal(t, int64(1), w.CreatedUserID, "me is the connected user")
	require.Zero(t, w.AssigneeID)
	require.Equal(t, spaceHost, w.SpaceHost)
	require.Nil(t, w.Cursor)
	list, err := r.svc.ListWatches(r.ctx, "ws-1")
	require.NoError(t, err)
	require.Equal(t, []IssueWatch{w}, list)

	_, err = r.svc.SaveWatch(r.ctx, "ws-1", IssueWatchInput{Name: "OPEN BUGS", ProjectKey: "PROJ", StatusIDs: []int64{1}, WorkflowID: "wf-1"})
	require.Equal(t, FieldName, fieldOf(t, err), "names are unique, ignoring case")
	_, err = r.svc.SaveWatch(r.ctx, "ws-1", IssueWatchInput{ID: "missing", Name: "x", ProjectKey: "PROJ", StatusIDs: []int64{1}, WorkflowID: "wf-1"})
	require.ErrorIs(t, err, ErrNotFound)
	empty, err := r.svc.ListWatches(r.ctx, "ws-2")
	require.NoError(t, err)
	require.NotNil(t, empty)
}

// BR3.8: a filter change clears the cursor; other edits keep it and the counts.
func TestIssueWatch_EditKeepsOrClearsTheCursor(t *testing.T) {
	r := newRig(t)
	w := r.saveIssueWatch(t, nil)
	cursor := &IssueCursor{Created: time.Date(2026, 10, 1, 0, 1, 0, 0, time.UTC), IssueID: 7001}
	r.setCursor(t, w.ID, cursor)
	edit := func(mut func(*IssueWatchInput)) IssueWatch {
		in := IssueWatchInput{ID: w.ID, Name: "Open bugs", ProjectKey: "PROJ", StatusIDs: []int64{1}, Assignee: WhoAnyone,
			Creator: WhoAnyone, WorkflowID: "wf-1"}
		mut(&in)
		got, err := r.svc.SaveWatch(r.ctx, "ws-1", in)
		require.NoError(t, err)
		return got
	}
	got := edit(func(in *IssueWatchInput) {
		in.Name, in.WorkflowID = "Renamed", "wf-2"
		in.IntervalMinutes = minutes(30)
	})
	require.Equal(t, cursor.IssueID, got.Cursor.IssueID, "name, workflow and interval keep the cursor")
	require.Equal(t, 30, got.IntervalMinutes)
	for name, mut := range map[string]func(*IssueWatchInput){
		"statuses": func(in *IssueWatchInput) { in.StatusIDs = []int64{1, 2} },
		"assignee": func(in *IssueWatchInput) { in.Assignee = WhoMe },
		"creator":  func(in *IssueWatchInput) { in.Creator = WhoMe },
		"project":  func(in *IssueWatchInput) { in.ProjectKey = "DEMO" },
	} {
		r.setCursor(t, w.ID, cursor)
		require.Nil(t, edit(mut).Cursor, name)
	}
}

// BR3.9, SM1: pause and resume; run now only for an active watch; a lost
// connection remembers the state and a reconnect restores it.
func TestIssueWatch_StatesFollowPauseAndTheConnection(t *testing.T) {
	r := newRig(t)
	w := r.saveIssueWatch(t, nil)
	y := NewWatcher(r.svc, r.log)
	paused, err := r.svc.PauseWatch(r.ctx, "ws-1", w.ID)
	require.NoError(t, err)
	require.Equal(t, StatePaused, paused.State)
	require.ErrorIs(t, y.Run(r.ctx, "ws-1", w.ID), ErrConflict, "run now needs an active watch")
	require.ErrorIs(t, y.Run(r.ctx, "ws-1", "missing"), ErrNotFound)

	r.svc.OnConnectionChanged(r.ctx, connection.ConnectionChanged{WorkspaceID: "ws-1", ConnectionEpoch: 2,
		Reason: connection.ReasonDisconnected})
	got := r.issueWatch(t, w.ID)
	require.Equal(t, StateNotConnected, got.State)
	require.Equal(t, StatePaused, got.StateBeforeDisconnect)
	_, err = r.svc.ResumeWatch(r.ctx, "ws-1", w.ID)
	require.ErrorIs(t, err, ErrConflict, "a not-connected watch cannot be resumed")

	r.svc.OnConnectionChanged(r.ctx, connection.ConnectionChanged{WorkspaceID: "ws-1", ConnectionEpoch: 3,
		SpaceHost: spaceHost, SelectedProjects: []string{"PROJ"}, Restore: true})
	got = r.issueWatch(t, w.ID)
	require.Equal(t, StatePaused, got.State, "reconnecting restores the state before")
	require.Empty(t, got.StateBeforeDisconnect)
	resumed, err := r.svc.ResumeWatch(r.ctx, "ws-1", w.ID)
	require.NoError(t, err)
	require.Equal(t, StateActive, resumed.State)
	require.NoError(t, y.Run(r.ctx, "ws-1", w.ID))
}

// BR3.10: deleting a watch deletes its ledger; its tasks and links stay.
func TestIssueWatch_DeleteKeepsTasksAndLinks(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		r := newRig(t)
		watchIssues(r.gw, 1)
		w := r.saveIssueWatch(t, nil)
		_, err := r.svc.runIssueWatch(r.ctx, "ws-1", w.ID)
		require.NoError(t, err)
		require.Len(t, r.ledger(t, w.ID), 1)
		require.NoError(t, r.svc.DeleteWatch(r.ctx, "ws-1", w.ID))
		require.Empty(t, r.ledger(t, w.ID))
		require.Len(t, r.links(t), 1, "the link stays")
		require.Equal(t, 1, r.host.createCount())
		require.ErrorIs(t, r.svc.DeleteWatch(r.ctx, "ws-1", w.ID), ErrNotFound)
	})
}

// BR3.4, BR3.5, FR3.3: existing issues are picked oldest first, one task per run.
func TestIssueWatchRun_OneTaskPerRunOldestFirst(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		r := newRig(t)
		watchIssues(r.gw, 3)
		w := r.saveIssueWatch(t, nil)
		for i, key := range []string{"PROJ-1", "PROJ-2", "PROJ-3"} {
			n, err := r.svc.runIssueWatch(r.ctx, "ws-1", w.ID)
			require.NoError(t, err)
			require.Equal(t, 1, n, "one task per run")
			got := r.issueWatch(t, w.ID)
			require.Equal(t, i+1, got.CreatedCount)
			require.Equal(t, 2-i, got.PendingCount, "matching issues still waiting")
			require.Equal(t, int64(7001+i), got.Cursor.IssueID)
			require.NotEmpty(t, got.LastRunAt)
			l := byTask(r.links(t), fmt.Sprintf("task-%d", 19+i))
			require.Equal(t, key, l.IssueKey, "linked like Create task from issue")
		}
		n, err := r.svc.runIssueWatch(r.ctx, "ws-1", w.ID)
		require.NoError(t, err)
		require.Zero(t, n, "no new issue, no task")
		require.Equal(t, NewTask{WorkspaceID: "ws-1", WorkflowID: "wf-1", WorkflowStepID: "step-1", Title: "Bug 1",
			Description: "Backlog: https://" + spaceHost + "/view/PROJ-1", Priority: "medium"}, r.host.creates[0])
		q := r.issueQueries()[0]
		require.Equal(t, backlog.IssueQuery{ProjectIDs: []int64{101}, StatusIDs: []int64{1}, Sort: "created", Order: "asc", Count: 100}, q)
		require.Equal(t, "2026-09-30", r.issueQueries()[1].CreatedSince, "later runs start at the cursor's day")
		for _, c := range r.gw.classes["issues"] {
			require.Equal(t, backlog.Background, c)
		}
		for _, e := range r.ledger(t, w.ID) {
			require.Equal(t, OutcomeCreated, e.Outcome)
			require.NotEmpty(t, e.TaskID)
		}
	})
}

// BR3.6: an issue with a link is skipped in the same run and never gets a task.
func TestIssueWatchRun_SkipsLinkedIssues(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		r := newRig(t)
		watchIssues(r.gw, 2)
		r.link(t, Link{IssueKey: "PROJ-1", IssueID: 7001, TaskID: "task-17"})
		w := r.saveIssueWatch(t, func(in *IssueWatchInput) { in.Creator = WhoMe })
		r.gw.set(func(g *fakeGateway) { g.creators[7001], g.creators[7002] = 1, 1 })
		n, err := r.svc.runIssueWatch(r.ctx, "ws-1", w.ID)
		require.NoError(t, err)
		require.Equal(t, 1, n)
		require.Equal(t, "Bug 2", r.host.creates[0].Title)
		l := r.ledger(t, w.ID)
		require.Equal(t, OutcomeSkippedLinked, l[0].Outcome)
		require.Equal(t, spaceHost+"|7001", l[0].Key)
		require.Equal(t, []int64{1}, r.issueQueries()[0].CreatedUserIDs, "creator me filters by the resolved id")
	})
}

// BR3.7: a failed create removes the reservation and keeps the cursor; a
// reservation left by a crash is finished when the link exists, else retried.
func TestIssueWatchRun_ReserveCreateMark(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		r := newRig(t)
		watchIssues(r.gw, 2)
		w := r.saveIssueWatch(t, nil)
		r.host.failNth = 1
		n, err := r.svc.runIssueWatch(r.ctx, "ws-1", w.ID)
		require.Error(t, err)
		require.Zero(t, n)
		require.Empty(t, r.ledger(t, w.ID), "the reservation is removed")
		got := r.issueWatch(t, w.ID)
		require.Nil(t, got.Cursor, "the cursor is kept")
		require.Equal(t, ErrorUnavailable, got.LastError)

		// A crash left PROJ-1 reserved and PROJ-1 is linked meanwhile; PROJ-2 reserved without a link.
		require.NoError(t, r.store.UpdateLedger(r.ctx, "ws-1", w.ID, func([]IssueWatchLedgerEntry) ([]IssueWatchLedgerEntry, error) {
			return []IssueWatchLedgerEntry{{Key: spaceHost + "|7001", Outcome: OutcomeReserved}, {Key: spaceHost + "|7002", Outcome: OutcomeReserved}}, nil
		}))
		r.link(t, Link{IssueKey: "PROJ-1", IssueID: 7001, TaskID: "task-17"})
		n, err = r.svc.runIssueWatch(r.ctx, "ws-1", w.ID)
		require.NoError(t, err)
		require.Equal(t, 1, n)
		l := r.ledger(t, w.ID)
		require.Equal(t, IssueWatchLedgerEntry{Key: spaceHost + "|7001", WatchID: w.ID, Outcome: OutcomeCreated, TaskID: "task-17", At: l[0].At}, l[0])
		require.Equal(t, OutcomeCreated, l[1].Outcome, "the unlinked reservation was retried")
		require.Equal(t, "Bug 2", r.host.creates[len(r.host.creates)-1].Title)
		require.Empty(t, r.issueWatch(t, w.ID).LastError, "a good run clears the error")
	})
}

// BR3.12: at most 5 pages of 100 per run; the cursor moves past every issue
// passed, so handled issues never block progress.
func TestIssueWatchRun_AtMostFivePagesAndTheCursorMoves(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		r := newRig(t)
		watchIssues(r.gw, 620)
		w := r.saveIssueWatch(t, nil)
		handled := make([]IssueWatchLedgerEntry, 0, 600)
		for i := 1; i <= 600; i++ {
			handled = append(handled, IssueWatchLedgerEntry{Key: fmt.Sprintf("%s|%d", spaceHost, 7000+i), Outcome: OutcomeCreated, TaskID: "old"})
		}
		require.NoError(t, r.store.UpdateLedger(r.ctx, "ws-1", w.ID, func([]IssueWatchLedgerEntry) ([]IssueWatchLedgerEntry, error) {
			return handled, nil
		}))
		n, err := r.svc.runIssueWatch(r.ctx, "ws-1", w.ID)
		require.NoError(t, err)
		require.Zero(t, n)
		require.Len(t, r.issueQueries(), 5, "five pages at most")
		c := r.issueWatch(t, w.ID).Cursor
		require.Equal(t, int64(7500), c.IssueID, "the cursor passed every handled issue")
		require.Equal(t, 499, c.DayOffset)

		n, err = r.svc.runIssueWatch(r.ctx, "ws-1", w.ID)
		require.NoError(t, err)
		require.Equal(t, 1, n)
		next := r.issueQueries()[5]
		require.Equal(t, 494, next.Offset, "the next run starts near the cursor")
		require.Equal(t, "Bug 601", r.host.creates[0].Title)
	})
}

// R-09: the list shrank by more than 5 before the cursor, so the start offset
// overshoots; the run restarts once from 0 and skips no issue.
func TestIssueWatchRun_OvershootRestartsFromZero(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		r := newRig(t)
		watchIssues(r.gw, 5)
		w := r.saveIssueWatch(t, nil)
		r.setCursor(t, w.ID, &IssueCursor{Created: time.Date(2026, 10, 1, 0, 3, 0, 0, time.UTC), IssueID: 7003, DayOffset: 20})
		n, err := r.svc.runIssueWatch(r.ctx, "ws-1", w.ID)
		require.NoError(t, err)
		require.Equal(t, 1, n)
		require.Equal(t, "Bug 4", r.host.creates[0].Title, "the issue right after the cursor")
		q := r.issueQueries()
		require.Equal(t, []int{15, 0}, []int{q[0].Offset, q[1].Offset})
	})
}

// R-10: a cursor just after midnight UTC still sees later issues of that day.
func TestIssueWatchRun_CreatedSinceIsTheDayBefore(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		r := newRig(t)
		r.gw.set(func(g *fakeGateway) {
			g.issues = []backlog.Issue{
				{ID: 8001, ProjectID: 101, IssueKey: "PROJ-1", Summary: "Old", StatusID: 1, Created: "2026-10-07T00:30:00Z"},
				{ID: 8002, ProjectID: 101, IssueKey: "PROJ-2", Summary: "New", StatusID: 1, Created: "2026-10-07T05:00:00Z"},
			}
		})
		w := r.saveIssueWatch(t, nil)
		r.setCursor(t, w.ID, &IssueCursor{Created: time.Date(2026, 10, 7, 0, 30, 0, 0, time.UTC), IssueID: 8001})
		n, err := r.svc.runIssueWatch(r.ctx, "ws-1", w.ID)
		require.NoError(t, err)
		require.Equal(t, 1, n)
		require.Equal(t, "New", r.host.creates[0].Title)
		require.Equal(t, "2026-10-06", r.issueQueries()[0].CreatedSince)
	})
}

// BR3.11, BR3.13, BR3.14: errors are stored on the watch and no task is made.
func TestIssueWatchRun_ErrorsAreRecorded(t *testing.T) {
	cases := map[string]struct {
		setup func(r *rig, w IssueWatch)
		want  string
	}{
		"401": {func(r *rig, _ IssueWatch) {
			r.gw.errs["issues"] = &backlog.Error{Kind: backlog.KindUnauthorized, Status: 401}
		}, ErrorUnauthorized},
		"429": {func(r *rig, _ IssueWatch) { r.gw.errs["issues"] = rateLimited(30 * time.Second) }, ErrorRateLimited},
		"ledger full": {func(r *rig, w IssueWatch) {
			full := make([]IssueWatchLedgerEntry, maxLedgerEntries)
			for i := range full {
				full[i] = IssueWatchLedgerEntry{Key: fmt.Sprint(i), Outcome: OutcomeCreated}
			}
			require.NoError(t, r.store.UpdateLedger(r.ctx, "ws-1", w.ID, func([]IssueWatchLedgerEntry) ([]IssueWatchLedgerEntry, error) {
				return full, nil
			}))
		}, ErrorLedgerFull},
		"workflow removed": {func(r *rig, _ IssueWatch) {
			r.host.failNth, r.host.failErr = 1, fmt.Errorf("create task: %w", ErrWorkflowMissing)
		}, ErrorWorkflowMissing},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				r := newRig(t)
				watchIssues(r.gw, 1)
				w := r.saveIssueWatch(t, nil)
				tc.setup(r, w)
				n, _ := r.svc.runIssueWatch(r.ctx, "ws-1", w.ID)
				require.Zero(t, n)
				got := r.issueWatch(t, w.ID)
				require.Equal(t, tc.want, got.LastError)
				require.Equal(t, StateActive, got.State, "the watch stays active")
				require.Nil(t, got.Cursor)
				require.NotContains(t, r.logs.String(), r.conn.apiKey)
			})
		})
	}
}

// BR3.3, BR3.9, NFR6: the watcher ticks every minute and runs each active
// watch whose own interval has passed; it keeps going when the host (its
// store) is not ready yet at the first tick.
func TestIssueWatcher_RunsDueWatchesOnTheirInterval(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		r := newRig(t)
		r.svc.Now = time.Now
		watchIssues(r.gw, 10)
		w := r.saveIssueWatch(t, func(in *IssueWatchInput) { in.IntervalMinutes = minutes(2) })
		paused := r.saveIssueWatch(t, func(in *IssueWatchInput) { in.Name = "Paused" })
		_, err := r.svc.PauseWatch(r.ctx, "ws-1", paused.ID)
		require.NoError(t, err)
		r.state.mu.Lock()
		r.state.failGet = true // the host is not injected yet
		r.state.mu.Unlock()
		y := NewWatcher(r.svc, r.log)
		y.Start()
		defer y.Stop()
		after(time.Minute)
		require.Zero(t, r.host.createCount())
		r.state.mu.Lock()
		r.state.failGet = false
		r.state.mu.Unlock()
		after(time.Minute)
		require.Equal(t, 1, r.host.createCount(), "the first run after the host arrives")
		after(time.Minute)
		require.Equal(t, 1, r.host.createCount(), "not due before 2 minutes")
		after(time.Minute)
		require.Equal(t, 2, r.host.createCount(), "due again")
		require.Equal(t, 2, r.issueWatch(t, w.ID).CreatedCount)
		require.Zero(t, r.issueWatch(t, paused.ID).CreatedCount, "a paused watch never runs")

		require.NoError(t, y.Run(r.ctx, "ws-1", w.ID))
		after(2 * time.Second) // the fake keeps issue searches 1 s apart
		require.Equal(t, 3, r.host.createCount(), "run now does not wait for the interval")
		require.Contains(t, r.logs.String(), `"event":"issue_watch_cycle"`)

		r.conn.set(func(c *fakeConn) { c.disabled = true })
		after(2 * time.Minute)
		require.Equal(t, 3, r.host.createCount(), "BR3.9: a disabled workspace does not run")
		require.Equal(t, StateNotConnected, r.issueWatch(t, w.ID).State)
		r.conn.set(func(c *fakeConn) { c.disabled = false })
		after(time.Minute)
		require.Equal(t, StateActive, r.issueWatch(t, w.ID).State, "re-enabled: the state before comes back")
	})
}

// BR3.9, AC1.8.3: a run does nothing for a watch that is off, gone, paused
// or no longer covered by the connection.
func TestIssueWatchRun_SkipsWatchesThatCannotRun(t *testing.T) {
	cases := map[string]func(r *rig, w IssueWatch){
		"switch off": func(r *rig, _ IssueWatch) { r.conn.set(func(c *fakeConn) { c.disabled = true }) },
		"deleted":    func(r *rig, w IssueWatch) { require.NoError(t, r.svc.DeleteWatch(r.ctx, "ws-1", w.ID)) },
		"paused": func(r *rig, w IssueWatch) {
			_, err := r.svc.PauseWatch(r.ctx, "ws-1", w.ID)
			require.NoError(t, err)
		},
		"project deselected": func(r *rig, _ IssueWatch) {
			r.conn.set(func(c *fakeConn) { c.snap.SelectedProjects = []string{"DEMO"} })
		},
	}
	for name, setup := range cases {
		t.Run(name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				r := newRig(t)
				watchIssues(r.gw, 1)
				w := r.saveIssueWatch(t, nil)
				setup(r, w)
				n, err := r.svc.runIssueWatch(r.ctx, "ws-1", w.ID)
				require.NoError(t, err)
				require.Zero(t, n)
				require.Zero(t, r.gw.count("issues"), "no Backlog search")
			})
		})
	}
	t.Run("switch unreadable", func(t *testing.T) {
		r := newRig(t)
		w := r.saveIssueWatch(t, nil)
		r.conn.set(func(c *fakeConn) { c.enabledErr = connection.ErrStore })
		_, err := r.svc.runIssueWatch(r.ctx, "ws-1", w.ID)
		require.ErrorIs(t, err, connection.ErrStore, "fail closed")
	})
}
