package backlog

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/testutil"
)

// BR3.5, BR3.12: the issue watch lists issues oldest created first, from a
// createdSince date, filtered by creator; other callers keep updated/desc.
func TestIssueQuery_EncodesTheWatchOptions(t *testing.T) {
	watch := IssueQuery{ProjectIDs: []int64{101}, StatusIDs: []int64{1}, CreatedUserIDs: []int64{7},
		CreatedSince: "2026-10-06", Sort: "created", Order: "asc", Offset: 95, Count: 100}
	require.Equal(t, "count=100&createdSince=2026-10-06&createdUserId%5B%5D=7&offset=95&order=asc&projectId%5B%5D=101&sort=created&statusId%5B%5D=1",
		watch.Values().Encode())
	require.Equal(t, "count=20&order=desc&projectId%5B%5D=101&sort=updated",
		IssueQuery{ProjectIDs: []int64{101}, Count: 20}.Values().Encode(), "the default order is unchanged")
}

func TestIssues_ReturnTheCreatedTime(t *testing.T) {
	c, _ := fakeBacklog(t, serveFixture(t, 200, "issues_ok.json"))
	issues, err := c.Issues(context.Background(), creds(testutil.APIKey(t)), Background, IssueQuery{Count: 100})
	require.NoError(t, err)
	require.Equal(t, "2026-09-20T09:00:00Z", issues[0].Created)
}

// BR4.1: the PR list total comes from the pull request count call, with the
// list's filters and no paging.
func TestPullRequestCount_SendsTheFilters(t *testing.T) {
	key := testutil.APIKey(t)
	var got *http.Request
	c, _ := fakeBacklog(t, func(w http.ResponseWriter, r *http.Request) {
		got = r.Clone(context.Background())
		_, _ = w.Write(readFixture(t, "pullrequests_count_ok.json"))
	})
	n, err := c.PullRequestCount(context.Background(), creds(key), Interactive, "PROJ", "web-app",
		PullRequestQuery{StatusIDs: []int64{1}, CreatedUserIDs: []int64{7}, Offset: 20, Count: 20})
	require.NoError(t, err)
	require.Equal(t, 45, n)
	require.Equal(t, http.MethodGet, got.Method)
	require.Equal(t, "/api/v2/projects/PROJ/git/repositories/web-app/pullRequests/count", got.URL.Path)
	require.Equal(t, "createdUserId%5B%5D=7&statusId%5B%5D=1", got.URL.RawQuery)
	require.Equal(t, key, got.Header.Get("Backlog-API-Key"))
}

// NFR3: 401 and 429 (with Retry-After) map to the client's error type.
func TestPullRequestCount_MapsErrors(t *testing.T) {
	c, _ := fakeBacklog(t, serveFixture(t, 401, "error_401.json"))
	_, err := c.PullRequestCount(context.Background(), creds(testutil.APIKey(t)), Interactive, "PROJ", "web-app", PullRequestQuery{})
	require.Equal(t, KindUnauthorized, kindOf(t, err))

	c, _ = fakeBacklog(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Retry-After", "40")
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write(readFixture(t, "error_429.json"))
	})
	_, err = c.PullRequestCount(context.Background(), creds(testutil.APIKey(t)), Interactive, "PROJ", "web-app", PullRequestQuery{})
	var be *Error
	require.ErrorAs(t, err, &be)
	require.Equal(t, KindRateLimited, be.Kind)
	require.Equal(t, int64(40), int64(be.RetryAfter.Seconds()))
	_, err = c.PullRequestCount(context.Background(), creds("k"), Interactive, "proj", "web-app", PullRequestQuery{})
	require.Equal(t, KindInvalid, kindOf(t, err), "a bad project key never reaches a request path")
}
