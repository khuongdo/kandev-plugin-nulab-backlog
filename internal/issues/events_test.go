package issues

import (
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/connection"
)

func eventRig(t *testing.T) *rig {
	r := newRig(t)
	r.link(t, Link{IssueKey: "PROJ-118", TaskID: "task-17"})
	r.link(t, Link{IssueKey: "DEMO-1", TaskID: "task-18"})
	return r
}

func states(r *rig, t *testing.T) map[string]string {
	out := map[string]string{}
	for _, l := range r.links(t) {
		out[l.TaskID] = l.State
	}
	return out
}

func change(reason connection.Reason, epoch int, host string, selected ...string) connection.ConnectionChanged {
	return connection.ConnectionChanged{WorkspaceID: "ws-1", Reason: reason, ConnectionEpoch: epoch, SpaceHost: host, SelectedProjects: selected}
}

func TestU3_Events_DisconnectTurnsLinksOff(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		r := eventRig(t)
		r.svc.OnConnectionChanged(r.ctx, change(connection.ReasonDisconnected, 2, ""))
		r.conn.set(func(c *fakeConn) { c.currentErr = connection.ErrNotConnected })
		require.Equal(t, map[string]string{"task-17": StateNotConnected, "task-18": StateNotConnected}, states(r, t))
		y := NewSyncer(r.svc, r.log)
		y.Start()
		defer y.Stop()
		after(time.Minute)
		require.Zero(t, r.gw.count("issue"), "AC1.5.4: not-connected links are not polled")
	})
}

func TestU3_Events_SpaceChangeTurnsOldHostOff(t *testing.T) {
	r := eventRig(t)
	r.svc.OnConnectionChanged(r.ctx, change(connection.ReasonSpaceChanged, 2, "other.backlog.jp", "PROJ", "DEMO"))
	require.Equal(t, map[string]string{"task-17": StateNotConnected, "task-18": StateNotConnected}, states(r, t), "AC1.8.2")
}

func TestU3_Events_DeselectOnlyThoseProjects(t *testing.T) {
	r := eventRig(t)
	r.svc.OnConnectionChanged(r.ctx, change(connection.ReasonProjectsChanged, 2, spaceHost, "PROJ"))
	require.Equal(t, map[string]string{"task-17": StateActive, "task-18": StateNotConnected}, states(r, t), "AC1.9.2")
}

func TestU3_Events_RestoreBringsLinksBack(t *testing.T) {
	r := eventRig(t)
	r.svc.OnConnectionChanged(r.ctx, change(connection.ReasonDisconnected, 2, ""))
	e := change(connection.ReasonConnected, 3, spaceHost, "PROJ")
	e.Restore = true
	r.svc.OnConnectionChanged(r.ctx, e)
	require.Equal(t, map[string]string{"task-17": StateActive, "task-18": StateNotConnected}, states(r, t), "M12")
}

func TestU3_Events_OlderEpochIsIgnored(t *testing.T) {
	r := eventRig(t)
	r.svc.OnConnectionChanged(r.ctx, change(connection.ReasonProjectsChanged, 5, spaceHost, "PROJ", "DEMO"))
	r.svc.OnConnectionChanged(r.ctx, change(connection.ReasonDisconnected, 4, ""))
	require.Equal(t, map[string]string{"task-17": StateActive, "task-18": StateActive}, states(r, t))
}

func TestU3_Events_ReconcileWithCurrent(t *testing.T) {
	r := eventRig(t)
	r.conn.set(func(c *fakeConn) { c.snap.SelectedProjects = []string{"DEMO"} })
	require.NoError(t, r.svc.ReconcileAll(r.ctx))
	require.Equal(t, map[string]string{"task-17": StateNotConnected, "task-18": StateActive}, states(r, t))
	r.conn.set(func(c *fakeConn) { c.snap.SelectedProjects = []string{"PROJ", "DEMO"} })
	require.NoError(t, r.svc.ReconcileAll(r.ctx))
	require.Equal(t, map[string]string{"task-17": StateActive, "task-18": StateActive}, states(r, t), "a reselected project comes back")
	r.conn.set(func(c *fakeConn) { c.currentErr = connection.ErrNotConnected })
	require.NoError(t, r.svc.ReconcileAll(r.ctx))
	require.Equal(t, map[string]string{"task-17": StateNotConnected, "task-18": StateNotConnected}, states(r, t))
	r.state.failGet = true
	require.Error(t, r.svc.ReconcileAll(r.ctx))
}

func TestU3_Events_TaskDeletedRemovesLinks(t *testing.T) {
	r := eventRig(t)
	r.link(t, Link{IssueKey: "PROJ-120", TaskID: "task-17"})
	require.NoError(t, r.svc.OnTaskDeleted(r.ctx, "ws-1", "task-17"))
	require.Equal(t, map[string]string{"task-18": StateActive}, states(r, t), "R-05")
	require.NoError(t, r.svc.OnTaskDeleted(r.ctx, "ws-1", "task-17"), "idempotent")
	r.state.failGet = true
	require.ErrorIs(t, r.svc.OnTaskDeleted(r.ctx, "ws-1", "task-18"), connection.ErrStore, "a store failure is returned so Kandev retries")
}

func TestU3_Events_ListenSubscribes(t *testing.T) {
	r := eventRig(t)
	stop := r.svc.Listen()
	defer stop()
	require.Len(t, r.conn.subs, 1)
	r.conn.subs[0](change(connection.ReasonDisconnected, 2, ""))
	require.Equal(t, StateNotConnected, states(r, t)["task-17"])
}
