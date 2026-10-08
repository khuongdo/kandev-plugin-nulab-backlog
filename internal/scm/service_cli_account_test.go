package scm

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/testutil"
)

// withGHAccounts gives gh two github.com logins, alice (active) and bob;
// GitHub's /user names each token's login. It returns the tokens by login.
func (h *harness) withGHAccounts(t *testing.T) (*fakeCLI, map[string]string) {
	t.Helper()
	cli := h.withCLI("unused")
	toks := map[string]string{"alice": testutil.Token(t), "bob": testutil.Token(t)}
	cli.on(ghStatus, statusJSON("alice", "bob"), nil)
	cli.on(ghPlain, toks["alice"], nil)
	f := h.clients[GitHub]
	f.users = map[string]User{}
	for login, tok := range toks {
		cli.on(ghUser(login), tok+"\n", nil)
		f.users[tok] = User{ID: login, Name: strings.ToUpper(login[:1]) + login[1:]}
	}
	return cli, toks
}

// seenBy returns the fake GitHub calls since the last reset.
func (h *harness) seenBy() []string {
	f := h.clients[GitHub]
	f.mu.Lock()
	defer f.mu.Unlock()
	out := f.seen
	f.seen = nil
	return out
}

// FR2.1, FR2.3, AC1.1.1, AC1.1.2, AC1.2.1, AC1.2.2: no login means gh's
// active account; a chosen login is saved as AccountID; changing it keeps
// the mappings and the next call reads the new login.
func TestUseCLI_ConnectsTheChosenGHAccount(t *testing.T) {
	h := newHarness(t)
	cli, toks := h.withGHAccounts(t)
	v, err := h.svc.UseCLI(h.ctx, ws, GitHub, "")
	require.NoError(t, err)
	require.Equal(t, "alice", v.Login)
	require.Equal(t, "Alice", v.Account)
	require.Equal(t, "alice", h.settingsFor(t, GitHub).AccountID)
	_, err = h.svc.SetMapping(h.ctx, ws, MappingInput{Provider: GitHub, ProjectKey: "PROJ", Repos: []string{"acme/web"}})
	require.NoError(t, err)

	v, err = h.svc.UseCLI(h.ctx, ws, GitHub, "bob")
	require.NoError(t, err)
	require.Equal(t, ProviderView{Provider: GitHub, State: StateConnected, Method: MethodCLI, Account: "Bob", Login: "bob",
		Mappings: []Mapping{{ProjectKey: "PROJ", Repos: []string{"acme/web"}}}}, v)
	require.Equal(t, "bob", h.settingsFor(t, GitHub).AccountID)
	_, err = h.svc.SearchRepos(h.ctx, ws, GitHub, "")
	require.NoError(t, err)
	require.Equal(t, toks["bob"], h.lastCred(GitHub).Token)
	require.Contains(t, cli.log(), ghUser("bob"))
	for _, c := range cli.log() {
		require.NotContains(t, c, "switch", "gh's active account is never changed")
	}
}

// FR5.1, AC1.1.7, AC1.2.3: a login gh does not have, or a token GitHub says
// is someone else's, is account-missing and changes nothing.
func TestUseCLI_AMissingAccountChangesNothing(t *testing.T) {
	h := newHarness(t)
	cli, toks := h.withGHAccounts(t)
	_, err := h.svc.UseCLI(h.ctx, ws, GitHub, "alice")
	require.NoError(t, err)
	before := h.settingsFor(t, GitHub)

	cli.on(ghUser("carol"), "", errors.New("exit status 1: no account found for carol"))
	_, err = h.svc.UseCLI(h.ctx, ws, GitHub, "carol")
	require.ErrorIs(t, err, ErrCLIAccountMissing)
	require.Equal(t, before, h.settingsFor(t, GitHub))

	cli.on(ghUser("bob"), toks["alice"], nil) // gh hands out the wrong account's token
	_, err = h.svc.UseCLI(h.ctx, ws, GitHub, "bob")
	require.ErrorIs(t, err, ErrCLIAccountMissing)
	require.Equal(t, before, h.settingsFor(t, GitHub))
	require.False(t, h.cliCached(GitHub), "the wrong token is not kept")
}

// AC1.1.5, AC3.2.3: an invalid login, or any login for GitLab, is a field
// error before any CLI run, and nothing is saved.
func TestUseCLI_InvalidLoginIsAFieldError(t *testing.T) {
	for _, c := range []struct {
		p     Provider
		login string
	}{
		{GitHub, "-x"}, {GitHub, "a;b"}, {GitHub, " "}, {GitHub, " alice"}, {GitHub, strings.Repeat("a", 40)},
		{GitLab, "alice"},
	} {
		t.Run(string(c.p)+" "+c.login, func(t *testing.T) {
			h := newHarness(t)
			cli := h.withCLI("x")
			_, err := h.svc.UseCLI(h.ctx, ws, c.p, c.login)
			require.Equal(t, FieldLogin, fieldOf(t, err))
			require.Zero(t, cli.count())
			require.Equal(t, Settings{Provider: c.p}, h.settingsFor(t, c.p))
		})
	}
}

// FR3.1, AC1.2.2, AC2.1.2, AC2.1.3, AC2.1.5: with bob chosen and alice
// active, every GitHub call site of the workspace reads bob's token.
func TestCredential_EveryCallSiteUsesTheChosenAccount(t *testing.T) {
	for name, run := range map[string]func(h *harness) error{
		"test": func(h *harness) error { _, err := h.svc.Test(h.ctx, ws, GitHub); return err },
		"repo list": func(h *harness) error {
			_, err := h.svc.SearchRepos(h.ctx, ws, GitHub, "")
			return err
		},
		"pr list": func(h *harness) error {
			_, err := h.svc.ListPRs(h.ctx, ws, PRListInput{Provider: GitHub, ProjectKey: "PROJ", Repo: "acme/web",
				Statuses: []string{StateOpen}, Author: WhoAnyone})
			return err
		},
		"mine filter": func(h *harness) error {
			page, err := h.svc.ListPRs(h.ctx, ws, PRListInput{Provider: GitHub, ProjectKey: "PROJ", Repo: "acme/web",
				Statuses: []string{StateOpen}, Author: WhoMe})
			if err == nil && (len(page.Items) != 1 || page.Items[0].Number != 1) {
				err = fmt.Errorf("mine is not bob's PR: %+v", page.Items)
			}
			return err
		},
		"pr link": func(h *harness) error {
			_, err := h.svc.Link(h.ctx, ws, "task-1", "https://github.com/acme/web/pull/2")
			return err
		},
		"watch poll": func(h *harness) error {
			_, err := h.svc.SaveWatch(h.ctx, ws, watchInput(GitHub))
			if err == nil {
				h.seenBy()
				NewWatcher(h.svc, nil).cycle(h.ctx)
				if len(h.tasks.all()) != 1 {
					err = errors.New("the watch created no task")
				}
			}
			return err
		},
	} {
		t.Run(name, func(t *testing.T) {
			h := newHarness(t)
			h.withGHAccounts(t)
			_, err := h.svc.UseCLI(h.ctx, ws, GitHub, "bob")
			require.NoError(t, err)
			_, err = h.svc.SetMapping(h.ctx, ws, MappingInput{Provider: GitHub, ProjectKey: "PROJ", Repos: []string{"acme/web"}})
			require.NoError(t, err)
			bobs, alices := pr(GitHub, "acme/web", 1, "mine", "a", StateOpen), pr(GitHub, "acme/web", 2, "theirs", "b", StateOpen)
			bobs.AuthorID, alices.AuthorID = "bob", "alice"
			h.clients[GitHub].setPRs("acme/web", bobs, alices)
			h.seenBy()

			require.NoError(t, run(h))
			seen := h.seenBy()
			require.NotEmpty(t, seen)
			for _, s := range seen {
				require.True(t, strings.HasSuffix(s, " bob"), s)
			}
		})
	}
}

// FR2.4, AC2.1.1: two workspaces use two accounts at the same time.
func TestCredential_TwoWorkspacesTwoAccountsConcurrently(t *testing.T) {
	h := newHarness(t)
	h.withGHAccounts(t)
	chosen := map[string]string{"ws-1": "alice", "ws-2": "bob"}
	for w, login := range chosen {
		_, err := h.svc.UseCLI(h.ctx, w, GitHub, login)
		require.NoError(t, err)
	}
	h.seenBy()
	var wg sync.WaitGroup
	for range 10 {
		for w := range chosen {
			wg.Go(func() {
				if _, err := h.svc.SearchRepos(inWS(h.ctx, w), w, GitHub, ""); err != nil {
					t.Error(err)
				}
			})
		}
	}
	wg.Wait()
	seen := h.seenBy()
	require.Len(t, seen, 20)
	for _, s := range seen {
		w, _, _ := strings.Cut(s, " ")
		require.True(t, strings.HasSuffix(s, " "+chosen[w]), s)
	}
}

// FR5.1, FR5.2, AC3.1.1, AC3.1.2: Test keeps the chosen login after gh
// switches account, and records cli_account_missing once it is logged out.
func TestTest_KeepsTheChosenLogin(t *testing.T) {
	h := newHarness(t)
	cli, _ := h.withGHAccounts(t)
	_, err := h.svc.UseCLI(h.ctx, ws, GitHub, "bob")
	require.NoError(t, err)

	cli.on(ghStatus, statusJSON("alice", "bob"), nil) // `gh auth switch` to alice
	h.clients[GitHub].users[h.ghToken(t, "bob")] = User{ID: "bob", Name: "Bob Renamed"}
	v, err := h.svc.Test(h.ctx, ws, GitHub)
	require.NoError(t, err)
	require.Equal(t, StateConnected, v.State)
	require.Equal(t, "Bob Renamed", v.Account, "the display name of the same login is refreshed")
	require.Equal(t, "bob", h.settingsFor(t, GitHub).AccountID)

	cli.on(ghUser("bob"), "", errors.New("exit status 1"))
	cli.on(ghStatus, statusJSON("alice"), nil) // `gh auth logout --user bob`
	v, err = h.svc.Test(h.ctx, ws, GitHub)
	require.NoError(t, err, "a failed test is a result")
	require.Equal(t, StateError, v.State)
	require.Equal(t, ErrorCLIAccountMissing, v.LastError)
	require.Equal(t, "bob", v.Login)
	require.Equal(t, "bob", h.settingsFor(t, GitHub).AccountID)
	require.Equal(t, ErrorCLIAccountMissing, errorCode(fmt.Errorf("x: %w", ErrCLIAccountMissing)))
}

// ghToken is the token the fake gh prints for login.
func (h *harness) ghToken(t *testing.T, login string) string {
	t.Helper()
	for tok, u := range h.clients[GitHub].users {
		if u.ID == login {
			return tok
		}
	}
	t.Fatalf("no token for %s", login)
	return ""
}

// FR4.1, AC3.2.1, AC3.2.4: a v0.5.2 gh CLI record (frozen fixture) keeps
// working as its AccountID while gh has another account active; a record
// without AccountID is account-missing and never uses the active account.
func TestUpgrade_V052CLISettingsKeepTheirAccount(t *testing.T) {
	h := newHarness(t)
	cli, toks := h.withGHAccounts(t)
	cli.on(ghStatus, statusJSON("bob", "alice"), nil)
	cli.on(ghPlain, toks["bob"], nil)
	raw, err := os.ReadFile("testdata/v052-settings-cli.json")
	require.NoError(t, err)
	var doc map[string]any
	require.NoError(t, json.Unmarshal(raw, &doc))
	h.state.set("workspace/"+ws+"/"+keySettings, doc)

	_, err = h.svc.SearchRepos(h.ctx, ws, GitHub, "")
	require.NoError(t, err)
	require.Equal(t, toks["alice"], h.lastCred(GitHub).Token)
	views, err := h.svc.Providers(h.ctx, ws)
	require.NoError(t, err)
	require.Equal(t, ProviderView{Provider: GitHub, State: StateConnected, Method: MethodCLI, Account: "Alice", Login: "alice",
		Mappings: []Mapping{{ProjectKey: "PROJ", Repos: []string{"acme/web"}}}}, views[0])

	doc["items"].([]any)[0].(map[string]any)["accountId"] = ""
	h.state.set("workspace/"+ws+"/"+keySettings, doc)
	h.svc.forgetCLI(GitHub)
	runs := cli.count()
	_, err = h.svc.SearchRepos(h.ctx, ws, GitHub, "")
	require.ErrorIs(t, err, ErrCLIAccountMissing)
	v, err := h.svc.Test(h.ctx, ws, GitHub)
	require.NoError(t, err)
	require.Equal(t, ErrorCLIAccountMissing, v.LastError)
	require.Equal(t, runs, cli.count(), "no gh run without a login")
}

// NFR1, AC3.1.3: no gh failure puts a token or gh's output into an error,
// a log, the provider list or the stored state.
func TestRedaction_NoGHOutputInErrorsLogsOrViews(t *testing.T) {
	sentinel := testutil.Token(t)
	raw := "gh: token " + sentinel + " is invalid"
	for name, c := range map[string]struct{ user, status cliReply }{
		"not installed":    {cliReply{"", errors.New("exec: gh: not found")}, cliReply{"", errors.New("exec: gh: not found")}},
		"not logged in":    {cliReply{"", errors.New("exit status 1: " + raw)}, cliReply{`{"hosts":{}}`, nil}},
		"undecodable":      {cliReply{raw, errors.New("exit status 1")}, cliReply{raw, errors.New("exit status 1: " + raw)}},
		"timeout":          {cliReply{"", errors.New("signal: killed")}, cliReply{"", errors.New("signal: killed")}},
		"token read fails": {cliReply{raw, errors.New("exit status 1: " + raw)}, cliReply{statusJSON("alice"), nil}},
		"over the cap":     {cliReply{"", errors.New("exit status 1")}, cliReply{raw, errCLIOutputCut}},
	} {
		t.Run(name, func(t *testing.T) {
			h := newHarness(t)
			cli, _ := h.withGHAccounts(t)
			_, err := h.svc.UseCLI(h.ctx, ws, GitHub, "bob")
			require.NoError(t, err)
			h.svc.forgetCLI(GitHub)
			cli.on(ghUser("bob"), c.user.out, c.user.err)
			cli.on(ghUser("alice"), c.user.out, c.user.err)
			cli.on(ghStatus, c.status.out, c.status.err)
			cli.on(ghPlain, c.user.out, c.user.err)

			var texts []string
			_, err = h.svc.SearchRepos(h.ctx, ws, GitHub, "")
			require.Error(t, err)
			texts = append(texts, err.Error(), fmt.Sprintf("%+v", err))
			_, err = h.svc.UseCLI(h.ctx, ws, GitHub, "alice")
			require.Error(t, err)
			texts = append(texts, err.Error())
			_, err = h.svc.ListCLIAccounts(h.ctx)
			if err != nil {
				texts = append(texts, err.Error())
			}
			v, err := h.svc.Test(h.ctx, ws, GitHub)
			require.NoError(t, err)
			require.Contains(t, []string{ErrorCLIUnavailable, ErrorCLIAccountMissing}, v.LastError)
			views, _ := h.svc.Providers(h.ctx, ws)
			b, _ := json.Marshal(views)
			state, _ := json.Marshal(h.state.data)
			all := h.logs.String() + strings.Join(texts, "\n") + string(b) + string(state)
			testutil.AssertNoLeak(t, all, sentinel)
			require.NotContains(t, all, "is invalid")
		})
	}
}
