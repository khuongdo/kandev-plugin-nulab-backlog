package plugin

import (
	"context"
	"slices"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/testutil"
)

// storeLatency is the virtual time each state and secret operation takes.
// performance-requirements assumes Kandev's stores answer "well under
// 200 ms" on a self-hosted server; half of that is used here.
const storeLatency = 100 * time.Millisecond

// slowHost adds storeLatency to every state and secret operation.
type slowHost struct{ *fakeHost }

func (h slowHost) GetState(ctx context.Context, scope, scopeID, key string) (map[string]any, bool, error) {
	time.Sleep(storeLatency)
	return h.fakeHost.GetState(ctx, scope, scopeID, key)
}

func (h slowHost) SetState(ctx context.Context, scope, scopeID, key string, value map[string]any) error {
	time.Sleep(storeLatency)
	return h.fakeHost.SetState(ctx, scope, scopeID, key, value)
}

func (h slowHost) GetSecret(ctx context.Context, key string) (string, bool, error) {
	time.Sleep(storeLatency)
	return h.fakeHost.GetSecret(ctx, key)
}

func (h slowHost) SetSecret(ctx context.Context, key, value string) error {
	time.Sleep(storeLatency)
	return h.fakeHost.SetSecret(ctx, key, value)
}

// newSlowRig is a connected rig whose stores answer after storeLatency.
func newSlowRig(t *testing.T) *rig {
	t.Helper()
	r := newRig(t)
	r.rt.SetHost(slowHost{r.host})
	resp, _ := r.call(t, actionConnectAPIKey, connectBody(testutil.APIKey(t)))
	require.Equal(t, 200, resp.Status)
	r.gw.calls = 0
	return r
}

// p95Of calls the action 100 times and returns the 95th percentile duration,
// read from the synctest clock.
func p95Of(t *testing.T, r *rig, action string, body func(i int) any) time.Duration {
	t.Helper()
	var d []time.Duration
	for i := range 100 {
		start := time.Now()
		resp, _ := r.call(t, action, body(i))
		require.Equal(t, 200, resp.Status)
		d = append(d, time.Since(start))
	}
	slices.Sort(d)
	return d[94]
}

// NFR1.1 (T-PERF-03): connection.get p95 <= 500 ms over 100 calls, no Backlog call.
func TestConnectionGetP95Under500ms(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		r := newSlowRig(t)
		p95 := p95Of(t, r, actionGet, func(int) any { return nil })
		require.LessOrEqual(t, p95, 500*time.Millisecond)
		require.Positive(t, p95, "the store latency is counted")
		require.Zero(t, r.gw.calls, "connection.get never calls Backlog")
	})
}

// NFR1.3 (T-PERF-03): connection.set_enabled p95 <= 500 ms over 100 calls.
func TestSetEnabledP95Under500ms(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		r := newSlowRig(t)
		p95 := p95Of(t, r, actionSetEnabled, func(i int) any { return map[string]any{"enabled": i%2 == 1} })
		require.LessOrEqual(t, p95, 500*time.Millisecond)
		require.Positive(t, p95, "the store latency is counted")
		require.Zero(t, r.gw.calls, "connection.set_enabled never calls Backlog")
	})
}
