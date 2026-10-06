package git

import (
	"context"
	"errors"
)

// MetadataKey is the task metadata key a watch task carries; its value is
// the pull request's LinkKey, so a lost ledger write can be recovered.
const MetadataKey = "nulab_backlog_pr"

// ErrRepositoryNotFound is returned by HostPort.Repository for an unknown id.
var ErrRepositoryNotFound = errors.New("kandev repository not found")

// NewTask is a Kandev task a watch creates.
type NewTask struct {
	WorkspaceID    string
	WorkflowID     string
	WorkflowStepID string
	Title          string
	Description    string
	Metadata       map[string]any
}

// KandevRepository is a repository registered in a Kandev workspace.
type KandevRepository struct {
	ID                   string
	ProviderID           string
	ProviderRepositoryID string
	ProviderHost         string
	ProviderScope        string
	OwnerOrProject       string
	Name                 string
	DefaultBranch        string
}

// HostPort is the Kandev host API U4 needs (C2). internal/plugin implements
// it on pluginsdk.Host; tests swap in a fake.
type HostPort interface {
	CreateTask(ctx context.Context, in NewTask) (taskID string, err error)
	FindTaskByMetadata(ctx context.Context, workspaceID, key, value string) (taskID string, found bool, err error)
	Repository(ctx context.Context, workspaceID, repositoryID string) (KandevRepository, error)
}
