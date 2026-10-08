package issues

import (
	"errors"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/backlog"
	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/connection"
)

// Input field names reported with the validation code.
const (
	FieldPage        = "page"
	FieldPageSize    = "pageSize"
	FieldProjectKeys = "projectKeys"
	FieldStatusIDs   = "statusIds"
	FieldAssigneeIDs = "assigneeIds"
	FieldKeyword     = "keyword"
	FieldMinutes     = "minutes"
	FieldIssueKey    = "issueKey"
	FieldTaskID      = "taskId"
	FieldLimit       = "limit"
	FieldWorkflow    = "workflowId"
)

// Link states.
const (
	StateActive       = "active"
	StateNotConnected = "not_connected"
)

const (
	defaultPageSize = 20
	maxPageSize     = 100
	maxFilterIDs    = 50
	maxKeyword      = 100
	// DefaultPollMinutes is the sync interval until an admin sets one (FR4.2).
	DefaultPollMinutes = 5
	maxPollMinutes     = 1440
	// staleAfter is the number of failed checks in a row after which a link
	// "may be out of date" (AC4.1.4).
	staleAfter = 3
)

var errInvalid = errors.New("invalid value")

// ErrNoProject means no project is selected for the workspace (AC1.7.2).
var ErrNoProject = errors.New("no Backlog project is selected")

func invalid(field string) error { return &connection.FieldError{Field: field, Err: errInvalid} }

var issueKeyPattern = regexp.MustCompile(`^([A-Z][A-Z0-9_]*)-([1-9][0-9]{0,8})$`)

// ParseIssueKey splits PROJ-120 into its project key and number.
func ParseIssueKey(key string) (projectKey string, number int, err error) {
	m := issueKeyPattern.FindStringSubmatch(key)
	if m == nil {
		return "", 0, invalid(FieldIssueKey)
	}
	n, _ := strconv.Atoi(m[2]) // at most 9 digits: always fits
	return m[1], n, nil
}

// Query is the issues.list request (C5 IssueQuery).
type Query struct {
	Page        int      `json:"page"`
	PageSize    int      `json:"pageSize"`
	ProjectKeys []string `json:"projectKeys"`
	StatusIDs   []int64  `json:"statusIds"`
	AssigneeIDs []int64  `json:"assigneeIds"`
	Keyword     string   `json:"keyword"`
	// Assignee "me" filters on the connected user, resolved on the server
	// (FR4.2); it replaces AssigneeIDs. Empty means use AssigneeIDs.
	Assignee string `json:"assignee,omitempty"`
}

// ValidateQuery checks a query against the selected projects and fills the
// defaults: page 1, 20 per page, every selected project.
func ValidateQuery(q Query, selected []string) (Query, error) {
	if len(selected) == 0 {
		return Query{}, &connection.FieldError{Field: FieldProjectKeys, Err: ErrNoProject}
	}
	switch {
	case q.Page < 0:
		return Query{}, invalid(FieldPage)
	case q.PageSize < 0 || q.PageSize > maxPageSize:
		return Query{}, invalid(FieldPageSize)
	case len(q.StatusIDs) > maxFilterIDs:
		return Query{}, invalid(FieldStatusIDs)
	case len(q.AssigneeIDs) > maxFilterIDs:
		return Query{}, invalid(FieldAssigneeIDs)
	case q.Assignee != "" && q.Assignee != WhoMe:
		return Query{}, invalid(FieldAssignee)
	}
	for _, k := range q.ProjectKeys {
		if !slices.Contains(selected, k) {
			return Query{}, invalid(FieldProjectKeys)
		}
	}
	q.Keyword = strings.TrimSpace(q.Keyword)
	if utf8.RuneCountInString(q.Keyword) > maxKeyword {
		return Query{}, invalid(FieldKeyword)
	}
	if q.Page == 0 {
		q.Page = 1
	}
	if q.PageSize == 0 {
		q.PageSize = defaultPageSize
	}
	if len(q.ProjectKeys) == 0 {
		q.ProjectKeys = slices.Clone(selected)
	}
	return q, nil
}

// ShowingRange is the "Showing from-to of total" line of a page; zeros when
// the page is empty.
func ShowingRange(page, size, total int) (from, to int, hasNext bool) {
	from = (page-1)*size + 1
	if total == 0 || from > total {
		return 0, 0, false
	}
	to = min(page*size, total)
	return from, to, to < total
}

// PriorityFor maps a Backlog priority id to a Kandev task priority (AC3.1.1).
func PriorityFor(id int64) string {
	switch id {
	case 2:
		return "high"
	case 4:
		return "low"
	default:
		return "medium"
	}
}

// NewTask is a Kandev task created from an issue.
type NewTask struct {
	WorkspaceID    string
	WorkflowID     string
	WorkflowStepID string
	Title          string
	Description    string
	Priority       string
}

// IssueURL is the issue's page on the space.
func IssueURL(host, key string) string { return "https://" + host + "/view/" + key }

// NewTaskFor is the title, description and priority of a task created from
// issue (AC3.1.1). Comments and attachments are not copied (Q9).
func NewTaskFor(issue backlog.Issue, host string) NewTask {
	link := "Backlog: " + IssueURL(host, issue.IssueKey)
	desc := link
	if d := strings.TrimSpace(issue.Description); d != "" {
		desc = d + "\n\n" + link
	}
	return NewTask{Title: issue.Summary, Description: desc, Priority: PriorityFor(issue.PriorityID)}
}

// ValidatePollMinutes accepts a whole number of minutes from 1 to 1440 (AC4.2.2).
func ValidatePollMinutes(raw any) (int, error) {
	f, ok := raw.(float64)
	if !ok || f != float64(int(f)) || f < 1 || f > maxPollMinutes {
		return 0, invalid(FieldMinutes)
	}
	return int(f), nil
}

// Link associates a Kandev task with a Backlog issue (ADR-004).
type Link struct {
	IssueKey        string `json:"issueKey"`
	IssueID         int64  `json:"issueId"`
	ProjectKey      string `json:"projectKey"`
	SpaceHost       string `json:"spaceHost"`
	TaskID          string `json:"taskId"`
	TaskKey         string `json:"taskKey,omitempty"`
	Summary         string `json:"summary,omitempty"`
	State           string `json:"state"`
	Unavailable     bool   `json:"unavailable,omitempty"`
	LastKnownStatus string `json:"lastKnownStatus,omitempty"`
	StatusUpdatedAt string `json:"statusUpdatedAt,omitempty"`
	FailCount       int    `json:"failCount,omitempty"`
	ConnectionEpoch int    `json:"connectionEpoch"`
	CreatedAt       string `json:"createdAt"`
}

// Stale reports whether the last checks failed often enough that the shown
// status may be out of date (AC4.1.4).
func (l Link) Stale() bool { return l.FailCount >= staleAfter }

// Matches reports whether issue passes the query's project, status and
// assignee filters (AC2.2.1); an empty filter matches everything.
func Matches(issue backlog.Issue, q backlog.IssueQuery) bool {
	in := func(ids []int64, id int64) bool { return len(ids) == 0 || slices.Contains(ids, id) }
	return in(q.ProjectIDs, issue.ProjectID) && in(q.StatusIDs, issue.StatusID) && in(q.AssigneeIDs, issue.AssigneeID)
}

// SuggestMatch reports whether issue fits a `#` query: the key starts with
// q or the title contains it, ignoring case (AC3.4.1).
func SuggestMatch(issue backlog.Issue, q string) bool {
	q = strings.ToLower(q)
	return strings.HasPrefix(strings.ToLower(issue.IssueKey), q) || strings.Contains(strings.ToLower(issue.Summary), q)
}
