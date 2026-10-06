package connection

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/backlog"
	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/testutil"
)

// The base fakeGateway has no Git repositories; gitGateway scripts them.
func (g *fakeGateway) Repositories(context.Context, backlog.Credentials, string) ([]backlog.Repository, error) {
	return nil, nil
}

func (g *fakeGateway) CheckGitAccess(context.Context, string, string, string, string, string) error {
	return nil
}

type gitGateway struct {
	*fakeGateway
	mu        sync.Mutex
	repos     []backlog.Repository
	reposErr  error
	gitErr    error
	repoCalls []string // project keys
	gitCalls  []string // "host user project/repo"
}

func (g *gitGateway) Repositories(_ context.Context, _ backlog.Credentials, projectKey string) ([]backlog.Repository, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.repoCalls = append(g.repoCalls, projectKey)
	return g.repos, g.reposErr
}

func (g *gitGateway) CheckGitAccess(_ context.Context, host, user, _, project, repo string) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.gitCalls = append(g.gitCalls, fmt.Sprintf("%s %s %s/%s", host, user, project, repo))
	return g.gitErr
}

// newGit is a U2 harness with a scripted Git gateway, connected to
// a.backlog.com with project PROJ selected.
func newGit(t *testing.T) (*u2, *gitGateway) {
	t.Helper()
	u := newU2(t)
	gg := &gitGateway{fakeGateway: u.gw, repos: []backlog.Repository{{ID: 11, ProjectID: 101, Name: "web-app"}}}
	u.svc.gateway = gg
	u.connectKey(t, "a.backlog.com")
	u.gw.projects = []backlog.Project{{ID: 101, Key: "PROJ"}, {ID: 102, Key: "DEMO"}}
	_, err := u.svc.SetProjects(u.ctx, ws, []any{"PROJ"})
	require.NoError(t, err)
	return u, gg
}

func (u *u2) setGit(t *testing.T) string {
	t.Helper()
	pw := testutil.Token(t)
	_, err := u.svc.SetGitCredential(u.ctx, ws, GitCredentialInput{Username: "lan", Password: pw})
	require.NoError(t, err)
	return pw
}

func TestU4_GitAccess_SaveReturnsTheViewWithoutThePassword(t *testing.T) {
	u, _ := newGit(t)
	pw := testutil.Token(t)
	view, err := u.svc.SetGitCredential(u.ctx, ws, GitCredentialInput{Username: "lan", Password: pw})
	require.NoError(t, err)
	require.True(t, view.HasGitCredential)
	require.True(t, view.Connected)
	b, _ := json.Marshal(view)
	testutil.AssertNoLeak(t, string(b), pw, 8)
	require.NotContains(t, string(b), "lan", "the view has no user name field either")
	cred, _, err := u.svc.GitCredential(u.ctx, ws)
	require.NoError(t, err)
	require.Equal(t, GitCredential{Username: "lan", Password: pw, SpaceHost: "a.backlog.com"}, cred)
}

func TestU4_GitAccess_SaveRefusals(t *testing.T) {
	u, _ := newGit(t)
	_, err := u.svc.SetGitCredential(u.ctx, ws, GitCredentialInput{Username: "", Password: testutil.Token(t)})
	require.Equal(t, Outcome{Code: CodeValidation, Field: FieldGitUsername}, Classify(err))
	_, err = u.svc.SetGitCredential(u.ctx, ws, GitCredentialInput{Username: "lan"})
	require.Equal(t, Outcome{Code: CodeValidation, Field: FieldGitPassword}, Classify(err))
	_, err = u.svc.SetGitCredential(u.ctx, "ws-none", GitCredentialInput{Username: "lan", Password: testutil.Token(t)})
	require.ErrorIs(t, err, ErrNotConnected)
	require.Equal(t, CodeReconnectRequired, Classify(err).Code)
	require.Equal(t, Outcome{Code: CodeValidation, Field: FieldGitCredential}, Classify(ErrNoGitCredential))
}

func TestU4_GitAccess_BindingFollowsEveryChange(t *testing.T) {
	u, _ := newGit(t)
	_, _, err := u.svc.GitCredential(u.ctx, ws)
	require.ErrorIs(t, err, ErrNoGitCredential, "none stored")

	u.setGit(t)
	_, b1, err := u.svc.GitCredential(u.ctx, ws)
	require.NoError(t, err)
	require.Equal(t, "2.1", b1, "<connectionEpoch>.<revision>")

	u.setGit(t)
	_, b2, _ := u.svc.GitCredential(u.ctx, ws)
	require.Equal(t, "2.2", b2, "a new save changes it")

	u.connectKey(t, "a.backlog.com")
	_, b3, err := u.svc.GitCredential(u.ctx, ws)
	require.NoError(t, err, "a same-host replacement keeps the Git password")
	require.Equal(t, "3.2", b3)

	_, err = u.svc.Disconnect(u.ctx, ws)
	require.NoError(t, err)
	_, _, err = u.svc.GitCredential(u.ctx, ws)
	require.ErrorIs(t, err, ErrNotConnected)
}

func TestU4_GitAccess_AnotherHostIsNoCredential(t *testing.T) {
	u, _ := newGit(t)
	u.secrets.mu.Lock()
	u.secrets.data["backlog.git."+ws] = `{"username":"lan","password":"TESTSECRET-x","spaceHost":"b.backlog.com","revision":1}`
	u.secrets.mu.Unlock()
	_, _, err := u.svc.GitCredential(u.ctx, ws)
	require.ErrorIs(t, err, ErrNoGitCredential)
}

func TestU4_GitAccess_TestProbesTheFirstRepository(t *testing.T) {
	u, gg := newGit(t)
	view, err := u.svc.Test(u.ctx, ws)
	require.NoError(t, err)
	require.Empty(t, view.GitCheck, "no Git credential: no probe")
	require.Empty(t, gg.gitCalls)

	u.setGit(t)
	view, err = u.svc.Test(u.ctx, ws)
	require.NoError(t, err)
	require.Equal(t, GitCheckOK, view.GitCheck)
	require.Equal(t, []string{"PROJ"}, gg.repoCalls)
	require.Equal(t, []string{"a.backlog.com lan PROJ/web-app"}, gg.gitCalls)

	gg.gitErr = &backlog.Error{Kind: backlog.KindUnauthorized, Status: 401}
	view, err = u.svc.Test(u.ctx, ws)
	require.NoError(t, err, "a Git 401 does not fail the connection test (AC5.5.2)")
	require.True(t, view.Connected)
	require.Equal(t, GitCheckInvalid, view.GitCheck)

	for name, mut := range map[string]func(){
		"a probe outage": func() { gg.gitErr = &backlog.Error{Kind: backlog.KindUnreachable, Status: 503} },
		"no repository":  func() { gg.gitErr, gg.repos = nil, nil },
		"a list error":   func() { gg.reposErr = &backlog.Error{Kind: backlog.KindUnreachable} },
	} {
		mut()
		view, err = u.svc.Test(u.ctx, ws)
		require.NoError(t, err, name)
		require.Equal(t, GitCheckUntested, view.GitCheck, name)
	}
}

func TestU4_GitAccess_DisconnectAndSpaceChangeDropIt(t *testing.T) {
	u, _ := newGit(t)
	u.setGit(t)
	events := subscribe(t, u.svc)
	_, err := u.svc.Disconnect(u.ctx, ws)
	require.NoError(t, err)
	require.Empty(t, u.secrets.snapshot(), "no API key, token or Git password is left (AC1.5.4)")
	require.Equal(t, ReasonDisconnected, events.next(t).Reason)
	select {
	case e := <-events.ch:
		t.Fatalf("a second event: %v", e.Reason)
	default:
	}

	u2, _ := newGit(t)
	u2.setGit(t)
	u2.connectKey(t, "b.backlog.jp")
	_, ok := u2.secrets.snapshot()["backlog.git."+ws]
	require.False(t, ok, "the old space's Git password is gone (AC1.8.2)")
	_, _, err = u2.svc.GitCredential(u2.ctx, ws)
	require.ErrorIs(t, err, ErrNoGitCredential)
}

func TestU4_GitAccess_PasswordNeverLeaks(t *testing.T) {
	u, gg := newGit(t)
	events := subscribe(t, u.svc)
	var texts []string
	record := func(v any, err error) {
		b, _ := json.Marshal(v)
		texts = append(texts, string(b), fmt.Sprintf("%+v", v))
		if err != nil {
			texts = append(texts, err.Error())
		}
	}
	pw := testutil.Token(t)
	record(u.svc.SetGitCredential(u.ctx, ws, GitCredentialInput{Username: "lan", Password: pw}))
	record(u.svc.SetGitCredential(u.ctx, ws, GitCredentialInput{Username: "la\x00n", Password: pw}))
	record(u.svc.Test(u.ctx, ws))
	gg.gitErr = &backlog.Error{Kind: backlog.KindUnauthorized, Status: 401}
	record(u.svc.Test(u.ctx, ws))
	c, b, err := u.svc.GitCredential(u.ctx, ws)
	texts = append(texts, fmt.Sprintf("%v %+v %#v %s", c, c, c, b))
	record(nil, err)
	record(u.svc.Disconnect(u.ctx, ws))
	texts = append(texts, fmt.Sprintf("%+v", events.next(t)))
	testutil.AssertNoLeak(t, strings.Join(texts, "\n"), pw, 8)
	testutil.AssertNoLeak(t, u.logs.String(), pw, 8)
}
