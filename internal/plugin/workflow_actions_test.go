package plugin

import (
	"context"
	"errors"
	"testing"

	"github.com/kandev/kandev/pkg/pluginsdk"
	"github.com/stretchr/testify/require"
)

// Intent 261009 (FR3.2, FR3.3): the Host workflow read behind workflows.status.
func TestHostPort_HasWorkflow(t *testing.T) {
	cases := []struct {
		name      string
		workflows []pluginsdk.Workflow
		listErr   error
		want      bool
	}{
		{name: "one workflow means true", workflows: []pluginsdk.Workflow{{ID: "wf-1"}}, want: true},
		{name: "several workflows still read one", workflows: []pluginsdk.Workflow{{ID: "wf-1"}, {ID: "wf-2"}}, want: true},
		{name: "no workflow means false"},
		{name: "a host list error is wrapped", listErr: errors.New("rpc error: code = Unavailable")},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data := &hostData{workflows: tc.workflows, workflowsErr: tc.listErr}
			h := &u4Host{fakeHost: newFakeHost(), data: data}
			got, err := hostPort{host: func() pluginsdk.Host { return h }}.HasWorkflow(context.Background(), "ws-1")
			if tc.listErr != nil {
				require.ErrorIs(t, err, tc.listErr)
				require.ErrorContains(t, err, "list workflows")
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.want, got)
			}
			require.Equal(t, []workflowCall{{ws: "ws-1", page: pluginsdk.Page{Limit: 1}}}, data.workflowCalls, "one host call, page size 1 (NFR2)")
		})
	}
}

func TestHostPort_HasWorkflow_FailsWhenTheHostNeverArrives(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // the bounded wait ends at once, no real sleep
	p := hostPort{host: func() pluginsdk.Host { return nil }, ready: make(chan struct{})}
	_, err := p.HasWorkflow(ctx, "ws-1")
	require.ErrorIs(t, err, errNoHost)
}

func TestWorkflowsStatus_Action(t *testing.T) {
	for _, tc := range []struct {
		name      string
		workflows []pluginsdk.Workflow
		want      bool
	}{
		{name: "workspace with a workflow", workflows: []pluginsdk.Workflow{{ID: "wf-1", WorkspaceID: "ws-1"}}, want: true},
		{name: "workspace without a workflow", want: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := newU3Rig(t)
			r.data.workflows = tc.workflows
			resp, out := r.call(t, actionWorkflowsStatus, map[string]any{})
			require.Equal(t, 200, resp.Status)
			require.Equal(t, map[string]any{"hasWorkflow": tc.want}, out, "a boolean only, no workflow data")
			require.Equal(t, "ws-1", r.data.workflowCalls[0].ws, "the workspace comes from the verified context")
		})
	}
}

// NFR4: a host failure maps to an error code and never carries host text.
func TestWorkflowsStatus_HostErrorLeaksNoText(t *testing.T) {
	r := newU3Rig(t)
	r.data.workflowsErr = errors.New("rpc error: code = Unavailable desc = host-secret-detail")
	resp, out := r.call(t, actionWorkflowsStatus, nil)
	require.Equal(t, 500, resp.Status)
	require.Equal(t, "internal", out["error"].(map[string]any)["code"])
	require.NotContains(t, string(resp.Body), "host-secret-detail")
}

func TestWorkflowsStatus_Manifest(t *testing.T) {
	m := loadManifest(t)
	require.Equal(t, []any{"tasks", "repositories", "workflows"}, m.Capabilities["api_read"])
	var found bool
	for _, a := range m.Actions {
		if a.Key == actionWorkflowsStatus {
			found = true
			require.Equal(t, [3]any{"workspace", "authenticated", 8192}, [3]any{a.Scope, a.Access, a.MaxBodyBytes})
		}
	}
	require.True(t, found, "workflows.status is declared")
}
