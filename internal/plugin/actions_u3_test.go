package plugin

import (
	"context"
	"encoding/json"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/kandev/kandev/pkg/pluginsdk"
	"github.com/stretchr/testify/require"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/backlog"
	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/testutil"
)

// The base fakeGateway answers U3 calls with nothing; u3Gateway scripts them.
func (g *fakeGateway) Issues(context.Context, backlog.Credentials, backlog.CallClass, backlog.IssueQuery) ([]backlog.Issue, error) {
	return nil, nil
}

func (g *fakeGateway) IssueCount(context.Context, backlog.Credentials, backlog.CallClass, backlog.IssueQuery) (int, error) {
	return 0, nil
}

func (g *fakeGateway) IssueComments(context.Context, backlog.Credentials, backlog.CallClass, string, backlog.CommentQuery) ([]backlog.Comment, error) {
	return nil, nil
}

func (g *fakeGateway) IssueAttachments(context.Context, backlog.Credentials, backlog.CallClass, string) ([]backlog.Attachment, error) {
	return nil, nil
}

func (g *fakeGateway) ProjectStatuses(context.Context, backlog.Credentials, string) ([]backlog.Status, error) {
	return nil, nil
}

func (g *fakeGateway) ProjectUsers(context.Context, backlog.Credentials, string) ([]backlog.ProjectUser, error) {
	return nil, nil
}

type u3Gateway struct {
	*fakeGateway
	mu      sync.Mutex
	issues  []backlog.Issue
	listErr error
	getErr  error
	calls   map[string]int
	classes []backlog.CallClass // of Issue calls
}

func (g *u3Gateway) hit(op string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.calls[op]++
}

func (g *u3Gateway) count(op string) int {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.calls[op]
}

func (g *u3Gateway) Issues(context.Context, backlog.Credentials, backlog.CallClass, backlog.IssueQuery) ([]backlog.Issue, error) {
	g.hit("issues")
	g.mu.Lock()
	defer g.mu.Unlock()
	return slices.Clone(g.issues), g.listErr
}

func (g *u3Gateway) IssueCount(context.Context, backlog.Credentials, backlog.CallClass, backlog.IssueQuery) (int, error) {
	g.hit("count")
	g.mu.Lock()
	defer g.mu.Unlock()
	return len(g.issues), g.listErr
}

func (g *u3Gateway) Issue(_ context.Context, _ backlog.Credentials, class backlog.CallClass, ref string) (backlog.Issue, error) {
	g.hit("issue")
	g.mu.Lock()
	defer g.mu.Unlock()
	g.classes = append(g.classes, class)
	if g.getErr != nil {
		return backlog.Issue{}, g.getErr
	}
	for _, i := range g.issues {
		if i.IssueKey == ref {
			return i, nil
		}
	}
	return backlog.Issue{}, &backlog.Error{Kind: backlog.KindNotFound, Status: 404}
}

func (g *u3Gateway) IssueComments(context.Context, backlog.Credentials, backlog.CallClass, string, backlog.CommentQuery) ([]backlog.Comment, error) {
	g.hit("comments")
	return []backlog.Comment{{ID: 2, Content: "Fixed", AuthorName: "Lan"}}, nil
}

func (g *u3Gateway) setIssue(fn func(i *backlog.Issue)) {
	g.mu.Lock()
	defer g.mu.Unlock()
	fn(&g.issues[0])
}

type u3rig struct {
	*rig
	gw3  *u3Gateway
	data *hostData
}

// newU3Rig is connected to example-space.backlog.com with PROJ selected; the
// workspace has task-17 (T-17).
func newU3Rig(t *testing.T) *u3rig {
	t.Helper()
	gw := &u3Gateway{fakeGateway: &fakeGateway{projects: []backlog.Project{{ID: 101, Key: "PROJ", Name: "Test Project"}}},
		calls: map[string]int{}, issues: []backlog.Issue{
			{ID: 5118, ProjectID: 101, IssueKey: "PROJ-118", Summary: "Fix login timeout", Description: "Steps.",
				StatusID: 2, StatusName: "In Progress", PriorityID: 2, PriorityName: "High", AssigneeName: "Lan", Created: "2026-10-01T09:00:00Z"},
			{ID: 5120, ProjectID: 101, IssueKey: "PROJ-120", Summary: "Login page", StatusID: 3, StatusName: "Resolved", PriorityID: 3,
				Created: "2026-10-02T09:00:00Z"},
		}}
	logs, host := &syncBuffer{}, newFakeHost()
	switchOn(t, host, "ws-1")
	rt := newRuntime(gw, logs, "debug")
	data := &hostData{tasks: []pluginsdk.Task{{ID: "task-17", WorkspaceID: "ws-1", Title: "Login work", Identifier: "T-17"}}}
	rt.SetHost(&u4Host{fakeHost: host, data: data})
	r := &u3rig{rig: &rig{rt: rt, host: host, gw: gw.fakeGateway, logs: logs}, gw3: gw, data: data}
	r.connected(t)
	resp, _ := r.call(t, keySetProjects, map[string]any{"projectKeys": []string{"PROJ"}})
	require.Equal(t, 200, resp.Status)
	return r
}

// onTask invokes a task-scoped action with a verified task id.
func (r *u3rig) onTask(t *testing.T, key, taskID string, body any) (*pluginsdk.PluginActionResponse, map[string]any) {
	t.Helper()
	raw, err := json.Marshal(body)
	require.NoError(t, err)
	resp, err := r.rt.HandleAction(context.Background(), &pluginsdk.PluginActionRequest{
		ActionKey: key, Body: raw,
		Context: pluginsdk.VerifiedActionContext{WorkspaceID: "ws-1", ActorID: "user-1", TaskID: taskID},
	})
	require.NoError(t, err)
	var out map[string]any
	require.NoError(t, json.Unmarshal(resp.Body, &out))
	return resp, out
}

func TestU3_Actions_ListAndFilters(t *testing.T) {
	r := newU3Rig(t)
	resp, out := r.call(t, actionIssuesList, map[string]any{"page": 1, "pageSize": 20})
	require.Equal(t, 200, resp.Status)
	require.EqualValues(t, 2, out["total"])
	item := out["items"].([]any)[0].(map[string]any)
	require.Equal(t, "PROJ-118", item["issueKey"])
	require.Equal(t, "https://"+spaceHost+"/view/PROJ-118", item["url"])

	resp, out = r.call(t, actionIssuesFilters, nil)
	require.Equal(t, 200, resp.Status)
	require.Equal(t, []any{map[string]any{"key": "PROJ", "name": "Test Project"}}, out["projects"])

	resp, out = r.call(t, actionIssuesList, map[string]any{"projectKeys": []string{"OTHER"}})
	require.Equal(t, 400, resp.Status)
	require.Equal(t, "projectKeys", errorOf(t, out)["field"])
	resp, out = r.call(t, actionIssuesList, []byte(`{"page":"one"}`))
	require.Equal(t, 400, resp.Status)
	require.Equal(t, "validation", errorOf(t, out)["code"])
}

func TestU3_Actions_CreateTaskAndConflict(t *testing.T) {
	r := newU3Rig(t)
	resp, out := r.call(t, actionIssuesCreate, map[string]any{"issueKey": "PROJ-118", "workflowId": "wf-1", "workflowStepId": "step-1"})
	require.Equal(t, 200, resp.Status)
	require.Equal(t, map[string]any{"taskId": "task-2", "taskKey": "T-2", "issueKey": "PROJ-118"}, out)
	step := "step-1"
	require.Equal(t, pluginsdk.CreateTaskInput{WorkspaceID: "ws-1", WorkflowID: "wf-1", WorkflowStepID: &step, Title: "Fix login timeout",
		Description: "Steps.\n\nBacklog: https://" + spaceHost + "/view/PROJ-118", Priority: "high"}, r.data.creates[0])

	resp, out = r.call(t, actionIssuesCreate, map[string]any{"issueKey": "PROJ-118", "workflowId": "wf-1"})
	require.Equal(t, 409, resp.Status, "AC3.1.3")
	require.Equal(t, "conflict", errorOf(t, out)["code"])
	resp, _ = r.call(t, actionIssuesCreate, map[string]any{"issueKey": "PROJ-118", "workflowId": "wf-1", "force": true})
	require.Equal(t, 200, resp.Status)

	resp, out = r.call(t, actionIssuesLinks, nil)
	require.Equal(t, 200, resp.Status)
	require.Len(t, out["links"], 2)
	resp, out = r.call(t, actionIssuesImpact, map[string]any{"projectKeys": []string{"PROJ"}})
	require.Equal(t, 200, resp.Status)
	require.EqualValues(t, 2, out["issueLinks"])
}

func TestU3_Actions_TaskScopedLinkUnlink(t *testing.T) {
	r := newU3Rig(t)
	resp, out := r.onTask(t, actionIssuesLink, "task-17", map[string]any{"issueKey": "PROJ-120", "taskId": "evil"})
	require.Equal(t, 200, resp.Status)
	require.Equal(t, "task-17", out["taskId"], "the task comes from the verified context, never the body")
	require.Equal(t, "T-17", out["taskKey"])

	resp, out = r.onTask(t, actionIssuesLink, "task-17", map[string]any{"issueKey": "PROJ-118"})
	require.Equal(t, 409, resp.Status, "AC3.3.3")
	require.Equal(t, "conflict", errorOf(t, out)["code"])

	resp, out = r.call(t, actionTasksSearch, map[string]any{"query": "login"})
	require.Equal(t, 200, resp.Status)
	require.Equal(t, []any{map[string]any{"taskId": "task-17", "taskKey": "T-17", "title": "Login work", "linkedIssueKey": "PROJ-120"}}, out["tasks"])

	resp, out = r.onTask(t, actionIssuesGet, "task-17", nil)
	require.Equal(t, 200, resp.Status)
	require.Equal(t, "PROJ-120", out["issue"].(map[string]any)["key"])
	resp, out = r.onTask(t, actionIssuesComments, "task-17", map[string]any{"maxId": 0})
	require.Equal(t, 200, resp.Status)
	require.Len(t, out["comments"], 1)

	resp, _ = r.onTask(t, actionIssuesUnlink, "task-17", nil)
	require.Equal(t, 200, resp.Status)
	resp, out = r.onTask(t, actionIssuesUnlink, "task-17", nil)
	require.Equal(t, 404, resp.Status)
	require.Equal(t, "not_found", errorOf(t, out)["code"])
	resp, _ = r.onTask(t, actionIssuesGet, "task-17", nil)
	require.Equal(t, 404, resp.Status, "unlinked")
}

func TestU3_Actions_SettingsAndInterval(t *testing.T) {
	r := newU3Rig(t)
	resp, out := r.call(t, actionIssuesSettings, nil)
	require.Equal(t, 200, resp.Status)
	require.EqualValues(t, 5, out["pollMinutes"])
	resp, out = r.call(t, actionSetPollInterval, map[string]any{"minutes": 0.5})
	require.Equal(t, 400, resp.Status)
	require.Equal(t, "minutes", errorOf(t, out)["field"], "AC4.2.2")
	resp, out = r.call(t, actionSetPollInterval, map[string]any{"minutes": 2})
	require.Equal(t, 200, resp.Status)
	require.EqualValues(t, 2, out["pollMinutes"])
}

func TestU3_Actions_ErrorCodes(t *testing.T) {
	r := newU3Rig(t)
	r.gw3.listErr = &backlog.Error{Kind: backlog.KindRateLimited, Status: 429, RetryAfter: 9 * time.Second}
	resp, out := r.call(t, actionIssuesList, nil)
	require.Equal(t, 429, resp.Status)
	require.Equal(t, "9", resp.Headers["Retry-After"])
	require.EqualValues(t, 9, errorOf(t, out)["retryAfterSeconds"])

	r.gw3.listErr = nil
	resp, _ = r.onTask(t, actionIssuesLink, "task-17", map[string]any{"issueKey": "PROJ-120"})
	require.Equal(t, 200, resp.Status)
	r.gw3.getErr = &backlog.Error{Kind: backlog.KindForbidden, Status: 403}
	resp, out = r.onTask(t, actionIssuesGet, "task-17", nil)
	require.Equal(t, 404, resp.Status, "an unavailable issue is not_found, never reconnect_required")
	require.Equal(t, "not_found", errorOf(t, out)["code"])

	resp, out = r.call(t, actionIssuesCreate, map[string]any{"issueKey": "PROJ-118"})
	require.Equal(t, 400, resp.Status)
	require.Equal(t, "workflowId", errorOf(t, out)["field"])
}

func TestU3_Actions_GuardRefusesAllWhileOff(t *testing.T) {
	r := newU3Rig(t)
	r.call(t, actionSetEnabled, map[string]any{"enabled": false})
	for _, key := range u3Actions {
		resp, out := r.onTask(t, key, "task-17", map[string]any{})
		require.Equal(t, 409, resp.Status, key)
		require.Equal(t, "integration_disabled", errorOf(t, out)["code"], key)
	}
	require.Zero(t, r.gw3.count("issues")+r.gw3.count("issue"))
}

func TestU3_Actions_FailureLogHasNoSecret(t *testing.T) {
	r := newU3Rig(t)
	key := testutil.APIKey(t)
	r.call(t, actionConnectAPIKey, connectBody(key))
	r.call(t, keySetProjects, map[string]any{"projectKeys": []string{"PROJ"}})
	r.gw3.listErr = &backlog.Error{Kind: backlog.KindUnreachable, Status: 500}
	resp, out := r.call(t, actionIssuesList, nil)
	require.Equal(t, 503, resp.Status)
	var failed int
	for _, line := range strings.Split(strings.TrimSpace(r.logs.String()), "\n") {
		if m := parseLine(t, line); m["event"] == "action_failed" && m["action"] == actionIssuesList {
			failed++
			require.Equal(t, errorOf(t, out)["requestId"], m["requestId"])
		}
	}
	require.Equal(t, 1, failed, "AC8.3.1")
	testutil.AssertNoLeak(t, r.logs.String()+string(resp.Body), key)
}

// u3Actions lists every action U3 adds.
var u3Actions = []string{
	actionIssuesList, actionIssuesFilters, actionIssuesCreate, actionTasksSearch, actionIssuesLinks, actionIssuesRefresh,
	actionIssuesImpact, actionIssuesSettings, actionIssuesLink, actionIssuesUnlink, actionIssuesGet, actionIssuesComments,
	actionSetPollInterval,
	// Intent 261007: issue watches (FR3).
	actionIssueWatchesList, actionIssueWatchesSave, actionIssueWatchesDelete, actionIssueWatchesRun,
	actionIssueWatchesPause, actionIssueWatchesResume,
	// Intent 261007-github-parity-actions: quick actions and saved issue queries.
	actionQuickActionsGet, actionQuickActionsSave,
	actionIssueQueriesList, actionIssueQueriesSave, actionIssueQueriesDelete, actionIssueQueriesDefault,
}

// FR2.2-FR2.4: quick actions come back resolved; a save replaces one kind.
func TestQuickActionActions_GetSaveAndValidate(t *testing.T) {
	r := newU3Rig(t)
	resp, out := r.call(t, actionQuickActionsGet, nil)
	require.Equal(t, 200, resp.Status)
	require.Len(t, out["issue"], 3)
	require.Len(t, out["pr"], 3)
	require.Equal(t, "implement", out["issue"].([]any)[0].(map[string]any)["id"])

	resp, out = r.call(t, actionQuickActionsSave, map[string]any{"kind": "issue", "actions": []map[string]any{
		{"label": "Triage", "hint": "Sort it", "icon": "check", "promptTemplate": "Triage {{url}}"}}})
	require.Equal(t, 200, resp.Status)
	require.Len(t, out["issue"], 1)
	first := out["issue"].([]any)[0].(map[string]any)
	require.Equal(t, "Triage", first["label"])
	require.NotEmpty(t, first["id"])
	require.Len(t, out["pr"], 3, "the PR defaults stay")

	resp, out = r.call(t, actionQuickActionsSave, map[string]any{"kind": "issue", "actions": []map[string]any{{"label": "", "icon": "eye"}}})
	require.Equal(t, 400, resp.Status)
	require.Equal(t, "label", errorOf(t, out)["field"])
	resp, out = r.call(t, actionQuickActionsSave, []byte(`{"kind":1}`))
	require.Equal(t, 400, resp.Status)
	require.Equal(t, "validation", errorOf(t, out)["code"])
}

// FR4.3, FR4.4: saved issue queries with one starred default.
func TestIssueQueryActions_CRUDAndDefault(t *testing.T) {
	r := newU3Rig(t)
	resp, out := r.call(t, actionIssueQueriesSave, map[string]any{"name": "My bugs", "assignee": "me", "statusIds": []int{1, 2}, "keyword": "login"})
	require.Equal(t, 200, resp.Status)
	id := out["id"].(string)
	require.Equal(t, "me", out["assignee"])

	resp, out = r.call(t, actionIssueQueriesDefault, map[string]any{"id": id, "isDefault": true})
	require.Equal(t, 200, resp.Status)
	queries := out["queries"].([]any)
	require.Equal(t, true, queries[0].(map[string]any)["isDefault"])

	_, out = r.call(t, actionIssueQueriesList, nil)
	require.Len(t, out["queries"], 1)

	resp, out = r.call(t, actionIssueQueriesSave, map[string]any{"name": "Bad", "assignee": "someone"})
	require.Equal(t, 400, resp.Status)
	require.Equal(t, "assignee", errorOf(t, out)["field"])
	resp, out = r.call(t, actionIssueQueriesDefault, map[string]any{"id": "missing", "isDefault": true})
	require.Equal(t, 404, resp.Status)
	require.Equal(t, "not_found", errorOf(t, out)["code"])

	resp, _ = r.call(t, actionIssueQueriesDelete, map[string]string{"id": id})
	require.Equal(t, 200, resp.Status)
	resp, out = r.call(t, actionIssueQueriesDelete, map[string]string{"id": id})
	require.Equal(t, 404, resp.Status)
	require.Equal(t, "not_found", errorOf(t, out)["code"])

	resp, out = r.call(t, actionIssuesList, map[string]any{"assignee": "nobody"})
	require.Equal(t, 400, resp.Status, "FR4.2: issues.list only accepts me")
	require.Equal(t, "assignee", errorOf(t, out)["field"])
	resp, _ = r.call(t, actionIssuesList, map[string]any{"assignee": "me"})
	require.Equal(t, 200, resp.Status)
}
