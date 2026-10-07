package scm

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func query(name string, p Provider) QueryInput {
	return QueryInput{Name: name, Provider: p, ProjectKey: "PROJ", Repo: "acme/web", Statuses: []string{"open"}}
}

// FR4.2: save, list, rename, delete; one default per provider.
func TestQueries_CRUDAndOneDefaultPerProvider(t *testing.T) {
	h := newHarness(t)
	h.mapRepo(t, GitHub, "PROJ", "acme/web")
	h.mapRepo(t, GitLab, "PROJ", "acme/web")
	a, err := h.svc.SaveQuery(h.ctx, ws, query("A", GitHub))
	require.NoError(t, err)
	require.NotEmpty(t, a.ID)
	b, err := h.svc.SaveQuery(h.ctx, ws, query("B", GitHub))
	require.NoError(t, err)
	c, err := h.svc.SaveQuery(h.ctx, ws, query("C", GitLab))
	require.NoError(t, err)

	_, err = h.svc.SetQueryDefault(h.ctx, ws, c.ID, true)
	require.NoError(t, err)
	_, err = h.svc.SetQueryDefault(h.ctx, ws, a.ID, true)
	require.NoError(t, err)
	list, err := h.svc.SetQueryDefault(h.ctx, ws, b.ID, true)
	require.NoError(t, err)
	defaults := map[string]bool{}
	for _, q := range list {
		defaults[q.Name] = q.IsDefault
	}
	require.Equal(t, map[string]bool{"A": false, "B": true, "C": true}, defaults, "the star moves within GitHub only")

	renamed := query("B2", GitHub)
	renamed.ID = b.ID
	got, err := h.svc.SaveQuery(h.ctx, ws, renamed)
	require.NoError(t, err)
	require.True(t, got.IsDefault, "an edit keeps the star")

	require.NoError(t, h.svc.DeleteQuery(h.ctx, ws, a.ID))
	require.ErrorIs(t, h.svc.DeleteQuery(h.ctx, ws, a.ID), ErrNotFound)
	_, err = h.svc.SetQueryDefault(h.ctx, ws, "missing", true)
	require.ErrorIs(t, err, ErrNotFound)
	renamed.ID = "missing"
	_, err = h.svc.SaveQuery(h.ctx, ws, renamed)
	require.ErrorIs(t, err, ErrNotFound)
	qs, err := h.svc.ListQueries(h.ctx, ws)
	require.NoError(t, err)
	require.Len(t, qs, 2)
}

// FR3.3: a query covers a mapped repository only.
func TestSaveQuery_NeedsAMappedRepository(t *testing.T) {
	h := newHarness(t)
	h.mapRepo(t, GitHub, "PROJ", "acme/web")
	q := query("A", GitHub)
	q.Repo = "acme/api"
	_, err := h.svc.SaveQuery(h.ctx, ws, q)
	require.Equal(t, FieldRepository, fieldOf(t, err))
	q.Name = ""
	_, err = h.svc.SaveQuery(h.ctx, ws, q)
	require.Equal(t, FieldName, fieldOf(t, err))
}

// FR3.4: removing a mapping disables, never deletes, its queries and watches.
func TestUnmapping_DisablesQueriesAndWatches(t *testing.T) {
	h := newHarness(t)
	h.mapRepo(t, GitHub, "PROJ", "acme/web")
	q, err := h.svc.SaveQuery(h.ctx, ws, query("A", GitHub))
	require.NoError(t, err)
	w, err := h.svc.SaveWatch(h.ctx, ws, WatchInput{QueryInput: query("W", GitHub), WorkflowID: "wf-1"})
	require.NoError(t, err)

	_, err = h.svc.SetMapping(h.ctx, ws, MappingInput{Provider: GitHub, ProjectKey: "PROJ"})
	require.NoError(t, err)
	qs, err := h.svc.ListQueries(h.ctx, ws)
	require.NoError(t, err)
	require.True(t, qs[0].Unmapped)
	ws2, err := h.svc.ListWatches(h.ctx, ws)
	require.NoError(t, err)
	require.True(t, ws2[0].Unmapped)
	_, err = h.svc.RunQuery(h.ctx, ws, q.ID)
	require.ErrorIs(t, err, ErrConflict)
	_, err = h.svc.RunWatch(h.ctx, ws, w.ID)
	require.ErrorIs(t, err, ErrConflict)

	h.mapRepo(t, GitHub, "PROJ", "acme/web")
	qs, _ = h.svc.ListQueries(h.ctx, ws)
	require.False(t, qs[0].Unmapped, "mapping again enables them")
}

// FR4.2: a run shows the first page with the query's filters.
func TestRunQuery(t *testing.T) {
	h := newHarness(t)
	h.mapRepo(t, GitHub, "PROJ", "acme/web")
	h.clients[GitHub].setPRs("acme/web", pr(GitHub, "acme/web", 1, "x", "b", StateOpen))
	q, err := h.svc.SaveQuery(h.ctx, ws, query("A", GitHub))
	require.NoError(t, err)
	page, err := h.svc.RunQuery(h.ctx, ws, q.ID)
	require.NoError(t, err)
	require.Len(t, page.Items, 1)
	_, err = h.svc.RunQuery(h.ctx, ws, "missing")
	require.ErrorIs(t, err, ErrNotFound)
}
