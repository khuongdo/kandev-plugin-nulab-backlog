package plugin

import (
	"context"
	"testing"
	"testing/synctest"
	"time"

	"github.com/kandev/kandev/pkg/pluginsdk"
	"github.com/stretchr/testify/require"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/backlog"
	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/testutil"
)

// Intent 261007-opt-in-default: Backlog is off after installation until an
// admin turns it on. A workspace without a switch record is either freshly
// installed or upgraded from v0.1.x without ever touching the switch.

const switchStateKey = "workspace/ws-1/integration"

// forgetSwitch removes the switch record, as on a workspace that connected
// under v0.1.x and never touched the switch.
func forgetSwitch(r *rig) {
	r.host.mu.Lock()
	defer r.host.mu.Unlock()
	delete(r.host.state, switchStateKey)
}

func TestOptIn_FreshInstallIsOffAndRefusesConnect(t *testing.T) {
	r := newFreshRig()
	resp, out := r.call(t, actionGet, nil)
	require.Equal(t, 200, resp.Status)
	require.Equal(t, "not_connected", out["state"])
	require.Equal(t, false, out["enabled"], "a fresh install reports Backlog off")

	resp, out = r.call(t, actionConnectAPIKey, connectBody(testutil.APIKey(t)))
	require.Equal(t, 409, resp.Status)
	require.Equal(t, "integration_disabled", errorOf(t, out)["code"])
	require.Zero(t, r.gw.calls, "no Backlog request while off")
	require.Zero(t, r.host.writes, "nothing is stored while off")
}

func TestOptIn_AdminTurnsOnThenConnectsAndOnSurvivesARestart(t *testing.T) {
	r := newFreshRig()
	resp, out := r.call(t, actionSetEnabled, map[string]any{"enabled": true})
	require.Equal(t, 200, resp.Status)
	require.Equal(t, true, out["enabled"])

	resp, out = r.call(t, actionConnectAPIKey, connectBody(testutil.APIKey(t)))
	require.Equal(t, 200, resp.Status)
	require.Equal(t, "connected", out["state"])

	// An explicit on record stays on for a new plugin process on the same host.
	rt := newRuntime(r.gw, r.logs, "debug")
	rt.SetHost(r.host)
	restarted := &rig{rt: rt, host: r.host, gw: r.gw, logs: r.logs}
	_, out = restarted.call(t, actionGet, nil)
	require.Equal(t, true, out["enabled"])
	require.Equal(t, "connected", out["state"])
}

func TestOptIn_UpgradedWorkspaceWithoutARecordKeepsTheConnectionButIsOff(t *testing.T) {
	r := newRig(t)
	r.connected(t)
	forgetSwitch(r)
	_, out := r.call(t, actionGet, nil)
	require.Equal(t, false, out["enabled"])
	require.Equal(t, "connected", out["state"], "the saved connection is kept")
	resp, out := r.call(t, actionTest, nil)
	require.Equal(t, 409, resp.Status)
	require.Equal(t, "integration_disabled", errorOf(t, out)["code"])

	resp, _ = r.call(t, actionSetEnabled, map[string]any{"enabled": true})
	require.Equal(t, 200, resp.Status)
	resp, _ = r.call(t, actionTest, nil)
	require.Equal(t, 200, resp.Status, "turning it on resumes without a reconnect")
}

func TestOptIn_GitCredentialRefusedWithoutASwitchRecord(t *testing.T) {
	r := newU4Rig(t)
	r.gitPassword(t)
	forgetSwitch(r.rig)
	_, err := r.rt.ResolveGitCredential(context.Background(), credRequest())
	require.ErrorContains(t, err, "integration_disabled")
	resp, err := r.rt.GetGitCredentialBinding(context.Background(), &pluginsdk.GitCredentialBindingRequest{ProviderID: "nulab-backlog", WorkspaceID: "ws-1"})
	require.NoError(t, err)
	require.Empty(t, resp.Binding)
}

func TestOptIn_IssueSyncIdleWithoutASwitchRecord(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		r := newU3Rig(t)
		resp, _ := r.onTask(t, actionIssuesLink, "task-17", map[string]any{"issueKey": "PROJ-118"})
		require.Equal(t, 200, resp.Status)
		forgetSwitch(r.rig)
		r.rt.Start()
		defer r.rt.Close()
		time.Sleep(11 * time.Minute)
		synctest.Wait()
		require.Equal(t, 1, r.gw3.count("issue"), "no issue read while off")
	})
}

func TestOptIn_PRWatcherIdleWithoutASwitchRecord(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		r := newU4Rig(t)
		r.gw4.prs = []backlog.PullRequest{{RepositoryID: 11, Number: 1, Summary: "First", StatusID: 1}}
		resp, _ := r.call(t, actionWatchesSave, map[string]any{"name": "Reviews", "projectKey": "PROJ", "repoName": "web-app",
			"statuses": []string{"open"}, "assignee": "anyone", "creator": "anyone", "workflowId": "wf-1"})
		require.Equal(t, 200, resp.Status)
		forgetSwitch(r.rig)
		r.rt.watcher.Every = time.Minute
		paths := len(r.gw4.pathLog())
		r.rt.Start()
		defer r.rt.Close()
		time.Sleep(time.Minute)
		synctest.Wait()
		require.Len(t, r.gw4.pathLog(), paths, "no Backlog call while off")
		r.data.mu.Lock()
		defer r.data.mu.Unlock()
		require.Empty(t, r.data.creates)
	})
}
