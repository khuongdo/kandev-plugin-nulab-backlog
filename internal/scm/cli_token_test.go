package scm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/testutil"
)

// Commands the fake CLI accepts (FR3.3): anything else, `gh auth switch`
// included, fails the test.
var allowedCLI = regexp.MustCompile(`^(gh auth status --json hosts --hostname github\.com` +
	`|gh auth token --hostname github\.com( --user \S+)?` +
	`|glab config get token --host gitlab\.com)$`)

const (
	ghStatus = "gh auth status --json hosts --hostname github.com"
	ghPlain  = "gh auth token --hostname github.com"
)

func ghUser(login string) string { return ghPlain + " --user " + login }

type cliReply struct {
	out string
	err error
}

// fakeCLI stands in for gh / glab: it records each call and returns the
// reply scripted for that command, or the default one. No test runs a real
// CLI (NFR3).
type fakeCLI struct {
	t       testing.TB
	mu      sync.Mutex
	replies map[string]cliReply // by "name arg arg"; "" is the default
	calls   []string
	limits  []int
}

func (f *fakeCLI) run(_ context.Context, limit int, name string, args ...string) ([]byte, error) {
	cmd := strings.Join(append([]string{name}, args...), " ")
	if !allowedCLI.MatchString(cmd) {
		f.t.Errorf("unexpected CLI command %q", cmd)
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, cmd)
	f.limits = append(f.limits, limit)
	r, ok := f.replies[cmd]
	if !ok {
		r = f.replies[""]
	}
	return []byte(r.out), r.err
}

// set changes the default reply.
func (f *fakeCLI) set(out string, err error) { f.on("", out, err) }

// on scripts the reply of one command.
func (f *fakeCLI) on(cmd, out string, err error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.replies[cmd] = cliReply{out, err}
}

func (f *fakeCLI) count() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.calls)
}

func (f *fakeCLI) log() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string(nil), f.calls...)
}

// withCLI puts a fake CLI into the harness: every token read prints tok and
// gh lists the fake client's user (lan-id) as its only, active account.
func (h *harness) withCLI(tok string) *fakeCLI {
	f := &fakeCLI{t: h.t, replies: map[string]cliReply{"": {out: tok + "\n"}}}
	f.on(ghStatus, statusJSON("lan-id"), nil)
	h.svc.CLI = f.run
	return f
}

// statusJSON is `gh auth status --json hosts` output; the first login is active.
func statusJSON(logins ...string) string {
	var entries []string
	for i, l := range logins {
		entries = append(entries, fmt.Sprintf(`{"active":%t,"host":"github.com","login":%q,"state":"success"}`, i == 0, l))
	}
	return `{"hosts":{"github.com":[` + strings.Join(entries, ",") + `]}}`
}

func readTestdata(t *testing.T, name string) string {
	t.Helper()
	b, err := os.ReadFile("testdata/" + name) //nolint:gosec // G304: fixed fixture names from this file
	require.NoError(t, err)
	return string(b)
}

// AC1.1.5: the GitHub login rule.
func TestValidLogin(t *testing.T) {
	for login, want := range map[string]bool{
		"alice": true, "Bob-2": true, "a": true, "a1-b2-c3": true, strings.Repeat("a", 39): true,
		"": false, " ": false, " alice": false, "-x": false, "x-": false, "a--b": false, "a;b": false,
		"a b": false, "a_b": false, "alice\n": false, strings.Repeat("a", 40): false, "ålice": false,
	} {
		require.Equal(t, want, validLogin(login), "%q", login)
	}
}

// FR1.2, FR2.1, NFR2: fixed commands per provider; stdout is trimmed; gh reads the chosen login.
func TestCLIToken_RunsTheFixedCommandAndTrims(t *testing.T) {
	h := newHarness(t)
	cli := h.withCLI("  gho_fake_cli_token \n")
	tok, err := h.svc.cliToken(h.ctx, GitHub, "bob")
	require.NoError(t, err)
	require.Equal(t, "gho_fake_cli_token", tok)
	tok, err = h.svc.cliToken(h.ctx, GitLab, "")
	require.NoError(t, err)
	require.Equal(t, "gho_fake_cli_token", tok)
	require.Equal(t, []string{ghUser("bob"), "glab config get token --host gitlab.com"}, cli.log())
	require.Equal(t, []int{maxCLIOutput, maxCLIOutput}, cli.limits)
}

// FR4.1, NFR1, AC3.1.3: every CLI failure is ErrCLIUnavailable without stderr or token text.
func TestCLIToken_EveryFailureIsCLIUnavailable(t *testing.T) {
	secret := testutil.Token(t)
	for name, c := range map[string]struct {
		out string
		err error
	}{
		"empty output":      {"\n", nil},
		"space inside":      {secret + " extra", nil},
		"control character": {secret + "\x00", nil},
		"too long":          {strings.Repeat("x", maxToken+1), nil},
		"non-zero exit":     {secret, fmt.Errorf("exit status 1: not logged in to github.com %s", secret)},
		"missing binary":    {"", fmt.Errorf("gh: %w", exec.ErrNotFound)},
		"timeout":           {"", context.DeadlineExceeded},
		"output cut off":    {secret, errCLIOutputCut},
	} {
		t.Run(name, func(t *testing.T) {
			for _, p := range []Provider{GitHub, GitLab} {
				h := newHarness(t)
				cli := h.withCLI("")
				cli.set(c.out, c.err)
				cli.on(ghStatus, c.out, c.err)
				tok, err := h.svc.cliToken(h.ctx, p, "alice")
				require.ErrorIs(t, err, ErrCLIUnavailable)
				require.Empty(t, tok)
				require.NotContains(t, err.Error(), secret)
				require.NotContains(t, err.Error(), "not logged in to github.com")
				testutil.AssertNoLeak(t, h.logs.String(), secret)
			}
		})
	}
}

// The caller's cancellation is returned unchanged.
func TestCLIToken_CallerCancellationIsReturned(t *testing.T) {
	h := newHarness(t)
	h.withCLI("").set("", context.Canceled)
	ctx, cancel := context.WithCancel(h.ctx)
	cancel()
	_, err := h.svc.cliToken(ctx, GitHub, "alice")
	require.ErrorIs(t, err, context.Canceled)
	_, err = h.svc.ListCLIAccounts(ctx)
	require.ErrorIs(t, err, context.Canceled)
}

// FR2.3: Bitbucket has no CLI login.
func TestCLIToken_BitbucketIsRefused(t *testing.T) {
	h := newHarness(t)
	cli := h.withCLI("x")
	_, err := h.svc.cliToken(h.ctx, Bitbucket, "")
	require.Equal(t, FieldProvider, fieldOf(t, err))
	require.Zero(t, cli.count())
}

// AC3.2.4, AC1.1.5: a GitHub read without a valid login is account-missing and runs nothing.
func TestCLIToken_GitHubWithoutAValidLoginRunsNothing(t *testing.T) {
	for _, login := range []string{"", "-x", "a;b", " ", strings.Repeat("a", 40)} {
		h := newHarness(t)
		cli := h.withCLI("x")
		_, err := h.svc.cliToken(h.ctx, GitHub, login)
		require.ErrorIs(t, err, ErrCLIAccountMissing, "%q", login)
		require.Zero(t, cli.count(), "%q", login)
	}
}

// FR3.2, NFR4, AC1.2.2, AC2.1.7: one CLI run per provider and login per 5
// minutes; one login's token is never served to another; forget clears
// every login of one provider.
func TestCLIToken_CachesPerLoginForFiveMinutes(t *testing.T) {
	h := newHarness(t)
	cli := h.withCLI("x")
	cli.on(ghUser("alice"), "tok-alice", nil)
	cli.on(ghUser("bob"), "tok-bob", nil)
	for range 3 {
		tok, err := h.svc.cliToken(h.ctx, GitHub, "alice")
		require.NoError(t, err)
		require.Equal(t, "tok-alice", tok)
	}
	require.Equal(t, 1, cli.count())
	tok, err := h.svc.cliToken(h.ctx, GitHub, "bob")
	require.NoError(t, err)
	require.Equal(t, "tok-bob", tok, "alice's cache entry is not served to bob")
	require.Equal(t, 2, cli.count())

	cli.on(ghUser("bob"), "tok-bob-2", nil)
	h.now = h.now.Add(cliTTL - time.Second)
	tok, _ = h.svc.cliToken(h.ctx, GitHub, "bob")
	require.Equal(t, "tok-bob", tok, "still cached")
	h.now = h.now.Add(time.Second)
	tok, _ = h.svc.cliToken(h.ctx, GitHub, "bob")
	require.Equal(t, "tok-bob-2", tok, "asked again after expiry")
	require.Equal(t, 3, cli.count())

	_, _ = h.svc.cliToken(h.ctx, GitLab, "")
	h.svc.forgetCLI(GitHub)
	_, _ = h.svc.cliToken(h.ctx, GitLab, "")
	require.Equal(t, 4, cli.count(), "GitLab stays cached")
	_, _ = h.svc.cliToken(h.ctx, GitHub, "alice")
	_, _ = h.svc.cliToken(h.ctx, GitHub, "bob")
	require.Equal(t, 6, cli.count(), "every GitHub login was forgotten")
}

// A failure is not cached: the next call asks the CLI again.
func TestCLIToken_FailureIsNotCached(t *testing.T) {
	h := newHarness(t)
	cli := h.withCLI("")
	cli.set("", errors.New("exit status 1"))
	cli.on(ghStatus, "", errors.New("exit status 1"))
	_, err := h.svc.cliToken(h.ctx, GitHub, "lan-id")
	require.ErrorIs(t, err, ErrCLIUnavailable)
	cli.set("now-logged-in", nil)
	tok, err := h.svc.cliToken(h.ctx, GitHub, "lan-id")
	require.NoError(t, err)
	require.Equal(t, "now-logged-in", tok)
}

// FR3.4, FR5.1, AC2.1.4, AC2.1.6, AC3.1.5: when the --user read fails, the
// account list decides: no list is cli_unavailable, a missing login is
// cli_account_missing, a listed login is read with the plain command and
// used only when GitHub says the token is that login.
func TestCLIToken_UserReadFailureIsClassified(t *testing.T) {
	failUser := errors.New("exit status 1: unknown flag: --user")
	for name, c := range map[string]struct {
		status  cliReply
		plainAs string // the /user login of the plain token
		want    error
		calls   []string
	}{
		"list fails": {cliReply{"", errors.New("exit status 1")}, "", ErrCLIUnavailable,
			[]string{ghUser("bob"), ghStatus, ghPlain}},
		"login not listed": {cliReply{statusJSON("alice"), nil}, "", ErrCLIAccountMissing,
			[]string{ghUser("bob"), ghStatus}},
		"listed, another account active": {cliReply{statusJSON("alice", "bob"), nil}, "alice", ErrCLIAccountMissing,
			[]string{ghUser("bob"), ghStatus, ghPlain}},
		"listed and active": {cliReply{statusJSON("bob", "alice"), nil}, "BOB", nil,
			[]string{ghUser("bob"), ghStatus, ghPlain}},
	} {
		t.Run(name, func(t *testing.T) {
			h := newHarness(t)
			cli := h.withCLI("plain-token")
			cli.on(ghUser("bob"), "", failUser)
			cli.on(ghStatus, c.status.out, c.status.err)
			h.clients[GitHub].user = User{ID: c.plainAs, Name: "Someone"}
			tok, err := h.svc.cliToken(h.ctx, GitHub, "bob")
			require.Equal(t, c.calls, cli.log())
			if c.want != nil {
				require.ErrorIs(t, err, c.want)
				require.Empty(t, tok)
				require.False(t, h.cliCached(GitHub))
				return
			}
			require.NoError(t, err)
			require.Equal(t, "plain-token", tok)
			require.Equal(t, "plain-token", h.lastCred(GitHub).Token)
		})
	}
}

// FR1.1, AC1.1.1, AC1.1.6, AC1.1.8: the accounts are github.com's usable
// logins with the active one marked, and nothing else.
func TestListCLIAccounts_ParsesGHStatus(t *testing.T) {
	for name, c := range map[string]struct {
		out  string
		err  error
		want []CLIAccount
	}{
		"two accounts": {readTestdata(t, "gh-auth-status-two.json"), nil,
			[]CLIAccount{{Login: "alice", Active: true}, {Login: "bob"}}},
		"error state and invalid logins are dropped": {readTestdata(t, "gh-auth-status-error-state.json"), nil,
			[]CLIAccount{{Login: "alice", Active: true}}},
		"non-zero exit with decodable stdout": {statusJSON("bob", "alice"), errors.New("exit status 1"),
			[]CLIAccount{{Login: "bob", Active: true}, {Login: "alice"}}},
	} {
		t.Run(name, func(t *testing.T) {
			h := newHarness(t)
			cli := h.withCLI("x")
			cli.on(ghStatus, c.out, c.err)
			got, err := h.svc.ListCLIAccounts(h.ctx)
			require.NoError(t, err)
			require.Equal(t, c.want, got)
			require.Equal(t, []string{ghStatus}, cli.log())
			require.Equal(t, []int{maxCLIListOutput}, cli.limits)
			b, err := json.Marshal(got[0])
			require.NoError(t, err)
			require.JSONEq(t, fmt.Sprintf(`{"login":%q,"active":true}`, got[0].Login), string(b), "login and active only")
		})
	}
}

// AC1.1.3, AC1.1.8, AC3.1.3: a cut-off list, an empty usable list or a dead
// gh is cli_unavailable, never a partial list or gh's output.
func TestListCLIAccounts_FailuresAreCLIUnavailable(t *testing.T) {
	secret := testutil.Token(t)
	two := readTestdata(t, "gh-auth-status-two.json")
	for name, c := range map[string]struct {
		status, plain cliReply
	}{
		"cut off":            {cliReply{two, errCLIOutputCut}, cliReply{"x", nil}},
		"no usable account":  {cliReply{`{"hosts":{"github.com":[{"login":"carol","state":"error","active":true}]}}`, nil}, cliReply{"x", nil}},
		"no github.com host": {cliReply{`{"hosts":{}}`, nil}, cliReply{"x", nil}},
		"not installed":      {cliReply{"", exec.ErrNotFound}, cliReply{"", exec.ErrNotFound}},
		"timeout":            {cliReply{"", context.DeadlineExceeded}, cliReply{"", context.DeadlineExceeded}},
		"not logged in":      {cliReply{"", errors.New("exit status 1: " + secret)}, cliReply{"", errors.New("exit status 1: " + secret)}},
	} {
		t.Run(name, func(t *testing.T) {
			h := newHarness(t)
			cli := h.withCLI("")
			cli.on(ghStatus, c.status.out, c.status.err)
			cli.on(ghPlain, c.plain.out, c.plain.err)
			got, err := h.svc.ListCLIAccounts(h.ctx)
			require.ErrorIs(t, err, ErrCLIUnavailable)
			require.Nil(t, got)
			require.NotContains(t, err.Error(), secret)
		})
	}
}

// AC1.1.9: gh without `auth status --json` offers only the active account,
// found with the plain token read and GitHub's /user.
func TestListCLIAccounts_OldGHOffersTheActiveAccountOnly(t *testing.T) {
	h := newHarness(t)
	cli := h.withCLI("plain-token")
	cli.on(ghStatus, "", errors.New("exit status 1: unknown flag: --json"))
	h.clients[GitHub].user = User{ID: "alice", Name: "Alice"}
	got, err := h.svc.ListCLIAccounts(h.ctx)
	require.NoError(t, err)
	require.Equal(t, []CLIAccount{{Login: "alice", Active: true}}, got)
	require.Equal(t, []string{ghStatus, ghPlain}, cli.log())
	require.Equal(t, "plain-token", h.lastCred(GitHub).Token)

	h.clients[GitHub].user = User{ID: "not a login"}
	_, err = h.svc.ListCLIAccounts(h.ctx)
	require.ErrorIs(t, err, ErrCLIUnavailable, "a /user login that breaks the rule is not offered")
	h.clients[GitHub].setErr("CurrentUser", &HTTPError{Provider: GitHub, Status: 401})
	_, err = h.svc.ListCLIAccounts(h.ctx)
	require.ErrorIs(t, err, ErrCLIUnavailable)
}

// cliHelperEnv is the helper process mode of TestRunCLI_*.
const cliHelperEnv = "SCM_CLI_HELPER"

// NFR2, AC1.1.8: the default runner reads at most limit bytes of stdout,
// reports a cut-off and a missing binary.
func TestRunCLI_LimitsOutputAndReportsMissingBinary(t *testing.T) {
	if os.Getenv(cliHelperEnv) == "big" {
		fmt.Print(strings.Repeat("x", 3*maxCLIOutput))
		fmt.Fprint(os.Stderr, "stderr is discarded")
		os.Exit(0)
	}
	t.Setenv(cliHelperEnv, "big")
	self := []string{"-test.run=^TestRunCLI_LimitsOutputAndReportsMissingBinary$"}
	out, err := runCLI(context.Background(), maxCLIOutput, os.Args[0], self...)
	require.ErrorIs(t, err, errCLIOutputCut)
	require.Equal(t, strings.Repeat("x", maxCLIOutput), string(out))
	out, err = runCLI(context.Background(), maxCLIListOutput, os.Args[0], self...)
	require.NoError(t, err)
	require.Equal(t, strings.Repeat("x", 3*maxCLIOutput), string(out))

	_, err = runCLI(context.Background(), maxCLIOutput, "kandev-no-such-cli-binary")
	require.ErrorIs(t, err, exec.ErrNotFound)
}

// AC1.1.6, AC2.1.8: gh never sees the server's token variables, so it reads
// its stored logins instead of an ambient token.
func TestRunCLI_StripsTokenVariables(t *testing.T) {
	if os.Getenv(cliHelperEnv) == "env" {
		for _, k := range []string{"GH_TOKEN", "GITHUB_TOKEN", "GH_ENTERPRISE_TOKEN", "GITHUB_ENTERPRISE_TOKEN", "KEEP_ME"} {
			if v, ok := os.LookupEnv(k); ok {
				fmt.Printf("%s=%s\n", k, v)
			}
		}
		os.Exit(0)
	}
	t.Setenv(cliHelperEnv, "env")
	secret := testutil.Token(t)
	for _, k := range []string{"GH_TOKEN", "GITHUB_TOKEN", "GH_ENTERPRISE_TOKEN", "GITHUB_ENTERPRISE_TOKEN"} {
		t.Setenv(k, secret)
	}
	t.Setenv("KEEP_ME", "yes")
	out, err := runCLI(context.Background(), maxCLIOutput, os.Args[0], "-test.run=^TestRunCLI_StripsTokenVariables$")
	require.NoError(t, err)
	require.Equal(t, "KEEP_ME=yes\n", string(out))
	require.Equal(t, []string{"A=1", "B=GH_TOKEN"}, cliEnv([]string{"A=1", "GH_TOKEN=x", "gh_token=x", "B=GH_TOKEN"}),
		"names match without case, values never")
}
