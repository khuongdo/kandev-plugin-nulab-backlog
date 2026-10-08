package issues

import (
	"encoding/json"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestU3_Links_SearchTasks(t *testing.T) {
	r := newRig(t)
	for i := range 30 {
		r.host.tasks = append(r.host.tasks, TaskInfo{ID: fmt.Sprintf("bulk-%d", i), Key: fmt.Sprintf("T-%d", 100+i), Title: "Bulk login"})
	}
	r.link(t, Link{IssueKey: "PROJ-118", TaskID: "task-17", TaskKey: "T-17"})
	got, err := r.svc.SearchTasks(r.ctx, "ws-1", "LOGIN")
	require.NoError(t, err)
	require.Len(t, got, 20, "at most 20")
	require.Equal(t, TaskItem{TaskID: "task-17", TaskKey: "T-17", Title: "Login work", LinkedIssueKey: "PROJ-118"}, got[0])
	got, err = r.svc.SearchTasks(r.ctx, "ws-1", "t-18")
	require.NoError(t, err)
	require.Equal(t, []TaskItem{{TaskID: "task-18", TaskKey: "T-18", Title: "Other"}}, got, "by task key too")
	got, err = r.svc.SearchTasks(r.ctx, "ws-1", "nothing")
	require.NoError(t, err)
	require.Empty(t, got)
	require.NotNil(t, got)
}

func TestU3_Links_LinkChecksTheIssue(t *testing.T) {
	r := newRig(t)
	l, err := r.svc.Link(r.ctx, "ws-1", "task-17", "PROJ-118")
	require.NoError(t, err)
	require.Equal(t, "T-17", l.TaskKey)
	require.Equal(t, "In Progress", l.LastKnownStatus)
	require.Equal(t, 1, r.gw.count("issue"))

	_, err = r.svc.Link(r.ctx, "ws-1", "task-18", "PROJ-999")
	require.ErrorIs(t, err, ErrNotFound)
	_, err = r.svc.Link(r.ctx, "ws-1", "task-18", "OTHER-1")
	require.Equal(t, FieldIssueKey, fieldOf(t, err), "only selected projects")
	_, err = r.svc.Link(r.ctx, "ws-1", "", "PROJ-118")
	require.Equal(t, FieldTaskID, fieldOf(t, err))
}

func TestU3_Links_SecondTaskSameIssue(t *testing.T) {
	r := newRig(t)
	_, err := r.svc.Link(r.ctx, "ws-1", "task-17", "PROJ-118")
	require.NoError(t, err)
	_, err = r.svc.Link(r.ctx, "ws-1", "task-18", "PROJ-118")
	require.NoError(t, err, "AC3.3.2")
	require.Len(t, r.links(t), 2)
}

func TestU3_Links_OneIssuePerTask(t *testing.T) {
	r := newRig(t)
	_, err := r.svc.Link(r.ctx, "ws-1", "task-17", "PROJ-118")
	require.NoError(t, err)
	_, err = r.svc.Link(r.ctx, "ws-1", "task-17", "PROJ-120")
	require.ErrorIs(t, err, ErrConflict, "AC3.3.3, FR3.4")
	again, err := r.svc.Link(r.ctx, "ws-1", "task-17", "PROJ-118")
	require.NoError(t, err, "the same pair is idempotent")
	require.Equal(t, "PROJ-118", again.IssueKey)
	require.Len(t, r.links(t), 1)
}

func TestU3_Links_UnlinkOnlyThatTask(t *testing.T) {
	r := newRig(t)
	r.link(t, Link{IssueKey: "PROJ-118", TaskID: "task-17"})
	r.link(t, Link{IssueKey: "PROJ-118", TaskID: "task-18"})
	require.NoError(t, r.svc.Unlink(r.ctx, "ws-1", "task-17"))
	ls := r.links(t)
	require.Len(t, ls, 1)
	require.Equal(t, "task-18", ls[0].TaskID)
	require.Zero(t, r.gw.total(), "AC3.3.5: no Backlog write, no Backlog call")
	require.ErrorIs(t, r.svc.Unlink(r.ctx, "ws-1", "task-17"), ErrNotFound)
}

func TestU3_Links_ListsBadges(t *testing.T) {
	r := newRig(t)
	r.link(t, Link{IssueKey: "PROJ-118", TaskID: "task-17", TaskKey: "T-17", LastKnownStatus: "Open", StatusUpdatedAt: "2026-10-06T00:00:00Z", FailCount: 3})
	r.link(t, Link{IssueKey: "PROJ-120", TaskID: "task-18", State: StateNotConnected, Unavailable: true})
	got, err := r.svc.Links(r.ctx, "ws-1")
	require.NoError(t, err)
	require.Equal(t, []LinkView{
		{TaskID: "task-17", TaskKey: "T-17", IssueKey: "PROJ-118", SpaceHost: spaceHost, State: StateActive, Status: "Open",
			StatusUpdatedAt: "2026-10-06T00:00:00Z", Stale: true, URL: "https://" + spaceHost + "/view/PROJ-118"},
		{TaskID: "task-18", IssueKey: "PROJ-120", SpaceHost: spaceHost, State: StateNotConnected, Unavailable: true,
			URL: "https://" + spaceHost + "/view/PROJ-120"},
	}, got)
	empty, err := (newRig(t)).svc.Links(r.ctx, "ws-1")
	require.NoError(t, err)
	require.NotNil(t, empty)
}

func TestU3_Links_ParallelLinksAreKept(t *testing.T) {
	r := newRig(t)
	for i := range 10 {
		id := fmt.Sprintf("p-%d", i)
		r.host.tasks = append(r.host.tasks, TaskInfo{ID: id, Key: id})
	}
	var wg sync.WaitGroup
	for i := range 10 {
		wg.Go(func() {
			_, err := r.svc.Link(r.ctx, "ws-1", fmt.Sprintf("p-%d", i), "PROJ-118")
			require.NoError(t, err)
		})
	}
	wg.Wait()
	require.Len(t, r.links(t), 10)
}

func TestFR2_Links_ViewCarriesTheIssueSummary(t *testing.T) {
	r := newRig(t)
	l, err := r.svc.Link(r.ctx, "ws-1", "task-17", "PROJ-118")
	require.NoError(t, err)
	require.Equal(t, "Fix login timeout", l.Summary, "FR2.1: stored when the link is made")
	got, err := r.svc.Links(r.ctx, "ws-1")
	require.NoError(t, err)
	require.Len(t, got, 1)
	require.Equal(t, "Fix login timeout", got[0].Summary, "FR2.3")
}

func TestFR2_Links_OldStoredLinkWithoutSummaryStillReads(t *testing.T) {
	var l Link
	require.NoError(t, json.Unmarshal([]byte(`{"issueKey":"PROJ-12","taskId":"task-1","state":"active","lastKnownStatus":"Open","connectionEpoch":1,"createdAt":"2026-10-01T00:00:00Z"}`), &l))
	require.Equal(t, "PROJ-12", l.IssueKey)
	require.Empty(t, l.Summary, "NFR5: the new field is optional")
	b, err := json.Marshal(LinkView{TaskID: "task-1", IssueKey: "PROJ-12"})
	require.NoError(t, err)
	require.NotContains(t, string(b), "summary", "omitted when unknown")
}
