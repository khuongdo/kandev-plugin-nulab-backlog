package git

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/backlog"
	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/connection"
)

// ProviderID is the repository and review provider id this plugin owns (C8).
const ProviderID = "nulab-backlog"

// Input fields reported with CodeValidation by U4 actions (C5).
const (
	FieldReference  = "reference"
	FieldName       = "name"
	FieldRepository = "repository"
	FieldStatuses   = "statuses"
	FieldAssignee   = "assignee"
	FieldCreator    = "creator"
	FieldIssueKey   = "issueKey"
	FieldWorkflow   = "workflowId"
	FieldTitle      = "title"
	FieldBranch     = "branch"
	FieldBaseBranch = "baseBranch"
	FieldCursor     = "cursor"
	FieldURL        = "url"
	FieldID         = "id"
)

// Who values for the watch and query assignee/creator filters.
const (
	WhoAnyone = "anyone"
	WhoMe     = "me"
)

var (
	projectKeyRe = regexp.MustCompile(`^[A-Z][A-Z0-9_]{0,24}$`)
	repoNameRe   = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,99}$`)
	issueKeyRe   = regexp.MustCompile(`^([A-Z][A-Z0-9_]*)-[1-9][0-9]*$`)
	issueKeyText = regexp.MustCompile(`\b([A-Z][A-Z0-9_]*)-[1-9][0-9]*\b`)
	prNumberRe   = regexp.MustCompile(`^[1-9][0-9]{0,8}$`)

	errReference = errors.New("must be a pull request URL in this space, <repository>#<number> or <number>")
	errInvalid   = errors.New("invalid value")
)

func invalid(field string) error { return &connection.FieldError{Field: field, Err: errInvalid} }

// RepoRef is a Backlog repository by project key and name.
type RepoRef struct {
	ProjectKey string
	Name       string
}

// Reference is a parsed pull request reference.
type Reference struct {
	ProjectKey string
	Repo       string
	Number     int
}

// ParseReference accepts https://<space>/git/<PROJ>/<repo>/pullRequests/<n>,
// <repo>#<n> and <n>. The short forms need def: the task's one Backlog
// repository. Any other input is a validation error on field reference.
func ParseReference(raw, spaceHost string, def *RepoRef) (Reference, error) {
	s := strings.TrimSpace(raw)
	bad := &connection.FieldError{Field: FieldReference, Err: errReference}
	if strings.Contains(s, "://") {
		u, err := url.Parse(s)
		if err != nil || u.Scheme != "https" || !strings.EqualFold(u.Host, spaceHost) || u.User != nil ||
			u.RawQuery != "" || u.ForceQuery || u.Fragment != "" {
			return Reference{}, bad
		}
		parts := strings.Split(strings.TrimPrefix(u.Path, "/"), "/")
		if len(parts) != 5 || parts[0] != "git" || parts[3] != "pullRequests" ||
			!projectKeyRe.MatchString(parts[1]) || !repoNameRe.MatchString(parts[2]) || !prNumberRe.MatchString(parts[4]) {
			return Reference{}, bad
		}
		n, _ := strconv.Atoi(parts[4])
		return Reference{ProjectKey: parts[1], Repo: parts[2], Number: n}, nil
	}
	repo, num, found := strings.Cut(s, "#")
	if !found {
		repo, num = "", s
	}
	if def == nil || !prNumberRe.MatchString(num) || (repo != "" && !repoNameRe.MatchString(repo)) {
		return Reference{}, bad
	}
	if repo == "" {
		repo = def.Name
	}
	n, _ := strconv.Atoi(num)
	return Reference{ProjectKey: def.ProjectKey, Repo: repo, Number: n}, nil
}

// ParseClonePath reads /git/<PROJ>/<repo>.git or /git/<PROJ>/<repo>.
func ParseClonePath(path string) (project, repo string, ok bool) {
	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
	if len(parts) != 3 || parts[0] != "git" || !projectKeyRe.MatchString(parts[1]) {
		return "", "", false
	}
	repo = strings.TrimSuffix(parts[2], ".git")
	if !repoNameRe.MatchString(repo) {
		return "", "", false
	}
	return parts[1], repo, true
}

// LinkKey identifies a pull request: spaceHost|repositoryId|number. It is the
// reviewKey of a link and the key of a watch ledger entry.
func LinkKey(spaceHost string, repositoryID int64, number int) string {
	return fmt.Sprintf("%s|%d|%d", spaceHost, repositoryID, number)
}

// prURL is the pull request page in Backlog.
func prURL(spaceHost, projectKey, repo string, number int) string {
	return fmt.Sprintf("https://%s/git/%s/%s/pullRequests/%d", spaceHost, projectKey, url.PathEscape(repo), number)
}

// stateIDs maps a status filter value to the Backlog status id.
var stateIDs = map[string]int64{"open": 1, "closed": 2, "merged": 3}

// WatchInput is the git.watches.save body.
type WatchInput struct {
	ID             string   `json:"id,omitempty"`
	Name           string   `json:"name"`
	ProjectKey     string   `json:"projectKey"`
	RepoName       string   `json:"repoName"`
	Statuses       []string `json:"statuses"`
	Assignee       string   `json:"assignee"`
	Creator        string   `json:"creator"`
	IssueKey       string   `json:"issueKey,omitempty"`
	WorkflowID     string   `json:"workflowId"`
	WorkflowStepID string   `json:"workflowStepId,omitempty"`
}

// Validate trims the name and checks every field against the selected
// projects (AC6.1.2). The workflow is checked by the service.
func (w *WatchInput) Validate(selected []string) error {
	w.Name = strings.TrimSpace(w.Name)
	if n := utf8.RuneCountInString(w.Name); n == 0 || n > 100 {
		return invalid(FieldName)
	}
	if err := validateRepo(w.ProjectKey, w.RepoName, selected); err != nil {
		return err
	}
	statuses, err := validateStatuses(w.Statuses)
	if err != nil {
		return err
	}
	w.Statuses = statuses
	if !validWho(w.Assignee) {
		return invalid(FieldAssignee)
	}
	if !validWho(w.Creator) {
		return invalid(FieldCreator)
	}
	if w.IssueKey != "" {
		m := issueKeyRe.FindStringSubmatch(w.IssueKey)
		if m == nil || !slices.Contains(selected, m[1]) {
			return invalid(FieldIssueKey)
		}
	}
	return nil
}

// QueryInput is the git.queries.save body.
type QueryInput struct {
	ID         string   `json:"id,omitempty"`
	Name       string   `json:"name"`
	ProjectKey string   `json:"projectKey"`
	RepoName   string   `json:"repoName"`
	Statuses   []string `json:"statuses"`
	Assignee   string   `json:"assignee"`
}

// Validate trims the name and checks the fields.
func (q *QueryInput) Validate(selected []string) error {
	q.Name = strings.TrimSpace(q.Name)
	if n := utf8.RuneCountInString(q.Name); n == 0 || n > 100 {
		return invalid(FieldName)
	}
	if err := validateRepo(q.ProjectKey, q.RepoName, selected); err != nil {
		return err
	}
	statuses, err := validateStatuses(q.Statuses)
	if err != nil {
		return err
	}
	q.Statuses = statuses
	if !validWho(q.Assignee) {
		return invalid(FieldAssignee)
	}
	return nil
}

func validateRepo(project, repo string, selected []string) error {
	if !projectKeyRe.MatchString(project) || !slices.Contains(selected, project) || !repoNameRe.MatchString(repo) {
		return invalid(FieldRepository)
	}
	return nil
}

func validateStatuses(in []string) ([]string, error) {
	if len(in) == 0 {
		return nil, invalid(FieldStatuses)
	}
	var out []string
	for _, s := range in {
		if _, ok := stateIDs[s]; !ok {
			return nil, invalid(FieldStatuses)
		}
		if !slices.Contains(out, s) {
			out = append(out, s)
		}
	}
	return out, nil
}

func validWho(s string) bool { return s == WhoAnyone || s == WhoMe }

// RelatedIssueKey returns the first issue key in title, then body, whose
// project is selected, or "".
func RelatedIssueKey(title, body string, selected []string) string {
	for _, m := range issueKeyText.FindAllStringSubmatch(title+"\n"+body, -1) {
		if slices.Contains(selected, m[1]) {
			return m[0]
		}
	}
	return ""
}

// PRDescription keeps a non-empty body. An empty body becomes
// "Related: <KEY>", or the title without a key. The plugin never writes a
// closing keyword before the key (AC5.3.1).
func PRDescription(body, title, key string) string {
	switch {
	case strings.TrimSpace(body) != "":
		return body
	case key != "":
		return "Related: " + key
	default:
		return title
	}
}

// BranchCandidates lists the base and branch names of prs, de-duplicated,
// most used first, ties by name.
func BranchCandidates(prs []backlog.PullRequest) []string {
	counts := map[string]int{}
	for _, pr := range prs {
		for _, b := range []string{pr.Base, pr.Branch} {
			if b != "" {
				counts[b]++
			}
		}
	}
	return byCount(counts)
}

// DefaultBranch is the most used base branch of prs, or "master" (Backlog's
// default for a new repository) when no pull request names one. Backlog's
// API has no default-branch field.
func DefaultBranch(prs []backlog.PullRequest) string {
	counts := map[string]int{}
	for _, pr := range prs {
		if pr.Base != "" {
			counts[pr.Base]++
		}
	}
	if names := byCount(counts); len(names) > 0 {
		return names[0]
	}
	return "master"
}

func byCount(counts map[string]int) []string {
	out := make([]string, 0, len(counts))
	for name := range counts {
		out = append(out, name)
	}
	sort.Slice(out, func(i, j int) bool {
		if counts[out[i]] != counts[out[j]] {
			return counts[out[i]] > counts[out[j]]
		}
		return out[i] < out[j]
	})
	return out
}
