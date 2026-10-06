package issues

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

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/backlog"
	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/connection"
	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/redact"
	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/testutil"
)

const spaceHost = "example-space.backlog.com"

func notFound() error { return &backlog.Error{Kind: backlog.KindNotFound, Status: 404, Class: "http"} }
func forbidden() error {
	return &backlog.Error{Kind: backlog.KindForbidden, Status: 403, Class: "http"}
}
func rateLimited(d time.Duration) error {
	return &backlog.Error{Kind: backlog.KindRateLimited, Status: 429, RetryAfter: d, Class: "http"}
}

// fakeGateway is a scripted Backlog. It counts calls per operation and
// keeps Search calls one at a time, 1 s apart, like backlog.Client.
type fakeGateway struct {
	mu          sync.Mutex
	projects    []backlog.Project
	issues      []backlog.Issue // newest updated first
	comments    []backlog.Comment
	attachments []backlog.Attachment
	statuses    map[string][]backlog.Status
	users       map[string][]backlog.ProjectUser
	errs        map[string]error // by op, or "issue:<key>"
	delay       time.Duration    // every call takes this long (virtual time)
	block       chan struct{}    // when set, Issue calls wait for it to close

	calls       map[string]int
	classes     map[string][]backlog.CallClass
	queries     []backlog.IssueQuery
	inFlight    int
	maxInFlight int

	searchMu   sync.Mutex
	lastSearch time.Time
}

func newFakeGateway() *fakeGateway {
	g := &fakeGateway{errs: map[string]error{}, calls: map[string]int{}, classes: map[string][]backlog.CallClass{},
		statuses: map[string][]backlog.Status{}, users: map[string][]backlog.ProjectUser{}}
	g.projects = []backlog.Project{{ID: 101, Key: "PROJ", Name: "Test Project"}, {ID: 102, Key: "DEMO", Name: "Demo Project"}}
	g.issues = []backlog.Issue{
		{ID: 5118, ProjectID: 101, IssueKey: "PROJ-118", Summary: "Fix login timeout", Description: "Steps in the attachment.",
			StatusID: 2, StatusName: "In Progress", PriorityID: 2, PriorityName: "High", AssigneeID: 2, AssigneeName: "Lan",
			DueDate: "2026-10-10T00:00:00Z", Updated: "2026-10-01T09:00:00Z"},
		{ID: 5120, ProjectID: 101, IssueKey: "PROJ-120", Summary: "Login page", StatusID: 3, StatusName: "Resolved",
			PriorityID: 3, PriorityName: "Normal", Updated: "2026-09-30T09:00:00Z"},
		{ID: 5123, ProjectID: 101, IssueKey: "PROJ-123", Summary: "Session handling", StatusID: 1, StatusName: "Open",
			PriorityID: 4, PriorityName: "Low", AssigneeID: 1, AssigneeName: "Test User", Updated: "2026-09-29T09:00:00Z"},
		{ID: 6001, ProjectID: 102, IssueKey: "DEMO-1", Summary: "Demo login", StatusID: 1, StatusName: "Open", Updated: "2026-09-28T09:00:00Z"},
	}
	return g
}

// enter counts one call and waits the fake delay.
func (g *fakeGateway) enter(ctx context.Context, op string, class backlog.CallClass) error {
	g.mu.Lock()
	g.calls[op]++
	g.classes[op] = append(g.classes[op], class)
	g.inFlight++
	g.maxInFlight = max(g.maxInFlight, g.inFlight)
	delay, err := g.delay, g.errs[op]
	g.mu.Unlock()
	defer func() {
		g.mu.Lock()
		g.inFlight--
		g.mu.Unlock()
	}()
	if delay > 0 {
		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return err
}

// search keeps Search calls one at a time and 1 s apart.
func (g *fakeGateway) search(ctx context.Context, op string, class backlog.CallClass) error {
	g.searchMu.Lock()
	defer g.searchMu.Unlock()
	if !g.lastSearch.IsZero() {
		if d := time.Until(g.lastSearch.Add(time.Second)); d > 0 {
			time.Sleep(d)
		}
	}
	g.lastSearch = time.Now()
	return g.enter(ctx, op, class)
}

func (g *fakeGateway) count(op string) int {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.calls[op]
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

func (g *fakeGateway) peak() int {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.maxInFlight
}

func (g *fakeGateway) set(fn func(g *fakeGateway)) {
	g.mu.Lock()
	defer g.mu.Unlock()
	fn(g)
}

func (g *fakeGateway) Projects(ctx context.Context, _ backlog.Credentials) ([]backlog.Project, error) {
	if err := g.enter(ctx, "projects", backlog.Interactive); err != nil {
		return nil, err
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	return slices.Clone(g.projects), nil
}

func (g *fakeGateway) match(q backlog.IssueQuery) []backlog.Issue {
	var out []backlog.Issue
	kw := strings.ToLower(q.Keyword)
	for _, i := range g.issues {
		if Matches(i, q) && (kw == "" || strings.Contains(strings.ToLower(i.Summary), kw)) {
			out = append(out, i)
		}
	}
	return out
}

func (g *fakeGateway) Issues(ctx context.Context, _ backlog.Credentials, class backlog.CallClass, q backlog.IssueQuery) ([]backlog.Issue, error) {
	if err := g.search(ctx, "issues", class); err != nil {
		return nil, err
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	g.queries = append(g.queries, q)
	all := g.match(q)
	start := min(q.Offset, len(all))
	return slices.Clone(all[start:min(start+q.Count, len(all))]), nil
}

func (g *fakeGateway) IssueCount(ctx context.Context, _ backlog.Credentials, class backlog.CallClass, q backlog.IssueQuery) (int, error) {
	if err := g.search(ctx, "count", class); err != nil {
		return 0, err
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	return len(g.match(q)), nil
}

func (g *fakeGateway) Issue(ctx context.Context, _ backlog.Credentials, class backlog.CallClass, ref string) (backlog.Issue, error) {
	g.mu.Lock()
	block := g.block
	keyErr := g.errs["issue:"+ref]
	g.mu.Unlock()
	if block != nil {
		<-block
	}
	if err := g.enter(ctx, "issue", class); err != nil {
		return backlog.Issue{}, err
	}
	if keyErr != nil {
		return backlog.Issue{}, keyErr
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	for _, i := range g.issues {
		if i.IssueKey == ref {
			return i, nil
		}
	}
	return backlog.Issue{}, notFound()
}

func (g *fakeGateway) IssueComments(ctx context.Context, _ backlog.Credentials, class backlog.CallClass, _ string, q backlog.CommentQuery) ([]backlog.Comment, error) {
	if err := g.enter(ctx, "comments", class); err != nil {
		return nil, err
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	var out []backlog.Comment
	for _, c := range g.comments { // newest (highest id) first
		if (q.MaxID == 0 || c.ID <= q.MaxID) && len(out) < q.Count {
			out = append(out, c)
		}
	}
	return out, nil
}

func (g *fakeGateway) IssueAttachments(ctx context.Context, _ backlog.Credentials, class backlog.CallClass, _ string) ([]backlog.Attachment, error) {
	if err := g.enter(ctx, "attachments", class); err != nil {
		return nil, err
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	return slices.Clone(g.attachments), nil
}

func (g *fakeGateway) ProjectStatuses(ctx context.Context, _ backlog.Credentials, projectKey string) ([]backlog.Status, error) {
	if err := g.enter(ctx, "statuses", backlog.Interactive); err != nil {
		return nil, err
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	return slices.Clone(g.statuses[projectKey]), nil
}

func (g *fakeGateway) ProjectUsers(ctx context.Context, _ backlog.Credentials, projectKey string) ([]backlog.ProjectUser, error) {
	if err := g.enter(ctx, "users", backlog.Interactive); err != nil {
		return nil, err
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	return slices.Clone(g.users[projectKey]), nil
}

// fakeConn is the ConnectionReader.
type fakeConn struct {
	mu         sync.Mutex
	snap       connection.Snapshot
	currentErr error
	apiKey     string
	token      string
	disabled   bool
	enabledErr error
	subs       []func(connection.ConnectionChanged)
}

func (c *fakeConn) Current(context.Context, string) (connection.Snapshot, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return connection.Snapshot{SpaceHost: c.snap.SpaceHost, ConnectionEpoch: c.snap.ConnectionEpoch,
		SelectedProjects: slices.Clone(c.snap.SelectedProjects)}, c.currentErr
}

func (c *fakeConn) Credentials(context.Context, string) (backlog.Credentials, int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.currentErr != nil {
		return backlog.Credentials{}, 0, c.currentErr
	}
	return backlog.Credentials{SpaceHost: c.snap.SpaceHost, APIKey: c.apiKey, AccessToken: c.token}, c.snap.ConnectionEpoch, nil
}

func (c *fakeConn) Subscribe(fn func(connection.ConnectionChanged)) func() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.subs = append(c.subs, fn)
	return func() {}
}

func (c *fakeConn) RequireEnabled(context.Context, string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.enabledErr != nil {
		return c.enabledErr
	}
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

// fakeHost is the Kandev host: tasks get ids task-<n> and keys T-<n>.
type fakeHost struct {
	mu        sync.Mutex
	tasks     []TaskInfo
	creates   []NewTask
	next      int
	failNth   int // the Nth create fails (1-based); 0 = never
	listErr   error
	listCalls int
}

func (h *fakeHost) CreateTask(_ context.Context, in NewTask) (TaskRef, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.creates = append(h.creates, in)
	if h.failNth == len(h.creates) {
		return TaskRef{}, fmt.Errorf("rpc error: code = PermissionDenied")
	}
	h.next++
	ref := TaskRef{ID: fmt.Sprintf("task-%d", h.next), Key: fmt.Sprintf("T-%d", h.next)}
	h.tasks = append(h.tasks, TaskInfo{ID: ref.ID, Key: ref.Key, Title: in.Title})
	return ref, nil
}

func (h *fakeHost) ListTasks(context.Context, string) ([]TaskInfo, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.listCalls++
	return slices.Clone(h.tasks), h.listErr
}

func (h *fakeHost) lists() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.listCalls
}

func (h *fakeHost) createCount() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.creates)
}

func (h *fakeHost) remove(id string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.tasks = slices.DeleteFunc(h.tasks, func(t TaskInfo) bool { return t.ID == id })
}

type syncBuf struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (s *syncBuf) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.Write(p)
}

func (s *syncBuf) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.String()
}

type rig struct {
	svc   *Service
	gw    *fakeGateway
	conn  *fakeConn
	host  *fakeHost
	state *memState
	store *Store
	logs  *syncBuf
	log   *slog.Logger
	ctx   context.Context
}

// newRig is connected to example-space.backlog.com at epoch 1 with PROJ and
// DEMO selected, and has tasks task-17 (T-17) and task-18 (T-18).
func newRig(t *testing.T) *rig {
	t.Helper()
	gw, state, logs := newFakeGateway(), newMemState(), &syncBuf{}
	conn := &fakeConn{snap: connection.Snapshot{SpaceHost: spaceHost, ConnectionEpoch: 1, SelectedProjects: []string{"PROJ", "DEMO"}},
		apiKey: testutil.APIKey(t)}
	host := &fakeHost{tasks: []TaskInfo{{ID: "task-17", Key: "T-17", Title: "Login work"}, {ID: "task-18", Key: "T-18", Title: "Other"}}, next: 18}
	store := NewStore(state)
	log := slog.New(redact.NewHandler(slog.NewJSONHandler(logs, &slog.HandlerOptions{Level: slog.LevelDebug})))
	svc := NewService(gw, conn, host, store)
	return &rig{svc: svc, gw: gw, conn: conn, host: host, state: state, store: store, logs: logs, log: log,
		ctx: redact.WithLogger(context.Background(), log)}
}

// link stores a link directly, as an earlier action would have.
func (r *rig) link(t *testing.T, l Link) {
	t.Helper()
	if l.State == "" {
		l.State = StateActive
	}
	if l.SpaceHost == "" {
		l.SpaceHost = spaceHost
	}
	if l.ProjectKey == "" {
		l.ProjectKey, _, _ = ParseIssueKey(l.IssueKey)
	}
	if err := r.store.UpdateLinks(context.Background(), "ws-1", func(ls []Link) ([]Link, error) { return append(ls, l), nil }); err != nil {
		t.Fatal(err)
	}
}

func (r *rig) links(t *testing.T) []Link {
	t.Helper()
	ls, err := r.store.Links(context.Background(), "ws-1")
	if err != nil {
		t.Fatal(err)
	}
	return ls
}

// byTask returns the task's link, or a zero Link.
func byTask(ls []Link, taskID string) Link {
	for _, l := range ls {
		if l.TaskID == taskID {
			return l
		}
	}
	return Link{}
}
