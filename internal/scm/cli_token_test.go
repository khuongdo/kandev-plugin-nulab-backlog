package scm

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/testutil"
)

// fakeCLI stands in for gh / glab: it records each call and returns the
// scripted stdout or error. No test runs a real CLI (NFR3).
type fakeCLI struct {
	mu    sync.Mutex
	out   string
	err   error
	calls []string // "name arg arg"
}

func (f *fakeCLI) run(_ context.Context, name string, args ...string) ([]byte, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, strings.Join(append([]string{name}, args...), " "))
	return []byte(f.out), f.err
}

func (f *fakeCLI) set(out string, err error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.out, f.err = out, err
}

func (f *fakeCLI) count() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.calls)
}

// withCLI puts a fake CLI that prints tok into the harness.
func (h *harness) withCLI(tok string) *fakeCLI {
	f := &fakeCLI{out: tok + "\n"}
	h.svc.CLI = f.run
	return f
}

// FR1.2, FR2.1, NFR2: fixed commands per provider; stdout is trimmed.
func TestCLIToken_RunsTheFixedCommandAndTrims(t *testing.T) {
	h := newHarness(t)
	cli := h.withCLI("  gho_fake_cli_token \n")
	tok, err := h.svc.cliToken(h.ctx, GitHub)
	require.NoError(t, err)
	require.Equal(t, "gho_fake_cli_token", tok)
	tok, err = h.svc.cliToken(h.ctx, GitLab)
	require.NoError(t, err)
	require.Equal(t, "gho_fake_cli_token", tok)
	require.Equal(t, []string{"gh auth token --hostname github.com", "glab config get token --host gitlab.com"}, cli.calls)
}

// FR4.1, NFR1: every CLI failure is ErrCLIUnavailable without stderr or token text.
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
	} {
		t.Run(name, func(t *testing.T) {
			h := newHarness(t)
			cli := h.withCLI("")
			cli.set(c.out, c.err)
			tok, err := h.svc.cliToken(h.ctx, GitHub)
			require.ErrorIs(t, err, ErrCLIUnavailable)
			require.Empty(t, tok)
			require.NotContains(t, err.Error(), secret)
			require.NotContains(t, err.Error(), "not logged in to github.com")
		})
	}
}

// The caller's cancellation is returned unchanged.
func TestCLIToken_CallerCancellationIsReturned(t *testing.T) {
	h := newHarness(t)
	h.withCLI("").set("", context.Canceled)
	ctx, cancel := context.WithCancel(h.ctx)
	cancel()
	_, err := h.svc.cliToken(ctx, GitHub)
	require.ErrorIs(t, err, context.Canceled)
}

// FR2.3: Bitbucket has no CLI login.
func TestCLIToken_BitbucketIsRefused(t *testing.T) {
	h := newHarness(t)
	cli := h.withCLI("x")
	_, err := h.svc.cliToken(h.ctx, Bitbucket)
	require.Equal(t, FieldProvider, fieldOf(t, err))
	require.Zero(t, cli.count())
}

// FR3.2, FR3.3, NFR4: one CLI run per provider per 5 minutes; forget clears one provider.
func TestCLIToken_CachesForFiveMinutes(t *testing.T) {
	h := newHarness(t)
	cli := h.withCLI("first")
	for range 3 {
		tok, err := h.svc.cliToken(h.ctx, GitHub)
		require.NoError(t, err)
		require.Equal(t, "first", tok)
	}
	require.Equal(t, 1, cli.count())

	cli.set("second", nil)
	h.now = h.now.Add(cliTTL - time.Second)
	tok, _ := h.svc.cliToken(h.ctx, GitHub)
	require.Equal(t, "first", tok, "still cached")
	h.now = h.now.Add(time.Second)
	tok, _ = h.svc.cliToken(h.ctx, GitHub)
	require.Equal(t, "second", tok, "asked again after expiry")
	require.Equal(t, 2, cli.count())

	_, _ = h.svc.cliToken(h.ctx, GitLab)
	h.svc.forgetCLI(GitHub)
	_, _ = h.svc.cliToken(h.ctx, GitLab)
	require.Equal(t, 3, cli.count(), "GitLab stays cached")
	_, _ = h.svc.cliToken(h.ctx, GitHub)
	require.Equal(t, 4, cli.count(), "GitHub was forgotten")
}

// A failure is not cached: the next call asks the CLI again.
func TestCLIToken_FailureIsNotCached(t *testing.T) {
	h := newHarness(t)
	cli := h.withCLI("")
	cli.set("", errors.New("exit status 1"))
	_, err := h.svc.cliToken(h.ctx, GitHub)
	require.ErrorIs(t, err, ErrCLIUnavailable)
	cli.set("now-logged-in", nil)
	tok, err := h.svc.cliToken(h.ctx, GitHub)
	require.NoError(t, err)
	require.Equal(t, "now-logged-in", tok)
}

// NFR2: the default runner reads at most 4 KiB of stdout and reports a missing binary.
func TestRunCLI_LimitsOutputAndReportsMissingBinary(t *testing.T) {
	if os.Getenv("SCM_CLI_HELPER") == "1" {
		fmt.Print(strings.Repeat("x", 3*maxCLIOutput))
		fmt.Fprint(os.Stderr, "stderr is discarded")
		os.Exit(0)
	}
	t.Setenv("SCM_CLI_HELPER", "1")
	out, err := runCLI(context.Background(), os.Args[0], "-test.run=^TestRunCLI_LimitsOutputAndReportsMissingBinary$")
	require.NoError(t, err)
	require.Equal(t, strings.Repeat("x", maxCLIOutput), string(out))

	_, err = runCLI(context.Background(), "kandev-no-such-cli-binary")
	require.ErrorIs(t, err, exec.ErrNotFound)
}
