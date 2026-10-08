package scm

import (
	"context"
	"slices"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/connection"
)

// PRListInput is the scm.prs.list body.
type PRListInput struct {
	Provider   Provider `json:"provider"`
	ProjectKey string   `json:"projectKey"`
	Repo       string   `json:"repo"`
	Statuses   []string `json:"statuses"`
	Author     string   `json:"author"`
	Page       int      `json:"page"`
}

// PRRow is one row of the PR list.
type PRRow struct {
	PullRequest
	LinkedTaskIDs []string `json:"linkedTaskIds"`
}

// PRListPage is the scm.prs.list reply. The providers give no total.
type PRListPage struct {
	Items    []PRRow `json:"items"`
	Page     int     `json:"page"`
	PageSize int     `json:"pageSize"`
	HasNext  bool    `json:"hasNext"`
}

// ListPRs returns one page of a mapped repository's pull requests, and
// auto-links the issues it sees (FR3.3, FR4.1, FR5.2).
func (s *Service) ListPRs(ctx context.Context, ws string, in PRListInput) (PRListPage, error) {
	q := QueryInput{Name: "list", Provider: in.Provider, ProjectKey: in.ProjectKey, Repo: in.Repo,
		Statuses: in.Statuses, Author: in.Author}
	if err := q.Validate(); err != nil {
		return PRListPage{}, err
	}
	st, err := s.settings(ctx, ws, in.Provider)
	if err != nil {
		return PRListPage{}, err
	}
	if !isMapped(st, q.ProjectKey, q.Repo) {
		return PRListPage{}, &connection.FieldError{Field: FieldRepository, Err: errUnmapped}
	}
	return s.listPRs(ctx, ws, st, q, max(in.Page, 1))
}

func (s *Service) listPRs(ctx context.Context, ws string, st Settings, q QueryInput, page int) (PRListPage, error) {
	ctx, cred, err := s.credential(ctx, ws, q.Provider)
	if err != nil {
		return PRListPage{}, err
	}
	res, err := s.clients[q.Provider].ListPRs(ctx, cred, q.Repo, ListQuery{States: q.Statuses, Page: page, PerPage: prPageSize})
	if err != nil {
		return PRListPage{}, err
	}
	prs := byAuthor(res.Items, q.Author, st.AccountID)
	s.autoLink(ctx, ws, st, q.Repo, prs)
	links, err := s.store.Links(ctx, ws)
	if err != nil {
		return PRListPage{}, err
	}
	out := PRListPage{Items: []PRRow{}, Page: page, PageSize: prPageSize, HasNext: res.HasNext}
	for _, pr := range prs {
		row := PRRow{PullRequest: pr, LinkedTaskIDs: []string{}}
		for _, l := range links {
			if l.TaskID != "" && l.Key() == pr.Key() {
				row.LinkedTaskIDs = append(row.LinkedTaskIDs, l.TaskID)
			}
		}
		out.Items = append(out.Items, row)
	}
	return out, nil
}

// byAuthor keeps the token account's pull requests for "me".
func byAuthor(prs []PullRequest, author, accountID string) []PullRequest {
	if author != WhoMe {
		return prs
	}
	return slices.DeleteFunc(slices.Clone(prs), func(pr PullRequest) bool { return pr.AuthorID != accountID })
}
