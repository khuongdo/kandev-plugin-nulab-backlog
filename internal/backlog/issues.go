package backlog

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
)

// Issue is a Backlog issue (GET /api/v2/issues, /api/v2/issues/:issueIdOrKey).
type Issue struct {
	ID           int64
	ProjectID    int64
	IssueKey     string
	Summary      string
	Description  string // user content: never logged or formatted
	StatusID     int64
	StatusName   string
	PriorityID   int64
	PriorityName string
	AssigneeID   int64 // 0 = no assignee
	AssigneeName string
	DueDate      string
	Updated      string
}

// String hides the description, which is user content.
func (i Issue) String() string {
	return fmt.Sprintf("backlog.Issue{IssueKey: %s, ID: %d, StatusID: %d}", i.IssueKey, i.ID, i.StatusID)
}

// GoString hides the description for %#v as well.
func (i Issue) GoString() string { return i.String() }

// Format hides the description for every verb, including %+v.
func (i Issue) Format(f fmt.State, _ rune) { _, _ = fmt.Fprint(f, i.String()) }

// Comment is one issue comment, newest first in Backlog's order.
type Comment struct {
	ID         int64
	Content    string // user content: never logged or formatted
	AuthorName string
	Created    string
}

// String hides the content.
func (c Comment) String() string { return fmt.Sprintf("backlog.Comment{ID: %d}", c.ID) }

// GoString hides the content for %#v as well.
func (c Comment) GoString() string { return c.String() }

// Format hides the content for every verb.
func (c Comment) Format(f fmt.State, _ rune) { _, _ = fmt.Fprint(f, c.String()) }

// Attachment is one file attached to an issue; Size is in bytes.
type Attachment struct {
	ID   int64
	Name string
	Size int64
}

// Status is one status of a project.
type Status struct {
	ID   int64
	Name string
}

// ProjectUser is one member of a project.
type ProjectUser struct {
	ID   int64
	Name string
}

// IssueQuery filters GET /api/v2/issues and /api/v2/issues/count. Results
// are always newest-updated first.
type IssueQuery struct {
	ProjectIDs  []int64
	StatusIDs   []int64
	AssigneeIDs []int64
	IDs         []int64
	Keyword     string
	Offset      int
	Count       int // clamped to 1-100
}

// Values encodes the query string.
func (q IssueQuery) Values() url.Values {
	v := url.Values{}
	add := func(key string, ids []int64) {
		for _, id := range ids {
			v.Add(key, strconv.FormatInt(id, 10))
		}
	}
	add("projectId[]", q.ProjectIDs)
	add("statusId[]", q.StatusIDs)
	add("assigneeId[]", q.AssigneeIDs)
	add("id[]", q.IDs)
	if q.Keyword != "" {
		v.Set("keyword", q.Keyword)
	}
	v.Set("sort", "updated")
	v.Set("order", "desc")
	if q.Offset > 0 {
		v.Set("offset", strconv.Itoa(q.Offset))
	}
	v.Set("count", strconv.Itoa(min(max(q.Count, 1), 100)))
	return v
}

// CommentQuery pages GET .../comments, newest first: MaxID 0 means the newest.
type CommentQuery struct {
	MaxID int64
	Count int // clamped to 1-100
}

// Values encodes the query string.
func (q CommentQuery) Values() url.Values {
	v := url.Values{"order": {"desc"}, "count": {strconv.Itoa(min(max(q.Count, 1), 100))}}
	if q.MaxID > 0 {
		v.Set("maxId", strconv.FormatInt(q.MaxID, 10))
	}
	return v
}

var issueRefPattern = regexp.MustCompile(`^(?:[A-Z][A-Z0-9_]*-[1-9][0-9]{0,8}|[1-9][0-9]{0,18})$`)

// ValidIssueRef reports whether ref is an issue key (PROJ-120) or a numeric
// id, the only forms put into a request path.
func ValidIssueRef(ref string) bool { return issueRefPattern.MatchString(ref) }

type rawUser struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type rawIssue struct {
	ID          *int64 `json:"id"`
	ProjectID   int64  `json:"projectId"`
	IssueKey    string `json:"issueKey"`
	Summary     string `json:"summary"`
	Description string `json:"description"`
	Status      *struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	} `json:"status"`
	Priority *struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	} `json:"priority"`
	Assignee *rawUser `json:"assignee"`
	DueDate  *string  `json:"dueDate"`
	Updated  string   `json:"updated"`
}

func (r rawIssue) issue() (Issue, bool) {
	if r.ID == nil || r.IssueKey == "" {
		return Issue{}, false
	}
	i := Issue{ID: *r.ID, ProjectID: r.ProjectID, IssueKey: r.IssueKey, Summary: r.Summary,
		Description: r.Description, Updated: r.Updated}
	if r.Status != nil {
		i.StatusID, i.StatusName = r.Status.ID, r.Status.Name
	}
	if r.Priority != nil {
		i.PriorityID, i.PriorityName = r.Priority.ID, r.Priority.Name
	}
	if r.Assignee != nil {
		i.AssigneeID, i.AssigneeName = r.Assignee.ID, r.Assignee.Name
	}
	if r.DueDate != nil {
		i.DueDate = *r.DueDate
	}
	return i, true
}

func parseIssue(body []byte) (Issue, error) {
	var raw rawIssue
	if err := json.Unmarshal(body, &raw); err != nil {
		return Issue{}, errBody()
	}
	i, ok := raw.issue()
	if !ok {
		return Issue{}, errBody()
	}
	return i, nil
}

func parseIssues(body []byte) ([]Issue, error) {
	var raw []rawIssue
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, errBody()
	}
	out := make([]Issue, 0, len(raw))
	for _, r := range raw {
		i, ok := r.issue()
		if !ok {
			return nil, errBody()
		}
		out = append(out, i)
	}
	return out, nil
}

func parseCount(body []byte) (int, error) {
	var raw struct {
		Count *int `json:"count"`
	}
	if err := json.Unmarshal(body, &raw); err != nil || raw.Count == nil || *raw.Count < 0 {
		return 0, errBody()
	}
	return *raw.Count, nil
}

func parseComments(body []byte) ([]Comment, error) {
	var raw []struct {
		ID          *int64   `json:"id"`
		Content     string   `json:"content"`
		CreatedUser *rawUser `json:"createdUser"`
		Created     string   `json:"created"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, errBody()
	}
	out := make([]Comment, 0, len(raw))
	for _, r := range raw {
		if r.ID == nil {
			return nil, errBody()
		}
		c := Comment{ID: *r.ID, Content: r.Content, Created: r.Created}
		if r.CreatedUser != nil {
			c.AuthorName = r.CreatedUser.Name
		}
		out = append(out, c)
	}
	return out, nil
}

func parseAttachments(body []byte) ([]Attachment, error) {
	var raw []struct {
		ID   *int64 `json:"id"`
		Name string `json:"name"`
		Size int64  `json:"size"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, errBody()
	}
	out := make([]Attachment, 0, len(raw))
	for _, r := range raw {
		if r.ID == nil || r.Name == "" {
			return nil, errBody()
		}
		out = append(out, Attachment{ID: *r.ID, Name: r.Name, Size: r.Size})
	}
	return out, nil
}

// parseNamed decodes a list of {id, name} objects; each needs a numeric id.
func parseNamed(body []byte) ([]rawUser, error) {
	var raw []struct {
		ID   *int64 `json:"id"`
		Name string `json:"name"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, errBody()
	}
	out := make([]rawUser, 0, len(raw))
	for _, r := range raw {
		if r.ID == nil {
			return nil, errBody()
		}
		out = append(out, rawUser{ID: *r.ID, Name: r.Name})
	}
	return out, nil
}

func parseStatuses(body []byte) ([]Status, error) {
	named, err := parseNamed(body)
	if err != nil {
		return nil, err
	}
	out := make([]Status, 0, len(named))
	for _, n := range named {
		out = append(out, Status(n))
	}
	return out, nil
}

func parseProjectUsers(body []byte) ([]ProjectUser, error) {
	named, err := parseNamed(body)
	if err != nil {
		return nil, err
	}
	out := make([]ProjectUser, 0, len(named))
	for _, n := range named {
		out = append(out, ProjectUser(n))
	}
	return out, nil
}

const issuesPath = "/api/v2/issues"

var projectKeyPattern = regexp.MustCompile(`^[A-Z][A-Z0-9_]*$`)

// errBadRef is a fresh Invalid error for a reference that is never put into
// a request path.
func errBadRef() error { return &Error{Kind: KindInvalid, Class: "request"} }

// get sends one GET and returns its body.
func (c *Client) get(ctx context.Context, creds Credentials, class CallClass, path string, q url.Values) ([]byte, error) {
	return c.send(ctx, request{method: http.MethodGet, path: path, query: q, creds: creds, class: class})
}

// Issues lists issues, newest updated first (GET /api/v2/issues, Search group).
func (c *Client) Issues(ctx context.Context, creds Credentials, class CallClass, q IssueQuery) ([]Issue, error) {
	body, err := c.get(ctx, creds, class, issuesPath, q.Values())
	if err != nil {
		return nil, err
	}
	return parseIssues(body)
}

// IssueCount counts the issues a query matches (GET /api/v2/issues/count, Search group).
func (c *Client) IssueCount(ctx context.Context, creds Credentials, class CallClass, q IssueQuery) (int, error) {
	v := q.Values()
	for _, k := range []string{"sort", "order", "offset", "count"} {
		v.Del(k)
	}
	body, err := c.get(ctx, creds, class, issuesPath+"/count", v)
	if err != nil {
		return 0, err
	}
	return parseCount(body)
}

// Issue returns one issue by key or id (GET /api/v2/issues/:issueIdOrKey, C1).
// An invalid ref is refused before any request.
func (c *Client) Issue(ctx context.Context, creds Credentials, class CallClass, ref string) (Issue, error) {
	if !ValidIssueRef(ref) {
		return Issue{}, errBadRef()
	}
	body, err := c.get(ctx, creds, class, issuesPath+"/"+ref, nil)
	if err != nil {
		return Issue{}, err
	}
	return parseIssue(body)
}

// IssueComments returns one page of an issue's comments, newest first.
func (c *Client) IssueComments(ctx context.Context, creds Credentials, class CallClass, ref string, q CommentQuery) ([]Comment, error) {
	if !ValidIssueRef(ref) {
		return nil, errBadRef()
	}
	body, err := c.get(ctx, creds, class, issuesPath+"/"+ref+"/comments", q.Values())
	if err != nil {
		return nil, err
	}
	return parseComments(body)
}

// IssueAttachments lists the files attached to an issue.
func (c *Client) IssueAttachments(ctx context.Context, creds Credentials, class CallClass, ref string) ([]Attachment, error) {
	if !ValidIssueRef(ref) {
		return nil, errBadRef()
	}
	body, err := c.get(ctx, creds, class, issuesPath+"/"+ref+"/attachments", nil)
	if err != nil {
		return nil, err
	}
	return parseAttachments(body)
}

// ProjectStatuses lists a project's statuses (Interactive).
func (c *Client) ProjectStatuses(ctx context.Context, creds Credentials, projectKey string) ([]Status, error) {
	if !projectKeyPattern.MatchString(projectKey) {
		return nil, errBadRef()
	}
	body, err := c.get(ctx, creds, Interactive, projectsPath+"/"+projectKey+"/statuses", nil)
	if err != nil {
		return nil, err
	}
	return parseStatuses(body)
}

// ProjectUsers lists a project's members (Interactive).
func (c *Client) ProjectUsers(ctx context.Context, creds Credentials, projectKey string) ([]ProjectUser, error) {
	if !projectKeyPattern.MatchString(projectKey) {
		return nil, errBadRef()
	}
	body, err := c.get(ctx, creds, Interactive, projectsPath+"/"+projectKey+"/users", nil)
	if err != nil {
		return nil, err
	}
	return parseProjectUsers(body)
}
