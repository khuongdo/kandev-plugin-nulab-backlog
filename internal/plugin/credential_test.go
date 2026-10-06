package plugin

import (
	"context"
	"testing"
	"time"

	"github.com/kandev/kandev/pkg/pluginsdk"
	"github.com/stretchr/testify/require"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/testutil"
)

var _ pluginsdk.GitCredentialHandler = (*Runtime)(nil)

func credRequest() *pluginsdk.ResolveGitCredentialRequest {
	return &pluginsdk.ResolveGitCredentialRequest{ProviderID: "nulab-backlog", WorkspaceID: "ws-1", TaskID: "task-17",
		SessionID: "session-1", RepositoryID: "repo-k1", Host: spaceHost, Path: "/git/PROJ/web-app.git"}
}

func (r *u4rig) gitPassword(t *testing.T) string {
	t.Helper()
	pw := testutil.Token(t)
	resp, _ := r.call(t, actionSetGitCredential, map[string]string{"gitUsername": "lan", "gitPassword": pw})
	require.Equal(t, 200, resp.Status)
	return pw
}

func TestU4_Credential_ResolvesAValidScope(t *testing.T) {
	r := newU4Rig(t)
	pw := r.gitPassword(t)
	resp, err := r.rt.ResolveGitCredential(context.Background(), credRequest())
	require.NoError(t, err)
	require.Equal(t, "lan", resp.Username)
	require.Equal(t, pw, resp.Secret)
	exp, err := time.Parse(time.RFC3339, resp.ExpiresAt)
	require.NoError(t, err)
	require.WithinDuration(t, time.Now().Add(15*time.Minute), exp, time.Minute)
	testutil.AssertNoLeak(t, r.logs.String(), pw, 8)
}

func TestU4_Credential_RefusalsHoldNoSecret(t *testing.T) {
	r := newU4Rig(t)
	pw := r.gitPassword(t)
	for reason, mut := range map[string]func(*pluginsdk.ResolveGitCredentialRequest){
		"provider": func(q *pluginsdk.ResolveGitCredentialRequest) { q.ProviderID = "github" },
		"task":     func(q *pluginsdk.ResolveGitCredentialRequest) { q.TaskID = "" },
		"host":     func(q *pluginsdk.ResolveGitCredentialRequest) { q.Host = "evil.example.com" },
		"project":  func(q *pluginsdk.ResolveGitCredentialRequest) { q.Path = "/git/DEMO/x.git" },
	} {
		q := credRequest()
		mut(q)
		resp, err := r.rt.ResolveGitCredential(context.Background(), q)
		require.Error(t, err, reason)
		require.Nil(t, resp)
		testutil.AssertNoLeak(t, err.Error(), pw, 8)
		require.Contains(t, r.logs.String(), `"reason":"`+reason+`"`)
	}
	require.Contains(t, r.logs.String(), `"event":"git_credential_refused"`)
	testutil.AssertNoLeak(t, r.logs.String(), pw, 8)

	r2 := newU4Rig(t)
	_, err := r2.rt.ResolveGitCredential(context.Background(), credRequest())
	require.ErrorContains(t, err, "Backlog settings", "no Git access: the error says where to fix it")
	require.Contains(t, r2.logs.String(), `"reason":"no_git_credential"`)
}

func TestU4_Credential_BindingFollowsTheConnection(t *testing.T) {
	r := newU4Rig(t)
	r.gitPassword(t)
	req := &pluginsdk.GitCredentialBindingRequest{ProviderID: "nulab-backlog", WorkspaceID: "ws-1", TaskID: "task-17",
		SessionID: "session-1", RepositoryID: "repo-k1", Host: spaceHost, Path: "/git/PROJ/web-app.git"}
	resp, err := r.rt.GetGitCredentialBinding(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, "2.1", resp.Binding)
	r.call(t, keyDisconnect, nil)
	resp, err = r.rt.GetGitCredentialBinding(context.Background(), req)
	require.NoError(t, err)
	require.Empty(t, resp.Binding, "revoked after a disconnect")
}

func TestU4_Credential_StoreFailureIsAnError(t *testing.T) {
	r := newU4Rig(t)
	r.host.failRead = true
	_, err := r.rt.GetGitCredentialBinding(context.Background(), &pluginsdk.GitCredentialBindingRequest{ProviderID: "nulab-backlog", WorkspaceID: "ws-1"})
	require.Error(t, err)
	_, err = r.rt.ResolveGitCredential(context.Background(), credRequest())
	require.Error(t, err)
}

func TestU4_Credential_RefusedWhileBacklogIsOff(t *testing.T) {
	r := newU4Rig(t)
	r.gitPassword(t)
	r.call(t, actionSetEnabled, map[string]any{"enabled": false})
	_, err := r.rt.ResolveGitCredential(context.Background(), credRequest())
	require.ErrorContains(t, err, "integration_disabled")
	resp, err := r.rt.GetGitCredentialBinding(context.Background(), &pluginsdk.GitCredentialBindingRequest{ProviderID: "nulab-backlog", WorkspaceID: "ws-1"})
	require.NoError(t, err)
	require.Empty(t, resp.Binding, "off revokes leases (BR7.3)")
}
