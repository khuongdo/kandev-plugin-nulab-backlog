package scm

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/connection"
	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/redact"
	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/testutil"
)

// fakeClient scripts one provider and counts calls per method.
type fakeClient struct {
	mu     sync.Mutex
	p      Provider
	user   User
	users  map[string]User // by token; else user
	seen   []string        // "<workspace tag> <method> <login of the token>"
	repos  []Repo
	prs    map[string][]PullRequest // by repo
	err    map[string]error         // by method
	calls  map[string]int
	creds  []Credential
	delay  time.Duration
	lastQ  ListQuery
	onCall func(method string)
}

func newFakeClient(p Provider) *fakeClient {
	return &fakeClient{p: p, user: User{ID: "lan-id", Name: "Lan"}, prs: map[string][]PullRequest{},
		err: map[string]error{}, calls: map[string]int{},
		repos: []Repo{{FullName: "acme/web", URL: "https://example.test/acme/web"}, {FullName: "acme/api"}}}
}

// wsTag marks a test context with its workspace, so a fake client call
// records which workspace made it.
type wsTag struct{}

func inWS(ctx context.Context, ws string) context.Context { return context.WithValue(ctx, wsTag{}, ws) }

func (f *fakeClient) hit(ctx context.Context, method string, cred Credential) error {
	if f.delay > 0 {
		time.Sleep(f.delay)
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls[method]++
	f.creds = append(f.creds, cred)
	tag, _ := ctx.Value(wsTag{}).(string)
	login := "?"
	if u, ok := f.users[cred.Token]; ok {
		login = u.ID
	}
	f.seen = append(f.seen, tag+" "+method+" "+login)
	if f.onCall != nil {
		f.onCall(method)
	}
	return f.err[method]
}

func (f *fakeClient) count(method string) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.calls[method]
}

func (f *fakeClient) setErr(method string, err error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.err[method] = err
}

func (f *fakeClient) setPRs(repo string, prs ...PullRequest) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.prs[repo] = prs
}

func (f *fakeClient) CurrentUser(ctx context.Context, c Credential) (User, error) {
	if err := f.hit(ctx, "CurrentUser", c); err != nil {
		return User{}, err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if u, ok := f.users[c.Token]; ok {
		return u, nil
	}
	return f.user, nil
}

func (f *fakeClient) SearchRepos(ctx context.Context, c Credential, q string) ([]Repo, error) {
	if err := f.hit(ctx, "SearchRepos", c); err != nil {
		return nil, err
	}
	var out []Repo
	for _, r := range f.repos {
		if strings.Contains(r.FullName, q) {
			out = append(out, r)
		}
	}
	return out, nil
}

func (f *fakeClient) GetRepo(ctx context.Context, c Credential, repo string) (Repo, error) {
	if err := f.hit(ctx, "GetRepo", c); err != nil {
		return Repo{}, err
	}
	for _, r := range f.repos {
		if strings.EqualFold(r.FullName, repo) {
			return r, nil
		}
	}
	return Repo{}, &HTTPError{Provider: f.p, Status: 404}
}

func (f *fakeClient) ListPRs(ctx context.Context, c Credential, repo string, q ListQuery) (PRPage, error) {
	if err := f.hit(ctx, "ListPRs", c); err != nil {
		return PRPage{}, err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.lastQ = q
	page := PRPage{Items: []PullRequest{}}
	for _, pr := range f.prs[repo] {
		if q.Wants(pr.State) {
			page.Items = append(page.Items, pr)
		}
	}
	return page, nil
}

func (f *fakeClient) GetPR(ctx context.Context, c Credential, repo string, n int) (PullRequest, error) {
	if err := f.hit(ctx, "GetPR", c); err != nil {
		return PullRequest{}, err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, pr := range f.prs[repo] {
		if pr.Number == n {
			return pr, nil
		}
	}
	return PullRequest{}, &HTTPError{Provider: f.p, Status: 404}
}

// fakeConn is the Backlog connection: connected to PROJ and DEMO unless changed.
type fakeConn struct {
	mu       sync.Mutex
	snap     connection.Snapshot
	err      error
	disabled bool
}

func (c *fakeConn) Current(context.Context, string) (connection.Snapshot, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.snap, c.err
}

func (c *fakeConn) RequireEnabled(context.Context, string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.disabled {
		return connection.ErrIntegrationDisabled
	}
	return nil
}

func (c *fakeConn) set(fn func(c *fakeConn)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	fn(c)
}

type fakeSecrets struct {
	mu      sync.Mutex
	data    map[string]string
	failGet bool
}

func (s *fakeSecrets) GetSecret(_ context.Context, key string) (string, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.failGet {
		return "", false, errInjected
	}
	v, ok := s.data[key]
	return v, ok, nil
}

func (s *fakeSecrets) SetSecret(_ context.Context, key, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
	return nil
}

func (s *fakeSecrets) DeleteSecret(_ context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, key)
	return nil
}

type fakeTasks struct {
	mu      sync.Mutex
	created []NewTask
	err     error
}

func (f *fakeTasks) CreateTask(_ context.Context, in NewTask) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.err != nil {
		return "", f.err
	}
	f.created = append(f.created, in)
	return fmt.Sprintf("task-%d", len(f.created)), nil
}

func (f *fakeTasks) all() []NewTask {
	f.mu.Lock()
	defer f.mu.Unlock()
	return slices.Clone(f.created)
}

type harness struct {
	t       *testing.T
	svc     *Service
	clients map[Provider]*fakeClient
	conn    *fakeConn
	secrets *fakeSecrets
	tasks   *fakeTasks
	state   *fakeState
	logs    *bytes.Buffer
	ctx     context.Context
	now     time.Time
	tokens  map[Provider]string
}

const ws = "ws-1"

func newHarness(t *testing.T) *harness {
	t.Helper()
	h := &harness{
		t:       t,
		clients: map[Provider]*fakeClient{},
		conn:    &fakeConn{snap: connection.Snapshot{SpaceHost: "example-space.backlog.com", SelectedProjects: []string{"PROJ", "DEMO"}}},
		secrets: &fakeSecrets{data: map[string]string{}},
		tasks:   &fakeTasks{},
		state:   newFakeState(),
		logs:    &bytes.Buffer{},
		now:     time.Date(2026, 10, 7, 9, 0, 0, 0, time.UTC),
		tokens:  map[Provider]string{},
	}
	clients := map[Provider]Client{}
	for _, p := range Providers {
		h.clients[p] = newFakeClient(p)
		clients[p] = h.clients[p]
		h.tokens[p] = testutil.Token(t)
	}
	h.svc = NewService(clients, h.conn, h.secrets, h.tasks, NewStore(h.state))
	h.svc.Now = func() time.Time { return h.now }
	h.ctx = redact.WithLogger(context.Background(), slog.New(redact.NewHandler(
		slog.NewJSONHandler(h.logs, &slog.HandlerOptions{Level: slog.LevelDebug}))))
	return h
}

// connect stores a token for p (Bitbucket with a user name). It makes p
// active for the call, then clears the stored service, so the active one is
// derived: p alone, or pending (all usable) when several are connected.
func (h *harness) connect(t *testing.T, p Provider) {
	t.Helper()
	in := TokenInput{Provider: p, Token: h.tokens[p]}
	if p == Bitbucket {
		in.Username = "lan@example.com"
	}
	h.use(t, p)
	_, err := h.svc.SetToken(h.ctx, ws, in)
	require.NoError(t, err)
	require.NoError(t, h.svc.store.SetActive(h.ctx, ws, ""))
}

// use stores p as the active service.
func (h *harness) use(t *testing.T, p Provider) {
	t.Helper()
	require.NoError(t, h.svc.SetActive(h.ctx, ws, p))
}

// mapRepo connects p and maps repos to project.
func (h *harness) mapRepo(t *testing.T, p Provider, project string, repos ...string) {
	t.Helper()
	if !h.hasToken(p) {
		h.connect(t, p)
	}
	_, err := h.svc.SetMapping(h.ctx, ws, MappingInput{Provider: p, ProjectKey: project, Repos: repos})
	require.NoError(t, err)
}

func (h *harness) hasToken(p Provider) bool {
	_, ok, _ := h.secrets.GetSecret(context.Background(), SecretKey(p, ws))
	return ok
}

func pr(p Provider, repo string, n int, title, branch, state string) PullRequest {
	ref := PRRef{Provider: p, Repo: repo, Number: n}
	return PullRequest{PRRef: ref, Title: title, State: state, Author: "Lan", AuthorID: "lan-id",
		SourceBranch: branch, TargetBranch: "main", UpdatedAt: "2026-10-06T09:00:00Z", URL: PRURL(ref)}
}

func (h *harness) allTokens() []string {
	var out []string
	for _, t := range h.tokens {
		out = append(out, t)
	}
	return out
}

// assertNoTokenLeak checks the logs and the given texts for every token (NFR1).
func (h *harness) assertNoTokenLeak(t *testing.T, texts ...string) {
	t.Helper()
	all := h.logs.String() + strings.Join(texts, "\n")
	for _, tok := range h.allTokens() {
		testutil.AssertNoLeak(t, all, tok)
	}
}
