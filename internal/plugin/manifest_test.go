package plugin

import (
	"os"
	"path/filepath"
	"slices"
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
	Capabilities        map[string]any `yaml:"capabilities"`
	RepositoryProviders []string       `yaml:"repository_providers"`
	Actions             []struct {
		Key          string `yaml:"key"`
		Scope        string `yaml:"scope"`
		Access       string `yaml:"access"`
		MaxBodyBytes int    `yaml:"max_body_bytes"`
	} `yaml:"actions"`
	UI struct {
		Bundle string `yaml:"bundle"`
	} `yaml:"ui"`
	Webhooks []struct {
		Key    string `yaml:"key"`
		Method string `yaml:"method"`
		Access string `yaml:"access"`
	} `yaml:"webhooks"`
	ConfigSchema struct {
		Type       string `yaml:"type"`
		Properties map[string]struct {
			Type   string `yaml:"type"`
			Secret bool   `yaml:"secret"`
		} `yaml:"properties"`
		Required []string `yaml:"required"`
	} `yaml:"config_schema"`
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
	require.Len(t, m.Capabilities, 4, "U1 declares state and secrets; U4 adds api_read and api_write")
}

func TestManifestActionsAndAccess(t *testing.T) {
	m := loadManifest(t)
	access := map[string]string{}
	for _, a := range m.Actions {
		require.Regexp(t, `^[a-z0-9][a-z0-9._-]*$`, a.Key, "Kandev's action key rule")
		if slices.Contains(u4Actions, a.Key) {
			continue // TestU4_Manifest_ActionsAndProvider
		}
		require.Equal(t, "workspace", a.Scope, a.Key)
		require.Equal(t, 8192, a.MaxBodyBytes, a.Key)
		access[a.Key] = a.Access
	}
	require.Equal(t, map[string]string{
		actionGet:           "authenticated",
		actionConnectAPIKey: "admin",
		actionSetEnabled:    "admin",
		// U2 (R-01): the three that change the shared connection are admin.
		keyStartOAuth:   "admin",
		keyTest:         "authenticated",
		keyDisconnect:   "admin",
		keyListProjects: "authenticated",
		keySetProjects:  "admin",
	}, access) // NFR3.5, BR2.1, BR7.2
}

func TestU2_ManifestWebhookAndConfigSchema(t *testing.T) {
	m := loadManifest(t)
	require.Len(t, m.Webhooks, 1)
	require.Equal(t, "oauth-callback", m.Webhooks[0].Key)
	require.Equal(t, "GET", m.Webhooks[0].Method)
	require.Equal(t, "public", m.Webhooks[0].Access, "api_version 2 defaults to authenticated, so public is explicit")
	require.Equal(t, "object", m.ConfigSchema.Type)
	require.Len(t, m.ConfigSchema.Properties, 3)
	for _, k := range []string{"oauth_client_id", "oauth_client_secret", "public_base_url"} {
		require.Equal(t, "string", m.ConfigSchema.Properties[k].Type, k)
	}
	require.True(t, m.ConfigSchema.Properties["oauth_client_secret"].Secret)
	require.False(t, m.ConfigSchema.Properties["oauth_client_id"].Secret)
	require.Empty(t, m.ConfigSchema.Required, "OAuth is optional: an API key works without it")
}

func TestU2_ManifestActionKeysMatchTheRuntime(t *testing.T) {
	m := loadManifest(t)
	for _, a := range m.Actions {
		_, ok := handlers[a.Key]
		require.True(t, ok, a.Key)
	}
	require.Len(t, handlers, len(m.Actions))
}

func TestU4_Manifest_ActionsAndProvider(t *testing.T) {
	m := loadManifest(t)
	got := map[string]string{}
	for _, a := range m.Actions {
		if slices.Contains(u4Actions, a.Key) {
			require.Equal(t, 16384, a.MaxBodyBytes, a.Key)
			got[a.Key] = a.Scope + "/" + a.Access
		}
	}
	want := map[string]string{actionSetGitCredential: "workspace/admin"}
	for _, k := range u4Actions[1:] {
		want[k] = "workspace/authenticated"
	}
	for _, k := range []string{actionPRLink, actionPRUnlink, actionPRCreate, actionPRStatus} {
		want[k] = "task/authenticated"
	}
	require.Equal(t, want, got)
	require.Equal(t, []string{"nulab-backlog"}, m.RepositoryProviders)
	require.Equal(t, []any{"tasks", "repositories"}, m.Capabilities["api_read"])
	require.Equal(t, []any{"tasks"}, m.Capabilities["api_write"])
}
