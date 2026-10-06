package git

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestU4_Impact_CountsActiveItems(t *testing.T) {
	r := eventRig(t)
	r.addLink(t, Link{TaskID: "task-3", SpaceHost: host, ProjectKey: "PROJ", RepoName: "web-app", RepositoryID: 11, Number: 3, Status: StatusNotConnected})
	got, err := r.svc.Impact(r.ctx, ws, nil)
	require.NoError(t, err)
	require.Equal(t, Impact{PRLinks: 2, PRWatches: 2}, got)
	got, err = r.svc.Impact(r.ctx, ws, []string{"DEMO"})
	require.NoError(t, err)
	require.Equal(t, Impact{PRLinks: 1, PRWatches: 1}, got)
}

func TestU4_Impact_EmptyWorkspaceAndPausedWatches(t *testing.T) {
	r := newRig(t)
	got, err := r.svc.Impact(r.ctx, "ws-empty", nil)
	require.NoError(t, err)
	require.Equal(t, Impact{}, got)
	w := r.saveWatch(t, nil)
	_, err = r.svc.PauseWatch(r.ctx, ws, w.ID)
	require.NoError(t, err)
	got, err = r.svc.Impact(r.ctx, ws, nil)
	require.NoError(t, err)
	require.Equal(t, 1, got.PRWatches, "a paused watch is still turned off by a disconnect")
}
