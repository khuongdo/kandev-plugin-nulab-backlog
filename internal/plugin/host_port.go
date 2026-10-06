package plugin

import (
	"context"
	"fmt"

	"github.com/kandev/kandev/pkg/pluginsdk"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/git"
)

// hostPort adapts the injected Host's data API to git.HostPort (C2). It
// needs api_read: [tasks, repositories] and api_write: [tasks].
type hostPort struct {
	host func() pluginsdk.Host
}

const hostPageSize = 100

func (p hostPort) get() (pluginsdk.Host, error) {
	if h := p.host(); h != nil {
		return h, nil
	}
	return nil, errNoHost
}

// CreateTask creates a task; an empty step means the workflow's default.
func (p hostPort) CreateTask(ctx context.Context, in git.NewTask) (string, error) {
	h, err := p.get()
	if err != nil {
		return "", err
	}
	var step *string
	if in.WorkflowStepID != "" {
		step = &in.WorkflowStepID
	}
	t, err := h.Tasks().Create(ctx, pluginsdk.CreateTaskInput{WorkspaceID: in.WorkspaceID, WorkflowID: in.WorkflowID,
		WorkflowStepID: step, Title: in.Title, Description: in.Description, Metadata: in.Metadata})
	if err != nil {
		return "", fmt.Errorf("create task: %w", err)
	}
	return t.ID, nil
}

// FindTaskByMetadata scans the workspace's tasks, archived ones included.
// ponytail: a paged scan per reserved ledger entry; fine for rare crash recovery.
func (p hostPort) FindTaskByMetadata(ctx context.Context, ws, key, value string) (string, bool, error) {
	h, err := p.get()
	if err != nil {
		return "", false, err
	}
	page := pluginsdk.Page{Limit: hostPageSize}
	for {
		tasks, info, err := h.Tasks().List(ctx, pluginsdk.TaskFilter{WorkspaceIDs: []string{ws}, IncludeArchived: true}, page)
		if err != nil {
			return "", false, fmt.Errorf("list tasks: %w", err)
		}
		for _, t := range tasks {
			if v, _ := t.Metadata[key].(string); v == value {
				return t.ID, true, nil
			}
		}
		if info == nil || !info.HasMore || info.NextCursor == "" {
			return "", false, nil
		}
		page.Cursor = info.NextCursor
	}
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
					ProviderHost: r.ProviderHost, ProviderScope: r.ProviderScope, OwnerOrProject: r.OwnerOrProject, Name: r.Name}
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
