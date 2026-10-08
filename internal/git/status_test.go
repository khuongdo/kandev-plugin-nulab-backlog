package git

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/backlog"
	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/connection"
)

func statusRig(t *testing.T, statusID int, assignee string) *rig {
	t.Helper()
	r := newRig(t)
	r.gw.setPRs("PROJ/web-app", []backlog.PullRequest{{RepositoryID: 11, Number: 42, Summary: "Add login page",
		Base: "main", Branch: "feature/login", StatusID: statusID, AssigneeName: assignee}})
	r.addLink(t, Link{TaskID: "task-17", SpaceHost: host, ProjectKey: "PROJ", RepoName: "web-app", RepositoryID: 11, Number: 42, Status: StatusActive})
	return r
}

func TestU4_Status_OneSummaryPerLinkForEachState(t *testing.T) {
	for _, tc := range []struct {
		statusID int
		assignee string
		state    string
		label    string
	}{
		{1, "Lan", "open", "Open – Lan"},
		{2, "", "closed", "Closed"},
		{3, "Lan", "merged", "Merged – Lan"},
	} {
		t.Run(tc.state, func(t *testing.T) {
			r := statusRig(t, tc.statusID, tc.assignee)
			got, err := r.svc.Status(r.ctx, ws, "task-17")
			require.NoError(t, err)
			require.Equal(t, []ReviewSummary{{
				ProviderID: ProviderID, ReviewKey: LinkKey(host, 11, 42), Title: "Add login page",
				URL: "https://" + host + "/git/PROJ/web-app/pullRequests/42", ConnectionScope: host, RepositoryID: "11",
				ChangeRequestNumber: 42, State: tc.state, StatusBadge: &StatusBadge{Label: tc.label},
				TaskStatus: &TaskStatus{Number: 42, State: tc.state, PipelineState: "neutral", Checks: []any{}},
				Base:       "main", Branch: "feature/login", Assignee: tc.assignee,
			}}, got)
		})
	}
}

func TestU4_Status_FailedFetchIsStatusUnknown(t *testing.T) {
	r := statusRig(t, 1, "Lan")
	r.addLink(t, Link{TaskID: "task-17", SpaceHost: host, ProjectKey: "PROJ", RepoName: "web-app", RepositoryID: 11, Number: 7, Status: StatusActive})
	got, err := r.svc.Status(r.ctx, ws, "task-17")
	require.NoError(t, err)
	require.Len(t, got, 2, "the other link is still returned")
	require.Equal(t, "Open – Lan", got[0].StatusBadge.Label)
	require.Equal(t, "Status unknown", got[1].StatusBadge.Label)
	require.Equal(t, "not_found", got[1].TaskStatus.Error)
}

func TestU4_Status_NotConnectedLinksMakeNoCall(t *testing.T) {
	r := statusRig(t, 1, "Lan")
	require.NoError(t, r.store.UpdateLinks(r.ctx, ws, func(ls []Link) ([]Link, error) {
		ls[0].Status = StatusNotConnected
		return ls, nil
	}))
	got, err := r.svc.Status(r.ctx, ws, "task-17")
	require.NoError(t, err)
	require.Equal(t, "Not connected", got[0].StatusBadge.Label)
	require.Equal(t, "not_connected", got[0].TaskStatus.Error)
	require.Zero(t, r.gw.count("PullRequest"))

	r = statusRig(t, 1, "Lan")
	r.conn.update(func(c *fakeConn) { c.err = connection.ErrNotConnected })
	got, err = r.svc.Status(r.ctx, ws, "task-17")
	require.NoError(t, err)
	require.Equal(t, "Not connected", got[0].StatusBadge.Label)
	require.Zero(t, r.gw.count("PullRequest"))
}

func TestU4_Status_OnlyThatTasksLinks(t *testing.T) {
	r := statusRig(t, 1, "Lan")
	got, err := r.svc.Status(r.ctx, ws, "task-99")
	require.NoError(t, err)
	require.Empty(t, got)
	require.NotNil(t, got)
}

// TestFR5_Status_EveryLinkedPRCarriesTaskStatus: the host shows a Backlog PR in
// the task top bar only when its summary carries taskStatus; one PR is one
// button, two or more are grouped into the host dropdown (FR5.4).
func TestFR5_Status_EveryLinkedPRCarriesTaskStatus(t *testing.T) {
	for _, n := range []int{1, 2} {
		t.Run(fmt.Sprintf("%d linked PRs", n), func(t *testing.T) {
			r := newRig(t)
			r.gw.setPRs("PROJ/web-app", []backlog.PullRequest{
				{RepositoryID: 11, Number: 42, Summary: "Add login page", StatusID: 1},
				{RepositoryID: 11, Number: 43, Summary: "Fix logout", StatusID: 3},
			})
			for i := range n {
				r.addLink(t, Link{TaskID: "task-17", SpaceHost: host, ProjectKey: "PROJ", RepoName: "web-app",
					RepositoryID: 11, Number: 42 + i, Status: StatusActive})
			}
			got, err := r.svc.Status(r.ctx, ws, "task-17")
			require.NoError(t, err)
			require.Len(t, got, n)
			for _, s := range got {
				require.NotNil(t, s.TaskStatus, "PR %d", s.ChangeRequestNumber)
				require.Equal(t, s.ChangeRequestNumber, s.TaskStatus.Number)
				require.Equal(t, s.State, s.TaskStatus.State)
			}
		})
	}
}
