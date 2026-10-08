package github

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/scm"
	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/testutil"
)

var _ scm.Client = (*Client)(nil)

type hit struct {
	path  string
	query string
	auth  string
}

// fake serves testdata files by path and records every request.
func fake(t *testing.T, routes map[string]string, status map[string]int, headers map[string]string) (*Client, *[]hit) {
	t.Helper()
	var hits []hit
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits = append(hits, hit{path: r.URL.EscapedPath(), query: r.URL.RawQuery, auth: r.Header.Get("Authorization")})
		file, ok := routes[r.URL.EscapedPath()]
		if !ok {
			file = "error_404.json"
			w.WriteHeader(404)
		} else if s := status[r.URL.EscapedPath()]; s != 0 {
			for k, v := range headers {
				w.Header().Set(k, v)
			}
			w.WriteHeader(s)
		}
		b, err := os.ReadFile(filepath.Join("testdata", file)) //nolint:gosec // fixture names are test constants
		require.NoError(t, err)
		_, _ = w.Write(b)
	}))
	t.Cleanup(srv.Close)
	c := New()
	c.API.BaseURL, c.API.HTTP = srv.URL, srv.Client()
	return c, &hits
}

func TestNew_UsesTheFixedHost(t *testing.T) {
	require.Equal(t, "https://api.github.com", New().API.BaseURL) // NFR2
}

// FR2.2, FR2.4: the token goes as a Bearer header; the account is the login.
func TestCurrentUser(t *testing.T) {
	tok := testutil.Token(t)
	c, hits := fake(t, map[string]string{"/user": "user.json"}, nil, nil)
	u, err := c.CurrentUser(context.Background(), scm.Credential{Token: tok})
	require.NoError(t, err)
	require.Equal(t, scm.User{ID: "lan-dev", Name: "Lan Nguyen"}, u)
	require.Equal(t, "Bearer "+tok, (*hits)[0].auth)
}

// FR3.2, NFR4: one page of 100 repositories the token can read, filtered by the search text.
func TestSearchRepos(t *testing.T) {
	c, hits := fake(t, map[string]string{"/user/repos": "repos.json"}, nil, nil)
	repos, err := c.SearchRepos(context.Background(), scm.Credential{Token: "x"}, "WEB")
	require.NoError(t, err)
	require.Equal(t, []scm.Repo{
		{FullName: "acme/web-app", URL: "https://github.com/acme/web-app"},
		{FullName: "other/Web-Docs", URL: "https://github.com/other/Web-Docs"},
	}, repos)
	require.Equal(t, "per_page=100&sort=updated", (*hits)[0].query)
	all, err := c.SearchRepos(context.Background(), scm.Credential{Token: "x"}, "")
	require.NoError(t, err)
	require.Len(t, all, 3)
}

func TestGetRepo(t *testing.T) {
	c, _ := fake(t, map[string]string{"/repos/acme/web-app": "repo.json"}, nil, nil)
	r, err := c.GetRepo(context.Background(), scm.Credential{Token: "x"}, "acme/web-app")
	require.NoError(t, err)
	require.Equal(t, scm.Repo{FullName: "acme/web-app", URL: "https://github.com/acme/web-app"}, r)
	_, err = c.GetRepo(context.Background(), scm.Credential{Token: "x"}, "acme/missing")
	var he *scm.HTTPError
	require.ErrorAs(t, err, &he)
	require.Equal(t, 404, he.Status)
}

// FR4.1, A3: states with draft, merged and closed; paging of at most 100.
func TestListPRs(t *testing.T) {
	c, hits := fake(t, map[string]string{"/repos/acme/web-app/pulls": "pulls.json"}, nil, nil)
	page, err := c.ListPRs(context.Background(), scm.Credential{Token: "x"}, "acme/web-app",
		scm.ListQuery{States: []string{"open", "closed", "merged"}, Page: 2, PerPage: 4})
	require.NoError(t, err)
	require.Equal(t, "direction=desc&page=2&per_page=4&sort=updated&state=all", (*hits)[0].query)
	require.True(t, page.HasNext, "a full page may have a next one")
	require.Len(t, page.Items, 4)
	require.Equal(t, scm.PullRequest{PRRef: scm.PRRef{Provider: scm.GitHub, Repo: "acme/web-app", Number: 42},
		Title: "PROJ-12 Add login page", State: "open", Author: "lan-dev", AuthorID: "lan-dev",
		SourceBranch: "feature/PROJ-12-login", TargetBranch: "main", UpdatedAt: "2026-10-06T09:00:00Z",
		URL: "https://github.com/acme/web-app/pull/42"}, page.Items[0])
	var states []string
	for _, pr := range page.Items {
		states = append(states, pr.State)
	}
	require.Equal(t, []string{"open", "draft", "merged", "closed"}, states)

	for _, c2 := range []struct {
		states []string
		param  string
		want   []int
	}{
		{[]string{"open"}, "state=open", []int{42, 41}},
		{[]string{"merged"}, "state=closed", []int{40}},
		{[]string{"closed"}, "state=closed", []int{39}},
	} {
		*hits = nil
		page, err := c.ListPRs(context.Background(), scm.Credential{Token: "x"}, "acme/web-app",
			scm.ListQuery{States: c2.states, Page: 1, PerPage: 20})
		require.NoError(t, err)
		require.Contains(t, (*hits)[0].query, c2.param)
		require.False(t, page.HasNext)
		var nums []int
		for _, pr := range page.Items {
			nums = append(nums, pr.Number)
		}
		require.Equal(t, c2.want, nums, c2.states)
	}
	_, err = c.ListPRs(context.Background(), scm.Credential{Token: "x"}, "acme/web-app",
		scm.ListQuery{States: []string{"open"}, Page: 1, PerPage: 500})
	require.NoError(t, err)
	require.Contains(t, (*hits)[len(*hits)-1].query, "per_page=100", "NFR4: at most 100 per page")
}

func TestGetPR(t *testing.T) {
	c, hits := fake(t, map[string]string{"/repos/acme/web-app/pulls/42": "pull.json"}, nil, nil)
	pr, err := c.GetPR(context.Background(), scm.Credential{Token: "x"}, "acme/web-app", 42)
	require.NoError(t, err)
	require.Equal(t, 42, pr.Number)
	require.Equal(t, "open", pr.State)
	require.Equal(t, "/repos/acme/web-app/pulls/42", (*hits)[0].path)
}

// NFR3, NFR5: 401, 403, 404, 429 and GitHub's 403 rate limit.
func TestErrors(t *testing.T) {
	reset := time.Now().Add(90 * time.Second).Unix()
	for _, c2 := range []struct {
		file    string
		status  int
		headers map[string]string
		want    int
	}{
		{"error_401.json", 401, nil, 401},
		{"error_403_rate_limit.json", 403, nil, 403},
		{"error_404.json", 404, nil, 404},
		{"error_429.json", 429, map[string]string{"Retry-After": "7"}, 429},
		{"error_403_rate_limit.json", 403, map[string]string{"X-RateLimit-Remaining": "0",
			"X-RateLimit-Reset": strconv.FormatInt(reset, 10)}, 429},
	} {
		c, _ := fake(t, map[string]string{"/user": c2.file}, map[string]int{"/user": c2.status}, c2.headers)
		_, err := c.CurrentUser(context.Background(), scm.Credential{Token: "x"})
		var he *scm.HTTPError
		require.ErrorAs(t, err, &he, c2.file)
		require.Equal(t, c2.want, he.Status, c2.file)
		require.NotContains(t, err.Error(), "Bad credentials")
		if c2.want == 429 {
			require.Positive(t, he.RetryAfter)
		}
	}
}
