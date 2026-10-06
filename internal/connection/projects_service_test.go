package connection

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/backlog"
)

var spaceProjects = []backlog.Project{{ID: 101, Key: "PROJ", Name: "Test Project"}, {ID: 102, Key: "DEMO", Name: "Demo Project"}}

func TestProjectsListMarksTheSelection(t *testing.T) {
	u := newU2(t)
	u.connectKey(t, "a.backlog.com")
	_, _ = u.svc.store.SaveProjects(u.ctx, ws, []string{"PROJ"})
	u.gw.projects = spaceProjects
	items, err := u.svc.ListProjects(u.ctx, ws)
	require.NoError(t, err)
	require.Equal(t, []ProjectItem{
		{ProjectKey: "PROJ", ProjectID: 101, ProjectName: "Test Project", Selected: true},
		{ProjectKey: "DEMO", ProjectID: 102, ProjectName: "Demo Project"},
	}, items)
}

func TestProjectsListEmptyIsValid(t *testing.T) {
	u := newU2(t)
	u.connectKey(t, "a.backlog.com")
	items, err := u.svc.ListProjects(u.ctx, ws)
	require.NoError(t, err)
	require.NotNil(t, items)
	require.Empty(t, items, "AC1.7.3")
}

func TestProjectsListNeedsAConnection(t *testing.T) {
	u := newU2(t)
	_, err := u.svc.ListProjects(u.ctx, ws)
	require.Equal(t, CodeReconnectRequired, Classify(err).Code)
	u.connectKey(t, "a.backlog.com")
	u.gw.projectsErr = &backlog.Error{Kind: backlog.KindForbidden, Status: 403}
	_, err = u.svc.ListProjects(u.ctx, ws)
	require.Equal(t, CodeReconnectRequired, Classify(err).Code, "R-06: 403 on a U2 call")
}

func TestProjectsSetStoresTheSelectionAndSendsAnEvent(t *testing.T) {
	u := newU2(t)
	u.connectKey(t, "a.backlog.com")
	u.gw.projects = spaceProjects
	events := subscribe(t, u.svc)

	view, err := u.svc.SetProjects(u.ctx, ws, []any{"proj"})
	require.NoError(t, err)
	require.Equal(t, []string{"PROJ"}, view.SelectedProjects)
	_, projects, _, _ := u.gw.counts()
	require.Equal(t, 1, projects, "Projects is called once")
	e := events.next(t)
	require.Equal(t, ReasonProjectsChanged, e.Reason)
	require.Equal(t, []string{"PROJ"}, e.SelectedProjects)
	require.True(t, e.Restore, "a key was added")

	_, err = u.svc.SetProjects(u.ctx, ws, []any{"PROJ"})
	require.NoError(t, err)
	require.Len(t, u.changedLogs(t), 2, "an unchanged selection sends no event")

	_, err = u.svc.SetProjects(u.ctx, ws, []any{})
	require.NoError(t, err)
	e = events.next(t)
	require.Equal(t, ReasonProjectsChanged, e.Reason)
	require.False(t, e.Restore, "only removed (AC1.9.2)")
	require.Empty(t, e.SelectedProjects)
}

func TestProjectsSetRejectsKeysNotOnTheSpace(t *testing.T) {
	u := newU2(t)
	u.connectKey(t, "a.backlog.com")
	u.gw.projects = spaceProjects
	before := u.state.snapshot()
	_, err := u.svc.SetProjects(u.ctx, ws, []any{"PROJ", "OTHER"})
	require.Equal(t, Outcome{Code: CodeValidation, Field: FieldProjectKeys}, Classify(err))
	require.Equal(t, before, u.state.snapshot())
}

func TestProjectsSetBadBodyMakesNoCall(t *testing.T) {
	u := newU2(t)
	u.connectKey(t, "a.backlog.com")
	_, err := u.svc.SetProjects(u.ctx, ws, "PROJ")
	require.ErrorIs(t, err, ErrInvalidProjects)
	_, projects, _, _ := u.gw.counts()
	require.Zero(t, projects)
}
