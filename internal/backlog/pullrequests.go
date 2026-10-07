package backlog

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

// PullRequest is a Backlog pull request (C7).
type PullRequest struct {
	ID           int64
	RepositoryID int64
	Number       int
	Summary      string
	Description  string // never logged or formatted
	Base         string
	Branch       string
	StatusID     int
	AssigneeName string
	IssueID      int64
	Created      string
	AuthorName   string // the creator's display name (PR list, FR4.2)
	Updated      string
}

// String hides the description, which is user content.
func (p PullRequest) String() string {
	return fmt.Sprintf("backlog.PullRequest{Number: %d, RepositoryID: %d, StatusID: %d}", p.Number, p.RepositoryID, p.StatusID)
}

// GoString hides the description for %#v as well.
func (p PullRequest) GoString() string { return p.String() }

// Format hides the description for every verb, including %+v.
func (p PullRequest) Format(f fmt.State, _ rune) { _, _ = fmt.Fprint(f, p.String()) }

// PullRequestState maps a Backlog status id to open, closed, merged or unknown.
// Only 1 = Open is confirmed by the Backlog docs (plan assumption).
func PullRequestState(statusID int) string {
	switch statusID {
	case 1:
		return "open"
	case 2:
		return "closed"
	case 3:
		return "merged"
	default:
		return "unknown"
	}
}

// PullRequestQuery filters GET .../pullRequests.
type PullRequestQuery struct {
	StatusIDs      []int64
	AssigneeIDs    []int64
	IssueIDs       []int64
	CreatedUserIDs []int64
	Offset         int
	Count          int // clamped to 1-100
}

// Values encodes the query string.
func (q PullRequestQuery) Values() url.Values {
	v := url.Values{}
	add := func(key string, ids []int64) {
		for _, id := range ids {
			v.Add(key, strconv.FormatInt(id, 10))
		}
	}
	add("statusId[]", q.StatusIDs)
	add("assigneeId[]", q.AssigneeIDs)
	add("issueId[]", q.IssueIDs)
	add("createdUserId[]", q.CreatedUserIDs)
	if q.Offset > 0 {
		v.Set("offset", strconv.Itoa(q.Offset))
	}
	v.Set("count", strconv.Itoa(min(max(q.Count, 1), 100)))
	return v
}

// NewPullRequest is the Add Pull Request form.
type NewPullRequest struct {
	Summary     string
	Description string
	Base        string
	Branch      string
	IssueID     int64 // 0 = none
}

// Form encodes the form body; issueId only when set.
func (n NewPullRequest) Form() url.Values {
	v := url.Values{"summary": {n.Summary}, "description": {n.Description}, "base": {n.Base}, "branch": {n.Branch}}
	if n.IssueID != 0 {
		v.Set("issueId", strconv.FormatInt(n.IssueID, 10))
	}
	return v
}

type rawPullRequest struct {
	ID           int64  `json:"id"`
	RepositoryID *int64 `json:"repositoryId"`
	Number       *int   `json:"number"`
	Summary      string `json:"summary"`
	Description  string `json:"description"`
	Base         string `json:"base"`
	Branch       string `json:"branch"`
	Status       struct {
		ID int `json:"id"`
	} `json:"status"`
	Assignee *struct {
		Name string `json:"name"`
	} `json:"assignee"`
	Issue *struct {
		ID int64 `json:"id"`
	} `json:"issue"`
	Created     string `json:"created"`
	Updated     string `json:"updated"`
	CreatedUser *struct {
		Name string `json:"name"`
	} `json:"createdUser"`
}

func (r rawPullRequest) pullRequest() (PullRequest, bool) {
	if r.Number == nil || *r.Number <= 0 || r.RepositoryID == nil {
		return PullRequest{}, false
	}
	pr := PullRequest{ID: r.ID, RepositoryID: *r.RepositoryID, Number: *r.Number, Summary: r.Summary,
		Description: r.Description, Base: r.Base, Branch: r.Branch, StatusID: r.Status.ID, Created: r.Created, Updated: r.Updated}
	if r.CreatedUser != nil {
		pr.AuthorName = r.CreatedUser.Name
	}
	if r.Assignee != nil {
		pr.AssigneeName = r.Assignee.Name
	}
	if r.Issue != nil {
		pr.IssueID = r.Issue.ID
	}
	return pr, true
}

// errBody is a fresh Unreachable body error.
func errBody() error { return &Error{Kind: KindUnreachable, Status: 200, Class: "body"} }

func parsePullRequests(body []byte) ([]PullRequest, error) {
	var raw []rawPullRequest
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, errBody()
	}
	out := make([]PullRequest, 0, len(raw))
	for _, r := range raw {
		pr, ok := r.pullRequest()
		if !ok {
			return nil, errBody()
		}
		out = append(out, pr)
	}
	return out, nil
}

func parsePullRequest(body []byte) (PullRequest, error) {
	var raw rawPullRequest
	if err := json.Unmarshal(body, &raw); err != nil {
		return PullRequest{}, errBody()
	}
	pr, ok := raw.pullRequest()
	if !ok {
		return PullRequest{}, errBody()
	}
	return pr, nil
}
