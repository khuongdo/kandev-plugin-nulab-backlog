package issues

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func seedQueries(t *testing.T, r *rig, qs ...IssueQuery) {
	t.Helper()
	require.NoError(t, r.store.UpdateQueries(context.Background(), "ws-1", func([]IssueQuery) ([]IssueQuery, error) { return qs, nil }))
}

func TestIssueQueriesStore_SaveListDeleteAndCap(t *testing.T) {
	s := NewStore(newMemState())
	ctx := context.Background()
	q := IssueQuery{ID: "q1", Name: "My bugs", ProjectKey: "PROJ", StatusIDs: []int64{1, 2}, Assignee: WhoMe, Keyword: "login"}
	require.NoError(t, s.UpdateQueries(ctx, "ws-1", func(l []IssueQuery) ([]IssueQuery, error) { return append(l, q), nil }))
	got, err := s.Queries(ctx, "ws-1")
	require.NoError(t, err)
	require.Equal(t, []IssueQuery{q}, got)
	require.NoError(t, s.UpdateQueries(ctx, "ws-1", func([]IssueQuery) ([]IssueQuery, error) { return nil, nil }))
	got, err = s.Queries(ctx, "ws-1")
	require.NoError(t, err)
	require.Empty(t, got)

	require.NoError(t, s.UpdateQueries(ctx, "ws-1", func([]IssueQuery) ([]IssueQuery, error) { return make([]IssueQuery, 50), nil }))
	err = s.UpdateQueries(ctx, "ws-1", func(l []IssueQuery) ([]IssueQuery, error) { return append(l, IssueQuery{}), nil })
	require.Equal(t, FieldLimit, fieldOf(t, err))
}

func TestIssueQueriesStore_MissingIsDefaultReadsFalse(t *testing.T) {
	state := newMemState()
	state.put("workspace", "ws-1", "issues.queries", map[string]any{"schemaVersion": float64(1),
		"items": []any{map[string]any{"id": "q1", "name": "Old", "statusIds": []any{}, "assignee": "", "keyword": ""}}})
	got, err := NewStore(state).Queries(context.Background(), "ws-1")
	require.NoError(t, err)
	require.Len(t, got, 1)
	require.False(t, got[0].IsDefault)
}

func TestIssueQueries_SetDefaultLeavesExactlyOne(t *testing.T) {
	r := newRig(t)
	seedQueries(t, r, IssueQuery{ID: "a", Name: "A", IsDefault: true}, IssueQuery{ID: "b", Name: "B"}, IssueQuery{ID: "c", Name: "C"})
	defaults := func(list []IssueQuery) (out []string) {
		for _, q := range list {
			if q.IsDefault {
				out = append(out, q.ID)
			}
		}
		return out
	}
	list, err := r.svc.SetQueryDefault(r.ctx, "ws-1", "b", true)
	require.NoError(t, err)
	require.Equal(t, []string{"b"}, defaults(list), "starring moves the star")
	list, err = r.svc.SetQueryDefault(r.ctx, "ws-1", "c", false)
	require.NoError(t, err)
	require.Equal(t, []string{"b"}, defaults(list), "un-starring another query keeps the star")
	list, err = r.svc.SetQueryDefault(r.ctx, "ws-1", "b", false)
	require.NoError(t, err)
	require.Empty(t, defaults(list), "un-starring the default leaves none")
	_, err = r.svc.SetQueryDefault(r.ctx, "ws-1", "missing", true)
	require.ErrorIs(t, err, ErrNotFound)
}

func TestIssueQuery_Validate(t *testing.T) {
	selected := []string{"PROJ", "DEMO"}
	ids := make([]int64, 51)
	for i := range ids {
		ids[i] = int64(i + 1)
	}
	cases := []struct {
		name  string
		q     IssueQuery
		field string
	}{
		{"blank name", IssueQuery{Name: " "}, FieldName},
		{"name over 100 runes", IssueQuery{Name: strings.Repeat("n", 101)}, FieldName},
		{"project not selected", IssueQuery{Name: "Q", ProjectKey: "OTHER"}, FieldProjectKey},
		{"more than 50 statuses", IssueQuery{Name: "Q", StatusIDs: ids}, FieldStatusIDs},
		{"zero status id", IssueQuery{Name: "Q", StatusIDs: []int64{0}}, FieldStatusIDs},
		{"unknown assignee", IssueQuery{Name: "Q", Assignee: "anyone-else"}, FieldAssignee},
		{"negative assignee id", IssueQuery{Name: "Q", Assignee: "-3"}, FieldAssignee},
		{"keyword over 100 runes", IssueQuery{Name: "Q", Keyword: strings.Repeat("k", 101)}, FieldKeyword},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			require.Equal(t, c.field, fieldOf(t, c.q.Validate(selected)))
		})
	}
	for _, a := range []string{"", WhoMe, "42"} {
		q := IssueQuery{Name: " Mine ", Assignee: a, ProjectKey: "DEMO", StatusIDs: []int64{1}, Keyword: " x "}
		require.NoError(t, q.Validate(selected), a)
		require.Equal(t, "Mine", q.Name)
		require.Equal(t, "x", q.Keyword)
	}
}

func TestIssueQueries_SaveListDeleteKeepTheStar(t *testing.T) {
	r := newRig(t)
	list, err := r.svc.ListQueries(r.ctx, "ws-1")
	require.NoError(t, err)
	require.NotNil(t, list)
	require.Empty(t, list)

	q, err := r.svc.SaveQuery(r.ctx, "ws-1", IssueQuery{Name: "My bugs", Assignee: WhoMe, StatusIDs: []int64{1}, IsDefault: true})
	require.NoError(t, err)
	require.NotEmpty(t, q.ID)
	require.False(t, q.IsDefault, "save never sets the star")
	_, err = r.svc.SetQueryDefault(r.ctx, "ws-1", q.ID, true)
	require.NoError(t, err)
	edited, err := r.svc.SaveQuery(r.ctx, "ws-1", IssueQuery{ID: q.ID, Name: "Renamed", Assignee: WhoMe})
	require.NoError(t, err)
	require.True(t, edited.IsDefault, "an edit keeps the star")
	list, err = r.svc.ListQueries(r.ctx, "ws-1")
	require.NoError(t, err)
	require.Equal(t, []IssueQuery{edited}, list)

	_, err = r.svc.SaveQuery(r.ctx, "ws-1", IssueQuery{ID: "missing", Name: "X"})
	require.ErrorIs(t, err, ErrNotFound)
	_, err = r.svc.SaveQuery(r.ctx, "ws-1", IssueQuery{Name: ""})
	require.Equal(t, FieldName, fieldOf(t, err))

	require.NoError(t, r.svc.DeleteQuery(r.ctx, "ws-1", q.ID))
	require.ErrorIs(t, r.svc.DeleteQuery(r.ctx, "ws-1", q.ID), ErrNotFound)
	list, err = r.svc.ListQueries(r.ctx, "ws-1")
	require.NoError(t, err)
	require.Empty(t, list, "deleting the default removes the default")
}
