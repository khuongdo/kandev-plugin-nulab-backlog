package ci

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClassifyChanges(t *testing.T) {
	cases := []struct {
		name  string
		paths []string
		app   bool
	}{
		{"records and docs only are non-app", []string{"aidlc/spaces/default/intents/x/aidlc-state.md", "docs/brand/note.md"}, false},
		{"root docs, license, gitignore and .claude are non-app", []string{"README.md", "LICENSE", ".gitignore", ".claude/x"}, false},
		{"docs with app code is app", []string{"docs/brand/note.md", "internal/backlog/client.go"}, true},
		{"a workflow change is app", []string{".github/workflows/ci.yml"}, true},
		{"prefix look-alikes are app", []string{"docsite/a.go"}, true},
		{"a root file named like a directory is app", []string{"aidlc.go"}, true},
		{"a nested README is app", []string{"ui/README.md"}, true},
		{"no changed files is app (fail-safe)", nil, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.app, ClassifyChanges(tc.paths))
		})
	}
}

const (
	baseSHA = "1111111111111111111111111111111111111111"
	headSHA = "2222222222222222222222222222222222222222"
)

// fakeDiff replaces the git runner for one test.
func fakeDiff(t *testing.T, names []string, err error) *[]string {
	t.Helper()
	var got []string
	old := diffNames
	diffNames = func(_ context.Context, base, head string) ([]string, error) {
		got = []string{base, head}
		return names, err
	}
	t.Cleanup(func() { diffNames = old })
	return &got
}

func TestChangesCommand(t *testing.T) {
	t.Run("a records-only diff prints app=false", func(t *testing.T) {
		got := fakeDiff(t, []string{"aidlc/x.md", "docs/y.md"}, nil)
		code, out, errOut := run("changes", "-base", baseSHA, "-head", headSHA)
		require.Equal(t, 0, code)
		require.Equal(t, "app=false\n", out)
		require.Empty(t, errOut)
		require.Equal(t, []string{baseSHA, headSHA}, *got)
	})
	t.Run("an app diff prints app=true", func(t *testing.T) {
		fakeDiff(t, []string{"docs/y.md", "internal/backlog/client.go"}, nil)
		code, out, _ := run("changes", "-base", baseSHA, "-head", headSHA)
		require.Equal(t, 0, code)
		require.Equal(t, "app=true\n", out)
	})
	for name, args := range map[string][]string{
		"an empty base":    {"-head", headSHA},
		"an all-zero base": {"-base", "0000000000000000000000000000000000000000", "-head", headSHA},
		"a non-SHA base":   {"-base", "--output=x", "-head", headSHA},
		"a non-SHA head":   {"-base", baseSHA, "-head", "main"},
	} {
		t.Run(name+" fails safe to app=true without running git", func(t *testing.T) {
			got := fakeDiff(t, []string{"docs/y.md"}, nil)
			code, out, errOut := run(append([]string{"changes"}, args...)...)
			require.Equal(t, 0, code)
			require.Equal(t, "app=true\n", out)
			require.Contains(t, errOut, "warning")
			require.Nil(t, *got, "git must not run")
		})
	}
	t.Run("a git error fails safe to app=true", func(t *testing.T) {
		fakeDiff(t, nil, errors.New("fatal: bad object"))
		code, out, errOut := run("changes", "-base", baseSHA, "-head", headSHA)
		require.Equal(t, 0, code)
		require.Equal(t, "app=true\n", out)
		require.Contains(t, errOut, "bad object")
	})
	t.Run("an empty diff fails safe to app=true", func(t *testing.T) {
		fakeDiff(t, nil, nil)
		code, out, _ := run("changes", "-base", baseSHA, "-head", headSHA)
		require.Equal(t, 0, code)
		require.Equal(t, "app=true\n", out)
	})
	t.Run("a missing -head is a usage error", func(t *testing.T) {
		fakeDiff(t, nil, nil)
		code, out, _ := run("changes", "-base", baseSHA)
		require.Equal(t, 2, code)
		require.Empty(t, out)
	})
}

// The real runner lists the files changed between two commits of this repo.
func TestDiffNamesReadsGit(t *testing.T) {
	dir := t.TempDir()
	git := func(args ...string) string {
		cmd := exec.Command("git", append([]string{"-C", dir, "-c", "user.name=t", "-c", "user.email=t@example.invalid", "-c", "commit.gpgsign=false"}, args...)...) //nolint:gosec // G204: fixed git arguments from this test
		out, err := cmd.CombinedOutput()
		require.NoError(t, err, string(out))
		return strings.TrimSpace(string(out))
	}
	git("init", "-q")
	git("commit", "-q", "--allow-empty", "-m", "a")
	base := git("rev-parse", "HEAD")
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "docs"), 0o750))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "docs", "n.md"), []byte("x\n"), 0o600))
	git("add", ".")
	git("commit", "-q", "-m", "b")
	t.Chdir(dir)

	names, err := diffNames(context.Background(), base, git("rev-parse", "HEAD"))
	require.NoError(t, err)
	require.Equal(t, []string{"docs/n.md"}, names)

	_, err = diffNames(context.Background(), baseSHA, headSHA)
	require.Error(t, err)
}
