package ci

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// Every bait string is assembled at run time, so this file never matches the
// scanner itself (AC7.3.4).
var (
	baitToken    = strings.Repeat("aB3", 12) // 36 characters: upper, lower and digit
	testPrefix   = strings.Join([]string{"TEST", "SECRET-"}, "")
	passwordWord = strings.Join([]string{"pass", "word"}, "")
	apiKeyParam  = strings.Join([]string{"api", "Key="}, "")
)

func writeTree(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for name, body := range files {
		p := filepath.Join(root, filepath.FromSlash(name))
		require.NoError(t, os.MkdirAll(filepath.Dir(p), 0o750))
		require.NoError(t, os.WriteFile(p, []byte(body), 0o600))
	}
	return root
}

func TestScanSecretsFindsCredentialShapedTokenInTestdataWithoutEchoingIt(t *testing.T) {
	root := writeTree(t, map[string]string{"internal/x/testdata/a.json": "{\n  \"key\": \"" + baitToken + "\"\n}\n"})
	got, err := ScanSecrets(root)
	require.NoError(t, err)
	require.Equal(t, []Finding{{Path: "internal/x/testdata/a.json", Line: 2, Rule: RuleToken}}, got)
	require.NotContains(t, got[0].String(), baitToken)
	require.NotContains(t, got[0].String(), baitToken[:12])
}

func TestScanSecretsAllowsTheTestSecretPrefix(t *testing.T) {
	root := writeTree(t, map[string]string{
		"internal/x/testdata/a.json": "\"" + testPrefix + baitToken + "\"\n",
		"internal/x/a_test.go":       passwordWord + " := \"" + testPrefix + "hunter\"\nurl := \"/x?" + apiKeyParam + testPrefix + "abcdefgh\"\n",
	})
	got, err := ScanSecrets(root)
	require.NoError(t, err)
	require.Empty(t, got)
}

func TestScanSecretsFindsQuotedPasswordLiteral(t *testing.T) {
	root := writeTree(t, map[string]string{
		"internal/x/a_test.go":       "x := 1\n" + passwordWord + " := \"hunter2\"\n",
		"internal/x/testdata/b.json": "{\"passwd\": \"letmein\"}\n",
	})
	got, err := ScanSecrets(root)
	require.NoError(t, err)
	require.ElementsMatch(t, []Finding{
		{Path: "internal/x/a_test.go", Line: 2, Rule: RulePassword},
		{Path: "internal/x/testdata/b.json", Line: 1, Rule: RulePassword},
	}, got)
}

func TestScanSecretsFindsLongAPIKeyQueryValue(t *testing.T) {
	root := writeTree(t, map[string]string{"docs/manual-checks/2026-10-06-first-release.md": "see https://a.backlog.com/api/v2/users/myself?" + apiKeyParam + "abcd1234efgh\n"})
	got, err := ScanSecrets(root)
	require.NoError(t, err)
	require.Equal(t, []Finding{{Path: "docs/manual-checks/2026-10-06-first-release.md", Line: 1, Rule: RuleAPIKeyQuery}}, got)
}

func TestScanSecretsIgnoresKnownSafeShapes(t *testing.T) {
	safe := []string{
		strings.Repeat("0123456789abcdef", 4), // SHA-256 digest
		strings.Repeat("f099a46dc7", 4),       // commit SHA
		"func TestScanSecretsIgnoresKnownSafeShapesOfManyKindsInTestNames(t *testing.T) {}",
		"func TestIntegrationDisabledMapsToStatusCode409(t *testing.T) {}", // digits only at the end
		"\"k-123456\"",
		"\"test-api-key-0000\"",
		apiKeyParam + "abc123",
		passwordWord + " := \"\"",
	}
	for _, line := range safe {
		t.Run(line, func(t *testing.T) {
			root := writeTree(t, map[string]string{"internal/x/a_test.go": line + "\n"})
			got, err := ScanSecrets(root)
			require.NoError(t, err)
			require.Empty(t, got)
		})
	}
}

func TestScanSecretsReadsOnlyTheScanScope(t *testing.T) {
	bait := "\"" + baitToken + "\"\n"
	inScope := []string{
		"internal/x/testdata/a.json", "internal/x/a_test.go", "ui/src/a/b.test.ts", "ui/src/a/b.test.tsx",
		"ui/src/testing/fake.ts", "docs/manual-checks/r.md", "coverage.out", "build/coverage.filtered.out",
	}
	outOfScope := []string{
		"go.sum", "ui/package-lock.json", "ui/node_modules/p/testdata/a.json", ".git/testdata/a",
		"dist/checksums.txt", "internal/x/a.go", "ui/src/a/b.ts", "build/other.out", "docs/other.md",
	}
	files := map[string]string{}
	for _, p := range append(append([]string{}, inScope...), outOfScope...) {
		files[p] = bait
	}
	got, err := ScanSecrets(writeTree(t, files))
	require.NoError(t, err)
	var paths []string
	for _, f := range got {
		paths = append(paths, f.Path)
	}
	require.ElementsMatch(t, inScope, paths)
}

func TestScanSecretsFailsOnMissingRoot(t *testing.T) {
	_, err := ScanSecrets(filepath.Join(t.TempDir(), "absent"))
	require.Error(t, err)
}

func TestScanSecretsFindsNothingInTheRepository(t *testing.T) {
	got, err := ScanSecrets("../..")
	require.NoError(t, err)
	require.Empty(t, got, "credential-shaped strings in test data or test artifacts")
}
