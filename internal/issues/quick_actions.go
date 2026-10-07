package issues

import (
	"context"
	"errors"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/connection"
)

const keyQuickActions = "issues.quick_actions"

// Quick action kinds: the row type a quick action starts a task from.
const (
	KindIssue = "issue"
	KindPR    = "pr"
)

// Quick action input fields reported with the validation code (FR2.4).
const (
	FieldKind    = "kind"
	FieldActions = "actions"
	FieldLabel   = "label"
	FieldHint    = "hint"
	FieldIcon    = "icon"
	FieldPrompt  = "promptTemplate"
)

const (
	maxQuickActions = 20
	maxActionLabel  = 100
	maxActionHint   = 100
	maxPrompt       = 4000
)

// QuickActionIcons are the icon names Kandev's start-task menu can draw.
var QuickActionIcons = []string{"eye", "message", "tool", "code", "search", "bug", "sparkle", "check"}

// QuickAction is one "+ Task" menu entry: a prompt template with {{url}} and
// {{title}} placeholders, filled in by the UI (FR1.2).
type QuickAction struct {
	ID             string `json:"id"`
	Label          string `json:"label"`
	Hint           string `json:"hint"`
	Icon           string `json:"icon"`
	PromptTemplate string `json:"promptTemplate"`
}

// QuickActions are a workspace's quick actions per kind. An empty list means
// the defaults of that kind (FR2.2).
type QuickActions struct {
	Issue []QuickAction `json:"issue"`
	PR    []QuickAction `json:"pr"`
}

// DefaultQuickActions are Kandev's GitHub defaults with Backlog wording
// (FR2.1). They are a persisted seed, so they are never translated.
func DefaultQuickActions() QuickActions {
	return QuickActions{
		Issue: []QuickAction{
			{ID: "implement", Label: "Implement", Hint: "Build and open a PR", Icon: "code",
				PromptTemplate: `Implement the changes described in the Backlog issue at {{url}} (title: "{{title}}"). Open a pull request when complete.`},
			{ID: "investigate", Label: "Investigate", Hint: "Find the root cause", Icon: "search",
				PromptTemplate: `Investigate the Backlog issue at {{url}} (title: "{{title}}"). Identify root cause and summarize findings.`},
			{ID: "reproduce", Label: "Reproduce", Hint: "Document repro steps", Icon: "bug",
				PromptTemplate: `Reproduce the bug described in the Backlog issue at {{url}} (title: "{{title}}"). Document the reproduction steps.`},
		},
		PR: []QuickAction{
			{ID: "review", Label: "Review", Hint: "Read the diff, flag issues", Icon: "eye",
				PromptTemplate: "Review the Backlog pull request at {{url}}. Provide feedback on code quality, correctness, and suggest improvements."},
			{ID: "address_feedback", Label: "Address feedback", Hint: "Apply review comments", Icon: "message",
				PromptTemplate: "Review the feedback on the Backlog pull request at {{url}}. Evaluate each comment critically - apply changes that improve the code, push back on suggestions that are unnecessary or harmful, and explain your reasoning. Push the changes when done."},
			{ID: "fix_ci", Label: "Fix CI", Hint: "Diagnose failing checks", Icon: "tool",
				PromptTemplate: "Investigate and fix the CI failures and merge conflicts on the Backlog pull request at {{url}}. Run the failing checks locally, resolve any conflicts, diagnose issues, and push fixes."},
		},
	}
}

// Resolved replaces an empty kind with its defaults (FR2.2).
func (q QuickActions) Resolved() QuickActions {
	d := DefaultQuickActions()
	if len(q.Issue) > 0 {
		d.Issue = q.Issue
	}
	if len(q.PR) > 0 {
		d.PR = q.PR
	}
	return d
}

// QuickActionsInput is the issues.quick_actions.save body: the full list of
// one kind. An empty list resets that kind to its defaults.
type QuickActionsInput struct {
	Kind    string        `json:"kind"`
	Actions []QuickAction `json:"actions"`
}

// Validate checks the kind and every action (FR2.4), trims the texts and
// gives a new action (no id) a fresh id.
func (in *QuickActionsInput) Validate() error {
	if in.Kind != KindIssue && in.Kind != KindPR {
		return invalid(FieldKind)
	}
	if len(in.Actions) > maxQuickActions {
		return &connection.FieldError{Field: FieldActions, Err: errLimit}
	}
	for i := range in.Actions {
		a := &in.Actions[i]
		a.Label, a.Hint = strings.TrimSpace(a.Label), strings.TrimSpace(a.Hint)
		switch n := utf8.RuneCountInString(a.Label); {
		case n == 0 || n > maxActionLabel:
			return invalid(FieldLabel)
		case utf8.RuneCountInString(a.Hint) > maxActionHint:
			return invalid(FieldHint)
		case utf8.RuneCountInString(a.PromptTemplate) > maxPrompt:
			return invalid(FieldPrompt)
		case !slices.Contains(QuickActionIcons, a.Icon):
			return invalid(FieldIcon)
		}
		if a.ID = strings.TrimSpace(a.ID); a.ID == "" {
			a.ID = newID()
		}
	}
	return nil
}

// QuickActions returns the workspace's stored quick actions (empty kinds
// stay empty; see Resolved).
func (s *Store) QuickActions(ctx context.Context, ws string) (QuickActions, error) {
	var q QuickActions
	err := s.load(ctx, "workspace", ws, keyQuickActions, &q)
	return q, err
}

// UpdateQuickActions replaces the stored quick actions with fn's result.
func (s *Store) UpdateQuickActions(ctx context.Context, ws string, fn func(QuickActions) (QuickActions, error)) error {
	defer s.lock("quick_actions/" + ws)()
	q, err := s.QuickActions(ctx, ws)
	if err != nil {
		return err
	}
	next, err := fn(q)
	if errors.Is(err, errUnchanged) {
		return nil
	}
	if err != nil {
		return err
	}
	return s.put(ctx, "workspace", ws, keyQuickActions, next)
}

// QuickActions returns the workspace's quick actions with the defaults
// filled in (FR2.2).
func (s *Service) QuickActions(ctx context.Context, ws string) (QuickActions, error) {
	q, err := s.store.QuickActions(ctx, ws)
	if err != nil {
		return QuickActions{}, err
	}
	return q.Resolved(), nil
}

// SaveQuickActions replaces one kind's quick actions and returns all of
// them resolved (FR2.3). Invalid input saves nothing.
func (s *Service) SaveQuickActions(ctx context.Context, ws string, in QuickActionsInput) (QuickActions, error) {
	if err := in.Validate(); err != nil {
		return QuickActions{}, err
	}
	var out QuickActions
	err := s.store.UpdateQuickActions(ctx, ws, func(q QuickActions) (QuickActions, error) {
		if in.Kind == KindIssue {
			q.Issue = in.Actions
		} else {
			q.PR = in.Actions
		}
		out = q
		return q, nil
	})
	if err != nil {
		return QuickActions{}, err
	}
	return out.Resolved(), nil
}
