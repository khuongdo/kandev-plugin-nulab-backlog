package git

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/backlog"
	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/connection"
)

func saveQuery(t *testing.T, r *rig, mut func(*QueryInput)) Query {
	t.Helper()
	in := QueryInput{Name: "Open PRs", ProjectKey: "PROJ", RepoName: "web-app", Statuses: []string{"open"}, Assignee: WhoAnyone}
	if mut != nil {
		mut(&in)
	}
	q, err := r.svc.SaveQuery(r.ctx, ws, in)
	require.NoError(t, err)
	return q
}

func TestU4_Queries_SaveListDelete(t *testing.T) {
	r := newRig(t)
	q := saveQuery(t, r, nil)
	require.NotEmpty(t, q.ID)
	edited := saveQuery(t, r, func(in *QueryInput) { in.ID, in.Name = q.ID, "Renamed" })
	require.Equal(t, q.ID, edited.ID)
	list, err := r.svc.ListQueries(r.ctx, ws)
	require.NoError(t, err)
	require.Equal(t, []Query{edited}, list)
	require.NoError(t, r.svc.DeleteQuery(r.ctx, ws, q.ID))
	list, err = r.svc.ListQueries(r.ctx, ws)
	require.NoError(t, err)
	require.Empty(t, list)
	require.ErrorIs(t, r.svc.DeleteQuery(r.ctx, ws, q.ID), ErrNotFound)
	_, err = r.svc.SaveQuery(r.ctx, ws, QueryInput{Name: ""})
	require.Equal(t, FieldName, fieldOf(t, err))
}

func TestU4_Queries_RunReturnsAtMostTwentyRows(t *testing.T) {
	r := newRig(t)
	r.gw.setPRs("PROJ/web-app", prs(30))
	r.addLink(t, Link{TaskID: "task-9", SpaceHost: host, ProjectKey: "PROJ", RepoName: "web-app", RepositoryID: 11, Number: 30, Status: StatusActive})
	q := saveQuery(t, r, func(in *QueryInput) { in.Assignee = WhoMe })
	rows, err := r.svc.RunQuery(r.ctx, ws, q.ID)
	require.NoError(t, err)
	require.Len(t, rows, 20)
	require.Equal(t, QueryRow{Number: 30, Title: "Change 30", State: "open", Assignee: "Lan", RepoName: "web-app",
		URL: "https://" + host + "/git/PROJ/web-app/pullRequests/30", LinkedTaskIDs: []string{"task-9"}}, rows[0])
	require.Equal(t, []string{}, rows[1].LinkedTaskIDs)
	last := r.gw.queries[len(r.gw.queries)-1]
	require.Equal(t, 20, last.Count)
	require.Equal(t, []int64{1}, last.StatusIDs)
	require.Equal(t, []int64{1234}, last.AssigneeIDs, "me is resolved with Myself")
}

func TestU4_Queries_EmptyIsAnEmptyList(t *testing.T) {
	r := newRig(t)
	q := saveQuery(t, r, nil)
	rows, err := r.svc.RunQuery(r.ctx, ws, q.ID)
	require.NoError(t, err)
	require.NotNil(t, rows)
	require.Empty(t, rows)
	_, err = r.svc.RunQuery(r.ctx, ws, "missing")
	require.ErrorIs(t, err, ErrNotFound)
}

func TestU4_Queries_ErrorsKeepTheirCodes(t *testing.T) {
	r := newRig(t)
	q := saveQuery(t, r, nil)
	for _, e := range []error{backlogErr(backlog.KindRateLimited, 429), backlogErr(backlog.KindUnreachable, 503)} {
		r.gw.setErr("PullRequests", e)
		_, err := r.svc.RunQuery(r.ctx, ws, q.ID)
		require.ErrorIs(t, err, e)
	}
	r.conn.update(func(c *fakeConn) { c.err = connection.ErrNotConnected })
	_, err := r.svc.RunQuery(r.ctx, ws, q.ID)
	require.ErrorIs(t, err, connection.ErrNotConnected)
}

func TestU4_Queries_UnselectedProjectIsRefused(t *testing.T) {
	r := newRig(t)
	_, err := r.svc.SaveQuery(r.ctx, ws, QueryInput{Name: "x", ProjectKey: "DEMO", RepoName: "demo-app", Statuses: []string{"open"}, Assignee: WhoAnyone})
	require.Equal(t, FieldRepository, fieldOf(t, err))
}

func TestQueries_SetDefaultMovesTheStarAndSaveKeepsIt(t *testing.T) {
	r := newRig(t)
	a := saveQuery(t, r, func(in *QueryInput) { in.Name = "A" })
	b := saveQuery(t, r, func(in *QueryInput) { in.Name = "B"; in.IsDefault = true })
	require.False(t, b.IsDefault, "save never sets the star")
	defaults := func(list []Query) (out []string) {
		for _, q := range list {
			if q.IsDefault {
				out = append(out, q.ID)
			}
		}
		return out
	}
	list, err := r.svc.SetQueryDefault(r.ctx, ws, a.ID, true)
	require.NoError(t, err)
	require.Equal(t, []string{a.ID}, defaults(list))
	list, err = r.svc.SetQueryDefault(r.ctx, ws, b.ID, true)
	require.NoError(t, err)
	require.Equal(t, []string{b.ID}, defaults(list), "at most one default")
	saveQuery(t, r, func(in *QueryInput) { in.ID, in.Name = b.ID, "B renamed" })
	list, err = r.svc.ListQueries(r.ctx, ws)
	require.NoError(t, err)
	require.Equal(t, []string{b.ID}, defaults(list), "an edit keeps the star")
	list, err = r.svc.SetQueryDefault(r.ctx, ws, b.ID, false)
	require.NoError(t, err)
	require.Empty(t, defaults(list))
	_, err = r.svc.SetQueryDefault(r.ctx, ws, "missing", true)
	require.ErrorIs(t, err, ErrNotFound)
}

func TestQueries_OldDocumentWithoutIsDefaultStillReads(t *testing.T) {
	r := newRig(t)
	r.state.set("workspace/"+ws+"/git.queries", map[string]any{"schemaVersion": float64(1), "items": []any{
		map[string]any{"id": "q1", "name": "Old", "projectKey": "PROJ", "repoName": "web-app", "statuses": []any{"open"}, "assignee": "anyone"}}})
	list, err := r.svc.ListQueries(r.ctx, ws)
	require.NoError(t, err)
	require.Len(t, list, 1)
	require.False(t, list[0].IsDefault)
}
