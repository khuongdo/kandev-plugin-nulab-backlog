package plugin

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/kandev/kandev/pkg/pluginsdk"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/backlog"
	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/connection"
	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/git"
	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/redact"
)

// Build information, set with -ldflags -X by the Makefile.
var (
	Version = "dev"
	SDKRef  = "unknown"
)

// Action keys declared in manifest.yaml.
const (
	actionGet           = "connection.get"
	actionConnectAPIKey = "connection.connect_api_key"
	actionSetEnabled    = "connection.set_enabled"
	actionStartOAuth    = "connection.start_oauth"
	actionTest          = "connection.test"
	actionDisconnect    = "connection.disconnect"
	actionListProjects  = "connection.list_projects"
	actionSetProjects   = "connection.set_projects"
)

// guarded reports whether an action is refused while Backlog is off for the
// workspace (BR7.3). Every action is, except the two the UI needs to show the
// state and turn Backlog back on, so later units' actions are guarded by default.
func guarded(action string) bool {
	return action != actionGet && action != actionSetEnabled
}

// errBadBody marks a request body that is not the expected JSON object.
var errBadBody = errors.New("malformed action body")

// handler runs one action for the verified workspace and returns the JSON reply.
type handler func(r *Runtime, ctx context.Context, workspaceID string, body []byte) (any, error)

var handlers = map[string]handler{
	actionGet: func(r *Runtime, ctx context.Context, ws string, _ []byte) (any, error) {
		return r.service.Get(ctx, ws)
	},
	actionConnectAPIKey: func(r *Runtime, ctx context.Context, ws string, body []byte) (any, error) {
		var in connection.ConnectInput
		if err := json.Unmarshal(body, &in); err != nil {
			return nil, errBadBody
		}
		return r.service.Connect(ctx, ws, in)
	},
	actionSetEnabled: func(r *Runtime, ctx context.Context, ws string, body []byte) (any, error) {
		// The body has no workspace field: the workspace is always the verified one (NFR3.8).
		var in struct {
			Enabled *bool `json:"enabled"`
		}
		if err := json.Unmarshal(body, &in); err != nil || in.Enabled == nil {
			return nil, &connection.FieldError{Field: connection.FieldEnabled, Err: errors.New("must be a boolean")}
		}
		return r.service.SetEnabled(ctx, ws, *in.Enabled)
	},
	actionStartOAuth: func(r *Runtime, ctx context.Context, ws string, body []byte) (any, error) {
		var in connection.StartInput
		if err := json.Unmarshal(body, &in); err != nil {
			return nil, errBadBody
		}
		return r.service.StartOAuth(ctx, ws, in)
	},
	actionTest: func(r *Runtime, ctx context.Context, ws string, _ []byte) (any, error) {
		return r.service.Test(ctx, ws)
	},
	actionDisconnect: func(r *Runtime, ctx context.Context, ws string, _ []byte) (any, error) {
		return r.service.Disconnect(ctx, ws)
	},
	actionListProjects: func(r *Runtime, ctx context.Context, ws string, _ []byte) (any, error) {
		items, err := r.service.ListProjects(ctx, ws)
		if err != nil {
			return nil, err
		}
		return map[string]any{"projects": items}, nil
	},
	actionSetProjects: func(r *Runtime, ctx context.Context, ws string, body []byte) (any, error) {
		var in struct {
			ProjectKeys any `json:"projectKeys"`
		}
		if err := json.Unmarshal(body, &in); err != nil {
			return nil, connection.ErrInvalidProjects
		}
		return r.service.SetProjects(ctx, ws, in.ProjectKeys)
	},
}

const codeNotFound = "not_found"

// Runtime is the Kandev plugin implementation for Nulab Backlog.
type Runtime struct {
	pluginsdk.UnimplementedPlugin
	service *connection.Service
	git     *git.Service
	watcher *git.Watcher
	log     *slog.Logger

	lifeMu   sync.Mutex
	started  bool
	unlisten func()
}

var _ pluginsdk.ActionHandler = (*Runtime)(nil)

// gateway is every Backlog call the plugin makes. backlog.Client implements it.
type gateway interface {
	connection.Gateway
	git.Gateway
}

// NewRuntime returns the plugin served by server/main.go, with the Git
// service listening to ConnectionChanged and the PR watcher running. It logs
// JSON to standard error at the level in KANDEV_PLUGIN_LOG_LEVEL (default info).
func NewRuntime() *Runtime {
	r := newRuntime(backlog.NewClient(), os.Stderr, os.Getenv("KANDEV_PLUGIN_LOG_LEVEL"))
	r.Start()
	return r
}

// Start subscribes the Git service to ConnectionChanged and starts the PR
// watcher. It starts nothing twice.
func (r *Runtime) Start() {
	r.lifeMu.Lock()
	defer r.lifeMu.Unlock()
	if r.started {
		return
	}
	r.started = true
	r.unlisten = r.git.Listen()
	r.watcher.Start()
}

// Close stops the watcher and the subscription. Safe to call more than once.
func (r *Runtime) Close() {
	r.lifeMu.Lock()
	defer r.lifeMu.Unlock()
	if !r.started {
		return
	}
	r.started = false
	r.watcher.Stop()
	r.unlisten()
	r.watcher = git.NewWatcher(r.git, r.log) // a stopped watcher can be started again
}

func newRuntime(gateway gateway, logOut io.Writer, level string) *Runtime {
	lvl := slog.LevelInfo
	if strings.EqualFold(level, "debug") {
		lvl = slog.LevelDebug
	}
	r := &Runtime{log: slog.New(redact.NewHandler(slog.NewJSONHandler(logOut, &slog.HandlerOptions{Level: lvl})))}
	// The stores resolve the injected Host on every call, because Serve
	// injects it from a background goroutine after NewRuntime returns.
	stores := hostStores{host: r.Host}
	r.service = connection.NewService(gateway, connection.NewStore(stores, stores))
	r.service.Config = stores
	r.git = git.NewService(gateway, r.service, hostPort{host: r.Host}, git.NewStore(stores))
	r.watcher = git.NewWatcher(r.git, r.log)
	r.log.Info("plugin started", "event", "plugin_started", "version", Version,
		"platform", runtime.GOOS+"-"+runtime.GOARCH, "sdkRef", SDKRef)
	return r
}

// HandleAction routes a verified browser action. It always returns a
// response with a status and never a Go error, so the UI keeps the error
// code (contract C5).
func (r *Runtime) HandleAction(ctx context.Context, req *pluginsdk.PluginActionRequest) (resp *pluginsdk.PluginActionResponse, err error) {
	start := time.Now()
	requestID := newRequestID()
	log := r.log.With("workspaceId", req.Context.WorkspaceID, "requestId", requestID)
	ctx = redact.WithLogger(ctx, log)
	ctx = context.WithValue(ctx, verifiedKey{}, req.Context)
	defer func() {
		if p := recover(); p != nil {
			log.ErrorContext(ctx, "action panicked", "event", "action_panic", "action", req.ActionKey)
			resp, err = errorResponse(outcome{Outcome: connection.Outcome{Code: connection.CodeInternal}}, requestID), nil
		}
	}()

	h, ok := handlers[req.ActionKey]
	if !ok {
		return reject(ctx, log, req.ActionKey, codeNotFound, requestID), nil
	}
	var actErr error
	if guarded(req.ActionKey) {
		// The action deadline starts before the guard, so the Connect budget
		// (NFR1.4: 12 s plus a 2 s rollback) also covers the guard's read.
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, r.service.Deadline)
		defer cancel()
		// The one guard: read the switch before anything else; fail closed (NFR3.9).
		actErr = r.service.RequireEnabled(ctx, req.Context.WorkspaceID)
	}
	var reply any
	if actErr == nil {
		reply, actErr = h(r, ctx, req.Context.WorkspaceID, req.Body)
	}
	if errors.Is(actErr, errBadBody) {
		return reject(ctx, log, req.ActionKey, connection.CodeValidation, requestID), nil
	}
	if actErr != nil {
		out := classify(actErr)
		logFailure(ctx, log, req.ActionKey, out.Code, actErr, time.Since(start))
		return errorResponse(out, requestID), nil
	}
	return jsonResponse(200, reply, nil), nil
}

// logFailure writes one action_failed line for every failure except a
// validation error (AC8.3.1): ERROR for internal, WARN otherwise. The error
// text goes through the redacting handler.
func logFailure(ctx context.Context, log *slog.Logger, action, code string, err error, d time.Duration) {
	if code == connection.CodeValidation {
		return
	}
	level := slog.LevelWarn
	if code == connection.CodeInternal {
		level = slog.LevelError
	}
	log.Log(ctx, level, "action failed", "event", "action_failed", "action", action, "errorCode", code,
		"durationMs", d.Milliseconds(), "err", err)
}

// reject answers a request refused before it reaches the service, and logs
// it so its requestId can be found in the logs.
func reject(ctx context.Context, log *slog.Logger, action, code, requestID string) *pluginsdk.PluginActionResponse {
	log.WarnContext(ctx, "action rejected", "event", "action_rejected", "action", action, "errorCode", code)
	return errorResponse(outcome{Outcome: connection.Outcome{Code: code}}, requestID)
}

// statusFor is the one table from action error code to HTTP status.
func statusFor(code string) int {
	switch code {
	case connection.CodeValidation:
		return 400
	case connection.CodeReconnectRequired:
		return 401
	case codeNotFound:
		return 404
	case connection.CodeConflict, connection.CodeIntegrationDisabled:
		return 409
	case connection.CodeRateLimited:
		return 429
	case connection.CodeUnreachable:
		return 503
	default:
		return 500
	}
}

type actionError struct {
	Code              string `json:"code"`
	RetryAfterSeconds int    `json:"retryAfterSeconds,omitempty"`
	Field             string `json:"field,omitempty"`
	PullRequestNumber int    `json:"pullRequestNumber,omitempty"` // U4: the open PR of a create conflict
	RequestID         string `json:"requestId"`
}

// outcome is an action error: the connection outcome plus U4's open PR number.
type outcome struct {
	connection.Outcome
	PullRequestNumber int
}

// classify maps an action error to its code: U4's Git errors first, then
// the connection's table.
func classify(err error) outcome {
	var open *git.OpenPRExistsError
	switch {
	case errors.As(err, &open):
		return outcome{Outcome: connection.Outcome{Code: connection.CodeConflict}, PullRequestNumber: open.Number}
	case errors.Is(err, git.ErrNotFound):
		return outcome{Outcome: connection.Outcome{Code: codeNotFound}}
	case errors.Is(err, git.ErrConflict), errors.Is(err, git.ErrStale):
		return outcome{Outcome: connection.Outcome{Code: connection.CodeConflict}}
	}
	return outcome{Outcome: connection.Classify(err)}
}

func errorResponse(out outcome, requestID string) *pluginsdk.PluginActionResponse {
	var headers map[string]string
	if out.RetryAfterSeconds > 0 {
		headers = map[string]string{"Retry-After": strconv.Itoa(out.RetryAfterSeconds)}
	}
	body := map[string]actionError{"error": {
		Code: out.Code, RetryAfterSeconds: out.RetryAfterSeconds, Field: out.Field,
		PullRequestNumber: out.PullRequestNumber, RequestID: requestID,
	}}
	return jsonResponse(statusFor(out.Code), body, headers)
}

func jsonResponse(status int, v any, headers map[string]string) *pluginsdk.PluginActionResponse {
	body, err := json.Marshal(v)
	if err != nil {
		status, body = 500, []byte(`{"error":{"code":"internal"}}`)
	}
	h := map[string]string{"Content-Type": "application/json"}
	for k, val := range headers {
		h[k] = val
	}
	return &pluginsdk.PluginActionResponse{Status: status, Body: body, Headers: h}
}

func newRequestID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b) // crypto/rand.Read never fails on supported platforms
	return hex.EncodeToString(b)
}

// errNoHost is returned while Kandev has not injected the Host yet.
var errNoHost = errors.New("kandev host not connected")

// hostStores adapts the injected pluginsdk.Host to the connection stores.
type hostStores struct {
	host func() pluginsdk.Host
}

func (s hostStores) get() (pluginsdk.Host, error) {
	if h := s.host(); h != nil {
		return h, nil
	}
	return nil, errNoHost
}

func (s hostStores) GetSecret(ctx context.Context, key string) (string, bool, error) {
	h, err := s.get()
	if err != nil {
		return "", false, err
	}
	return h.GetSecret(ctx, key)
}

func (s hostStores) SetSecret(ctx context.Context, key, value string) error {
	h, err := s.get()
	if err != nil {
		return err
	}
	return h.SetSecret(ctx, key, value)
}

func (s hostStores) DeleteSecret(ctx context.Context, key string) error {
	h, err := s.get()
	if err != nil {
		return err
	}
	return h.DeleteSecret(ctx, key)
}

func (s hostStores) GetState(ctx context.Context, scope, scopeID, key string) (map[string]any, bool, error) {
	h, err := s.get()
	if err != nil {
		return nil, false, err
	}
	return h.GetState(ctx, scope, scopeID, key)
}

func (s hostStores) SetState(ctx context.Context, scope, scopeID, key string, value map[string]any) error {
	h, err := s.get()
	if err != nil {
		return err
	}
	return h.SetState(ctx, scope, scopeID, key, value)
}

func (s hostStores) DeleteState(ctx context.Context, scope, scopeID, key string) error {
	h, err := s.get()
	if err != nil {
		return err
	}
	return h.DeleteState(ctx, scope, scopeID, key)
}
