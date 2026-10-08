package scm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/connection"
	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/testutil"
)

// FR2.1: every provider is listed, not configured until it has a token.
func TestProviders_ListsAllThreeNotConfigured(t *testing.T) {
	h := newHarness(t)
	views, err := h.svc.Providers(h.ctx, ws)
	require.NoError(t, err)
	require.Len(t, views, 3)
	for i, p := range Providers {
		require.Equal(t, ProviderView{Provider: p, State: StateNotConfigured, Mappings: []Mapping{}}, views[i])
	}
}

// FR2.2, FR2.4, NFR1: the token is checked, stored only as a secret, never shown.
func TestSetToken_StoresTheSecretAndTheAccount(t *testing.T) {
	h := newHarness(t)
	h.connect(t, GitHub)
	raw, ok, _ := h.secrets.GetSecret(context.Background(), "backlog.scm.github.ws-1")
	require.True(t, ok)
	var cred Credential
	require.NoError(t, json.Unmarshal([]byte(raw), &cred))
	require.Equal(t, h.tokens[GitHub], cred.Token)
	require.Equal(t, 1, h.clients[GitHub].count("CurrentUser"))

	views, err := h.svc.Providers(h.ctx, ws)
	require.NoError(t, err)
	require.Equal(t, ProviderView{Provider: GitHub, State: StateConnected, Method: MethodToken, Account: "Lan",
		Mappings: []Mapping{}}, views[0])
	b, _ := json.Marshal(views)
	state, _ := json.Marshal(h.state.data)
	h.assertNoTokenLeak(t, string(b), string(state))
}

func TestSetToken_ValidatesBeforeAnyCall(t *testing.T) {
	h := newHarness(t)
	for name, c := range map[string]struct {
		in    TokenInput
		field string
	}{
		"unknown provider":       {TokenInput{Provider: "gitea", Token: "abcd"}, FieldProvider},
		"empty token":            {TokenInput{Provider: GitHub, Token: " "}, FieldToken},
		"token with a space":     {TokenInput{Provider: GitHub, Token: "ab cd"}, FieldToken},
		"bitbucket without user": {TokenInput{Provider: Bitbucket, Token: "abcd"}, FieldUsername},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := h.svc.SetToken(h.ctx, ws, c.in)
			require.Equal(t, c.field, fieldOf(t, err))
		})
	}
	require.Zero(t, h.clients[GitHub].count("CurrentUser"))
}

// FR2.4: a refused token is not stored; the error carries no secret.
func TestSetToken_ARefusedTokenIsNotStored(t *testing.T) {
	h := newHarness(t)
	h.clients[GitLab].setErr("CurrentUser", &HTTPError{Provider: GitLab, Status: 401})
	_, err := h.svc.SetToken(h.ctx, ws, TokenInput{Provider: GitLab, Token: h.tokens[GitLab]})
	require.True(t, IsStatus(err, 401))
	require.False(t, h.hasToken(GitLab))
	h.assertNoTokenLeak(t, err.Error())
}

// FR2.4: test shows the account, or records a plain error code.
func TestTest_ReturnsTheAccountOrRecordsTheError(t *testing.T) {
	h := newHarness(t)
	_, err := h.svc.Test(h.ctx, ws, GitHub)
	require.ErrorIs(t, err, ErrNoToken)

	h.connect(t, GitHub)
	v, err := h.svc.Test(h.ctx, ws, GitHub)
	require.NoError(t, err)
	require.Equal(t, StateConnected, v.State)
	require.Equal(t, "Lan", v.Account)

	for status, code := range map[int]string{401: ErrorInvalidToken, 403: ErrorMissingScope, 429: ErrorRateLimited, 0: ErrorUnreachable} {
		h.clients[GitHub].setErr("CurrentUser", &HTTPError{Provider: GitHub, Status: status})
		v, err = h.svc.Test(h.ctx, ws, GitHub)
		require.NoError(t, err, "a failed test is a result, not an action error")
		require.Equal(t, StateError, v.State)
		require.Equal(t, code, v.LastError)
	}
	h.clients[GitHub].setErr("CurrentUser", nil)
	v, err = h.svc.Test(h.ctx, ws, GitHub)
	require.NoError(t, err)
	require.Equal(t, StateConnected, v.State)
	require.Empty(t, v.LastError)
}

// FR2.2, FR3.4: remove deletes the secret and keeps mappings, queries and links.
func TestRemoveToken_KeepsEverythingElse(t *testing.T) {
	h := newHarness(t)
	h.mapRepo(t, GitHub, "PROJ", "acme/web")
	_, err := h.svc.SaveQuery(h.ctx, ws, QueryInput{Name: "Open", Provider: GitHub, ProjectKey: "PROJ", Repo: "acme/web",
		Statuses: []string{"open"}})
	require.NoError(t, err)
	v, err := h.svc.RemoveToken(h.ctx, ws, GitHub)
	require.NoError(t, err)
	require.Equal(t, StateNotConfigured, v.State)
	require.Empty(t, v.Account)
	require.Equal(t, []Mapping{{ProjectKey: "PROJ", Repos: []string{"acme/web"}}}, v.Mappings)
	require.False(t, h.hasToken(GitHub))
	qs, err := h.svc.ListQueries(h.ctx, ws)
	require.NoError(t, err)
	require.Len(t, qs, 1)
	_, err = h.svc.ListPRs(h.ctx, ws, PRListInput{Provider: GitHub, ProjectKey: "PROJ", Repo: "acme/web", Statuses: []string{"open"}})
	require.ErrorIs(t, err, ErrNoToken, "disabled until a token exists")
}

// FR3.1, FR3.2: repositories of a selected project, validated by the provider.
func TestSetMapping_ValidatesThroughTheProvider(t *testing.T) {
	h := newHarness(t)
	_, err := h.svc.SetMapping(h.ctx, ws, MappingInput{Provider: GitHub, ProjectKey: "PROJ", Repos: []string{"acme/web"}})
	require.ErrorIs(t, err, ErrNoToken)
	h.connect(t, GitHub)
	v, err := h.svc.SetMapping(h.ctx, ws, MappingInput{Provider: GitHub, ProjectKey: "PROJ",
		Repos: []string{"ACME/web", "acme/web", "acme/api"}})
	require.NoError(t, err)
	require.Equal(t, []Mapping{{ProjectKey: "PROJ", Repos: []string{"acme/web", "acme/api"}}}, v.Mappings,
		"the provider's spelling, de-duplicated")

	for name, c := range map[string]struct {
		in    MappingInput
		field string
	}{
		"project not selected": {MappingInput{Provider: GitHub, ProjectKey: "OTHER", Repos: []string{"acme/web"}}, FieldProjectKey},
		"bad name":             {MappingInput{Provider: GitHub, ProjectKey: "PROJ", Repos: []string{"acme"}}, FieldRepos},
		"missing repository":   {MappingInput{Provider: GitHub, ProjectKey: "PROJ", Repos: []string{"acme/gone"}}, FieldRepos},
		"too many":             {MappingInput{Provider: GitHub, ProjectKey: "PROJ", Repos: manyRepos(21)}, FieldRepos},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := h.svc.SetMapping(h.ctx, ws, c.in)
			require.Equal(t, c.field, fieldOf(t, err))
		})
	}
	h.clients[GitHub].setErr("GetRepo", &HTTPError{Provider: GitHub, Status: 429, RetryAfter: time.Second})
	_, err = h.svc.SetMapping(h.ctx, ws, MappingInput{Provider: GitHub, ProjectKey: "PROJ", Repos: []string{"acme/web"}})
	require.True(t, IsStatus(err, 429), "a provider failure is not a validation error")

	v, err = h.svc.SetMapping(h.ctx, ws, MappingInput{Provider: GitHub, ProjectKey: "PROJ"})
	require.NoError(t, err)
	require.Empty(t, v.Mappings, "no repositories removes the mapping")
}

func manyRepos(n int) []string {
	out := make([]string, n)
	for i := range out {
		out[i] = fmt.Sprintf("acme/r%d", i)
	}
	return out
}

// FR3.2: the repository picker searches what the token can read.
func TestSearchRepos(t *testing.T) {
	h := newHarness(t)
	h.connect(t, GitLab)
	repos, err := h.svc.SearchRepos(h.ctx, ws, GitLab, "api")
	require.NoError(t, err)
	require.Equal(t, []Repo{{FullName: "acme/api"}}, repos)
	_, err = h.svc.SearchRepos(h.ctx, ws, "nope", "")
	require.Equal(t, FieldProvider, fieldOf(t, err))
}

// FR3.3, FR4.1: one mapped repository, 20 per page, with linked tasks.
func TestListPRs_MappedRepositoryOnly(t *testing.T) {
	h := newHarness(t)
	h.mapRepo(t, GitHub, "PROJ", "acme/web")
	h.clients[GitHub].setPRs("acme/web",
		pr(GitHub, "acme/web", 2, "Add search", "search", StateDraft),
		pr(GitHub, "acme/web", 1, "Fix", "fix", StateMerged))
	page, err := h.svc.ListPRs(h.ctx, ws, PRListInput{Provider: GitHub, ProjectKey: "PROJ", Repo: "acme/web",
		Statuses: []string{"open"}, Page: 1})
	require.NoError(t, err)
	require.Equal(t, 20, h.clients[GitHub].lastQ.PerPage)
	require.Equal(t, 1, page.Page)
	require.Equal(t, 20, page.PageSize)
	require.Len(t, page.Items, 1)
	require.Equal(t, "draft", page.Items[0].State)
	require.Equal(t, []string{}, page.Items[0].LinkedTaskIDs)

	_, err = h.svc.ListPRs(h.ctx, ws, PRListInput{Provider: GitHub, ProjectKey: "DEMO", Repo: "acme/web", Statuses: []string{"open"}})
	require.Equal(t, FieldRepository, fieldOf(t, err), "acme/web is mapped to PROJ, not DEMO")
	_, err = h.svc.ListPRs(h.ctx, ws, PRListInput{Provider: GitHub, ProjectKey: "PROJ", Repo: "acme/api", Statuses: []string{"open"}})
	require.Equal(t, FieldRepository, fieldOf(t, err))
	_, err = h.svc.ListPRs(h.ctx, ws, PRListInput{Provider: GitHub, ProjectKey: "PROJ", Repo: "acme/web", Statuses: []string{"x"}})
	require.Equal(t, FieldStatuses, fieldOf(t, err))
	_, err = h.svc.ListPRs(h.ctx, ws, PRListInput{Provider: GitHub, ProjectKey: "PROJ", Repo: "acme/web",
		Statuses: []string{"open"}, Author: "x"})
	require.Equal(t, FieldAuthor, fieldOf(t, err))
}

// FR4.1: "me" keeps the token account's pull requests.
func TestListPRs_AuthorMe(t *testing.T) {
	h := newHarness(t)
	h.mapRepo(t, GitLab, "PROJ", "acme/web")
	mine, theirs := pr(GitLab, "acme/web", 2, "Mine", "a", StateOpen), pr(GitLab, "acme/web", 1, "Theirs", "b", StateOpen)
	theirs.AuthorID = "someone"
	h.clients[GitLab].setPRs("acme/web", mine, theirs)
	page, err := h.svc.ListPRs(h.ctx, ws, PRListInput{Provider: GitLab, ProjectKey: "PROJ", Repo: "acme/web",
		Statuses: []string{"open"}, Author: WhoMe})
	require.NoError(t, err)
	require.Len(t, page.Items, 1)
	require.Equal(t, "Mine", page.Items[0].Title)
}

// NFR3: a provider's rate limit is passed through with its wait.
func TestListPRs_RateLimitPassesThrough(t *testing.T) {
	h := newHarness(t)
	h.mapRepo(t, Bitbucket, "PROJ", "acme/web")
	h.clients[Bitbucket].setErr("ListPRs", &HTTPError{Provider: Bitbucket, Status: 429, RetryAfter: 9 * time.Second})
	_, err := h.svc.ListPRs(h.ctx, ws, PRListInput{Provider: Bitbucket, ProjectKey: "PROJ", Repo: "acme/web", Statuses: []string{"open"}})
	var he *HTTPError
	require.ErrorAs(t, err, &he)
	require.Equal(t, 9*time.Second, he.RetryAfter)
}

// NFR8: the first page shows within 3 s when the provider answers within 1 s.
func TestListPRs_FirstPageWithinThreeSeconds(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		h := newHarness(t)
		h.mapRepo(t, GitHub, "PROJ", "acme/web")
		h.clients[GitHub].setPRs("acme/web", pr(GitHub, "acme/web", 1, "PROJ-1 x", "b", StateOpen))
		h.clients[GitHub].delay = time.Second
		start := time.Now()
		page, err := h.svc.ListPRs(h.ctx, ws, PRListInput{Provider: GitHub, ProjectKey: "PROJ", Repo: "acme/web",
			Statuses: []string{"open"}})
		require.NoError(t, err)
		require.Len(t, page.Items, 1)
		require.LessOrEqual(t, time.Since(start), 3*time.Second)
	})
}

// FR6.2: no ConnectionChanged reaction: items survive a disconnect, a space
// change and a project deselection, and keep working.
func TestCoupling_ItemsSurviveBacklogChanges(t *testing.T) {
	h := newHarness(t)
	h.mapRepo(t, GitHub, "PROJ", "acme/web")
	_, err := h.svc.SaveQuery(h.ctx, ws, QueryInput{Name: "Open", Provider: GitHub, ProjectKey: "PROJ", Repo: "acme/web",
		Statuses: []string{"open"}})
	require.NoError(t, err)
	_, err = h.svc.Link(h.ctx, ws, "task-1", "https://github.com/acme/web/pull/1")
	require.ErrorIs(t, err, ErrNotFound)
	h.clients[GitHub].setPRs("acme/web", pr(GitHub, "acme/web", 1, "x", "b", StateOpen))
	_, err = h.svc.Link(h.ctx, ws, "task-1", "https://github.com/acme/web/pull/1")
	require.NoError(t, err)

	for name, change := range map[string]func(c *fakeConn){
		"space change":       func(c *fakeConn) { c.snap.SpaceHost = "other.backlog.jp" },
		"project deselected": func(c *fakeConn) { c.snap.SelectedProjects = []string{"DEMO"} },
		"disconnect":         func(c *fakeConn) { c.err = connection.ErrNotConnected },
	} {
		t.Run(name, func(t *testing.T) {
			h.conn.set(change)
			views, err := h.svc.Providers(h.ctx, ws)
			require.NoError(t, err)
			require.Len(t, views[0].Mappings, 1)
			qs, err := h.svc.ListQueries(h.ctx, ws)
			require.NoError(t, err)
			require.Len(t, qs, 1)
			require.False(t, qs[0].Unmapped)
			links, err := h.svc.Links(h.ctx, ws, "task-1", "")
			require.NoError(t, err)
			require.Len(t, links, 1)
			rows, err := h.svc.RunQuery(h.ctx, ws, qs[0].ID)
			require.NoError(t, err)
			require.Len(t, rows.Items, 1)
		})
	}
}

// NFR1: every service path keeps tokens out of logs, errors and replies.
func TestRedaction_NoTokenInLogsErrorsOrReplies(t *testing.T) {
	h := newHarness(t)
	var texts []string
	for _, p := range Providers {
		h.mapRepo(t, p, "PROJ", "acme/web")
		h.clients[p].setErr("ListPRs", &HTTPError{Provider: p, Status: 401})
		_, err := h.svc.ListPRs(h.ctx, ws, PRListInput{Provider: p, ProjectKey: "PROJ", Repo: "acme/web", Statuses: []string{"open"}})
		require.Error(t, err)
		texts = append(texts, err.Error(), fmt.Sprintf("%v %+v", err, err))
		v, err := h.svc.Test(h.ctx, ws, p)
		require.NoError(t, err)
		b, _ := json.Marshal(v)
		texts = append(texts, string(b))
	}
	h.svc.RefreshLinks(h.ctx, ws)
	require.NotEmpty(t, h.logs.String())
	h.assertNoTokenLeak(t, texts...)
}

func TestCredentialRead_FailuresAreStoreErrors(t *testing.T) {
	h := newHarness(t)
	h.connect(t, GitHub)
	h.secrets.failGet = true
	_, err := h.svc.Test(h.ctx, ws, GitHub)
	require.ErrorIs(t, err, connection.ErrStore)
	h.secrets.failGet = false
	h.secrets.data[SecretKey(GitHub, ws)] = "{broken"
	_, err = h.svc.Test(h.ctx, ws, GitHub)
	require.ErrorIs(t, err, connection.ErrStore)
}

// settingsFor reads p's stored settings.
func (h *harness) settingsFor(t *testing.T, p Provider) Settings {
	t.Helper()
	st, err := h.svc.settings(h.ctx, ws, p)
	require.NoError(t, err)
	return st
}

func (h *harness) cliCached(p Provider) bool {
	h.svc.cli.mu.Lock()
	defer h.svc.cli.mu.Unlock()
	_, ok := h.svc.cli.tokens[p]
	return ok
}

// lastCred is the credential of p's last provider call.
func (h *harness) lastCred(p Provider) Credential {
	f := h.clients[p]
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.creds[len(f.creds)-1]
}

// FR1.2, FR1.3, FR2.1: the CLI token is checked, a typed token is deleted,
// only the method and account are stored.
func TestUseCLI_ConnectsGitHubAndGitLab(t *testing.T) {
	for _, p := range []Provider{GitHub, GitLab} {
		t.Run(string(p), func(t *testing.T) {
			h := newHarness(t)
			h.connect(t, p)
			cliTok := testutil.Token(t)
			cli := h.withCLI(cliTok)
			v, err := h.svc.UseCLI(h.ctx, ws, p)
			require.NoError(t, err)
			require.Equal(t, ProviderView{Provider: p, State: StateConnected, Method: MethodCLI, Account: "Lan",
				Mappings: []Mapping{}}, v)
			require.Equal(t, cliTok, h.lastCred(p).Token, "checked with the current user call")
			require.False(t, h.hasToken(p), "the typed token is deleted")
			st := h.settingsFor(t, p)
			require.Equal(t, Settings{Provider: p, Source: MethodCLI, HasToken: true, Account: "Lan", AccountID: "lan-id"}, st)
			require.Equal(t, 1, cli.count())
			state, _ := json.Marshal(h.state.data)
			testutil.AssertNoLeak(t, string(state)+h.logs.String(), cliTok)
		})
	}
}

// FR2.3: Bitbucket has no CLI login.
func TestUseCLI_BitbucketIsAFieldError(t *testing.T) {
	h := newHarness(t)
	h.withCLI("x")
	_, err := h.svc.UseCLI(h.ctx, ws, Bitbucket)
	require.Equal(t, FieldProvider, fieldOf(t, err))
	_, err = h.svc.UseCLI(h.ctx, ws, "gitea")
	require.Equal(t, FieldProvider, fieldOf(t, err))
}

// FR4.1: a failing CLI or a refused CLI token changes nothing.
func TestUseCLI_FailuresChangeNothing(t *testing.T) {
	h := newHarness(t)
	h.connect(t, GitHub)
	before := h.settingsFor(t, GitHub)
	cli := h.withCLI("")
	cli.set("", errors.New("exit status 1"))
	_, err := h.svc.UseCLI(h.ctx, ws, GitHub)
	require.ErrorIs(t, err, ErrCLIUnavailable)

	cli.set(testutil.Token(t), nil)
	h.clients[GitHub].setErr("CurrentUser", &HTTPError{Provider: GitHub, Status: 401})
	_, err = h.svc.UseCLI(h.ctx, ws, GitHub)
	require.True(t, IsStatus(err, 401))
	require.False(t, h.cliCached(GitHub), "a refused token is not kept")

	require.True(t, h.hasToken(GitHub), "the typed token stays")
	require.Equal(t, before, h.settingsFor(t, GitHub))
}

// FR3.1: a CLI provider's calls use the CLI token, never the secret store.
func TestCredential_CLIProviderNeverReadsTheSecretStore(t *testing.T) {
	h := newHarness(t)
	cliTok := testutil.Token(t)
	h.withCLI(cliTok)
	_, err := h.svc.UseCLI(h.ctx, ws, GitHub)
	require.NoError(t, err)
	h.secrets.failGet = true
	_, err = h.svc.SearchRepos(h.ctx, ws, GitHub, "web")
	require.NoError(t, err)
	require.Equal(t, cliTok, h.lastCred(GitHub).Token)
}

// FR1.3, FR5.3: a typed token replaces the CLI; remove clears the source and keeps mappings.
func TestSetTokenAndRemove_SwitchAwayFromTheCLI(t *testing.T) {
	h := newHarness(t)
	h.withCLI(testutil.Token(t))
	_, err := h.svc.UseCLI(h.ctx, ws, GitHub)
	require.NoError(t, err)
	_, err = h.svc.SetMapping(h.ctx, ws, MappingInput{Provider: GitHub, ProjectKey: "PROJ", Repos: []string{"acme/web"}})
	require.NoError(t, err)
	require.True(t, h.cliCached(GitHub))

	v, err := h.svc.SetToken(h.ctx, ws, TokenInput{Provider: GitHub, Token: h.tokens[GitHub]})
	require.NoError(t, err)
	require.Equal(t, MethodToken, v.Method)
	require.False(t, h.cliCached(GitHub))
	_, err = h.svc.SearchRepos(h.ctx, ws, GitHub, "")
	require.NoError(t, err)
	require.Equal(t, h.tokens[GitHub], h.lastCred(GitHub).Token)

	_, err = h.svc.UseCLI(h.ctx, ws, GitHub)
	require.NoError(t, err)
	v, err = h.svc.RemoveToken(h.ctx, ws, GitHub)
	require.NoError(t, err)
	require.Equal(t, StateNotConfigured, v.State)
	require.Empty(t, v.Method)
	require.Equal(t, []Mapping{{ProjectKey: "PROJ", Repos: []string{"acme/web"}}}, v.Mappings)
	require.False(t, h.cliCached(GitHub))
	require.Empty(t, h.settingsFor(t, GitHub).Source)
	_, err = h.svc.SearchRepos(h.ctx, ws, GitHub, "")
	require.ErrorIs(t, err, ErrNoToken)
}

// FR4.2, FR4.3: Test asks the CLI again, records cli_unavailable, and clears it once the CLI works.
func TestTest_CLIProviderRecordsAndClearsCLIUnavailable(t *testing.T) {
	h := newHarness(t)
	cli := h.withCLI(testutil.Token(t))
	_, err := h.svc.UseCLI(h.ctx, ws, GitHub)
	require.NoError(t, err)

	cli.set("", errors.New("exit status 1"))
	v, err := h.svc.Test(h.ctx, ws, GitHub)
	require.NoError(t, err, "a failed test is a result")
	require.Equal(t, StateError, v.State)
	require.Equal(t, ErrorCLIUnavailable, v.LastError)
	require.Equal(t, MethodCLI, v.Method)

	fresh := testutil.Token(t)
	cli.set(fresh, nil)
	v, err = h.svc.Test(h.ctx, ws, GitHub)
	require.NoError(t, err)
	require.Equal(t, StateConnected, v.State)
	require.Empty(t, v.LastError)
	require.Equal(t, fresh, h.lastCred(GitHub).Token, "the test does not reuse the cache")
}

// FR3.3: after the server's CLI login changes, calls use the new token once the cache expires.
func TestCredential_NewCLITokenAfterExpiry(t *testing.T) {
	h := newHarness(t)
	first, second := testutil.Token(t), testutil.Token(t)
	cli := h.withCLI(first)
	_, err := h.svc.UseCLI(h.ctx, ws, GitLab)
	require.NoError(t, err)
	cli.set(second, nil)
	_, err = h.svc.SearchRepos(h.ctx, ws, GitLab, "")
	require.NoError(t, err)
	require.Equal(t, first, h.lastCred(GitLab).Token)
	h.now = h.now.Add(cliTTL)
	_, err = h.svc.SearchRepos(h.ctx, ws, GitLab, "")
	require.NoError(t, err)
	require.Equal(t, second, h.lastCred(GitLab).Token)
}

// FR3.2: the link refresh forgets a CLI token the provider rejects.
func TestRefreshLinks_ForgetsARejectedCLIToken(t *testing.T) {
	h := newHarness(t)
	h.withCLI(testutil.Token(t))
	_, err := h.svc.UseCLI(h.ctx, ws, GitHub)
	require.NoError(t, err)
	_, err = h.svc.SetMapping(h.ctx, ws, MappingInput{Provider: GitHub, ProjectKey: "PROJ", Repos: []string{"acme/web"}})
	require.NoError(t, err)
	h.clients[GitHub].setPRs("acme/web", pr(GitHub, "acme/web", 1, "x", "b", StateOpen))
	_, err = h.svc.Link(h.ctx, ws, "task-1", "https://github.com/acme/web/pull/1")
	require.NoError(t, err)
	require.True(t, h.cliCached(GitHub))
	h.clients[GitHub].setErr("GetPR", &HTTPError{Provider: GitHub, Status: 401})
	h.svc.RefreshLinks(h.ctx, ws)
	require.False(t, h.cliCached(GitHub))
}

// NFR1: no CLI token in logs, errors or the provider list.
func TestRedaction_NoCLITokenInLogsErrorsOrReplies(t *testing.T) {
	h := newHarness(t)
	cliTok := testutil.Token(t)
	cli := h.withCLI(cliTok)
	var texts []string
	for _, p := range []Provider{GitHub, GitLab} {
		_, err := h.svc.UseCLI(h.ctx, ws, p)
		require.NoError(t, err)
		_, err = h.svc.SetMapping(h.ctx, ws, MappingInput{Provider: p, ProjectKey: "PROJ", Repos: []string{"acme/web"}})
		require.NoError(t, err)
		h.clients[p].setPRs("acme/web", pr(p, "acme/web", 1, "x", "b", StateOpen))
		_, err = h.svc.Link(h.ctx, ws, "task-1", PRURL(PRRef{Provider: p, Repo: "acme/web", Number: 1}))
		require.NoError(t, err)
		h.clients[p].setErr("GetPR", &HTTPError{Provider: p, Status: 401})
	}
	h.svc.RefreshLinks(h.ctx, ws)
	cli.set(cliTok, errors.New("exit status 1: "+cliTok))
	_, err := h.svc.Test(h.ctx, ws, GitHub)
	require.NoError(t, err)
	_, err = h.svc.UseCLI(h.ctx, ws, GitLab)
	require.Error(t, err)
	texts = append(texts, err.Error(), fmt.Sprintf("%+v", err))
	views, err := h.svc.Providers(h.ctx, ws)
	require.NoError(t, err)
	b, _ := json.Marshal(views)
	state, _ := json.Marshal(h.state.data)
	texts = append(texts, string(b), string(state))
	require.NotEmpty(t, h.logs.String())
	testutil.AssertNoLeak(t, h.logs.String()+strings.Join(texts, "\n"), cliTok)
}
