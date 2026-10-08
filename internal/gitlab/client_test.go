package gitlab

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/scm"
	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/testutil"
)

var _ scm.Client = (*Client)(nil)

type hit struct{ path, query, token, auth string }

const base = "/api/v4"

func fake(t *testing.T, routes map[string]string, status map[string]int, headers map[string]string) (*Client, *[]hit) {
	t.Helper()
	var hits []hit
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.EscapedPath()
		hits = append(hits, hit{p, r.URL.RawQuery, r.Header.Get("PRIVATE-TOKEN"), r.Header.Get("Authorization")})
		file, ok := routes[p]
		if !ok {
			file = "error_404.json"
			w.WriteHeader(404)
		} else if s := status[p]; s != 0 {
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
	c.API.BaseURL, c.API.HTTP = srv.URL+base, srv.Client()
	return c, &hits
}

func TestNew_UsesTheFixedHost(t *testing.T) {
	require.Equal(t, "https://gitlab.com/api/v4", New().API.BaseURL) // NFR2
}

// FR2.2 (intent 261008-gh-cli-auth): Authorization: Bearer, which works for
// personal access tokens and glab's OAuth tokens; never PRIVATE-TOKEN.
func TestCurrentUser(t *testing.T) {
	tok := testutil.Token(t)
	c, hits := fake(t, map[string]string{base + "/user": "user.json"}, nil, nil)
	u, err := c.CurrentUser(context.Background(), scm.Credential{Token: tok})
	require.NoError(t, err)
	require.Equal(t, scm.User{ID: "lan.dev", Name: "Lan Nguyen"}, u)
	require.Equal(t, "Bearer "+tok, (*hits)[0].auth)
	require.Empty(t, (*hits)[0].token)
}

// FR3.2, NFR4: the projects the token is a member of, searched by GitLab.
func TestSearchRepos(t *testing.T) {
	c, hits := fake(t, map[string]string{base + "/projects": "projects.json"}, nil, nil)
	repos, err := c.SearchRepos(context.Background(), scm.Credential{Token: "x"}, "web")
	require.NoError(t, err)
	require.Equal(t, scm.Repo{FullName: "grp/sub/web-app", URL: "https://gitlab.com/grp/sub/web-app"}, repos[0])
	require.Equal(t, "membership=true&order_by=last_activity_at&per_page=100&search=web&simple=true", (*hits)[0].query)
	_, err = c.SearchRepos(context.Background(), scm.Credential{Token: "x"}, "")
	require.NoError(t, err)
	require.NotContains(t, (*hits)[1].query, "search=")
}

// FR3.2: a group/subgroup path is one URL-encoded id.
func TestGetRepo(t *testing.T) {
	c, hits := fake(t, map[string]string{base + "/projects/grp%2Fsub%2Fweb-app": "project.json"}, nil, nil)
	r, err := c.GetRepo(context.Background(), scm.Credential{Token: "x"}, "grp/sub/web-app")
	require.NoError(t, err)
	require.Equal(t, "grp/sub/web-app", r.FullName)
	require.Equal(t, base+"/projects/grp%2Fsub%2Fweb-app", (*hits)[0].path)
}

// FR4.1, A3: opened, draft, merged, closed; locked counts as closed.
func TestListPRs(t *testing.T) {
	path := base + "/projects/grp%2Fsub%2Fweb-app/merge_requests"
	c, hits := fake(t, map[string]string{path: "merge_requests.json"}, nil, nil)
	page, err := c.ListPRs(context.Background(), scm.Credential{Token: "x"}, "grp/sub/web-app",
		scm.ListQuery{States: []string{"open", "merged", "closed"}, Page: 1, PerPage: 5})
	require.NoError(t, err)
	require.Equal(t, "order_by=updated_at&page=1&per_page=5&sort=desc&state=all", (*hits)[0].query)
	require.True(t, page.HasNext)
	require.Equal(t, scm.PullRequest{PRRef: scm.PRRef{Provider: scm.GitLab, Repo: "grp/sub/web-app", Number: 7},
		Title: "PROJ-12 Add login page", State: "open", Author: "Lan Nguyen", AuthorID: "lan.dev",
		SourceBranch: "feature/PROJ-12", TargetBranch: "main", UpdatedAt: "2026-10-06T09:00:00.000Z",
		URL: "https://gitlab.com/grp/sub/web-app/-/merge_requests/7"}, page.Items[0])
	var states []string
	for _, pr := range page.Items {
		states = append(states, pr.State)
	}
	require.Equal(t, []string{"open", "draft", "merged", "closed", "closed"}, states)

	for states, param := range map[string]string{"open": "state=opened", "merged": "state=merged", "closed": "state=closed"} {
		*hits = nil
		_, err := c.ListPRs(context.Background(), scm.Credential{Token: "x"}, "grp/sub/web-app",
			scm.ListQuery{States: []string{states}, Page: 1, PerPage: 200})
		require.NoError(t, err)
		require.Contains(t, (*hits)[0].query, param)
		require.Contains(t, (*hits)[0].query, "per_page=100", "NFR4")
	}
}

func TestGetPR(t *testing.T) {
	c, _ := fake(t, map[string]string{base + "/projects/grp%2Fsub%2Fweb-app/merge_requests/7": "merge_request.json"}, nil, nil)
	pr, err := c.GetPR(context.Background(), scm.Credential{Token: "x"}, "grp/sub/web-app", 7)
	require.NoError(t, err)
	require.Equal(t, 7, pr.Number)
	require.Equal(t, "https://gitlab.com/grp/sub/web-app/-/merge_requests/7", pr.URL)
}

// NFR3, NFR5.
func TestErrors(t *testing.T) {
	for file, status := range map[string]int{"error_401.json": 401, "error_403.json": 403, "error_404.json": 404, "error_429.json": 429} {
		c, _ := fake(t, map[string]string{base + "/user": file}, map[string]int{base + "/user": status},
			map[string]string{"Retry-After": "11"})
		_, err := c.CurrentUser(context.Background(), scm.Credential{Token: "x"})
		var he *scm.HTTPError
		require.ErrorAs(t, err, &he, file)
		require.Equal(t, status, he.Status)
		require.Equal(t, scm.GitLab, he.Provider)
		require.NotContains(t, err.Error(), "insufficient_scope")
	}
}
