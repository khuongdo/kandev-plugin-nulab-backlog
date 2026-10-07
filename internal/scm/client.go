package scm

import (
	"context"
	"fmt"
)

// Credential is a provider token, plus the user name Bitbucket's Basic auth
// needs. It is never stored outside Kandev's secret store (NFR1).
type Credential struct {
	Token    string `json:"token"` //nolint:gosec // G117: goes only to Kandev's encrypted secret store
	Username string `json:"username,omitempty"`
}

// String hides the token.
func (c Credential) String() string { return fmt.Sprintf("scm.Credential{Username: %s}", c.Username) }

// GoString hides the token for %#v.
func (c Credential) GoString() string { return c.String() }

// Format hides the token for every verb, including %+v.
func (c Credential) Format(f fmt.State, _ rune) { _, _ = fmt.Fprint(f, c.String()) }

// User is the account a token belongs to (FR2.4). ID matches
// PullRequest.AuthorID for the "me" filter.
type User struct {
	ID   string
	Name string
}

// Repo is one repository a token can read.
type Repo struct {
	FullName string `json:"fullName"`
	URL      string `json:"url"`
}

// PullRequest is one pull or merge request (FR4.1).
type PullRequest struct {
	PRRef
	Title        string `json:"title"`
	State        string `json:"state"` // open, draft, closed, declined or merged
	Author       string `json:"author"`
	AuthorID     string `json:"-"`
	SourceBranch string `json:"sourceBranch"`
	TargetBranch string `json:"targetBranch"`
	UpdatedAt    string `json:"updatedAt"`
	URL          string `json:"url"`
}

// PRPage is one page of pull requests, newest update first.
type PRPage struct {
	Items   []PullRequest
	HasNext bool
}

// ListQuery filters a pull request list. States are filter states (open,
// closed, merged); Page starts at 1; PerPage is capped at 100 (NFR4).
type ListQuery struct {
	States  []string
	Page    int
	PerPage int
}

// PageSize returns PerPage within 1..limit.
func (q ListQuery) PageSize(limit int) int { return min(max(q.PerPage, 1), limit) }

// PageNumber returns Page, at least 1.
func (q ListQuery) PageNumber() int { return max(q.Page, 1) }

// Wants reports whether the filter states include the shown state.
func (q ListQuery) Wants(state string) bool {
	for _, s := range q.States {
		if s == FilterState(state) {
			return true
		}
	}
	return false
}

// Client is one provider's read-only API (FR4.5). internal/github,
// internal/gitlab and internal/bitbucket implement it; tests swap in a fake.
type Client interface {
	CurrentUser(ctx context.Context, cred Credential) (User, error)
	// SearchRepos lists up to 100 readable repositories whose name contains query.
	SearchRepos(ctx context.Context, cred Credential, query string) ([]Repo, error)
	GetRepo(ctx context.Context, cred Credential, repo string) (Repo, error)
	ListPRs(ctx context.Context, cred Credential, repo string, q ListQuery) (PRPage, error)
	GetPR(ctx context.Context, cred Credential, repo string, number int) (PullRequest, error)
}
