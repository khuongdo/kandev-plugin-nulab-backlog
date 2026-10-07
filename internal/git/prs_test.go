package git

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/backlog"
)

func prListInput(page int) PRListInput {
	return PRListInput{ProjectKey: "PROJ", RepoName: "web-app", Page: page}
}

// FR4.1, BR4.1: 45 PRs are pages of 20, 20 and 5; the total comes from the
// count call and hasNext from page*20 < total.
func TestPRList_PagesOfTwenty(t *testing.T) {
	r := newRig(t)
	r.gw.setPRs("PROJ/web-app", prs(45))
	for page, want := range map[int]struct {
		n, first int
		next     bool
	}{1: {20, 45, true}, 2: {20, 25, true}, 3: {5, 5, false}} {
		got, err := r.svc.ListPullRequests(r.ctx, ws, prListInput(page))
		require.NoError(t, err)
		require.Len(t, got.Items, want.n, "page %d", page)
		require.Equal(t, want.first, got.Items[0].Number, "newest first, page %d", page)
		require.Equal(t, PullRequestPage{Items: got.Items, Page: page, PageSize: 20, Total: 45, HasNext: want.next}, got)
	}
	for _, q := range r.gw.queries {
		require.Equal(t, 20, q.Count)
	}
	require.Equal(t, 3, r.gw.count("PullRequestCount"))
}

// FR4.2, BR4.2: each row has its fields and the task linked to it.
func TestPRList_RowsCarryTheLinkedTask(t *testing.T) {
	r := newRig(t)
	r.gw.setPRs("PROJ/web-app", []backlog.PullRequest{{RepositoryID: 11, Number: 42, Summary: "Add login", StatusID: 3,
		AuthorName: "Test User", AssigneeName: "Lan", Updated: "2026-10-02T09:00:00Z"}, {RepositoryID: 11, Number: 41, Summary: "Typo", StatusID: 1}})
	r.addLink(t, Link{TaskID: "task-9", SpaceHost: host, ProjectKey: "PROJ", RepoName: "web-app", RepositoryID: 11, Number: 42, Status: StatusActive})
	r.addLink(t, Link{TaskID: "task-8", SpaceHost: host, ProjectKey: "PROJ", RepoName: "web-app", RepositoryID: 11, Number: 41, Status: StatusNotConnected})
	got, err := r.svc.ListPullRequests(r.ctx, ws, prListInput(1))
	require.NoError(t, err)
	require.Equal(t, PullRequestRow{Number: 42, Title: "Add login", Status: "merged", Author: "Test User", Assignee: "Lan",
		Updated: "2026-10-02T09:00:00Z", URL: "https://" + host + "/git/PROJ/web-app/pullRequests/42",
		LinkedTaskIDs: []string{"task-9"}}, got.Items[0])
	require.Equal(t, []string{}, got.Items[1].LinkedTaskIDs, "only active links count")
}

func TestPRList_FiltersAndMe(t *testing.T) {
	r := newRig(t)
	in := prListInput(1)
	in.Statuses, in.Assignee, in.Creator = []string{"open", "merged"}, WhoMe, WhoMe
	_, err := r.svc.ListPullRequests(r.ctx, ws, in)
	require.NoError(t, err)
	q := r.gw.queries[0]
	require.Equal(t, backlog.PullRequestQuery{StatusIDs: []int64{1, 3}, AssigneeIDs: []int64{1234}, CreatedUserIDs: []int64{1234}, Count: 20}, q)
	require.Equal(t, []backlog.CallClass{backlog.Interactive}, r.gw.classes)
	require.Equal(t, 1, r.gw.count("Myself"), "me is resolved once")
}

func TestPRList_RejectsBadInput(t *testing.T) {
	r := newRig(t)
	for field, in := range map[string]PRListInput{
		FieldRepository: {ProjectKey: "DEMO", RepoName: "demo-app", Page: 1},
		FieldPage:       {ProjectKey: "PROJ", RepoName: "web-app", Page: 0},
		FieldStatuses:   {ProjectKey: "PROJ", RepoName: "web-app", Page: 1, Statuses: []string{"draft"}},
		FieldCreator:    {ProjectKey: "PROJ", RepoName: "web-app", Page: 1, Creator: "lan"},
	} {
		_, err := r.svc.ListPullRequests(r.ctx, ws, in)
		require.Equal(t, field, fieldOf(t, err), field)
	}
	require.Zero(t, r.gw.total(), "nothing is sent for a bad input")
	r.gw.setErr("PullRequestCount", backlogErr(backlog.KindRateLimited, 429))
	_, err := r.svc.ListPullRequests(r.ctx, ws, prListInput(1))
	require.Error(t, err)
}
