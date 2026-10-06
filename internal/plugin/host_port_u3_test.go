package plugin

import (
	"context"
	"testing"

	"github.com/kandev/kandev/pkg/pluginsdk"
	"github.com/stretchr/testify/require"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/issues"
)

func TestU3_HostPort_CreatesWithIdentifier(t *testing.T) {
	r := newU3Rig(t)
	p := issueHost{hostPort{host: r.rt.Host}}
	ref, err := p.CreateTask(context.Background(), issues.NewTask{WorkspaceID: "ws-1", WorkflowID: "wf-1", Title: "T", Description: "D", Priority: "low"})
	require.NoError(t, err)
	require.Equal(t, issues.TaskRef{ID: "task-2", Key: "T-2"}, ref)
	require.Equal(t, pluginsdk.CreateTaskInput{WorkspaceID: "ws-1", WorkflowID: "wf-1", Title: "T", Description: "D", Priority: "low"},
		r.data.creates[0], "no step, no metadata, no repository")
	r.data.failCreate = true
	_, err = p.CreateTask(context.Background(), issues.NewTask{WorkspaceID: "ws-1"})
	require.Error(t, err)
}

func TestU3_HostPort_ListsEveryPage(t *testing.T) {
	r := newU3Rig(t)
	r.data.tasks = append(r.data.tasks,
		pluginsdk.Task{ID: "task-2", Identifier: "T-2", Title: "B"},
		pluginsdk.Task{ID: "task-3", Identifier: "T-3", Title: "C"})
	got, err := issueHost{hostPort{host: r.rt.Host}}.ListTasks(context.Background(), "ws-1")
	require.NoError(t, err)
	require.Equal(t, []issues.TaskInfo{{ID: "task-17", Key: "T-17", Title: "Login work"}, {ID: "task-2", Key: "T-2", Title: "B"},
		{ID: "task-3", Key: "T-3", Title: "C"}}, got)
	require.Len(t, r.data.listPages, 2, "two tasks per page, archived included (the fake refuses otherwise)")
}

func TestU3_HostPort_NoHostIsAnError(t *testing.T) {
	p := issueHost{hostPort{host: func() pluginsdk.Host { return nil }}}
	_, err := p.CreateTask(context.Background(), issues.NewTask{})
	require.ErrorIs(t, err, errNoHost)
	_, err = p.ListTasks(context.Background(), "ws-1")
	require.ErrorIs(t, err, errNoHost)
}

// filterSpy records the TaskFilter of every Tasks().List call.
type filterSpy struct {
	pluginsdk.Host
	filters *[]pluginsdk.TaskFilter
}

func (s filterSpy) Tasks() pluginsdk.TaskReader { return spyTasks{s.Host.Tasks(), s.filters} }

type spyTasks struct {
	pluginsdk.TaskReader
	filters *[]pluginsdk.TaskFilter
}

func (s spyTasks) List(ctx context.Context, f pluginsdk.TaskFilter, p pluginsdk.Page) ([]pluginsdk.Task, *pluginsdk.PageInfo, error) {
	*s.filters = append(*s.filters, f)
	return s.TaskReader.List(ctx, f, p)
}

// R-02 (review 1): ephemeral tasks are listed, so their links get task keys.
func TestU3_HostPort_ListsEphemeralTasks(t *testing.T) {
	r := newU3Rig(t)
	var filters []pluginsdk.TaskFilter
	p := issueHost{hostPort{host: func() pluginsdk.Host { return filterSpy{r.rt.Host(), &filters} }}}
	_, err := p.ListTasks(context.Background(), "ws-1")
	require.NoError(t, err)
	require.Equal(t, []pluginsdk.TaskFilter{{WorkspaceIDs: []string{"ws-1"}, IncludeArchived: true, IncludeEphemeral: true}}, filters)
}
