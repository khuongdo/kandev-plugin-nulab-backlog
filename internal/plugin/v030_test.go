package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kandev/kandev/pkg/pluginsdk"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/backlog"
	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/git"
)

func v030(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", "v030", name)) //nolint:gosec // fixture names are test constants
	require.NoError(t, err)
	return b
}

func v030Doc(t *testing.T, name string) map[string]any {
	t.Helper()
	var m map[string]any
	require.NoError(t, json.Unmarshal(v030(t, name), &m))
	return m
}

// FR7.2: every v0.3.0 action keeps its key, scope, access and body limit.
func TestV030_ActionKeysAreUnchanged(t *testing.T) {
	var old manifest
	require.NoError(t, yaml.Unmarshal(v030(t, "manifest.yaml"), &old))
	require.Len(t, old.Actions, 55)
	now := map[string]string{}
	for _, a := range loadManifest(t).Actions {
		now[a.Key] = fmt.Sprintf("%s/%s/%d", a.Scope, a.Access, a.MaxBodyBytes)
	}
	for _, a := range old.Actions {
		require.Equal(t, fmt.Sprintf("%s/%s/%d", a.Scope, a.Access, a.MaxBodyBytes), now[a.Key], a.Key)
		require.False(t, strings.HasPrefix(a.Key, "scm."), "new behaviour uses new keys only")
	}
}

// FR1.3, FR7.1: data stored by v0.3.0 is read as Backlog Git data and keeps working.
func TestV030_StoredGitDataStillWorks(t *testing.T) {
	r := newSCMRig(t)
	r.host.mu.Lock()
	for _, key := range []string{"git.links", "git.queries", "git.watches", "git.ledger"} {
		r.host.state["workspace/ws-1/"+key] = v030Doc(t, key+".json")
	}
	r.host.secrets["backlog.git.ws-1"] = strings.TrimSpace(string(v030(t, "backlog.git.ws-1.json")))
	r.host.mu.Unlock()
	r.data.mu.Lock()
	r.data.tasks = append(r.data.tasks, pluginsdk.Task{ID: "task-7", Metadata: v030Doc(t, "task-metadata.json")})
	r.data.mu.Unlock()
	r.gw4.prs = []backlog.PullRequest{{RepositoryID: 11, Number: 42, Summary: "Add login page", StatusID: 3}}

	_, out := r.call(t, actionLinksList, nil)
	require.Equal(t, []any{map[string]any{"providerId": "nulab-backlog", "taskId": "task-17", "reviewKey": spaceHost + "|11|42",
		"connectionScope": spaceHost, "repositoryId": "11", "changeRequestNumber": float64(42)}}, out["associations"])
	_, out = r.call(t, actionQueriesList, nil)
	q := out["queries"].([]any)[0].(map[string]any)
	require.Equal(t, []any{"Merged by me", true}, []any{q["name"], q["isDefault"]})
	_, out = r.call(t, actionWatchesList, nil)
	w := out["watches"].([]any)[0].(map[string]any)
	require.Equal(t, []any{"Reviews", "active", float64(2)}, []any{w["name"], w["state"], w["createdCount"]})

	resp, out := r.taskCall(t, actionPRStatus, nil)
	require.Equal(t, 200, resp.Status)
	s := out["summaries"].([]any)[0].(map[string]any)
	require.Equal(t, "Merged", s["statusBadge"].(map[string]any)["label"], "the review key format is unchanged")

	cred, err := r.rt.ResolveGitCredential(context.Background(), &pluginsdk.ResolveGitCredentialRequest{
		ProviderID: git.ProviderID, WorkspaceID: "ws-1", TaskID: "task-17", SessionID: "s-1", RepositoryID: "repo-k1",
		Host: spaceHost, Path: "/git/PROJ/web-app.git"})
	require.NoError(t, err)
	require.Equal(t, "lan", cred.Username, "the backlog.git.<ws> secret still leases")

	id, found, err := hostPort{host: r.rt.Host}.FindTaskByMetadata(context.Background(), "ws-1", git.MetadataKey, spaceHost+"|11|7")
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, "task-7", id, "nulab_backlog_pr metadata is still found")

	resp, _ = r.taskCall(t, actionPRUnlink, map[string]string{"reviewKey": spaceHost + "|11|42"})
	require.Equal(t, 200, resp.Status)
	_, out = r.call(t, actionSCMProviders, nil)
	for _, p := range out["providers"].([]any) {
		require.Equal(t, "not_configured", p.(map[string]any)["state"], "v0.3.0 data creates no provider settings")
	}
}
