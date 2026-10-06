package issues

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/backlog"
	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/connection"
)

func fieldOf(t *testing.T, err error) string {
	t.Helper()
	var fe *connection.FieldError
	require.True(t, errors.As(err, &fe), "want a FieldError, got %v", err)
	return fe.Field
}

func TestU3_Types_ParseIssueKey(t *testing.T) {
	cases := []struct {
		key     string
		project string
		number  int
		ok      bool
	}{
		{"PROJ-120", "PROJ", 120, true},
		{"MY_APP2-7", "MY_APP2", 7, true},
		{"proj-120", "", 0, false},
		{"PROJ-0", "", 0, false},
		{"PROJ", "", 0, false},
		{"PROJ-12a", "", 0, false},
		{"", "", 0, false},
	}
	for _, tc := range cases {
		t.Run(tc.key, func(t *testing.T) {
			project, n, err := ParseIssueKey(tc.key)
			if !tc.ok {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.project, project)
			require.Equal(t, tc.number, n)
		})
	}
}

func TestU3_Types_ValidateQuery(t *testing.T) {
	selected := []string{"PROJ", "DEMO"}
	many := make([]int64, 51)
	cases := []struct {
		name  string
		in    Query
		want  Query
		field string
	}{
		{"defaults", Query{}, Query{Page: 1, PageSize: 20, ProjectKeys: selected}, ""},
		{"trimmed keyword and subset", Query{Page: 3, PageSize: 100, ProjectKeys: []string{"DEMO"}, Keyword: "  login  "},
			Query{Page: 3, PageSize: 100, ProjectKeys: []string{"DEMO"}, Keyword: "login"}, ""},
		{"negative page", Query{Page: -1}, Query{}, FieldPage},
		{"page size over 100", Query{PageSize: 101}, Query{}, FieldPageSize},
		{"negative page size", Query{PageSize: -1}, Query{}, FieldPageSize},
		{"unselected project", Query{ProjectKeys: []string{"OTHER"}}, Query{}, FieldProjectKeys},
		{"too many statuses", Query{StatusIDs: many}, Query{}, FieldStatusIDs},
		{"too many assignees", Query{AssigneeIDs: many}, Query{}, FieldAssigneeIDs},
		{"long keyword", Query{Keyword: strings.Repeat("あ", 101)}, Query{}, FieldKeyword},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ValidateQuery(tc.in, selected)
			if tc.field != "" {
				require.Equal(t, tc.field, fieldOf(t, err))
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
	_, err := ValidateQuery(Query{}, nil)
	require.ErrorIs(t, err, ErrNoProject, "AC1.7.2: no selected project")
	require.Equal(t, FieldProjectKeys, fieldOf(t, err))
	ok, err := ValidateQuery(Query{Keyword: strings.Repeat("あ", 100)}, selected)
	require.NoError(t, err, "100 runes is fine")
	require.Len(t, []rune(ok.Keyword), 100)
}

func TestU3_Types_ShowingRange(t *testing.T) {
	cases := []struct {
		page, size, total, from, to int
		next                        bool
	}{
		{1, 20, 57, 1, 20, true},
		{2, 20, 57, 21, 40, true},
		{3, 20, 57, 41, 57, false},
		{1, 20, 20, 1, 20, false},
		{1, 20, 0, 0, 0, false},
		{9, 20, 57, 0, 0, false},
	}
	for _, tc := range cases {
		from, to, next := ShowingRange(tc.page, tc.size, tc.total)
		require.Equal(t, [3]any{tc.from, tc.to, tc.next}, [3]any{from, to, next}, "%+v", tc)
	}
}

func TestU3_Types_PriorityAndNewTask(t *testing.T) {
	for id, want := range map[int64]string{2: "high", 3: "medium", 4: "low", 1: "medium", 0: "medium", 99: "medium"} {
		require.Equal(t, want, PriorityFor(id), "priority %d", id)
	}
	task := NewTaskFor(backlog.Issue{IssueKey: "PROJ-120", Summary: "Login page", Description: "Users cannot log in.", PriorityID: 2},
		"example-space.backlog.com")
	require.Equal(t, NewTask{Title: "Login page", Priority: "high",
		Description: "Users cannot log in.\n\nBacklog: https://example-space.backlog.com/view/PROJ-120"}, task)
	empty := NewTaskFor(backlog.Issue{IssueKey: "PROJ-1", Summary: "S"}, "example-space.backlog.com")
	require.Equal(t, "Backlog: https://example-space.backlog.com/view/PROJ-1", empty.Description)
	require.Equal(t, "medium", empty.Priority)
}

func TestU3_Types_ValidatePollMinutes(t *testing.T) {
	for _, raw := range []any{0.5, 0.0, -1.0, 1441.0, "abc", "", "5", nil, true} {
		_, err := ValidatePollMinutes(raw)
		require.Equal(t, FieldMinutes, fieldOf(t, err), "%v", raw)
	}
	for raw, want := range map[float64]int{1: 1, 5: 5, 1440: 1440} {
		got, err := ValidatePollMinutes(raw)
		require.NoError(t, err)
		require.Equal(t, want, got)
	}
	require.Equal(t, 5, DefaultPollMinutes)
}

func TestU3_Types_LinkStale(t *testing.T) {
	require.False(t, Link{FailCount: 2}.Stale())
	require.True(t, Link{FailCount: 3}.Stale(), "AC4.1.4: may be out of date after 3 failed cycles")
}

func TestU3_Types_Matches(t *testing.T) {
	issue := backlog.Issue{ProjectID: 101, StatusID: 2, AssigneeID: 7}
	cases := map[string]struct {
		q    backlog.IssueQuery
		want bool
	}{
		"no filter":        {backlog.IssueQuery{}, true},
		"project matches":  {backlog.IssueQuery{ProjectIDs: []int64{101}}, true},
		"other project":    {backlog.IssueQuery{ProjectIDs: []int64{102}}, false},
		"status matches":   {backlog.IssueQuery{StatusIDs: []int64{1, 2}}, true},
		"other status":     {backlog.IssueQuery{StatusIDs: []int64{3}}, false},
		"assignee matches": {backlog.IssueQuery{AssigneeIDs: []int64{7}}, true},
		"other assignee":   {backlog.IssueQuery{AssigneeIDs: []int64{8}}, false},
	}
	for name, tc := range cases {
		require.Equal(t, tc.want, Matches(issue, tc.q), name)
	}
}

func TestU3_Types_SuggestMatch(t *testing.T) {
	issue := backlog.Issue{IssueKey: "PROJ-123", Summary: "Fix Login timeout"}
	for q, want := range map[string]bool{"PROJ-12": true, "proj-1": true, "login": true, "TIMEOUT": true, "": true,
		"DEMO": false, "123": false, "logout": false} {
		require.Equal(t, want, SuggestMatch(issue, q), q)
	}
}
