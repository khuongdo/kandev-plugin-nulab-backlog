package git

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/backlog"
	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/connection"
	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/redact"
	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/testutil"
)

// fakeGateway scripts Backlog and counts requests per method.
type fakeGateway struct {
	mu        sync.Mutex
	calls     map[string]int
	classes   []backlog.CallClass // class of each PullRequests call
	repos     map[string][]backlog.Repository
	prs       map[string][]backlog.PullRequest // by "PROJ/repo"
	queries   []backlog.PullRequestQuery
	err       map[string]error // by method
	created   []backlog.NewPullRequest
	issues    map[string]backlog.Issue
	myself    backlog.User
	onPRs     func() // runs inside PullRequests
	onPR      func() // runs inside PullRequest
	logWait   bool   // PullRequests logs a backlog_wait line, as the client does for a 429
	nextPRNum int
}

func newFakeGateway() *fakeGateway {
	return &fakeGateway{
		calls: map[string]int{},
		err:   map[string]error{},
		repos: map[string][]backlog.Repository{
			"PROJ": {{ID: 11, ProjectID: 101, Name: "web-app", HTTPURL: "https://" + host + "/git/PROJ/web-app.git"},
				{ID: 12, ProjectID: 101, Name: "api", HTTPURL: "https://" + host + "/git/PROJ/api.git"}},
			"DEMO": {{ID: 21, ProjectID: 102, Name: "demo-app", HTTPURL: "https://" + host + "/git/DEMO/demo-app.git"}},
		},
		prs:       map[string][]backlog.PullRequest{},
		issues:    map[string]backlog.Issue{"PROJ-120": {ID: 5120, IssueKey: "PROJ-120"}},
		myself:    backlog.User{ID: 1234, Name: "Test User"},
		nextPRNum: 100,
	}
}

func (g *fakeGateway) hit(method string) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.calls[method]++
	return g.err[method]
}

func (g *fakeGateway) count(method string) int {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.calls[method]
}

func (g *fakeGateway) total() int {
	g.mu.Lock()
	defer g.mu.Unlock()
	n := 0
	for _, c := range g.calls {
		n += c
	}
	return n
}

func (g *fakeGateway) setErr(method string, err error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.err[method] = err
}

func (g *fakeGateway) setPRs(repo string, prs []backlog.PullRequest) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.prs[repo] = prs
}

func (g *fakeGateway) Myself(context.Context, backlog.Credentials) (backlog.User, error) {
	if err := g.hit("Myself"); err != nil {
		return backlog.User{}, err
	}
	return g.myself, nil
}

func (g *fakeGateway) Repositories(_ context.Context, _ backlog.Credentials, project string) ([]backlog.Repository, error) {
	if err := g.hit("Repositories"); err != nil {
		return nil, err
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.repos[project], nil
}

func (g *fakeGateway) PullRequests(ctx context.Context, _ backlog.Credentials, class backlog.CallClass, project, repo string, q backlog.PullRequestQuery) ([]backlog.PullRequest, error) {
	err := g.hit("PullRequests")
	g.mu.Lock()
	g.classes = append(g.classes, class)
	g.queries = append(g.queries, q)
	hook := g.onPRs
	all := slices.Clone(g.prs[project+"/"+repo])
	g.mu.Unlock()
	if hook != nil {
		hook()
	}
	if g.logWait {
		redact.Logger(ctx).InfoContext(ctx, "backlog wait", "event", "backlog_wait", "reason", "rate_limited")
	}
	if err != nil {
		return nil, err
	}
	var out []backlog.PullRequest
	for _, pr := range all {
		if len(q.StatusIDs) > 0 && !slices.Contains(q.StatusIDs, int64(pr.StatusID)) {
			continue
		}
		out = append(out, pr)
	}
	out = out[min(q.Offset, len(out)):]
	if len(out) > q.Count && q.Count > 0 {
		out = out[:q.Count]
	}
	return out, nil
}

func (g *fakeGateway) PullRequestCount(_ context.Context, _ backlog.Credentials, _ backlog.CallClass, project, repo string, q backlog.PullRequestQuery) (int, error) {
	if err := g.hit("PullRequestCount"); err != nil {
		return 0, err
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	n := 0
	for _, pr := range g.prs[project+"/"+repo] {
		if len(q.StatusIDs) == 0 || slices.Contains(q.StatusIDs, int64(pr.StatusID)) {
			n++
		}
	}
	return n, nil
}

func (g *fakeGateway) PullRequest(_ context.Context, _ backlog.Credentials, _ backlog.CallClass, project, repo string, number int) (backlog.PullRequest, error) {
	if err := g.hit("PullRequest"); err != nil {
		return backlog.PullRequest{}, err
	}
	g.mu.Lock()
	hook := g.onPR
	g.mu.Unlock()
	if hook != nil {
		hook()
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	for _, pr := range g.prs[project+"/"+repo] {
		if pr.Number == number {
			return pr, nil
		}
	}
	return backlog.PullRequest{}, &backlog.Error{Kind: backlog.KindNotFound, Status: 404}
}

func (g *fakeGateway) CreatePullRequest(_ context.Context, _ backlog.Credentials, project, repo string, in backlog.NewPullRequest) (backlog.PullRequest, error) {
	if err := g.hit("CreatePullRequest"); err != nil {
		return backlog.PullRequest{}, err
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	g.created = append(g.created, in)
	g.nextPRNum++
	pr := backlog.PullRequest{ID: 9000, RepositoryID: 11, Number: g.nextPRNum, Summary: in.Summary, Base: in.Base, Branch: in.Branch, StatusID: 1}
	g.prs[project+"/"+repo] = append(g.prs[project+"/"+repo], pr)
	return pr, nil
}

func (g *fakeGateway) Issue(_ context.Context, _ backlog.Credentials, _ backlog.CallClass, key string) (backlog.Issue, error) {
	if err := g.hit("Issue"); err != nil {
		return backlog.Issue{}, err
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	issue, ok := g.issues[key]
	if !ok {
		return backlog.Issue{}, &backlog.Error{Kind: backlog.KindNotFound, Status: 404}
	}
	return issue, nil
}

// fakeConn is a scripted ConnectionReader.
type fakeConn struct {
	mu      sync.Mutex
	snap    connection.Snapshot
	err     error
	creds   backlog.Credentials
	git     connection.GitCredential
	binding string
	gitErr  error
	subs    []func(connection.ConnectionChanged)
	// disabled is the workspace's Backlog switch turned off; switchErr is
	// the switch store failing.
	disabled  bool
	switchErr error
}

func (c *fakeConn) RequireEnabled(context.Context, string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.switchErr != nil {
		return c.switchErr
	}
	if c.disabled {
		return connection.ErrIntegrationDisabled
	}
	return nil
}

func newFakeConn(t *testing.T) *fakeConn {
	return &fakeConn{
		snap:    connection.Snapshot{SpaceHost: host, AuthMethod: "api_key", ConnectionEpoch: 2, SelectedProjects: []string{"PROJ"}},
		creds:   backlog.Credentials{SpaceHost: host, APIKey: testutil.APIKey(t)},
		git:     connection.GitCredential{Username: "lan", Password: testutil.Token(t), SpaceHost: host},
		binding: "2.1",
	}
}

func (c *fakeConn) Current(context.Context, string) (connection.Snapshot, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.snap, c.err
}

func (c *fakeConn) Credentials(context.Context, string) (backlog.Credentials, int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.creds, c.snap.ConnectionEpoch, c.err
}

func (c *fakeConn) GitCredential(context.Context, string) (connection.GitCredential, string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.err != nil {
		return connection.GitCredential{}, "", c.err
	}
	return c.git, c.binding, c.gitErr
}

func (c *fakeConn) Subscribe(fn func(connection.ConnectionChanged)) func() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.subs = append(c.subs, fn)
	return func() {}
}

func (c *fakeConn) update(fn func(*fakeConn)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	fn(c)
}

// fakeHost is a HostPort that records created tasks and can fail.
type fakeHost struct {
	mu        sync.Mutex
	tasks     []NewTask
	taskIDs   []string
	deleted   map[string]bool
	failOn    int // fail the Nth CreateTask (1-based, counted over the fake's life)
	creates   int
	repos     map[string]KandevRepository
	findCalls int
	findErr   error
}

func newFakeHost() *fakeHost {
	return &fakeHost{deleted: map[string]bool{}, repos: map[string]KandevRepository{
		"repo-k1": {ID: "repo-k1", ProviderID: ProviderID, ProviderRepositoryID: "11", ProviderHost: "https://" + host,
			ProviderScope: host, OwnerOrProject: "PROJ", Name: "web-app", DefaultBranch: "main"},
		"repo-gh": {ID: "repo-gh", ProviderID: "github", OwnerOrProject: "acme", Name: "web"},
	}}
}

func (h *fakeHost) CreateTask(_ context.Context, in NewTask) (string, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.creates++
	if h.failOn != 0 && h.creates == h.failOn {
		return "", errInjected
	}
	id := fmt.Sprintf("task-%d", len(h.tasks)+1)
	h.tasks = append(h.tasks, in)
	h.taskIDs = append(h.taskIDs, id)
	return id, nil
}

func (h *fakeHost) FindTaskByMetadata(_ context.Context, ws, key, value string) (string, bool, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.findCalls++
	if h.findErr != nil {
		return "", false, h.findErr
	}
	for i, t := range h.tasks {
		if t.WorkspaceID == ws && t.Metadata[key] == value && !h.deleted[h.taskIDs[i]] {
			return h.taskIDs[i], true, nil
		}
	}
	return "", false, nil
}

func (h *fakeHost) Repository(_ context.Context, _ string, id string) (KandevRepository, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	r, ok := h.repos[id]
	if !ok {
		return KandevRepository{}, ErrRepositoryNotFound
	}
	return r, nil
}

func (h *fakeHost) taskCount() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.tasks)
}

func (h *fakeHost) titles() []string {
	h.mu.Lock()
	defer h.mu.Unlock()
	var out []string
	for _, t := range h.tasks {
		out = append(out, t.Title)
	}
	return out
}

// rig wires a Service over the fakes.
type rig struct {
	svc   *Service
	gw    *fakeGateway
	conn  *fakeConn
	host  *fakeHost
	state *fakeState
	store *Store
	logs  *syncBuffer
	ctx   context.Context
}

var testNow = time.Date(2026, 10, 6, 12, 0, 0, 0, time.UTC)

func newRig(t *testing.T) *rig {
	t.Helper()
	r := &rig{gw: newFakeGateway(), conn: newFakeConn(t), host: newFakeHost(), state: newFakeState(), logs: &syncBuffer{}}
	r.store = NewStore(r.state)
	r.svc = NewService(r.gw, r.conn, r.host, r.store)
	r.svc.Now = func() time.Time { return testNow }
	log := slog.New(redact.NewHandler(slog.NewJSONHandler(r.logs, &slog.HandlerOptions{Level: slog.LevelDebug})))
	r.ctx = redact.WithLogger(context.Background(), log)
	return r
}

// prs makes n open PRs on PROJ/web-app numbered 1..n, newest first like Backlog.
func prs(n int) []backlog.PullRequest {
	out := make([]backlog.PullRequest, 0, n)
	for i := n; i >= 1; i-- {
		out = append(out, backlog.PullRequest{ID: int64(1000 + i), RepositoryID: 11, Number: i,
			Summary: fmt.Sprintf("Change %d", i), Base: "main", Branch: fmt.Sprintf("feature/%d", i), StatusID: 1, AssigneeName: "Lan"})
	}
	return out
}

func (r *rig) links(t *testing.T) []Link {
	t.Helper()
	l, err := r.store.Links(context.Background(), ws)
	require.NoError(t, err)
	return l
}

func (r *rig) watches(t *testing.T) []Watch {
	t.Helper()
	w, err := r.store.Watches(context.Background(), ws)
	require.NoError(t, err)
	return w
}

func (r *rig) addLink(t *testing.T, l Link) {
	t.Helper()
	require.NoError(t, r.store.UpdateLinks(context.Background(), ws, func(ls []Link) ([]Link, error) { return append(ls, l), nil }))
}

func (r *rig) saveWatch(t *testing.T, mut func(*WatchInput)) Watch {
	t.Helper()
	in := WatchInput{Name: "Reviews", ProjectKey: "PROJ", RepoName: "web-app", Statuses: []string{"open"},
		Assignee: WhoAnyone, Creator: WhoAnyone, WorkflowID: "wf-1", WorkflowStepID: "step-1"}
	if mut != nil {
		mut(&in)
	}
	w, err := r.svc.SaveWatch(r.ctx, ws, in)
	require.NoError(t, err)
	return w
}

// events parses the JSON log lines named name.
func (r *rig) events(t *testing.T, name string) []map[string]any {
	t.Helper()
	var out []map[string]any
	for _, line := range strings.Split(strings.TrimSpace(r.logs.String()), "\n") {
		if line == "" {
			continue
		}
		var m map[string]any
		require.NoError(t, json.Unmarshal([]byte(line), &m))
		if m["event"] == name {
			out = append(out, m)
		}
	}
	return out
}

type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *syncBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *syncBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

func backlogErr(kind backlog.Kind, status int) error {
	return &backlog.Error{Kind: kind, Status: status}
}

// startWatcher starts a watcher over r's service; it stops at cleanup.
func (r *rig) startWatcher(t *testing.T) *Watcher {
	t.Helper()
	w := NewWatcher(r.svc, redact.Logger(r.ctx))
	w.Start()
	t.Cleanup(w.Stop)
	return w
}
