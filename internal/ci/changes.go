package ci

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"slices"
	"strings"
)

// nonAppDirs and nonAppFiles are the only paths whose changes skip the app
// CI jobs (FR1.1). Every other path is app (FR1.2). This is the one place
// the split is defined (FR1.3).
var (
	nonAppDirs  = []string{"aidlc/", ".claude/", "docs/"}
	nonAppFiles = map[string]bool{"README.md": true, "LICENSE": true, ".gitignore": true}
)

// ClassifyChanges reports whether the changed paths (slash-separated,
// relative to the repository root) need the app CI jobs. An empty list is
// app, so an unknown change set never skips the checks (FR2.4).
func ClassifyChanges(paths []string) bool {
	return len(paths) == 0 || slices.ContainsFunc(paths, isApp)
}

func isApp(path string) bool {
	return !nonAppFiles[path] && !slices.ContainsFunc(nonAppDirs, func(d string) bool { return strings.HasPrefix(path, d) })
}

// commitSHA is a full commit SHA. Anything else (an empty or all-zero push
// "before", a branch name, an option) never reaches git.
var commitSHA = regexp.MustCompile(`^[0-9a-f]{40}$`)

// diffNames lists the paths changed between the merge base of base and head,
// and head. Tests replace it.
var diffNames = func(ctx context.Context, base, head string) ([]string, error) {
	out, err := exec.CommandContext(ctx, "git", "diff", "--name-only", base+"..."+head).Output() //nolint:gosec // G204: base and head are checked full SHAs
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return nil, fmt.Errorf("git diff: %w: %s", err, strings.TrimSpace(string(exitErr.Stderr)))
		}
		return nil, fmt.Errorf("git diff: %w", err)
	}
	// One path per line; a path git quotes (unusual characters) never matches
	// a non-app prefix, so it counts as app.
	return strings.FieldsFunc(string(out), func(r rune) bool { return r == '\n' }), nil
}

// changesCommand prints app=true or app=false for $GITHUB_OUTPUT. Any doubt
// about the change set prints app=true with a warning and exits 0, so CI
// never skips the app jobs by mistake (FR2.4).
func changesCommand(fs *flag.FlagSet) func(io.Writer, io.Writer) int {
	base := fs.String("base", "", "base commit SHA (pull request base, or the push's previous commit)")
	head := fs.String("head", "", "head commit SHA")
	return func(stdout, stderr io.Writer) int {
		if *head == "" {
			return usage(stderr, "changes", "-head")
		}
		names, err := changedPaths(*base, *head)
		if err != nil {
			_, _ = fmt.Fprintf(stderr, "ci changes: warning: %v; running the app jobs\n", err)
		}
		_, _ = fmt.Fprintf(stdout, "app=%t\n", err != nil || ClassifyChanges(names))
		return 0
	}
}

func changedPaths(base, head string) ([]string, error) {
	if !commitSHA.MatchString(base) || strings.Trim(base, "0") == "" || !commitSHA.MatchString(head) {
		return nil, fmt.Errorf("base %q or head %q is not a usable commit SHA", base, head)
	}
	return diffNames(context.Background(), base, head)
}
