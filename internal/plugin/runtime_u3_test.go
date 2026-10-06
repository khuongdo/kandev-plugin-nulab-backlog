package plugin

import (
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/backlog"
)

func TestU3_Runtime_StartWiresSyncAndEvents(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		r := newU3Rig(t)
		r.rt.Start()
		defer r.rt.Close()
		resp, _ := r.onTask(t, actionIssuesLink, "task-17", map[string]any{"issueKey": "PROJ-118"})
		require.Equal(t, 200, resp.Status)
		r.gw3.setIssue(func(i *backlog.Issue) { i.StatusName = "Closed" })
		time.Sleep(time.Minute)
		synctest.Wait()
		require.Equal(t, 2, r.gw3.count("issue"), "the link's read, then the first cycle")
		r.gw3.mu.Lock()
		require.Equal(t, backlog.Background, r.gw3.classes[1])
		r.gw3.mu.Unlock()
		_, out := r.call(t, actionIssuesLinks, nil)
		require.Equal(t, "Closed", out["links"].([]any)[0].(map[string]any)["status"])

		resp, out = r.call(t, actionIssuesRefresh, nil)
		require.Equal(t, 200, resp.Status)
		require.EqualValues(t, 0, out["updatedCount"])
		require.NotEmpty(t, out["refreshedAt"])

		r.call(t, keyDisconnect, nil)
		synctest.Wait()
		_, out = r.call(t, actionIssuesLinks, nil)
		require.Equal(t, "not_connected", out["links"].([]any)[0].(map[string]any)["state"], "the issue service hears ConnectionChanged")
	})
}

func TestU3_Runtime_NewRuntimeStartsNothing(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		r := newU3Rig(t)
		resp, _ := r.onTask(t, actionIssuesLink, "task-17", map[string]any{"issueKey": "PROJ-118"})
		require.Equal(t, 200, resp.Status)
		time.Sleep(10 * time.Minute)
		require.Equal(t, 1, r.gw3.count("issue"), "no cycle without Start")
		start := time.Now()
		resp, out := r.call(t, actionIssuesRefresh, nil)
		require.Equal(t, 503, resp.Status)
		require.Equal(t, "unreachable", errorOf(t, out)["code"])
		require.Equal(t, 10*time.Second, time.Since(start), "a refresh waits at most 10 s")
	})
}

func TestU3_Runtime_SwitchOffStopsTheCycle(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		r := newU3Rig(t)
		r.rt.Start()
		defer r.rt.Close()
		resp, _ := r.onTask(t, actionIssuesLink, "task-17", map[string]any{"issueKey": "PROJ-118"})
		require.Equal(t, 200, resp.Status)
		r.call(t, actionSetEnabled, map[string]any{"enabled": false})
		time.Sleep(11 * time.Minute)
		synctest.Wait()
		require.Equal(t, 1, r.gw3.count("issue"), "AC1.5.4: no read while Backlog is off")
		r.call(t, actionSetEnabled, map[string]any{"enabled": true})
		time.Sleep(time.Minute)
		synctest.Wait()
		require.Equal(t, 2, r.gw3.count("issue"), "back on: the cycle resumes")
	})
}

func TestU3_Runtime_CloseIsSafeTwice(t *testing.T) {
	r := newU3Rig(t)
	r.rt.Start()
	r.rt.Close()
	r.rt.Close()
	r.rt.Start()
	r.rt.Close()
}
