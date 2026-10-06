package backlog

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/redact"
	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/testutil"
)

// handlerTransport serves requests in memory, so timing tests run inside a
// synctest bubble with no socket and no real wait.
type handlerTransport struct{ h http.HandlerFunc }

func (t handlerTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	rec := httptest.NewRecorder()
	t.h(rec, r)
	return rec.Result(), nil
}

func memClient(h http.HandlerFunc) *Client { return newClient(handlerTransport{h: h}) }

func ok(w http.ResponseWriter) { _, _ = w.Write([]byte(`{"id":1,"name":"Test User"}`)) }

func tooMany(w http.ResponseWriter, headers map[string]string) {
	for k, v := range headers {
		w.Header().Set(k, v)
	}
	w.WriteHeader(http.StatusTooManyRequests)
}

func bg(host, method, path string) request {
	return request{method: method, path: path, creds: Credentials{SpaceHost: host, APIKey: "k-1234"}, class: Background}
}

func TestRateLimitWaitsForTheHeaderThenRetries(t *testing.T) {
	cases := []struct {
		name    string
		headers func(now time.Time) map[string]string
		want    time.Duration
	}{
		{"until X-RateLimit-Reset", func(now time.Time) map[string]string {
			return map[string]string{"X-RateLimit-Reset": strconv.FormatInt(now.Unix()+30, 10), "Retry-After": "5"}
		}, 30 * time.Second},
		{"for Retry-After", func(time.Time) map[string]string { return map[string]string{"Retry-After": "12"} }, 12 * time.Second},
		{"60 seconds without a header", func(time.Time) map[string]string { return nil }, 60 * time.Second},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				var starts []time.Time
				c := memClient(func(w http.ResponseWriter, _ *http.Request) {
					starts = append(starts, time.Now())
					if len(starts) == 1 {
						tooMany(w, tc.headers(time.Now()))
						return
					}
					ok(w)
				})
				_, err := c.send(context.Background(), bg(spaceHost, http.MethodGet, myselfPath))
				require.NoError(t, err)
				require.Len(t, starts, 2)
				require.Equal(t, tc.want, starts[1].Sub(starts[0]), "the retry comes only after the virtual wait")
			})
		})
	}
}

func TestRateLimitStopsAfterThreeRetries(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var hits atomic.Int32
		c := memClient(func(w http.ResponseWriter, _ *http.Request) {
			hits.Add(1)
			tooMany(w, map[string]string{"Retry-After": "2"})
		})
		_, err := c.send(context.Background(), bg(spaceHost, http.MethodGet, myselfPath))
		var be *Error
		require.True(t, errors.As(err, &be))
		require.Equal(t, KindRateLimited, be.Kind)
		require.Equal(t, 2*time.Second, be.RetryAfter)
		require.EqualValues(t, 4, hits.Load(), "1 attempt + 3 retries")
	})
}

func TestRateLimitCancelDuringTheWaitReturnsCanceled(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var hits atomic.Int32
		c := memClient(func(w http.ResponseWriter, _ *http.Request) {
			hits.Add(1)
			tooMany(w, map[string]string{"Retry-After": "60"})
		})
		ctx, cancel := context.WithCancel(context.Background())
		go func() { time.Sleep(time.Second); cancel() }()
		_, err := c.send(ctx, bg(spaceHost, http.MethodGet, myselfPath))
		require.Equal(t, context.Canceled, err)
		require.EqualValues(t, 1, hits.Load(), "no further request")
	})
}

func TestRateLimitInteractiveBudget(t *testing.T) {
	cases := []struct {
		name       string
		retryAfter string
		hits       int32
		elapsed    time.Duration
		wantErr    bool
	}{
		{"a wait over 3 s returns at once", "5", 1, 0, true},
		{"a wait of 3 s or less is waited and retried", "2", 2, 2 * time.Second, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				var hits atomic.Int32
				c := memClient(func(w http.ResponseWriter, _ *http.Request) {
					if hits.Add(1) == 1 {
						tooMany(w, map[string]string{"Retry-After": tc.retryAfter})
						return
					}
					ok(w)
				})
				start := time.Now()
				_, err := c.Myself(context.Background(), creds("k-1234"))
				require.Equal(t, tc.hits, hits.Load())
				require.Equal(t, tc.elapsed, time.Since(start))
				if tc.wantErr {
					var be *Error
					require.True(t, errors.As(err, &be))
					require.Equal(t, KindRateLimited, be.Kind)
					require.Equal(t, 5*time.Second, be.RetryAfter)
					return
				}
				require.NoError(t, err)
			})
		})
	}
}

func TestRateLimitBackgroundWaitsTheFullTime(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var hits atomic.Int32
		c := memClient(func(w http.ResponseWriter, _ *http.Request) {
			if hits.Add(1) == 1 {
				tooMany(w, map[string]string{"Retry-After": "60"})
				return
			}
			ok(w)
		})
		start := time.Now()
		_, err := c.send(context.Background(), bg(spaceHost, http.MethodGet, myselfPath))
		require.NoError(t, err)
		require.Equal(t, 60*time.Second, time.Since(start))
	})
}

func TestRateLimitUsesTheInjectedWait(t *testing.T) {
	var waits []time.Duration
	var hits atomic.Int32
	c := memClient(func(w http.ResponseWriter, _ *http.Request) {
		if hits.Add(1) == 1 {
			tooMany(w, map[string]string{"Retry-After": "9"})
			return
		}
		ok(w)
	})
	c.Wait = func(_ context.Context, d time.Duration) error { waits = append(waits, d); return nil }
	_, err := c.send(context.Background(), bg(spaceHost, http.MethodGet, myselfPath))
	require.NoError(t, err)
	require.Equal(t, []time.Duration{9 * time.Second}, waits)
}

// groupRecorder counts requests in flight per group and records start times.
type groupRecorder struct {
	mu       sync.Mutex
	inFlight map[string]int
	maxIn    map[string]int
	starts   map[string][]time.Time
}

func newGroupRecorder() *groupRecorder {
	return &groupRecorder{inFlight: map[string]int{}, maxIn: map[string]int{}, starts: map[string][]time.Time{}}
}

func (g *groupRecorder) handler(hold time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.Host + " " + group(r.Method, r.URL.Path).String()
		g.mu.Lock()
		g.inFlight[key]++
		g.maxIn[key] = max(g.maxIn[key], g.inFlight[key])
		g.starts[key] = append(g.starts[key], time.Now())
		g.mu.Unlock()
		time.Sleep(hold) // virtual time: the request takes a while
		g.mu.Lock()
		g.inFlight[key]--
		g.mu.Unlock()
		ok(w)
	}
}

func TestQueueSearchAndUpdateNeverRunInParallelAndAreSpaced(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		rec := newGroupRecorder()
		c := memClient(rec.handler(100 * time.Millisecond))
		var wg sync.WaitGroup
		for range 3 {
			wg.Go(func() {
				for i := range 10 {
					path, method := "/api/v2/issues", http.MethodGet
					if i%2 == 1 {
						path, method = "/api/v2/issues/PROJ-1", http.MethodPatch
					}
					_, err := c.send(context.Background(), bg(spaceHost, method, path))
					require.NoError(t, err)
				}
			})
		}
		wg.Wait()
		for _, key := range []string{spaceHost + " search", spaceHost + " update"} {
			require.Equal(t, 1, rec.maxIn[key], key+": never in parallel")
			starts := rec.starts[key]
			require.Len(t, starts, 15)
			for i := 1; i < len(starts); i++ {
				require.GreaterOrEqual(t, starts[i].Sub(starts[i-1]), time.Second, key+": at least 1 s apart")
			}
		}
	})
}

func TestQueueReadCallsAreNotQueued(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		rec := newGroupRecorder()
		c := memClient(rec.handler(time.Second))
		var wg sync.WaitGroup
		start := time.Now()
		for range 3 {
			wg.Go(func() {
				_, err := c.send(context.Background(), bg(spaceHost, http.MethodGet, "/api/v2/projects"))
				require.NoError(t, err)
			})
		}
		wg.Wait()
		require.Equal(t, 3, rec.maxIn[spaceHost+" read"])
		require.Equal(t, time.Second, time.Since(start))
	})
}

func TestQueueTwoHostsDoNotBlockEachOther(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		rec := newGroupRecorder()
		c := memClient(rec.handler(5 * time.Second))
		var wg sync.WaitGroup
		start := time.Now()
		for _, host := range []string{"a.backlog.com", "b.backlog.com"} {
			wg.Go(func() {
				_, err := c.send(context.Background(), bg(host, http.MethodPost, "/api/v2/issues"))
				require.NoError(t, err)
			})
		}
		wg.Wait()
		require.Equal(t, 5*time.Second, time.Since(start), "both ran at the same time")
	})
}

func TestQueueTimeCountsTowardTheInteractiveBudget(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		rec := newGroupRecorder()
		c := memClient(rec.handler(5 * time.Second))
		done := make(chan struct{})
		go func() {
			defer close(done)
			_, _ = c.send(context.Background(), bg(spaceHost, http.MethodGet, "/api/v2/issues"))
		}()
		synctest.Wait() // the background search holds the slot
		start := time.Now()
		r := bg(spaceHost, http.MethodGet, "/api/v2/issues")
		r.class = Interactive
		_, err := c.send(context.Background(), r)
		var be *Error
		require.True(t, errors.As(err, &be))
		require.Equal(t, KindRateLimited, be.Kind)
		require.Equal(t, 3*time.Second, time.Since(start), "gives up when the 3 s budget is spent")
		<-done
	})
}

func TestQueueWaitsAreLoggedWithoutQueryOrSecret(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		key := testutil.APIKey(t)
		var hits atomic.Int32
		c := memClient(func(w http.ResponseWriter, _ *http.Request) {
			if hits.Add(1) == 2 {
				tooMany(w, map[string]string{"Retry-After": "2"})
				return
			}
			ok(w)
		})
		var buf bytes.Buffer
		log := slog.New(redact.NewHandler(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})))
		ctx := redact.WithLogger(context.Background(), log)
		r := bg(spaceHost, http.MethodPost, "/api/v2/issues")
		r.creds.APIKey = key
		_, err := c.send(ctx, r)
		require.NoError(t, err)
		_, err = c.send(ctx, r) // spacing wait, then a 429 wait
		require.NoError(t, err)
		out := buf.String()
		require.Contains(t, out, `"event":"backlog_wait","group":"update","reason":"spacing","waitMs":1000,"attempt":1`)
		require.Contains(t, out, `"event":"backlog_wait","group":"update","reason":"rate_limited","waitMs":2000,"attempt":1`)
		testutil.AssertNoLeak(t, out, key, 8, "?")
	})
}
