package plugin

import (
	"context"
	"testing"
	"testing/synctest"
	"time"

	"github.com/kandev/kandev/pkg/pluginsdk"
	"github.com/stretchr/testify/require"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/backlog"
	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/git"
)

func TestU4_Runtime_StartWiresEventsAndTheWatcher(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		r := newU4Rig(t)
		r.rt.watcher.Every = time.Minute
		r.rt.Start()
		defer r.rt.Close()
		r.gw4.prs = []backlog.PullRequest{{RepositoryID: 11, Number: 1, Summary: "First", StatusID: 1}}
		resp, _ := r.call(t, actionWatchesSave, map[string]any{"name": "Reviews", "projectKey": "PROJ", "repoName": "web-app",
			"statuses": []string{"open"}, "assignee": "anyone", "creator": "anyone", "workflowId": "wf-1", "workflowStepId": "step-2"})
		require.Equal(t, 200, resp.Status)
		time.Sleep(time.Minute)
		synctest.Wait()
		r.data.mu.Lock()
		creates := r.data.creates
		r.data.mu.Unlock()
		require.Len(t, creates, 1, "the watcher runs with the injected interval")
		step := "step-2"
		require.Equal(t, pluginsdk.CreateTaskInput{WorkspaceID: "ws-1", WorkflowID: "wf-1", WorkflowStepID: &step,
			Title: "Review PR #1: First", Description: "https://" + spaceHost + "/git/PROJ/web-app/pullRequests/1",
			Metadata: map[string]any{git.MetadataKey: spaceHost + "|11|1"}}, creates[0])

		r.call(t, keyDisconnect, nil)
		synctest.Wait()
		_, out := r.call(t, actionWatchesList, nil)
		require.Equal(t, "not_connected", out["watches"].([]any)[0].(map[string]any)["state"], "the Git service hears ConnectionChanged")
	})
}

func TestU4_Runtime_CloseIsSafeTwiceAndBeforeStart(t *testing.T) {
	r := newU4Rig(t)
	r.rt.Close()
	r.rt.Start()
	r.rt.Close()
	r.rt.Close()
}

func TestU4_HostPort_MapsTheHostDataAPI(t *testing.T) {
	r := newU4Rig(t)
	p := hostPort{host: r.rt.Host}
	ctx := context.Background()
	id, err := p.CreateTask(ctx, git.NewTask{WorkspaceID: "ws-1", WorkflowID: "wf-1", Title: "T", Metadata: map[string]any{git.MetadataKey: "k1"}})
	require.NoError(t, err)
	require.Equal(t, "task-1", id)
	require.Nil(t, r.data.creates[0].WorkflowStepID, "no step: the workflow's default")
	for _, k := range []string{"k2", "k3", "k4"} {
		_, err := p.CreateTask(ctx, git.NewTask{WorkspaceID: "ws-1", Title: k, Metadata: map[string]any{git.MetadataKey: k}})
		require.NoError(t, err)
	}
	found, ok, err := p.FindTaskByMetadata(ctx, "ws-1", git.MetadataKey, "k4")
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, "task-4", found)
	require.Len(t, r.data.listPages, 2, "it pages through the task list")
	_, ok, err = p.FindTaskByMetadata(ctx, "ws-1", git.MetadataKey, "nope")
	require.NoError(t, err)
	require.False(t, ok)

	repo, err := p.Repository(ctx, "ws-1", "repo-k1")
	require.NoError(t, err)
	require.Equal(t, git.KandevRepository{ID: "repo-k1", ProviderID: "nulab-backlog", ProviderRepositoryID: "11",
		ProviderHost: "https://" + spaceHost, ProviderScope: spaceHost, OwnerOrProject: "PROJ", Name: "web-app", DefaultBranch: "main"}, repo)
	_, err = p.Repository(ctx, "ws-1", "missing")
	require.ErrorIs(t, err, git.ErrRepositoryNotFound)

	r.data.failCreate = true
	_, err = p.CreateTask(ctx, git.NewTask{WorkspaceID: "ws-1"})
	require.Error(t, err)
}

func TestU4_HostPort_NoHostIsAnError(t *testing.T) {
	p := hostPort{host: func() pluginsdk.Host { return nil }}
	_, err := p.CreateTask(context.Background(), git.NewTask{})
	require.ErrorIs(t, err, errNoHost)
	_, _, err = p.FindTaskByMetadata(context.Background(), "ws", "k", "v")
	require.ErrorIs(t, err, errNoHost)
	_, err = p.Repository(context.Background(), "ws", "r")
	require.ErrorIs(t, err, errNoHost)
}
