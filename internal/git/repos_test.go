package git

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/backlog"
	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/connection"
)

func TestU4_Repos_ListsOnlySelectedProjects(t *testing.T) {
	r := newRig(t)
	page, err := r.svc.ListRepositories(r.ctx, ws, "", "")
	require.NoError(t, err)
	require.Empty(t, page.NextCursor)
	require.Equal(t, []RepositoryInspection{
		{ProviderID: ProviderID, ProviderHost: "https://" + host, ProviderScope: host, OwnerOrProject: "PROJ",
			RepositoryID: "11", RepositoryName: "web-app", CloneURL: "https://" + host + "/git/PROJ/web-app.git"},
		{ProviderID: ProviderID, ProviderHost: "https://" + host, ProviderScope: host, OwnerOrProject: "PROJ",
			RepositoryID: "12", RepositoryName: "api", CloneURL: "https://" + host + "/git/PROJ/api.git"},
	}, page.Repositories, "DEMO is not selected (AC1.7.1)")
	require.Equal(t, 1, r.gw.count("Repositories"))
}

func TestU4_Repos_QueryFiltersByName(t *testing.T) {
	r := newRig(t)
	page, err := r.svc.ListRepositories(r.ctx, ws, " WEB ", "")
	require.NoError(t, err)
	require.Len(t, page.Repositories, 1)
	require.Equal(t, "web-app", page.Repositories[0].RepositoryName)
}

func TestU4_Repos_CursorIsBoundToQueryAndScope(t *testing.T) {
	r := newRig(t)
	var many []backlog.Repository
	for i := range 60 {
		many = append(many, backlog.Repository{ID: int64(100 + i), Name: fmt.Sprintf("repo-%02d", i), HTTPURL: "https://" + host + "/git/PROJ/x.git"})
	}
	r.gw.repos["PROJ"] = many
	first, err := r.svc.ListRepositories(r.ctx, ws, "repo", "")
	require.NoError(t, err)
	require.Len(t, first.Repositories, 50)
	require.NotEmpty(t, first.NextCursor)
	second, err := r.svc.ListRepositories(r.ctx, ws, "repo", first.NextCursor)
	require.NoError(t, err)
	require.Len(t, second.Repositories, 10)
	require.Empty(t, second.NextCursor)

	calls := r.gw.count("Repositories")
	for name, tc := range map[string]struct{ query, cursor string }{
		"another query": {"other", first.NextCursor},
		"garbage":       {"repo", "not-a-cursor"},
	} {
		_, err := r.svc.ListRepositories(r.ctx, ws, tc.query, tc.cursor)
		require.Equal(t, FieldCursor, fieldOf(t, err), name)
	}
	r.conn.update(func(c *fakeConn) { c.snap.SelectedProjects = []string{"PROJ", "DEMO"} })
	_, err = r.svc.ListRepositories(r.ctx, ws, "repo", first.NextCursor)
	require.Equal(t, FieldCursor, fieldOf(t, err), "another project scope")
	require.Equal(t, calls, r.gw.count("Repositories"), "a bad cursor is refused before any request")
}

func TestU4_Repos_ErrorsPassThrough(t *testing.T) {
	r := newRig(t)
	rl := backlogErr(backlog.KindRateLimited, 429)
	r.gw.setErr("Repositories", rl)
	_, err := r.svc.ListRepositories(r.ctx, ws, "", "")
	require.ErrorIs(t, err, rl)
	r.conn.update(func(c *fakeConn) { c.err = connection.ErrNotConnected })
	_, err = r.svc.ListRepositories(r.ctx, ws, "", "")
	require.ErrorIs(t, err, connection.ErrNotConnected)
}

func TestU4_Repos_InspectReturnsTheDescriptor(t *testing.T) {
	r := newRig(t)
	r.gw.setPRs("PROJ/web-app", []backlog.PullRequest{{Number: 1, RepositoryID: 11, Base: "develop"}, {Number: 2, RepositoryID: 11, Base: "develop"}})
	for _, raw := range []string{"https://" + host + "/git/PROJ/web-app.git", "https://" + host + "/git/PROJ/web-app"} {
		d, err := r.svc.Inspect(r.ctx, ws, raw)
		require.NoError(t, err)
		require.Equal(t, &Descriptor{ProviderID: ProviderID, ProviderHost: "https://" + host, ProviderScope: host,
			ProviderRepositoryID: "11", OwnerOrProject: "PROJ", Name: "web-app",
			CloneURL: "https://" + host + "/git/PROJ/web-app.git", DefaultBranch: "develop"}, d, raw)
	}
	for _, raw := range []string{
		"https://other.backlog.com/git/PROJ/web-app.git",
		"https://" + host + "/git/DEMO/demo-app.git",
		"https://github.com/acme/web.git",
		"https://" + host + "/git/PROJ/missing.git",
		"https://" + host + "/view/PROJ-1",
		"not a url",
	} {
		d, err := r.svc.Inspect(r.ctx, ws, raw)
		require.NoError(t, err, raw)
		require.Nil(t, d, raw+" is matched:false")
	}
}

func TestU4_Repos_BranchesFromRecentPRs(t *testing.T) {
	r := newRig(t)
	d := Descriptor{ProviderID: ProviderID, ProviderHost: "https://" + host, ProviderScope: host, ProviderRepositoryID: "11", OwnerOrProject: "PROJ", Name: "web-app"}
	branches, err := r.svc.Branches(r.ctx, ws, d)
	require.NoError(t, err)
	require.Empty(t, branches)
	require.NotNil(t, branches, "an empty list, not null")

	r.gw.setPRs("PROJ/web-app", prs(3))
	branches, err = r.svc.Branches(r.ctx, ws, d)
	require.NoError(t, err)
	require.Equal(t, Branch{Name: "main", IsDefault: true}, branches[0])
	require.Len(t, branches, 4)
	require.Equal(t, 100, r.gw.queries[len(r.gw.queries)-1].Count, "the newest 100 PRs")

	for name, mut := range map[string]func(*Descriptor){
		"another provider":   func(d *Descriptor) { d.ProviderID = "github" },
		"another host":       func(d *Descriptor) { d.ProviderScope = "other.backlog.com" },
		"unselected project": func(d *Descriptor) { d.OwnerOrProject = "DEMO" },
	} {
		bad := d
		mut(&bad)
		_, err := r.svc.Branches(r.ctx, ws, bad)
		var fe *connection.FieldError
		require.True(t, errors.As(err, &fe), name)
		require.Equal(t, FieldRepository, fe.Field, name)
	}
}
