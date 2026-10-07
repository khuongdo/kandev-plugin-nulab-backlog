package scm

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/redact"
)

const (
	// maxBody is the response size limit (NFR4): 100 repositories or pull
	// requests of GitHub's verbose JSON fit well within it.
	maxBody = 8 << 20
	// defaultRateLimitWait applies when a 429 names no wait.
	defaultRateLimitWait = 60 * time.Second
)

// API is the stdlib HTTP helper the three clients share (NFR6): GET only,
// https to one fixed host, no redirects, a size-limited body, no retries.
type API struct {
	Provider Provider
	// BaseURL is the fixed https origin plus path prefix, for example
	// https://gitlab.com/api/v4. Tests point it at a fake server.
	BaseURL string
	HTTP    *http.Client
	Now     func() time.Time // clock for rate-limit reset headers
	Auth    func(*http.Request, Credential)
}

// NewAPI returns the production helper: TLS 1.2 or later, a 10-second call
// limit, and redirects returned instead of followed (NFR2).
func NewAPI(p Provider, baseURL string, auth func(*http.Request, Credential)) *API {
	tr := http.DefaultTransport.(*http.Transport).Clone()
	tr.TLSClientConfig = &tls.Config{MinVersion: tls.VersionTLS12}
	return &API{
		Provider: p, BaseURL: baseURL, Now: time.Now, Auth: auth,
		HTTP: &http.Client{Transport: tr, Timeout: 10 * time.Second, CheckRedirect: noRedirect},
	}
}

func noRedirect(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }

// Get calls BaseURL+path (path already escaped) and decodes a 2xx JSON body
// into out (nil skips it). Any other status is an *HTTPError; the caller's
// cancellation or deadline is returned unchanged.
func (a *API) Get(ctx context.Context, cred Credential, path string, query url.Values, out any) (http.Header, error) {
	u, err := a.target(path)
	if err != nil {
		return nil, err
	}
	u.RawQuery = query.Encode()
	ctx = redact.WithSecrets(ctx, cred.Token)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, &HTTPError{Provider: a.Provider}
	}
	req.Header.Set("Accept", "application/json")
	if a.Auth != nil {
		a.Auth(req, cred)
	}
	start := time.Now()
	resp, err := a.HTTP.Do(req)
	if err != nil {
		a.log(ctx, path, 0, start)
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, &HTTPError{Provider: a.Provider} // never the *url.Error, which holds the URL
	}
	defer func() { _ = resp.Body.Close() }()
	a.log(ctx, path, resp.StatusCode, start)
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, maxBody)) // drained, never echoed (NFR5)
		return nil, a.statusError(resp)
	}
	if out == nil {
		return resp.Header, nil
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBody+1))
	if err != nil || len(body) > maxBody || json.Unmarshal(body, out) != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, &HTTPError{Provider: a.Provider}
	}
	return resp.Header, nil
}

// target parses BaseURL+path and refuses anything that is not https on
// BaseURL's own host, so no path can steer a request elsewhere (NFR2).
func (a *API) target(path string) (*url.URL, error) {
	base, err := url.Parse(a.BaseURL)
	if err != nil || base.Scheme != "https" {
		return nil, ErrHostRefused
	}
	u, err := url.Parse(a.BaseURL + path)
	if err != nil || u.Scheme != "https" || u.Host != base.Host || u.User != nil {
		return nil, ErrHostRefused
	}
	return u, nil
}

// statusError maps a non-2xx response; GitHub's 403 with no quota left is a
// rate limit like 429 (NFR3).
func (a *API) statusError(resp *http.Response) error {
	status := resp.StatusCode
	if status == http.StatusForbidden && resp.Header.Get("X-RateLimit-Remaining") == "0" {
		status = http.StatusTooManyRequests
	}
	e := &HTTPError{Provider: a.Provider, Status: status}
	if status == http.StatusTooManyRequests {
		e.RetryAfter = a.rateLimitWait(resp.Header)
	}
	return e
}

// rateLimitWait reads Retry-After (seconds), then X-RateLimit-Reset or
// RateLimit-Reset (epoch seconds), else 60 s; never less than 1 s.
func (a *API) rateLimitWait(h http.Header) time.Duration {
	if secs, err := strconv.Atoi(h.Get("Retry-After")); err == nil {
		return max(time.Duration(secs)*time.Second, time.Second)
	}
	for _, k := range []string{"X-RateLimit-Reset", "RateLimit-Reset"} {
		if reset, err := strconv.ParseInt(h.Get(k), 10, 64); err == nil {
			return max(time.Unix(reset, 0).Sub(a.Now()).Round(time.Second), time.Second)
		}
	}
	return defaultRateLimitWait
}

// log writes one debug scm_call line: no query string, no header, no body.
func (a *API) log(ctx context.Context, path string, status int, start time.Time) {
	redact.Logger(ctx).DebugContext(ctx, "scm call", "event", "scm_call", "provider", string(a.Provider),
		"path", path, "status", status, "durationMs", time.Since(start).Milliseconds())
}

// IsStatus reports whether err is an HTTPError with the given status.
func IsStatus(err error, status int) bool {
	var he *HTTPError
	return errors.As(err, &he) && he.Status == status
}
