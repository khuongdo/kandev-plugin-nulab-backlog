package backlog

import (
	"errors"
	"fmt"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestU4_PRParse_DecodesTheList(t *testing.T) {
	got, err := parsePullRequests(readFixture(t, "pullrequests_ok.json"))
	require.NoError(t, err)
	require.Len(t, got, 3)
	require.Equal(t, PullRequest{
		ID: 2001, RepositoryID: 11, Number: 42, Summary: "Add login page", Description: "Related: PROJ-120 {{BAIT}}",
		Base: "main", Branch: "feature/login", StatusID: 1, AssigneeName: "Lan", IssueID: 5120,
		Created: "2026-10-01T09:00:00Z",
	}, got[0])
	require.Empty(t, got[1].AssigneeName, "a null assignee is empty")
	require.Zero(t, got[1].IssueID, "a null issue is 0")
}

func TestU4_PRParse_RejectsUnusableBodies(t *testing.T) {
	for name, body := range map[string]string{
		"a missing number":  `[{"id":1,"repositoryId":11,"summary":"x","status":{"id":1}}]`,
		"a string number":   `[{"id":1,"repositoryId":11,"number":"4","status":{"id":1}}]`,
		"a missing repo id": `[{"id":1,"number":4,"status":{"id":1}}]`,
		"not a list":        `{"id":1}`,
	} {
		t.Run(name, func(t *testing.T) {
			_, err := parsePullRequests([]byte(body))
			var be *Error
			require.True(t, errors.As(err, &be))
			require.Equal(t, "body", be.Class)
		})
	}
	_, err := parsePullRequest([]byte(`{"id":1}`))
	require.Error(t, err)
}

func TestU4_PRState_MapsStatusIDs(t *testing.T) {
	for id, want := range map[int]string{1: "open", 2: "closed", 3: "merged", 4: "unknown", 0: "unknown"} {
		require.Equal(t, want, PullRequestState(id), id)
	}
}

func TestU4_PRQuery_EncodesFiltersAndClampsCount(t *testing.T) {
	q := PullRequestQuery{StatusIDs: []int64{1, 2}, AssigneeIDs: []int64{7}, IssueIDs: []int64{5120},
		CreatedUserIDs: []int64{9}, Offset: 20, Count: 500}
	require.Equal(t, url.Values{
		"statusId[]": {"1", "2"}, "assigneeId[]": {"7"}, "issueId[]": {"5120"}, "createdUserId[]": {"9"},
		"offset": {"20"}, "count": {"100"},
	}, q.Values())
	require.Equal(t, url.Values{"count": {"1"}}, PullRequestQuery{}.Values(), "count is at least 1")
}

func TestU4_PRForm_SendsTheRequiredFields(t *testing.T) {
	in := NewPullRequest{Summary: "Add search", Description: "Related: PROJ-120", Base: "main", Branch: "feature/search"}
	require.Equal(t, url.Values{"summary": {"Add search"}, "description": {"Related: PROJ-120"},
		"base": {"main"}, "branch": {"feature/search"}}, in.Form())
	in.IssueID = 5120
	require.Equal(t, []string{"5120"}, in.Form()["issueId"])
}

func TestU4_PRFormat_HidesTheDescription(t *testing.T) {
	pr := PullRequest{Number: 42, Summary: "Add login page", Description: "secret-body-text"}
	for _, s := range []string{fmt.Sprintf("%v", pr), fmt.Sprintf("%+v", pr), fmt.Sprintf("%#v", pr)} {
		require.NotContains(t, s, "secret-body-text")
		require.Contains(t, s, "42")
	}
}

func TestU4_IssueParse_DecodesTheIssue(t *testing.T) {
	got, err := parseIssue(readFixture(t, "issue_ok.json"))
	require.NoError(t, err)
	require.Equal(t, Issue{ID: 5120, IssueKey: "PROJ-120", Summary: "Login page"}, got)
	_, err = parseIssue([]byte(`{"issueKey":"PROJ-1"}`))
	require.Error(t, err)
}
