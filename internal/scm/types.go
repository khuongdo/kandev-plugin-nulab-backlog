package scm

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/connection"
)

// Provider is a source control service this package supports (FR1.1). It is
// the plugin's own name, never a Kandev repository provider id (C1).
type Provider string

// The three cloud providers (FR1.2). Backlog Git is internal/git's.
const (
	GitHub    Provider = "github"
	GitLab    Provider = "gitlab"
	Bitbucket Provider = "bitbucket"
)

// Providers lists every provider in display order.
var Providers = []Provider{GitHub, GitLab, Bitbucket}

// Input fields reported with connection.FieldError (validation).
const (
	FieldProvider   = "provider"
	FieldToken      = "token"
	FieldUsername   = "username"
	FieldRepository = "repository"
	FieldRepos      = "repos"
	FieldProjectKey = "projectKey"
	FieldURL        = "url"
	FieldName       = "name"
	FieldStatuses   = "statuses"
	FieldAuthor     = "author"
	FieldWorkflow   = "workflowId"
	FieldInterval   = "intervalMinutes"
	FieldLimit      = "limit"
	FieldKey        = "key"
)

// Pull request states. Draft and declined are shown as their own labels and
// filtered as open and closed (A3).
const (
	StateOpen     = "open"
	StateClosed   = "closed"
	StateMerged   = "merged"
	StateDraft    = "draft"
	StateDeclined = "declined"
)

// Who values of the author filter.
const (
	WhoAnyone = "anyone"
	WhoMe     = "me"
)

var (
	errInvalid    = errors.New("invalid value")
	projectKeyRe  = regexp.MustCompile(`^[A-Z][A-Z0-9_]{0,24}$`)
	segmentRe     = regexp.MustCompile(`^[A-Za-z0-9_][A-Za-z0-9._-]{0,99}$`)
	issueKeyText  = regexp.MustCompile(`\b([A-Z][A-Z0-9_]*)-[1-9][0-9]*\b`)
	prNumberRe    = regexp.MustCompile(`^[1-9][0-9]{0,8}$`)
	filterStates  = []string{StateOpen, StateClosed, StateMerged}
	webHosts      = map[Provider]string{GitHub: "github.com", GitLab: "gitlab.com", Bitbucket: "bitbucket.org"}
	maxRepoDepth  = map[Provider]int{GitHub: 2, GitLab: 20, Bitbucket: 2}
	errPRURL      = errors.New("must be a pull request page on github.com, gitlab.com or bitbucket.org")
	errUnmapped   = errors.New("the repository is not mapped to a selected Backlog project")
	errNoProvider = errors.New("must be github, gitlab or bitbucket")
)

func invalid(field string) error { return &connection.FieldError{Field: field, Err: errInvalid} }

// ParseProvider accepts exactly github, gitlab or bitbucket.
func ParseProvider(s string) (Provider, error) {
	p := Provider(s)
	if !slices.Contains(Providers, p) {
		return "", &connection.FieldError{Field: FieldProvider, Err: errNoProvider}
	}
	return p, nil
}

// ValidRepo reports whether name is a repository full name of p: owner/name
// (GitHub), group[/subgroup...]/project (GitLab) or workspace/repo (Bitbucket).
func ValidRepo(p Provider, name string) bool {
	parts := strings.Split(name, "/")
	depth, ok := maxRepoDepth[p]
	if !ok || len(parts) < 2 || len(parts) > depth {
		return false
	}
	for _, s := range parts {
		if !segmentRe.MatchString(s) || s == "." || s == ".." || strings.HasSuffix(s, ".git") {
			return false
		}
	}
	return true
}

// PRRef identifies one pull request of one provider.
type PRRef struct {
	Provider Provider `json:"provider"`
	Repo     string   `json:"repo"`
	Number   int      `json:"number"`
}

// Key is provider|repo|number: the link, dismissal and ledger key.
func (r PRRef) Key() string { return fmt.Sprintf("%s|%s|%d", r.Provider, r.Repo, r.Number) }

// PRURL is the pull request's page on the provider's website.
func PRURL(r PRRef) string {
	switch r.Provider {
	case GitLab:
		return fmt.Sprintf("https://gitlab.com/%s/-/merge_requests/%d", r.Repo, r.Number)
	case Bitbucket:
		return fmt.Sprintf("https://bitbucket.org/%s/pull-requests/%d", r.Repo, r.Number)
	default:
		return fmt.Sprintf("https://github.com/%s/pull/%d", r.Repo, r.Number)
	}
}

// ParsePRURL reads the web URL of a pull request on github.com, gitlab.com
// or bitbucket.org (FR5.1). Any other host, a port, userinfo, a query or a
// fragment is a validation error on field url (NFR2).
func ParsePRURL(raw string) (PRRef, error) {
	bad := &connection.FieldError{Field: FieldURL, Err: errPRURL}
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u.Scheme != "https" || u.User != nil || u.Port() != "" || u.RawQuery != "" ||
		u.ForceQuery || u.Fragment != "" {
		return PRRef{}, bad
	}
	p := providerOfHost(u.Host)
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	var repo []string
	var num string
	switch p {
	case GitHub, Bitbucket: // owner/repo/pull/N[/...], workspace/repo/pull-requests/N[/...]
		marker := map[Provider]string{GitHub: "pull", Bitbucket: "pull-requests"}[p]
		if len(parts) < 4 || parts[2] != marker {
			return PRRef{}, bad
		}
		repo, num = parts[:2], parts[3]
	case GitLab: // group/.../project/-/merge_requests/N[/...]
		i := slices.Index(parts, "-")
		if i < 2 || len(parts) < i+3 || parts[i+1] != "merge_requests" {
			return PRRef{}, bad
		}
		repo, num = parts[:i], parts[i+2]
	default:
		return PRRef{}, bad
	}
	name := strings.Join(repo, "/")
	if !ValidRepo(p, name) || !prNumberRe.MatchString(num) {
		return PRRef{}, bad
	}
	n, _ := strconv.Atoi(num)
	return PRRef{Provider: p, Repo: name, Number: n}, nil
}

func providerOfHost(host string) Provider {
	for p, h := range webHosts {
		if strings.EqualFold(host, h) {
			return p
		}
	}
	return ""
}

// FilterState is the saved-query state of a shown state (A3).
func FilterState(state string) string {
	switch state {
	case StateClosed, StateDeclined:
		return StateClosed
	case StateMerged:
		return StateMerged
	default:
		return StateOpen
	}
}

// IssueKeys returns the distinct issue keys of the selected projects in
// texts, in order of appearance (FR5.2, A4).
func IssueKeys(texts []string, selected []string) []string {
	var out []string
	for _, t := range texts {
		for _, m := range issueKeyText.FindAllStringSubmatch(t, -1) {
			if slices.Contains(selected, m[1]) && !slices.Contains(out, m[0]) {
				out = append(out, m[0])
			}
		}
	}
	return out
}

// QueryInput is the scm.queries.save body and the filter part of a watch.
type QueryInput struct {
	ID         string   `json:"id,omitempty"`
	Name       string   `json:"name"`
	Provider   Provider `json:"provider"`
	ProjectKey string   `json:"projectKey"`
	Repo       string   `json:"repo"`
	Statuses   []string `json:"statuses"`
	Author     string   `json:"author"`
}

// Validate trims the name and checks every field (FR4.2): a name of 1-100
// characters, a known provider, a repository of that provider, at least one
// state and an author filter. No author means anyone.
func (q *QueryInput) Validate() error {
	q.Name = strings.TrimSpace(q.Name)
	if n := utf8.RuneCountInString(q.Name); n == 0 || n > 100 {
		return invalid(FieldName)
	}
	if _, err := ParseProvider(string(q.Provider)); err != nil {
		return err
	}
	if !projectKeyRe.MatchString(q.ProjectKey) || !ValidRepo(q.Provider, q.Repo) {
		return invalid(FieldRepository)
	}
	if len(q.Statuses) == 0 {
		return invalid(FieldStatuses)
	}
	var states []string
	for _, s := range q.Statuses {
		if !slices.Contains(filterStates, s) {
			return invalid(FieldStatuses)
		}
		if !slices.Contains(states, s) {
			states = append(states, s)
		}
	}
	q.Statuses = states
	if q.Author == "" {
		q.Author = WhoAnyone
	}
	if q.Author != WhoAnyone && q.Author != WhoMe {
		return invalid(FieldAuthor)
	}
	return nil
}

const (
	defaultWatchMinutes = 5
	maxWatchMinutes     = 1440
)

// WatchInput is the scm.watches.save body (FR4.3).
type WatchInput struct {
	QueryInput
	WorkflowID      string `json:"workflowId"`
	WorkflowStepID  string `json:"workflowStepId,omitempty"`
	IntervalMinutes int    `json:"intervalMinutes,omitempty"` // 0 = 5
}

// Validate checks the filters, the workflow and the interval (1-1440
// minutes, 5 when absent).
func (w *WatchInput) Validate() error {
	if err := w.QueryInput.Validate(); err != nil {
		return err
	}
	if strings.TrimSpace(w.WorkflowID) == "" {
		return invalid(FieldWorkflow)
	}
	if w.IntervalMinutes == 0 {
		w.IntervalMinutes = defaultWatchMinutes
	}
	if w.IntervalMinutes < 1 || w.IntervalMinutes > maxWatchMinutes {
		return invalid(FieldInterval)
	}
	return nil
}
