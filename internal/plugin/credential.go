package plugin

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/kandev/kandev/pkg/pluginsdk"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/connection"
	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/git"
	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/redact"
)

var _ pluginsdk.GitCredentialHandler = (*Runtime)(nil)

// ResolveGitCredential hands Kandev the Backlog Git credential for a fetch
// or push of a selected project's repository (US5.6). Every refusal logs
// git_credential_refused with a fixed reason and returns an error with no
// secret; gRPC carries it to Kandev.
func (r *Runtime) ResolveGitCredential(ctx context.Context, req *pluginsdk.ResolveGitCredentialRequest) (*pluginsdk.ResolveGitCredentialResponse, error) {
	ctx, log := r.rpcContext(ctx, req.WorkspaceID)
	sc := git.Scope{ProviderID: req.ProviderID, WorkspaceID: req.WorkspaceID, TaskID: req.TaskID,
		SessionID: req.SessionID, RepositoryID: req.RepositoryID, Host: req.Host, Path: req.Path}
	if err := r.service.RequireEnabled(ctx, req.WorkspaceID); err != nil {
		return nil, refused(ctx, log, err)
	}
	lease, err := r.git.ResolveCredential(ctx, sc)
	if err != nil {
		return nil, refused(ctx, log, err)
	}
	return &pluginsdk.ResolveGitCredentialResponse{Username: lease.Username, Secret: lease.Secret,
		ExpiresAt: lease.ExpiresAt.UTC().Format(time.RFC3339)}, nil
}

// GetGitCredentialBinding returns the credential's non-secret revision, or
// empty (revoked) when there is no usable credential.
func (r *Runtime) GetGitCredentialBinding(ctx context.Context, req *pluginsdk.GitCredentialBindingRequest) (*pluginsdk.GitCredentialBindingResponse, error) {
	ctx, log := r.rpcContext(ctx, req.WorkspaceID)
	err := r.service.RequireEnabled(ctx, req.WorkspaceID)
	if errors.Is(err, connection.ErrIntegrationDisabled) {
		return &pluginsdk.GitCredentialBindingResponse{}, nil
	}
	if err != nil {
		return nil, refused(ctx, log, err)
	}
	b, err := r.git.Binding(ctx, git.Scope{ProviderID: req.ProviderID, WorkspaceID: req.WorkspaceID, TaskID: req.TaskID,
		SessionID: req.SessionID, RepositoryID: req.RepositoryID, Host: req.Host, Path: req.Path})
	if err != nil {
		return nil, refused(ctx, log, err)
	}
	return &pluginsdk.GitCredentialBindingResponse{Binding: b}, nil
}

func (r *Runtime) rpcContext(ctx context.Context, ws string) (context.Context, *slog.Logger) {
	log := r.log.With("workspaceId", ws, "requestId", newRequestID())
	return redact.WithLogger(ctx, log), log
}

// refused logs the refusal and returns an error whose text holds only a fixed
// reason. No Git access keeps its own text, which says where to fix it.
func refused(ctx context.Context, log *slog.Logger, err error) error {
	var re *git.RefusedError
	reason := connection.Classify(err).Code
	switch {
	case errors.As(err, &re):
		reason = re.Reason
	case errors.Is(err, connection.ErrNoGitCredential):
		reason = "no_git_credential"
	}
	log.WarnContext(ctx, "git credential refused", "event", "git_credential_refused", "reason", reason)
	if errors.Is(err, connection.ErrNoGitCredential) {
		return connection.ErrNoGitCredential
	}
	return fmt.Errorf("nulab-backlog: git credential refused (%s)", reason)
}
