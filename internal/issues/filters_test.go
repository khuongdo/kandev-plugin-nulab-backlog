package issues

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/backlog"
	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/connection"
)

func filterRig(t *testing.T) *rig {
	r := newRig(t)
	r.gw.statuses["PROJ"] = []backlog.Status{{ID: 1, Name: "Open"}, {ID: 2, Name: "In Progress"}, {ID: 7, Name: "Review"}}
	r.gw.statuses["DEMO"] = []backlog.Status{{ID: 1, Name: "Open"}, {ID: 2, Name: "In Progress"}}
	r.gw.users["PROJ"] = []backlog.ProjectUser{{ID: 1, Name: "Test User"}, {ID: 2, Name: "Lan"}}
	r.gw.users["DEMO"] = []backlog.ProjectUser{{ID: 2, Name: "Lan"}}
	return r
}

func TestU3_Filters_ListsProjectsStatusesUsers(t *testing.T) {
	r := filterRig(t)
	f, err := r.svc.Filters(r.ctx, "ws-1")
	require.NoError(t, err)
	require.Equal(t, []Option{{Key: "PROJ", Name: "Test Project"}, {Key: "DEMO", Name: "Demo Project"}}, f.Projects)
	require.Equal(t, []Option{{ID: 1, Name: "Open"}, {ID: 2, Name: "In Progress"}, {ID: 7, Name: "Review"}}, f.Statuses, "deduplicated by id")
	require.Equal(t, []Option{{ID: 1, Name: "Test User"}, {ID: 2, Name: "Lan"}}, f.Assignees, "the union of the members")
}

func TestU3_Filters_OneCallPerProject(t *testing.T) {
	r := filterRig(t)
	_, err := r.svc.Filters(r.ctx, "ws-1")
	require.NoError(t, err)
	require.Equal(t, 2, r.gw.count("statuses"))
	require.Equal(t, 2, r.gw.count("users"))
	require.Equal(t, 1, r.gw.count("projects"))
}

func TestU3_Filters_KeepsErrorCodes(t *testing.T) {
	r := filterRig(t)
	r.gw.errs["users"] = forbidden()
	_, err := r.svc.Filters(r.ctx, "ws-1")
	require.Equal(t, "reconnect_required", connection.Classify(err).Code)
}

func TestU3_Filters_NeedAProject(t *testing.T) {
	r := filterRig(t)
	r.conn.set(func(c *fakeConn) { c.snap.SelectedProjects = nil })
	_, err := r.svc.Filters(r.ctx, "ws-1")
	require.ErrorIs(t, err, ErrNoProject)
	require.Zero(t, r.gw.total())
}
