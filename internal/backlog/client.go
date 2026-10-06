package backlog

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/redact"
)

const (
	// maxBody is the response size limit (BR4.2).
	maxBody = 1 << 20
	// defaultRateLimitWait applies when Backlog sends no wait header (AC8.4.1).
	defaultRateLimitWait = 60 * time.Second
	// maxRetries is the number of retries after a 429 (AC8.4.2).
	maxRetries = 3
	// interactiveBudget is the most an Interactive call waits on queues and
	// 429s in total.
	interactiveBudget = 3 * time.Second
	// groupSpacing is the gap between two Search or Update calls (AC8.4.3).
	groupSpacing = time.Second

	myselfPath   = "/api/v2/users/myself"
	projectsPath = "/api/v2/projects"
	tokenPath    = "/api/v2/oauth2/token" //nolint:gosec // G101: an API path, not a credential
)

// Client calls the Backlog API v2. It holds no credentials. Search and
// Update calls run one at a time per space host and group, 1 s apart.
type Client struct {
	http    *http.Client
	Timeout time.Duration                                    // limit per attempt (BR4.1); callers may set a shorter context deadline
	Now     func() time.Time                                 // clock for X-RateLimit-Reset and queue spacing
	Wait    func(ctx context.Context, d time.Duration) error // waits d or until ctx ends

	mu    sync.Mutex
	slots map[slotKey]*slot
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
		Wait:    sleepCtx,
		slots:   map[slotKey]*slot{},
	}
}

// sleepCtx waits d, or returns the context's error when it ends first.
func sleepCtx(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
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
	body, err := c.send(ctx, request{method: http.MethodGet, path: myselfPath, creds: creds, class: Interactive})
	if err != nil {
		return User{}, err
	}
	return parseUser(body)
}

// Projects returns the projects the credentials can see (GET /api/v2/projects).
func (c *Client) Projects(ctx context.Context, creds Credentials) ([]Project, error) {
	body, err := c.send(ctx, request{method: http.MethodGet, path: projectsPath, creds: creds, class: Interactive})
	if err != nil {
		return nil, err
	}
	return parseProjects(body)
}

// Repositories returns a project's Git repositories (GET
// /api/v2/projects/:key/git/repositories). Each httpUrl must be https on the space host.
func (c *Client) Repositories(ctx context.Context, creds Credentials, projectKey string) ([]Repository, error) {
	body, err := c.send(ctx, request{method: http.MethodGet, path: repositoriesPath(projectKey), creds: creds, class: Interactive})
	if err != nil {
		return nil, err
	}
	return parseRepositories(body, creds.SpaceHost)
}

// PullRequests lists a repository's pull requests, newest first.
func (c *Client) PullRequests(ctx context.Context, creds Credentials, class CallClass, projectKey, repo string, q PullRequestQuery) ([]PullRequest, error) {
	body, err := c.send(ctx, request{method: http.MethodGet, path: pullRequestsPath(projectKey, repo), query: q.Values(), creds: creds, class: class})
	if err != nil {
		return nil, err
	}
	return parsePullRequests(body)
}

// PullRequest returns one pull request by number.
func (c *Client) PullRequest(ctx context.Context, creds Credentials, class CallClass, projectKey, repo string, number int) (PullRequest, error) {
	body, err := c.send(ctx, request{method: http.MethodGet, path: fmt.Sprintf("%s/%d", pullRequestsPath(projectKey, repo), number), creds: creds, class: class})
	if err != nil {
		return PullRequest{}, err
	}
	return parsePullRequest(body)
}

// CreatePullRequest adds a pull request (an Update call, queued and spaced).
func (c *Client) CreatePullRequest(ctx context.Context, creds Credentials, projectKey, repo string, in NewPullRequest) (PullRequest, error) {
	body, err := c.send(ctx, request{method: http.MethodPost, path: pullRequestsPath(projectKey, repo), form: in.Form(), creds: creds, class: Interactive})
	if err != nil {
		return PullRequest{}, err
	}
	return parsePullRequest(body)
}

// CheckGitAccess asks Backlog's Git smart-HTTP endpoint whether username and
// password may fetch the repository. A refusal (401 or 403) is Unauthorized;
// the password and the Basic header value are redacted from logs and errors.
func (c *Client) CheckGitAccess(ctx context.Context, spaceHost, username, password, projectKey, repo string) error {
	_, err := c.send(ctx, request{
		method: http.MethodGet, path: "/git/" + projectKey + "/" + repo + ".git/info/refs",
		query: url.Values{"service": {"git-upload-pack"}}, creds: Credentials{SpaceHost: spaceHost},
		basicUser: username, basicPass: password, statusOnly: true, class: Interactive,
	})
	var be *Error
	if errors.As(err, &be) && be.Kind == KindForbidden {
		be.Kind = KindUnauthorized
	}
	return err
}

func repositoriesPath(projectKey string) string {
	return "/api/v2/projects/" + projectKey + "/git/repositories"
}

func pullRequestsPath(projectKey, repo string) string {
	return repositoriesPath(projectKey) + "/" + repo + "/pullRequests"
}

// ExchangeOAuthCode trades an authorization code for tokens.
func (c *Client) ExchangeOAuthCode(ctx context.Context, spaceHost string, client OAuthClient, code, redirectURI string) (TokenSet, error) {
	return c.token(ctx, spaceHost, url.Values{
		"grant_type": {"authorization_code"}, "code": {code}, "redirect_uri": {redirectURI},
		"client_id": {client.ClientID}, "client_secret": {client.ClientSecret},
	}, client.ClientSecret, code)
}

// RefreshToken trades a refresh token for new tokens; Backlog rotates the
// refresh token on every call.
func (c *Client) RefreshToken(ctx context.Context, spaceHost string, client OAuthClient, refreshToken string) (TokenSet, error) {
	return c.token(ctx, spaceHost, url.Values{
		"grant_type": {"refresh_token"}, "refresh_token": {refreshToken},
		"client_id": {client.ClientID}, "client_secret": {client.ClientSecret},
	}, client.ClientSecret, refreshToken)
}

func (c *Client) token(ctx context.Context, spaceHost string, form url.Values, secrets ...string) (TokenSet, error) {
	body, err := c.send(ctx, request{
		method: http.MethodPost, path: tokenPath, creds: Credentials{SpaceHost: spaceHost},
		form: form, class: Interactive, secrets: secrets,
	})
	if err != nil {
		return TokenSet{}, err
	}
	return parseTokenSet(body, c.Now())
}

// request is one Backlog call.
type request struct {
	method  string
	path    string
	creds   Credentials // SpaceHost plus at most one credential
	form    url.Values  // a form body (token endpoint); nil for none
	class   CallClass   // 0 counts as Interactive (R-07)
	secrets []string    // extra values to redact: client secret, code, refresh token

	// U4 fields.
	query      url.Values // the query string; never logged
	basicUser  string     // HTTP Basic auth for the Git endpoint, instead of creds
	basicPass  string
	statusOnly bool // a 2xx body is not read (Git ref lists can be large)
}

// errOverBudget is returned by pause when an Interactive call may not wait.
var errOverBudget = errors.New("backlog: over the interactive wait budget")

// send runs a call through its group's queue and retries a 429 up to
// maxRetries times, waiting as Backlog asks (AC8.4.1, AC8.4.2).
func (c *Client) send(ctx context.Context, r request) ([]byte, error) {
	u := &url.URL{Scheme: "https", Host: r.creds.SpaceHost, Path: r.path}
	if r.query != nil {
		u.RawQuery = r.query.Encode()
	}
	if r.creds.SpaceHost == "" || u.Hostname() != r.creds.SpaceHost || u.Port() != "" {
		return nil, &Error{Kind: KindUnreachable, Class: "host"}
	}
	ctx = redact.WithSecrets(ctx, r.redactions()...)
	g := group(r.method, r.path)
	b := &budget{limited: r.class != Background, left: interactiveBudget}
	for attempt := 1; ; attempt++ {
		release, err := c.acquire(ctx, r.creds.SpaceHost, g, attempt, b)
		if err != nil {
			return nil, err
		}
		body, err := c.attempt(ctx, r, u)
		release()
		var be *Error
		if !errors.As(err, &be) || be.Kind != KindRateLimited || attempt > maxRetries {
			return body, err
		}
		if err := c.pause(ctx, g, "rate_limited", be.RetryAfter, attempt, b); err != nil {
			if errors.Is(err, errOverBudget) {
				return nil, be
			}
			return nil, err
		}
	}
}

// redactions lists every secret of the request, raw and query-escaped.
func (r request) redactions() []string {
	var out []string
	secrets := append([]string{r.creds.APIKey, r.creds.AccessToken}, r.secrets...)
	if r.basicPass != "" {
		basic := base64.StdEncoding.EncodeToString([]byte(r.basicUser + ":" + r.basicPass))
		secrets = append(secrets, r.basicPass, basic)
	}
	for _, s := range secrets {
		out = append(out, s, url.QueryEscape(s))
	}
	return out
}

// budget is the wait an Interactive call has left; Background is unlimited.
type budget struct {
	limited bool
	left    time.Duration
}

func (b *budget) take(d time.Duration) bool {
	if !b.limited {
		return true
	}
	if d > b.left {
		return false
	}
	b.left -= d
	return true
}

// pause logs one backlog_wait line and waits d, within the budget.
func (c *Client) pause(ctx context.Context, g Group, reason string, d time.Duration, attempt int, b *budget) error {
	if !b.take(d) {
		return errOverBudget
	}
	redact.Logger(ctx).InfoContext(ctx, "backlog wait", "event", "backlog_wait", "group", g.String(),
		"reason", reason, "waitMs", d.Milliseconds(), "attempt", attempt)
	if err := c.Wait(ctx, d); err != nil {
		if errors.Is(ctx.Err(), context.Canceled) {
			return context.Canceled
		}
		return &Error{Kind: KindUnreachable, Class: "timeout"}
	}
	return nil
}

// slot is the FIFO queue of one (host, group): sem holds the one running
// call, and last is guarded by holding sem.
type slot struct {
	sem  chan struct{}
	last time.Time
}

type slotKey struct {
	host  string
	group Group
}

func (c *Client) slot(host string, g Group) *slot {
	c.mu.Lock()
	defer c.mu.Unlock()
	s, ok := c.slots[slotKey{host, g}]
	if !ok {
		s = &slot{sem: make(chan struct{}, 1)}
		c.slots[slotKey{host, g}] = s
	}
	return s
}

// acquire takes the group's slot for one attempt and keeps the 1 s spacing.
// Read calls are not queued. The returned func releases the slot.
// ponytail: one process-wide queue per host; Backlog's limit is per user, so
// two Kandev instances on one account still share it.
func (c *Client) acquire(ctx context.Context, host string, g Group, attempt int, b *budget) (func(), error) {
	if g == GroupRead {
		return func() {}, nil
	}
	s := c.slot(host, g)
	if err := c.lock(ctx, s, b); err != nil {
		return nil, err
	}
	if d := s.last.Add(groupSpacing).Sub(c.Now()); !s.last.IsZero() && d > 0 {
		if err := c.pause(ctx, g, "spacing", d, attempt, b); err != nil {
			<-s.sem
			if errors.Is(err, errOverBudget) {
				return nil, &Error{Kind: KindRateLimited, RetryAfter: groupSpacing, Class: "queue"}
			}
			return nil, err
		}
	}
	s.last = c.Now()
	return func() { <-s.sem }, nil
}

// lock waits for the slot; an Interactive call waits at most its budget.
func (c *Client) lock(ctx context.Context, s *slot, b *budget) error {
	select {
	case s.sem <- struct{}{}:
		return nil
	default:
	}
	var timeout <-chan time.Time
	if b.limited {
		t := time.NewTimer(b.left)
		defer t.Stop()
		timeout = t.C
	}
	start := time.Now()
	select {
	case s.sem <- struct{}{}:
		b.left -= time.Since(start)
		return nil
	case <-ctx.Done():
		if errors.Is(ctx.Err(), context.Canceled) {
			return context.Canceled
		}
		return &Error{Kind: KindUnreachable, Class: "timeout"}
	case <-timeout:
		return &Error{Kind: KindRateLimited, RetryAfter: groupSpacing, Class: "queue"}
	}
}

// attempt performs one HTTP request and returns the body of a 2xx response.
func (c *Client) attempt(ctx context.Context, r request, u *url.URL) ([]byte, error) {
	callCtx, cancel := context.WithTimeout(ctx, c.Timeout)
	defer cancel()
	start := time.Now()
	body, status, err := c.do(callCtx, r, u)
	if err != nil {
		err = c.mapTransportError(ctx, err)
	}
	logCall(ctx, r.method, r.path, status, time.Since(start), err)
	return body, err
}

func (c *Client) do(ctx context.Context, r request, u *url.URL) ([]byte, int, error) {
	var reqBody io.Reader
	if r.form != nil {
		reqBody = strings.NewReader(r.form.Encode())
	}
	req, err := http.NewRequestWithContext(ctx, r.method, u.String(), reqBody)
	if err != nil {
		return nil, 0, &Error{Kind: KindUnreachable, Class: "request"}
	}
	if r.form != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	// One credential per request, never in the URL (NFR3, AC1.1.7).
	switch {
	case r.basicPass != "":
		req.SetBasicAuth(r.basicUser, r.basicPass)
	case r.creds.AccessToken != "":
		req.Header.Set("Authorization", "Bearer "+r.creds.AccessToken)
	case r.creds.APIKey != "":
		req.Header.Set("Backlog-API-Key", r.creds.APIKey)
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
	if r.statusOnly {
		return nil, resp.StatusCode, nil
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

func logCall(ctx context.Context, method, path string, status int, d time.Duration, err error) {
	attrs := []any{"event", "backlog_call", "method", method, "path", path,
		"status", status, "durationMs", d.Milliseconds()}
	var be *Error
	if errors.As(err, &be) {
		attrs = append(attrs, "errorKind", be.Kind.String(), "errorClass", be.Class)
	}
	redact.Logger(ctx).DebugContext(ctx, "backlog call", attrs...)
}
