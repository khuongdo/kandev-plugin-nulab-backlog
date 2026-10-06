package git

import (
	"context"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/connection"
)

func TestU4_Watch_SaveResolvesMeAndTheIssue(t *testing.T) {
	r := newRig(t)
	w := r.saveWatch(t, func(in *WatchInput) { in.Assignee, in.Creator, in.IssueKey = WhoMe, WhoMe, "PROJ-120" })
	require.NotEmpty(t, w.ID)
	require.Equal(t, StatusActive, w.State)
	require.Equal(t, host, w.SpaceHost)
	require.EqualValues(t, 1234, w.AssigneeID)
	require.EqualValues(t, 1234, w.CreatedUserID)
	require.EqualValues(t, 5120, w.IssueID)
	require.Equal(t, "wf-1", w.WorkflowID)
	require.Equal(t, "step-1", w.WorkflowStepID)
	require.Equal(t, 1, r.gw.count("Myself"), "one Myself call for both filters")
	require.Equal(t, 1, r.gw.count("Issue"))
	require.Equal(t, []Watch{w}, r.watches(t))
}

func TestU4_Watch_SaveRefusals(t *testing.T) {
	r := newRig(t)
	in := WatchInput{Name: "x", ProjectKey: "PROJ", RepoName: "web-app", Statuses: []string{"open"}, Assignee: WhoAnyone, Creator: WhoAnyone}
	_, err := r.svc.SaveWatch(r.ctx, ws, in)
	require.Equal(t, FieldWorkflow, fieldOf(t, err))
	in.WorkflowID, in.Name = "wf-1", ""
	_, err = r.svc.SaveWatch(r.ctx, ws, in)
	require.Equal(t, FieldName, fieldOf(t, err))
	in.Name, in.IssueKey = "x", "PROJ-999"
	_, err = r.svc.SaveWatch(r.ctx, ws, in)
	require.Equal(t, FieldIssueKey, fieldOf(t, err), "an unknown linked issue")
	in.IssueKey, in.ID = "", "missing"
	_, err = r.svc.SaveWatch(r.ctx, ws, in)
	require.ErrorIs(t, err, ErrNotFound)
	require.Empty(t, r.watches(t))
}

func TestU4_Watch_EditKeepsCountsAndLedger(t *testing.T) {
	r := newRig(t)
	w := r.saveWatch(t, nil)
	require.NoError(t, r.store.UpdateWatches(r.ctx, ws, func(ws []Watch) ([]Watch, error) {
		ws[0].CreatedCount, ws[0].PendingCount = 10, 15
		return ws, nil
	}))
	require.NoError(t, r.store.UpdateLedger(r.ctx, ws, func(l []LedgerEntry) ([]LedgerEntry, error) {
		return append(l, LedgerEntry{Key: LinkKey(host, 11, 1), TaskID: "task-1", WatchID: w.ID}), nil
	}))
	edited := r.saveWatch(t, func(in *WatchInput) { in.ID, in.Name = w.ID, "Renamed" })
	require.Equal(t, "Renamed", edited.Name)
	require.Equal(t, 10, edited.CreatedCount)
	require.Equal(t, 15, edited.PendingCount)
	require.Len(t, r.watches(t), 1)
	ledger, _ := r.store.Ledger(r.ctx, ws)
	require.Len(t, ledger, 1)
}

func TestU4_Watch_PauseStopsRequestsResumeRestarts(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		r := newRig(t)
		r.gw.setPRs("PROJ/web-app", prs(1))
		w := r.saveWatch(t, nil)
		_, err := r.svc.PauseWatch(r.ctx, ws, w.ID)
		require.NoError(t, err)
		watcher := r.startWatcher(t)
		time.Sleep(5*watcher.Every + time.Second)
		synctest.Wait()
		require.Zero(t, r.gw.count("PullRequests"), "a paused watch makes no request over 5 cycles")

		got, err := r.svc.ResumeWatch(r.ctx, ws, w.ID)
		require.NoError(t, err)
		require.Equal(t, StatusActive, got.State)
		time.Sleep(watcher.Every)
		synctest.Wait()
		require.Equal(t, 1, r.gw.count("PullRequests"))
		require.Equal(t, 1, r.host.taskCount())
	})
}

func TestU4_Watch_DeleteKeepsLedgerAndTasks(t *testing.T) {
	r := newRig(t)
	w := r.saveWatch(t, nil)
	require.NoError(t, r.store.UpdateLedger(r.ctx, ws, func(l []LedgerEntry) ([]LedgerEntry, error) {
		return append(l, LedgerEntry{Key: "k", TaskID: "task-1", WatchID: w.ID}), nil
	}))
	require.NoError(t, r.svc.DeleteWatch(r.ctx, ws, w.ID))
	require.Empty(t, r.watches(t))
	ledger, _ := r.store.Ledger(r.ctx, ws)
	require.Len(t, ledger, 1)
	require.ErrorIs(t, r.svc.DeleteWatch(r.ctx, ws, w.ID), ErrNotFound)
}

func TestU4_Watch_NotConnectedCannotResume(t *testing.T) {
	r := newRig(t)
	w := r.saveWatch(t, nil)
	r.svc.OnConnectionChanged(context.Background(), connection.ConnectionChanged{WorkspaceID: ws, Reason: connection.ReasonDisconnected, ConnectionEpoch: 3})
	_, err := r.svc.ResumeWatch(r.ctx, ws, w.ID)
	require.ErrorIs(t, err, ErrConflict)
	_, err = r.svc.PauseWatch(r.ctx, ws, "missing")
	require.ErrorIs(t, err, ErrNotFound)
	list, err := r.svc.ListWatches(r.ctx, ws)
	require.NoError(t, err)
	require.Equal(t, StatusNotConnected, list[0].State)
}
