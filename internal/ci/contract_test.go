package ci

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/testutil"
)

// fakeKandev scripts the Kandev v0.96.0 HTTP surface the contract driver
// uses, and records what it was sent.
type fakeKandev struct {
	mu sync.Mutex

	readyVersion  string
	notReadyFor   int // /ready answers 503 "starting" this many times first
	installStatus int
	installBody   string
	statuses      []string // GET /api/plugins/{id} answers, the last one repeats
	workspaces    string
	getStatus     int
	getBody       string
	connectStatus int
	connectBody   string // "%KEY%" is replaced with the key the driver sent

	order       []string
	readyCalls  int
	statusCalls int
	fieldName   string
	packageData []byte
	sentKey     string
	envelopes   []map[string]any
}

func newFakeKandev() *fakeKandev {
	return &fakeKandev{
		readyVersion:  "v0.96.0",
		installStatus: http.StatusCreated,
		installBody:   `{"plugin":{"id":"nulab-backlog","status":"registered"}}`,
		statuses:      []string{"registered", "active"},
		workspaces:    `{"workspaces":[{"id":"ws-1"}],"total":1}`,
		getStatus:     http.StatusOK,
		getBody:       `{"connected":false,"enabled":true,"state":"not_connected","hasApiKey":false}`,
		connectStatus: http.StatusBadRequest,
		connectBody:   `{"error":{"code":"validation","field":"spaceUrl","requestId":"r-1"}}`,
	}
}

func (f *fakeKandev) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.order = append(f.order, r.Method+" "+r.URL.Path)
	reply := func(status int, body string) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = io.WriteString(w, body)
	}
	switch {
	case r.URL.Path == "/ready":
		f.readyCalls++
		if f.readyCalls <= f.notReadyFor {
			reply(http.StatusServiceUnavailable, `{"status":"starting","service":"kandev","version":"`+f.readyVersion+`"}`)
			return
		}
		reply(http.StatusOK, `{"status":"ok","service":"kandev","version":"`+f.readyVersion+`"}`)
	case r.URL.Path == "/api/plugins/install":
		if mr, err := r.MultipartReader(); err == nil {
			if part, err := mr.NextPart(); err == nil {
				f.fieldName = part.FormName()
				f.packageData, _ = io.ReadAll(io.LimitReader(part, 1<<20))
			}
		}
		reply(f.installStatus, f.installBody)
	case r.URL.Path == "/api/plugins/nulab-backlog":
		i := min(f.statusCalls, len(f.statuses)-1)
		f.statusCalls++
		reply(http.StatusOK, `{"id":"nulab-backlog","status":"`+f.statuses[i]+`","last_error":"boom"}`)
	case r.URL.Path == "/api/v1/workspaces":
		reply(http.StatusOK, f.workspaces)
	case strings.HasPrefix(r.URL.Path, "/api/plugins/nulab-backlog/actions/"):
		var env map[string]any
		_ = json.NewDecoder(r.Body).Decode(&env)
		f.envelopes = append(f.envelopes, env)
		if strings.HasSuffix(r.URL.Path, "/connection.get") {
			reply(f.getStatus, f.getBody)
			return
		}
		body, _ := env["body"].(map[string]any)
		f.sentKey, _ = body["apiKey"].(string)
		reply(f.connectStatus, strings.ReplaceAll(f.connectBody, "%KEY%", f.sentKey))
	default:
		reply(http.StatusNotFound, `{"error":"not found"}`)
	}
}

func startFake(t *testing.T, f *fakeKandev) (Config, *fakeKandev) {
	t.Helper()
	srv := httptest.NewServer(f)
	t.Cleanup(srv.Close)
	pkg := filepath.Join(t.TempDir(), "nulab-backlog-0.0.1.tar.gz")
	require.NoError(t, os.WriteFile(pkg, []byte("dummy package bytes"), 0o600))
	return Config{
		BaseURL: srv.URL, PackagePath: pkg, PluginID: "nulab-backlog", HostVersion: "v0.96.0",
		Wait:         func(context.Context, time.Duration) error { return nil },
		PollInterval: time.Millisecond, ReadyTimeout: 5 * time.Millisecond, ActiveTimeout: 5 * time.Millisecond,
	}, f
}

func TestRunContractHappyPathInstallsAndRunsThePlugin(t *testing.T) {
	fake := newFakeKandev()
	fake.notReadyFor = 2
	cfg, _ := startFake(t, fake)

	require.NoError(t, RunContract(context.Background(), cfg))

	require.Equal(t, []string{
		"GET /ready", "GET /ready", "GET /ready",
		"POST /api/plugins/install",
		"GET /api/plugins/nulab-backlog", "GET /api/plugins/nulab-backlog",
		"GET /api/v1/workspaces",
		"POST /api/plugins/nulab-backlog/actions/connection.get",
		"POST /api/plugins/nulab-backlog/actions/connection.connect_api_key",
	}, fake.order)
	require.Equal(t, "package", fake.fieldName)
	require.Equal(t, []byte("dummy package bytes"), fake.packageData)
	require.True(t, strings.HasPrefix(fake.sentKey, "TESTSECRET-"), "the driver sends only a bait key")
	require.Equal(t, "ws-1", fake.envelopes[0]["workspaceId"])
	require.Equal(t, map[string]any{}, fake.envelopes[0]["body"])
}

func TestRunContractFailsWhenInstallIsRefused(t *testing.T) {
	cases := []struct {
		name   string
		status int
		body   string
		want   string
	}{
		{"bad package", http.StatusBadRequest, `{"error":"plugins: manifest invalid: no id"}`, "manifest invalid: no id"},
		{"version exists", http.StatusConflict, `{"error":"plugins: version already installed"}`, "version already installed"},
		{"installed with a warning", http.StatusCreated, `{"plugin":{"id":"nulab-backlog"},"warning":"runtime did not start"}`, "runtime did not start"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fake := newFakeKandev()
			fake.installStatus, fake.installBody = tc.status, tc.body
			cfg, _ := startFake(t, fake)
			require.ErrorContains(t, RunContract(context.Background(), cfg), tc.want)
		})
	}
}

func TestRunContractFailsAtOnceOnErrorStatus(t *testing.T) {
	fake := newFakeKandev()
	fake.statuses = []string{"error"}
	cfg, _ := startFake(t, fake)
	err := RunContract(context.Background(), cfg)
	require.ErrorContains(t, err, "error")
	require.Equal(t, 1, fake.statusCalls)
}

func TestRunContractFailsWhenThePluginNeverBecomesActive(t *testing.T) {
	fake := newFakeKandev()
	fake.statuses = []string{"registered"}
	cfg, _ := startFake(t, fake)
	require.ErrorContains(t, RunContract(context.Background(), cfg), "not active")
	require.Equal(t, 5, fake.statusCalls, "ActiveTimeout / PollInterval attempts")
}

func TestRunContractFailsWhenTheHostIsNotReadyOrNotTheMinimumVersion(t *testing.T) {
	notReady := newFakeKandev()
	notReady.notReadyFor = 100
	cfg, _ := startFake(t, notReady)
	require.ErrorContains(t, RunContract(context.Background(), cfg), "not ready")

	wrongVersion := newFakeKandev()
	wrongVersion.readyVersion = "v0.97.0"
	cfg, _ = startFake(t, wrongVersion)
	err := RunContract(context.Background(), cfg)
	require.ErrorContains(t, err, "v0.97.0")
	require.ErrorContains(t, err, "v0.96.0")
	require.Equal(t, []string{"GET /ready"}, wrongVersion.order, "nothing is installed on the wrong host")
}

func TestRunContractFailsOnWorkspaceOrConnectionGetProblems(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*fakeKandev)
		want   string
	}{
		{"no workspace", func(f *fakeKandev) { f.workspaces = `{"workspaces":[],"total":0}` }, "no workspace"},
		{"connection.get fails", func(f *fakeKandev) { f.getStatus = http.StatusServiceUnavailable }, "503"},
		{"connection.get wrong state", func(f *fakeKandev) {
			f.getBody = `{"connected":true,"enabled":true,"state":"connected"}`
		}, "not_connected"},
		{"connection.get disabled", func(f *fakeKandev) {
			f.getBody = `{"connected":false,"enabled":false,"state":"not_connected"}`
		}, "enabled"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fake := newFakeKandev()
			tc.mutate(fake)
			cfg, _ := startFake(t, fake)
			require.ErrorContains(t, RunContract(context.Background(), cfg), tc.want)
		})
	}
}

func TestRunContractFailsOnWrongValidationReplyWithoutLeakingTheKey(t *testing.T) {
	cases := []struct {
		name   string
		status int
		body   string
	}{
		{"accepted", http.StatusOK, `{"connected":true,"echo":"%KEY%"}`},
		{"wrong field", http.StatusBadRequest, `{"error":{"code":"validation","field":"apiKey"},"echo":"%KEY%"}`},
		{"server error", http.StatusInternalServerError, `{"error":"%KEY%"}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fake := newFakeKandev()
			fake.connectStatus, fake.connectBody = tc.status, tc.body
			cfg, _ := startFake(t, fake)
			err := RunContract(context.Background(), cfg)
			require.ErrorContains(t, err, "connection.connect_api_key")

			testutil.AssertNoLeak(t, err.Error(), fake.sentKey)
		})
	}
}

func TestRunContractStopsOnCanceledContextAndMissingPackage(t *testing.T) {
	fake := newFakeKandev()
	fake.notReadyFor = 100
	cfg, _ := startFake(t, fake)
	ctx, cancel := context.WithCancel(context.Background())
	cfg.Wait = func(context.Context, time.Duration) error { cancel(); return context.Canceled }
	require.ErrorIs(t, RunContract(ctx, cfg), context.Canceled)

	cfg, _ = startFake(t, newFakeKandev())
	cfg.PackagePath = filepath.Join(t.TempDir(), "absent.tar.gz")
	require.ErrorContains(t, RunContract(context.Background(), cfg), "package")
}
