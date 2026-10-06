package backlog

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/redact"
	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/testutil"
)

const spaceHost = "example-space.backlog.com"

// fakeBacklog starts a TLS fake Backlog and a client whose dialer sends
// example-space.backlog.com:443 to it. Other hosts are dialled for real, so a
// followed redirect would reach them.
func fakeBacklog(t *testing.T, h http.HandlerFunc) (*Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewTLSServer(h)
	t.Cleanup(srv.Close)
	tr := srv.Client().Transport.(*http.Transport).Clone()
	tr.TLSClientConfig.ServerName = "example.com" // a SAN of the httptest certificate
	backlogAddr := srv.Listener.Addr().String()
	tr.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		if addr == spaceHost+":443" {
			addr = backlogAddr
		}
		return (&net.Dialer{}).DialContext(ctx, network, addr)
	}
	return newClient(tr), srv
}

func serveFixture(t *testing.T, status int, fixture string) http.HandlerFunc {
	body := readFixture(t, fixture)
	return func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(status)
		_, _ = w.Write(body)
	}
}

func creds(key string) Credentials { return Credentials{SpaceHost: spaceHost, APIKey: key} }

func TestMyselfReturnsTheUser(t *testing.T) {
	c, _ := fakeBacklog(t, serveFixture(t, 200, "myself_ok.json"))
	u, err := c.Myself(context.Background(), creds(testutil.APIKey(t)))
	require.NoError(t, err)
	require.Equal(t, User{ID: 1234, UserID: "test.user", Name: "Test User"}, u)
}

func TestMyselfSendsTheKeyAsAQueryParameterOverHTTPS(t *testing.T) {
	key := testutil.APIKey(t) + "+/="
	var got *http.Request
	c, _ := fakeBacklog(t, func(w http.ResponseWriter, r *http.Request) {
		got = r.Clone(context.Background())
		_, _ = w.Write(readFixture(t, "myself_ok.json"))
	})
	_, err := c.Myself(context.Background(), creds(key))
	require.NoError(t, err)
	require.NotNil(t, got.TLS)
	require.Equal(t, spaceHost, got.Host)
	require.Equal(t, "/api/v2/users/myself", got.URL.Path)
	require.Equal(t, key, got.URL.Query().Get("apiKey"))
	require.Contains(t, got.URL.RawQuery, url.QueryEscape(key))
}

func TestMyselfDoesNotFollowRedirects(t *testing.T) {
	var otherHits atomic.Int32
	other := httptest.NewTLSServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { otherHits.Add(1) }))
	defer other.Close()
	c, _ := fakeBacklog(t, func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, other.URL+"/steal", http.StatusFound)
	})
	_, err := c.Myself(context.Background(), creds(testutil.APIKey(t)))
	var be *Error
	require.True(t, errors.As(err, &be))
	require.Equal(t, KindUnreachable, be.Kind)
	require.Equal(t, 302, be.Status)
	require.Zero(t, otherHits.Load(), "the other host must receive 0 requests")
}

func TestMyselfStatusMapping(t *testing.T) {
	cases := []struct {
		status int
		kind   Kind
	}{
		{401, KindUnauthorized}, {403, KindForbidden}, {404, KindNotFound}, {409, KindConflict},
		{400, KindInvalid}, {422, KindInvalid}, {418, KindUnreachable}, {500, KindUnreachable}, {503, KindUnreachable},
	}
	for _, tc := range cases {
		t.Run(strconv.Itoa(tc.status), func(t *testing.T) {
			c, _ := fakeBacklog(t, serveFixture(t, tc.status, "error_401.json"))
			_, err := c.Myself(context.Background(), creds(testutil.APIKey(t)))
			var be *Error
			require.True(t, errors.As(err, &be))
			require.Equal(t, tc.kind, be.Kind)
			require.Equal(t, tc.status, be.Status)
			require.NotContains(t, err.Error(), "SECRET-BODY-MARKER", "the body is never echoed (NFR3.6)")
		})
	}
}

func TestMyselfUnusableBodiesAreUnreachable(t *testing.T) {
	for _, fixture := range []string{"myself_missing_id.json", "myself_empty_name.json", "myself_string_id.json"} {
		t.Run(fixture, func(t *testing.T) {
			c, _ := fakeBacklog(t, serveFixture(t, 200, fixture))
			_, err := c.Myself(context.Background(), creds(testutil.APIKey(t)))
			var be *Error
			require.True(t, errors.As(err, &be))
			require.Equal(t, KindUnreachable, be.Kind)
		})
	}
}

func TestMyselfBodyOverOneMiBIsUnreachable(t *testing.T) {
	c, _ := fakeBacklog(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"id":1,"name":"` + strings.Repeat("x", 2<<20) + `"}`))
	})
	_, err := c.Myself(context.Background(), creds(testutil.APIKey(t)))
	var be *Error
	require.True(t, errors.As(err, &be))
	require.Equal(t, KindUnreachable, be.Kind)
	require.Equal(t, "body", be.Class)
}

func TestMyselfTimeoutIsUnreachable(t *testing.T) {
	release := make(chan struct{})
	c, _ := fakeBacklog(t, func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-release:
		}
	})
	defer close(release)
	c.Timeout = 50 * time.Millisecond
	start := time.Now()
	_, err := c.Myself(context.Background(), creds(testutil.APIKey(t)))
	var be *Error
	require.True(t, errors.As(err, &be))
	require.Equal(t, KindUnreachable, be.Kind)
	require.Equal(t, "timeout", be.Class)
	require.ErrorIs(t, err, context.DeadlineExceeded)
	require.Less(t, time.Since(start), 2*time.Second)
}

func TestMyselfReturnsCancellationUnchanged(t *testing.T) {
	entered := make(chan struct{})
	c, _ := fakeBacklog(t, func(w http.ResponseWriter, r *http.Request) {
		close(entered)
		<-r.Context().Done()
	})
	ctx, cancel := context.WithCancel(context.Background())
	go func() { <-entered; cancel() }()
	_, err := c.Myself(ctx, creds(testutil.APIKey(t)))
	require.Equal(t, context.Canceled, err)
}

func TestMyselfRateLimitWait(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	cases := []struct {
		name    string
		headers map[string]string
		want    time.Duration
	}{
		{"uses X-RateLimit-Reset first", map[string]string{"X-RateLimit-Reset": strconv.FormatInt(now.Unix()+30, 10), "Retry-After": "5"}, 30 * time.Second},
		{"falls back to Retry-After", map[string]string{"Retry-After": "12"}, 12 * time.Second},
		{"defaults to 60 seconds", nil, 60 * time.Second},
		{"clamps a past reset to 1 second", map[string]string{"X-RateLimit-Reset": strconv.FormatInt(now.Unix()-10, 10)}, time.Second},
		{"ignores a malformed reset", map[string]string{"X-RateLimit-Reset": "soon", "Retry-After": "7"}, 7 * time.Second},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var hits atomic.Int32
			c, _ := fakeBacklog(t, func(w http.ResponseWriter, _ *http.Request) {
				hits.Add(1)
				for k, v := range tc.headers {
					w.Header().Set(k, v)
				}
				w.WriteHeader(http.StatusTooManyRequests)
				_, _ = w.Write(readFixture(t, "error_429.json"))
			})
			c.Now = func() time.Time { return now }
			_, err := c.Myself(context.Background(), creds(testutil.APIKey(t)))
			var be *Error
			require.True(t, errors.As(err, &be))
			require.Equal(t, KindRateLimited, be.Kind)
			require.Equal(t, tc.want, be.RetryAfter)
			require.EqualValues(t, 1, hits.Load(), "no retry (BR2.10)")
		})
	}
}

func TestMyselfTransportErrorsHideTheURLAndKey(t *testing.T) {
	key := testutil.APIKey(t)
	c, srv := fakeBacklog(t, func(http.ResponseWriter, *http.Request) {})
	srv.Close() // connection refused
	_, err := c.Myself(context.Background(), creds(key))
	require.Error(t, err)
	var ue *url.Error
	require.False(t, errors.As(err, &ue), "a *url.Error must never escape")
	var be *Error
	require.True(t, errors.As(err, &be))
	require.Equal(t, KindUnreachable, be.Kind)
	require.Zero(t, be.Status)
	testutil.AssertNoLeak(t, err.Error(), key, 8, "apiKey", "?")
}

func TestMyselfRejectsAHostThatIsNotABareHost(t *testing.T) {
	var hits atomic.Int32
	c, _ := fakeBacklog(t, func(http.ResponseWriter, *http.Request) { hits.Add(1) })
	for _, host := range []string{"", "evil.io/x", "a@b.backlog.com", "b.backlog.com:80"} {
		_, err := c.Myself(context.Background(), Credentials{SpaceHost: host, APIKey: "k"})
		var be *Error
		require.True(t, errors.As(err, &be), host)
	}
	require.Zero(t, hits.Load())
}

func TestHTTPSOnlyTransportRefusesPlainHTTP(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { hits.Add(1) }))
	defer srv.Close()
	req, err := http.NewRequest(http.MethodGet, srv.URL, nil)
	require.NoError(t, err)
	_, err = httpsOnly{next: http.DefaultTransport}.RoundTrip(req)
	require.Error(t, err)
	require.Zero(t, hits.Load())
}

func TestMyselfLogsBacklogCallWithoutQuery(t *testing.T) {
	key := testutil.APIKey(t)
	c, _ := fakeBacklog(t, serveFixture(t, 401, "error_401.json"))
	var buf bytes.Buffer
	log := slog.New(redact.NewHandler(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})))
	ctx := redact.WithLogger(context.Background(), log)
	_, _ = c.Myself(ctx, creds(key))
	out := buf.String()
	require.Contains(t, out, `"event":"backlog_call"`)
	require.Contains(t, out, `"path":"/api/v2/users/myself"`)
	require.Contains(t, out, `"status":401`)
	require.Contains(t, out, `"errorKind":"unauthorized"`)
	require.Contains(t, out, `"durationMs"`)
	testutil.AssertNoLeak(t, out, key, 8, "apiKey", "SECRET-BODY-MARKER")
}

func TestNewClientUsesProductionDefaults(t *testing.T) {
	c := NewClient()
	require.Equal(t, 10*time.Second, c.Timeout)
	require.NotNil(t, c.Now)
	require.NotNil(t, c.http.CheckRedirect)
}
