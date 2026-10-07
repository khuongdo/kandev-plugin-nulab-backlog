package backlog

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestU3_ParseIssues_DecodesTheListFields(t *testing.T) {
	issues, err := parseIssues(readFixture(t, "issues_ok.json"))
	require.NoError(t, err)
	require.Len(t, issues, 3)
	require.Equal(t, Issue{ID: 5118, ProjectID: 101, IssueKey: "PROJ-118", Summary: "Fix login timeout",
		Description: "Steps to reproduce are in the attachment.", StatusID: 2, StatusName: "In Progress",
		PriorityID: 2, PriorityName: "High", AssigneeID: 2, AssigneeName: "Lan", DueDate: "2026-10-10T00:00:00Z",
		Updated: "2026-10-01T09:00:00Z", Created: "2026-09-20T09:00:00Z"}, issues[0])
	require.Zero(t, issues[1].AssigneeID, "no assignee")
	require.Empty(t, issues[1].DueDate)
	require.Equal(t, 200, len([]rune(issues[2].Summary)), "the long Japanese summary survives")
}

func TestU3_ParseIssue_DetailAndCount(t *testing.T) {
	issue, err := parseIssue(readFixture(t, "issue_detail_ok.json"))
	require.NoError(t, err)
	require.Equal(t, "PROJ-118", issue.IssueKey)
	require.Equal(t, int64(2), issue.PriorityID)

	n, err := parseCount(readFixture(t, "issues_count_ok.json"))
	require.NoError(t, err)
	require.Equal(t, 57, n)
}

func TestU3_ParseSmallTypes(t *testing.T) {
	comments, err := parseComments(readFixture(t, "comments_ok.json"))
	require.NoError(t, err)
	require.Equal(t, []Comment{{ID: 902, Content: "Fixed on staging.", AuthorName: "Lan", Created: "2026-10-01T09:00:00Z"},
		{ID: 901, Content: "I can reproduce this.", AuthorName: "Test User", Created: "2026-09-30T09:00:00Z"}}, comments)

	atts, err := parseAttachments(readFixture(t, "attachments_ok.json"))
	require.NoError(t, err)
	require.Equal(t, []Attachment{{ID: 8, Name: "spec.pdf", Size: 1258291}, {ID: 9, Name: "dump.zip", Size: 50331648}}, atts)

	statuses, err := parseStatuses(readFixture(t, "statuses_ok.json"))
	require.NoError(t, err)
	require.Equal(t, Status{ID: 2, Name: "In Progress"}, statuses[1])

	users, err := parseProjectUsers(readFixture(t, "project_users_ok.json"))
	require.NoError(t, err)
	require.Equal(t, []ProjectUser{{ID: 1, Name: "Test User"}, {ID: 2, Name: "Lan"}}, users)
}

func TestU3_Parse_RejectsBadBodies(t *testing.T) {
	cases := map[string]struct {
		parse func([]byte) error
		body  string
	}{
		"issue string id":    {func(b []byte) error { _, err := parseIssue(b); return err }, `{"id":"5118","issueKey":"PROJ-118"}`},
		"issue empty key":    {func(b []byte) error { _, err := parseIssue(b); return err }, `{"id":5118,"issueKey":""}`},
		"issue list no id":   {func(b []byte) error { _, err := parseIssues(b); return err }, `[{"issueKey":"PROJ-1"}]`},
		"issue list object":  {func(b []byte) error { _, err := parseIssues(b); return err }, `{}`},
		"count missing":      {func(b []byte) error { _, err := parseCount(b); return err }, `{}`},
		"comment no id":      {func(b []byte) error { _, err := parseComments(b); return err }, `[{"content":"x"}]`},
		"attachment no name": {func(b []byte) error { _, err := parseAttachments(b); return err }, `[{"id":1,"size":2}]`},
		"status no id":       {func(b []byte) error { _, err := parseStatuses(b); return err }, `[{"name":"Open"}]`},
		"user no id":         {func(b []byte) error { _, err := parseProjectUsers(b); return err }, `[{"name":"Lan"}]`},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := tc.parse([]byte(tc.body))
			var be *Error
			require.ErrorAs(t, err, &be)
			require.Equal(t, KindUnreachable, be.Kind)
			require.Equal(t, "body", be.Class)
		})
	}
}

func TestU3_IssueQuery_EncodesTheFilters(t *testing.T) {
	q := IssueQuery{ProjectIDs: []int64{101, 102}, StatusIDs: []int64{1}, AssigneeIDs: []int64{2}, Keyword: "login", Offset: 40, Count: 20}
	v := q.Values()
	require.Equal(t, []string{"101", "102"}, v["projectId[]"])
	require.Equal(t, []string{"1"}, v["statusId[]"])
	require.Equal(t, []string{"2"}, v["assigneeId[]"])
	require.Equal(t, "login", v.Get("keyword"))
	require.Equal(t, "updated", v.Get("sort"))
	require.Equal(t, "desc", v.Get("order"))
	require.Equal(t, "40", v.Get("offset"))
	require.Equal(t, "20", v.Get("count"))

	for count, want := range map[int]string{0: "1", -5: "1", 100: "100", 500: "100"} {
		require.Equal(t, want, IssueQuery{Count: count}.Values().Get("count"), "count %d", count)
	}
	empty := IssueQuery{Count: 20}.Values()
	require.NotContains(t, empty, "keyword")
	require.NotContains(t, empty, "offset")
}

func TestU3_IssueQuery_EncodesKeywords(t *testing.T) {
	// AC2.2.4: every keyword reaches Backlog unchanged.
	for _, kw := range []string{"ログイン", "đăng nhập", "a&b", "100%", "two words", "x=1#y"} {
		t.Run(kw, func(t *testing.T) {
			raw := IssueQuery{Keyword: kw, Count: 20}.Values().Encode()
			back, err := url.ParseQuery(raw)
			require.NoError(t, err)
			require.Equal(t, kw, back.Get("keyword"))
			require.Len(t, back, 4, "no extra parameter leaks out of the keyword")
		})
	}
}

func TestU3_ValidIssueRef(t *testing.T) {
	for ref, want := range map[string]bool{
		"PROJ-120": true, "P_2-1": true, "A-999999999": true, "5118": true,
		"": false, "proj-120": false, "PROJ-0": false, "PROJ-01": false, "PROJ-1234567890": false,
		"PROJ-1/../x": false, "../PROJ-1": false, "PROJ-1?x=1": false, "PROJ 1": false, "-1": false, "0": false,
		"1PROJ-1": false, "PROJ-1%2F": false,
	} {
		require.Equal(t, want, ValidIssueRef(ref), ref)
	}
}

func TestU3_Issue_FormatsHideUserContent(t *testing.T) {
	i := Issue{ID: 1, IssueKey: "PROJ-1", Summary: "s", Description: "secret-description-body"}
	c := Comment{ID: 2, Content: "secret-comment-body", AuthorName: "Lan"}
	for _, verb := range []string{"%v", "%+v", "%#v", "%s"} {
		require.NotContains(t, fmt.Sprintf(verb, i), "secret-description-body", verb)
		require.NotContains(t, fmt.Sprintf(verb, c), "secret-comment-body", verb)
	}
	require.Contains(t, fmt.Sprint(i), "PROJ-1")
}
