package issues

import (
	"encoding/json"
	"fmt"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/backlog"
	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/testutil"
)

// TestU3_Leak_NoSecretAnywhere drives every service path with an API key, an
// OAuth token and a failing Backlog, then checks logs, errors, replies, task
// descriptions and candidates.
func TestU3_Leak_NoSecretAnywhere(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		r := newRig(t)
		token := testutil.Token(t)
		r.conn.token = token
		key := r.conn.apiKey
		var out []any
		record := func(v any, err error) {
			out = append(out, v)
			if err != nil {
				out = append(out, err.Error())
			}
		}
		record(r.svc.List(r.ctx, "ws-1", Query{Keyword: "PROJ-118"}))
		record(r.svc.Filters(r.ctx, "ws-1"))
		record(r.svc.CreateTask(r.ctx, "ws-1", createIn("PROJ-118")))
		record(r.svc.Link(r.ctx, "ws-1", "task-17", "PROJ-120"))
		record(r.svc.Detail(r.ctx, "ws-1", "task-17"))
		record(r.svc.Comments(r.ctx, "ws-1", "task-17", 0))
		record(r.svc.Suggest(r.ctx, "ws-1", "login", 5))
		record(r.svc.Links(r.ctx, "ws-1"))
		y := NewSyncer(r.svc, r.log)
		y.Start()
		after(time.Minute)
		record(y.Refresh(r.ctx, "ws-1"))
		y.Stop()

		failing := &backlog.Error{Kind: backlog.KindUnreachable, Status: 500, Class: "http"}
		for _, op := range []string{"issues", "issue", "comments", "attachments", "projects"} {
			r.gw.errs[op] = failing
		}
		record(r.svc.List(r.ctx, "ws-1", Query{}))
		record(r.svc.Detail(r.ctx, "ws-1", "task-17"))
		record(r.svc.Comments(r.ctx, "ws-1", "task-17", 0))
		record(r.svc.CreateTask(r.ctx, "ws-1", createIn("PROJ-120")))

		b, err := json.Marshal(out)
		require.NoError(t, err)
		text := string(b) + fmt.Sprint(out) + r.logs.String() + fmt.Sprint(r.host.creates)
		testutil.AssertNoLeak(t, text, key)
		testutil.AssertNoLeak(t, text, token)
	})
}

// TestU3_Leak_CycleLogHasNoSecret checks the cycle and failure logs alone.
func TestU3_Leak_CycleLogHasNoSecret(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		r, _ := syncRig(t)
		r.gw.set(func(g *fakeGateway) { g.errs["issue"] = rateLimited(30 * time.Second) })
		after(time.Minute)
		require.NotEmpty(t, cycleLines(t, r.logs.String()))
		testutil.AssertNoLeak(t, r.logs.String(), r.conn.apiKey)
	})
}

// TestNFR1_Leak_SummaryNeverInErrorsOrLogs keeps the stored issue summary (Backlog
// content) out of every error message and log line.
func TestNFR1_Leak_SummaryNeverInErrorsOrLogs(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		r, _ := syncRig(t)
		after(time.Minute)
		summary := byTask(r.links(t), "task-17").Summary
		require.Equal(t, "Fix login timeout", summary)
		failing := &backlog.Error{Kind: backlog.KindUnreachable, Status: 500, Class: "http"}
		r.gw.set(func(g *fakeGateway) {
			for _, op := range []string{"issues", "issue", "comments", "attachments", "projects"} {
				g.errs[op] = failing
			}
		})
		var errs []string
		keep := func(_ any, err error) {
			if err != nil {
				errs = append(errs, err.Error())
			}
		}
		keep(r.svc.Detail(r.ctx, "ws-1", "task-17"))
		keep(r.svc.Comments(r.ctx, "ws-1", "task-17", 0))
		keep(r.svc.Link(r.ctx, "ws-1", "task-19", "PROJ-118"))
		after(5 * time.Minute)
		require.NotEmpty(t, errs)
		testutil.AssertNoLeak(t, fmt.Sprint(errs)+r.logs.String(), "", summary)
	})
}
