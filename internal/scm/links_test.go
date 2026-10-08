package scm

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// FR5.1: a pasted URL of a mapped repository links the task.
func TestLink_ByURLToAMappedRepository(t *testing.T) {
	h := newHarness(t)
	h.mapRepo(t, GitLab, "PROJ", "acme/web")
	h.clients[GitLab].setPRs("acme/web", pr(GitLab, "acme/web", 7, "Add login", "login", StateOpen))
	l, err := h.svc.Link(h.ctx, ws, "task-1", " https://gitlab.com/acme/web/-/merge_requests/7 ")
	require.NoError(t, err)
	require.Equal(t, Link{PRRef: PRRef{GitLab, "acme/web", 7}, TaskID: "task-1", ProjectKey: "PROJ", Title: "Add login",
		State: StateOpen, URL: "https://gitlab.com/acme/web/-/merge_requests/7"}, l)
	_, err = h.svc.Link(h.ctx, ws, "task-1", "https://gitlab.com/acme/web/-/merge_requests/7")
	require.NoError(t, err, "linking twice is a no-op")
	links, err := h.svc.Links(h.ctx, ws, "task-1", "")
	require.NoError(t, err)
	require.Len(t, links, 1)

	_, err = h.svc.Link(h.ctx, ws, "task-1", "https://gitlab.com/acme/api/-/merge_requests/7")
	require.Equal(t, FieldURL, fieldOf(t, err), "acme/api is not mapped")
	_, err = h.svc.Link(h.ctx, ws, "task-1", "https://evil.example/acme/web/-/merge_requests/7")
	require.Equal(t, FieldURL, fieldOf(t, err))
	_, err = h.svc.Link(h.ctx, ws, "", "https://gitlab.com/acme/web/-/merge_requests/7")
	require.Equal(t, FieldURL, fieldOf(t, err), "a task is required")
	_, err = h.svc.Link(h.ctx, ws, "task-1", "https://gitlab.com/acme/web/-/merge_requests/8")
	require.ErrorIs(t, err, ErrNotFound)
}

// FR5.2, A4: a key of a selected, mapped project in the branch or title
// links the issue automatically when a list refresh first sees the PR.
func TestAutoLink_OnListRefresh(t *testing.T) {
	h := newHarness(t)
	h.mapRepo(t, GitHub, "PROJ", "acme/web")
	h.mapRepo(t, GitHub, "DEMO", "acme/api")
	h.clients[GitHub].setPRs("acme/web",
		pr(GitHub, "acme/web", 3, "PROJ-12 login", "feature/PROJ-12-login", StateOpen), // one key, twice
		pr(GitHub, "acme/web", 2, "DEMO-3 is not mapped to acme/web", "x", StateOpen),
		pr(GitHub, "acme/web", 1, "Fix OTHER-1", "y", StateOpen))
	_, err := h.svc.ListPRs(h.ctx, ws, PRListInput{Provider: GitHub, ProjectKey: "PROJ", Repo: "acme/web", Statuses: []string{"open"}})
	require.NoError(t, err)
	links, err := h.svc.Links(h.ctx, ws, "", "PROJ-12")
	require.NoError(t, err)
	require.Equal(t, []Link{{PRRef: PRRef{GitHub, "acme/web", 3}, IssueKey: "PROJ-12", Auto: true, ProjectKey: "PROJ",
		Title: "PROJ-12 login", State: StateOpen, URL: "https://github.com/acme/web/pull/3"}}, links)
	all, err := h.store().Links(h.ctx, ws)
	require.NoError(t, err)
	require.Len(t, all, 1, "DEMO is selected but not mapped to acme/web; OTHER is not selected")

	_, err = h.svc.ListPRs(h.ctx, ws, PRListInput{Provider: GitHub, ProjectKey: "PROJ", Repo: "acme/web", Statuses: []string{"open"}})
	require.NoError(t, err)
	all, _ = h.store().Links(h.ctx, ws)
	require.Len(t, all, 1, "never created twice")
}

// FR5.3: a removed auto-link is never created again.
func TestUnlink_AnAutoLinkIsDismissedForGood(t *testing.T) {
	h := newHarness(t)
	h.mapRepo(t, GitHub, "PROJ", "acme/web")
	h.clients[GitHub].setPRs("acme/web", pr(GitHub, "acme/web", 3, "PROJ-12 login", "b", StateOpen))
	in := PRListInput{Provider: GitHub, ProjectKey: "PROJ", Repo: "acme/web", Statuses: []string{"open"}}
	_, err := h.svc.ListPRs(h.ctx, ws, in)
	require.NoError(t, err)
	key := PRRef{GitHub, "acme/web", 3}.Key()

	require.ErrorIs(t, h.svc.Unlink(h.ctx, ws, "task-9", key, ""), ErrNotFound, "the task has no such link")
	require.NoError(t, h.svc.Unlink(h.ctx, ws, "task-9", key, "PROJ-12"))
	dismissed, err := h.store().Dismissed(h.ctx, ws)
	require.NoError(t, err)
	require.Equal(t, []string{key + "|PROJ-12"}, dismissed)

	_, err = h.svc.ListPRs(h.ctx, ws, in)
	require.NoError(t, err)
	links, err := h.svc.Links(h.ctx, ws, "", "PROJ-12")
	require.NoError(t, err)
	require.Empty(t, links)
}

// Removing the auto-link of a PR keeps the task's manual link to it, and back.
func TestUnlink_TargetsOneKindOfLink(t *testing.T) {
	h := newHarness(t)
	h.mapRepo(t, GitHub, "PROJ", "acme/web")
	h.clients[GitHub].setPRs("acme/web", pr(GitHub, "acme/web", 3, "PROJ-12 login", "b", StateOpen))
	_, err := h.svc.ListPRs(h.ctx, ws, PRListInput{Provider: GitHub, ProjectKey: "PROJ", Repo: "acme/web", Statuses: []string{"open"}})
	require.NoError(t, err)
	l, err := h.svc.Link(h.ctx, ws, "task-1", "https://github.com/acme/web/pull/3")
	require.NoError(t, err)
	require.NoError(t, h.svc.Unlink(h.ctx, ws, "task-1", l.Key(), "PROJ-12"))
	links, _ := h.svc.Links(h.ctx, ws, "task-1", "PROJ-12")
	require.Len(t, links, 1)
	require.False(t, links[0].Auto)
	require.NoError(t, h.svc.Unlink(h.ctx, ws, "task-1", l.Key(), ""))
	links, _ = h.svc.Links(h.ctx, ws, "task-1", "PROJ-12")
	require.Empty(t, links)
}

func TestUnlink_AManualLinkIsRemovedWithoutDismissal(t *testing.T) {
	h := newHarness(t)
	h.mapRepo(t, Bitbucket, "PROJ", "acme/web")
	h.clients[Bitbucket].setPRs("acme/web", pr(Bitbucket, "acme/web", 4, "x", "b", StateOpen))
	l, err := h.svc.Link(h.ctx, ws, "task-1", "https://bitbucket.org/acme/web/pull-requests/4")
	require.NoError(t, err)
	require.NoError(t, h.svc.Unlink(h.ctx, ws, "task-1", l.Key(), ""))
	links, _ := h.svc.Links(h.ctx, ws, "task-1", "")
	require.Empty(t, links)
	dismissed, _ := h.store().Dismissed(h.ctx, ws)
	require.Empty(t, dismissed)
}

// The issue panel shows the task's manual links and its issue's auto-links.
func TestLinks_FiltersByTaskOrIssue(t *testing.T) {
	h := newHarness(t)
	h.use(t, GitHub)
	require.NoError(t, h.store().UpdateLinks(h.ctx, ws, func([]Link) ([]Link, error) {
		return []Link{
			{PRRef: PRRef{GitHub, "a/b", 1}, TaskID: "task-1"},
			{PRRef: PRRef{GitHub, "a/b", 2}, IssueKey: "PROJ-1", Auto: true},
			{PRRef: PRRef{GitHub, "a/b", 3}, TaskID: "task-2"},
		}, nil
	}))
	links, err := h.svc.Links(h.ctx, ws, "task-1", "PROJ-1")
	require.NoError(t, err)
	require.Len(t, links, 2)
	none, err := h.svc.Links(h.ctx, ws, "", "")
	require.NoError(t, err)
	require.Equal(t, []Link{}, none, "no filter shows nothing")
}

func (h *harness) store() *Store { return h.svc.store }
