package plugin

import (
	"context"

	"github.com/kandev/kandev/pkg/pluginsdk"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/redact"
)

const eventTaskDeleted = "task.deleted"

// OnEvent removes the links of a deleted task (R-05, C8). It is the only
// removal: the sync cycle keeps links of tasks missing from the task list
// (review 1, R-02) and only counts them. A returned error makes Kandev retry after 5, 15 and 45 s; other events and
// deliveries without a task or workspace are accepted and ignored.
func (r *Runtime) OnEvent(ctx context.Context, e *pluginsdk.Event) error {
	if e == nil || e.EventType != eventTaskDeleted {
		return nil
	}
	taskID, _ := e.Payload["task_id"].(string)
	if taskID == "" || e.WorkspaceID == "" {
		return nil
	}
	ctx = redact.WithLogger(ctx, r.log.With("workspaceId", e.WorkspaceID))
	return r.issues.OnTaskDeleted(ctx, e.WorkspaceID, taskID)
}
