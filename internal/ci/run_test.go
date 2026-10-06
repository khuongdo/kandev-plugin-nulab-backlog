package ci

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func run(args ...string) (int, string, string) {
	var out, errOut bytes.Buffer
	code := Run(args, &out, &errOut)
	return code, out.String(), errOut.String()
}

func TestRunRejectsUnknownSubcommandAndMissingFlags(t *testing.T) {
	for _, args := range [][]string{
		{}, {"bogus"}, {"secrets"}, {"preflight", "-version", "0.1.0"}, {"marketplace"}, {"workflows"},
		{"contract", "-base-url", "http://127.0.0.1:1"}, {"secrets", "-no-such-flag"},
	} {
		code, _, _ := run(args...)
		require.Equal(t, 2, code, "%v", args)
	}
}

func TestRunSecrets(t *testing.T) {
	code, out, _ := run("secrets", "-root", writeTree(t, map[string]string{"a_test.go": "x := 1\n"}))
	require.Equal(t, 0, code)
	require.Contains(t, out, "OK")

	code, out, errOut := run("secrets", "-root", writeTree(t, map[string]string{"a_test.go": "\"" + baitToken + "\"\n"}))
	require.Equal(t, 1, code)
	require.Contains(t, errOut, "a_test.go:1: "+RuleToken)
	require.NotContains(t, out+errOut, baitToken[:8])
}

func TestRunPreflight(t *testing.T) {
	dir := passingRecordDir(t, "pass")
	base := []string{"preflight", "-tag", "v0.1.0", "-version", "0.1.0", "-manual-checks", dir}
	code, out, _ := run(append(base, "-on-main=true", "-releases", "[]")...)
	require.Equal(t, 0, code)
	require.Contains(t, out, "OK")

	code, _, errOut := run(append(base, "-on-main=false", "-releases", "[]")...)
	require.Equal(t, 1, code)
	require.Contains(t, errOut, "not on main")

	code, _, errOut = run(append(base, "-on-main=true", "-releases", `[{"tagName":"v0.1.0","isDraft":false,"isPrerelease":false}]`)...)
	require.Equal(t, 1, code)
	require.Contains(t, errOut, "already has a GitHub Release")
}

// R-03: when gh release list fails the Makefile passes no list; the preflight
// fails closed instead of reading that as "no Release".
func TestRunPreflightRefusesWhenTheReleaseListIsUnknown(t *testing.T) {
	base := []string{"preflight", "-tag", "v0.1.0", "-version", "0.1.0", "-on-main=true", "-manual-checks", passingRecordDir(t, "pass")}
	for _, extra := range [][]string{nil, {"-releases", ""}, {"-releases", "error connecting to api.github.com"}} {
		code, out, errOut := run(append(base, extra...)...)
		require.Equal(t, 1, code, "%v", extra)
		require.Empty(t, out)
		require.Contains(t, errOut, "could not list GitHub Releases")
		require.Equal(t, 1, strings.Count(errOut, "\n"), "one-line reason")
	}
}

func TestRunMarketplace(t *testing.T) {
	manifest := writeManifest(t, "id: nulab-backlog\nversion: 0.1.0\nmin_kandev_version: 0.96.0\n")
	base := []string{"marketplace", "-manifest", manifest, "-repo", thisRepo, "-categories", "integrations"}
	code, out, _ := run(base...)
	require.Equal(t, 0, code)
	require.Contains(t, out, "  - id: nulab-backlog\n")

	reg := writeTree(t, map[string]string{"plugins.yaml": "plugins:\n  - id: backlog\n    repo: " + thisRepo + "\n"})
	code, _, errOut := run(append(base, "-registry", reg+"/plugins.yaml")...)
	require.Equal(t, 1, code)
	require.Contains(t, errOut, "differs from the manifest id")
}

func TestRunWorkflowsAndContract(t *testing.T) {
	code, _, _ := run("workflows", "-dir", writeTree(t, map[string]string{"ci.yml": compliantCI, "release.yml": compliantRelease}))
	require.Equal(t, 0, code)
	code, _, errOut := run("workflows", "-dir", writeTree(t, map[string]string{"ci.yml": "on: pull_request_target\njobs: {}\n"}))
	require.Equal(t, 1, code)
	require.Contains(t, errOut, "pull_request_target")

	cfg, _ := startFake(t, newFakeKandev())
	code, out, errOut := run("contract", "-base-url", cfg.BaseURL, "-package", cfg.PackagePath, "-plugin-id", "nulab-backlog", "-host-version", "v0.96.0")
	require.Equal(t, 0, code, errOut)
	require.Contains(t, out, "OK")
}
