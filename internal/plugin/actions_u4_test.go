package plugin

import (
	"context"
	"encoding/json"
	"fmt"
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

const spaceHost = "example-space.backlog.com"

// The base fakeGateway answers U4 calls with nothing; u4Gateway scripts them.
func (g *fakeGateway) Repositories(context.Context, backlog.Credentials, string) ([]backlog.Repository, error) {
	return nil, nil
}

func (g *fakeGateway) CheckGitAccess(context.Context, string, string, string, string, string) error {
	return nil
}

func (g *fakeGateway) PullRequests(context.Context, backlog.Credentials, backlog.CallClass, string, string, backlog.PullRequestQuery) ([]backlog.PullRequest, error) {
	return nil, nil
}

func (g *fakeGateway) PullRequestCount(context.Context, backlog.Credentials, backlog.CallClass, string, string, backlog.PullRequestQuery) (int, error) {
	return 0, nil
}

func (g *fakeGateway) PullRequest(context.Context, backlog.Credentials, backlog.CallClass, string, string, int) (backlog.PullRequest, error) {
	return backlog.PullRequest{}, &backlog.Error{Kind: backlog.KindNotFound, Status: 404}
}

func (g *fakeGateway) CreatePullRequest(context.Context, backlog.Credentials, string, string, backlog.NewPullRequest) (backlog.PullRequest, error) {
	return backlog.PullRequest{}, &backlog.Error{Kind: backlog.KindUnreachable, Status: 503}
}

func (g *fakeGateway) Issue(context.Context, backlog.Credentials, backlog.CallClass, string) (backlog.Issue, error) {
	return backlog.Issue{}, &backlog.Error{Kind: backlog.KindNotFound, Status: 404}
}

type u4Gateway struct {
	*fakeGateway
	mu      sync.Mutex
	prs     []backlog.PullRequest // PROJ/web-app
	prsErr  error
	created []backlog.NewPullRequest
	paths   []string // "project/repo" of every PR call, in order
}

func (g *u4Gateway) hitPath(project, repo string) {
	g.paths = append(g.paths, project+"/"+repo)
}

func (g *u4Gateway) pathLog() []string {
	g.mu.Lock()
	defer g.mu.Unlock()
	return slices.Clone(g.paths)
}

func (g *u4Gateway) Repositories(context.Context, backlog.Credentials, string) ([]backlog.Repository, error) {
	return []backlog.Repository{{ID: 11, ProjectID: 101, Name: "web-app", HTTPURL: "https://" + spaceHost + "/git/PROJ/web-app.git"}}, nil
}

func (g *u4Gateway) PullRequests(_ context.Context, _ backlog.Credentials, _ backlog.CallClass, project, repo string, q backlog.PullRequestQuery) ([]backlog.PullRequest, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.hitPath(project, repo)
	if g.prsErr != nil {
		return nil, g.prsErr
	}
	var out []backlog.PullRequest
	for _, pr := range g.prs {
		if len(q.StatusIDs) == 0 || slices.Contains(q.StatusIDs, int64(pr.StatusID)) {
			out = append(out, pr)
		}
	}
	return out, nil
}

func (g *u4Gateway) PullRequest(_ context.Context, _ backlog.Credentials, _ backlog.CallClass, project, repo string, n int) (backlog.PullRequest, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.hitPath(project, repo)
	for _, pr := range g.prs {
		if pr.Number == n {
			return pr, nil
		}
	}
	return backlog.PullRequest{}, &backlog.Error{Kind: backlog.KindNotFound, Status: 404}
}

func (g *u4Gateway) CreatePullRequest(_ context.Context, _ backlog.Credentials, project, repo string, in backlog.NewPullRequest) (backlog.PullRequest, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.hitPath(project, repo)
	g.created = append(g.created, in)
	return backlog.PullRequest{RepositoryID: 11, Number: 50, Summary: in.Summary, Branch: in.Branch, StatusID: 1}, nil
}

// u4Host adds the Host data API (Tasks, Repositories) to the fake Host.
type u4Host struct {
	*fakeHost
	data *hostData
}

type hostData struct {
	mu         sync.Mutex
	tasks      []pluginsdk.Task
	creates    []pluginsdk.CreateTaskInput
	listPages  []pluginsdk.Page
	repos      []pluginsdk.Repository
	repoPages  []pluginsdk.Page
	failCreate bool
	createErr  error // when set, Create fails with it
	// crashAfterCreate makes every ledger write fail once a task is created,
	// like a crash between Tasks().Create and storing the task id.
	crashAfterCreate bool
	crashed          bool
}

// SetState fails ledger writes after a crash.
func (h *u4Host) SetState(ctx context.Context, scope, scopeID, key string, value map[string]any) error {
	h.data.mu.Lock()
	crashed := h.data.crashed
	h.data.mu.Unlock()
	if crashed && key == "git.ledger" {
		return fmt.Errorf("state unavailable")
	}
	return h.fakeHost.SetState(ctx, scope, scopeID, key, value)
}

// kandevMetadata stores plugin metadata the way Kandev v0.96.0 does
// (pluginTaskMetadata): nested under "plugin:<id>" next to "source".
func kandevMetadata(in map[string]any) map[string]any {
	md := map[string]any{"source": "plugin:nulab-backlog"}
	if len(in) > 0 {
		md["plugin:nulab-backlog"] = in
	}
	return md
}

func (h *u4Host) Tasks() pluginsdk.TaskReader              { return fakeTasks{d: h.data} }
func (h *u4Host) Repositories() pluginsdk.RepositoryReader { return fakeRepos{h.data} }

type fakeTasks struct {
	pluginsdk.TaskReader
	d *hostData
}

func (f fakeTasks) Create(_ context.Context, in pluginsdk.CreateTaskInput) (*pluginsdk.Task, error) {
	f.d.mu.Lock()
	defer f.d.mu.Unlock()
	if f.d.failCreate {
		return nil, fmt.Errorf("rpc error: code = PermissionDenied")
	}
	if f.d.createErr != nil {
		return nil, f.d.createErr
	}
	f.d.creates = append(f.d.creates, in)
	n := len(f.d.tasks) + 1 // U3: Kandev returns the task's human key as Identifier
	t := pluginsdk.Task{ID: fmt.Sprintf("task-%d", n), Identifier: fmt.Sprintf("T-%d", n), WorkspaceID: in.WorkspaceID,
		Title: in.Title, Metadata: kandevMetadata(in.Metadata)}
	f.d.tasks = append(f.d.tasks, t)
	f.d.crashed = f.d.crashAfterCreate
	return &t, nil
}

// List pages two tasks at a time.
func (f fakeTasks) List(_ context.Context, filter pluginsdk.TaskFilter, page pluginsdk.Page) ([]pluginsdk.Task, *pluginsdk.PageInfo, error) {
	f.d.mu.Lock()
	defer f.d.mu.Unlock()
	f.d.listPages = append(f.d.listPages, page)
	if !filter.IncludeArchived || len(filter.WorkspaceIDs) != 1 {
		return nil, nil, fmt.Errorf("unexpected filter")
	}
	start := 0
	if page.Cursor != "" {
		_, _ = fmt.Sscan(page.Cursor, &start)
	}
	end := min(start+2, len(f.d.tasks))
	info := &pluginsdk.PageInfo{HasMore: end < len(f.d.tasks), NextCursor: fmt.Sprint(end)}
	return slices.Clone(f.d.tasks[start:end]), info, nil
}

type fakeRepos struct{ d *hostData }

// List pages one repository at a time.
func (f fakeRepos) List(_ context.Context, _ string, page pluginsdk.Page) ([]pluginsdk.Repository, *pluginsdk.PageInfo, error) {
	f.d.mu.Lock()
	defer f.d.mu.Unlock()
	f.d.repoPages = append(f.d.repoPages, page)
	start := 0
	if page.Cursor != "" {
		_, _ = fmt.Sscan(page.Cursor, &start)
	}
	end := min(start+1, len(f.d.repos))
	return slices.Clone(f.d.repos[start:end]), &pluginsdk.PageInfo{HasMore: end < len(f.d.repos), NextCursor: fmt.Sprint(end)}, nil
}

type u4rig struct {
	*rig
	gw4  *u4Gateway
	data *hostData
}

// backlogRepo is shaped like Kandev v0.96.0 returns a provider repository:
// Name is "<owner>/<providerName>", the Backlog name is ProviderName.
func backlogRepo() pluginsdk.Repository {
	main := "main"
	return pluginsdk.Repository{ID: "repo-k1", WorkspaceID: "ws-1", Name: "PROJ/web-app", DefaultBranch: &main, ProviderID: "nulab-backlog",
		ProviderRepositoryID: "11", ProviderHost: "https://" + spaceHost, ProviderScope: spaceHost, OwnerOrProject: "PROJ", ProviderName: "web-app",
		RemoteURL: "https://" + spaceHost + "/git/PROJ/web-app.git"}
}

// newU4Rig is connected to example-space.backlog.com with PROJ selected.
func newU4Rig(t *testing.T) *u4rig {
	t.Helper()
	gw := &u4Gateway{fakeGateway: &fakeGateway{projects: []backlog.Project{{ID: 101, Key: "PROJ"}}}}
	logs, host := &syncBuffer{}, newFakeHost()
	rt := newRuntime(gw, logs, "debug")
	data := &hostData{repos: []pluginsdk.Repository{{ID: "repo-gh", ProviderID: "github"}, backlogRepo()}}
	rt.SetHost(&u4Host{fakeHost: host, data: data})
	r := &u4rig{rig: &rig{rt: rt, host: host, gw: gw.fakeGateway, logs: logs}, gw4: gw, data: data}
	r.connected(t)
	resp, _ := r.call(t, keySetProjects, map[string]any{"projectKeys": []string{"PROJ"}})
	require.Equal(t, 200, resp.Status)
	return r
}

// taskCall invokes a task-scoped action with a verified task, repository,
// session and head branch.
func (r *u4rig) taskCall(t *testing.T, key string, body any) (*pluginsdk.PluginActionResponse, map[string]any) {
	t.Helper()
	raw, err := json.Marshal(body)
	require.NoError(t, err)
	resp, err := r.rt.HandleAction(context.Background(), &pluginsdk.PluginActionRequest{
		ActionKey: key, Body: raw,
		Context: pluginsdk.VerifiedActionContext{WorkspaceID: "ws-1", ActorID: "user-1", TaskID: "task-17",
			RepositoryID: "repo-k1", SessionID: "session-1", HeadBranch: "feature/search"},
	})
	require.NoError(t, err)
	var out map[string]any
	require.NoError(t, json.Unmarshal(resp.Body, &out))
	return resp, out
}

func TestU4_Actions_SetGitCredential(t *testing.T) {
	r := newU4Rig(t)
	pw := testutil.Token(t)
	resp, out := r.call(t, actionSetGitCredential, map[string]string{"gitUsername": "lan", "gitPassword": pw})
	require.Equal(t, 200, resp.Status)
	require.Equal(t, true, out["hasGitCredential"])
	resp, out = r.call(t, actionSetGitCredential, map[string]string{"gitUsername": "", "gitPassword": pw})
	require.Equal(t, 400, resp.Status)
	require.Equal(t, "gitUsername", errorOf(t, out)["field"])
	resp, _ = r.call(t, actionSetGitCredential, []byte("{nope"))
	require.Equal(t, 400, resp.Status)
	testutil.AssertNoLeak(t, r.logs.String(), pw)
}

func TestU4_Actions_RepositorySource(t *testing.T) {
	r := newU4Rig(t)
	resp, out := r.call(t, actionReposInspect, map[string]string{"url": "https://" + spaceHost + "/git/PROJ/web-app.git"})
	require.Equal(t, 200, resp.Status)
	repo := out["repository"].(map[string]any)
	require.Equal(t, "nulab-backlog", repo["provider_id"])
	require.Equal(t, "https://"+spaceHost, repo["provider_host"])
	require.Equal(t, "master", repo["default_branch"])
	_, out = r.call(t, actionReposInspect, map[string]string{"url": "https://github.com/acme/web.git"})
	require.Equal(t, map[string]any{"matched": false}, out)

	r.gw4.prs = []backlog.PullRequest{{RepositoryID: 11, Number: 1, Base: "main", Branch: "feature/a", StatusID: 1}}
	resp, out = r.call(t, actionReposBranches, map[string]any{"repository": repo})
	require.Equal(t, 200, resp.Status)
	require.Equal(t, []any{map[string]any{"name": "feature/a"}, map[string]any{"name": "main", "is_default": true}}, out["branches"])

	resp, out = r.call(t, actionReposList, map[string]any{"query": "web"})
	require.Equal(t, 200, resp.Status)
	require.Len(t, out["repositories"], 1)
	resp, out = r.call(t, actionReposList, map[string]any{"cursor": "bad"})
	require.Equal(t, 400, resp.Status)
	require.Equal(t, "cursor", errorOf(t, out)["field"])
}

func TestU4_Actions_LinkStatusAndUnlink(t *testing.T) {
	r := newU4Rig(t)
	r.gw4.prs = []backlog.PullRequest{{RepositoryID: 11, Number: 42, Summary: "Add login page", StatusID: 1, AssigneeName: "Lan"}}
	resp, out := r.taskCall(t, actionPRLink, map[string]string{"reference": "42"})
	require.Equal(t, 200, resp.Status)
	require.Equal(t, "task-17", out["taskId"])

	resp, out = r.taskCall(t, actionPRStatus, map[string]any{})
	require.Equal(t, 200, resp.Status)
	summary := out["summaries"].([]any)[0].(map[string]any)
	require.Equal(t, "Open – Lan", summary["statusBadge"].(map[string]any)["label"])

	resp, out = r.call(t, actionLinksList, nil)
	require.Equal(t, 200, resp.Status)
	require.Len(t, out["associations"], 1)
	resp, out = r.call(t, actionImpact, map[string]any{"projectKeys": []string{"PROJ"}})
	require.Equal(t, 200, resp.Status)
	require.EqualValues(t, 1, out["prLinks"])

	resp, out = r.taskCall(t, actionPRLink, map[string]string{"reference": "999"})
	require.Equal(t, 404, resp.Status)
	require.Equal(t, "not_found", errorOf(t, out)["code"])
	resp, out = r.taskCall(t, actionPRLink, map[string]string{"reference": "abc"})
	require.Equal(t, 400, resp.Status)
	require.Equal(t, "reference", errorOf(t, out)["field"])

	resp, _ = r.taskCall(t, actionPRUnlink, map[string]string{"reviewKey": spaceHost + "|11|42"})
	require.Equal(t, 200, resp.Status)
	_, out = r.call(t, actionLinksList, nil)
	require.Empty(t, out["associations"])
}

func TestU4_Actions_CreateUsesTheVerifiedContext(t *testing.T) {
	r := newU4Rig(t)
	resp, out := r.taskCall(t, actionPRCreate, map[string]any{"title": "Add search", "body": "", "branch": "evil", "repositoryId": "repo-gh"})
	require.Equal(t, 200, resp.Status)
	require.Equal(t, "https://"+spaceHost+"/git/PROJ/web-app/pullRequests/50", out["url"])
	require.Equal(t, true, out["linked"])
	require.Equal(t, "feature/search", r.gw4.created[0].Branch, "HeadBranch comes from the verified context, never the body")
	require.Equal(t, "main", r.gw4.created[0].Base)
	require.Equal(t, []string{"PROJ/web-app", "PROJ/web-app"}, r.gw4.pathLog(), "the Backlog repo is ProviderName, not Kandev's owner/name")

	r.gw4.prs = []backlog.PullRequest{{RepositoryID: 11, Number: 7, Branch: "feature/search", StatusID: 1}}
	resp, out = r.taskCall(t, actionPRCreate, map[string]any{"title": "Again"})
	require.Equal(t, 409, resp.Status)
	require.Equal(t, "conflict", errorOf(t, out)["code"])
	require.EqualValues(t, 7, errorOf(t, out)["pullRequestNumber"])

	resp, out = r.taskCall(t, actionPRCreate, map[string]any{"title": " "})
	require.Equal(t, 400, resp.Status)
	require.Equal(t, "title", errorOf(t, out)["field"])
}

func TestU4_Actions_WatchesAndQueries(t *testing.T) {
	r := newU4Rig(t)
	resp, out := r.call(t, actionWatchesSave, map[string]any{"name": "Reviews", "projectKey": "PROJ", "repoName": "web-app",
		"statuses": []string{"open"}, "assignee": "anyone", "creator": "anyone", "workflowId": "wf-1"})
	require.Equal(t, 200, resp.Status)
	id := out["id"].(string)
	require.Equal(t, "active", out["state"])
	for _, step := range [][2]string{{actionWatchesPause, "paused"}, {actionWatchesResume, "active"}} {
		resp, out = r.call(t, step[0], map[string]string{"id": id})
		require.Equal(t, 200, resp.Status, step[0])
		require.Equal(t, step[1], out["state"], step[0])
	}
	resp, out = r.call(t, actionWatchesRun, map[string]string{"id": id})
	require.Equal(t, 200, resp.Status)
	require.Equal(t, true, out["queued"])
	_, out = r.call(t, actionWatchesList, nil)
	require.Len(t, out["watches"], 1)
	resp, _ = r.call(t, actionWatchesDelete, map[string]string{"id": id})
	require.Equal(t, 200, resp.Status)
	resp, out = r.call(t, actionWatchesDelete, map[string]string{"id": id})
	require.Equal(t, 404, resp.Status)
	require.Equal(t, "not_found", errorOf(t, out)["code"])
	resp, out = r.call(t, actionWatchesSave, map[string]any{"name": ""})
	require.Equal(t, 400, resp.Status)
	require.Equal(t, "name", errorOf(t, out)["field"])

	resp, out = r.call(t, actionQueriesSave, map[string]any{"name": "Open", "projectKey": "PROJ", "repoName": "web-app",
		"statuses": []string{"open"}, "assignee": "anyone"})
	require.Equal(t, 200, resp.Status)
	qid := out["id"].(string)
	r.gw4.prs = []backlog.PullRequest{{RepositoryID: 11, Number: 3, Summary: "x", StatusID: 1}}
	resp, out = r.call(t, actionQueriesRun, map[string]string{"id": qid})
	require.Equal(t, 200, resp.Status)
	require.Len(t, out["rows"], 1)
	_, out = r.call(t, actionQueriesList, nil)
	require.Len(t, out["queries"], 1)
	resp, _ = r.call(t, actionQueriesDelete, map[string]string{"id": qid})
	require.Equal(t, 200, resp.Status)

	r.gw4.prsErr = &backlog.Error{Kind: backlog.KindRateLimited, Status: 429, RetryAfter: 9 * time.Second}
	resp, out = r.call(t, actionQueriesSave, map[string]any{"name": "Open", "projectKey": "PROJ", "repoName": "web-app",
		"statuses": []string{"open"}, "assignee": "anyone"})
	require.Equal(t, 200, resp.Status)
	resp, _ = r.call(t, actionQueriesRun, map[string]string{"id": out["id"].(string)})
	require.Equal(t, 429, resp.Status)
	require.Equal(t, "9", resp.Headers["Retry-After"])
}

func TestU4_Actions_GuardRefusesAllWhileOff(t *testing.T) {
	r := newU4Rig(t)
	r.call(t, actionSetEnabled, map[string]any{"enabled": false})
	for _, key := range u4Actions {
		resp, out := r.taskCall(t, key, map[string]any{})
		require.Equal(t, 409, resp.Status, key)
		require.Equal(t, "integration_disabled", errorOf(t, out)["code"], key)
	}
}

func TestU4_Actions_FailureLogHasNoSecret(t *testing.T) {
	r := newU4Rig(t)
	r.gw4.prsErr = &backlog.Error{Kind: backlog.KindUnreachable, Status: 500}
	resp, out := r.call(t, actionReposInspect, map[string]string{"url": "https://" + spaceHost + "/git/PROJ/web-app.git"})
	require.Equal(t, 503, resp.Status)
	var failed int
	for _, line := range strings.Split(strings.TrimSpace(r.logs.String()), "\n") {
		if m := parseLine(t, line); m["event"] == "action_failed" && m["action"] == actionReposInspect {
			failed++
			require.Equal(t, errorOf(t, out)["requestId"], m["requestId"])
		}
	}
	require.Equal(t, 1, failed)
}

// u4Actions lists every action U4 adds.
var u4Actions = []string{
	actionSetGitCredential, actionReposInspect, actionReposBranches, actionReposList,
	actionPRLink, actionPRUnlink, actionPRCreate, actionPRStatus, actionLinksList, actionImpact,
	actionWatchesList, actionWatchesSave, actionWatchesDelete, actionWatchesRun, actionWatchesPause, actionWatchesResume,
	actionQueriesList, actionQueriesSave, actionQueriesDelete, actionQueriesRun,
	actionPRList,         // intent 261007 (FR4)
	actionQueriesDefault, // intent 261007-github-parity-actions (FR3.3)
}

// FR3.3: one starred PR query; an unknown id is not_found.
func TestQueryDefaultAction_MovesTheStar(t *testing.T) {
	r := newU4Rig(t)
	body := map[string]any{"projectKey": "PROJ", "repoName": "web-app", "statuses": []string{"open"}, "assignee": "me"}
	ids := []string{}
	for _, name := range []string{"A", "B"} {
		body["name"] = name
		resp, out := r.call(t, actionQueriesSave, body)
		require.Equal(t, 200, resp.Status)
		ids = append(ids, out["id"].(string))
	}
	for _, id := range ids {
		resp, out := r.call(t, actionQueriesDefault, map[string]any{"id": id, "isDefault": true})
		require.Equal(t, 200, resp.Status)
		for _, q := range out["queries"].([]any) {
			m := q.(map[string]any)
			require.Equal(t, m["id"] == id, m["isDefault"] == true, "only %s is starred", id)
		}
	}
	resp, out := r.call(t, actionQueriesDefault, map[string]any{"id": "missing", "isDefault": true})
	require.Equal(t, 404, resp.Status)
	require.Equal(t, "not_found", errorOf(t, out)["code"])
	resp, out = r.call(t, actionQueriesDefault, []byte(`{"id":5}`))
	require.Equal(t, 400, resp.Status)
	require.Equal(t, "validation", errorOf(t, out)["code"])
}
