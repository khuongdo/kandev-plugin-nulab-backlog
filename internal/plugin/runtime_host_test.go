package plugin

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/kandev/kandev/pkg/pluginsdk"
)

// Loop-back 2 (T-COMPAT-01, R-04): Serve injects the Host from a background
// goroutine, so an action can arrive before SetHost. These tests build the
// runtime without a Host and call SetHost themselves.

// hostlessRig is a rig whose Host has not been injected yet.
func hostlessRig() *rig {
	gw, logs, host := &fakeGateway{}, &syncBuffer{}, newFakeHost()
	return &rig{rt: newRuntime(gw, logs, "debug"), host: host, gw: gw, logs: logs}
}

// handle calls an action with ctx and returns its status and error code.
func (r *rig) handle(ctx context.Context, key string) (int, string) {
	resp, err := r.rt.HandleAction(ctx, &pluginsdk.PluginActionRequest{
		ActionKey: key,
		Context:   pluginsdk.VerifiedActionContext{WorkspaceID: "ws-1", ActorID: "user-1"},
	})
	if err != nil || resp == nil {
		return 0, "go-error"
	}
	var out struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	_ = json.Unmarshal(resp.Body, &out)
	return resp.Status, out.Error.Code
}

// (a) An action that arrives before SetHost waits for it, then succeeds.
func TestHost_EarlyActionWaitsForHost(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		r, t0 := hostlessRig(), time.Now()
		var status int
		done := make(chan struct{})
		go func() {
			defer close(done)
			status, _ = r.handle(context.Background(), actionGet)
		}()
		time.Sleep(time.Second)
		synctest.Wait()
		select {
		case <-done:
			t.Fatalf("the action answered %d before the Host was set", status)
		default:
		}
		r.rt.SetHost(r.host)
		<-done
		require.Equal(t, 200, status)
		require.Equal(t, time.Second, time.Since(t0), "it answered as soon as the Host arrived")
	})
}

// (b) A context deadline before SetHost fails closed with internal, and no
// store is touched when the Host arrives later.
func TestHost_DeadlineFailsClosed(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		r, t0 := hostlessRig(), time.Now()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		status, code := r.handle(ctx, actionGet)
		require.Equal(t, 500, status)
		require.Equal(t, "internal", code)
		require.Equal(t, time.Second, time.Since(t0), "the call's deadline bounds the wait")
		require.Contains(t, r.logs.String(), errNoHost.Error())

		r.rt.SetHost(r.host)
		r.host.mu.Lock()
		defer r.host.mu.Unlock()
		require.Empty(t, r.host.reads, "no store call after the failed wait")
		require.Zero(t, r.host.writes)
	})
}

// (b) Without a deadline, the wait stops at the 5-second cap.
func TestHost_WaitIsCappedAt5s(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		r, t0 := hostlessRig(), time.Now()
		status, code := r.handle(context.Background(), actionGet)
		require.Equal(t, 500, status)
		require.Equal(t, "internal", code)
		require.Equal(t, hostWait, time.Since(t0))
		require.Equal(t, 5*time.Second, hostWait)
	})
}

// (b) A nil Host does not open the gate.
func TestHost_NilHostDoesNotOpen(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		r, t0 := hostlessRig(), time.Now()
		r.rt.SetHost(nil)
		status, code := r.handle(context.Background(), actionGet)
		require.Equal(t, 500, status)
		require.Equal(t, "internal", code)
		require.Equal(t, hostWait, time.Since(t0))
	})
}

// (c) Once the Host is set, store calls never wait.
func TestHost_SetHostMeansNoWait(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		r, t0 := hostlessRig(), time.Now()
		r.rt.SetHost(r.host)
		for range 3 {
			status, _ := r.handle(context.Background(), actionGet)
			require.Equal(t, 200, status)
		}
		require.Zero(t, time.Since(t0))
	})
}

// (c) SetHost is safe to call again and from many goroutines: the gate
// opens once and the latest Host is used.
func TestHost_SetHostTwiceIsSafe(t *testing.T) {
	r := hostlessRig()
	other := newFakeHost()
	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() { r.rt.SetHost(other) })
		wg.Go(func() { _, _ = r.handle(context.Background(), actionGet) })
	}
	wg.Wait()
	r.rt.SetHost(r.host)
	status, _ := r.handle(context.Background(), actionGet)
	require.Equal(t, 200, status)
	r.host.mu.Lock()
	defer r.host.mu.Unlock()
	require.NotEmpty(t, r.host.reads, "the latest Host serves the call")
}

// (d) Workers that tick before SetHost skip the cycle: no ERROR line and at
// most one cycle line per worker tick. Once the Host is set they run clean.
func TestHost_WorkersWaitWithoutStorm(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		r := hostlessRig()
		r.rt.watcher.Every = time.Minute
		r.rt.syncer.Tick = time.Minute
		r.rt.Start()
		defer r.rt.Close()
		// The first ticks wait inside their 1 s store call; the Host arrives.
		time.Sleep(time.Minute + 500*time.Millisecond)
		r.rt.SetHost(r.host)
		time.Sleep(time.Minute)
		synctest.Wait()

		logs := r.logs.String()
		require.NotContains(t, logs, `"level":"ERROR"`)
		for _, event := range []string{"watch_cycle", "issue_sync_cycle"} {
			require.LessOrEqual(t, strings.Count(logs, `"event":"`+event+`"`), 2, event)
		}
		require.NotRegexp(t, `"errors":[1-9]`, logs, "the first cycle waited for the Host")
	})
}
