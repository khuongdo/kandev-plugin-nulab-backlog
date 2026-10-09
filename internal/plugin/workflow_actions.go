package plugin

import (
	"context"
	"maps"
)

// actionWorkflowsStatus tells the "+ Task" menu whether the workspace has a
// workflow when Kandev's browser context has none loaded (intent 261009, FR3).
const actionWorkflowsStatus = "workflows.status"

func init() {
	maps.Copy(handlers, map[string]handler{
		actionWorkflowsStatus: func(r *Runtime, ctx context.Context, ws string, _ []byte) (any, error) {
			has, err := r.ports.HasWorkflow(ctx, ws)
			return map[string]bool{"hasWorkflow": has}, err
		},
	})
}
