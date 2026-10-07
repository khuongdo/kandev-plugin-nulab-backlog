package issues

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/connection"
)

func TestQuickActions_DefaultsMatchKandevWithBacklogWording(t *testing.T) {
	d := DefaultQuickActions()
	ids := func(list []QuickAction) (out []string) {
		for _, a := range list {
			out = append(out, a.ID+"/"+a.Label+"/"+a.Icon)
		}
		return out
	}
	require.Equal(t, []string{"implement/Implement/code", "investigate/Investigate/search", "reproduce/Reproduce/bug"}, ids(d.Issue))
	require.Equal(t, []string{"review/Review/eye", "address_feedback/Address feedback/message", "fix_ci/Fix CI/tool"}, ids(d.PR))
	for _, a := range append(d.Issue, d.PR...) {
		require.Contains(t, a.PromptTemplate, "{{url}}")
		require.Contains(t, a.PromptTemplate, "Backlog")
		require.NotContains(t, a.PromptTemplate, "GitHub")
		require.NotEmpty(t, a.Hint)
	}
}

func TestQuickActionsStore_RoundTripsAndFallsBackPerKind(t *testing.T) {
	state := newMemState()
	s := NewStore(state)
	ctx := context.Background()

	got, err := s.QuickActions(ctx, "ws-1")
	require.NoError(t, err)
	require.Equal(t, DefaultQuickActions(), got.Resolved(), "nothing stored: the defaults")

	mine := QuickAction{ID: "a1", Label: "Triage", Hint: "Sort it", Icon: "check", PromptTemplate: "Triage {{url}}"}
	require.NoError(t, s.UpdateQuickActions(ctx, "ws-1", func(q QuickActions) (QuickActions, error) {
		q.Issue = []QuickAction{mine}
		return q, nil
	}))
	got, err = s.QuickActions(ctx, "ws-1")
	require.NoError(t, err)
	require.Equal(t, []QuickAction{mine}, got.Issue)
	require.EqualValues(t, 1, state.raw("workspace", "ws-1", "issues.quick_actions")["schemaVersion"])
	resolved := got.Resolved()
	require.Equal(t, []QuickAction{mine}, resolved.Issue)
	require.Equal(t, DefaultQuickActions().PR, resolved.PR, "an empty PR list falls back to the PR defaults")
}

func TestQuickActionsStore_BrokenDocumentIsAnError(t *testing.T) {
	state := newMemState()
	state.put("workspace", "ws-1", "issues.quick_actions", map[string]any{"schemaVersion": float64(9)})
	_, err := NewStore(state).QuickActions(context.Background(), "ws-1")
	require.Error(t, err)
}

func TestQuickActionsInput_Validate(t *testing.T) {
	ok := func() QuickAction { return QuickAction{Label: "Do it", Icon: "sparkle", PromptTemplate: "Go {{url}}"} }
	many := make([]QuickAction, 21)
	for i := range many {
		many[i] = ok()
	}
	cases := []struct {
		name  string
		in    QuickActionsInput
		field string
	}{
		{"unknown kind", QuickActionsInput{Kind: "task", Actions: []QuickAction{ok()}}, FieldKind},
		{"more than 20 actions", QuickActionsInput{Kind: KindPR, Actions: many}, FieldActions},
		{"blank label", QuickActionsInput{Kind: KindIssue, Actions: []QuickAction{func() QuickAction { a := ok(); a.Label = "  "; return a }()}}, FieldLabel},
		{"label over 100 runes", QuickActionsInput{Kind: KindIssue, Actions: []QuickAction{func() QuickAction { a := ok(); a.Label = strings.Repeat("ä", 101); return a }()}}, FieldLabel},
		{"hint over 100 runes", QuickActionsInput{Kind: KindIssue, Actions: []QuickAction{func() QuickAction { a := ok(); a.Hint = strings.Repeat("h", 101); return a }()}}, FieldHint},
		{"prompt over 4000", QuickActionsInput{Kind: KindIssue, Actions: []QuickAction{func() QuickAction { a := ok(); a.PromptTemplate = strings.Repeat("p", 4001); return a }()}}, FieldPrompt},
		{"unknown icon", QuickActionsInput{Kind: KindIssue, Actions: []QuickAction{func() QuickAction { a := ok(); a.Icon = "rocket"; return a }()}}, FieldIcon},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			require.Equal(t, c.field, fieldOf(t, c.in.Validate()))
		})
	}
	t.Run("valid input is trimmed and new actions get an id", func(t *testing.T) {
		a := ok()
		a.Label, a.PromptTemplate = " Do it ", strings.Repeat("p", 4000)
		in := QuickActionsInput{Kind: KindIssue, Actions: []QuickAction{a, {ID: "keep", Label: "X", Icon: "eye"}}}
		require.NoError(t, in.Validate())
		require.Equal(t, "Do it", in.Actions[0].Label)
		require.NotEmpty(t, in.Actions[0].ID)
		require.Equal(t, "keep", in.Actions[1].ID)
	})
}

func TestQuickActions_SaveReplacesOneKindAndEmptyResets(t *testing.T) {
	r := newRig(t)
	got, err := r.svc.QuickActions(r.ctx, "ws-1")
	require.NoError(t, err)
	require.Equal(t, DefaultQuickActions(), got)

	edited := DefaultQuickActions().Issue
	edited[0].PromptTemplate = "Build {{title}} now"
	got, err = r.svc.SaveQuickActions(r.ctx, "ws-1", QuickActionsInput{Kind: KindIssue, Actions: edited})
	require.NoError(t, err)
	require.Equal(t, "Build {{title}} now", got.Issue[0].PromptTemplate)
	require.Equal(t, DefaultQuickActions().PR, got.PR)

	got, err = r.svc.SaveQuickActions(r.ctx, "ws-1", QuickActionsInput{Kind: KindPR, Actions: []QuickAction{}})
	require.NoError(t, err)
	require.Equal(t, DefaultQuickActions().PR, got.PR, "deleting every PR action falls back to the defaults")
	require.Equal(t, "Build {{title}} now", got.Issue[0].PromptTemplate, "the other kind is kept")

	_, err = r.svc.SaveQuickActions(r.ctx, "ws-1", QuickActionsInput{Kind: KindIssue, Actions: []QuickAction{{Label: "", Icon: "eye"}}})
	require.Equal(t, FieldLabel, fieldOf(t, err))
	stored, err := r.store.QuickActions(r.ctx, "ws-1")
	require.NoError(t, err)
	require.Equal(t, "Build {{title}} now", stored.Issue[0].PromptTemplate, "invalid input saves nothing")

	r.state.failSet["issues.quick_actions"] = true
	_, err = r.svc.SaveQuickActions(r.ctx, "ws-1", QuickActionsInput{Kind: KindPR})
	require.ErrorIs(t, err, connection.ErrStore)
}
