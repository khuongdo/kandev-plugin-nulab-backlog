package issues

import (
	"testing"
	"testing/synctest"

	"github.com/stretchr/testify/require"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/backlog"
	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/connection"
)

func createIn(key string) CreateInput {
	return CreateInput{IssueKey: key, WorkflowID: "wf-1", WorkflowStepID: "step-1"}
}

func TestU3_Create_MakesTheTaskAndLink(t *testing.T) {
	r := newRig(t)
	got, err := r.svc.CreateTask(r.ctx, "ws-1", createIn("PROJ-118"))
	require.NoError(t, err)
	require.Equal(t, CreateResult{TaskID: "task-19", TaskKey: "T-19", IssueKey: "PROJ-118"}, got)
	require.Equal(t, []NewTask{{WorkspaceID: "ws-1", WorkflowID: "wf-1", WorkflowStepID: "step-1", Title: "Fix login timeout",
		Description: "Steps in the attachment.\n\nBacklog: https://" + spaceHost + "/view/PROJ-118", Priority: "high"}}, r.host.creates)
	l := byTask(r.links(t), "task-19")
	require.Equal(t, Link{IssueKey: "PROJ-118", IssueID: 5118, ProjectKey: "PROJ", SpaceHost: spaceHost, TaskID: "task-19",
		TaskKey: "T-19", Summary: "Fix login timeout", State: StateActive, LastKnownStatus: "In Progress", StatusUpdatedAt: l.StatusUpdatedAt,
		ConnectionEpoch: 1, CreatedAt: l.CreatedAt}, l, "AC3.1.1")
	require.NotEmpty(t, l.CreatedAt)
	require.Equal(t, []backlog.CallClass{backlog.Interactive}, r.gw.classes["issue"])
}

func TestU3_Create_DoubleClickMakesOneTask(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		r := newRig(t)
		block := make(chan struct{})
		r.gw.block = block
		results := make(chan error, 2)
		for range 2 {
			go func() {
				_, err := r.svc.CreateTask(r.ctx, "ws-1", createIn("PROJ-118"))
				results <- err
			}()
		}
		synctest.Wait() // one call waits on Backlog holding the lock; the other is done
		require.ErrorIs(t, <-results, ErrConflict, "AC3.1.2: the second click is refused")
		close(block)
		require.NoError(t, <-results)
		require.Equal(t, 1, r.host.createCount())
	})
}

func TestU3_Create_LinkedIssueNeedsForce(t *testing.T) {
	r := newRig(t)
	r.link(t, Link{IssueKey: "PROJ-118", TaskID: "task-17", TaskKey: "T-17"})
	_, err := r.svc.CreateTask(r.ctx, "ws-1", createIn("PROJ-118"))
	require.ErrorIs(t, err, ErrConflict, "AC3.1.3")
	require.Zero(t, r.host.createCount())
	in := createIn("PROJ-118")
	in.Force = true
	got, err := r.svc.CreateTask(r.ctx, "ws-1", in)
	require.NoError(t, err)
	require.Equal(t, "T-19", got.TaskKey)
	require.Len(t, r.links(t), 2, "a second task for the same issue")
}

func TestU3_Create_FailuresLeaveNothing(t *testing.T) {
	cases := map[string]func(r *rig){
		"backlog 404":    func(r *rig) { r.gw.errs["issue:PROJ-118"] = notFound() },
		"backlog 429":    func(r *rig) { r.gw.errs["issue"] = rateLimited(5e9) },
		"kandev refuses": func(r *rig) { r.host.failNth = 1 },
	}
	for name, setup := range cases {
		t.Run(name, func(t *testing.T) {
			r := newRig(t)
			setup(r)
			_, err := r.svc.CreateTask(r.ctx, "ws-1", createIn("PROJ-118"))
			require.Error(t, err, "AC3.1.4")
			require.Empty(t, r.links(t))
			if name != "kandev refuses" {
				require.Zero(t, r.host.createCount())
			}
		})
	}
}

func TestU3_Create_LinkWriteFailureIsInternal(t *testing.T) {
	r := newRig(t)
	r.state.failSet["issues.links"] = true
	_, err := r.svc.CreateTask(r.ctx, "ws-1", createIn("PROJ-118"))
	require.ErrorIs(t, err, connection.ErrStore)
	require.Equal(t, 1, r.host.createCount(), "the task stays; the user can link it by hand")
	require.Contains(t, r.logs.String(), `"event":"issue_link_write_failed"`)
	require.Contains(t, r.logs.String(), `"taskId":"task-19"`)
}

func TestU3_Create_ValidatesAndDropsLateResults(t *testing.T) {
	r := newRig(t)
	for in, field := range map[CreateInput]string{
		{IssueKey: "bad", WorkflowID: "wf-1"}:      FieldIssueKey,
		{IssueKey: "OTHER-1", WorkflowID: "wf-1"}:  FieldIssueKey,
		{IssueKey: "PROJ-118"}:                     FieldWorkflow,
		{IssueKey: "PROJ-118", WorkflowID: " \t "}: FieldWorkflow,
	} {
		_, err := r.svc.CreateTask(r.ctx, "ws-1", in)
		require.Equal(t, field, fieldOf(t, err), "%+v", in)
	}
	require.Zero(t, r.gw.total())

	synctest.Test(t, func(t *testing.T) {
		r := newRig(t)
		block := make(chan struct{})
		r.gw.block = block
		done := make(chan error)
		go func() {
			_, err := r.svc.CreateTask(r.ctx, "ws-1", createIn("PROJ-118"))
			done <- err
		}()
		synctest.Wait()
		r.conn.set(func(c *fakeConn) { c.snap.ConnectionEpoch = 2 })
		close(block)
		require.ErrorIs(t, <-done, ErrStale, "AC1.8.3")
		require.Zero(t, r.host.createCount())
	})
}
