package issues

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestU3_Impact_CountsActiveLinks(t *testing.T) {
	r := newRig(t)
	r.link(t, Link{IssueKey: "PROJ-118", TaskID: "task-17"})
	r.link(t, Link{IssueKey: "PROJ-120", TaskID: "task-18"})
	r.link(t, Link{IssueKey: "DEMO-1", TaskID: "task-19"})
	r.link(t, Link{IssueKey: "DEMO-2", TaskID: "task-20", State: StateNotConnected})
	got, err := r.svc.Impact(r.ctx, "ws-1", nil)
	require.NoError(t, err)
	require.Equal(t, Impact{IssueLinks: 3}, got, "AC1.8.1")
	got, err = r.svc.Impact(r.ctx, "ws-1", []string{"DEMO"})
	require.NoError(t, err)
	require.Equal(t, Impact{IssueLinks: 1}, got, "AC1.9.1")
}

func TestU3_Impact_StoreErrorIsReturned(t *testing.T) {
	r := newRig(t)
	r.state.failGet = true
	_, err := r.svc.Impact(r.ctx, "ws-1", nil)
	require.Error(t, err)
}
