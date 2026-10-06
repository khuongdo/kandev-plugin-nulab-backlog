package plugin

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

type manifest struct {
	ID               string `yaml:"id"`
	APIVersion       int    `yaml:"api_version"`
	Version          string `yaml:"version"`
	MinKandevVersion string `yaml:"min_kandev_version"`
	Runtime          struct {
		Type        string            `yaml:"type"`
		Executables map[string]string `yaml:"executables"`
	} `yaml:"runtime"`
	Capabilities map[string]any `yaml:"capabilities"`
	Actions      []struct {
		Key          string `yaml:"key"`
		Scope        string `yaml:"scope"`
		Access       string `yaml:"access"`
		MaxBodyBytes int    `yaml:"max_body_bytes"`
	} `yaml:"actions"`
	UI struct {
		Bundle string `yaml:"bundle"`
	} `yaml:"ui"`
}

func loadManifest(t *testing.T) manifest {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("..", "..", "manifest.yaml"))
	require.NoError(t, err)
	var m manifest
	require.NoError(t, yaml.Unmarshal(raw, &m))
	return m
}

func TestManifestIdentity(t *testing.T) {
	m := loadManifest(t)
	require.Equal(t, "nulab-backlog", m.ID)
	require.Equal(t, 2, m.APIVersion)
	require.NotEmpty(t, m.Version)
	require.Equal(t, "/ui/bundle.js", m.UI.Bundle)
}

func TestManifestDeclaresTheMinimumKandevVersion(t *testing.T) {
	require.Equal(t, "0.96.0", loadManifest(t).MinKandevVersion) // NFR6.1
}

func TestManifestListsExactlyTheFiveExecutables(t *testing.T) {
	m := loadManifest(t)
	require.Equal(t, "binary", m.Runtime.Type)
	require.Equal(t, map[string]string{
		"linux-amd64":   "server/plugin-linux-amd64",
		"linux-arm64":   "server/plugin-linux-arm64",
		"darwin-amd64":  "server/plugin-darwin-amd64",
		"darwin-arm64":  "server/plugin-darwin-arm64",
		"windows-amd64": "server/plugin-windows-amd64.exe",
	}, m.Runtime.Executables) // BR5.1
}

func TestManifestCapabilities(t *testing.T) {
	m := loadManifest(t)
	require.Equal(t, true, m.Capabilities["state"])
	require.Equal(t, true, m.Capabilities["secrets"])
	require.Len(t, m.Capabilities, 2, "U1 declares only state and secrets")
}

func TestManifestActionsAndAccess(t *testing.T) {
	m := loadManifest(t)
	access := map[string]string{}
	for _, a := range m.Actions {
		require.Regexp(t, `^[a-z0-9][a-z0-9._-]*$`, a.Key, "Kandev's action key rule")
		require.Equal(t, "workspace", a.Scope, a.Key)
		require.Equal(t, 8192, a.MaxBodyBytes, a.Key)
		access[a.Key] = a.Access
	}
	require.Equal(t, map[string]string{
		actionGet:           "authenticated",
		actionConnectAPIKey: "admin",
		actionSetEnabled:    "admin",
	}, access) // NFR3.5, BR2.1, BR7.2
}
