package ci

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// maxResponse bounds every response body the driver reads.
const maxResponse = 1 << 20

// invalidSpaceURL fails the plugin's address check (https and a Backlog
// domain only) before any network call.
const invalidSpaceURL = "http://not-a-backlog-space.invalid"

// Config drives one packaged-host contract run against a throwaway Kandev.
type Config struct {
	BaseURL     string // for example http://127.0.0.1:38529
	PackagePath string // dist/<id>-<version>.tar.gz
	PluginID    string
	HostVersion string // the exact /ready version, for example v0.96.0

	// Wait sleeps between polls; nil uses a timer. Tests inject a no-op.
	Wait          func(ctx context.Context, d time.Duration) error
	PollInterval  time.Duration
	ReadyTimeout  time.Duration // until /ready reports ok
	ActiveTimeout time.Duration // until the plugin status is active
	Client        *http.Client  // nil uses a client with a 30 s timeout
}

// RunContract checks that the packaged plugin installs and runs on the host:
// the host is ready at exactly HostVersion (AC7.4.2), the package installs
// through the multipart upload with no warning, the plugin becomes active,
// connection.get answers not_connected and enabled for the default
// workspace, and the admin action connection.connect_api_key maps a bad
// space address to a validation error on spaceUrl.
func RunContract(ctx context.Context, cfg Config) error {
	d := driver{cfg: cfg, client: cfg.Client, wait: cfg.Wait}
	if d.client == nil {
		d.client = &http.Client{Timeout: 30 * time.Second}
	}
	if d.wait == nil {
		d.wait = sleep
	}
	steps := []func(context.Context) error{d.ready, d.install, d.active, d.actions}
	for _, step := range steps {
		if err := step(ctx); err != nil {
			return err
		}
	}
	return nil
}

func sleep(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

type driver struct {
	cfg    Config
	client *http.Client
	wait   func(context.Context, time.Duration) error
}

// poll calls try up to timeout/PollInterval times (at least once), waiting
// between attempts. try returns done, or a fatal error that stops at once.
func (d driver) poll(ctx context.Context, timeout time.Duration, try func() (bool, error)) (bool, error) {
	attempts := 1
	if d.cfg.PollInterval > 0 {
		attempts = max(1, int(timeout/d.cfg.PollInterval))
	}
	for i := range attempts {
		if i > 0 {
			if err := d.wait(ctx, d.cfg.PollInterval); err != nil {
				return false, err
			}
		}
		if err := ctx.Err(); err != nil {
			return false, err
		}
		done, err := try()
		if done || err != nil {
			return done, err
		}
	}
	return false, nil
}

func (d driver) ready(ctx context.Context) error {
	var body struct{ Status, Version string }
	ok, err := d.poll(ctx, d.cfg.ReadyTimeout, func() (bool, error) {
		status, raw, err := d.do(ctx, http.MethodGet, "/ready", nil, "")
		if err != nil || status != http.StatusOK || json.Unmarshal(raw, &body) != nil {
			return false, nil // still starting
		}
		return body.Status == "ok", nil
	})
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("contract: host not ready within %s", d.cfg.ReadyTimeout)
	}
	if body.Version != d.cfg.HostVersion {
		return fmt.Errorf("contract: host version is %q, want exactly %q (min_kandev_version)", body.Version, d.cfg.HostVersion)
	}
	return nil
}

func (d driver) install(ctx context.Context) error {
	pkg, err := os.ReadFile(d.cfg.PackagePath)
	if err != nil {
		return fmt.Errorf("contract: read package: %w", err)
	}
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	part, err := mw.CreateFormFile("package", filepath.Base(d.cfg.PackagePath))
	if err != nil {
		return fmt.Errorf("contract: build upload: %w", err)
	}
	if _, err := part.Write(pkg); err != nil {
		return fmt.Errorf("contract: build upload: %w", err)
	}
	if err := mw.Close(); err != nil {
		return fmt.Errorf("contract: build upload: %w", err)
	}
	status, raw, err := d.do(ctx, http.MethodPost, "/api/plugins/install", &buf, mw.FormDataContentType())
	if err != nil {
		return err
	}
	var reply struct {
		Error   string `json:"error"`
		Warning string `json:"warning"`
	}
	_ = json.Unmarshal(raw, &reply)
	switch {
	case status != http.StatusCreated:
		return fmt.Errorf("contract: install answered %d: %s", status, message(reply.Error, raw))
	case reply.Warning != "":
		return fmt.Errorf("contract: install answered with a warning: %s", reply.Warning)
	}
	return nil
}

func (d driver) active(ctx context.Context) error {
	var rec struct {
		Status    string `json:"status"`
		LastError string `json:"last_error"`
	}
	ok, err := d.poll(ctx, d.cfg.ActiveTimeout, func() (bool, error) {
		status, raw, err := d.do(ctx, http.MethodGet, "/api/plugins/"+d.cfg.PluginID, nil, "")
		if err != nil {
			return false, err
		}
		if status != http.StatusOK || json.Unmarshal(raw, &rec) != nil {
			return false, fmt.Errorf("contract: plugin status answered %d: %s", status, message("", raw))
		}
		if rec.Status == "error" {
			return false, fmt.Errorf("contract: plugin status is error: %s", rec.LastError)
		}
		return rec.Status == "active", nil
	})
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("contract: plugin not active within %s (status %q)", d.cfg.ActiveTimeout, rec.Status)
	}
	return nil
}

func (d driver) actions(ctx context.Context) error {
	workspace, err := d.firstWorkspace(ctx)
	if err != nil {
		return err
	}
	status, raw, err := d.action(ctx, workspace, "connection.get", map[string]any{})
	if err != nil {
		return err
	}
	var view struct {
		State   string `json:"state"`
		Enabled bool   `json:"enabled"`
	}
	if status != http.StatusOK || json.Unmarshal(raw, &view) != nil {
		return fmt.Errorf("contract: connection.get answered %d", status)
	}
	if view.State != "not_connected" || !view.Enabled {
		return fmt.Errorf("contract: connection.get gave state %q enabled %t, want not_connected enabled true", view.State, view.Enabled)
	}
	return d.connectValidation(ctx, workspace)
}

func (d driver) firstWorkspace(ctx context.Context) (string, error) {
	status, raw, err := d.do(ctx, http.MethodGet, "/api/v1/workspaces", nil, "")
	if err != nil {
		return "", err
	}
	var list struct {
		Workspaces []struct {
			ID string `json:"id"`
		} `json:"workspaces"`
	}
	if status != http.StatusOK || json.Unmarshal(raw, &list) != nil {
		return "", fmt.Errorf("contract: list workspaces answered %d", status)
	}
	if len(list.Workspaces) == 0 || list.Workspaces[0].ID == "" {
		return "", errors.New("contract: the host has no workspace")
	}
	return list.Workspaces[0].ID, nil
}

// connectValidation sends a bait key with a bad space address. The reply is
// never echoed, and the key is redacted from the message (project.md Mandated).
func (d driver) connectValidation(ctx context.Context, workspace string) error {
	key, err := baitKey()
	if err != nil {
		return err
	}
	status, raw, err := d.action(ctx, workspace, "connection.connect_api_key", map[string]any{"spaceUrl": invalidSpaceURL, "apiKey": key})
	if err != nil {
		return err
	}
	var reply struct {
		Error struct{ Code, Field string } `json:"error"`
	}
	_ = json.Unmarshal(raw, &reply)
	if status == http.StatusBadRequest && reply.Error.Code == "validation" && reply.Error.Field == "spaceUrl" {
		return nil
	}
	msg := fmt.Sprintf("contract: connection.connect_api_key answered %d code %q field %q, want 400 validation spaceUrl",
		status, reply.Error.Code, reply.Error.Field)
	return errors.New(strings.ReplaceAll(msg, key, "[redacted]"))
}

func baitKey() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("contract: generate bait key: %w", err)
	}
	return allowedPrefix + hex.EncodeToString(b), nil
}

func (d driver) action(ctx context.Context, workspace, key string, body any) (int, []byte, error) {
	env, err := json.Marshal(map[string]any{"workspaceId": workspace, "body": body})
	if err != nil {
		return 0, nil, fmt.Errorf("contract: encode %s: %w", key, err)
	}
	return d.do(ctx, http.MethodPost, "/api/plugins/"+d.cfg.PluginID+"/actions/"+key, bytes.NewReader(env), "application/json")
}

// do sends one request and reads at most maxResponse bytes of the reply.
func (d driver) do(ctx context.Context, method, path string, body io.Reader, contentType string) (int, []byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, strings.TrimRight(d.cfg.BaseURL, "/")+path, body)
	if err != nil {
		return 0, nil, fmt.Errorf("contract: %s %s: %w", method, path, err)
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	resp, err := d.client.Do(req)
	if err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return 0, nil, ctxErr
		}
		return 0, nil, fmt.Errorf("contract: %s %s: %w", method, path, err)
	}
	defer func() { _ = resp.Body.Close() }()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, maxResponse))
	if err != nil {
		return 0, nil, fmt.Errorf("contract: read %s %s: %w", method, path, err)
	}
	return resp.StatusCode, raw, nil
}

// message is Kandev's error text, or the start of the raw reply.
func message(errText string, raw []byte) string {
	if errText != "" {
		return errText
	}
	if len(raw) > 300 {
		raw = raw[:300]
	}
	return strings.TrimSpace(string(raw))
}
