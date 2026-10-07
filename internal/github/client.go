package github

import (
	"context"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/scm"
)

// Client calls the GitHub REST API. It holds no credentials.
type Client struct {
	API *scm.API // BaseURL is overridden only by tests
}

// New returns the production client for https://api.github.com (NFR2).
func New() *Client {
	return &Client{API: scm.NewAPI(scm.GitHub, "https://api.github.com", func(r *http.Request, c scm.Credential) {
		r.Header.Set("Authorization", "Bearer "+c.Token) // fine-grained or classic token (FR2.2)
		r.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	})}
}

type user struct {
	Login string `json:"login"`
	Name  string `json:"name"`
}

type repo struct {
	FullName string `json:"full_name"`
	HTMLURL  string `json:"html_url"`
}

type pull struct {
	Number   int     `json:"number"`
	Title    string  `json:"title"`
	State    string  `json:"state"`
	Draft    bool    `json:"draft"`
	MergedAt *string `json:"merged_at"`
	User     user    `json:"user"`
	Head     struct {
		Ref string `json:"ref"`
	} `json:"head"`
	Base struct {
		Ref string `json:"ref"`
	} `json:"base"`
	UpdatedAt string `json:"updated_at"`
}

// CurrentUser reads GET /user (FR2.4).
func (c *Client) CurrentUser(ctx context.Context, cred scm.Credential) (scm.User, error) {
	var u user
	if _, err := c.API.Get(ctx, cred, "/user", nil, &u); err != nil {
		return scm.User{}, err
	}
	name := u.Name
	if name == "" {
		name = u.Login
	}
	return scm.User{ID: u.Login, Name: name}, nil
}

// SearchRepos reads one page of GET /user/repos and keeps the names that
// contain query. GitHub has no member-repository search.
// ponytail: the 100 most recently updated repositories; type owner/name for others.
func (c *Client) SearchRepos(ctx context.Context, cred scm.Credential, query string) ([]scm.Repo, error) {
	var list []repo
	q := url.Values{"per_page": {"100"}, "sort": {"updated"}}
	if _, err := c.API.Get(ctx, cred, "/user/repos", q, &list); err != nil {
		return nil, err
	}
	out := []scm.Repo{}
	for _, r := range list {
		if strings.Contains(strings.ToLower(r.FullName), strings.ToLower(query)) {
			out = append(out, scm.Repo{FullName: r.FullName, URL: r.HTMLURL})
		}
	}
	return out, nil
}

// GetRepo reads GET /repos/{owner}/{repo}.
func (c *Client) GetRepo(ctx context.Context, cred scm.Credential, name string) (scm.Repo, error) {
	var r repo
	if _, err := c.API.Get(ctx, cred, "/repos/"+name, nil, &r); err != nil {
		return scm.Repo{}, err
	}
	return scm.Repo{FullName: r.FullName, URL: r.HTMLURL}, nil
}

// ListPRs reads one page of GET /repos/{owner}/{repo}/pulls, newest update
// first. GitHub filters only open or closed, so merged and closed are told
// apart here; a page may hold fewer rows than PerPage.
func (c *Client) ListPRs(ctx context.Context, cred scm.Credential, name string, q scm.ListQuery) (scm.PRPage, error) {
	size := q.PageSize(100)
	params := url.Values{"state": {apiState(q.States)}, "per_page": {strconv.Itoa(size)},
		"page": {strconv.Itoa(q.PageNumber())}, "sort": {"updated"}, "direction": {"desc"}}
	var list []pull
	if _, err := c.API.Get(ctx, cred, "/repos/"+name+"/pulls", params, &list); err != nil {
		return scm.PRPage{}, err
	}
	page := scm.PRPage{Items: []scm.PullRequest{}, HasNext: len(list) == size}
	for _, p := range list {
		if pr := toPR(name, p); q.Wants(pr.State) {
			page.Items = append(page.Items, pr)
		}
	}
	return page, nil
}

// GetPR reads GET /repos/{owner}/{repo}/pulls/{number}.
func (c *Client) GetPR(ctx context.Context, cred scm.Credential, name string, number int) (scm.PullRequest, error) {
	var p pull
	if _, err := c.API.Get(ctx, cred, "/repos/"+name+"/pulls/"+strconv.Itoa(number), nil, &p); err != nil {
		return scm.PullRequest{}, err
	}
	return toPR(name, p), nil
}

func apiState(states []string) string {
	switch {
	case !slices.Contains(states, scm.StateOpen):
		return "closed"
	case len(states) == 1:
		return "open"
	default:
		return "all"
	}
}

func toPR(name string, p pull) scm.PullRequest {
	state := scm.StateOpen
	switch {
	case p.MergedAt != nil:
		state = scm.StateMerged
	case p.State == "closed":
		state = scm.StateClosed
	case p.Draft:
		state = scm.StateDraft
	}
	ref := scm.PRRef{Provider: scm.GitHub, Repo: name, Number: p.Number}
	return scm.PullRequest{PRRef: ref, Title: p.Title, State: state, Author: p.User.Login, AuthorID: p.User.Login,
		SourceBranch: p.Head.Ref, TargetBranch: p.Base.Ref, UpdatedAt: p.UpdatedAt, URL: scm.PRURL(ref)}
}
