package issues

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/backlog"
	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/connection"
)

func detailRig(t *testing.T) *rig {
	r := newRig(t)
	r.link(t, Link{IssueKey: "PROJ-118", TaskID: "task-17", TaskKey: "T-17", StatusUpdatedAt: "2026-10-06T00:00:00Z"})
	r.gw.attachments = []backlog.Attachment{{ID: 8, Name: "spec.pdf", Size: 1258291}, {ID: 9, Name: "dump.zip", Size: 50331648}}
	return r
}

func TestU3_Detail_ShowsTheIssue(t *testing.T) {
	r := detailRig(t)
	d, err := r.svc.Detail(r.ctx, "ws-1", "task-17")
	require.NoError(t, err)
	require.Equal(t, DetailView{IssueKey: "PROJ-118", LinkState: StateActive, StatusUpdatedAt: "2026-10-06T00:00:00Z",
		Issue: &IssueDetail{Key: "PROJ-118", Summary: "Fix login timeout", Status: "In Progress", Assignee: "Lan",
			Priority: "High", DueDate: "2026-10-10T00:00:00Z", URL: "https://" + spaceHost + "/view/PROJ-118"},
		Attachments: []AttachmentView{{Name: "spec.pdf", Size: 1258291}, {Name: "dump.zip", Size: 50331648, TooLarge: true}}}, d,
		"AC3.2.1, AC3.5.2")
	require.Equal(t, 1, r.gw.count("issue"))
	require.Equal(t, 1, r.gw.count("attachments"))
	require.Equal(t, []backlog.CallClass{backlog.Interactive}, r.gw.classes["attachments"])
}

func TestU3_Detail_ReadsLiveEveryTime(t *testing.T) {
	r := detailRig(t)
	_, err := r.svc.Detail(r.ctx, "ws-1", "task-17")
	require.NoError(t, err)
	r.gw.set(func(g *fakeGateway) { g.issues[0].DueDate = "2026-12-24T00:00:00Z" })
	d, err := r.svc.Detail(r.ctx, "ws-1", "task-17")
	require.NoError(t, err)
	require.Equal(t, "2026-12-24T00:00:00Z", d.Issue.DueDate, "AC3.2.2")
	require.Zero(t, r.host.createCount(), "the plugin never writes the task")
}

func TestU3_Detail_UnavailableIsNotFound(t *testing.T) {
	for name, err := range map[string]error{"404": notFound(), "403": forbidden()} {
		t.Run(name, func(t *testing.T) {
			r := detailRig(t)
			r.gw.errs["issue:PROJ-118"] = err
			_, got := r.svc.Detail(r.ctx, "ws-1", "task-17")
			require.ErrorIs(t, got, ErrIssueUnavailable)
			require.ErrorIs(t, got, ErrNotFound, "answered not_found, never reconnect_required")
		})
	}
}

func TestU3_Detail_NotConnectedMakesNoCall(t *testing.T) {
	r := newRig(t)
	r.link(t, Link{IssueKey: "PROJ-118", TaskID: "task-17", State: StateNotConnected, SpaceHost: "old.backlog.jp"})
	d, err := r.svc.Detail(r.ctx, "ws-1", "task-17")
	require.NoError(t, err)
	require.Equal(t, DetailView{IssueKey: "PROJ-118", LinkState: StateNotConnected, SpaceHost: "old.backlog.jp"}, d)
	require.Zero(t, r.gw.total())
	_, err = r.svc.Detail(r.ctx, "ws-1", "task-99")
	require.ErrorIs(t, err, ErrNotFound, "no link")
}

func TestU3_Detail_AttachmentErrorKeepsIssue(t *testing.T) {
	r := detailRig(t)
	r.gw.errs["attachments"] = rateLimited(5e9)
	d, err := r.svc.Detail(r.ctx, "ws-1", "task-17")
	require.NoError(t, err)
	require.NotNil(t, d.Issue)
	require.Equal(t, "rate_limited", d.AttachmentsError)
	require.Empty(t, d.Attachments)
}

func TestU3_Comments_PagesNewestFirst(t *testing.T) {
	r := detailRig(t)
	for id := int64(150); id >= 1; id-- {
		r.gw.comments = append(r.gw.comments, backlog.Comment{ID: id, Content: "c", AuthorName: "Lan"})
	}
	pages, maxID := 0, int64(0)
	var seen []int64
	for {
		p, err := r.svc.Comments(r.ctx, "ws-1", "task-17", maxID)
		require.NoError(t, err)
		pages++
		for _, c := range p.Comments {
			seen = append(seen, c.ID)
		}
		if p.NextMaxID == 0 {
			break
		}
		maxID = p.NextMaxID
	}
	require.Equal(t, 8, pages, "AC3.5.1: 150 comments, 20 per page")
	require.Len(t, seen, 150)
	require.Equal(t, int64(150), seen[0], "newest first")
	require.Equal(t, int64(1), seen[149])
}

func TestU3_Comments_KeepErrorCodes(t *testing.T) {
	r := detailRig(t)
	r.gw.errs["comments"] = &backlog.Error{Kind: backlog.KindUnreachable, Status: 500}
	_, err := r.svc.Comments(r.ctx, "ws-1", "task-17", 0)
	require.Equal(t, "unreachable", connection.Classify(err).Code, "AC3.5.3")
	_, err = r.svc.Comments(r.ctx, "ws-1", "task-99", 0)
	require.ErrorIs(t, err, ErrNotFound)
}
