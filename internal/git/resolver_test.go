package git

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/connection"
	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/testutil"
)

func scope() Scope {
	return Scope{ProviderID: ProviderID, WorkspaceID: ws, TaskID: "task-17", SessionID: "session-1",
		RepositoryID: "repo-k1", Host: host, Path: "/git/PROJ/web-app.git"}
}

func TestU4_Resolver_ReturnsAFifteenMinuteLease(t *testing.T) {
	r := newRig(t)
	lease, err := r.svc.ResolveCredential(r.ctx, scope())
	require.NoError(t, err)
	require.Equal(t, Lease{Username: "lan", Secret: r.conn.git.Password, ExpiresAt: testNow.Add(15 * time.Minute)}, lease)
}

func TestU4_Resolver_RefusesABadScope(t *testing.T) {
	r := newRig(t)
	for reason, mut := range map[string]func(*Scope){
		"provider":   func(s *Scope) { s.ProviderID = "github" },
		"task":       func(s *Scope) { s.TaskID = "" },
		"session":    func(s *Scope) { s.SessionID = " " },
		"repository": func(s *Scope) { s.RepositoryID = "" },
		"host":       func(s *Scope) { s.Host = "other.backlog.com" },
		"path":       func(s *Scope) { s.Path = "/PROJ/web-app.git" },
		"project":    func(s *Scope) { s.Path = "/git/DEMO/demo-app.git" },
	} {
		sc := scope()
		mut(&sc)
		_, err := r.svc.ResolveCredential(r.ctx, sc)
		var refused *RefusedError
		require.ErrorAs(t, err, &refused, reason)
		require.Equal(t, reason, refused.Reason)
		require.ErrorIs(t, err, ErrScopeRefused)
		testutil.AssertNoLeak(t, err.Error(), r.conn.git.Password, 8)
	}
	sc := scope()
	sc.Host = "EXAMPLE-SPACE.backlog.com"
	_, err := r.svc.ResolveCredential(r.ctx, sc)
	require.NoError(t, err, "the host is compared without case")
}

func TestU4_Resolver_NoGitCredentialSaysWhereToFixIt(t *testing.T) {
	r := newRig(t)
	r.conn.update(func(c *fakeConn) { c.gitErr = connection.ErrNoGitCredential })
	_, err := r.svc.ResolveCredential(r.ctx, scope())
	require.ErrorIs(t, err, connection.ErrNoGitCredential)
	require.Contains(t, err.Error(), "Backlog settings")

	r.conn.update(func(c *fakeConn) { c.err = connection.ErrNotConnected })
	_, err = r.svc.ResolveCredential(r.ctx, scope())
	var refused *RefusedError
	require.ErrorAs(t, err, &refused)
	require.Equal(t, "not_connected", refused.Reason)
}

func TestU4_Resolver_BindingIsEmptyWhenRevoked(t *testing.T) {
	r := newRig(t)
	b, err := r.svc.Binding(r.ctx, scope())
	require.NoError(t, err)
	require.Equal(t, "2.1", b)

	r.conn.update(func(c *fakeConn) { c.gitErr = connection.ErrNoGitCredential })
	b, err = r.svc.Binding(r.ctx, scope())
	require.NoError(t, err)
	require.Empty(t, b)

	r.conn.update(func(c *fakeConn) { c.err = connection.ErrNotConnected })
	b, err = r.svc.Binding(r.ctx, scope())
	require.NoError(t, err)
	require.Empty(t, b, "not connected: empty binding revokes leases")

	r.conn.update(func(c *fakeConn) { c.err = connection.ErrStore })
	_, err = r.svc.Binding(r.ctx, scope())
	require.ErrorIs(t, err, connection.ErrStore)
}
