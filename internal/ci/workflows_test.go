package ci

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

const pinned = "actions/checkout@3d3c42e5aac5ba805825da76410c181273ba90b1"

const compliantCI = `name: ci
on:
  pull_request:
    branches: [main]
permissions:
  contents: read
jobs:
  checks:
    runs-on: ubuntu-latest
    steps:
      - uses: ` + pinned + ` # v7.0.1
      - run: make test
`

const compliantRelease = `name: release
on:
  push:
    tags: ['v*']
permissions:
  contents: read
jobs:
  verify:
    runs-on: ubuntu-latest
    steps:
      - uses: ` + pinned + `
  publish:
    needs: [verify]
    runs-on: ubuntu-latest
    permissions:
      contents: write
      id-token: write
      attestations: write
    steps:
      - uses: actions/attest-build-provenance@0123456789abcdef0123456789abcdef01234567
`

func TestCheckWorkflowsPassesACompliantPair(t *testing.T) {
	dir := writeTree(t, map[string]string{"ci.yml": compliantCI, "release.yml": compliantRelease})
	require.NoError(t, CheckWorkflows(dir))
}

func TestCheckWorkflowsFlagsViolations(t *testing.T) {
	cases := []struct {
		name, file, body, want string
	}{
		{"tag pin", "ci.yml", strings.Replace(compliantCI, pinned, "actions/checkout@v4", 1), "not pinned"},
		{"short sha", "ci.yml", strings.Replace(compliantCI, pinned, "actions/checkout@3d3c42e", 1), "not pinned"},
		{"missing permissions", "ci.yml", strings.Replace(compliantCI, "permissions:\n  contents: read\n", "", 1), "permissions"},
		{"broader permissions", "ci.yml", strings.Replace(compliantCI, "contents: read", "contents: write", 1), "permissions"},
		{"extra permission", "ci.yml", strings.Replace(compliantCI, "  contents: read\n", "  contents: read\n  packages: read\n", 1), "permissions"},
		{"write-all", "ci.yml", strings.Replace(compliantCI, "permissions:\n  contents: read\n", "permissions: write-all\n", 1), "permissions"},
		{"pull_request_target", "ci.yml", strings.Replace(compliantCI, "pull_request:", "pull_request_target:", 1), "pull_request_target"},
		{"pull_request_target in a list", "ci.yml", strings.Replace(compliantCI, "on:\n  pull_request:\n    branches: [main]\n", "on: [push, pull_request_target]\n", 1), "pull_request_target"},
		{"id-token outside publish", "release.yml", strings.Replace(compliantRelease, "  verify:\n    runs-on: ubuntu-latest\n", "  verify:\n    runs-on: ubuntu-latest\n    permissions:\n      id-token: write\n", 1), "verify"},
		{"attestations in ci.yml", "ci.yml", strings.Replace(compliantCI, "    runs-on: ubuntu-latest\n", "    runs-on: ubuntu-latest\n    permissions:\n      attestations: write\n", 1), "checks"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			files := map[string]string{"ci.yml": compliantCI, "release.yml": compliantRelease}
			files[tc.file] = tc.body
			err := CheckWorkflows(writeTree(t, files))
			require.ErrorContains(t, err, tc.want)
			require.ErrorContains(t, err, tc.file)
		})
	}
}

func TestCheckWorkflowsFailsOnUnreadableInput(t *testing.T) {
	require.Error(t, CheckWorkflows(writeTree(t, map[string]string{"ci.yml": "jobs: [unclosed\n"})))
	require.Error(t, CheckWorkflows(writeTree(t, map[string]string{"notes.txt": "x"})), "a directory with no workflow is an error")
}

func TestRepositoryWorkflowsFollowThePolicy(t *testing.T) {
	require.NoError(t, CheckWorkflows("../../.github/workflows"))

	raw, err := os.ReadFile("../../.github/workflows/release.yml")
	require.NoError(t, err, "release.yml must exist")
	var wf struct {
		Jobs map[string]struct {
			Permissions map[string]string `yaml:"permissions"`
		} `yaml:"jobs"`
	}
	require.NoError(t, yaml.Unmarshal(raw, &wf))
	publish, ok := wf.Jobs["publish"]
	require.True(t, ok, "release.yml needs a publish job")
	require.Equal(t, "write", publish.Permissions["id-token"])
	require.Equal(t, "write", publish.Permissions["attestations"])
}
