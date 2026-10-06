package backlog

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/testutil"
)

// baitHandler answers every request with status and a body holding bait.
func baitHandler(t *testing.T, status int, bait string) http.HandlerFunc {
	body := bytes.ReplaceAll(readFixture(t, "error_500_bait.json"), []byte("{{BAIT}}"), []byte(bait))
	return func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(status)
		_, _ = w.Write(body)
	}
}

func kindOf(t *testing.T, err error) Kind {
	t.Helper()
	var be *Error
	require.True(t, errors.As(err, &be), "want *backlog.Error, got %v", err)
	return be.Kind
}

func TestU4_Repositories_CallsTheProjectRepositoryList(t *testing.T) {
	key := testutil.APIKey(t)
	var got *http.Request
	c, _ := fakeBacklog(t, func(w http.ResponseWriter, r *http.Request) {
		got = r.Clone(context.Background())
		_, _ = w.Write(readFixture(t, "repositories_ok.json"))
	})
	repos, err := c.Repositories(context.Background(), creds(key), "PROJ")
	require.NoError(t, err)
	require.Len(t, repos, 2)
	require.Equal(t, "/api/v2/projects/PROJ/git/repositories", got.URL.Path)
	require.Equal(t, key, got.Header.Get("Backlog-API-Key"))
	require.Equal(t, GroupRead, group(http.MethodGet, got.URL.Path), "a Read call is not queued")
}

func TestU4_GitCalls_ErrorsNeverEchoTheBody(t *testing.T) {
	calls := map[string]func(c *Client, ctx context.Context, cr Credentials) error{
		"repositories": func(c *Client, ctx context.Context, cr Credentials) error {
			_, err := c.Repositories(ctx, cr, "PROJ")
			return err
		},
		"pull requests": func(c *Client, ctx context.Context, cr Credentials) error {
			_, err := c.PullRequests(ctx, cr, Interactive, "PROJ", "web-app", PullRequestQuery{Count: 20})
			return err
		},
		"pull request": func(c *Client, ctx context.Context, cr Credentials) error {
			_, err := c.PullRequest(ctx, cr, Interactive, "PROJ", "web-app", 42)
			return err
		},
		"issue": func(c *Client, ctx context.Context, cr Credentials) error {
			_, err := c.Issue(ctx, cr, Interactive, "PROJ-120")
			return err
		},
	}
	for name, call := range calls {
		for status, kind := range map[int]Kind{401: KindUnauthorized, 403: KindForbidden, 404: KindNotFound, 500: KindUnreachable} {
			t.Run(name+" "+http.StatusText(status), func(t *testing.T) {
				bait, key := testutil.Token(t), testutil.APIKey(t)
				c, _ := fakeBacklog(t, baitHandler(t, status, bait))
				ctx, logs := oauthLogs()
				err := call(c, ctx, creds(key))
				require.Equal(t, kind, kindOf(t, err))
				testutil.AssertNoLeak(t, err.Error()+logs.String(), bait, 8)
				testutil.AssertNoLeak(t, err.Error()+logs.String(), key, 8)
			})
		}
	}
}

func TestU4_PullRequests_SendsTheExactQuery(t *testing.T) {
	var got *http.Request
	c, _ := fakeBacklog(t, func(w http.ResponseWriter, r *http.Request) {
		got = r.Clone(context.Background())
		_, _ = w.Write(readFixture(t, "pullrequests_ok.json"))
	})
	ctx, logs := oauthLogs()
	prs, err := c.PullRequests(ctx, creds(testutil.APIKey(t)), Background, "PROJ", "web-app",
		PullRequestQuery{StatusIDs: []int64{1}, AssigneeIDs: []int64{7}, Count: 100})
	require.NoError(t, err)
	require.Len(t, prs, 3)
	require.Equal(t, "/api/v2/projects/PROJ/git/repositories/web-app/pullRequests", got.URL.Path)
	require.Equal(t, "assigneeId%5B%5D=7&count=100&statusId%5B%5D=1", got.URL.RawQuery)
	require.NotContains(t, logs.String(), "assigneeId", "the query is never logged")
}

func TestU4_PullRequests_BackgroundWaitsInteractiveDoesNot(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var hits atomic.Int32
		c := memClient(func(w http.ResponseWriter, _ *http.Request) {
			if hits.Add(1) == 1 {
				tooMany(w, map[string]string{"Retry-After": "40"})
				return
			}
			_, _ = w.Write([]byte(`[]`))
		})
		start := time.Now()
		_, err := c.PullRequests(context.Background(), creds("k-1234"), Background, "PROJ", "web-app", PullRequestQuery{Count: 10})
		require.NoError(t, err)
		require.Equal(t, 40*time.Second, time.Since(start), "Background waits the full 429 time")

		hits.Store(0)
		start = time.Now()
		_, err = c.PullRequests(context.Background(), creds("k-1234"), Interactive, "PROJ", "web-app", PullRequestQuery{Count: 10})
		require.Equal(t, KindRateLimited, kindOf(t, err))
		require.Zero(t, time.Since(start), "Interactive keeps its 3 s budget")
	})
}

func TestU4_PullRequest_GetsOneAndMapsNotFound(t *testing.T) {
	var path string
	c, _ := fakeBacklog(t, func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		_, _ = w.Write(readFixture(t, "pullrequest_ok.json"))
	})
	pr, err := c.PullRequest(context.Background(), creds(testutil.APIKey(t)), Interactive, "PROJ", "web-app", 42)
	require.NoError(t, err)
	require.Equal(t, 42, pr.Number)
	require.Equal(t, "/api/v2/projects/PROJ/git/repositories/web-app/pullRequests/42", path)

	c, _ = fakeBacklog(t, serveFixture(t, 404, "error_401.json"))
	_, err = c.PullRequest(context.Background(), creds(testutil.APIKey(t)), Interactive, "PROJ", "web-app", 999)
	require.Equal(t, KindNotFound, kindOf(t, err))
}

func TestU4_CreatePR_PostsTheExactForm(t *testing.T) {
	var got *http.Request
	var form map[string][]string
	c, _ := fakeBacklog(t, func(w http.ResponseWriter, r *http.Request) {
		got = r.Clone(context.Background())
		require.NoError(t, r.ParseForm())
		form = r.PostForm
		_, _ = w.Write(readFixture(t, "pullrequest_created.json"))
	})
	pr, err := c.CreatePullRequest(context.Background(), creds(testutil.APIKey(t)), "PROJ", "web-app",
		NewPullRequest{Summary: "Add search", Description: "Related: PROJ-120", Base: "main", Branch: "feature/search", IssueID: 5120})
	require.NoError(t, err)
	require.Equal(t, 43, pr.Number)
	require.Equal(t, http.MethodPost, got.Method)
	require.Equal(t, "application/x-www-form-urlencoded", got.Header.Get("Content-Type"))
	require.Equal(t, "/api/v2/projects/PROJ/git/repositories/web-app/pullRequests", got.URL.Path)
	require.Equal(t, map[string][]string{"summary": {"Add search"}, "description": {"Related: PROJ-120"},
		"base": {"main"}, "branch": {"feature/search"}, "issueId": {"5120"}}, form)

	c, _ = fakeBacklog(t, serveFixture(t, 400, "error_401.json"))
	_, err = c.CreatePullRequest(context.Background(), creds(testutil.APIKey(t)), "PROJ", "web-app", NewPullRequest{Summary: "x"})
	require.Equal(t, KindInvalid, kindOf(t, err))
}

func TestU4_CreatePR_QueuedOneSecondFromARefresh(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		rec := newGroupRecorder()
		c := memClient(func(w http.ResponseWriter, r *http.Request) {
			rec.handler(100*time.Millisecond)(httptestDiscard{}, r)
			if r.URL.Path == tokenPath {
				_, _ = w.Write([]byte(`{"access_token":"a-1234","token_type":"Bearer","expires_in":3600,"refresh_token":"r-1234"}`))
				return
			}
			_, _ = w.Write(readFixture(t, "pullrequest_created.json"))
		})
		var wg sync.WaitGroup
		wg.Go(func() {
			_, err := c.RefreshToken(context.Background(), spaceHost, OAuthClient{ClientID: "id", ClientSecret: "s-1234"}, "r-0000")
			require.NoError(t, err)
		})
		wg.Go(func() {
			_, err := c.CreatePullRequest(context.Background(), creds("k-1234"), "PROJ", "web-app", NewPullRequest{Summary: "x", Base: "main", Branch: "b"})
			require.NoError(t, err)
		})
		wg.Wait()
		key := spaceHost + " update"
		require.Equal(t, 1, rec.maxIn[key], "never in parallel")
		require.Len(t, rec.starts[key], 2)
		require.GreaterOrEqual(t, rec.starts[key][1].Sub(rec.starts[key][0]), time.Second)
	})
}

// httptestDiscard is a ResponseWriter that drops everything.
type httptestDiscard struct{}

func (httptestDiscard) Header() http.Header         { return http.Header{} }
func (httptestDiscard) Write(b []byte) (int, error) { return len(b), nil }
func (httptestDiscard) WriteHeader(int)             {}

func TestU4_Issue_GetsByKey(t *testing.T) {
	var path string
	c, _ := fakeBacklog(t, func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		_, _ = w.Write(readFixture(t, "issue_ok.json"))
	})
	issue, err := c.Issue(context.Background(), creds(testutil.APIKey(t)), Interactive, "PROJ-120")
	require.NoError(t, err)
	require.Equal(t, int64(5120), issue.ID)
	require.Equal(t, "/api/v2/issues/PROJ-120", path)
	require.Equal(t, GroupRead, group(http.MethodGet, path))
}

func TestU4_GitAccess_ProbesWithBasicAuth(t *testing.T) {
	pw := testutil.Token(t)
	var got *http.Request
	c, _ := fakeBacklog(t, func(w http.ResponseWriter, r *http.Request) {
		got = r.Clone(context.Background())
		w.Header().Set("Content-Type", "application/x-git-upload-pack-advertisement")
		_, _ = w.Write(bytes.Repeat([]byte("0"), maxBody+10)) // a large ref list is fine: only the status matters
	})
	require.NoError(t, c.CheckGitAccess(context.Background(), spaceHost, "lan", pw, "PROJ", "web-app"))
	require.NotNil(t, got.TLS)
	require.Equal(t, "/git/PROJ/web-app.git/info/refs", got.URL.Path)
	require.Equal(t, "service=git-upload-pack", got.URL.RawQuery)
	user, pass, ok := got.BasicAuth()
	require.True(t, ok)
	require.Equal(t, "lan", user)
	require.Equal(t, pw, pass)
	require.Empty(t, got.Header.Get("Backlog-API-Key"))
}

func TestU4_GitAccess_RefusedIsUnauthorizedAndRedacted(t *testing.T) {
	for _, status := range []int{401, 403} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			pw := testutil.Token(t)
			basic := base64.StdEncoding.EncodeToString([]byte("lan:" + pw))
			c, _ := fakeBacklog(t, func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(status)
				_, _ = w.Write([]byte("denied " + basic)) // a server that echoes the header
			})
			ctx, logs := oauthLogs()
			err := c.CheckGitAccess(ctx, spaceHost, "lan", pw, "PROJ", "web-app")
			require.Equal(t, KindUnauthorized, kindOf(t, err))
			testutil.AssertNoLeak(t, err.Error()+logs.String(), pw, 8, basic)
			require.Contains(t, logs.String(), "backlog_call")
		})
	}
	c, _ := fakeBacklog(t, baitHandler(t, 500, "x"))
	err := c.CheckGitAccess(context.Background(), spaceHost, "lan", testutil.Token(t), "PROJ", "web-app")
	require.Equal(t, KindUnreachable, kindOf(t, err))
}
