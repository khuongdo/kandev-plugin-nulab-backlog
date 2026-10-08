package plugin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync"
	"testing"
	"testing/synctest"
	"time"

	"github.com/kandev/kandev/pkg/pluginsdk"
	"github.com/stretchr/testify/require"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/issues"
	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/scm"
	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/testutil"
)

// scmFake is one provider's scripted API for the action tests.
type scmFake struct {
	mu   sync.Mutex
	p    scm.Provider
	prs  []scm.PullRequest
	err  error     // every call fails with it when set
	user *scm.User // CurrentUser's answer when set; else Lan
}

func (f *scmFake) fail() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.err
}

func (f *scmFake) CurrentUser(context.Context, scm.Credential) (scm.User, error) {
	if f.user != nil {
		return *f.user, f.fail()
	}
	return scm.User{ID: "lan-id", Name: "Lan"}, f.fail()
}

func (f *scmFake) SearchRepos(context.Context, scm.Credential, string) ([]scm.Repo, error) {
	return []scm.Repo{{FullName: "acme/web"}}, f.fail()
}

func (f *scmFake) GetRepo(_ context.Context, _ scm.Credential, repo string) (scm.Repo, error) {
	return scm.Repo{FullName: repo}, f.fail()
}

func (f *scmFake) ListPRs(context.Context, scm.Credential, string, scm.ListQuery) (scm.PRPage, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return scm.PRPage{Items: slices.Clone(f.prs)}, f.err
}

func (f *scmFake) GetPR(_ context.Context, _ scm.Credential, _ string, n int) (scm.PullRequest, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, pr := range f.prs {
		if pr.Number == n {
			return pr, f.err
		}
	}
	return scm.PullRequest{}, &scm.HTTPError{Provider: f.p, Status: 404}
}

// Get serves Tasks().Get for scm.task_prs.list (FR5.4).
func (f fakeTasks) Get(_ context.Context, id string) (*pluginsdk.Task, error) {
	f.d.mu.Lock()
	defer f.d.mu.Unlock()
	for _, t := range f.d.tasks {
		if t.ID == id {
			return &t, nil
		}
	}
	return nil, fmt.Errorf("rpc error: code = NotFound")
}

type scmRig struct {
	*u4rig
	fakes map[scm.Provider]*scmFake
	token string
}

// newSCMRig is the U4 rig (connected, PROJ selected) with fake providers.
func newSCMRig(t *testing.T) *scmRig {
	t.Helper()
	r := &scmRig{u4rig: newU4Rig(t), fakes: map[scm.Provider]*scmFake{}, token: testutil.Token(t)}
	clients := map[scm.Provider]scm.Client{}
	for _, p := range scm.Providers {
		r.fakes[p] = &scmFake{p: p}
		clients[p] = r.fakes[p]
	}
	r.rt.wireSCM(clients)
	return r
}

// setToken connects p: it makes p active for the call, then clears the
// stored service, so it is derived like an upgraded workspace's (p alone, or
// pending when several are connected).
func (r *scmRig) setToken(t *testing.T, p scm.Provider) {
	t.Helper()
	r.use(t, string(p))
	resp, out := r.call(t, actionSCMSetToken, map[string]string{"provider": string(p), "token": r.token, "username": "lan"})
	require.Equal(t, 200, resp.Status, out)
	require.NoError(t, scm.NewStore(r.host).SetActive(context.Background(), "ws-1", ""))
}

// use makes service the active source control service.
func (r *scmRig) use(t *testing.T, service string) {
	t.Helper()
	resp, out := r.call(t, actionSCMActiveSet, map[string]string{"service": service})
	require.Equal(t, 200, resp.Status, out)
}

func (r *scmRig) mapWeb(t *testing.T, p scm.Provider) {
	t.Helper()
	r.setToken(t, p)
	resp, out := r.call(t, actionSCMMapping, map[string]any{"provider": p, "projectKey": "PROJ", "repos": []string{"acme/web"}})
	require.Equal(t, 200, resp.Status, out)
}

func scmPR(n int, title, branch string) scm.PullRequest {
	ref := scm.PRRef{Provider: scm.GitHub, Repo: "acme/web", Number: n}
	return scm.PullRequest{PRRef: ref, Title: title, State: "open", Author: "Lan", AuthorID: "lan-id",
		SourceBranch: branch, TargetBranch: "main", URL: scm.PRURL(ref)}
}

// FR2.1, FR2.2, NFR1: the token is stored as a secret and never returned or logged.
func TestSCM_Actions_TokenIsStoredButNeverReturned(t *testing.T) {
	r := newSCMRig(t)
	var replies []string
	resp, out := r.call(t, actionSCMProviders, nil)
	require.Equal(t, 200, resp.Status)
	require.Len(t, out["providers"], 3)
	r.setToken(t, scm.GitHub)
	for _, key := range []string{actionSCMProviders, actionSCMTest} {
		body := map[string]string{"provider": "github"}
		resp, out := r.call(t, key, body)
		require.Equal(t, 200, resp.Status, key)
		replies = append(replies, string(resp.Body))
		if key == actionSCMTest {
			require.Equal(t, "connected", out["state"])
			require.Equal(t, "Lan", out["account"])
		}
	}
	r.host.mu.Lock()
	stored := r.host.secrets["backlog.scm.github.ws-1"]
	state, _ := json.Marshal(r.host.state)
	r.host.mu.Unlock()
	require.Contains(t, stored, r.token)
	testutil.AssertNoLeak(t, strings.Join(replies, "\n")+r.logs.String()+string(state), r.token)

	resp, out = r.call(t, actionSCMRemove, map[string]string{"provider": "github"})
	require.Equal(t, 200, resp.Status)
	require.Equal(t, "not_configured", out["state"])
}

// FR6.1: while Backlog is off every scm action is refused except reading settings.
func TestSCM_Actions_GuardRefusesAllButProvidersList(t *testing.T) {
	r := newSCMRig(t)
	r.call(t, actionSetEnabled, map[string]any{"enabled": false})
	for _, key := range scmActions {
		resp, out := r.taskCall(t, key, map[string]any{})
		if key == actionSCMProviders {
			require.Equal(t, 200, resp.Status)
			continue
		}
		require.Equal(t, 409, resp.Status, key)
		require.Equal(t, "integration_disabled", errorOf(t, out)["code"], key)
	}
	require.False(t, guarded(actionSCMProviders))
}

// NFR5, NFR1: provider errors map to the existing codes with no body or token.
func TestSCM_Actions_ErrorMapping(t *testing.T) {
	cases := []struct {
		err    error
		status int
		code   string
	}{
		{&scm.HTTPError{Provider: scm.GitHub, Status: 401}, 401, "reconnect_required"},
		{&scm.HTTPError{Provider: scm.GitHub, Status: 403}, 401, "reconnect_required"},
		{&scm.HTTPError{Provider: scm.GitHub, Status: 404}, 404, "not_found"},
		{&scm.HTTPError{Provider: scm.GitHub, Status: 429, RetryAfter: 8500 * time.Millisecond}, 429, "rate_limited"},
		{&scm.HTTPError{Provider: scm.GitHub, Status: 502}, 503, "unreachable"},
		{&scm.HTTPError{Provider: scm.GitHub}, 503, "unreachable"},
		{scm.ErrHostRefused, 503, "unreachable"},
	}
	for _, c := range cases {
		r := newSCMRig(t)
		r.use(t, "github")
		r.fakes[scm.GitHub].err = c.err
		resp, out := r.call(t, actionSCMSetToken, map[string]string{"provider": "github", "token": r.token})
		require.Equal(t, c.status, resp.Status, c.err.Error())
		require.Equal(t, c.code, errorOf(t, out)["code"], c.err.Error())
		if c.code == "rate_limited" {
			require.Equal(t, "9", resp.Headers["Retry-After"])
		}
		testutil.AssertNoLeak(t, string(resp.Body)+r.logs.String(), r.token)
	}
	r := newSCMRig(t)
	r.use(t, "gitlab")
	resp, out := r.call(t, actionSCMRepos, map[string]string{"provider": "gitlab"})
	require.Equal(t, 400, resp.Status)
	require.Equal(t, "token", errorOf(t, out)["field"], "no token says where to add one")
	resp, out = r.call(t, actionSCMQueriesRun, map[string]string{"id": "missing"})
	require.Equal(t, 404, resp.Status)
	require.Equal(t, "not_found", errorOf(t, out)["code"])
	resp, _ = r.call(t, actionSCMPRList, []byte("{nope"))
	require.Equal(t, 400, resp.Status)
}

// FR3, FR4, FR5: mapping, the PR list with auto-links, links, queries and watches.
func TestSCM_Actions_EndToEnd(t *testing.T) {
	r := newSCMRig(t)
	r.mapWeb(t, scm.GitHub)
	r.fakes[scm.GitHub].prs = []scm.PullRequest{scmPR(42, "PROJ-1 Add login", "feature/PROJ-1")}

	resp, out := r.call(t, actionSCMRepos, map[string]string{"provider": "github", "query": "web"})
	require.Equal(t, 200, resp.Status)
	require.Len(t, out["repos"], 1)
	resp, out = r.call(t, actionSCMPRList, map[string]any{"provider": "github", "projectKey": "PROJ", "repo": "acme/web",
		"statuses": []string{"open"}, "author": "anyone", "page": 1})
	require.Equal(t, 200, resp.Status)
	row := out["items"].([]any)[0].(map[string]any)
	require.EqualValues(t, 42, row["number"])
	require.Equal(t, "feature/PROJ-1", row["sourceBranch"])

	resp, out = r.call(t, actionSCMLinks, map[string]string{"issueKey": "PROJ-1"})
	require.Equal(t, 200, resp.Status)
	auto := out["links"].([]any)[0].(map[string]any)
	require.Equal(t, true, auto["auto"], "FR5.2: auto-linked on the list refresh")

	resp, out = r.taskCall(t, actionSCMLink, map[string]string{"url": "https://github.com/acme/web/pull/42"})
	require.Equal(t, 200, resp.Status)
	require.Equal(t, "task-17", out["taskId"], "the task comes from the verified context")
	resp, out = r.call(t, actionSCMLinks, map[string]string{"taskId": "task-17", "issueKey": "PROJ-1"})
	require.Equal(t, 200, resp.Status)
	require.Len(t, out["links"], 2)
	resp, _ = r.taskCall(t, actionSCMUnlink, map[string]string{"key": "github|acme/web|42", "issueKey": "PROJ-1"})
	require.Equal(t, 200, resp.Status)
	resp, _ = r.taskCall(t, actionSCMUnlink, map[string]string{"key": "github|acme/web|42"})
	require.Equal(t, 200, resp.Status)

	query := map[string]any{"name": "Open", "provider": "github", "projectKey": "PROJ", "repo": "acme/web", "statuses": []string{"open"}}
	resp, out = r.call(t, actionSCMQueriesSave, query)
	require.Equal(t, 200, resp.Status)
	qid := out["id"].(string)
	resp, out = r.call(t, actionSCMQueriesDefault, map[string]any{"id": qid, "isDefault": true})
	require.Equal(t, 200, resp.Status)
	require.Equal(t, true, out["queries"].([]any)[0].(map[string]any)["isDefault"])
	resp, out = r.call(t, actionSCMQueriesRun, map[string]string{"id": qid})
	require.Equal(t, 200, resp.Status)
	require.Len(t, out["items"], 1)
	_, out = r.call(t, actionSCMQueriesList, nil)
	require.Len(t, out["queries"], 1)
	resp, _ = r.call(t, actionSCMQueriesDelete, map[string]string{"id": qid})
	require.Equal(t, 200, resp.Status)

	query["workflowId"] = "wf-1"
	resp, out = r.call(t, actionSCMWatchesSave, query)
	require.Equal(t, 200, resp.Status)
	wid := out["id"].(string)
	require.EqualValues(t, 5, out["intervalMinutes"])
	resp, out = r.call(t, actionSCMWatchesRun, map[string]string{"id": wid})
	require.Equal(t, 200, resp.Status)
	require.EqualValues(t, 1, out["created"], "FR4.3: one task per run")
	for _, step := range [][2]string{{actionSCMWatchesPause, "paused"}, {actionSCMWatchesResume, "active"}} {
		resp, out = r.call(t, step[0], map[string]string{"id": wid})
		require.Equal(t, 200, resp.Status, step[0])
		require.Equal(t, step[1], out["state"])
	}
	_, out = r.call(t, actionSCMWatchesList, nil)
	require.Len(t, out["watches"], 1)
	resp, _ = r.call(t, actionSCMWatchesDelete, map[string]string{"id": wid})
	require.Equal(t, 200, resp.Status)
	r.data.mu.Lock()
	require.Equal(t, "Review PR #42: PROJ-1 Add login", r.data.creates[0].Title)
	r.data.mu.Unlock()

	resp, out = r.call(t, actionSCMMapping, map[string]any{"provider": "github", "projectKey": "NOPE", "repos": []string{}})
	require.Equal(t, 400, resp.Status)
	require.Equal(t, "projectKey", errorOf(t, out)["field"])
}

// FR5.4: Kandev's own GitHub/GitLab PRs of every task linked to the task's
// issue, without any plugin token.
func TestSCM_Actions_TaskPRsFromKandev(t *testing.T) {
	r := newSCMRig(t)
	r.data.mu.Lock()
	r.data.tasks = []pluginsdk.Task{
		{ID: "task-17", PullRequests: []pluginsdk.TaskPullRequest{
			{Number: 5, URL: "https://github.com/acme/web/pull/5", Title: "Fix", State: "open", Provider: "github", HeadBranch: "fix", BaseBranch: "main"},
			{Number: 9, URL: "https://dev.azure.com/x/_git/y/pullrequest/9", State: "open", Provider: "azure_devops"},
		}},
		{ID: "task-2", PullRequests: []pluginsdk.TaskPullRequest{
			{Number: 3, URL: "https://gitlab.com/g/p/-/merge_requests/3", Title: "MR", State: "merged", Provider: "gitlab", IsDraft: true},
		}},
		{ID: "task-3", PullRequests: []pluginsdk.TaskPullRequest{{Number: 1, Provider: "github", URL: "https://github.com/x/y/pull/1"}}},
	}
	r.data.mu.Unlock()
	r.setToken(t, scm.GitHub)
	r.setToken(t, scm.GitLab) // two connected: pending, both shown (FR3.3)
	require.NoError(t, issues.NewStore(r.host).UpdateLinks(context.Background(), "ws-1", func([]issues.Link) ([]issues.Link, error) {
		return []issues.Link{
			{IssueKey: "PROJ-1", TaskID: "task-17", SpaceHost: spaceHost, State: "active"},
			{IssueKey: "PROJ-1", TaskID: "task-2", SpaceHost: spaceHost, State: "active"},
			{IssueKey: "PROJ-2", TaskID: "task-3", SpaceHost: spaceHost, State: "active"},
			{IssueKey: "PROJ-1", TaskID: "task-gone", SpaceHost: spaceHost, State: "active"},
		}, nil
	}))
	resp, out := r.taskCall(t, actionSCMTaskPRs, nil)
	require.Equal(t, 200, resp.Status)
	got := out["pullRequests"].([]any)
	require.Len(t, got, 2, "github and gitlab of task-17 and task-2; azure_devops, task-3 and a deleted task are left out")
	require.Equal(t, map[string]any{"taskId": "task-17", "number": float64(5), "url": "https://github.com/acme/web/pull/5",
		"title": "Fix", "state": "open", "provider": "github", "headBranch": "fix", "baseBranch": "main"}, got[0])
	require.Equal(t, true, got[1].(map[string]any)["isDraft"])

	r.use(t, "github")
	_, out = r.taskCall(t, actionSCMTaskPRs, nil)
	require.Len(t, out["pullRequests"], 1, "FR1.5: only the active service's pull requests")
	require.Equal(t, "github", out["pullRequests"].([]any)[0].(map[string]any)["provider"])
	r.use(t, "backlog_git")
	_, out = r.taskCall(t, actionSCMTaskPRs, nil)
	require.Empty(t, out["pullRequests"], "FR1.6: Backlog Git active, no GitHub or GitLab pull request")

	r2 := newSCMRig(t)
	resp, out = r2.taskCall(t, actionSCMTaskPRs, nil)
	require.Equal(t, 200, resp.Status)
	require.Empty(t, out["pullRequests"], "a task with no issue has none")
}

// The scm watcher starts and stops with the plugin and creates watch tasks.
func TestSCM_Runtime_WatcherStartsWithThePlugin(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		r := newSCMRig(t)
		r.mapWeb(t, scm.GitLab)
		f := r.fakes[scm.GitLab]
		f.prs = []scm.PullRequest{scmPR(1, "x", "b")}
		f.prs[0].Provider = scm.GitLab
		resp, _ := r.call(t, actionSCMWatchesSave, map[string]any{"name": "W", "provider": "gitlab", "projectKey": "PROJ",
			"repo": "acme/web", "statuses": []string{"open"}, "workflowId": "wf-1"})
		require.Equal(t, 200, resp.Status)
		r.rt.Start()
		time.Sleep(r.rt.scmWatcher.Every + time.Second)
		synctest.Wait()
		r.rt.Close()
		r.data.mu.Lock()
		defer r.data.mu.Unlock()
		require.Len(t, r.data.creates, 1)
		require.Equal(t, map[string]any{scm.MetadataKey: "gitlab|acme/web|1"}, r.data.creates[0].Metadata)
	})
}

func TestSCM_Actions_RefusedActionKeysAreValidation(t *testing.T) {
	r := newSCMRig(t)
	for _, key := range []string{actionSCMSetToken, actionSCMMapping, actionSCMQueriesSave, actionSCMWatchesSave} {
		resp, out := r.call(t, key, map[string]any{"provider": "azure_devops"})
		require.Equal(t, 400, resp.Status, key)
		require.Equal(t, "validation", errorOf(t, out)["code"], key)
	}
}

// scmActions lists every action this intent adds.
var scmActions = []string{
	actionSCMProviders, actionSCMSetToken, actionSCMTest, actionSCMRemove, actionSCMRepos, actionSCMMapping,
	actionSCMPRList, actionSCMLink, actionSCMUnlink, actionSCMLinks, actionSCMTaskPRs,
	actionSCMQueriesList, actionSCMQueriesSave, actionSCMQueriesDelete, actionSCMQueriesRun, actionSCMQueriesDefault,
	actionSCMWatchesList, actionSCMWatchesSave, actionSCMWatchesDelete, actionSCMWatchesRun, actionSCMWatchesPause,
	actionSCMWatchesResume, actionSCMUseCLI, actionSCMCLIAccounts, actionSCMActiveSet,
}

// FR2.6, FR4, FR5: every scm action is declared; settings changes are admin.
func TestSCM_Manifest_Actions(t *testing.T) {
	m := loadManifest(t)
	got := map[string]string{}
	for _, a := range m.Actions {
		if slices.Contains(scmActions, a.Key) {
			require.Equal(t, 16384, a.MaxBodyBytes, a.Key)
			got[a.Key] = a.Scope + "/" + a.Access
		}
	}
	want := map[string]string{}
	for _, k := range scmActions {
		want[k] = "workspace/authenticated"
	}
	for _, k := range []string{actionSCMSetToken, actionSCMTest, actionSCMRemove, actionSCMRepos, actionSCMMapping,
		actionSCMUseCLI, actionSCMCLIAccounts, actionSCMActiveSet} {
		want[k] = "workspace/admin"
	}
	for _, k := range []string{actionSCMLink, actionSCMUnlink, actionSCMTaskPRs} {
		want[k] = "task/authenticated"
	}
	require.Equal(t, want, got)
	require.Len(t, scmActions, 25)
}

// FR1.2, FR1.4, FR4.1, NFR1 (intent 261008-gh-cli-auth): scm.providers.use_cli
// connects with the server's CLI login; a missing CLI is cli_unavailable;
// the CLI token and its stderr never reach a reply or a log.
func TestSCM_Actions_UseCLI(t *testing.T) {
	r := newSCMRig(t)
	cliTok := testutil.Token(t)
	var calls []string
	r.rt.scm.CLI = func(_ context.Context, _ int, name string, args ...string) ([]byte, error) {
		calls = append(calls, name+" "+strings.Join(args, " "))
		return []byte(cliTok + "\n"), nil
	}
	r.use(t, "github")
	resp, out := r.call(t, actionSCMUseCLI, map[string]string{"provider": "github"})
	require.Equal(t, 200, resp.Status, out)
	require.Equal(t, "cli", out["method"])
	require.Equal(t, "connected", out["state"])
	require.Equal(t, "Lan", out["account"])
	require.Equal(t, "lan-id", out["login"])
	require.Equal(t, []string{"gh auth status --json hosts --hostname github.com", // no --json: the active account
		"gh auth token --hostname github.com", "gh auth token --hostname github.com --user lan-id"}, calls)
	replies := []string{string(resp.Body)}
	resp, _ = r.call(t, actionSCMProviders, nil)
	replies = append(replies, string(resp.Body))

	r.rt.scm.CLI = func(context.Context, int, string, ...string) ([]byte, error) {
		return nil, errors.New("exit status 1: glab not logged in " + cliTok)
	}
	r.use(t, "gitlab")
	resp, out = r.call(t, actionSCMUseCLI, map[string]string{"provider": "gitlab"})
	require.Equal(t, 503, resp.Status)
	require.Equal(t, "cli_unavailable", errorOf(t, out)["code"])
	replies = append(replies, string(resp.Body))

	r.use(t, "bitbucket")
	resp, out = r.call(t, actionSCMUseCLI, map[string]string{"provider": "bitbucket"})
	require.Equal(t, 400, resp.Status)
	require.Equal(t, "provider", errorOf(t, out)["field"])

	r.host.mu.Lock()
	state, _ := json.Marshal(r.host.state)
	_, hasSecret := r.host.secrets["backlog.scm.github.ws-1"]
	r.host.mu.Unlock()
	require.False(t, hasSecret)
	all := strings.Join(replies, "\n") + r.logs.String() + string(state)
	testutil.AssertNoLeak(t, all, cliTok)
	require.NotContains(t, all, "glab not logged in", "the CLI output is never surfaced")
}

// FR4.1: classifySCM maps ErrCLIUnavailable to cli_unavailable.
func TestSCM_ClassifyCLIUnavailable(t *testing.T) {
	out, ok := classifySCM(fmt.Errorf("read token: %w", scm.ErrCLIUnavailable))
	require.True(t, ok)
	require.Equal(t, codeCLIUnavailable, out.Code)
	require.Equal(t, 503, statusFor(out.Code))
}

// ghCLI is a gh with alice (active) and bob logged in; a --user read prints
// tok, and any login but alice and bob is unknown.
func ghCLI(tok string) scm.CLIRunner {
	return func(_ context.Context, _ int, _ string, args ...string) ([]byte, error) {
		cmd := strings.Join(args, " ")
		switch {
		case strings.HasPrefix(cmd, "auth status"):
			return []byte(`{"hosts":{"github.com":[{"login":"alice","active":true,"state":"success"},` +
				`{"login":"bob","active":false,"state":"success"}]}}`), nil
		case strings.HasSuffix(cmd, "--user alice"), strings.HasSuffix(cmd, "--user bob"):
			return []byte(tok), nil
		}
		return nil, errors.New("exit status 1: no such account " + tok)
	}
}

// FR1.1, FR1.2, AC1.1.1, AC1.1.6 (intent 261008-gh-cli-profile):
// scm.providers.cli_accounts lists gh's logins with login and active only; a
// dead gh is cli_unavailable.
func TestSCM_Actions_CLIAccounts(t *testing.T) {
	r := newSCMRig(t)
	tok := testutil.Token(t)
	r.rt.scm.CLI = ghCLI(tok)
	resp, _ := r.call(t, actionSCMCLIAccounts, nil)
	require.Equal(t, 200, resp.Status)
	require.JSONEq(t, `{"accounts":[{"login":"alice","active":true},{"login":"bob","active":false}]}`, string(resp.Body))

	r.rt.scm.CLI = func(context.Context, int, string, ...string) ([]byte, error) {
		return nil, errors.New("exit status 1: " + tok)
	}
	resp, out := r.call(t, actionSCMCLIAccounts, nil)
	require.Equal(t, 503, resp.Status)
	require.Equal(t, "cli_unavailable", errorOf(t, out)["code"])
	testutil.AssertNoLeak(t, string(resp.Body)+r.logs.String(), tok)
}

// FR2.1, FR5.1, AC1.1.5, AC1.1.7, AC3.1.1, AC3.2.3: use_cli takes a login;
// an invalid one is a login field error, an unknown one cli_account_missing.
func TestSCM_Actions_UseCLIWithLogin(t *testing.T) {
	r := newSCMRig(t)
	tok := testutil.Token(t)
	r.rt.scm.CLI = ghCLI(tok)
	r.use(t, "github")
	r.fakes[scm.GitHub].user = &scm.User{ID: "bob", Name: "Bob"}
	resp, out := r.call(t, actionSCMUseCLI, map[string]string{"provider": "github", "login": "bob"})
	require.Equal(t, 200, resp.Status, out)
	require.Equal(t, "bob", out["login"])
	require.Equal(t, "Bob", out["account"])

	for _, c := range []struct{ provider, login string }{{"github", "-x"}, {"github", "a;b"}, {"gitlab", "bob"}} {
		resp, out = r.call(t, actionSCMUseCLI, map[string]string{"provider": c.provider, "login": c.login})
		require.Equal(t, 400, resp.Status, c)
		require.Equal(t, "validation", errorOf(t, out)["code"])
		require.Equal(t, "login", errorOf(t, out)["field"])
	}

	resp, out = r.call(t, actionSCMUseCLI, map[string]string{"provider": "github", "login": "carol"})
	require.Equal(t, 409, resp.Status)
	require.Equal(t, "cli_account_missing", errorOf(t, out)["code"])
	resp, _ = r.call(t, actionSCMProviders, nil)
	require.Equal(t, 200, resp.Status)
	require.Contains(t, string(resp.Body), `"login":"bob"`, "the failed change kept bob")
	testutil.AssertNoLeak(t, string(resp.Body)+r.logs.String(), tok)
}

// FR5.1: classifySCM maps ErrCLIAccountMissing to cli_account_missing (409).
func TestSCM_ClassifyCLIAccountMissing(t *testing.T) {
	out, ok := classifySCM(fmt.Errorf("read token: %w", scm.ErrCLIAccountMissing))
	require.True(t, ok)
	require.Equal(t, codeCLIAccountMissing, out.Code)
	require.Equal(t, 409, statusFor(out.Code))
}

// FR1.1, FR1.3, FR2.2 (intent 261008-source-control-settings): scm.active.set
// switches the service and answers with the providers; scm.providers.list
// returns the active service; there is no "none".
func TestSCM_Actions_ActiveServiceIsSetAndListed(t *testing.T) {
	r := newSCMRig(t)
	resp, out := r.call(t, actionSCMProviders, nil)
	require.Equal(t, 200, resp.Status)
	require.Equal(t, "backlog_git", out["active"], "FR1.2: nothing connected is Backlog Git")
	resp, out = r.call(t, actionSCMActiveSet, map[string]string{"service": "gitlab"})
	require.Equal(t, 200, resp.Status, out)
	require.Equal(t, "gitlab", out["active"])
	require.Len(t, out["providers"], 3)
	for _, bad := range []string{"", "none", "svn"} {
		resp, out = r.call(t, actionSCMActiveSet, map[string]string{"service": bad})
		require.Equal(t, 400, resp.Status, bad)
		require.Equal(t, "service", errorOf(t, out)["field"])
	}
	_, out = r.call(t, actionSCMProviders, nil)
	require.Equal(t, "gitlab", out["active"])

	r.setToken(t, scm.GitHub)
	r.setToken(t, scm.Bitbucket)
	_, out = r.call(t, actionSCMProviders, nil)
	require.Equal(t, "", out["active"], "FR3.3: several connected and none picked is pending")
}

// FR1.4, NFR1: an action of a non-active provider is service_inactive (409)
// naming the active service, with no token in the reply or the logs.
func TestSCM_Actions_InactiveProviderIsRefused(t *testing.T) {
	r := newSCMRig(t)
	r.use(t, "github")
	for key, body := range map[string]map[string]any{
		actionSCMSetToken: {"provider": "gitlab", "token": r.token},
		actionSCMTest:     {"provider": "bitbucket"},
		actionSCMPRList: {"provider": "gitlab", "projectKey": "PROJ", "repo": "acme/web",
			"statuses": []string{"open"}},
	} {
		resp, out := r.call(t, key, body)
		require.Equal(t, 409, resp.Status, key)
		e := errorOf(t, out)
		require.Equal(t, "service_inactive", e["code"], key)
		require.Equal(t, "github", e["activeService"], key)
		testutil.AssertNoLeak(t, string(resp.Body)+r.logs.String(), r.token)
	}
	resp, _ := r.call(t, actionSCMRemove, map[string]string{"provider": "gitlab"})
	require.Equal(t, 200, resp.Status, "removing a token stays allowed")
}

// FR1.6: while an external service is active, every Backlog Git feature is
// off: actions are refused or list nothing, and the Git credential is refused.
func TestSCM_Actions_BacklogGitIsOffWhileAnotherServiceIsActive(t *testing.T) {
	r := newSCMRig(t)
	pw := r.gitPassword(t)
	r.use(t, "github")
	for _, key := range []string{actionPRList, actionSetGitCredential, actionWatchesSave, actionQueriesRun} {
		resp, out := r.call(t, key, map[string]any{})
		require.Equal(t, 409, resp.Status, key)
		require.Equal(t, "service_inactive", errorOf(t, out)["code"], key)
		require.Equal(t, "github", errorOf(t, out)["activeService"], key)
	}
	for key, field := range map[string]string{actionWatchesList: "watches", actionQueriesList: "queries",
		actionLinksList: "associations", actionPRStatus: "summaries"} {
		resp, out := r.taskCall(t, key, nil)
		require.Equal(t, 200, resp.Status, key)
		require.Equal(t, []any{}, out[field], key)
	}
	_, err := r.rt.ResolveGitCredential(context.Background(), credRequest())
	require.Error(t, err)
	testutil.AssertNoLeak(t, err.Error()+r.logs.String(), pw)
	b, err := r.rt.GetGitCredentialBinding(context.Background(), &pluginsdk.GitCredentialBindingRequest{
		ProviderID: "nulab-backlog", WorkspaceID: "ws-1", TaskID: "task-17", Host: spaceHost, Path: "/git/PROJ/web-app.git"})
	require.NoError(t, err)
	require.Empty(t, b.Binding, "revoked while Backlog Git is off")

	r.use(t, "backlog_git")
	_, err = r.rt.ResolveGitCredential(context.Background(), credRequest())
	require.NoError(t, err, "Backlog Git active again: the credential works")
	resp, _ := r.call(t, actionPRList, map[string]any{})
	require.NotEqual(t, 409, resp.Status)
}
