package bitbucket

import (
	"context"
	"encoding/base64"
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

type hit struct{ path, query, auth string }

const base = "/2.0"

func fake(t *testing.T, routes map[string]string, status map[string]int) (*Client, *[]hit) {
	t.Helper()
	var hits []hit
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.EscapedPath()
		hits = append(hits, hit{p, r.URL.RawQuery, r.Header.Get("Authorization")})
		file, ok := routes[p]
		if !ok {
			file = "error_404.json"
			w.WriteHeader(404)
		} else if s := status[p]; s != 0 {
			w.Header().Set("Retry-After", "5")
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
	require.Equal(t, "https://api.bitbucket.org/2.0", New().API.BaseURL) // NFR2
}

// FR2.2: Basic auth with the user name and the API token or app password.
func TestCurrentUser(t *testing.T) {
	tok := testutil.Token(t)
	c, hits := fake(t, map[string]string{base + "/user": "user.json"}, nil)
	u, err := c.CurrentUser(context.Background(), scm.Credential{Token: tok, Username: "lan@example.com"})
	require.NoError(t, err)
	require.Equal(t, scm.User{ID: "557058:0000-fake-account", Name: "Lan Nguyen"}, u)
	want := "Basic " + base64.StdEncoding.EncodeToString([]byte("lan@example.com:"+tok))
	require.Equal(t, want, (*hits)[0].auth)
}

// FR3.2, NFR4: the member repositories, one page of 100, filtered locally.
func TestSearchRepos(t *testing.T) {
	c, hits := fake(t, map[string]string{base + "/repositories": "repositories.json"}, nil)
	repos, err := c.SearchRepos(context.Background(), scm.Credential{Token: "x", Username: "u"}, "WEB")
	require.NoError(t, err)
	require.Equal(t, []scm.Repo{{FullName: "ws/web-app", URL: "https://bitbucket.org/ws/web-app"}}, repos)
	require.Equal(t, "pagelen=100&role=member&sort=-updated_on", (*hits)[0].query)
}

func TestGetRepo(t *testing.T) {
	c, _ := fake(t, map[string]string{base + "/repositories/ws/web-app": "repository.json"}, nil)
	r, err := c.GetRepo(context.Background(), scm.Credential{Token: "x", Username: "u"}, "ws/web-app")
	require.NoError(t, err)
	require.Equal(t, "ws/web-app", r.FullName)
}

// FR4.1, A3: declined is its own label; draft too.
func TestListPRs(t *testing.T) {
	path := base + "/repositories/ws/web-app/pullrequests"
	c, hits := fake(t, map[string]string{path: "pullrequests.json"}, nil)
	page, err := c.ListPRs(context.Background(), scm.Credential{Token: "x", Username: "u"}, "ws/web-app",
		scm.ListQuery{States: []string{"open", "closed", "merged"}, Page: 1, PerPage: 4})
	require.NoError(t, err)
	require.Equal(t, "page=1&pagelen=4&sort=-updated_on&state=OPEN&state=DECLINED&state=SUPERSEDED&state=MERGED", (*hits)[0].query)
	require.True(t, page.HasNext)
	require.Equal(t, scm.PullRequest{PRRef: scm.PRRef{Provider: scm.Bitbucket, Repo: "ws/web-app", Number: 3},
		Title: "PROJ-12 Add login page", State: "open", Author: "Lan Nguyen", AuthorID: "557058:0000-fake-account",
		SourceBranch: "feature/PROJ-12", TargetBranch: "main", UpdatedAt: "2026-10-06T09:00:00.000000+00:00",
		URL: "https://bitbucket.org/ws/web-app/pull-requests/3"}, page.Items[0])
	var states []string
	for _, pr := range page.Items {
		states = append(states, pr.State)
	}
	require.Equal(t, []string{"open", "draft", "declined", "merged"}, states)

	*hits = nil
	_, err = c.ListPRs(context.Background(), scm.Credential{Token: "x", Username: "u"}, "ws/web-app",
		scm.ListQuery{States: []string{"merged"}, Page: 2, PerPage: 1000})
	require.NoError(t, err)
	require.Equal(t, "page=2&pagelen=50&sort=-updated_on&state=MERGED", (*hits)[0].query, "Bitbucket pages at most 50 pull requests")
}

func TestGetPR(t *testing.T) {
	c, _ := fake(t, map[string]string{base + "/repositories/ws/web-app/pullrequests/3": "pullrequest.json"}, nil)
	pr, err := c.GetPR(context.Background(), scm.Credential{Token: "x", Username: "u"}, "ws/web-app", 3)
	require.NoError(t, err)
	require.Equal(t, 3, pr.Number)
	require.Equal(t, "open", pr.State)
}

// NFR3, NFR5.
func TestErrors(t *testing.T) {
	for file, status := range map[string]int{"error_401.json": 401, "error_403.json": 403, "error_404.json": 404, "error_429.json": 429} {
		c, _ := fake(t, map[string]string{base + "/user": file}, map[string]int{base + "/user": status})
		_, err := c.CurrentUser(context.Background(), scm.Credential{Token: "x", Username: "u"})
		var he *scm.HTTPError
		require.ErrorAs(t, err, &he, file)
		require.Equal(t, status, he.Status)
		require.NotContains(t, err.Error(), "privilege")
	}
}
