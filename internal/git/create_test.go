package git

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/backlog"
)

func createInput() CreateInput {
	return CreateInput{TaskID: "task-17", RepositoryID: "repo-k1", HeadBranch: "feature/search", Title: "Add search PROJ-120"}
}

func TestU4_CreatePR_CreatesLinksAndReturnsTheURL(t *testing.T) {
	r := newRig(t)
	res, err := r.svc.CreatePR(r.ctx, ws, createInput())
	require.NoError(t, err)
	require.Equal(t, CreateResult{URL: "https://" + host + "/git/PROJ/web-app/pullRequests/101", Number: 101, Linked: true}, res)
	require.Equal(t, []backlog.NewPullRequest{{Summary: "Add search PROJ-120", Description: "Related: PROJ-120",
		Base: "main", Branch: "feature/search", IssueID: 5120}}, r.gw.created)
	require.Equal(t, 1, r.gw.count("Issue"))
	links := r.links(t)
	require.Len(t, links, 1)
	require.Equal(t, "task-17", links[0].TaskID)
	require.Equal(t, 101, links[0].Number)
}

func TestU4_CreatePR_ValidationMakesNoCall(t *testing.T) {
	r := newRig(t)
	for field, mut := range map[string]func(*CreateInput){
		FieldTitle:      func(in *CreateInput) { in.Title = "  " },
		FieldBranch:     func(in *CreateInput) { in.HeadBranch = "" },
		FieldRepository: func(in *CreateInput) { in.RepositoryID = "repo-gh" },
	} {
		in := createInput()
		mut(&in)
		_, err := r.svc.CreatePR(r.ctx, ws, in)
		require.Equal(t, field, fieldOf(t, err), field)
	}
	in := createInput()
	in.RepositoryID = "repo-missing"
	_, err := r.svc.CreatePR(r.ctx, ws, in)
	require.Equal(t, FieldRepository, fieldOf(t, err))
	require.Zero(t, r.gw.total())
}

func TestU4_CreatePR_AnOpenPRForTheBranchIsAConflict(t *testing.T) {
	r := newRig(t)
	r.gw.setPRs("PROJ/web-app", []backlog.PullRequest{{RepositoryID: 11, Number: 7, Branch: "feature/search", StatusID: 1},
		{RepositoryID: 11, Number: 6, Branch: "feature/other", StatusID: 1}})
	_, err := r.svc.CreatePR(r.ctx, ws, createInput())
	var open *OpenPRExistsError
	require.ErrorAs(t, err, &open)
	require.Equal(t, 7, open.Number)
	require.Zero(t, r.gw.count("CreatePullRequest"), "nothing is created (AC5.3.4)")
	require.Equal(t, []int64{1}, r.gw.queries[0].StatusIDs, "only open PRs are checked")
}

func TestU4_CreatePR_BaseBranchAndBody(t *testing.T) {
	r := newRig(t)
	in := createInput()
	in.BaseBranch, in.Body, in.Title = "develop", "My description", "No key here"
	_, err := r.svc.CreatePR(r.ctx, ws, in)
	require.NoError(t, err)
	require.Equal(t, backlog.NewPullRequest{Summary: "No key here", Description: "My description", Base: "develop", Branch: "feature/search"}, r.gw.created[0])
	require.Zero(t, r.gw.count("Issue"), "no key: no issue lookup")
}

func TestU4_CreatePR_UnknownIssueIsSkipped(t *testing.T) {
	r := newRig(t)
	in := createInput()
	in.Title = "Add search PROJ-999"
	_, err := r.svc.CreatePR(r.ctx, ws, in)
	require.NoError(t, err)
	require.Zero(t, r.gw.created[0].IssueID)
	require.Equal(t, "Related: PROJ-999", r.gw.created[0].Description)
}

func TestU4_CreatePR_LinkFailureNeverCreatesAgain(t *testing.T) {
	r := newRig(t)
	r.state.failSet = true
	res, err := r.svc.CreatePR(r.ctx, ws, createInput())
	require.NoError(t, err)
	require.False(t, res.Linked)
	require.NotEmpty(t, res.URL)
	require.Equal(t, 1, r.gw.count("CreatePullRequest"))
}

func TestU4_CreatePR_BackendErrorsPassThrough(t *testing.T) {
	r := newRig(t)
	rl := backlogErr(backlog.KindRateLimited, 429)
	r.gw.setErr("CreatePullRequest", rl)
	_, err := r.svc.CreatePR(r.ctx, ws, createInput())
	require.ErrorIs(t, err, rl)
	require.Empty(t, r.links(t))
}
