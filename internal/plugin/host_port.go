package plugin

import (
	"context"
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/kandev/kandev/pkg/pluginsdk"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/git"
	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/issues"
)

// hostPort adapts the injected Host's data API to git.HostPort (C2). It
// needs api_read: [tasks, repositories] and api_write: [tasks].
type hostPort struct {
	host func() pluginsdk.Host
}

const hostPageSize = 100

// metadataNamespace is where Kandev v0.96.0 keeps this plugin's task
// metadata: Tasks().Create nests it under "plugin:<plugin id>" next to
// "source" (pluginTaskMetadata), and List returns it nested.
const metadataNamespace = "plugin:nulab-backlog"

func (p hostPort) get() (pluginsdk.Host, error) {
	if h := p.host(); h != nil {
		return h, nil
	}
	return nil, errNoHost
}

// CreateTask creates a task; an empty step means the workflow's default.
func (p hostPort) CreateTask(ctx context.Context, in git.NewTask) (string, error) {
	t, err := p.create(ctx, pluginsdk.CreateTaskInput{WorkspaceID: in.WorkspaceID, WorkflowID: in.WorkflowID,
		WorkflowStepID: optional(in.WorkflowStepID), Title: in.Title, Description: in.Description, Metadata: in.Metadata})
	if err != nil {
		return "", err
	}
	return t.ID, nil
}

// FindTaskByMetadata scans the workspace's tasks, archived ones included.
// ponytail: a paged scan per reserved ledger entry; fine for rare crash recovery.
func (p hostPort) FindTaskByMetadata(ctx context.Context, ws, key, value string) (string, bool, error) {
	var id string
	err := p.eachTask(ctx, pluginsdk.TaskFilter{WorkspaceIDs: []string{ws}, IncludeArchived: true}, func(t pluginsdk.Task) bool {
		ns, _ := t.Metadata[metadataNamespace].(map[string]any)
		if v, _ := ns[key].(string); v == value {
			id = t.ID
			return false
		}
		return true
	})
	return id, id != "", err
}

// create is the one Tasks().Create call of both adapters (api_write: [tasks]).
func (p hostPort) create(ctx context.Context, in pluginsdk.CreateTaskInput) (*pluginsdk.Task, error) {
	h, err := p.get()
	if err != nil {
		return nil, err
	}
	t, err := h.Tasks().Create(ctx, in)
	if err != nil {
		return nil, fmt.Errorf("create task: %w", err)
	}
	return t, nil
}

// eachTask pages through the tasks the filter matches until fn returns false.
func (p hostPort) eachTask(ctx context.Context, filter pluginsdk.TaskFilter, fn func(pluginsdk.Task) bool) error {
	h, err := p.get()
	if err != nil {
		return err
	}
	page := pluginsdk.Page{Limit: hostPageSize}
	for {
		tasks, info, err := h.Tasks().List(ctx, filter, page)
		if err != nil {
			return fmt.Errorf("list tasks: %w", err)
		}
		for _, t := range tasks {
			if !fn(t) {
				return nil
			}
		}
		if info == nil || !info.HasMore || info.NextCursor == "" {
			return nil
		}
		page.Cursor = info.NextCursor
	}
}

// optional is nil for an empty string: Kandev then uses the default.
func optional(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// issueHost adapts the same Host data API to issues.HostPort (C2) for U3.
// It writes no task metadata and no repository.
type issueHost struct{ hostPort }

// CreateTask creates a task from an issue and returns its id and key.
func (p issueHost) CreateTask(ctx context.Context, in issues.NewTask) (issues.TaskRef, error) {
	t, err := p.create(ctx, pluginsdk.CreateTaskInput{WorkspaceID: in.WorkspaceID, WorkflowID: in.WorkflowID,
		WorkflowStepID: optional(in.WorkflowStepID), Title: in.Title, Description: in.Description, Priority: in.Priority})
	if err != nil {
		return issues.TaskRef{}, err
	}
	return issues.TaskRef{ID: t.ID, Key: t.Identifier}, nil
}

// ListTasks lists the workspace's tasks, archived and ephemeral ones included.
// ponytail: a full paged scan; fine for one per sync cycle and per link.
func (p issueHost) ListTasks(ctx context.Context, ws string) ([]issues.TaskInfo, error) {
	var out []issues.TaskInfo
	f := pluginsdk.TaskFilter{WorkspaceIDs: []string{ws}, IncludeArchived: true, IncludeEphemeral: true}
	err := p.eachTask(ctx, f, func(t pluginsdk.Task) bool {
		out = append(out, issues.TaskInfo{ID: t.ID, Key: t.Identifier, Title: t.Title})
		return true
	})
	return out, err
}

// Repository finds a workspace repository by id.
func (p hostPort) Repository(ctx context.Context, ws, id string) (git.KandevRepository, error) {
	h, err := p.get()
	if err != nil {
		return git.KandevRepository{}, err
	}
	page := pluginsdk.Page{Limit: hostPageSize}
	for {
		repos, info, err := h.Repositories().List(ctx, ws, page)
		if err != nil {
			return git.KandevRepository{}, fmt.Errorf("list repositories: %w", err)
		}
		for _, r := range repos {
			if r.ID == id {
				out := git.KandevRepository{ID: r.ID, ProviderID: r.ProviderID, ProviderRepositoryID: r.ProviderRepositoryID,
					ProviderHost: r.ProviderHost, ProviderScope: r.ProviderScope, OwnerOrProject: r.OwnerOrProject, Name: repoName(r)}
				if out.ProviderID == git.ProviderID && !git.ValidRepoName(out.Name) {
					return git.KandevRepository{}, fmt.Errorf("repository has no valid Backlog name: %w", git.ErrRepositoryNotFound)
				}
				if r.DefaultBranch != nil {
					out.DefaultBranch = *r.DefaultBranch
				}
				return out, nil
			}
		}
		if info == nil || !info.HasMore || info.NextCursor == "" {
			return git.KandevRepository{}, git.ErrRepositoryNotFound
		}
		page.Cursor = info.NextCursor
	}
}

// repoName is the Backlog repository name. Kandev sets Name to
// "<owner>/<providerName>" for provider repositories, so it is never used:
// ProviderName first, else the clone URL's last path segment.
func repoName(r pluginsdk.Repository) string {
	if r.ProviderName != "" {
		return r.ProviderName
	}
	u, err := url.Parse(r.RemoteURL)
	if err != nil || u.Path == "" {
		return ""
	}
	return strings.TrimSuffix(path.Base(u.Path), ".git")
}
