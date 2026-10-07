package issues

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/backlog"
	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/connection"
	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/redact"
)

// Errors the plugin maps to action error codes.
var (
	// ErrNotFound is an issue or link that does not exist (not_found).
	ErrNotFound = errors.New("not found")
	// ErrIssueUnavailable is an issue Backlog answers 404 or 403 for: deleted,
	// moved or hidden. It is never reconnect_required.
	ErrIssueUnavailable = fmt.Errorf("issue unavailable: %w", ErrNotFound)
	// ErrNotLinked is a task without an issue link.
	ErrNotLinked = fmt.Errorf("task has no Backlog issue link: %w", ErrNotFound)
	// ErrConflict is an action the current links do not allow (conflict).
	ErrConflict = errors.New("not allowed in the current state")
	// ErrStale means the connection changed while the action ran, so its
	// result was dropped (AC1.8.3).
	ErrStale = errors.New("the Backlog connection changed meanwhile")
	// ErrWorkflowMissing is a task creation Kandev refused because the
	// workflow or step no longer exists (BR3.14). HostPort wraps it.
	ErrWorkflowMissing = errors.New("the workflow or step does not exist")
)

// Gateway is the Backlog calls U3 makes; all are GETs. backlog.Client implements it.
type Gateway interface {
	Myself(ctx context.Context, creds backlog.Credentials) (backlog.User, error)
	Projects(ctx context.Context, creds backlog.Credentials) ([]backlog.Project, error)
	Issues(ctx context.Context, creds backlog.Credentials, class backlog.CallClass, q backlog.IssueQuery) ([]backlog.Issue, error)
	IssueCount(ctx context.Context, creds backlog.Credentials, class backlog.CallClass, q backlog.IssueQuery) (int, error)
	Issue(ctx context.Context, creds backlog.Credentials, class backlog.CallClass, ref string) (backlog.Issue, error)
	IssueComments(ctx context.Context, creds backlog.Credentials, class backlog.CallClass, ref string, q backlog.CommentQuery) ([]backlog.Comment, error)
	IssueAttachments(ctx context.Context, creds backlog.Credentials, class backlog.CallClass, ref string) ([]backlog.Attachment, error)
	ProjectStatuses(ctx context.Context, creds backlog.Credentials, projectKey string) ([]backlog.Status, error)
	ProjectUsers(ctx context.Context, creds backlog.Credentials, projectKey string) ([]backlog.ProjectUser, error)
}

// Connection is the ConnectionReader U3 uses (C3). connection.Service implements it.
type Connection interface {
	Current(ctx context.Context, workspaceID string) (connection.Snapshot, error)
	Credentials(ctx context.Context, workspaceID string) (backlog.Credentials, int, error)
	Subscribe(fn func(connection.ConnectionChanged)) (unsubscribe func())
	// RequireEnabled returns connection.ErrIntegrationDisabled while the
	// workspace's Backlog switch is off, or the store error (fail closed).
	RequireEnabled(ctx context.Context, workspaceID string) error
}

// TaskRef is a created Kandev task: its id and its human key (T-17).
type TaskRef struct {
	ID  string
	Key string
}

// TaskInfo is one task of a workspace.
type TaskInfo struct {
	ID    string
	Key   string
	Title string
}

// HostPort is the Kandev host API U3 needs (C2). internal/plugin implements
// it on pluginsdk.Host; tests swap in a fake.
type HostPort interface {
	// CreateTask creates a task; a missing workflow or step wraps ErrWorkflowMissing.
	CreateTask(ctx context.Context, in NewTask) (TaskRef, error)
	// ListTasks lists the workspace's tasks, archived ones included.
	ListTasks(ctx context.Context, workspaceID string) ([]TaskInfo, error)
}

const (
	suggestBudget   = 1200 * time.Millisecond // inside Kandev's 1.5 s per search or authorization
	suggestSettle   = 250 * time.Millisecond  // a newer query in this time replaces the older one (AC3.4.2)
	suggestDefault  = 5
	suggestMax      = 10
	searchTaskRows  = 20
	commentsPerPage = 20
	previewLimit    = 10 << 20 // attachments over 10 MiB are "too large to preview" (AC3.5.2)
)

// Service is the IssueIntegration component.
type Service struct {
	gateway Gateway
	conn    Connection
	host    HostPort
	store   *Store
	Now     func() time.Time                                 // clock for link and cycle times
	Wait    func(ctx context.Context, d time.Duration) error // waits d or until ctx ends

	mu         sync.Mutex
	lastEpoch  map[string]int // the last ConnectionChanged epoch handled per workspace
	projects   map[string]projectCache
	creating   map[string]bool // (workspace, issue key) of running creates
	suggestSeq map[string]int  // the latest suggestion query per workspace
}

type projectCache struct {
	epoch int
	byKey map[string]backlog.Project
}

// NewService wires the IssueIntegration component.
func NewService(gateway Gateway, conn Connection, host HostPort, store *Store) *Service {
	return &Service{gateway: gateway, conn: conn, host: host, store: store, Now: time.Now, Wait: sleepCtx,
		lastEpoch: map[string]int{}, projects: map[string]projectCache{}, creating: map[string]bool{}, suggestSeq: map[string]int{}}
}

func sleepCtx(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

func (s *Service) now() string { return s.Now().UTC().Format(time.RFC3339) }

// credentials returns the connection's credentials and a context that redacts them.
func (s *Service) credentials(ctx context.Context, ws string) (context.Context, backlog.Credentials, error) {
	creds, _, err := s.conn.Credentials(ctx, ws)
	if err != nil {
		return ctx, creds, err
	}
	return redact.WithSecrets(ctx, creds.APIKey, creds.AccessToken), creds, nil
}

// unchanged returns ErrStale when the connection epoch moved since epoch.
func (s *Service) unchanged(ctx context.Context, ws string, epoch int) error {
	snap, err := s.conn.Current(ctx, ws)
	if err != nil {
		return err
	}
	if snap.ConnectionEpoch != epoch {
		return ErrStale
	}
	return nil
}

// unavailable reports a Backlog 404 or 403 for one issue (R-06 for U3).
func unavailable(err error) bool {
	var be *backlog.Error
	return errors.As(err, &be) && (be.Kind == backlog.KindNotFound || be.Kind == backlog.KindForbidden)
}

// issue reads one issue; a 404 or 403 is ErrIssueUnavailable.
func (s *Service) issue(ctx context.Context, creds backlog.Credentials, class backlog.CallClass, key string) (backlog.Issue, error) {
	i, err := s.gateway.Issue(ctx, creds, class, key)
	if unavailable(err) {
		return backlog.Issue{}, ErrIssueUnavailable
	}
	return i, err
}

// projectsFor returns the space's projects by key, cached per (workspace, epoch).
func (s *Service) projectsFor(ctx context.Context, ws string, epoch int, creds backlog.Credentials) (map[string]backlog.Project, error) {
	s.mu.Lock()
	c, ok := s.projects[ws]
	s.mu.Unlock()
	if ok && c.epoch == epoch {
		return c.byKey, nil
	}
	list, err := s.gateway.Projects(ctx, creds)
	if err != nil {
		return nil, err
	}
	byKey := make(map[string]backlog.Project, len(list))
	for _, p := range list {
		byKey[p.Key] = p
	}
	s.mu.Lock()
	s.projects[ws] = projectCache{epoch: epoch, byKey: byKey}
	s.mu.Unlock()
	return byKey, nil
}

func projectIDs(byKey map[string]backlog.Project, keys []string) []int64 {
	var ids []int64
	for _, k := range keys {
		if p, ok := byKey[k]; ok {
			ids = append(ids, p.ID)
		}
	}
	return ids
}

// TaskLink is one task linked to an issue row.
type TaskLink struct {
	TaskID  string `json:"taskId"`
	TaskKey string `json:"taskKey,omitempty"`
}

// IssueItem is one row of the issue list.
type IssueItem struct {
	IssueKey    string     `json:"issueKey"`
	Summary     string     `json:"summary"`
	Status      string     `json:"status"`
	StatusID    int64      `json:"statusId"`
	Assignee    string     `json:"assignee,omitempty"`
	UpdatedAt   string     `json:"updatedAt"`
	URL         string     `json:"url"`
	LinkedTasks []TaskLink `json:"linkedTasks"`
}

// IssuePage is the issues.list reply (C5).
type IssuePage struct {
	Items           []IssueItem `json:"items"`
	Total           int         `json:"total"`
	Page            int         `json:"page"`
	PageSize        int         `json:"pageSize"`
	RefreshedAt     string      `json:"refreshedAt"`
	ConnectionEpoch int         `json:"connectionEpoch"`
}

// List returns one page of the selected projects' issues, newest updated
// first, with the tasks linked to each (AC2.1.1, AC2.3.1). On page 1 an
// exact issue key typed as the keyword is added on top when it passes the
// filters (AC2.2.2).
func (s *Service) List(ctx context.Context, ws string, q Query) (IssuePage, error) {
	snap, err := s.conn.Current(ctx, ws)
	if err != nil {
		return IssuePage{}, err
	}
	q, err = ValidateQuery(q, snap.SelectedProjects)
	if err != nil {
		return IssuePage{}, err
	}
	ctx, creds, err := s.credentials(ctx, ws)
	if err != nil {
		return IssuePage{}, err
	}
	projects, err := s.projectsFor(ctx, ws, snap.ConnectionEpoch, creds)
	if err != nil {
		return IssuePage{}, err
	}
	bq := backlog.IssueQuery{ProjectIDs: projectIDs(projects, q.ProjectKeys), StatusIDs: q.StatusIDs, AssigneeIDs: q.AssigneeIDs,
		Keyword: q.Keyword, Offset: (q.Page - 1) * q.PageSize, Count: q.PageSize}
	if len(bq.ProjectIDs) == 0 {
		// Never send a query without a project: Backlog would search every project.
		return IssuePage{}, &connection.FieldError{Field: FieldProjectKeys, Err: ErrNoProject}
	}
	found, err := s.gateway.Issues(ctx, creds, backlog.Interactive, bq)
	if err != nil {
		return IssuePage{}, err
	}
	total, err := s.gateway.IssueCount(ctx, creds, backlog.Interactive, bq)
	if err != nil {
		return IssuePage{}, err
	}
	if q.Page == 1 {
		found = s.withExactKey(ctx, creds, q, bq, found)
	}
	links, err := s.store.Links(ctx, ws)
	if err != nil {
		return IssuePage{}, err
	}
	page := IssuePage{Items: make([]IssueItem, 0, len(found)), Total: total, Page: q.Page, PageSize: q.PageSize,
		RefreshedAt: s.now(), ConnectionEpoch: snap.ConnectionEpoch}
	for _, i := range found {
		item := IssueItem{IssueKey: i.IssueKey, Summary: i.Summary, Status: i.StatusName, StatusID: i.StatusID,
			Assignee: i.AssigneeName, UpdatedAt: i.Updated, URL: IssueURL(snap.SpaceHost, i.IssueKey), LinkedTasks: []TaskLink{}}
		for _, l := range links {
			if l.IssueKey == i.IssueKey && l.SpaceHost == snap.SpaceHost {
				item.LinkedTasks = append(item.LinkedTasks, TaskLink{TaskID: l.TaskID, TaskKey: l.TaskKey})
			}
		}
		page.Items = append(page.Items, item)
	}
	return page, nil
}

// withExactKey puts the issue whose key is the keyword on top of found. Any
// failure of that one extra read (an unknown key included) is ignored.
func (s *Service) withExactKey(ctx context.Context, creds backlog.Credentials, q Query, bq backlog.IssueQuery, found []backlog.Issue) []backlog.Issue {
	key := strings.ToUpper(q.Keyword)
	project, _, err := ParseIssueKey(key)
	if err != nil || !slices.Contains(q.ProjectKeys, project) ||
		slices.ContainsFunc(found, func(i backlog.Issue) bool { return i.IssueKey == key }) {
		return found
	}
	i, err := s.gateway.Issue(ctx, creds, backlog.Interactive, key)
	if err != nil || !Matches(i, bq) {
		return found
	}
	return append([]backlog.Issue{i}, found...)
}

// Option is one filter choice: a project (Key) or a status or user (ID).
type Option struct {
	ID   int64  `json:"id,omitempty"`
	Key  string `json:"key,omitempty"`
	Name string `json:"name"`
}

// Filters is the issues.filters reply.
type Filters struct {
	Projects  []Option `json:"projects"`
	Statuses  []Option `json:"statuses"`
	Assignees []Option `json:"assignees"`
}

// Filters returns the selected projects, their statuses (deduplicated by id)
// and the union of their members, with one call of each kind per project.
func (s *Service) Filters(ctx context.Context, ws string) (Filters, error) {
	snap, err := s.conn.Current(ctx, ws)
	if err != nil {
		return Filters{}, err
	}
	if len(snap.SelectedProjects) == 0 {
		return Filters{}, &connection.FieldError{Field: FieldProjectKeys, Err: ErrNoProject}
	}
	ctx, creds, err := s.credentials(ctx, ws)
	if err != nil {
		return Filters{}, err
	}
	projects, err := s.projectsFor(ctx, ws, snap.ConnectionEpoch, creds)
	if err != nil {
		return Filters{}, err
	}
	f := Filters{Projects: []Option{}, Statuses: []Option{}, Assignees: []Option{}}
	addOnce := func(list []Option, o Option) []Option {
		if slices.ContainsFunc(list, func(x Option) bool { return x.ID == o.ID }) {
			return list
		}
		return append(list, o)
	}
	for _, k := range snap.SelectedProjects {
		f.Projects = append(f.Projects, Option{Key: k, Name: projects[k].Name})
		statuses, err := s.gateway.ProjectStatuses(ctx, creds, k)
		if err != nil {
			return Filters{}, err
		}
		for _, st := range statuses {
			f.Statuses = addOnce(f.Statuses, Option{ID: st.ID, Name: st.Name})
		}
		users, err := s.gateway.ProjectUsers(ctx, creds, k)
		if err != nil {
			return Filters{}, err
		}
		for _, u := range users {
			f.Assignees = addOnce(f.Assignees, Option{ID: u.ID, Name: u.Name})
		}
	}
	return f, nil
}

// CreateInput is an issues.create_task request.
type CreateInput struct {
	IssueKey       string `json:"issueKey"`
	WorkflowID     string `json:"workflowId"`
	WorkflowStepID string `json:"workflowStepId"`
	// Force creates another task for an issue that already has one (AC3.1.3).
	Force bool `json:"force"`
}

// CreateResult is the issues.create_task reply.
type CreateResult struct {
	TaskID   string `json:"taskId"`
	TaskKey  string `json:"taskKey"`
	IssueKey string `json:"issueKey"`
}

// CreateTask creates a Kandev task from an issue and links them (AC3.1.1),
// with one Backlog read and one Kandev create. A second create of the same
// issue while one runs is a conflict (AC3.1.2), and so is an issue that
// already has a task unless Force is set (AC3.1.3). A failed read or create
// leaves nothing (AC3.1.4); a failed link write after the create is
// internal and the task stays unlinked.
// ponytail: no automatic relink of a task whose link write failed.
func (s *Service) CreateTask(ctx context.Context, ws string, in CreateInput) (CreateResult, error) {
	project, _, err := ParseIssueKey(in.IssueKey)
	if err != nil {
		return CreateResult{}, err
	}
	if strings.TrimSpace(in.WorkflowID) == "" {
		return CreateResult{}, invalid(FieldWorkflow)
	}
	unlock, ok := s.tryLock(ws + "/" + in.IssueKey)
	if !ok {
		return CreateResult{}, ErrConflict
	}
	defer unlock()
	snap, err := s.conn.Current(ctx, ws)
	if err != nil {
		return CreateResult{}, err
	}
	if !slices.Contains(snap.SelectedProjects, project) {
		return CreateResult{}, invalid(FieldIssueKey)
	}
	links, err := s.store.Links(ctx, ws)
	if err != nil {
		return CreateResult{}, err
	}
	if !in.Force && slices.ContainsFunc(links, func(l Link) bool {
		return l.IssueKey == in.IssueKey && l.SpaceHost == snap.SpaceHost && l.State == StateActive
	}) {
		return CreateResult{}, ErrConflict
	}
	ctx, creds, err := s.credentials(ctx, ws)
	if err != nil {
		return CreateResult{}, err
	}
	issue, err := s.issue(ctx, creds, backlog.Interactive, in.IssueKey)
	if err != nil {
		return CreateResult{}, err
	}
	if err := s.unchanged(ctx, ws, snap.ConnectionEpoch); err != nil {
		return CreateResult{}, err
	}
	ref, err := s.createLinkedTask(ctx, ws, snap, issue, in.WorkflowID, in.WorkflowStepID)
	if err != nil {
		return CreateResult{}, err
	}
	return CreateResult{TaskID: ref.ID, TaskKey: ref.Key, IssueKey: issue.IssueKey}, nil
}

// createLinkedTask creates the task of an issue (NewTaskFor) in the
// workflow and step, then stores its link. Used by issues.create_task and
// the issue watcher. A failed link write is logged and returns the created
// task with the error, so the caller knows the task exists.
func (s *Service) createLinkedTask(ctx context.Context, ws string, snap connection.Snapshot, issue backlog.Issue, workflowID, stepID string) (TaskRef, error) {
	task := NewTaskFor(issue, snap.SpaceHost)
	task.WorkspaceID, task.WorkflowID, task.WorkflowStepID = ws, workflowID, stepID
	ref, err := s.host.CreateTask(ctx, task)
	if err != nil {
		return TaskRef{}, fmt.Errorf("create task: %w", err)
	}
	if err := s.putLink(ctx, ws, s.newLink(snap, issue, ref)); err != nil {
		redact.Logger(ctx).ErrorContext(ctx, "issue link not stored", "event", "issue_link_write_failed",
			"taskId", ref.ID, "issueKey", issue.IssueKey)
		return ref, err
	}
	return ref, nil
}

func (s *Service) newLink(snap connection.Snapshot, issue backlog.Issue, task TaskRef) Link {
	now := s.now()
	project, _, _ := ParseIssueKey(issue.IssueKey)
	return Link{IssueKey: issue.IssueKey, IssueID: issue.ID, ProjectKey: project, SpaceHost: snap.SpaceHost,
		TaskID: task.ID, TaskKey: task.Key, State: StateActive, LastKnownStatus: issue.StatusName,
		StatusUpdatedAt: now, ConnectionEpoch: snap.ConnectionEpoch, CreatedAt: now}
}

func (s *Service) tryLock(key string) (func(), bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.creating[key] {
		return nil, false
	}
	s.creating[key] = true
	return func() {
		s.mu.Lock()
		delete(s.creating, key)
		s.mu.Unlock()
	}, true
}

// putLink adds l. A task links one issue only: a link of the task to
// another issue is a conflict (FR3.4); the same pair is kept as it is.
func (s *Service) putLink(ctx context.Context, ws string, l Link) error {
	return s.store.UpdateLinks(ctx, ws, func(links []Link) ([]Link, error) {
		i := slices.IndexFunc(links, func(x Link) bool { return x.TaskID == l.TaskID })
		switch {
		case i < 0:
			return append(links, l), nil
		case links[i].IssueKey != l.IssueKey:
			return nil, ErrConflict
		default:
			return nil, errUnchanged
		}
	})
}

// TaskItem is one task of the Link to task dialog (M3).
type TaskItem struct {
	TaskID         string `json:"taskId"`
	TaskKey        string `json:"taskKey,omitempty"`
	Title          string `json:"title"`
	LinkedIssueKey string `json:"linkedIssueKey,omitempty"`
}

// SearchTasks returns at most 20 workspace tasks whose title or key contains
// query, ignoring case, with the issue each is linked to.
func (s *Service) SearchTasks(ctx context.Context, ws, query string) ([]TaskItem, error) {
	tasks, err := s.host.ListTasks(ctx, ws)
	if err != nil {
		return nil, err
	}
	links, err := s.store.Links(ctx, ws)
	if err != nil {
		return nil, err
	}
	q := strings.ToLower(strings.TrimSpace(query))
	out := []TaskItem{}
	for _, t := range tasks {
		if len(out) == searchTaskRows {
			break
		}
		if q != "" && !strings.Contains(strings.ToLower(t.Title), q) && !strings.Contains(strings.ToLower(t.Key), q) {
			continue
		}
		item := TaskItem{TaskID: t.ID, TaskKey: t.Key, Title: t.Title}
		if l := byTaskID(links, t.ID); l != nil {
			item.LinkedIssueKey = l.IssueKey
		}
		out = append(out, item)
	}
	return out, nil
}

func byTaskID(links []Link, taskID string) *Link {
	i := slices.IndexFunc(links, func(l Link) bool { return l.TaskID == taskID })
	if i < 0 {
		return nil
	}
	return &links[i]
}

// Link links a task to an issue of a selected project, after one Backlog
// read checks it exists. Another task may link the same issue (AC3.3.2); a
// task already linked to another issue is a conflict (AC3.3.3); the same
// pair again is a no-op.
func (s *Service) Link(ctx context.Context, ws, taskID, issueKey string) (Link, error) {
	if taskID == "" {
		return Link{}, invalid(FieldTaskID)
	}
	project, _, err := ParseIssueKey(issueKey)
	if err != nil {
		return Link{}, err
	}
	snap, err := s.conn.Current(ctx, ws)
	if err != nil {
		return Link{}, err
	}
	if !slices.Contains(snap.SelectedProjects, project) {
		return Link{}, invalid(FieldIssueKey)
	}
	links, err := s.store.Links(ctx, ws)
	if err != nil {
		return Link{}, err
	}
	if l := byTaskID(links, taskID); l != nil {
		if l.IssueKey != issueKey {
			return Link{}, ErrConflict
		}
		return *l, nil
	}
	ctx, creds, err := s.credentials(ctx, ws)
	if err != nil {
		return Link{}, err
	}
	issue, err := s.issue(ctx, creds, backlog.Interactive, issueKey)
	if err != nil {
		return Link{}, err
	}
	if err := s.unchanged(ctx, ws, snap.ConnectionEpoch); err != nil {
		return Link{}, err
	}
	ref := TaskRef{ID: taskID, Key: s.taskKey(ctx, ws, taskID)}
	l := s.newLink(snap, issue, ref)
	return l, s.putLink(ctx, ws, l)
}

// taskKey finds the task's human key; empty when it cannot be read, as the
// next sync cycle fills it in.
// ponytail: one paged task list per link; cache the keys if linking gets frequent.
func (s *Service) taskKey(ctx context.Context, ws, taskID string) string {
	tasks, err := s.host.ListTasks(ctx, ws)
	if err != nil {
		return ""
	}
	for _, t := range tasks {
		if t.ID == taskID {
			return t.Key
		}
	}
	return ""
}

// Unlink removes the task's link. Backlog is not called (AC3.3.5).
func (s *Service) Unlink(ctx context.Context, ws, taskID string) error {
	return s.store.UpdateLinks(ctx, ws, func(links []Link) ([]Link, error) {
		n := len(links)
		links = slices.DeleteFunc(links, func(l Link) bool { return l.TaskID == taskID })
		if len(links) == n {
			return nil, ErrNotLinked
		}
		return links, nil
	})
}

// LinkView is one link for the card badges and the task menu (M6).
type LinkView struct {
	TaskID          string `json:"taskId"`
	TaskKey         string `json:"taskKey,omitempty"`
	IssueKey        string `json:"issueKey"`
	SpaceHost       string `json:"spaceHost"`
	State           string `json:"state"`
	Status          string `json:"status,omitempty"`
	StatusUpdatedAt string `json:"statusUpdatedAt,omitempty"`
	Stale           bool   `json:"stale"`
	Unavailable     bool   `json:"unavailable"`
	URL             string `json:"url"`
}

// Links returns every link of the workspace.
func (s *Service) Links(ctx context.Context, ws string) ([]LinkView, error) {
	links, err := s.store.Links(ctx, ws)
	if err != nil {
		return nil, err
	}
	out := make([]LinkView, 0, len(links))
	for _, l := range links {
		out = append(out, LinkView{TaskID: l.TaskID, TaskKey: l.TaskKey, IssueKey: l.IssueKey, SpaceHost: l.SpaceHost,
			State: l.State, Status: l.LastKnownStatus, StatusUpdatedAt: l.StatusUpdatedAt, Stale: l.Stale(),
			Unavailable: l.Unavailable, URL: IssueURL(l.SpaceHost, l.IssueKey)})
	}
	return out, nil
}

// IssueDetail is the read-only issue block of the task panel (M8).
type IssueDetail struct {
	Key      string `json:"key"`
	Summary  string `json:"summary"`
	Status   string `json:"status"`
	Assignee string `json:"assignee,omitempty"`
	Priority string `json:"priority,omitempty"`
	DueDate  string `json:"dueDate,omitempty"`
	URL      string `json:"url"`
}

// AttachmentView is one attachment; the file stays on Backlog.
type AttachmentView struct {
	Name     string `json:"name"`
	Size     int64  `json:"size"`
	TooLarge bool   `json:"tooLarge"`
}

// DetailView is the issues.get reply.
type DetailView struct {
	IssueKey         string           `json:"issueKey"`
	SpaceHost        string           `json:"spaceHost,omitempty"` // set when not connected, for "Reconnect <host>"
	LinkState        string           `json:"linkState"`
	StatusUpdatedAt  string           `json:"statusUpdatedAt,omitempty"`
	Issue            *IssueDetail     `json:"issue,omitempty"`
	Attachments      []AttachmentView `json:"attachments,omitempty"`
	AttachmentsError string           `json:"attachmentsError,omitempty"`
}

// Detail reads the task's issue live, with one issue and one attachments
// call (AC3.2.1, AC3.2.2). A not-connected link makes no call. An
// attachments failure still returns the issue.
func (s *Service) Detail(ctx context.Context, ws, taskID string) (DetailView, error) {
	l, err := s.linkOf(ctx, ws, taskID)
	if err != nil {
		return DetailView{}, err
	}
	if l.State != StateActive {
		return DetailView{IssueKey: l.IssueKey, SpaceHost: l.SpaceHost, LinkState: l.State}, nil
	}
	ctx, creds, err := s.credentials(ctx, ws)
	if err != nil {
		return DetailView{}, err
	}
	i, err := s.issue(ctx, creds, backlog.Interactive, l.IssueKey)
	if err != nil {
		return DetailView{}, err
	}
	d := DetailView{IssueKey: l.IssueKey, LinkState: l.State, StatusUpdatedAt: l.StatusUpdatedAt,
		Issue: &IssueDetail{Key: i.IssueKey, Summary: i.Summary, Status: i.StatusName, Assignee: i.AssigneeName,
			Priority: i.PriorityName, DueDate: i.DueDate, URL: IssueURL(l.SpaceHost, i.IssueKey)}}
	atts, err := s.gateway.IssueAttachments(ctx, creds, backlog.Interactive, l.IssueKey)
	if err != nil {
		d.AttachmentsError = connection.Classify(err).Code
		return d, nil
	}
	for _, a := range atts {
		d.Attachments = append(d.Attachments, AttachmentView{Name: a.Name, Size: a.Size, TooLarge: a.Size > previewLimit})
	}
	return d, nil
}

func (s *Service) linkOf(ctx context.Context, ws, taskID string) (Link, error) {
	links, err := s.store.Links(ctx, ws)
	if err != nil {
		return Link{}, err
	}
	if l := byTaskID(links, taskID); l != nil {
		return *l, nil
	}
	return Link{}, ErrNotLinked
}

// CommentView is one comment of the task panel.
type CommentView struct {
	ID      int64  `json:"id"`
	Author  string `json:"author"`
	Content string `json:"content"`
	Created string `json:"created"`
}

// CommentPage is the issues.comments reply: newest first, and NextMaxID
// while older comments remain.
type CommentPage struct {
	Comments  []CommentView `json:"comments"`
	NextMaxID int64         `json:"nextMaxId,omitempty"`
}

// Comments returns 20 comments of the task's issue, newest first, starting
// at maxID (0 = the newest) (AC3.5.1).
func (s *Service) Comments(ctx context.Context, ws, taskID string, maxID int64) (CommentPage, error) {
	l, err := s.linkOf(ctx, ws, taskID)
	if err != nil {
		return CommentPage{}, err
	}
	if l.State != StateActive {
		return CommentPage{}, connection.ErrNotConnected
	}
	ctx, creds, err := s.credentials(ctx, ws)
	if err != nil {
		return CommentPage{}, err
	}
	// One extra comment tells whether an older page exists.
	cs, err := s.gateway.IssueComments(ctx, creds, backlog.Interactive, l.IssueKey,
		backlog.CommentQuery{MaxID: max(maxID, 0), Count: commentsPerPage + 1})
	if err != nil {
		return CommentPage{}, err
	}
	page := CommentPage{Comments: make([]CommentView, 0, commentsPerPage)}
	if len(cs) > commentsPerPage {
		page.NextMaxID = cs[commentsPerPage].ID
		cs = cs[:commentsPerPage]
	}
	for _, c := range cs {
		page.Comments = append(page.Comments, CommentView{ID: c.ID, Author: c.AuthorName, Content: c.Content, Created: c.Created})
	}
	return page, nil
}

// Candidate is one `#` issue suggestion.
type Candidate struct {
	Key   string
	Title string
	URL   string
}

// Suggest returns at most limit (1-10, default 5) issues for a `#` query:
// the exact key first, then keyword matches (AC3.4.1). A newer query for the
// workspace within 250 ms replaces this one, which returns nothing
// (AC3.4.2). Every failure is an empty list, never an error (AC3.4.3), and
// the whole call stays inside 1.2 s.
func (s *Service) Suggest(ctx context.Context, ws, query string, limit int) ([]Candidate, error) {
	out := []Candidate{}
	if limit <= 0 {
		limit = suggestDefault
	}
	limit = min(limit, suggestMax)
	ctx, cancel := context.WithTimeout(ctx, suggestBudget)
	defer cancel()
	seq := s.nextSuggest(ws)
	if s.Wait(ctx, suggestSettle) != nil || !s.latestSuggest(ws, seq) {
		return out, nil
	}
	if s.conn.RequireEnabled(ctx, ws) != nil {
		return out, nil
	}
	snap, err := s.conn.Current(ctx, ws)
	if err != nil {
		return out, nil
	}
	ctx, creds, err := s.credentials(ctx, ws)
	if err != nil {
		return out, nil
	}
	q := strings.TrimSpace(query)
	add := func(i backlog.Issue) {
		if len(out) < limit && SuggestMatch(i, q) && !slices.ContainsFunc(out, func(c Candidate) bool { return c.Key == i.IssueKey }) {
			out = append(out, Candidate{Key: i.IssueKey, Title: i.Summary, URL: IssueURL(snap.SpaceHost, i.IssueKey)})
		}
	}
	key := strings.ToUpper(q)
	if project, _, err := ParseIssueKey(key); err == nil && slices.Contains(snap.SelectedProjects, project) {
		if i, err := s.gateway.Issue(ctx, creds, backlog.Interactive, key); err == nil {
			add(i)
		}
	}
	projects, err := s.projectsFor(ctx, ws, snap.ConnectionEpoch, creds)
	if err != nil {
		return out, nil
	}
	ids := projectIDs(projects, snap.SelectedProjects)
	if len(ids) == 0 {
		return out, nil
	}
	found, err := s.gateway.Issues(ctx, creds, backlog.Interactive, backlog.IssueQuery{ProjectIDs: ids, Keyword: q, Count: limit})
	if err != nil {
		redact.Logger(ctx).WarnContext(ctx, "issue suggestions failed", "event", "issue_suggest_failed",
			"errorCode", connection.Classify(err).Code)
		return out, nil
	}
	for _, i := range found {
		add(i)
	}
	return out, nil
}

func (s *Service) nextSuggest(ws string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.suggestSeq[ws]++
	return s.suggestSeq[ws]
}

func (s *Service) latestSuggest(ws string, seq int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.suggestSeq[ws] == seq
}

// Authorize allows a `#` reference only for an issue of a selected project
// of the connected space that Backlog returns live within 1.2 s. Every
// other case, an error or a timeout included, is a denial (fail closed).
func (s *Service) Authorize(ctx context.Context, ws, key string) bool {
	ctx, cancel := context.WithTimeout(ctx, suggestBudget)
	defer cancel()
	project, _, err := ParseIssueKey(key)
	if err != nil || s.conn.RequireEnabled(ctx, ws) != nil {
		return false
	}
	snap, err := s.conn.Current(ctx, ws)
	if err != nil || !slices.Contains(snap.SelectedProjects, project) {
		return false
	}
	ctx, creds, err := s.credentials(ctx, ws)
	if err != nil {
		return false
	}
	i, err := s.gateway.Issue(ctx, creds, backlog.Interactive, key)
	return err == nil && i.IssueKey == key
}

// Impact is the number of active issue links a disconnect, space change or
// project deselection turns off (AC1.8.1, AC1.9.1).
type Impact struct {
	IssueLinks int `json:"issueLinks"`
}

// Impact counts active links, of projectKeys only when given.
func (s *Service) Impact(ctx context.Context, ws string, projectKeys []string) (Impact, error) {
	links, err := s.store.Links(ctx, ws)
	if err != nil {
		return Impact{}, err
	}
	var out Impact
	for _, l := range links {
		if l.State == StateActive && (len(projectKeys) == 0 || slices.Contains(projectKeys, l.ProjectKey)) {
			out.IssueLinks++
		}
	}
	return out, nil
}

// Settings returns the workspace's sync interval and last cycle time.
func (s *Service) Settings(ctx context.Context, ws string) (Settings, error) {
	return s.store.Settings(ctx, ws)
}

// SetPollInterval stores a sync interval of 1 to 1440 minutes (AC4.2.2).
func (s *Service) SetPollInterval(ctx context.Context, ws string, raw any) (Settings, error) {
	minutes, err := ValidatePollMinutes(raw)
	if err != nil {
		return Settings{}, err
	}
	var out Settings
	err = s.store.UpdateSettings(ctx, ws, func(st Settings) (Settings, error) {
		st.PollMinutes = minutes
		out = st
		return st, nil
	})
	return out, err
}
