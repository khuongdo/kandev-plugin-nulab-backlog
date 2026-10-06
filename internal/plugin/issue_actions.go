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
}
