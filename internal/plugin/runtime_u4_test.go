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

func TestU4_HostPort_ReadsNestedMetadata(t *testing.T) {
	r := newU4Rig(t)
	p := hostPort{host: r.rt.Host}
	ctx := context.Background()
	_, err := p.CreateTask(ctx, git.NewTask{WorkspaceID: "ws-1", Title: "T", Metadata: map[string]any{git.MetadataKey: "k1"}})
	require.NoError(t, err)
	r.data.mu.Lock()
	r.data.tasks = append(r.data.tasks,
		pluginsdk.Task{ID: "flat", Metadata: map[string]any{git.MetadataKey: "k2"}},
		pluginsdk.Task{ID: "other", Metadata: map[string]any{"plugin:other": map[string]any{git.MetadataKey: "k3"}}})
	r.data.mu.Unlock()
	id, ok, err := p.FindTaskByMetadata(ctx, "ws-1", git.MetadataKey, "k1")
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, "task-1", id)
	for _, v := range []string{"k2", "k3"} {
		_, ok, err := p.FindTaskByMetadata(ctx, "ws-1", git.MetadataKey, v)
		require.NoError(t, err)
		require.False(t, ok, "only this plugin's namespace counts: %s", v)
	}
}

func TestU4_HostPort_RepoNameFromProvider(t *testing.T) {
	cases := map[string]struct {
		providerName, remoteURL, want string
	}{
		"provider name wins":     {providerName: "web-app", remoteURL: "https://" + spaceHost + "/git/PROJ/other.git", want: "web-app"},
		"clone URL fallback":     {remoteURL: "https://" + spaceHost + "/git/PROJ/api.git", want: "api"},
		"clone URL without .git": {remoteURL: "https://" + spaceHost + "/git/PROJ/api", want: "api"},
		"nothing usable":         {},
		"invalid provider name":  {providerName: "../web-app"},
		"invalid clone URL name": {remoteURL: "https://" + spaceHost + "/git/PROJ/.git"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r := newU4Rig(t)
			repo := backlogRepo()
			repo.ProviderName, repo.RemoteURL = tc.providerName, tc.remoteURL
			r.data.repos = []pluginsdk.Repository{repo}
			got, err := hostPort{host: r.rt.Host}.Repository(context.Background(), "ws-1", "repo-k1")
			if tc.want == "" {
				require.ErrorIs(t, err, git.ErrRepositoryNotFound, "fail closed, never Name %q", repo.Name)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.want, got.Name)
		})
	}
}

func TestU4_HostPort_OtherProviderKeepsWorking(t *testing.T) {
	r := newU4Rig(t)
	got, err := hostPort{host: r.rt.Host}.Repository(context.Background(), "ws-1", "repo-gh")
	require.NoError(t, err)
	require.Equal(t, "github", got.ProviderID)
}

func TestU4_Actions_ShortRefsUseProviderName(t *testing.T) {
	for _, ref := range []string{"42", "#42", "web-app#42"} {
		t.Run(ref, func(t *testing.T) {
			r := newU4Rig(t)
			r.gw4.prs = []backlog.PullRequest{{RepositoryID: 11, Number: 42, Summary: "Add login page", StatusID: 1}}
			resp, _ := r.taskCall(t, actionPRLink, map[string]string{"reference": ref})
			require.Equal(t, 200, resp.Status)
			require.Equal(t, []string{"PROJ/web-app"}, r.gw4.pathLog())
		})
	}
}

func TestU4_Watcher_CrashAfterCreateNoDuplicate(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		r := newU4Rig(t)
		r.gw4.prs = []backlog.PullRequest{{RepositoryID: 11, Number: 1, Summary: "First", StatusID: 1}}
		resp, _ := r.call(t, actionWatchesSave, map[string]any{"name": "Reviews", "projectKey": "PROJ", "repoName": "web-app",
			"statuses": []string{"open"}, "assignee": "anyone", "creator": "anyone", "workflowId": "wf-1"})
		require.Equal(t, 200, resp.Status)
		r.data.mu.Lock()
		r.data.crashAfterCreate = true
		r.data.mu.Unlock()
		r.rt.watcher.Every = time.Minute
		r.rt.Start()
		time.Sleep(time.Minute)
		synctest.Wait()
		r.rt.Close()
		r.data.mu.Lock()
		require.Len(t, r.data.tasks, 1, "the task was created, its id was not stored")
		r.data.crashAfterCreate, r.data.crashed = false, false
		r.data.mu.Unlock()

		// Restart: a fresh runtime over the same Kandev host and plugin storage.
		rt := newRuntime(r.gw4, r.logs, "debug")
		rt.SetHost(r.rt.Host())
		rt.watcher.Every = time.Minute
		rt.Start()
		defer rt.Close()
		time.Sleep(time.Minute)
		synctest.Wait()
		r.data.mu.Lock()
		defer r.data.mu.Unlock()
		require.Len(t, r.data.tasks, 1, "the reserved entry is found by its nested metadata")
		require.Len(t, r.data.creates, 1)
	})
}

func TestU4_Runtime_WatcherObeysTheSwitch(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		r := newU4Rig(t)
		r.gw4.prs = []backlog.PullRequest{{RepositoryID: 11, Number: 1, Summary: "First", StatusID: 1}}
		resp, _ := r.call(t, actionWatchesSave, map[string]any{"name": "Reviews", "projectKey": "PROJ", "repoName": "web-app",
			"statuses": []string{"open"}, "assignee": "anyone", "creator": "anyone", "workflowId": "wf-1"})
		require.Equal(t, 200, resp.Status)
		r.call(t, actionSetEnabled, map[string]any{"enabled": false})
		r.rt.watcher.Every = time.Minute
		r.rt.Start()
		defer r.rt.Close()
		paths := len(r.gw4.pathLog())
		time.Sleep(time.Minute)
		synctest.Wait()
		require.Len(t, r.gw4.pathLog(), paths, "no Backlog call while off")
		r.data.mu.Lock()
		require.Empty(t, r.data.creates)
		r.data.mu.Unlock()
		r.call(t, actionSetEnabled, map[string]any{"enabled": true})
		time.Sleep(time.Minute)
		synctest.Wait()
		r.data.mu.Lock()
		defer r.data.mu.Unlock()
		require.Len(t, r.data.creates, 1)
	})
}
