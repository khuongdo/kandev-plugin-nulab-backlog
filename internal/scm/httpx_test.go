package scm

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/redact"
	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/testutil"
)

var bearer = func(r *http.Request, c Credential) { r.Header.Set("Authorization", "Bearer "+c.Token) }

// testAPI points an API at a TLS fake server; clock is fixed at 2026-10-07 00:00 UTC.
func testAPI(t *testing.T, h http.HandlerFunc) (*API, *httptest.Server) {
	t.Helper()
	srv := httptest.NewTLSServer(h)
	t.Cleanup(srv.Close)
	a := NewAPI(GitHub, srv.URL, bearer)
	a.HTTP = srv.Client()
	a.HTTP.CheckRedirect = noRedirect
	a.Now = func() time.Time { return time.Date(2026, 10, 7, 0, 0, 0, 0, time.UTC) }
	return a, srv
}

func TestAPI_GetDecodesJSONAndSendsTheAuth(t *testing.T) {
	tok := testutil.Token(t)
	a, _ := testAPI(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "Bearer "+tok, r.Header.Get("Authorization"))
		require.Equal(t, "application/json", r.Header.Get("Accept"))
		require.Equal(t, "/user", r.URL.Path)
		require.Equal(t, "2", r.URL.Query().Get("page"))
		w.Header().Set("X-Total", "9")
		_, _ = w.Write([]byte(`{"login":"lan"}`))
	})
	var out struct {
		Login string `json:"login"`
	}
	h, err := a.Get(context.Background(), Credential{Token: tok}, "/user", map[string][]string{"page": {"2"}}, &out)
	require.NoError(t, err)
	require.Equal(t, "lan", out.Login)
	require.Equal(t, "9", h.Get("X-Total"))
}

// NFR2: only https to the configured host; nothing is dialled otherwise.
func TestAPI_RefusesAnyOtherHostOrSchemeBeforeDialing(t *testing.T) {
	var hits atomic.Int32
	a, srv := testAPI(t, func(http.ResponseWriter, *http.Request) { hits.Add(1) })
	for _, path := range []string{"@evil.example/x", ".evil.example/x", ":1/x"} {
		_, err := a.Get(context.Background(), Credential{Token: "t0ken"}, path, nil, nil)
		require.ErrorIs(t, err, ErrHostRefused, path)
	}
	plain := NewAPI(GitHub, strings.Replace(srv.URL, "https://", "http://", 1), bearer)
	_, err := plain.Get(context.Background(), Credential{}, "/user", nil, nil)
	require.ErrorIs(t, err, ErrHostRefused)
	require.Zero(t, hits.Load())
}

// NFR2: a redirect to another host is never followed.
func TestAPI_NeverFollowsARedirect(t *testing.T) {
	var other atomic.Int32
	elsewhere := httptest.NewTLSServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { other.Add(1) }))
	defer elsewhere.Close()
	a, _ := testAPI(t, func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, elsewhere.URL+"/steal", http.StatusFound)
	})
	_, err := a.Get(context.Background(), Credential{Token: "t0ken"}, "/user", nil, nil)
	var he *HTTPError
	require.ErrorAs(t, err, &he)
	require.Equal(t, 302, he.Status)
	require.Zero(t, other.Load())
}

// NFR3, NFR5: statuses become HTTPError; rate limits carry the wait.
func TestAPI_ErrorStatusesAndRateLimits(t *testing.T) {
	reset := time.Date(2026, 10, 7, 0, 0, 42, 0, time.UTC).Unix()
	cases := []struct {
		name    string
		status  int
		headers map[string]string
		want    int
		wait    time.Duration
	}{
		{"unauthorized", 401, nil, 401, 0},
		{"forbidden", 403, nil, 403, 0},
		{"not found", 404, nil, 404, 0},
		{"server error", 502, nil, 502, 0},
		{"429 with Retry-After", 429, map[string]string{"Retry-After": "30"}, 429, 30 * time.Second},
		{"429 with a reset", 429, map[string]string{"X-RateLimit-Reset": fmt.Sprint(reset)}, 429, 42 * time.Second},
		{"429 with GitLab's reset", 429, map[string]string{"RateLimit-Reset": fmt.Sprint(reset)}, 429, 42 * time.Second},
		{"429 without a header", 429, nil, 429, 60 * time.Second},
		{"GitHub 403 out of quota", 403, map[string]string{"X-RateLimit-Remaining": "0",
			"X-RateLimit-Reset": fmt.Sprint(reset)}, 429, 42 * time.Second},
		{"a reset in the past waits 1 s", 429, map[string]string{"X-RateLimit-Reset": "1"}, 429, time.Second},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var calls atomic.Int32
			a, _ := testAPI(t, func(w http.ResponseWriter, _ *http.Request) {
				calls.Add(1)
				for k, v := range c.headers {
					w.Header().Set(k, v)
				}
				w.WriteHeader(c.status)
				_, _ = w.Write([]byte(`{"message":"secret body text"}`))
			})
			_, err := a.Get(context.Background(), Credential{Token: "t0ken"}, "/user", nil, nil)
			var he *HTTPError
			require.ErrorAs(t, err, &he)
			require.Equal(t, c.want, he.Status)
			require.Equal(t, GitHub, he.Provider)
			require.Equal(t, c.wait, he.RetryAfter)
			require.NotContains(t, err.Error(), "secret body text", "NFR5: no response body in errors")
			require.EqualValues(t, 1, calls.Load(), "NFR3: never retried")
		})
	}
}

// NFR4: the body is read through a limit.
func TestAPI_RefusesAnOversizedBody(t *testing.T) {
	a, _ := testAPI(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(bytes.Repeat([]byte("x"), maxBody+1))
	})
	_, err := a.Get(context.Background(), Credential{}, "/user", nil, &struct{}{})
	var he *HTTPError
	require.ErrorAs(t, err, &he)
	require.Zero(t, he.Status)
}

func TestAPI_MalformedJSONIsUnreachable(t *testing.T) {
	a, _ := testAPI(t, func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte(`{`)) })
	_, err := a.Get(context.Background(), Credential{}, "/user", nil, &struct{}{})
	var he *HTTPError
	require.ErrorAs(t, err, &he)
	require.Zero(t, he.Status)
}

func TestAPI_ReturnsTheCallersCancellationAndDeadlineUnchanged(t *testing.T) {
	block := make(chan struct{})
	a, _ := testAPI(t, func(http.ResponseWriter, *http.Request) { <-block })
	defer close(block)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := a.Get(ctx, Credential{}, "/user", nil, nil)
	require.True(t, errors.Is(err, context.Canceled), err)

	ctx, cancel = context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	_, err = a.Get(ctx, Credential{}, "/user", nil, nil)
	require.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestAPI_TransportFailureIsUnreachableWithoutTheURL(t *testing.T) {
	a, srv := testAPI(t, func(http.ResponseWriter, *http.Request) {})
	srv.Close()
	_, err := a.Get(context.Background(), Credential{Token: "t0ken"}, "/user", nil, nil)
	var he *HTTPError
	require.ErrorAs(t, err, &he)
	require.Zero(t, he.Status)
	require.NotContains(t, err.Error(), "127.0.0.1")
	require.Equal(t, "github: unreachable", err.Error())
}

// NFR1: the token never reaches a log line or an error.
func TestAPI_LogsAndErrorsNeverHoldTheToken(t *testing.T) {
	tok := testutil.Token(t)
	var buf bytes.Buffer
	ctx := redact.WithLogger(context.Background(),
		slog.New(redact.NewHandler(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))))
	a, _ := testAPI(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(401)
		_, _ = w.Write([]byte("Bearer " + tok)) // a provider echoing the credential
	})
	_, err := a.Get(ctx, Credential{Token: tok}, "/user", nil, nil)
	require.Error(t, err)
	require.Contains(t, buf.String(), `"event":"scm_call"`)
	testutil.AssertNoLeak(t, buf.String()+err.Error(), tok)
}

func TestCredential_HidesTheTokenInEveryFormat(t *testing.T) {
	tok := testutil.Token(t)
	c := Credential{Token: tok, Username: "lan"}
	for _, s := range []string{fmt.Sprint(c), fmt.Sprintf("%v %+v %#v %s %q", c, c, c, c, c), c.String(), c.GoString()} {
		testutil.AssertNoLeak(t, s, tok)
		require.Contains(t, s, "lan")
	}
}

func TestHTTPError_Text(t *testing.T) {
	require.Equal(t, "gitlab: HTTP 404", (&HTTPError{Provider: GitLab, Status: 404}).Error())
	require.Equal(t, "bitbucket: unreachable", (&HTTPError{Provider: Bitbucket}).Error())
}

func TestNewAPI_ProductionDefaults(t *testing.T) {
	a := NewAPI(GitLab, "https://gitlab.com/api/v4", nil)
	require.NotNil(t, a.HTTP)
	require.NotNil(t, a.Now)
	require.ErrorIs(t, a.HTTP.CheckRedirect(nil, nil), http.ErrUseLastResponse)
	require.Equal(t, 10*time.Second, a.HTTP.Timeout)
}
