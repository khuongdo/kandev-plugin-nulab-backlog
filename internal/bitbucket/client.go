package bitbucket

import (
	"context"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/scm"
)

// Client calls the Bitbucket Cloud REST API 2.0. It holds no credentials.
type Client struct {
	API *scm.API // BaseURL is overridden only by tests
}

// New returns the production client for https://api.bitbucket.org/2.0 (NFR2).
func New() *Client {
	return &Client{API: scm.NewAPI(scm.Bitbucket, "https://api.bitbucket.org/2.0", func(r *http.Request, c scm.Credential) {
		r.SetBasicAuth(c.Username, c.Token) // an API token or app password with its user name (FR2.2)
	})}
}

// maxPRPage is Bitbucket's largest pull request page.
const maxPRPage = 50

type account struct {
	DisplayName string `json:"display_name"`
	AccountID   string `json:"account_id"`
}

type repository struct {
	FullName string `json:"full_name"`
	Links    struct {
		HTML struct {
			Href string `json:"href"`
		} `json:"html"`
	} `json:"links"`
}

type branch struct {
	Branch struct {
		Name string `json:"name"`
	} `json:"branch"`
}

type pullRequest struct {
	ID          int     `json:"id"`
	Title       string  `json:"title"`
	State       string  `json:"state"` // OPEN, MERGED, DECLINED, SUPERSEDED
	Draft       bool    `json:"draft"`
	Author      account `json:"author"`
	Source      branch  `json:"source"`
	Destination branch  `json:"destination"`
	UpdatedOn   string  `json:"updated_on"`
}

type paged[T any] struct {
	Values []T    `json:"values"`
	Next   string `json:"next"`
}

// CurrentUser reads GET /user (FR2.4).
func (c *Client) CurrentUser(ctx context.Context, cred scm.Credential) (scm.User, error) {
	var a account
	if _, err := c.API.Get(ctx, cred, "/user", nil, &a); err != nil {
		return scm.User{}, err
	}
	return scm.User{ID: a.AccountID, Name: a.DisplayName}, nil
}

// SearchRepos reads one page of the repositories the user is a member of
// and keeps the names that contain query.
// ponytail: the 100 most recently updated repositories; type workspace/repo for others.
func (c *Client) SearchRepos(ctx context.Context, cred scm.Credential, query string) ([]scm.Repo, error) {
	var p paged[repository]
	q := url.Values{"role": {"member"}, "pagelen": {"100"}, "sort": {"-updated_on"}}
	if _, err := c.API.Get(ctx, cred, "/repositories", q, &p); err != nil {
		return nil, err
	}
	out := []scm.Repo{}
	for _, r := range p.Values {
		if strings.Contains(strings.ToLower(r.FullName), strings.ToLower(query)) {
			out = append(out, scm.Repo{FullName: r.FullName, URL: r.Links.HTML.Href})
		}
	}
	return out, nil
}

// GetRepo reads GET /repositories/{workspace}/{repo}.
func (c *Client) GetRepo(ctx context.Context, cred scm.Credential, name string) (scm.Repo, error) {
	var r repository
	if _, err := c.API.Get(ctx, cred, "/repositories/"+name, nil, &r); err != nil {
		return scm.Repo{}, err
	}
	return scm.Repo{FullName: r.FullName, URL: r.Links.HTML.Href}, nil
}

// ListPRs reads one page of the repository's pull requests, newest update first.
func (c *Client) ListPRs(ctx context.Context, cred scm.Credential, name string, q scm.ListQuery) (scm.PRPage, error) {
	params := url.Values{"pagelen": {strconv.Itoa(q.PageSize(maxPRPage))}, "page": {strconv.Itoa(q.PageNumber())},
		"sort": {"-updated_on"}}
	for _, s := range []struct{ filter, api string }{
		{scm.StateOpen, "OPEN"}, {scm.StateClosed, "DECLINED"}, {scm.StateClosed, "SUPERSEDED"}, {scm.StateMerged, "MERGED"},
	} {
		if slices.Contains(q.States, s.filter) {
			params.Add("state", s.api)
		}
	}
	var p paged[pullRequest]
	if _, err := c.API.Get(ctx, cred, "/repositories/"+name+"/pullrequests", params, &p); err != nil {
		return scm.PRPage{}, err
	}
	page := scm.PRPage{Items: []scm.PullRequest{}, HasNext: p.Next != ""}
	for _, pr := range p.Values {
		page.Items = append(page.Items, toPR(name, pr))
	}
	return page, nil
}

// GetPR reads GET /repositories/{workspace}/{repo}/pullrequests/{id}.
func (c *Client) GetPR(ctx context.Context, cred scm.Credential, name string, number int) (scm.PullRequest, error) {
	var pr pullRequest
	if _, err := c.API.Get(ctx, cred, "/repositories/"+name+"/pullrequests/"+strconv.Itoa(number), nil, &pr); err != nil {
		return scm.PullRequest{}, err
	}
	return toPR(name, pr), nil
}

func toPR(name string, pr pullRequest) scm.PullRequest {
	state := scm.StateOpen
	switch {
	case pr.State == "MERGED":
		state = scm.StateMerged
	case pr.State == "DECLINED" || pr.State == "SUPERSEDED":
		state = scm.StateDeclined
	case pr.Draft:
		state = scm.StateDraft
	}
	ref := scm.PRRef{Provider: scm.Bitbucket, Repo: name, Number: pr.ID}
	return scm.PullRequest{PRRef: ref, Title: pr.Title, State: state, Author: pr.Author.DisplayName,
		AuthorID: pr.Author.AccountID, SourceBranch: pr.Source.Branch.Name, TargetBranch: pr.Destination.Branch.Name,
		UpdatedAt: pr.UpdatedOn, URL: scm.PRURL(ref)}
}
