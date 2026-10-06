package git

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/backlog"
	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/connection"
)

const host = "example-space.backlog.com"

func fieldOf(t *testing.T, err error) string {
	t.Helper()
	var fe *connection.FieldError
	require.True(t, errors.As(err, &fe), "want a FieldError, got %v", err)
	return fe.Field
}

func TestU4_Reference_ParsesEachForm(t *testing.T) {
	def := &RepoRef{ProjectKey: "PROJ", Name: "web-app"}
	cases := []struct {
		name string
		raw  string
		def  *RepoRef
		want Reference
	}{
		{"a PR URL", "https://example-space.backlog.com/git/PROJ/api/pullRequests/42", nil, Reference{"PROJ", "api", 42}},
		{"a PR URL with spaces", "  https://EXAMPLE-SPACE.backlog.com/git/PROJ/api/pullRequests/7  ", nil, Reference{"PROJ", "api", 7}},
		{"repo#n", "api#12", def, Reference{"PROJ", "api", 12}},
		{"a bare number", "42", def, Reference{"PROJ", "web-app", 42}},
		{"a hash number", "#42", def, Reference{"PROJ", "web-app", 42}},
		{"nine digits", "999999999", def, Reference{"PROJ", "web-app", 999999999}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseReference(tc.raw, host, tc.def)
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}

func TestU4_Reference_RejectsBadInput(t *testing.T) {
	def := &RepoRef{ProjectKey: "PROJ", Name: "web-app"}
	for name, tc := range map[string]struct {
		raw string
		def *RepoRef
	}{
		"another host":              {"https://other.backlog.com/git/PROJ/api/pullRequests/42", def},
		"plain http":                {"http://example-space.backlog.com/git/PROJ/api/pullRequests/42", def},
		"a non-PR URL":              {"https://example-space.backlog.com/view/PROJ-1", def},
		"zero":                      {"0", def},
		"a negative":                {"-3", def},
		"letters":                   {"abc", def},
		"ten digits":                {"1234567890", def},
		"a number with no repo":     {"42", nil},
		"repo#n with no project":    {"api#4", nil},
		"empty":                     {"   ", def},
		"a bad repo name":           {"a/b#4", def},
		"a URL with a query":        {"https://example-space.backlog.com/git/PROJ/api/pullRequests/42?x=1", def},
		"a URL with a bad project":  {"https://example-space.backlog.com/git/proj/api/pullRequests/42", def},
		"a URL with a zero number":  {"https://example-space.backlog.com/git/PROJ/api/pullRequests/0", def},
		"a URL with extra segments": {"https://example-space.backlog.com/git/PROJ/api/pullRequests/4/files", def},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := ParseReference(tc.raw, host, tc.def)
			require.Equal(t, "reference", fieldOf(t, err))
		})
	}
}

func TestU4_ClonePath_AcceptsBothForms(t *testing.T) {
	for raw, want := range map[string][2]string{
		"/git/PROJ/web-app.git": {"PROJ", "web-app"},
		"/git/PROJ/web-app":     {"PROJ", "web-app"},
		"/git/DEMO_2/api.git":   {"DEMO_2", "api"},
	} {
		p, r, ok := ParseClonePath(raw)
		require.True(t, ok, raw)
		require.Equal(t, want, [2]string{p, r})
	}
	for _, raw := range []string{"/git/PROJ", "/PROJ/web-app.git", "/git/proj/web-app.git", "/git/PROJ/web-app.git/info", "/git/PROJ/.git", ""} {
		_, _, ok := ParseClonePath(raw)
		require.False(t, ok, raw)
	}
}

func TestU4_LinkKey_JoinsHostRepoAndNumber(t *testing.T) {
	require.Equal(t, "example-space.backlog.com|11|42", LinkKey(host, 11, 42))
}

func TestU4_Watch_Validate(t *testing.T) {
	valid := func() WatchInput {
		return WatchInput{Name: " Reviews ", ProjectKey: "PROJ", RepoName: "web-app",
			Statuses: []string{"open"}, Assignee: "me", Creator: "anyone", IssueKey: "PROJ-120"}
	}
	in := valid()
	require.NoError(t, in.Validate([]string{"PROJ"}))
	require.Equal(t, "Reviews", in.Name, "the name is trimmed")

	cases := map[string]struct {
		mut   func(*WatchInput)
		field string
	}{
		"an empty name":            {func(w *WatchInput) { w.Name = "  " }, "name"},
		"a long name":              {func(w *WatchInput) { w.Name = strings.Repeat("a", 101) }, "name"},
		"an unselected project":    {func(w *WatchInput) { w.ProjectKey = "DEMO" }, "repository"},
		"no repository":            {func(w *WatchInput) { w.RepoName = "" }, "repository"},
		"a bad repository name":    {func(w *WatchInput) { w.RepoName = "a b" }, "repository"},
		"no status":                {func(w *WatchInput) { w.Statuses = nil }, "statuses"},
		"an unknown status":        {func(w *WatchInput) { w.Statuses = []string{"draft"} }, "statuses"},
		"an unknown assignee":      {func(w *WatchInput) { w.Assignee = "someone" }, "assignee"},
		"an unknown creator":       {func(w *WatchInput) { w.Creator = "" }, "creator"},
		"a bad issue key":          {func(w *WatchInput) { w.IssueKey = "proj-1" }, "issueKey"},
		"an unselected issue":      {func(w *WatchInput) { w.IssueKey = "DEMO-1" }, "issueKey"},
		"a missing workflow":       {func(w *WatchInput) { w.WorkflowID = "x"; w.WorkflowID = "" }, ""},
		"a name of exactly 100":    {func(w *WatchInput) { w.Name = strings.Repeat("a", 100) }, ""},
		"an empty linked issue ok": {func(w *WatchInput) { w.IssueKey = "" }, ""},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			w := valid()
			tc.mut(&w)
			err := w.Validate([]string{"PROJ"})
			if tc.field == "" {
				require.NoError(t, err)
				return
			}
			require.Equal(t, tc.field, fieldOf(t, err))
		})
	}
}

func TestU4_Query_Validate(t *testing.T) {
	ok := QueryInput{Name: "Open PRs", ProjectKey: "PROJ", RepoName: "api", Statuses: []string{"open", "merged"}, Assignee: "anyone"}
	require.NoError(t, ok.Validate([]string{"PROJ"}))
	for field, mut := range map[string]func(*QueryInput){
		"name":       func(q *QueryInput) { q.Name = "" },
		"repository": func(q *QueryInput) { q.ProjectKey = "DEMO" },
		"statuses":   func(q *QueryInput) { q.Statuses = []string{"x"} },
		"assignee":   func(q *QueryInput) { q.Assignee = "x" },
	} {
		q := ok
		mut(&q)
		require.Equal(t, field, fieldOf(t, q.Validate([]string{"PROJ"})), field)
	}
}

func TestU4_RelatedIssue_FirstSelectedKey(t *testing.T) {
	selected := []string{"PROJ", "API_2"}
	cases := []struct{ title, body, want string }{
		{"PROJ-120 Add login", "", "PROJ-120"},
		{"Add login", "Fixes DEMO-3 and API_2-7", "API_2-7"},
		{"DEMO-1 only", "nothing", ""},
		{"lowercase proj-1", "", ""},
		{"XPROJ-1 is not PROJ", "", ""},
		{"Both PROJ-2 and API_2-9", "", "PROJ-2"},
	}
	for _, tc := range cases {
		require.Equal(t, tc.want, RelatedIssueKey(tc.title, tc.body, selected), tc.title)
	}
}

func TestU4_PRDescription_NeverClosesTheIssue(t *testing.T) {
	require.Equal(t, "My body", PRDescription("My body", "Title", "PROJ-120"))
	require.Equal(t, "Related: PROJ-120", PRDescription("  ", "Title", "PROJ-120"))
	require.Equal(t, "Title", PRDescription("", "Title", ""))
	for _, body := range []string{"", "x"} {
		d := strings.ToLower(PRDescription(body, "Fix PROJ-120", "PROJ-120"))
		for _, verb := range []string{"close", "fix", "resolve"} {
			require.NotContains(t, d, verb+" proj-120")
			require.NotContains(t, d, verb+"s proj-120")
		}
	}
}

func TestU4_Branches_MostUsedFirst(t *testing.T) {
	prs := []backlog.PullRequest{
		{Base: "main", Branch: "feature/a"},
		{Base: "main", Branch: "feature/b"},
		{Base: "develop", Branch: "feature/a"},
		{Base: "main", Branch: ""},
	}
	require.Equal(t, []string{"main", "feature/a", "develop", "feature/b"}, BranchCandidates(prs))
	require.Empty(t, BranchCandidates(nil))
	require.Equal(t, "main", DefaultBranch(prs))
	require.Equal(t, "master", DefaultBranch(nil), "Backlog's own default when no PR names a base")
}
