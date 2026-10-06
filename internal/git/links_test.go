package git

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func linkRig(t *testing.T) *rig {
	t.Helper()
	r := newRig(t)
	r.gw.setPRs("PROJ/web-app", prs(50))
	r.gw.setPRs("PROJ/api", prs(5))
	return r
}

func TestU4_Link_ResolvesThePRAndStoresTheLink(t *testing.T) {
	r := linkRig(t)
	l, err := r.svc.Link(r.ctx, ws, "task-17", "https://"+host+"/git/PROJ/api/pullRequests/4", "")
	require.NoError(t, err)
	want := Link{TaskID: "task-17", SpaceHost: host, ProjectKey: "PROJ", RepoName: "api", RepositoryID: 11, Number: 4, Title: "Change 4", Status: StatusActive}
	require.Equal(t, want, l)
	require.Equal(t, []Link{want}, r.links(t))
	require.Equal(t, 1, r.gw.count("PullRequest"))
}

func TestU4_Link_ShortFormsUseTheTaskRepository(t *testing.T) {
	r := linkRig(t)
	l, err := r.svc.Link(r.ctx, ws, "task-17", "42", "repo-k1")
	require.NoError(t, err)
	require.Equal(t, "web-app", l.RepoName)
	l, err = r.svc.Link(r.ctx, ws, "task-17", "api#3", "repo-k1")
	require.NoError(t, err)
	require.Equal(t, "api", l.RepoName)
	for _, repoID := range []string{"", "repo-gh", "repo-missing"} {
		_, err = r.svc.Link(r.ctx, ws, "task-17", "42", repoID)
		require.Equal(t, FieldReference, fieldOf(t, err), repoID)
	}
}

func TestU4_Link_MissingPRIsNotFound(t *testing.T) {
	r := linkRig(t)
	_, err := r.svc.Link(r.ctx, ws, "task-17", "999", "repo-k1")
	require.ErrorIs(t, err, ErrPRNotFound)
	require.Empty(t, r.links(t))
}

func TestU4_Link_RefusesOtherSpacesAndProjects(t *testing.T) {
	r := linkRig(t)
	for _, ref := range []string{"https://other.backlog.com/git/PROJ/api/pullRequests/4", "https://" + host + "/git/DEMO/demo-app/pullRequests/4"} {
		_, err := r.svc.Link(r.ctx, ws, "task-17", ref, "")
		require.Equal(t, FieldReference, fieldOf(t, err), ref)
	}
	require.Zero(t, r.gw.count("PullRequest"))
}

func TestU4_Link_TwiceIsIdempotentAndUnlinkIsLocal(t *testing.T) {
	r := linkRig(t)
	for range 2 {
		_, err := r.svc.Link(r.ctx, ws, "task-17", "42", "repo-k1")
		require.NoError(t, err)
	}
	_, err := r.svc.Link(r.ctx, ws, "task-17", "41", "repo-k1")
	require.NoError(t, err)
	_, err = r.svc.Link(r.ctx, ws, "task-18", "42", "repo-k1")
	require.NoError(t, err)
	require.Len(t, r.links(t), 3)

	calls := r.gw.total()
	require.NoError(t, r.svc.Unlink(r.ctx, ws, "task-17", LinkKey(host, 11, 42)))
	require.Equal(t, calls, r.gw.total(), "unlink writes nothing to Backlog")
	left := r.links(t)
	require.Len(t, left, 2)
	require.ElementsMatch(t, []string{"task-17|41", "task-18|42"}, []string{
		fmt.Sprintf("%s|%d", left[0].TaskID, left[0].Number), fmt.Sprintf("%s|%d", left[1].TaskID, left[1].Number)})
	require.NoError(t, r.svc.Unlink(r.ctx, ws, "task-17", "nope"), "unknown links are a no-op")
}

func TestU4_Link_AssociationsListOnlyActiveLinks(t *testing.T) {
	r := linkRig(t)
	r.addLink(t, Link{TaskID: "task-1", SpaceHost: host, ProjectKey: "PROJ", RepoName: "web-app", RepositoryID: 11, Number: 1, Status: StatusActive})
	r.addLink(t, Link{TaskID: "task-2", SpaceHost: host, ProjectKey: "PROJ", RepoName: "web-app", RepositoryID: 11, Number: 2, Status: StatusNotConnected})
	got, err := r.svc.Associations(r.ctx, ws)
	require.NoError(t, err)
	require.Equal(t, []Association{{ProviderID: ProviderID, TaskID: "task-1", ReviewKey: LinkKey(host, 11, 1),
		ConnectionScope: host, RepositoryID: "11", ChangeRequestNumber: 1}}, got)
}

func TestU4_Link_ParallelWritesKeepEveryLink(t *testing.T) {
	r := linkRig(t)
	var wg sync.WaitGroup
	for i := 1; i <= 10; i++ {
		wg.Go(func() {
			_, err := r.svc.Link(r.ctx, ws, fmt.Sprintf("task-%d", i), fmt.Sprint(i), "repo-k1")
			require.NoError(t, err)
		})
	}
	wg.Wait()
	require.Len(t, r.links(t), 10)
}

func TestU4_Link_LateResultAfterAnEpochChange(t *testing.T) {
	r := linkRig(t)
	r.gw.onPR = func() { r.conn.update(func(c *fakeConn) { c.snap.ConnectionEpoch = 3 }) }
	_, err := r.svc.Link(r.ctx, ws, "task-17", "42", "repo-k1")
	require.ErrorIs(t, err, ErrStale)
	require.Empty(t, r.links(t), "a late result is dropped (AC1.8.3)")
}
