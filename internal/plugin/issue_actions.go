package plugin

import (
	"context"
	"maps"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/issues"
)

// U3 action keys declared in manifest.yaml (C5, R-04).
const (
	actionIssuesList      = "issues.list"
	actionIssuesFilters   = "issues.filters"
	actionIssuesCreate    = "issues.create_task"
	actionTasksSearch     = "issues.tasks.search"
	actionIssuesLinks     = "issues.links.list"
	actionIssuesRefresh   = "issues.refresh"
	actionIssuesImpact    = "issues.impact"
	actionIssuesSettings  = "issues.settings.get"
	actionIssuesLink      = "issues.link"
	actionIssuesUnlink    = "issues.unlink"
	actionIssuesGet       = "issues.get"
	actionIssuesComments  = "issues.comments"
	actionSetPollInterval = "issues.set_poll_interval"

	// Intent 261007: issue watches (FR3), for any signed-in member (FR1.4).
	actionIssueWatchesList   = "issues.watches.list"
	actionIssueWatchesSave   = "issues.watches.save"
	actionIssueWatchesDelete = "issues.watches.delete"
	actionIssueWatchesRun    = "issues.watches.run"
	actionIssueWatchesPause  = "issues.watches.pause"
	actionIssueWatchesResume = "issues.watches.resume"

	// Intent 261007-github-parity-actions: quick actions (FR2) and saved
	// issue queries (FR4), workspace-wide for any signed-in member (NFR1).
	actionQuickActionsGet     = "issues.quick_actions.get"
	actionQuickActionsSave    = "issues.quick_actions.save"
	actionIssueQueriesList    = "issues.queries.list"
	actionIssueQueriesSave    = "issues.queries.save"
	actionIssueQueriesDelete  = "issues.queries.delete"
	actionIssueQueriesDefault = "issues.queries.set_default"
)

func init() { maps.Copy(handlers, issueHandlers) }

// issueHandlers are the U3 actions. Task-scoped ones take the task from the
// verified action context, never from the body.
var issueHandlers = map[string]handler{
	actionIssuesList: func(r *Runtime, ctx context.Context, ws string, body []byte) (any, error) {
		var q issues.Query
		if err := decode(body, &q); err != nil {
			return nil, err
		}
		return r.issues.List(ctx, ws, q)
	},
	actionIssuesFilters: func(r *Runtime, ctx context.Context, ws string, _ []byte) (any, error) {
		return r.issues.Filters(ctx, ws)
	},
	actionIssuesCreate: func(r *Runtime, ctx context.Context, ws string, body []byte) (any, error) {
		var in issues.CreateInput
		if err := decode(body, &in); err != nil {
			return nil, err
		}
		return r.issues.CreateTask(ctx, ws, in)
	},
	actionTasksSearch: func(r *Runtime, ctx context.Context, ws string, body []byte) (any, error) {
		var in struct {
			Query string `json:"query"`
		}
		if err := decode(body, &in); err != nil {
			return nil, err
		}
		tasks, err := r.issues.SearchTasks(ctx, ws, in.Query)
		return map[string]any{"tasks": tasks}, err
	},
	actionIssuesLinks: func(r *Runtime, ctx context.Context, ws string, _ []byte) (any, error) {
		links, err := r.issues.Links(ctx, ws)
		return map[string]any{"links": links}, err
	},
	actionIssuesRefresh: func(r *Runtime, ctx context.Context, ws string, _ []byte) (any, error) {
		return r.syncer.Refresh(ctx, ws)
	},
	actionIssuesImpact: func(r *Runtime, ctx context.Context, ws string, body []byte) (any, error) {
		var in struct {
			ProjectKeys []string `json:"projectKeys"`
		}
		if err := decode(body, &in); err != nil {
			return nil, err
		}
		return r.issues.Impact(ctx, ws, in.ProjectKeys)
	},
	actionIssuesSettings: func(r *Runtime, ctx context.Context, ws string, _ []byte) (any, error) {
		return r.issues.Settings(ctx, ws)
	},
	actionIssuesLink: func(r *Runtime, ctx context.Context, ws string, body []byte) (any, error) {
		var in struct {
			IssueKey string `json:"issueKey"`
		}
		if err := decode(body, &in); err != nil {
			return nil, err
		}
		return r.issues.Link(ctx, ws, verified(ctx).TaskID, in.IssueKey)
	},
	actionIssuesUnlink: func(r *Runtime, ctx context.Context, ws string, _ []byte) (any, error) {
		return ok, r.issues.Unlink(ctx, ws, verified(ctx).TaskID)
	},
	actionIssuesGet: func(r *Runtime, ctx context.Context, ws string, _ []byte) (any, error) {
		return r.issues.Detail(ctx, ws, verified(ctx).TaskID)
	},
	actionIssuesComments: func(r *Runtime, ctx context.Context, ws string, body []byte) (any, error) {
		var in struct {
			MaxID int64 `json:"maxId"`
		}
		if err := decode(body, &in); err != nil {
			return nil, err
		}
		return r.issues.Comments(ctx, ws, verified(ctx).TaskID, in.MaxID)
	},
	actionSetPollInterval: func(r *Runtime, ctx context.Context, ws string, body []byte) (any, error) {
		var in struct {
			Minutes any `json:"minutes"`
		}
		if err := decode(body, &in); err != nil {
			return nil, err
		}
		return r.issues.SetPollInterval(ctx, ws, in.Minutes)
	},
	actionIssueWatchesList: func(r *Runtime, ctx context.Context, ws string, _ []byte) (any, error) {
		w, err := r.issues.ListWatches(ctx, ws)
		return map[string]any{"watches": w}, err
	},
	actionIssueWatchesSave: func(r *Runtime, ctx context.Context, ws string, body []byte) (any, error) {
		var in issues.IssueWatchInput
		if err := decode(body, &in); err != nil {
			return nil, err
		}
		return r.issues.SaveWatch(ctx, ws, in)
	},
	actionIssueWatchesDelete: withID(func(r *Runtime, ctx context.Context, ws, id string) (any, error) {
		return ok, r.issues.DeleteWatch(ctx, ws, id)
	}),
	actionIssueWatchesRun: withID(func(r *Runtime, ctx context.Context, ws, id string) (any, error) {
		return map[string]bool{"queued": true}, r.issueWatcher.Run(ctx, ws, id)
	}),
	actionIssueWatchesPause: withID(func(r *Runtime, ctx context.Context, ws, id string) (any, error) {
		return r.issues.PauseWatch(ctx, ws, id)
	}),
	actionIssueWatchesResume: withID(func(r *Runtime, ctx context.Context, ws, id string) (any, error) {
		return r.issues.ResumeWatch(ctx, ws, id)
	}),
	actionQuickActionsGet: func(r *Runtime, ctx context.Context, ws string, _ []byte) (any, error) {
		return r.issues.QuickActions(ctx, ws)
	},
	actionQuickActionsSave: func(r *Runtime, ctx context.Context, ws string, body []byte) (any, error) {
		var in issues.QuickActionsInput
		if err := decode(body, &in); err != nil {
			return nil, err
		}
		return r.issues.SaveQuickActions(ctx, ws, in)
	},
	actionIssueQueriesList: func(r *Runtime, ctx context.Context, ws string, _ []byte) (any, error) {
		q, err := r.issues.ListQueries(ctx, ws)
		return map[string]any{"queries": q}, err
	},
	actionIssueQueriesSave: func(r *Runtime, ctx context.Context, ws string, body []byte) (any, error) {
		var in issues.IssueQuery
		if err := decode(body, &in); err != nil {
			return nil, err
		}
		return r.issues.SaveQuery(ctx, ws, in)
	},
	actionIssueQueriesDelete: withID(func(r *Runtime, ctx context.Context, ws, id string) (any, error) {
		return ok, r.issues.DeleteQuery(ctx, ws, id)
	}),
	actionIssueQueriesDefault: withDefault(func(r *Runtime, ctx context.Context, ws string, in defaultBody) (any, error) {
		q, err := r.issues.SetQueryDefault(ctx, ws, in.ID, in.IsDefault)
		return map[string]any{"queries": q}, err
	}),
}
