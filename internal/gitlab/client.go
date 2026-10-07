package gitlab

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/scm"
)

// Client calls the GitLab REST API v4. It holds no credentials.
type Client struct {
	API *scm.API // BaseURL is overridden only by tests
}

// New returns the production client for https://gitlab.com/api/v4 (NFR2).
func New() *Client {
	return &Client{API: scm.NewAPI(scm.GitLab, "https://gitlab.com/api/v4", func(r *http.Request, c scm.Credential) {
		r.Header.Set("PRIVATE-TOKEN", c.Token) // a personal access token with read_api (FR2.2, FR2.3)
	})}
}

type user struct {
	Username string `json:"username"`
	Name     string `json:"name"`
}

type project struct {
	PathWithNamespace string `json:"path_with_namespace"`
	WebURL            string `json:"web_url"`
}

type mergeRequest struct {
	IID          int    `json:"iid"`
	Title        string `json:"title"`
	State        string `json:"state"` // opened, closed, locked, merged
	Draft        bool   `json:"draft"`
	Author       user   `json:"author"`
	SourceBranch string `json:"source_branch"`
	TargetBranch string `json:"target_branch"`
	UpdatedAt    string `json:"updated_at"`
}

// CurrentUser reads GET /user (FR2.4).
func (c *Client) CurrentUser(ctx context.Context, cred scm.Credential) (scm.User, error) {
	var u user
	if _, err := c.API.Get(ctx, cred, "/user", nil, &u); err != nil {
		return scm.User{}, err
	}
	name := u.Name
	if name == "" {
		name = u.Username
	}
	return scm.User{ID: u.Username, Name: name}, nil
}

// SearchRepos reads one page of the projects the token is a member of,
// searched by GitLab.
func (c *Client) SearchRepos(ctx context.Context, cred scm.Credential, query string) ([]scm.Repo, error) {
	q := url.Values{"membership": {"true"}, "simple": {"true"}, "per_page": {"100"}, "order_by": {"last_activity_at"}}
	if query != "" {
		q.Set("search", query)
	}
	var list []project
	if _, err := c.API.Get(ctx, cred, "/projects", q, &list); err != nil {
		return nil, err
	}
	out := make([]scm.Repo, 0, len(list))
	for _, p := range list {
		out = append(out, scm.Repo{FullName: p.PathWithNamespace, URL: p.WebURL})
	}
	return out, nil
}

// GetRepo reads GET /projects/{url-encoded path}.
func (c *Client) GetRepo(ctx context.Context, cred scm.Credential, name string) (scm.Repo, error) {
	var p project
	if _, err := c.API.Get(ctx, cred, projectPath(name), nil, &p); err != nil {
		return scm.Repo{}, err
	}
	return scm.Repo{FullName: p.PathWithNamespace, URL: p.WebURL}, nil
}

// ListPRs reads one page of the project's merge requests, newest update first.
func (c *Client) ListPRs(ctx context.Context, cred scm.Credential, name string, q scm.ListQuery) (scm.PRPage, error) {
	size := q.PageSize(100)
	params := url.Values{"state": {apiState(q.States)}, "per_page": {strconv.Itoa(size)},
		"page": {strconv.Itoa(q.PageNumber())}, "order_by": {"updated_at"}, "sort": {"desc"}}
	var list []mergeRequest
	if _, err := c.API.Get(ctx, cred, projectPath(name)+"/merge_requests", params, &list); err != nil {
		return scm.PRPage{}, err
	}
	page := scm.PRPage{Items: []scm.PullRequest{}, HasNext: len(list) == size}
	for _, m := range list {
		if pr := toPR(name, m); q.Wants(pr.State) {
			page.Items = append(page.Items, pr)
		}
	}
	return page, nil
}

// GetPR reads one merge request by its project-level iid.
func (c *Client) GetPR(ctx context.Context, cred scm.Credential, name string, number int) (scm.PullRequest, error) {
	var m mergeRequest
	if _, err := c.API.Get(ctx, cred, projectPath(name)+"/merge_requests/"+strconv.Itoa(number), nil, &m); err != nil {
		return scm.PullRequest{}, err
	}
	return toPR(name, m), nil
}

// projectPath is /projects/<group%2Fsub%2Fproject>: GitLab takes the full
// path as one encoded id.
func projectPath(name string) string { return "/projects/" + url.PathEscape(name) }

func apiState(states []string) string {
	if len(states) != 1 {
		return "all"
	}
	if states[0] == scm.StateOpen {
		return "opened"
	}
	return states[0]
}

func toPR(name string, m mergeRequest) scm.PullRequest {
	state := scm.StateOpen
	switch {
	case m.State == "merged":
		state = scm.StateMerged
	case m.State == "closed" || m.State == "locked":
		state = scm.StateClosed
	case m.Draft:
		state = scm.StateDraft
	}
	ref := scm.PRRef{Provider: scm.GitLab, Repo: name, Number: m.IID}
	author := m.Author.Name
	if author == "" {
		author = m.Author.Username
	}
	return scm.PullRequest{PRRef: ref, Title: m.Title, State: state, Author: author, AuthorID: m.Author.Username,
		SourceBranch: m.SourceBranch, TargetBranch: m.TargetBranch, UpdatedAt: m.UpdatedAt, URL: scm.PRURL(ref)}
}
