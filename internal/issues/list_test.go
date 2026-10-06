package issues

import (
	"fmt"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/backlog"
	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/connection"
)

// fiftySeven replaces the fake's issues with PROJ-1..PROJ-57, newest first.
func fiftySeven(g *fakeGateway) {
	g.issues = nil
	for n := 57; n >= 1; n-- {
		g.issues = append(g.issues, backlog.Issue{ID: int64(9000 + n), ProjectID: 101, IssueKey: fmt.Sprintf("PROJ-%d", n),
			Summary: fmt.Sprintf("Issue %d", n), StatusID: int64(1 + n%3), StatusName: "Open", AssigneeID: int64(n % 2)})
	}
}

func TestU3_List_ReturnsAPageWithLinks(t *testing.T) {
	r := newRig(t)
	r.link(t, Link{IssueKey: "PROJ-120", TaskID: "task-17", TaskKey: "T-17"})
	r.link(t, Link{IssueKey: "PROJ-120", TaskID: "task-15", TaskKey: "T-15"})
	page, err := r.svc.List(r.ctx, "ws-1", Query{})
	require.NoError(t, err)
	require.Equal(t, 4, page.Total)
	require.Equal(t, 1, page.ConnectionEpoch)
	require.NotEmpty(t, page.RefreshedAt)
	require.Len(t, page.Items, 4)
	require.Equal(t, IssueItem{IssueKey: "PROJ-118", Summary: "Fix login timeout", Status: "In Progress", StatusID: 2,
		Assignee: "Lan", UpdatedAt: "2026-10-01T09:00:00Z", URL: "https://" + spaceHost + "/view/PROJ-118",
		LinkedTasks: []TaskLink{}}, page.Items[0])
	require.Equal(t, []TaskLink{{TaskID: "task-17", TaskKey: "T-17"}, {TaskID: "task-15", TaskKey: "T-15"}}, page.Items[1].LinkedTasks, "AC2.3.1")
	require.Equal(t, []int64{101, 102}, r.gw.queries[0].ProjectIDs, "AC1.7.1: only the selected projects")
	require.Equal(t, []backlog.CallClass{backlog.Interactive}, r.gw.classes["issues"])
	require.Equal(t, []backlog.CallClass{backlog.Interactive}, r.gw.classes["count"])
}

func TestU3_List_PagesAndCachesProjectIDs(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		r := newRig(t)
		r.gw.set(fiftySeven)
		page, err := r.svc.List(r.ctx, "ws-1", Query{Page: 3})
		require.NoError(t, err)
		require.Equal(t, 57, page.Total)
		require.Len(t, page.Items, 17)
		require.Equal(t, "PROJ-17", page.Items[0].IssueKey)
		require.Equal(t, 40, r.gw.queries[0].Offset)
		require.Equal(t, 20, r.gw.queries[0].Count)
		_, err = r.svc.List(r.ctx, "ws-1", Query{Page: 1})
		require.NoError(t, err)
		require.Equal(t, 1, r.gw.count("projects"), "project ids are cached per workspace and epoch")
		r.conn.set(func(c *fakeConn) { c.snap.ConnectionEpoch = 2 })
		_, err = r.svc.List(r.ctx, "ws-1", Query{Page: 1})
		require.NoError(t, err)
		require.Equal(t, 2, r.gw.count("projects"), "a new epoch reloads them")
	})
}

func TestU3_List_FiltersEveryRow(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		r := newRig(t)
		r.gw.set(fiftySeven)
		cases := []Query{
			{StatusIDs: []int64{2}},
			{AssigneeIDs: []int64{1}},
			{StatusIDs: []int64{1, 3}, AssigneeIDs: []int64{0}},
			{ProjectKeys: []string{"DEMO"}},
		}
		for _, q := range cases {
			page, err := r.svc.List(r.ctx, "ws-1", q)
			require.NoError(t, err)
			bq := r.gw.queries[len(r.gw.queries)-1]
			for _, it := range page.Items {
				var issue backlog.Issue
				for _, i := range r.gw.issues {
					if i.IssueKey == it.IssueKey {
						issue = i
					}
				}
				require.True(t, Matches(issue, bq), "AC2.2.1: %s passes %+v", it.IssueKey, q)
			}
		}
		require.Equal(t, []int64{102}, r.gw.queries[3].ProjectIDs)
	})
}

func TestU3_List_ExactKeyGoesOnTop(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		r := newRig(t)
		r.gw.set(fiftySeven)
		page, err := r.svc.List(r.ctx, "ws-1", Query{Keyword: "proj-5"})
		require.NoError(t, err)
		require.Equal(t, "PROJ-5", page.Items[0].IssueKey, "AC2.2.2: the exact key row first")
		require.Equal(t, 1, r.gw.count("issue"))

		_, err = r.svc.List(r.ctx, "ws-1", Query{Keyword: "PROJ-5", StatusIDs: []int64{1}})
		require.NoError(t, err)
		page, _ = r.svc.List(r.ctx, "ws-1", Query{Keyword: "PROJ-5", StatusIDs: []int64{1}})
		for _, it := range page.Items {
			require.NotEqual(t, "PROJ-5", it.IssueKey, "PROJ-5 has status 3 and does not pass the filter")
		}
		page, err = r.svc.List(r.ctx, "ws-1", Query{Keyword: "PROJ-5", Page: 2})
		require.NoError(t, err)
		require.Equal(t, 3, r.gw.count("issue"), "page 2 makes no exact-key call")
		_ = page

		page, err = r.svc.List(r.ctx, "ws-1", Query{Keyword: "PROJ-999"})
		require.NoError(t, err, "AC2.2.3: an unknown key is an empty page")
		require.Empty(t, page.Items)
	})
}

func TestU3_List_NeedsProjectsAndAConnection(t *testing.T) {
	r := newRig(t)
	r.conn.set(func(c *fakeConn) { c.snap.SelectedProjects = nil })
	_, err := r.svc.List(r.ctx, "ws-1", Query{})
	require.ErrorIs(t, err, ErrNoProject)
	require.Equal(t, connection.Outcome{Code: "validation", Field: FieldProjectKeys}, connection.Classify(err), "AC1.7.2")

	r.conn.set(func(c *fakeConn) { c.currentErr = connection.ErrNotConnected })
	_, err = r.svc.List(r.ctx, "ws-1", Query{})
	require.Equal(t, "reconnect_required", connection.Classify(err).Code)
	require.Zero(t, r.gw.total())
}

func TestU3_List_KeepsErrorCodes(t *testing.T) {
	for op, code := range map[string]string{"issues": "rate_limited", "count": "unreachable", "projects": "reconnect_required"} {
		t.Run(op, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				r := newRig(t)
				r.gw.errs[op] = map[string]error{"issues": rateLimited(9 * time.Second),
					"count": &backlog.Error{Kind: backlog.KindUnreachable, Status: 500}, "projects": forbidden()}[op]
				_, err := r.svc.List(r.ctx, "ws-1", Query{})
				require.Equal(t, code, connection.Classify(err).Code, "AC2.1.4")
			})
		})
	}
}

func TestU3_List_MeetsTheLatencyBudget(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		r := newRig(t)
		r.gw.set(fiftySeven)
		r.gw.delay = 300 * time.Millisecond
		start := time.Now()
		_, err := r.svc.List(r.ctx, "ws-1", Query{Keyword: "PROJ-5"})
		require.NoError(t, err)
		require.LessOrEqual(t, time.Since(start), 2500*time.Millisecond, "AC8.1.1 with a 300 ms Backlog")
	})
}

func TestU3_List_RejectsBadQueries(t *testing.T) {
	r := newRig(t)
	_, err := r.svc.List(r.ctx, "ws-1", Query{ProjectKeys: []string{"OTHER"}})
	require.Equal(t, connection.Outcome{Code: "validation", Field: FieldProjectKeys}, connection.Classify(err))
	require.Zero(t, r.gw.total())
}
