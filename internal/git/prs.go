package git

import (
	"context"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/backlog"
)

// FieldPage is reported for a PR list page below 1.
const FieldPage = "page"

const prPageSize = 20

// PRListInput is the git.prs.list body (FR4.1).
type PRListInput struct {
	ProjectKey string   `json:"projectKey"`
	RepoName   string   `json:"repoName"`
	Statuses   []string `json:"statuses"` // empty = every status
	Assignee   string   `json:"assignee"` // anyone (default) or me
	Creator    string   `json:"creator"`  // anyone (default) or me
	Page       int      `json:"page"`
}

// PullRequestRow is one pull request of the PR list (FR4.2).
type PullRequestRow struct {
	Number        int      `json:"number"`
	Title         string   `json:"title"`
	Status        string   `json:"status"`
	Author        string   `json:"author"`
	Assignee      string   `json:"assignee,omitempty"`
	Updated       string   `json:"updated"`
	URL           string   `json:"url"`
	LinkedTaskIDs []string `json:"linkedTaskIds"`
}

// PullRequestPage is the git.prs.list reply.
type PullRequestPage struct {
	Items    []PullRequestRow `json:"items"`
	Page     int              `json:"page"`
	PageSize int              `json:"pageSize"`
	Total    int              `json:"total"`
	HasNext  bool             `json:"hasNext"`
}

func (in *PRListInput) validate(selected []string) error {
	if err := validateRepo(in.ProjectKey, in.RepoName, selected); err != nil {
		return err
	}
	for _, s := range in.Statuses {
		if _, ok := stateIDs[s]; !ok {
			return invalid(FieldStatuses)
		}
	}
	for _, w := range []struct {
		field string
		who   *string
	}{{FieldAssignee, &in.Assignee}, {FieldCreator, &in.Creator}} {
		if *w.who == "" {
			*w.who = WhoAnyone
		}
		if !validWho(*w.who) {
			return invalid(w.field)
		}
	}
	if in.Page < 1 {
		return invalid(FieldPage)
	}
	return nil
}

// ListPullRequests returns one page of 20 pull requests of a selected
// project's repository, newest first, with the total from the count call
// and the tasks linked to each (BR4.1, BR4.2).
func (s *Service) ListPullRequests(ctx context.Context, ws string, in PRListInput) (PullRequestPage, error) {
	snap, err := s.conn.Current(ctx, ws)
	if err != nil {
		return PullRequestPage{}, err
	}
	if err := in.validate(snap.SelectedProjects); err != nil {
		return PullRequestPage{}, err
	}
	ctx, creds, err := s.credentials(ctx, ws)
	if err != nil {
		return PullRequestPage{}, err
	}
	q := backlog.PullRequestQuery{Offset: (in.Page - 1) * prPageSize, Count: prPageSize}
	if len(in.Statuses) > 0 {
		q.StatusIDs = statusIDs(in.Statuses)
	}
	if in.Assignee == WhoMe || in.Creator == WhoMe {
		me, err := s.gateway.Myself(ctx, creds)
		if err != nil {
			return PullRequestPage{}, err
		}
		if in.Assignee == WhoMe {
			q.AssigneeIDs = []int64{me.ID}
		}
		if in.Creator == WhoMe {
			q.CreatedUserIDs = []int64{me.ID}
		}
	}
	total, err := s.gateway.PullRequestCount(ctx, creds, backlog.Interactive, in.ProjectKey, in.RepoName, q)
	if err != nil {
		return PullRequestPage{}, err
	}
	prs, err := s.gateway.PullRequests(ctx, creds, backlog.Interactive, in.ProjectKey, in.RepoName, q)
	if err != nil {
		return PullRequestPage{}, err
	}
	links, err := s.store.Links(ctx, ws)
	if err != nil {
		return PullRequestPage{}, err
	}
	page := PullRequestPage{Items: make([]PullRequestRow, 0, len(prs)), Page: in.Page, PageSize: prPageSize, Total: total,
		HasNext: in.Page*prPageSize < total}
	for _, pr := range prs[:min(len(prs), prPageSize)] {
		key := LinkKey(snap.SpaceHost, pr.RepositoryID, pr.Number)
		linked := []string{}
		for _, l := range links {
			if l.Status == StatusActive && l.Key() == key {
				linked = append(linked, l.TaskID)
			}
		}
		page.Items = append(page.Items, PullRequestRow{Number: pr.Number, Title: pr.Summary,
			Status: backlog.PullRequestState(pr.StatusID), Author: pr.AuthorName, Assignee: pr.AssigneeName,
			Updated: pr.Updated, URL: prURL(snap.SpaceHost, in.ProjectKey, in.RepoName, pr.Number), LinkedTaskIDs: linked})
	}
	return page, nil
}
