package plugin

import (
	"context"
	"testing"

	"github.com/kandev/kandev/pkg/pluginsdk"
	"github.com/stretchr/testify/require"
)

// taskDeleted is shaped like Kandev v0.96.0's task.deleted delivery: the
// workspace on the event, the task id in the payload.
func taskDeleted(taskID string) *pluginsdk.Event {
	return &pluginsdk.Event{EventID: "ev-1", EventType: "task.deleted", OccurredAt: "2026-10-06T00:00:00Z", WorkspaceID: "ws-1",
		Payload: map[string]any{"task_id": taskID, "workspace_id": "ws-1", "title": "Login work"}}
}

func TestU3_Events_TaskDeletedRemovesLinks(t *testing.T) {
	r := newU3Rig(t)
	resp, _ := r.onTask(t, actionIssuesLink, "task-17", map[string]any{"issueKey": "PROJ-120"})
	require.Equal(t, 200, resp.Status)
	require.NoError(t, r.rt.OnEvent(context.Background(), taskDeleted("task-17")))
	_, out := r.call(t, actionIssuesLinks, nil)
	require.Empty(t, out["links"], "R-05")
	require.NoError(t, r.rt.OnEvent(context.Background(), taskDeleted("task-17")), "idempotent")
}

func TestU3_Events_OtherEventsAreIgnored(t *testing.T) {
	r := newU3Rig(t)
	resp, _ := r.onTask(t, actionIssuesLink, "task-17", map[string]any{"issueKey": "PROJ-120"})
	require.Equal(t, 200, resp.Status)
	for _, e := range []*pluginsdk.Event{
		{EventType: "task.updated", WorkspaceID: "ws-1", Payload: map[string]any{"task_id": "task-17"}},
		{EventType: "task.deleted", WorkspaceID: "ws-1", Payload: map[string]any{}},
		{EventType: "task.deleted", Payload: map[string]any{"task_id": "task-17"}},
	} {
		require.NoError(t, r.rt.OnEvent(context.Background(), e))
	}
	_, out := r.call(t, actionIssuesLinks, nil)
	require.Len(t, out["links"], 1)
}

func TestU3_Events_StoreFailureIsRetried(t *testing.T) {
	r := newU3Rig(t)
	r.host.mu.Lock()
	r.host.failRead = true
	r.host.mu.Unlock()
	require.Error(t, r.rt.OnEvent(context.Background(), taskDeleted("task-17")), "an error makes Kandev retry")
}
