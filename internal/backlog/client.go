package backlog

import (
	"context"
	"crypto/tls"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/redact"
)

const (
	// maxBody is the response size limit (BR4.2).
	maxBody = 1 << 20
	// defaultRateLimitWait applies when Backlog sends no wait header (BR2.10).
	defaultRateLimitWait = 60 * time.Second
	myselfPath           = "/api/v2/users/myself"
)

// Client calls the Backlog API v2. It holds no credentials.
type Client struct {
	http    *http.Client
	Timeout time.Duration    // limit per call (BR4.1); callers may set a shorter context deadline
	Now     func() time.Time // clock for X-RateLimit-Reset
}

// NewClient returns the production client: one shared http.Client, https
// only, no redirects, 10-second call limit.
func NewClient() *Client {
	tr := http.DefaultTransport.(*http.Transport).Clone()
	tr.TLSClientConfig = &tls.Config{MinVersion: tls.VersionTLS12}
	return newClient(tr)
}

func newClient(tr http.RoundTripper) *Client {
	return &Client{
		http: &http.Client{
			Transport: httpsOnly{next: tr},
			// Never follow a redirect: the 3xx response itself is returned and
			// mapped to Unreachable (BR1.4).
			CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
		},
		Timeout: 10 * time.Second,
		Now:     time.Now,
	}
}

// httpsOnly refuses any request that is not https (NFR3.4).
type httpsOnly struct {
	next http.RoundTripper
}

var errNotHTTPS = errors.New("backlog: refusing a non-https request")

func (t httpsOnly) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.URL.Scheme != "https" {
		return nil, errNotHTTPS
	}
	return t.next.RoundTrip(req)
}

// Myself returns the user that owns the credentials (GET /api/v2/users/myself).
func (c *Client) Myself(ctx context.Context, creds Credentials) (User, error) {
	body, err := c.get(ctx, creds, myselfPath)
	if err != nil {
		return User{}, err
	}
	return parseUser(body)
}

// get performs one GET and returns the body of a 2xx response.
func (c *Client) get(ctx context.Context, creds Credentials, path string) ([]byte, error) {
	u := &url.URL{Scheme: "https", Host: creds.SpaceHost, Path: path}
	if creds.SpaceHost == "" || u.Hostname() != creds.SpaceHost || u.Port() != "" {
		return nil, &Error{Kind: KindUnreachable, Class: "host"}
	}
	u.RawQuery = url.Values{"apiKey": {creds.APIKey}}.Encode()
	ctx = redact.WithSecrets(ctx, creds.APIKey, url.QueryEscape(creds.APIKey))

	callCtx, cancel := context.WithTimeout(ctx, c.Timeout)
	defer cancel()
	start := time.Now()
	body, status, err := c.do(callCtx, u)
	if err != nil {
		err = c.mapTransportError(ctx, err)
	}
	logCall(ctx, path, status, time.Since(start), err)
	return body, err
}

func (c *Client) do(ctx context.Context, u *url.URL) ([]byte, int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, 0, &Error{Kind: KindUnreachable, Class: "request"}
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = resp.Body.Close() }()
	if kind, isErr := kindForStatus(resp.StatusCode); isErr {
		// Drain up to the limit and discard: a body is never echoed (NFR3.6).
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, maxBody))
		e := &Error{Kind: kind, Status: resp.StatusCode, Class: "http"}
		if kind == KindRateLimited {
			e.RetryAfter = c.rateLimitWait(resp.Header)
		}
		if resp.StatusCode >= 300 && resp.StatusCode < 400 {
			e.Class = "redirect"
		}
		return nil, resp.StatusCode, e
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBody+1))
	if err != nil {
		return nil, resp.StatusCode, err
	}
	if len(body) > maxBody {
		return nil, resp.StatusCode, &Error{Kind: KindUnreachable, Status: resp.StatusCode, Class: "body"}
	}
	return body, resp.StatusCode, nil
}

// mapTransportError turns any non-Error failure into Unreachable with only an
// error class, so a *url.Error (which carries the URL and key) never escapes.
// The caller's cancellation is returned unchanged.
func (c *Client) mapTransportError(ctx context.Context, err error) error {
	var be *Error
	if errors.As(err, &be) {
		return be
	}
	if errors.Is(ctx.Err(), context.Canceled) {
		return context.Canceled
	}
	class := "connection"
	var dnsErr *net.DNSError
	var certErr *tls.CertificateVerificationError
	switch {
	case errors.Is(err, context.DeadlineExceeded) || isTimeout(err):
		class = "timeout"
	case errors.As(err, &dnsErr):
		class = "dns"
	case errors.As(err, &certErr):
		class = "tls"
	}
	return &Error{Kind: KindUnreachable, Class: class}
}

func isTimeout(err error) bool {
	var ne net.Error
	return errors.As(err, &ne) && ne.Timeout()
}

// rateLimitWait reads X-RateLimit-Reset (epoch seconds), then Retry-After
// (seconds), else 60 s, and never returns less than 1 s (NFR2.1).
func (c *Client) rateLimitWait(h http.Header) time.Duration {
	wait := defaultRateLimitWait
	if reset, err := strconv.ParseInt(h.Get("X-RateLimit-Reset"), 10, 64); err == nil {
		wait = time.Unix(reset, 0).Sub(c.Now())
	} else if secs, err := strconv.Atoi(h.Get("Retry-After")); err == nil {
		wait = time.Duration(secs) * time.Second
	}
	return max(wait.Round(time.Second), time.Second)
}

// Is lets a timeout Error match context.DeadlineExceeded, so callers can
// still test for a deadline with errors.Is.
func (e *Error) Is(target error) bool {
	return target == context.DeadlineExceeded && e.Class == "timeout"
}

func logCall(ctx context.Context, path string, status int, d time.Duration, err error) {
	attrs := []any{"event", "backlog_call", "method", http.MethodGet, "path", path,
		"status", status, "durationMs", d.Milliseconds()}
	var be *Error
	if errors.As(err, &be) {
		attrs = append(attrs, "errorKind", be.Kind.String(), "errorClass", be.Class)
	}
	redact.Logger(ctx).DebugContext(ctx, "backlog call", attrs...)
}
