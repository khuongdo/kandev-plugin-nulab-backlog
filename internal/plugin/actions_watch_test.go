package plugin

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"testing"
	"testing/synctest"
	"time"

	"github.com/kandev/kandev/pkg/pluginsdk"
	"github.com/stretchr/testify/require"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/backlog"
	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/issues"
	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/testutil"
)

// PullRequestCount counts the scripted PROJ/web-app pull requests.
func (g *u4Gateway) PullRequestCount(_ context.Context, _ backlog.Credentials, _ backlog.CallClass, _, _ string, q backlog.PullRequestQuery) (int, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.prsErr != nil {
		return 0, g.prsErr
	}
	n := 0
	for _, pr := range g.prs {
		if len(q.StatusIDs) == 0 || slices.Contains(q.StatusIDs, int64(pr.StatusID)) {
			n++
		}
	}
	return n, nil
}

func issueWatchBody(mut func(map[string]any)) map[string]any {
	b := map[string]any{"name": "Open bugs", "projectKey": "PROJ", "statusIds": []int{2}, "assignee": "anyone",
		"creator": "me", "workflowId": "wf-1", "workflowStepId": "step-1"}
	if mut != nil {
		mut(b)
	}
	return b
}

// FR3.5, FR1.4: the issue watch actions for any member, mapped like PR watches.
func TestIssueWatchActions_CRUDAndErrors(t *testing.T) {
	r := newU3Rig(t)
	resp, out := r.call(t, actionIssueWatchesSave, issueWatchBody(nil))
	require.Equal(t, 200, resp.Status)
	id := out["id"].(string)
	require.Equal(t, "active", out["state"])
	require.EqualValues(t, 5, out["intervalMinutes"], "BR3.2")
	require.EqualValues(t, 1234, out["createdUserId"], "me is the connected user")

	resp, out = r.call(t, actionIssueWatchesList, nil)
	require.Equal(t, 200, resp.Status)
	require.Len(t, out["watches"], 1)
	for _, step := range [][2]string{{actionIssueWatchesPause, "paused"}, {actionIssueWatchesResume, "active"}} {
		resp, out = r.call(t, step[0], map[string]string{"id": id})
		require.Equal(t, 200, resp.Status, step[0])
		require.Equal(t, step[1], out["state"], step[0])
	}
	resp, out = r.call(t, actionIssueWatchesRun, map[string]string{"id": id})
	require.Equal(t, 200, resp.Status)
	require.Equal(t, true, out["queued"])
	r.call(t, actionIssueWatchesPause, map[string]string{"id": id})
	resp, out = r.call(t, actionIssueWatchesRun, map[string]string{"id": id})
	require.Equal(t, 409, resp.Status, "BR3.9: run now needs an active watch")
	require.Equal(t, "conflict", errorOf(t, out)["code"])

	for field, body := range map[string]map[string]any{
		"name":            issueWatchBody(func(b map[string]any) { b["name"] = "" }),
		"projectKey":      issueWatchBody(func(b map[string]any) { b["projectKey"] = "OTHER" }),
		"statusIds":       issueWatchBody(func(b map[string]any) { b["statusIds"] = []int{} }),
		"intervalMinutes": issueWatchBody(func(b map[string]any) { b["intervalMinutes"] = 1441 }),
		"workflowId":      issueWatchBody(func(b map[string]any) { b["workflowId"] = "" }),
	} {
		resp, out = r.call(t, actionIssueWatchesSave, body)
		require.Equal(t, 400, resp.Status, field)
		require.Equal(t, field, errorOf(t, out)["field"], field)
	}
	resp, out = r.call(t, actionIssueWatchesSave, []byte(`{"intervalMinutes":"five"}`))
	require.Equal(t, 400, resp.Status)
	require.Equal(t, "validation", errorOf(t, out)["code"])

	resp, _ = r.call(t, actionIssueWatchesDelete, map[string]string{"id": id})
	require.Equal(t, 200, resp.Status)
	resp, out = r.call(t, actionIssueWatchesDelete, map[string]string{"id": id})
	require.Equal(t, 404, resp.Status)
	require.Equal(t, "not_found", errorOf(t, out)["code"])
}

// FR4.1, FR4.3: git.prs.list pages a repository's pull requests.
func TestPRListAction_PagesAndValidates(t *testing.T) {
	r := newU4Rig(t)
	for n := 25; n >= 1; n-- {
		r.gw4.prs = append(r.gw4.prs, backlog.PullRequest{RepositoryID: 11, Number: n, Summary: fmt.Sprintf("Change %d", n), StatusID: 1,
			AuthorName: "Test User", Updated: "2026-10-02T09:00:00Z"})
	}
	resp, out := r.call(t, actionPRList, map[string]any{"projectKey": "PROJ", "repoName": "web-app", "statuses": []string{"open"}, "page": 1})
	require.Equal(t, 200, resp.Status)
	require.Len(t, out["items"], 20)
	require.EqualValues(t, 25, out["total"])
	require.Equal(t, true, out["hasNext"])
	first := out["items"].([]any)[0].(map[string]any)
	require.Equal(t, "Test User", first["author"])
	require.Equal(t, "open", first["status"])

	resp, out = r.call(t, actionPRList, map[string]any{"projectKey": "PROJ", "repoName": "web-app", "page": 0})
	require.Equal(t, 400, resp.Status)
	require.Equal(t, "page", errorOf(t, out)["field"])
	resp, out = r.call(t, actionPRList, map[string]any{"projectKey": "DEMO", "repoName": "x", "page": 1})
	require.Equal(t, 400, resp.Status)
	require.Equal(t, "repository", errorOf(t, out)["field"])

	resp, out = r.call(t, actionQueriesSave, map[string]any{"name": "Mine", "projectKey": "PROJ", "repoName": "web-app",
		"statuses": []string{"open"}, "assignee": "anyone", "creator": "me"})
	require.Equal(t, 200, resp.Status)
	require.Equal(t, "me", out["creator"], "FR2.6: a saved query keeps the creator filter")
}

// BR3.14: Kandev's refusal of a removed workflow reaches the watcher as ErrWorkflowMissing.
func TestIssueHost_MapsAMissingWorkflow(t *testing.T) {
	r := newU3Rig(t)
	p := issueHost{hostPort{host: r.rt.Host}}
	for msg, missing := range map[string]bool{
		"rpc error: code = NotFound desc = workflow not found":            true,
		"rpc error: code = InvalidArgument desc = step not in workflow":   true,
		"rpc error: code = PermissionDenied desc = plugin may not create": false,
	} {
		r.data.mu.Lock()
		r.data.createErr = errors.New(msg)
		r.data.mu.Unlock()
		_, err := p.CreateTask(context.Background(), issues.NewTask{WorkspaceID: "ws-1", WorkflowID: "wf-gone"})
		require.Error(t, err)
		require.Equal(t, missing, errors.Is(err, issues.ErrWorkflowMissing), msg)
	}
}

// NFR2, NFR1 (github-parity-actions): no API key in the new actions' replies
// and logs, nor in the watcher's lines.
func TestIssueWatchActions_NoSecretLeaks(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		r := newU3Rig(t)
		key := testutil.APIKey(t)
		r.call(t, actionConnectAPIKey, connectBody(key))
		r.call(t, keySetProjects, map[string]any{"projectKeys": []string{"PROJ"}})
		var bodies string
		resp, out := r.call(t, actionIssueWatchesSave, issueWatchBody(nil))
		require.Equal(t, 200, resp.Status)
		bodies += string(resp.Body)
		r.gw3.mu.Lock()
		r.gw3.listErr = &backlog.Error{Kind: backlog.KindUnauthorized, Status: 401}
		r.gw3.mu.Unlock()
		r.rt.issueWatcher.Tick = time.Minute
		r.rt.Start()
		defer r.rt.Close()
		time.Sleep(time.Minute)
		synctest.Wait()
		for _, k := range []string{actionIssueWatchesList, actionIssueWatchesRun} {
			resp, _ = r.call(t, k, map[string]string{"id": out["id"].(string)})
			bodies += string(resp.Body)
		}
		resp, _ = r.call(t, actionPRList, map[string]any{"projectKey": "PROJ", "repoName": "web-app", "page": 1})
		bodies += string(resp.Body)
		for k, b := range map[string]any{
			actionQuickActionsGet:     nil,
			actionQuickActionsSave:    map[string]any{"kind": "pr", "actions": []map[string]any{{"label": "", "icon": "eye"}}},
			actionIssueQueriesSave:    map[string]any{"name": "Mine", "assignee": "me"},
			actionIssueQueriesList:    nil,
			actionIssueQueriesDefault: map[string]any{"id": "missing", "isDefault": true},
			actionIssuesList:          map[string]any{"assignee": "me"},
		} {
			resp, _ = r.call(t, k, b)
			bodies += string(resp.Body)
		}
		require.Contains(t, r.logs.String(), `"event":"issue_watch_failed"`)
		require.Contains(t, bodies, `"lastError":"unauthorized"`)
		testutil.AssertNoLeak(t, bodies+r.logs.String(), key)
	})
}

// NFR6, FR3.2: the issue watcher starts and stops with the plugin; before
// Kandev injects the Host its store calls wait (bounded) for it.
func TestIssueWatcher_StartsWithThePluginAndWaitsForTheHost(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		r := newU3Rig(t)
		resp, _ := r.call(t, actionIssueWatchesSave, issueWatchBody(func(b map[string]any) { b["creator"] = "anyone" }))
		require.Equal(t, 200, resp.Status)

		rt := newRuntime(r.gw3, r.logs, "debug") // a restart: the Host comes later
		rt.issueWatcher.Tick = time.Minute
		rt.Start()
		defer rt.Close()
		time.Sleep(time.Minute) // the first tick: its store call now waits for the Host
		synctest.Wait()
		rt.SetHost(r.rt.Host())
		time.Sleep(2 * time.Second)
		synctest.Wait()
		r.data.mu.Lock()
		creates := slices.Clone(r.data.creates)
		r.data.mu.Unlock()
		require.Len(t, creates, 1, "the waiting tick ran once the Host arrived")
		step := "step-1"
		require.Equal(t, pluginsdk.CreateTaskInput{WorkspaceID: "ws-1", WorkflowID: "wf-1", WorkflowStepID: &step,
			Title: "Fix login timeout", Description: "Steps.\n\nBacklog: https://" + spaceHost + "/view/PROJ-118", Priority: "high"}, creates[0])
		rt.Close()
		time.Sleep(10 * time.Minute)
		synctest.Wait()
		r.data.mu.Lock()
		defer r.data.mu.Unlock()
		require.Len(t, r.data.creates, 1, "a closed plugin runs no watch")
	})
}
