package ci

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Rule names reported in a Finding.
const (
	RuleToken       = "credential-shaped-token" //nolint:gosec // G101: a rule name, not a credential
	RulePassword    = "password-literal"
	RuleAPIKeyQuery = "apikey-query-value" //nolint:gosec // G101: a rule name, not a credential
)

// allowedPrefix marks a fake secret in test data (AC7.3.4, stories bait-secret convention).
const allowedPrefix = "TESTSECRET-"

var (
	// alnumRun is a maximal run of 32 or more letters and digits; it is a
	// credential-shaped token when it mixes upper case, lower case and digits.
	alnumRun = regexp.MustCompile(`[A-Za-z0-9]{32,}`)
	upper    = regexp.MustCompile(`[A-Z]`)
	lower    = regexp.MustCompile(`[a-z]`)
	digit    = regexp.MustCompile(`[0-9]`)
	// identifier is letters with digits only at the end, such as the test
	// name TestIntegrationDisabledMapsTo409; generated keys mix digits in.
	identifier = regexp.MustCompile(`^[A-Za-z]+[0-9]*$`)
	// passwordLiteral is a password/passwd key or variable followed by a
	// quoted, non-empty value; group 1 is the value.
	passwordLiteral = regexp.MustCompile(`(?i)\b(?:password|passwd)\b["']?\s*(?::=|[:=])\s*["']([^"'\n]+)["']`)
	// apiKeyQuery is an apiKey= query value; group 1 is the value.
	apiKeyQuery = regexp.MustCompile(`apiKey=([^&\s"'` + "`" + `]+)`)
)

// skippedDirs are never read: VCS data, dependencies and build outputs.
var skippedDirs = map[string]bool{".git": true, "node_modules": true, "dist": true}

// Finding is one credential-shaped string. It never carries the matched
// text, so printing it cannot leak a secret (project.md Mandated).
type Finding struct {
	Path string // slash-separated, relative to the scan root
	Line int
	Rule string
}

func (f Finding) String() string { return fmt.Sprintf("%s:%d: %s", f.Path, f.Line, f.Rule) }

// ScanSecrets reads the test data and test artifacts under root and returns
// every credential-shaped string that does not carry the TESTSECRET- prefix.
func ScanSecrets(root string) ([]Finding, error) {
	var out []Finding
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if skippedDirs[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		rel, err := filepath.Rel(root, p)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if !inScanScope(rel) {
			return nil
		}
		raw, err := os.ReadFile(p) //nolint:gosec // G304: paths come from walking the operator's own tree
		if err != nil {
			return err
		}
		out = append(out, scanLines(rel, raw)...)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("scan %s: %w", root, err)
	}
	return out, nil
}

// inScanScope reports whether rel is a committed test file, test data, a
// manual-check record or a coverage profile.
func inScanScope(rel string) bool {
	switch {
	case rel == "coverage.out", rel == "build/coverage.filtered.out":
		return true
	case strings.HasPrefix(rel, "build/"):
		return false
	case strings.HasSuffix(rel, "_test.go"), strings.Contains("/"+rel, "/testdata/"):
		return true
	case strings.HasPrefix(rel, "docs/manual-checks/"), strings.HasPrefix(rel, "ui/src/testing/"):
		return true
	case strings.HasPrefix(rel, "ui/src/"):
		return strings.HasSuffix(rel, ".test.ts") || strings.HasSuffix(rel, ".test.tsx")
	}
	return false
}

func scanLines(rel string, raw []byte) []Finding {
	var out []Finding
	for i, line := range bytes.Split(raw, []byte("\n")) {
		if rule := lineRule(string(line)); rule != "" {
			out = append(out, Finding{Path: rel, Line: i + 1, Rule: rule})
		}
	}
	return out
}

// lineRule returns the first rule the line breaks, or "".
func lineRule(line string) string {
	for _, loc := range alnumRun.FindAllStringIndex(line, -1) {
		tok := line[loc[0]:loc[1]]
		if upper.MatchString(tok) && lower.MatchString(tok) && digit.MatchString(tok) &&
			!identifier.MatchString(tok) && !strings.HasSuffix(line[:loc[0]], allowedPrefix) {
			return RuleToken
		}
	}
	for _, m := range passwordLiteral.FindAllStringSubmatch(line, -1) {
		if !strings.HasPrefix(m[1], allowedPrefix) {
			return RulePassword
		}
	}
	for _, m := range apiKeyQuery.FindAllStringSubmatch(line, -1) {
		if len(m[1]) >= 8 && !strings.HasPrefix(m[1], allowedPrefix) {
			return RuleAPIKeyQuery
		}
	}
	return ""
}
