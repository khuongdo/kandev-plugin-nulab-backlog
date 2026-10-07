package scm

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/connection"
)

func fieldOf(t *testing.T, err error) string {
	t.Helper()
	var fe *connection.FieldError
	require.True(t, errors.As(err, &fe), "want a FieldError, got %v", err)
	return fe.Field
}

// FR1.1: exactly the three cloud providers; Backlog Git stays in internal/git.
func TestParseProvider_AcceptsOnlyTheThreeCloudProviders(t *testing.T) {
	require.Equal(t, []Provider{GitHub, GitLab, Bitbucket}, Providers)
	for _, p := range []string{"github", "gitlab", "bitbucket"} {
		got, err := ParseProvider(p)
		require.NoError(t, err)
		require.Equal(t, Provider(p), got)
	}
	for _, p := range []string{"", "GitHub", "nulab-backlog", "backlog", "azure_devops", "gitea"} {
		_, err := ParseProvider(p)
		require.Equal(t, FieldProvider, fieldOf(t, err), p)
	}
}

// FR3.2: the repository name each provider uses.
func TestValidRepo_PerProviderShapes(t *testing.T) {
	cases := []struct {
		p    Provider
		name string
		ok   bool
	}{
		{GitHub, "acme/web-app", true},
		{GitHub, "acme/web.app_2", true},
		{GitHub, "acme", false},
		{GitHub, "acme/web/app", false},
		{GitHub, "../etc/passwd", false},
		{GitHub, "acme/..", false},
		{GitLab, "group/project", true},
		{GitLab, "group/sub/deeper/project", true},
		{GitLab, "project", false},
		{GitLab, "group//project", false},
		{GitLab, "group/-/project", false},
		{Bitbucket, "workspace/repo-1", true},
		{Bitbucket, "workspace/repo/extra", false},
		{Bitbucket, "work space/repo", false},
		{Bitbucket, "workspace/repo?x=1", false},
	}
	for _, c := range cases {
		require.Equal(t, c.ok, ValidRepo(c.p, c.name), "%s %s", c.p, c.name)
	}
}

// FR1.2, FR5.1, NFR2: only the cloud web URL of a pull request is accepted.
func TestParsePRURL_AcceptsOnlyCloudPullRequestPages(t *testing.T) {
	ok := []struct {
		raw    string
		p      Provider
		repo   string
		number int
	}{
		{"https://github.com/acme/web/pull/42", GitHub, "acme/web", 42},
		{" https://github.com/acme/web/pull/42/files ", GitHub, "acme/web", 42},
		{"https://gitlab.com/grp/sub/proj/-/merge_requests/7", GitLab, "grp/sub/proj", 7},
		{"https://bitbucket.org/ws/repo/pull-requests/3", Bitbucket, "ws/repo", 3},
		{"https://bitbucket.org/ws/repo/pull-requests/3/diff", Bitbucket, "ws/repo", 3},
	}
	for _, c := range ok {
		ref, err := ParsePRURL(c.raw)
		require.NoError(t, err, c.raw)
		require.Equal(t, PRRef{Provider: c.p, Repo: c.repo, Number: c.number}, ref, c.raw)
	}
	bad := []string{
		"http://github.com/acme/web/pull/42",                // not https
		"https://evil.example/acme/web/pull/42",             // another host
		"https://github.com.evil.example/acme/web/pull/42",  // look-alike host
		"https://api.github.com/repos/acme/web/pulls/42",    // API, not the page
		"https://user:pw@github.com/acme/web/pull/42",       // userinfo
		"https://github.com/acme/web/pull/42?x=1",           // query
		"https://github.com/acme/web/pull/42#discussion",    // fragment
		"https://github.com:8443/acme/web/pull/42",          // port
		"https://github.com/acme/web/issues/42",             // not a PR
		"https://github.com/acme/web/pull/0",                // number
		"https://gitlab.com/proj/-/merge_requests/7",        // group missing
		"https://gitlab.example.com/g/p/-/merge_requests/7", // self-managed (FR1.2)
		"https://bitbucket.org/ws/repo/pull-requests/x",     // number
		"github.com/acme/web/pull/42",                       // no scheme
		"",
	}
	for _, raw := range bad {
		_, err := ParsePRURL(raw)
		require.Equal(t, FieldURL, fieldOf(t, err), raw)
	}
}

func TestPRURL_IsTheInverseOfParsePRURL(t *testing.T) {
	for _, ref := range []PRRef{{GitHub, "acme/web", 42}, {GitLab, "g/s/p", 7}, {Bitbucket, "ws/r", 3}} {
		got, err := ParsePRURL(PRURL(ref))
		require.NoError(t, err)
		require.Equal(t, ref, got)
	}
}

func TestLinkKey_IsProviderRepoNumber(t *testing.T) {
	require.Equal(t, "github|acme/web|42", PRRef{GitHub, "acme/web", 42}.Key())
}

// A3: draft is shown as draft and filtered as open; declined as closed.
func TestFilterState_MapsDraftAndDeclined(t *testing.T) {
	for state, want := range map[string]string{
		StateOpen: StateOpen, StateDraft: StateOpen, StateClosed: StateClosed,
		StateDeclined: StateClosed, StateMerged: StateMerged, "weird": StateOpen,
	} {
		require.Equal(t, want, FilterState(state), state)
	}
}

// FR4.2: the same limits as Backlog Git queries, plus a provider.
func TestQueryInput_Validate(t *testing.T) {
	valid := func() QueryInput {
		return QueryInput{Name: " Open ", Provider: GitHub, ProjectKey: "PROJ", Repo: "acme/web",
			Statuses: []string{"open", "open", "merged"}, Author: WhoMe}
	}
	q := valid()
	require.NoError(t, q.Validate())
	require.Equal(t, "Open", q.Name)
	require.Equal(t, []string{"open", "merged"}, q.Statuses, "de-duplicated")

	q = valid()
	q.Author = ""
	require.NoError(t, q.Validate())
	require.Equal(t, WhoAnyone, q.Author, "no author means anyone")

	long := make([]rune, 101)
	for i := range long {
		long[i] = 'é'
	}
	for name, c := range map[string]struct {
		mut   func(*QueryInput)
		field string
	}{
		"empty name":      {func(q *QueryInput) { q.Name = "  " }, FieldName},
		"101 runes":       {func(q *QueryInput) { q.Name = string(long) }, FieldName},
		"bad provider":    {func(q *QueryInput) { q.Provider = "backlog" }, FieldProvider},
		"bad project":     {func(q *QueryInput) { q.ProjectKey = "proj" }, FieldRepository},
		"bad repo":        {func(q *QueryInput) { q.Repo = "acme" }, FieldRepository},
		"no states":       {func(q *QueryInput) { q.Statuses = nil }, FieldStatuses},
		"unknown state":   {func(q *QueryInput) { q.Statuses = []string{"draft"} }, FieldStatuses},
		"author not who":  {func(q *QueryInput) { q.Author = "lan" }, FieldAuthor},
		"gitlab one part": {func(q *QueryInput) { q.Provider, q.Repo = GitLab, "proj" }, FieldRepository},
	} {
		t.Run(name, func(t *testing.T) {
			q := valid()
			c.mut(&q)
			require.Equal(t, c.field, fieldOf(t, q.Validate()))
		})
	}
}

// FR4.3: a watch has a workflow and an interval of 1-1440 minutes, 5 by default.
func TestWatchInput_Validate(t *testing.T) {
	w := WatchInput{QueryInput: QueryInput{Name: "Reviews", Provider: GitLab, ProjectKey: "PROJ", Repo: "g/p",
		Statuses: []string{"open"}}, WorkflowID: "wf-1"}
	require.NoError(t, w.Validate())
	require.Equal(t, 5, w.IntervalMinutes)
	for _, m := range []int{-1, 1441} {
		w.IntervalMinutes = m
		require.Equal(t, FieldInterval, fieldOf(t, w.Validate()), m)
	}
	w.IntervalMinutes = 60
	w.WorkflowID = " "
	require.Equal(t, FieldWorkflow, fieldOf(t, w.Validate()))
}

// FR5.2, A4: an issue key of a selected project in a branch or title.
func TestIssueKeys_FindsSelectedProjectKeys(t *testing.T) {
	got := IssueKeys([]string{"feature/PROJ-12-login", "Fix DEMO-3 and PROJ-12, OTHER-9"}, []string{"PROJ", "DEMO"})
	require.Equal(t, []string{"PROJ-12", "DEMO-3"}, got)
	require.Empty(t, IssueKeys([]string{"proj-12 PROJ-0 XPROJ-1x"}, []string{"PROJ"}))
}
