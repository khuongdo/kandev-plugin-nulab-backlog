package backlog

import (
	"context"
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

// getOnly fails the test on any request that is not a GET: U3 never writes
// to Backlog (AC3.2.3, FR4.3).
func getOnly(t *testing.T, h http.HandlerFunc) http.HandlerFunc {
	var writes atomic.Int32
	t.Cleanup(func() { require.Zero(t, writes.Load(), "U3 sent a non-GET request") })
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writes.Add(1)
		}
		h(w, r)
	}
}

// u3Calls is every U3 Backlog call, with its expected path.
var u3Calls = map[string]struct {
	path string
	call func(c *Client, ctx context.Context, cr Credentials) error
}{
	"issues": {"/api/v2/issues", func(c *Client, ctx context.Context, cr Credentials) error {
		_, err := c.Issues(ctx, cr, Interactive, IssueQuery{Count: 20})
		return err
	}},
	"count": {"/api/v2/issues/count", func(c *Client, ctx context.Context, cr Credentials) error {
		_, err := c.IssueCount(ctx, cr, Interactive, IssueQuery{})
		return err
	}},
	"issue": {"/api/v2/issues/PROJ-118", func(c *Client, ctx context.Context, cr Credentials) error {
		_, err := c.Issue(ctx, cr, Interactive, "PROJ-118")
		return err
	}},
	"comments": {"/api/v2/issues/PROJ-118/comments", func(c *Client, ctx context.Context, cr Credentials) error {
		_, err := c.IssueComments(ctx, cr, Interactive, "PROJ-118", CommentQuery{Count: 20})
		return err
	}},
	"attachments": {"/api/v2/issues/PROJ-118/attachments", func(c *Client, ctx context.Context, cr Credentials) error {
		_, err := c.IssueAttachments(ctx, cr, Interactive, "PROJ-118")
		return err
	}},
	"statuses": {"/api/v2/projects/PROJ/statuses", func(c *Client, ctx context.Context, cr Credentials) error {
		_, err := c.ProjectStatuses(ctx, cr, "PROJ")
		return err
	}},
	"users": {"/api/v2/projects/PROJ/users", func(c *Client, ctx context.Context, cr Credentials) error {
		_, err := c.ProjectUsers(ctx, cr, "PROJ")
		return err
	}},
}

func TestU3_Issues_SendsTheExactQuery(t *testing.T) {
	key := testutil.APIKey(t)
	var got []*http.Request
	c, _ := fakeBacklog(t, getOnly(t, func(w http.ResponseWriter, r *http.Request) {
		got = append(got, r.Clone(context.Background()))
		if r.URL.Path == "/api/v2/issues/count" {
			_, _ = w.Write(readFixture(t, "issues_count_ok.json"))
			return
		}
		_, _ = w.Write(readFixture(t, "issues_ok.json"))
	}))
	ctx, logs := oauthLogs()
	q := IssueQuery{ProjectIDs: []int64{101}, StatusIDs: []int64{2}, Keyword: "ログイン & co", Offset: 20, Count: 20}
	issues, err := c.Issues(ctx, creds(key), Interactive, q)
	require.NoError(t, err)
	require.Len(t, issues, 3)
	n, err := c.IssueCount(ctx, creds(key), Interactive, q)
	require.NoError(t, err)
	require.Equal(t, 57, n)
	require.Equal(t, "/api/v2/issues", got[0].URL.Path)
	require.Equal(t, q.Values().Encode(), got[0].URL.RawQuery)
	require.Equal(t, "keyword=%E3%83%AD%E3%82%B0%E3%82%A4%E3%83%B3+%26+co&projectId%5B%5D=101&statusId%5B%5D=2",
		got[1].URL.RawQuery, "the count takes only the filters")
	require.Equal(t, key, got[0].Header.Get("Backlog-API-Key"))
	require.NotContains(t, logs.String(), "keyword", "the query is never logged")
	require.Equal(t, GroupSearch, group(http.MethodGet, got[0].URL.Path))
	require.Equal(t, GroupSearch, group(http.MethodGet, got[1].URL.Path))
}

func TestU3_Issues_SearchCallsAreQueued(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		rec := newGroupRecorder()
		c := memClient(getOnly(t, func(w http.ResponseWriter, r *http.Request) {
			rec.handler(300*time.Millisecond)(httptestDiscard{}, r)
			if r.URL.Path == "/api/v2/issues/count" {
				_, _ = w.Write([]byte(`{"count":3}`))
				return
			}
			_, _ = w.Write([]byte(`[]`))
		}))
		var wg sync.WaitGroup
		wg.Go(func() {
			_, err := c.Issues(context.Background(), creds("k-1234"), Interactive, IssueQuery{Count: 20})
			require.NoError(t, err)
		})
		wg.Go(func() {
			_, err := c.IssueCount(context.Background(), creds("k-1234"), Interactive, IssueQuery{})
			require.NoError(t, err)
		})
		wg.Wait()
		key := spaceHost + " search"
		require.Equal(t, 1, rec.maxIn[key], "never in parallel")
		require.Len(t, rec.starts[key], 2)
		require.Equal(t, time.Second, rec.starts[key][1].Sub(rec.starts[key][0]), "1 s apart")
	})
}

func TestU3_ReadCalls_AreNotQueued(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		rec := newGroupRecorder()
		c := memClient(getOnly(t, func(w http.ResponseWriter, r *http.Request) {
			rec.handler(time.Second)(httptestDiscard{}, r)
			if r.URL.Path == "/api/v2/issues/PROJ-118" {
				_, _ = w.Write(readFixture(t, "issue_detail_ok.json"))
				return
			}
			_, _ = w.Write([]byte(`[]`))
		}))
		start := time.Now()
		var wg sync.WaitGroup
		for _, name := range []string{"issue", "comments", "attachments", "statuses", "users"} {
			wg.Go(func() { require.NoError(t, u3Calls[name].call(c, context.Background(), creds("k-1234")), name) })
		}
		wg.Wait()
		require.Equal(t, 5, rec.maxIn[spaceHost+" read"])
		require.Equal(t, time.Second, time.Since(start), "all five ran at once")
	})
}

func TestU3_IssueCalls_UseTheirPaths(t *testing.T) {
	for name, tc := range u3Calls {
		t.Run(name, func(t *testing.T) {
			var path, query string
			c, _ := fakeBacklog(t, getOnly(t, func(w http.ResponseWriter, r *http.Request) {
				path, query = r.URL.Path, r.URL.RawQuery
				switch r.URL.Path {
				case "/api/v2/issues/count":
					_, _ = w.Write([]byte(`{"count":1}`))
				case "/api/v2/issues/PROJ-118":
					_, _ = w.Write(readFixture(t, "issue_detail_ok.json"))
				default:
					_, _ = w.Write([]byte(`[]`))
				}
			}))
			require.NoError(t, tc.call(c, context.Background(), creds(testutil.APIKey(t))))
			require.Equal(t, tc.path, path)
			if name == "comments" {
				require.Equal(t, "count=20&order=desc", query)
			}
		})
	}
}

func TestU3_Comments_PagesNewestFirst(t *testing.T) {
	var query string
	c, _ := fakeBacklog(t, getOnly(t, func(w http.ResponseWriter, r *http.Request) {
		query = r.URL.RawQuery
		_, _ = w.Write(readFixture(t, "comments_ok.json"))
	}))
	got, err := c.IssueComments(context.Background(), creds(testutil.APIKey(t)), Interactive, "PROJ-118", CommentQuery{MaxID: 901, Count: 500})
	require.NoError(t, err)
	require.Len(t, got, 2)
	require.Equal(t, "count=100&maxId=901&order=desc", query)
}

func TestU3_IssueCalls_RefuseBadRefs(t *testing.T) {
	var hits atomic.Int32
	c, _ := fakeBacklog(t, func(http.ResponseWriter, *http.Request) { hits.Add(1) })
	ctx, cr := context.Background(), creds(testutil.APIKey(t))
	for _, ref := range []string{"../users/myself", "PROJ-1/comments", "PROJ-1?x", ""} {
		_, err := c.Issue(ctx, cr, Interactive, ref)
		require.Equal(t, KindInvalid, kindOf(t, err), ref)
		_, err = c.IssueComments(ctx, cr, Interactive, ref, CommentQuery{Count: 1})
		require.Equal(t, KindInvalid, kindOf(t, err), ref)
		_, err = c.IssueAttachments(ctx, cr, Interactive, ref)
		require.Equal(t, KindInvalid, kindOf(t, err), ref)
	}
	for _, project := range []string{"../issues", "proj", ""} {
		_, err := c.ProjectStatuses(ctx, cr, project)
		require.Equal(t, KindInvalid, kindOf(t, err), project)
		_, err = c.ProjectUsers(ctx, cr, project)
		require.Equal(t, KindInvalid, kindOf(t, err), project)
	}
	require.Zero(t, hits.Load(), "a bad reference makes no request")
}

func TestU3_IssueCalls_ErrorsNeverEchoTheBody(t *testing.T) {
	for name, tc := range u3Calls {
		for status, kind := range map[int]Kind{401: KindUnauthorized, 403: KindForbidden, 404: KindNotFound, 500: KindUnreachable} {
			t.Run(name+" "+http.StatusText(status), func(t *testing.T) {
				bait, key := testutil.Token(t), testutil.APIKey(t)
				c, _ := fakeBacklog(t, getOnly(t, baitHandler(t, status, bait)))
				ctx, logs := oauthLogs()
				err := tc.call(c, ctx, creds(key))
				require.Equal(t, kind, kindOf(t, err))
				testutil.AssertNoLeak(t, err.Error()+logs.String(), bait, 8)
				testutil.AssertNoLeak(t, err.Error()+logs.String(), key, 8)
			})
		}
	}
}

func TestU3_Issues_RateLimitByCallClass(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var hits atomic.Int32
		c := memClient(getOnly(t, func(w http.ResponseWriter, _ *http.Request) {
			if hits.Add(1) == 1 {
				tooMany(w, map[string]string{"Retry-After": "40"})
				return
			}
			_, _ = w.Write(readFixture(t, "issue_detail_ok.json"))
		}))
		start := time.Now()
		_, err := c.Issue(context.Background(), creds("k-1234"), Background, "PROJ-118")
		require.NoError(t, err)
		require.Equal(t, 40*time.Second, time.Since(start), "Background waits the full 429 time")

		hits.Store(0)
		start = time.Now()
		_, err = c.Issue(context.Background(), creds("k-1234"), Interactive, "PROJ-118")
		var be *Error
		require.True(t, errors.As(err, &be))
		require.Equal(t, KindRateLimited, be.Kind)
		require.Equal(t, 40*time.Second, be.RetryAfter)
		require.Zero(t, time.Since(start), "Interactive keeps its 3 s budget")
	})
}

func TestU3_Issue_TimeoutDoesNotBlockOthers(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		c := newClient(ctxTransport{h: getOnly(t, func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/api/v2/issues/PROJ-118/attachments" {
				<-r.Context().Done() // never answers
				return
			}
			_, _ = w.Write([]byte(`[]`))
		})})
		start := time.Now()
		done := make(chan error, 1)
		go func() {
			_, err := c.IssueAttachments(context.Background(), creds("k-1234"), Interactive, "PROJ-118")
			done <- err
		}()
		synctest.Wait()
		_, err := c.ProjectStatuses(context.Background(), creds("k-1234"), "PROJ")
		require.NoError(t, err, "another call is still served")
		require.Zero(t, time.Since(start))
		err = <-done
		var be *Error
		require.True(t, errors.As(err, &be))
		require.Equal(t, KindUnreachable, be.Kind)
		require.Equal(t, "timeout", be.Class)
		require.Equal(t, 10*time.Second, time.Since(start), "AC8.1.3: within the 10 s limit")
	})
}

// ctxTransport is handlerTransport that fails like a real transport once
// the request's context has ended.
type ctxTransport struct{ h http.HandlerFunc }

func (t ctxTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	resp, _ := handlerTransport(t).RoundTrip(r)
	if err := r.Context().Err(); err != nil {
		return nil, err
	}
	return resp, nil
}
