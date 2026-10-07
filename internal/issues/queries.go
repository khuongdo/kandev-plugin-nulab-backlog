package issues

import (
	"context"
	"errors"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/connection"
)

const (
	keyQueries   = "issues.queries"
	maxQueries   = 50
	maxQueryName = 100
)

// IssueQuery is a saved issue list filter (FR4.3). Assignee is "" (anyone),
// "me" (the connected user, resolved by issues.list) or a Backlog user id.
// A missing isDefault reads false, so older documents need no migration.
type IssueQuery struct {
	ID         string  `json:"id,omitempty"`
	Name       string  `json:"name"`
	ProjectKey string  `json:"projectKey,omitempty"`
	StatusIDs  []int64 `json:"statusIds"`
	Assignee   string  `json:"assignee"`
	Keyword    string  `json:"keyword"`
	IsDefault  bool    `json:"isDefault,omitempty"`
}

// Validate trims the name and keyword and checks every field against the
// selected projects, with the limits of saved PR queries and issues.list.
func (q *IssueQuery) Validate(selected []string) error {
	q.Name, q.Keyword = strings.TrimSpace(q.Name), strings.TrimSpace(q.Keyword)
	if n := utf8.RuneCountInString(q.Name); n == 0 || n > maxQueryName {
		return invalid(FieldName)
	}
	if q.ProjectKey != "" && !slices.Contains(selected, q.ProjectKey) {
		return invalid(FieldProjectKey)
	}
	if len(q.StatusIDs) > maxFilterIDs || slices.ContainsFunc(q.StatusIDs, func(id int64) bool { return id <= 0 }) {
		return invalid(FieldStatusIDs)
	}
	if q.Assignee != "" && q.Assignee != WhoMe {
		if id, err := strconv.ParseInt(q.Assignee, 10, 64); err != nil || id <= 0 {
			return invalid(FieldAssignee)
		}
	}
	if utf8.RuneCountInString(q.Keyword) > maxKeyword {
		return invalid(FieldKeyword)
	}
	return nil
}

// Queries returns the workspace's saved issue queries.
func (s *Store) Queries(ctx context.Context, ws string) ([]IssueQuery, error) {
	var d listDoc[IssueQuery]
	err := s.load(ctx, "workspace", ws, keyQueries, &d)
	return d.Items, err
}

// UpdateQueries replaces the saved issue queries with fn's result, at most 50.
func (s *Store) UpdateQueries(ctx context.Context, ws string, fn func([]IssueQuery) ([]IssueQuery, error)) error {
	defer s.lock("queries/" + ws)()
	list, err := s.Queries(ctx, ws)
	if err != nil {
		return err
	}
	next, err := fn(list)
	if errors.Is(err, errUnchanged) {
		return nil
	}
	if err != nil {
		return err
	}
	if len(next) > maxQueries {
		return &connection.FieldError{Field: FieldLimit, Err: errLimit}
	}
	return s.put(ctx, "workspace", ws, keyQueries, listDoc[IssueQuery]{Items: next})
}

// SetQueryDefault stars (isDefault) or un-stars a saved query and returns
// the list. Starring clears every other star in the same write, so at most
// one query is the default (FR4.4).
// ponytail: mirrors git.Service.SetQueryDefault; issues may not import git.
func (s *Service) SetQueryDefault(ctx context.Context, ws, id string, isDefault bool) ([]IssueQuery, error) {
	var out []IssueQuery
	err := s.store.UpdateQueries(ctx, ws, func(list []IssueQuery) ([]IssueQuery, error) {
		i := slices.IndexFunc(list, func(q IssueQuery) bool { return q.ID == id })
		if i < 0 {
			return nil, ErrNotFound
		}
		if isDefault {
			for j := range list {
				list[j].IsDefault = false
			}
		}
		list[i].IsDefault = isDefault
		out = list
		return list, nil
	})
	return out, err
}

// ListQueries returns the saved issue queries, never nil.
func (s *Service) ListQueries(ctx context.Context, ws string) ([]IssueQuery, error) {
	list, err := s.store.Queries(ctx, ws)
	if list == nil && err == nil {
		list = []IssueQuery{}
	}
	return list, err
}

// SaveQuery creates (no id) or replaces a saved issue query (FR4.3). Save
// never sets the star; an edit keeps the stored one.
func (s *Service) SaveQuery(ctx context.Context, ws string, in IssueQuery) (IssueQuery, error) {
	snap, err := s.conn.Current(ctx, ws)
	if err != nil {
		return IssueQuery{}, err
	}
	if err := in.Validate(snap.SelectedProjects); err != nil {
		return IssueQuery{}, err
	}
	in.IsDefault = false
	err = s.store.UpdateQueries(ctx, ws, func(list []IssueQuery) ([]IssueQuery, error) {
		if in.ID == "" {
			in.ID = newID()
			return append(list, in), nil
		}
		i := slices.IndexFunc(list, func(q IssueQuery) bool { return q.ID == in.ID })
		if i < 0 {
			return nil, ErrNotFound
		}
		in.IsDefault = list[i].IsDefault
		list[i] = in
		return list, nil
	})
	return in, err
}

// DeleteQuery removes a saved issue query; deleting the default leaves none.
func (s *Service) DeleteQuery(ctx context.Context, ws, id string) error {
	return s.store.UpdateQueries(ctx, ws, func(list []IssueQuery) ([]IssueQuery, error) {
		n := len(list)
		list = slices.DeleteFunc(list, func(q IssueQuery) bool { return q.ID == id })
		if len(list) == n {
			return nil, ErrNotFound
		}
		return list, nil
	})
}
