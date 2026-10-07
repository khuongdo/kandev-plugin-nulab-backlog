package git

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/backlog"
	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/connection"
	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/redact"
)

// Errors the plugin maps to action error codes.
var (
	// ErrNotFound is a watch, query or pull request that does not exist (not_found).
	ErrNotFound = errors.New("not found")
	// ErrPRNotFound is a pull request Backlog does not have (not_found).
	ErrPRNotFound = fmt.Errorf("pull request %w", ErrNotFound)
	// ErrConflict is an action the item's state does not allow (conflict).
	ErrConflict = errors.New("not allowed in the current state")
	// ErrStale means the connection changed while the action ran, so its
	// result was dropped (AC1.8.3).
	ErrStale = errors.New("the Backlog connection changed meanwhile")
	// ErrScopeRefused is any refused Git credential scope.
	ErrScopeRefused = errors.New("git credential scope refused")
)

// OpenPRExistsError is returned when an open pull request already uses the
// branch (AC5.3.4). The plugin answers conflict with pullRequestNumber.
type OpenPRExistsError struct{ Number int }

func (e *OpenPRExistsError) Error() string {
	return fmt.Sprintf("pull request #%d is already open for this branch", e.Number)
}

// Is makes an OpenPRExistsError match ErrConflict.
func (e *OpenPRExistsError) Is(target error) bool { return target == ErrConflict }

// RefusedError is a refused Git credential scope. Reason is a fixed word,
// never input or a secret.
type RefusedError struct{ Reason string }

func (e *RefusedError) Error() string { return "git credential refused: " + e.Reason }

// Unwrap makes a RefusedError match ErrScopeRefused.
func (e *RefusedError) Unwrap() error { return ErrScopeRefused }

// Gateway is the Backlog calls U4 makes. backlog.Client implements it.
type Gateway interface {
	Myself(ctx context.Context, creds backlog.Credentials) (backlog.User, error)
	Repositories(ctx context.Context, creds backlog.Credentials, projectKey string) ([]backlog.Repository, error)
	PullRequests(ctx context.Context, creds backlog.Credentials, class backlog.CallClass, projectKey, repo string, q backlog.PullRequestQuery) ([]backlog.PullRequest, error)
	PullRequestCount(ctx context.Context, creds backlog.Credentials, class backlog.CallClass, projectKey, repo string, q backlog.PullRequestQuery) (int, error)
	PullRequest(ctx context.Context, creds backlog.Credentials, class backlog.CallClass, projectKey, repo string, number int) (backlog.PullRequest, error)
	CreatePullRequest(ctx context.Context, creds backlog.Credentials, projectKey, repo string, in backlog.NewPullRequest) (backlog.PullRequest, error)
	Issue(ctx context.Context, creds backlog.Credentials, class backlog.CallClass, issueKey string) (backlog.Issue, error)
}

// Connection is the ConnectionReader U4 uses (C3). connection.Service implements it.
type Connection interface {
	Current(ctx context.Context, workspaceID string) (connection.Snapshot, error)
	Credentials(ctx context.Context, workspaceID string) (backlog.Credentials, int, error)
	GitCredential(ctx context.Context, workspaceID string) (connection.GitCredential, string, error)
	Subscribe(fn func(connection.ConnectionChanged)) (unsubscribe func())
	// RequireEnabled returns connection.ErrIntegrationDisabled while the
	// workspace's Backlog switch is off, or the store error (fail closed).
	RequireEnabled(ctx context.Context, workspaceID string) error
}

// Service is the GitIntegration component.
type Service struct {
	gateway Gateway
	conn    Connection
	host    HostPort
	store   *Store
	Now     func() time.Time // clock for leases and watch runs

	mu        sync.Mutex
	lastEpoch map[string]int // the last ConnectionChanged epoch handled per workspace
}

// NewService wires the GitIntegration component.
func NewService(gateway Gateway, conn Connection, host HostPort, store *Store) *Service {
	return &Service{gateway: gateway, conn: conn, host: host, store: store, Now: time.Now, lastEpoch: map[string]int{}}
}

const (
	repoPageSize = 50
	leaseTTL     = 15 * time.Minute
	statusPRs    = 100 // pull requests read for branches and the open-PR check
	queryRows    = 20
)

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

func isKind(err error, kind backlog.Kind) bool {
	var be *backlog.Error
	return errors.As(err, &be) && be.Kind == kind
}

func newID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b) // crypto/rand.Read never fails on supported platforms
	return hex.EncodeToString(b)
}

// RepositoryInspection is one repository for Kandev's repository picker (C8).
type RepositoryInspection struct {
	ProviderID     string `json:"providerId"`
	ProviderHost   string `json:"providerHost"`
	ProviderScope  string `json:"providerScope"`
	OwnerOrProject string `json:"ownerOrProject"`
	RepositoryID   string `json:"repositoryId"`
	RepositoryName string `json:"repositoryName"`
	CloneURL       string `json:"cloneUrl"`
}

// RepositoryPage is the git.repositories.list reply.
type RepositoryPage struct {
	Repositories []RepositoryInspection `json:"repositories"`
	NextCursor   string                 `json:"nextCursor,omitempty"`
}

type repoCursor struct {
	Query  string `json:"q"`
	Scope  string `json:"s"`
	Offset int    `json:"o"`
}

// ListRepositories lists the repositories of the selected projects whose
// name contains query (AC5.1.1, AC1.7.1). The cursor is bound to the query
// and the project scope; a mismatched one is refused before any request.
func (s *Service) ListRepositories(ctx context.Context, ws, query, cursor string) (RepositoryPage, error) {
	snap, err := s.conn.Current(ctx, ws)
	if err != nil {
		return RepositoryPage{}, err
	}
	q := strings.ToLower(strings.TrimSpace(query))
	scope := snap.SpaceHost + "|" + strings.Join(snap.SelectedProjects, ",")
	offset := 0
	if cursor != "" {
		var c repoCursor
		raw, err := base64.RawURLEncoding.DecodeString(cursor)
		if err != nil || json.Unmarshal(raw, &c) != nil || c.Query != q || c.Scope != scope || c.Offset < 0 {
			return RepositoryPage{}, invalid(FieldCursor)
		}
		offset = c.Offset
	}
	ctx, creds, err := s.credentials(ctx, ws)
	if err != nil {
		return RepositoryPage{}, err
	}
	all := []RepositoryInspection{}
	for _, project := range snap.SelectedProjects {
		repos, err := s.gateway.Repositories(ctx, creds, project)
		if err != nil {
			return RepositoryPage{}, err
		}
		for _, r := range repos {
			if q == "" || strings.Contains(strings.ToLower(r.Name), q) {
				all = append(all, RepositoryInspection{ProviderID: ProviderID, ProviderHost: "https://" + snap.SpaceHost,
					ProviderScope: snap.SpaceHost, OwnerOrProject: project, RepositoryID: strconv.FormatInt(r.ID, 10),
					RepositoryName: r.Name, CloneURL: r.HTTPURL})
			}
		}
	}
	offset = min(offset, len(all))
	end := min(offset+repoPageSize, len(all))
	page := RepositoryPage{Repositories: all[offset:end]}
	if end < len(all) {
		b, _ := json.Marshal(repoCursor{Query: q, Scope: scope, Offset: end})
		page.NextCursor = base64.RawURLEncoding.EncodeToString(b)
	}
	return page, nil
}

// Descriptor is the repositories.inspect repository, in Kandev's snake_case.
type Descriptor struct {
	ProviderID           string `json:"provider_id"`
	ProviderHost         string `json:"provider_host"`
	ProviderScope        string `json:"provider_scope"`
	ProviderRepositoryID string `json:"provider_repository_id"`
	OwnerOrProject       string `json:"owner_or_project"`
	Name                 string `json:"name"`
	CloneURL             string `json:"clone_url"`
	DefaultBranch        string `json:"default_branch"`
}

// Inspect resolves a Backlog clone or web URL of a selected project into a
// descriptor. nil means matched:false. Kandev requires a default branch and
// Backlog has none, so it is the most used base branch of the newest pull
// requests, else "master".
func (s *Service) Inspect(ctx context.Context, ws, raw string) (*Descriptor, error) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u.Scheme != "https" || u.User != nil {
		return nil, nil
	}
	snap, err := s.conn.Current(ctx, ws)
	if err != nil {
		return nil, err
	}
	project, name, ok := ParseClonePath(u.Path)
	if !ok || !strings.EqualFold(u.Host, snap.SpaceHost) || !slices.Contains(snap.SelectedProjects, project) {
		return nil, nil
	}
	ctx, creds, err := s.credentials(ctx, ws)
	if err != nil {
		return nil, err
	}
	repos, err := s.gateway.Repositories(ctx, creds, project)
	if err != nil {
		return nil, err
	}
	i := slices.IndexFunc(repos, func(r backlog.Repository) bool { return r.Name == name })
	if i < 0 {
		return nil, nil
	}
	prs, err := s.gateway.PullRequests(ctx, creds, backlog.Interactive, project, name, backlog.PullRequestQuery{Count: statusPRs})
	if err != nil {
		return nil, err
	}
	return &Descriptor{ProviderID: ProviderID, ProviderHost: "https://" + snap.SpaceHost, ProviderScope: snap.SpaceHost,
		ProviderRepositoryID: strconv.FormatInt(repos[i].ID, 10), OwnerOrProject: project, Name: name,
		CloneURL: repos[i].HTTPURL, DefaultBranch: DefaultBranch(prs)}, nil
}

// Branch is one repositories.branches entry.
type Branch struct {
	Name      string `json:"name"`
	IsDefault bool   `json:"is_default,omitempty"`
}

// Branches lists the branch names seen in the newest 100 pull requests;
// Backlog has no branch-list API.
func (s *Service) Branches(ctx context.Context, ws string, d Descriptor) ([]Branch, error) {
	snap, err := s.conn.Current(ctx, ws)
	if err != nil {
		return nil, err
	}
	if d.ProviderID != ProviderID || d.ProviderScope != snap.SpaceHost ||
		!slices.Contains(snap.SelectedProjects, d.OwnerOrProject) || !repoNameRe.MatchString(d.Name) {
		return nil, invalid(FieldRepository)
	}
	ctx, creds, err := s.credentials(ctx, ws)
	if err != nil {
		return nil, err
	}
	prs, err := s.gateway.PullRequests(ctx, creds, backlog.Interactive, d.OwnerOrProject, d.Name, backlog.PullRequestQuery{Count: statusPRs})
	if err != nil {
		return nil, err
	}
	out := []Branch{}
	if len(prs) == 0 {
		return out, nil
	}
	def := DefaultBranch(prs)
	for _, name := range BranchCandidates(prs) {
		out = append(out, Branch{Name: name, IsDefault: name == def})
	}
	return out, nil
}

// Link resolves a pull request reference with one Backlog call and links it
// to the task (AC5.2.1). taskRepoID is the verified Kandev repository of the
// action, used by the short reference forms. Linking twice is a no-op.
func (s *Service) Link(ctx context.Context, ws, taskID, reference, taskRepoID string) (Link, error) {
	if taskID == "" {
		return Link{}, invalid(FieldReference)
	}
	snap, err := s.conn.Current(ctx, ws)
	if err != nil {
		return Link{}, err
	}
	var def *RepoRef
	if taskRepoID != "" {
		kr, err := s.host.Repository(ctx, ws, taskRepoID)
		switch {
		case err == nil && kr.ProviderID == ProviderID:
			def = &RepoRef{ProjectKey: kr.OwnerOrProject, Name: kr.Name}
		case err != nil && !errors.Is(err, ErrRepositoryNotFound):
			return Link{}, err
		}
	}
	ref, err := ParseReference(reference, snap.SpaceHost, def)
	if err != nil {
		return Link{}, err
	}
	if !slices.Contains(snap.SelectedProjects, ref.ProjectKey) {
		return Link{}, &connection.FieldError{Field: FieldReference, Err: errReference}
	}
	ctx, creds, err := s.credentials(ctx, ws)
	if err != nil {
		return Link{}, err
	}
	pr, err := s.gateway.PullRequest(ctx, creds, backlog.Interactive, ref.ProjectKey, ref.Repo, ref.Number)
	if isKind(err, backlog.KindNotFound) {
		return Link{}, ErrPRNotFound
	}
	if err != nil {
		return Link{}, err
	}
	if err := s.unchanged(ctx, ws, snap.ConnectionEpoch); err != nil {
		return Link{}, err
	}
	l := Link{TaskID: taskID, SpaceHost: snap.SpaceHost, ProjectKey: ref.ProjectKey, RepoName: ref.Repo,
		RepositoryID: pr.RepositoryID, Number: pr.Number, Title: pr.Summary, Status: StatusActive}
	return l, s.putLink(ctx, ws, l)
}

// putLink adds l, or replaces the same task's link to the same pull request.
func (s *Service) putLink(ctx context.Context, ws string, l Link) error {
	return s.store.UpdateLinks(ctx, ws, func(links []Link) ([]Link, error) {
		i := slices.IndexFunc(links, func(x Link) bool { return x.TaskID == l.TaskID && x.Key() == l.Key() })
		if i >= 0 {
			links[i] = l
			return links, nil
		}
		return append(links, l), nil
	})
}

// Unlink removes the task's link with reviewKey. Backlog is not called (AC5.2.3).
func (s *Service) Unlink(ctx context.Context, ws, taskID, reviewKey string) error {
	return s.store.UpdateLinks(ctx, ws, func(links []Link) ([]Link, error) {
		n := len(links)
		links = slices.DeleteFunc(links, func(l Link) bool { return l.TaskID == taskID && l.Key() == reviewKey })
		if len(links) == n {
			return nil, errUnchanged
		}
		return links, nil
	})
}

// Association is one ReviewTaskAssociation.
type Association struct {
	ProviderID          string `json:"providerId"`
	TaskID              string `json:"taskId"`
	ReviewKey           string `json:"reviewKey"`
	ConnectionScope     string `json:"connectionScope"`
	RepositoryID        string `json:"repositoryId"`
	ChangeRequestNumber int    `json:"changeRequestNumber"`
}

// Associations lists the workspace's active links.
func (s *Service) Associations(ctx context.Context, ws string) ([]Association, error) {
	links, err := s.store.Links(ctx, ws)
	if err != nil {
		return nil, err
	}
	out := []Association{}
	for _, l := range links {
		if l.Status == StatusActive {
			out = append(out, Association{ProviderID: ProviderID, TaskID: l.TaskID, ReviewKey: l.Key(),
				ConnectionScope: l.SpaceHost, RepositoryID: strconv.FormatInt(l.RepositoryID, 10), ChangeRequestNumber: l.Number})
		}
	}
	return out, nil
}

// StatusBadge is the label Kandev shows on the card.
type StatusBadge struct {
	Label string `json:"label"`
	Tone  string `json:"tone,omitempty"`
}

// TaskStatus is the ReviewTaskStatus of a summary. Backlog has no pipeline
// or check data, so pipelineState is neutral and checks is empty.
type TaskStatus struct {
	Number        int    `json:"number"`
	State         string `json:"state"`
	PipelineState string `json:"pipelineState"`
	Checks        []any  `json:"checks"`
	Error         string `json:"error,omitempty"`
}

// ReviewSummary is one ReviewSummary for Kandev's review provider, plus the
// branches and assignee the review panel shows.
type ReviewSummary struct {
	ProviderID          string       `json:"providerId"`
	ReviewKey           string       `json:"reviewKey"`
	Title               string       `json:"title"`
	URL                 string       `json:"url"`
	ConnectionScope     string       `json:"connectionScope"`
	RepositoryID        string       `json:"repositoryId"`
	ChangeRequestNumber int          `json:"changeRequestNumber"`
	State               string       `json:"state"`
	StatusBadge         *StatusBadge `json:"statusBadge,omitempty"`
	TaskStatus          *TaskStatus  `json:"taskStatus,omitempty"`
	Base                string       `json:"base,omitempty"`
	Branch              string       `json:"branch,omitempty"`
	Assignee            string       `json:"assignee,omitempty"`
}

var stateLabels = map[string]string{"open": "Open", "closed": "Closed", "merged": "Merged"}

// Status returns one summary per link of the task (AC5.4.1). A failed fetch
// is "Status unknown" and never hides the other links (AC5.4.3).
func (s *Service) Status(ctx context.Context, ws, taskID string) ([]ReviewSummary, error) {
	links, err := s.store.Links(ctx, ws)
	if err != nil {
		return nil, err
	}
	out := []ReviewSummary{}
	snap, connErr := s.conn.Current(ctx, ws)
	var (
		creds    backlog.Credentials
		credsErr error
		haveCred bool
	)
	for _, l := range links {
		if l.TaskID != taskID {
			continue
		}
		sum := ReviewSummary{ProviderID: ProviderID, ReviewKey: l.Key(), Title: l.Title,
			URL: prURL(l.SpaceHost, l.ProjectKey, l.RepoName, l.Number), ConnectionScope: l.SpaceHost,
			RepositoryID: strconv.FormatInt(l.RepositoryID, 10), ChangeRequestNumber: l.Number, State: "unknown"}
		unknown := func(label, code string) {
			sum.StatusBadge = &StatusBadge{Label: label}
			sum.TaskStatus = &TaskStatus{Number: l.Number, State: "open", PipelineState: "neutral", Checks: []any{}, Error: code}
			out = append(out, sum)
		}
		if l.Status != StatusActive || connErr != nil || l.SpaceHost != snap.SpaceHost {
			unknown("Not connected", StatusNotConnected)
			continue
		}
		if !haveCred {
			ctx, creds, credsErr = s.credentials(ctx, ws)
			haveCred = true
		}
		var pr backlog.PullRequest
		err := credsErr
		if err == nil {
			pr, err = s.gateway.PullRequest(ctx, creds, backlog.Interactive, l.ProjectKey, l.RepoName, l.Number)
		}
		if err != nil {
			unknown("Status unknown", errorCode(err))
			continue
		}
		state := backlog.PullRequestState(pr.StatusID)
		label, ok := stateLabels[state]
		if !ok {
			label, state = "Open", "open"
		}
		if pr.AssigneeName != "" {
			label += " – " + pr.AssigneeName
		}
		sum.Title, sum.State, sum.Base, sum.Branch, sum.Assignee = pr.Summary, state, pr.Base, pr.Branch, pr.AssigneeName
		sum.StatusBadge = &StatusBadge{Label: label}
		sum.TaskStatus = &TaskStatus{Number: pr.Number, State: state, PipelineState: "neutral", Checks: []any{}}
		out = append(out, sum)
	}
	return out, nil
}

// errorCode is the action error code of err, with a Backlog 404 as not_found.
func errorCode(err error) string {
	if isKind(err, backlog.KindNotFound) || errors.Is(err, ErrNotFound) {
		return "not_found"
	}
	return connection.Classify(err).Code
}

// CreateInput is a git.prs.create request. RepositoryID and HeadBranch come
// from the verified action context, never from the browser body.
type CreateInput struct {
	TaskID       string
	RepositoryID string // Kandev repository id
	HeadBranch   string
	Title        string
	Body         string
	BaseBranch   string
}

// CreateResult is the git.prs.create reply.
type CreateResult struct {
	URL    string `json:"url"`
	Number int    `json:"number"`
	Linked bool   `json:"linked"`
}

// CreatePR creates a pull request from the task's pushed branch and links it
// (AC5.3.2). An open pull request on the branch is a conflict and nothing is
// created (AC5.3.4). If the link cannot be stored, the reply says
// linked:false and the pull request is never created again.
func (s *Service) CreatePR(ctx context.Context, ws string, in CreateInput) (CreateResult, error) {
	in.Title, in.HeadBranch = strings.TrimSpace(in.Title), strings.TrimSpace(in.HeadBranch)
	switch {
	case in.Title == "":
		return CreateResult{}, invalid(FieldTitle)
	case in.HeadBranch == "":
		return CreateResult{}, invalid(FieldBranch) // AC5.3.3: nothing was pushed
	}
	kr, err := s.host.Repository(ctx, ws, in.RepositoryID)
	if errors.Is(err, ErrRepositoryNotFound) || (err == nil && kr.ProviderID != ProviderID) {
		return CreateResult{}, invalid(FieldRepository)
	}
	if err != nil {
		return CreateResult{}, err
	}
	snap, err := s.conn.Current(ctx, ws)
	if err != nil {
		return CreateResult{}, err
	}
	if kr.ProviderScope != snap.SpaceHost || !slices.Contains(snap.SelectedProjects, kr.OwnerOrProject) {
		return CreateResult{}, invalid(FieldRepository)
	}
	base := strings.TrimSpace(in.BaseBranch)
	if base == "" {
		base = kr.DefaultBranch
	}
	if base == "" {
		return CreateResult{}, invalid(FieldBaseBranch)
	}
	project, repo := kr.OwnerOrProject, kr.Name
	ctx, creds, err := s.credentials(ctx, ws)
	if err != nil {
		return CreateResult{}, err
	}
	open, err := s.gateway.PullRequests(ctx, creds, backlog.Interactive, project, repo, backlog.PullRequestQuery{StatusIDs: []int64{1}, Count: statusPRs})
	if err != nil {
		return CreateResult{}, err
	}
	for _, pr := range open {
		if pr.Branch == in.HeadBranch {
			return CreateResult{}, &OpenPRExistsError{Number: pr.Number}
		}
	}
	key := RelatedIssueKey(in.Title, in.Body, snap.SelectedProjects)
	var issueID int64
	if key != "" {
		issue, err := s.gateway.Issue(ctx, creds, backlog.Interactive, key)
		switch {
		case err == nil:
			issueID = issue.ID
		case !isKind(err, backlog.KindNotFound):
			return CreateResult{}, err
		}
	}
	pr, err := s.gateway.CreatePullRequest(ctx, creds, project, repo, backlog.NewPullRequest{
		Summary: in.Title, Description: PRDescription(in.Body, in.Title, key), Base: base, Branch: in.HeadBranch, IssueID: issueID,
	})
	if err != nil {
		return CreateResult{}, err
	}
	res := CreateResult{URL: prURL(snap.SpaceHost, project, repo, pr.Number), Number: pr.Number}
	l := Link{TaskID: in.TaskID, SpaceHost: snap.SpaceHost, ProjectKey: project, RepoName: repo,
		RepositoryID: pr.RepositoryID, Number: pr.Number, Title: pr.Summary, Status: StatusActive}
	err = s.unchanged(ctx, ws, snap.ConnectionEpoch)
	if err == nil {
		err = s.putLink(ctx, ws, l)
	}
	if err != nil {
		redact.Logger(ctx).WarnContext(ctx, "pull request not linked", "event", "pr_link_failed", "errorCode", errorCode(err))
		return res, nil
	}
	res.Linked = true
	return res, nil
}

// Scope is the exact Git credential lease scope Kandev asks for.
type Scope struct {
	ProviderID   string
	WorkspaceID  string
	TaskID       string
	SessionID    string
	RepositoryID string
	Host         string
	Path         string
}

// Lease is a transient Git credential.
type Lease struct {
	Username  string
	Secret    string
	ExpiresAt time.Time
}

// ResolveCredential returns the Git credential for a fetch or push of a
// selected project's repository on the connected space (AC5.6.1). Every
// refusal is a RefusedError with a fixed reason; no credential is
// ErrNoGitCredential, which says to update Git access in the settings (AC5.6.3).
func (s *Service) ResolveCredential(ctx context.Context, sc Scope) (Lease, error) {
	switch {
	case sc.ProviderID != ProviderID:
		return Lease{}, &RefusedError{Reason: "provider"}
	case strings.TrimSpace(sc.TaskID) == "":
		return Lease{}, &RefusedError{Reason: "task"}
	case strings.TrimSpace(sc.SessionID) == "":
		return Lease{}, &RefusedError{Reason: "session"}
	case strings.TrimSpace(sc.RepositoryID) == "":
		return Lease{}, &RefusedError{Reason: "repository"}
	}
	snap, err := s.conn.Current(ctx, sc.WorkspaceID)
	if errors.Is(err, connection.ErrNotConnected) || errors.Is(err, connection.ErrReconnectRequired) {
		return Lease{}, &RefusedError{Reason: "not_connected"}
	}
	if err != nil {
		return Lease{}, err
	}
	if !strings.EqualFold(sc.Host, snap.SpaceHost) {
		return Lease{}, &RefusedError{Reason: "host"}
	}
	project, _, ok := ParseClonePath(sc.Path)
	if !ok {
		return Lease{}, &RefusedError{Reason: "path"}
	}
	if !slices.Contains(snap.SelectedProjects, project) {
		return Lease{}, &RefusedError{Reason: "project"}
	}
	cred, _, err := s.conn.GitCredential(ctx, sc.WorkspaceID)
	if err != nil {
		return Lease{}, err
	}
	return Lease{Username: cred.Username, Secret: cred.Password, ExpiresAt: s.Now().Add(leaseTTL)}, nil
}

// Binding is the non-secret revision of the scope's credential; empty when
// there is none, which revokes Kandev's leases.
func (s *Service) Binding(ctx context.Context, sc Scope) (string, error) {
	if sc.ProviderID != ProviderID {
		return "", nil
	}
	_, binding, err := s.conn.GitCredential(ctx, sc.WorkspaceID)
	if errors.Is(err, connection.ErrNotConnected) || errors.Is(err, connection.ErrReconnectRequired) ||
		errors.Is(err, connection.ErrNoGitCredential) {
		return "", nil
	}
	return binding, err
}

// Impact is the number of active PR links and watches a disconnect, space
// change or project deselection turns off (AC1.8.1, AC1.9.1).
type Impact struct {
	PRLinks   int `json:"prLinks"`
	PRWatches int `json:"prWatches"`
}

// Impact counts active items, of projectKeys only when given.
func (s *Service) Impact(ctx context.Context, ws string, projectKeys []string) (Impact, error) {
	links, err := s.store.Links(ctx, ws)
	if err != nil {
		return Impact{}, err
	}
	watches, err := s.store.Watches(ctx, ws)
	if err != nil {
		return Impact{}, err
	}
	in := func(p string) bool { return len(projectKeys) == 0 || slices.Contains(projectKeys, p) }
	var out Impact
	for _, l := range links {
		if l.Status == StatusActive && in(l.ProjectKey) {
			out.PRLinks++
		}
	}
	for _, w := range watches {
		if w.State != StatusNotConnected && in(w.ProjectKey) {
			out.PRWatches++
		}
	}
	return out, nil
}

// ListWatches returns the workspace's watches.
func (s *Service) ListWatches(ctx context.Context, ws string) ([]Watch, error) {
	watches, err := s.store.Watches(ctx, ws)
	if watches == nil && err == nil {
		watches = []Watch{}
	}
	return watches, err
}

// SaveWatch creates (no id) or edits a watch (AC6.1.1). "me" becomes the
// connected user's id with one Myself call and a linked issue key its id with
// one Issue call. An edit keeps the state, counts and ledger.
func (s *Service) SaveWatch(ctx context.Context, ws string, in WatchInput) (Watch, error) {
	snap, err := s.conn.Current(ctx, ws)
	if err != nil {
		return Watch{}, err
	}
	if err := in.Validate(snap.SelectedProjects); err != nil {
		return Watch{}, err
	}
	if strings.TrimSpace(in.WorkflowID) == "" {
		return Watch{}, invalid(FieldWorkflow)
	}
	ctx, creds, err := s.credentials(ctx, ws)
	if err != nil {
		return Watch{}, err
	}
	w := Watch{ID: in.ID, Name: in.Name, SpaceHost: snap.SpaceHost, ProjectKey: in.ProjectKey, RepoName: in.RepoName,
		Statuses: in.Statuses, Assignee: in.Assignee, Creator: in.Creator, IssueKey: in.IssueKey,
		WorkflowID: in.WorkflowID, WorkflowStepID: in.WorkflowStepID, State: StatusActive}
	if in.Assignee == WhoMe || in.Creator == WhoMe {
		me, err := s.gateway.Myself(ctx, creds)
		if err != nil {
			return Watch{}, err
		}
		if in.Assignee == WhoMe {
			w.AssigneeID = me.ID
		}
		if in.Creator == WhoMe {
			w.CreatedUserID = me.ID
		}
	}
	if in.IssueKey != "" {
		issue, err := s.gateway.Issue(ctx, creds, backlog.Interactive, in.IssueKey)
		if isKind(err, backlog.KindNotFound) {
			return Watch{}, invalid(FieldIssueKey)
		}
		if err != nil {
			return Watch{}, err
		}
		w.IssueID = issue.ID
	}
	err = s.store.UpdateWatches(ctx, ws, func(list []Watch) ([]Watch, error) {
		if w.ID == "" {
			w.ID = newID()
			return append(list, w), nil
		}
		i := slices.IndexFunc(list, func(x Watch) bool { return x.ID == w.ID })
		if i < 0 {
			return nil, ErrNotFound
		}
		w.State, w.CreatedCount, w.PendingCount, w.LastRunAt = list[i].State, list[i].CreatedCount, list[i].PendingCount, list[i].LastRunAt
		list[i] = w
		return list, nil
	})
	return w, err
}

// DeleteWatch removes a watch; its ledger and created tasks stay (AC6.1.5).
func (s *Service) DeleteWatch(ctx context.Context, ws, id string) error {
	return s.store.UpdateWatches(ctx, ws, func(list []Watch) ([]Watch, error) {
		n := len(list)
		list = slices.DeleteFunc(list, func(w Watch) bool { return w.ID == id })
		if len(list) == n {
			return nil, ErrNotFound
		}
		return list, nil
	})
}

// PauseWatch stops a watch from running (AC6.1.3).
func (s *Service) PauseWatch(ctx context.Context, ws, id string) (Watch, error) {
	return s.setWatchState(ctx, ws, id, StatusPaused)
}

// ResumeWatch lets a paused watch run again. A not-connected watch cannot
// be resumed (conflict).
func (s *Service) ResumeWatch(ctx context.Context, ws, id string) (Watch, error) {
	return s.setWatchState(ctx, ws, id, StatusActive)
}

func (s *Service) setWatchState(ctx context.Context, ws, id, state string) (Watch, error) {
	var out Watch
	err := s.store.UpdateWatches(ctx, ws, func(list []Watch) ([]Watch, error) {
		i := slices.IndexFunc(list, func(w Watch) bool { return w.ID == id })
		switch {
		case i < 0:
			return nil, ErrNotFound
		case list[i].State == StatusNotConnected:
			return nil, ErrConflict
		}
		list[i].State = state
		out = list[i]
		return list, nil
	})
	return out, err
}

func (s *Service) findWatch(ctx context.Context, ws, id string) (Watch, error) {
	list, err := s.store.Watches(ctx, ws)
	if err != nil {
		return Watch{}, err
	}
	i := slices.IndexFunc(list, func(w Watch) bool { return w.ID == id })
	if i < 0 {
		return Watch{}, ErrNotFound
	}
	return list[i], nil
}

// ListQueries returns the saved queries.
func (s *Service) ListQueries(ctx context.Context, ws string) ([]Query, error) {
	qs, err := s.store.Queries(ctx, ws)
	if qs == nil && err == nil {
		qs = []Query{}
	}
	return qs, err
}

// SaveQuery creates (no id) or replaces a saved query.
func (s *Service) SaveQuery(ctx context.Context, ws string, in QueryInput) (Query, error) {
	snap, err := s.conn.Current(ctx, ws)
	if err != nil {
		return Query{}, err
	}
	if err := in.Validate(snap.SelectedProjects); err != nil {
		return Query{}, err
	}
	err = s.store.UpdateQueries(ctx, ws, func(list []Query) ([]Query, error) {
		if in.ID == "" {
			in.ID = newID()
			return append(list, in), nil
		}
		i := slices.IndexFunc(list, func(q Query) bool { return q.ID == in.ID })
		if i < 0 {
			return nil, ErrNotFound
		}
		list[i] = in
		return list, nil
	})
	return in, err
}

// DeleteQuery removes a saved query.
func (s *Service) DeleteQuery(ctx context.Context, ws, id string) error {
	return s.store.UpdateQueries(ctx, ws, func(list []Query) ([]Query, error) {
		n := len(list)
		list = slices.DeleteFunc(list, func(q Query) bool { return q.ID == id })
		if len(list) == n {
			return nil, ErrNotFound
		}
		return list, nil
	})
}

// QueryRow is one pull request of a query run.
type QueryRow struct {
	Number        int      `json:"number"`
	Title         string   `json:"title"`
	State         string   `json:"state"`
	Assignee      string   `json:"assignee,omitempty"`
	RepoName      string   `json:"repoName"`
	URL           string   `json:"url"`
	LinkedTaskIDs []string `json:"linkedTaskIds"`
}

// RunQuery returns at most 20 pull requests with their linked tasks (AC6.3.1).
func (s *Service) RunQuery(ctx context.Context, ws, id string) ([]QueryRow, error) {
	qs, err := s.store.Queries(ctx, ws)
	if err != nil {
		return nil, err
	}
	i := slices.IndexFunc(qs, func(q Query) bool { return q.ID == id })
	if i < 0 {
		return nil, ErrNotFound
	}
	q := qs[i]
	snap, err := s.conn.Current(ctx, ws)
	if err != nil {
		return nil, err
	}
	ctx, creds, err := s.credentials(ctx, ws)
	if err != nil {
		return nil, err
	}
	pq := backlog.PullRequestQuery{StatusIDs: statusIDs(q.Statuses), Count: queryRows}
	if q.Assignee == WhoMe {
		me, err := s.gateway.Myself(ctx, creds)
		if err != nil {
			return nil, err
		}
		pq.AssigneeIDs = []int64{me.ID}
	}
	prs, err := s.gateway.PullRequests(ctx, creds, backlog.Interactive, q.ProjectKey, q.RepoName, pq)
	if err != nil {
		return nil, err
	}
	links, err := s.store.Links(ctx, ws)
	if err != nil {
		return nil, err
	}
	rows := []QueryRow{}
	for _, pr := range prs[:min(len(prs), queryRows)] {
		key := LinkKey(snap.SpaceHost, pr.RepositoryID, pr.Number)
		linked := []string{}
		for _, l := range links {
			if l.Status == StatusActive && l.Key() == key {
				linked = append(linked, l.TaskID)
			}
		}
		rows = append(rows, QueryRow{Number: pr.Number, Title: pr.Summary, State: backlog.PullRequestState(pr.StatusID),
			Assignee: pr.AssigneeName, RepoName: q.RepoName, URL: prURL(snap.SpaceHost, q.ProjectKey, q.RepoName, pr.Number),
			LinkedTaskIDs: linked})
	}
	return rows, nil
}

func statusIDs(statuses []string) []int64 {
	out := make([]int64, 0, len(statuses))
	for _, st := range statuses {
		out = append(out, stateIDs[st])
	}
	return out
}
