package issues

import (
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/backlog"
	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/connection"
)

func TestU3_Suggest_ExactKeyFirst(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		r := newRig(t)
		r.gw.set(func(g *fakeGateway) {
			g.issues = append(g.issues, backlog.Issue{ID: 1, ProjectID: 101, IssueKey: "PROJ-7", Summary: "About proj-120 too"})
		})
		got, err := r.svc.Suggest(r.ctx, "ws-1", "proj-120", 0)
		require.NoError(t, err)
		require.Equal(t, []Candidate{
			{Key: "PROJ-120", Title: "Login page", URL: "https://" + spaceHost + "/view/PROJ-120"},
			{Key: "PROJ-7", Title: "About proj-120 too", URL: "https://" + spaceHost + "/view/PROJ-7"},
		}, got, "AC3.4.1: the exact key first, then the keyword matches")
		require.Equal(t, 1, r.gw.count("issue"))
		require.Equal(t, 1, r.gw.count("issues"))
		require.Equal(t, 5, r.gw.queries[0].Count, "the default limit")
		require.Equal(t, []backlog.CallClass{backlog.Interactive}, r.gw.classes["issues"])
	})
}

func TestU3_Suggest_KeywordAndLimit(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		r := newRig(t)
		got, err := r.svc.Suggest(r.ctx, "ws-1", "login", 2)
		require.NoError(t, err)
		require.Equal(t, []string{"PROJ-118", "PROJ-120"}, keys(got))
		require.Zero(t, r.gw.count("issue"), "not a key: no exact-key call")
		for in, want := range map[int]int{50: 10, -1: 5} {
			_, err := r.svc.Suggest(r.ctx, "ws-1", "login", in)
			require.NoError(t, err)
			require.Equal(t, want, r.gw.queries[len(r.gw.queries)-1].Count, "limit %d", in)
		}
	})
}

func keys(cs []Candidate) []string {
	out := []string{}
	for _, c := range cs {
		out = append(out, c.Key)
	}
	return out
}

func TestU3_Suggest_EmptyNeverAnError(t *testing.T) {
	cases := map[string]func(r *rig){
		"not connected": func(r *rig) { r.conn.currentErr = connection.ErrNotConnected },
		"switch off":    func(r *rig) { r.conn.disabled = true },
		"switch broken": func(r *rig) { r.conn.enabledErr = connection.ErrStore },
		"no match":      func(*rig) {},
		"rate limited":  func(r *rig) { r.gw.errs["issues"] = rateLimited(30 * time.Second) },
	}
	for name, setup := range cases {
		t.Run(name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				r := newRig(t)
				setup(r)
				got, err := r.svc.Suggest(r.ctx, "ws-1", "zzz-nothing", 5)
				require.NoError(t, err, "AC3.4.3")
				require.Empty(t, got)
				require.NotNil(t, got)
				if name == "not connected" || name == "switch off" || name == "switch broken" {
					require.Zero(t, r.gw.total())
				}
			})
		})
	}
	synctest.Test(t, func(t *testing.T) {
		r := newRig(t)
		got, err := r.svc.Suggest(r.ctx, "ws-1", "PROJ-999", 5)
		require.NoError(t, err, "an unknown key")
		require.Empty(t, got)
	})
}

func TestU3_Suggest_LatestQueryWins(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		r := newRig(t)
		results := make(chan []Candidate, 5)
		for i, q := range []string{"l", "lo", "log", "logi", "login"} {
			go func() {
				got, err := r.svc.Suggest(r.ctx, "ws-1", q, 5)
				require.NoError(t, err)
				results <- got
			}()
			if i < 4 {
				time.Sleep(50 * time.Millisecond)
			}
		}
		var nonEmpty int
		for range 5 {
			if len(<-results) > 0 {
				nonEmpty++
			}
		}
		require.Equal(t, 1, r.gw.count("issues"), "AC3.4.2: one Backlog search")
		require.Equal(t, "login", r.gw.queries[0].Keyword, "for the last query")
		require.Equal(t, 1, nonEmpty)
	})
}

func TestU3_Suggest_StaysInsideTheDeadline(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		r := newRig(t)
		r.gw.delay = 5 * time.Second
		start := time.Now()
		got, err := r.svc.Suggest(r.ctx, "ws-1", "login", 5)
		require.NoError(t, err)
		require.Empty(t, got)
		require.LessOrEqual(t, time.Since(start), 1200*time.Millisecond, "inside Kandev's 1.5 s")
	})
}

func TestU3_Authorize_AllowsOnlyALiveIssue(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		r := newRig(t)
		require.True(t, r.svc.Authorize(r.ctx, "ws-1", "PROJ-118"))
		require.Equal(t, []backlog.CallClass{backlog.Interactive}, r.gw.classes["issue"])
	})
}

func TestU3_Authorize_FailsClosed(t *testing.T) {
	cases := map[string]struct {
		key   string
		setup func(r *rig)
	}{
		"unknown issue":      {"PROJ-999", func(*rig) {}},
		"unselected project": {"OTHER-1", func(*rig) {}},
		"not a key":          {"../x", func(*rig) {}},
		"rate limited":       {"PROJ-118", func(r *rig) { r.gw.errs["issue"] = rateLimited(time.Second) }},
		"switch off":         {"PROJ-118", func(r *rig) { r.conn.disabled = true }},
		"not connected":      {"PROJ-118", func(r *rig) { r.conn.currentErr = connection.ErrNotConnected }},
		"too slow":           {"PROJ-118", func(r *rig) { r.gw.delay = 2 * time.Second }},
		"other key returned": {"PROJ-118", func(r *rig) { r.gw.issues[0].IssueKey = "PROJ-1" }},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				r := newRig(t)
				tc.setup(r)
				start := time.Now()
				require.False(t, r.svc.Authorize(r.ctx, "ws-1", tc.key))
				require.LessOrEqual(t, time.Since(start), 1200*time.Millisecond)
			})
		})
	}
}
