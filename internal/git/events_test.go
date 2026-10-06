package git

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/connection"
)

// eventRig has links and active watches on PROJ and DEMO of host.
func eventRig(t *testing.T) *rig {
	t.Helper()
	r := newRig(t)
	r.conn.update(func(c *fakeConn) { c.snap.SelectedProjects = []string{"PROJ", "DEMO"} })
	r.addLink(t, Link{TaskID: "task-1", SpaceHost: host, ProjectKey: "PROJ", RepoName: "web-app", RepositoryID: 11, Number: 1, Status: StatusActive})
	r.addLink(t, Link{TaskID: "task-2", SpaceHost: host, ProjectKey: "DEMO", RepoName: "demo-app", RepositoryID: 21, Number: 2, Status: StatusActive})
	r.saveWatch(t, nil)
	r.saveWatch(t, func(in *WatchInput) { in.ProjectKey, in.RepoName = "DEMO", "demo-app" })
	return r
}

func states(r *rig, t *testing.T) (links, watches []string) {
	for _, l := range r.links(t) {
		links = append(links, l.ProjectKey+":"+l.Status)
	}
	for _, w := range r.watches(t) {
		watches = append(watches, w.ProjectKey+":"+w.State)
	}
	return links, watches
}

func (r *rig) change(e connection.ConnectionChanged) {
	e.WorkspaceID = ws
	r.svc.OnConnectionChanged(context.Background(), e)
}

func TestU4_Events_DisconnectTurnsEverythingOff(t *testing.T) {
	r := eventRig(t)
	r.change(connection.ConnectionChanged{Reason: connection.ReasonDisconnected, ConnectionEpoch: 3})
	links, watches := states(r, t)
	require.Equal(t, []string{"PROJ:not_connected", "DEMO:not_connected"}, links)
	require.Equal(t, []string{"PROJ:not_connected", "DEMO:not_connected"}, watches)
	r.conn.update(func(c *fakeConn) { c.err = connection.ErrNotConnected })
	calls := r.gw.total()
	w := NewWatcher(r.svc, nil)
	w.cycleAll(context.Background())
	require.Equal(t, calls, r.gw.total(), "the next cycle makes 0 requests")
}

func TestU4_Events_SpaceChangeTurnsOldHostItemsOff(t *testing.T) {
	r := eventRig(t)
	r.change(connection.ConnectionChanged{Reason: connection.ReasonSpaceChanged, ConnectionEpoch: 3, SpaceHost: "other.backlog.jp"})
	links, watches := states(r, t)
	require.Equal(t, []string{"PROJ:not_connected", "DEMO:not_connected"}, links)
	require.Equal(t, []string{"PROJ:not_connected", "DEMO:not_connected"}, watches)
}

func TestU4_Events_ProjectChangeTouchesOnlyDeselected(t *testing.T) {
	r := eventRig(t)
	r.change(connection.ConnectionChanged{Reason: connection.ReasonProjectsChanged, ConnectionEpoch: 3, SpaceHost: host, SelectedProjects: []string{"PROJ"}})
	links, watches := states(r, t)
	require.Equal(t, []string{"PROJ:active", "DEMO:not_connected"}, links)
	require.Equal(t, []string{"PROJ:active", "DEMO:not_connected"}, watches)
}

func TestU4_Events_RestoreBringsLinksBackAndPausesWatches(t *testing.T) {
	r := eventRig(t)
	r.change(connection.ConnectionChanged{Reason: connection.ReasonDisconnected, ConnectionEpoch: 3})
	r.change(connection.ConnectionChanged{Reason: connection.ReasonConnected, ConnectionEpoch: 4, SpaceHost: host,
		SelectedProjects: []string{"PROJ"}, Restore: true})
	links, watches := states(r, t)
	require.Equal(t, []string{"PROJ:active", "DEMO:not_connected"}, links)
	require.Equal(t, []string{"PROJ:paused", "DEMO:not_connected"}, watches, "M12: restored watches are Paused")

	r.change(connection.ConnectionChanged{Reason: connection.ReasonProjectsChanged, ConnectionEpoch: 5, SpaceHost: host,
		SelectedProjects: []string{"PROJ", "DEMO"}, Restore: true})
	links, watches = states(r, t)
	require.Equal(t, []string{"PROJ:active", "DEMO:active"}, links)
	require.Equal(t, []string{"PROJ:paused", "DEMO:paused"}, watches)
}

func TestU4_Events_ConnectWithoutRestoreKeepsItemsOff(t *testing.T) {
	r := eventRig(t)
	r.change(connection.ConnectionChanged{Reason: connection.ReasonDisconnected, ConnectionEpoch: 3})
	r.change(connection.ConnectionChanged{Reason: connection.ReasonConnected, ConnectionEpoch: 4, SpaceHost: host, SelectedProjects: []string{"PROJ"}})
	links, _ := states(r, t)
	require.Equal(t, []string{"PROJ:not_connected", "DEMO:not_connected"}, links)
}

func TestU4_Events_LowerEpochIsIgnored(t *testing.T) {
	r := eventRig(t)
	r.change(connection.ConnectionChanged{Reason: connection.ReasonProjectsChanged, ConnectionEpoch: 5, SpaceHost: host, SelectedProjects: []string{"PROJ", "DEMO"}})
	r.change(connection.ConnectionChanged{Reason: connection.ReasonDisconnected, ConnectionEpoch: 4})
	links, _ := states(r, t)
	require.Equal(t, []string{"PROJ:active", "DEMO:active"}, links)
}

func TestU4_Events_ReconcileAppliesTheSameRules(t *testing.T) {
	r := eventRig(t)
	r.conn.update(func(c *fakeConn) { c.snap.SelectedProjects = []string{"PROJ"} })
	require.NoError(t, r.svc.ReconcileAll(context.Background()))
	links, watches := states(r, t)
	require.Equal(t, []string{"PROJ:active", "DEMO:not_connected"}, links)
	require.Equal(t, []string{"PROJ:active", "DEMO:not_connected"}, watches)

	r.conn.update(func(c *fakeConn) { c.err = connection.ErrNotConnected })
	require.NoError(t, r.svc.ReconcileAll(context.Background()))
	links, _ = states(r, t)
	require.Equal(t, []string{"PROJ:not_connected", "DEMO:not_connected"}, links)

	r.conn.update(func(c *fakeConn) { c.err, c.snap.SelectedProjects = nil, []string{"PROJ", "DEMO"} })
	require.NoError(t, r.svc.ReconcileAll(context.Background()))
	links, watches = states(r, t)
	require.Equal(t, []string{"PROJ:active", "DEMO:active"}, links)
	require.Equal(t, []string{"PROJ:paused", "DEMO:paused"}, watches)
}

func TestU4_Events_ListenSubscribesOnce(t *testing.T) {
	r := newRig(t)
	stop := r.svc.Listen()
	defer stop()
	r.conn.mu.Lock()
	n := len(r.conn.subs)
	r.conn.mu.Unlock()
	require.Equal(t, 1, n)
}

func TestU4_Events_EachCycleReconcilesFirst(t *testing.T) {
	r := eventRig(t)
	r.conn.update(func(c *fakeConn) { c.snap.SelectedProjects = []string{"PROJ"} })
	NewWatcher(r.svc, nil).cycleAll(context.Background())
	links, watches := states(r, t)
	require.Equal(t, []string{"PROJ:active", "DEMO:not_connected"}, links, "a missed projects_changed is caught up")
	require.Equal(t, []string{"PROJ:active", "DEMO:not_connected"}, watches)
}
