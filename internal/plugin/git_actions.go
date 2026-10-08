package plugin

import (
	"bytes"
	"context"
	"encoding/json"
	"maps"

	"github.com/kandev/kandev/pkg/pluginsdk"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/connection"
	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/git"
)

// U4 action keys declared in manifest.yaml (C5). repositories.inspect and
// repositories.branches are the names Kandev calls for a repository provider.
const (
	actionSetGitCredential = "connection.set_git_credential"
	actionReposInspect     = "repositories.inspect"
	actionReposBranches    = "repositories.branches"
	actionReposList        = "git.repositories.list"
	actionPRLink           = "git.prs.link"
	actionPRUnlink         = "git.prs.unlink"
	actionPRCreate         = "git.prs.create"
	actionPRStatus         = "git.prs.status"
	actionLinksList        = "git.links.list"
	actionImpact           = "git.impact"
	actionWatchesList      = "git.watches.list"
	actionWatchesSave      = "git.watches.save"
	actionWatchesDelete    = "git.watches.delete"
	actionWatchesRun       = "git.watches.run"
	actionWatchesPause     = "git.watches.pause"
	actionWatchesResume    = "git.watches.resume"
	actionQueriesList      = "git.queries.list"
	actionQueriesSave      = "git.queries.save"
	actionQueriesDelete    = "git.queries.delete"
	actionQueriesRun       = "git.queries.run"
	actionPRList           = "git.prs.list"            // intent 261007 (FR4)
	actionQueriesDefault   = "git.queries.set_default" // intent 261007-github-parity-actions (FR3.3)
)

func init() { maps.Copy(handlers, gitHandlers) }

// gitOffReplies answer the Backlog Git reads with nothing, rather than an
// error, while another source control service is active (FR1.6), so the
// task panels and pickers just show no Backlog Git item.
var gitOffReplies = map[string]any{
	actionWatchesList:  map[string]any{"watches": []any{}},
	actionQueriesList:  map[string]any{"queries": []any{}},
	actionLinksList:    map[string]any{"associations": []any{}},
	actionPRStatus:     map[string]any{"summaries": []any{}},
	actionReposInspect: map[string]bool{"matched": false},
	actionReposList:    map[string]any{"repositories": []any{}},
	actionImpact:       git.Impact{},
}

// verifiedKey carries the verified action context to the U4 handlers, which
// take the task, repository and head branch from it, never from the body.
type verifiedKey struct{}

func verified(ctx context.Context) pluginsdk.VerifiedActionContext {
	v, _ := ctx.Value(verifiedKey{}).(pluginsdk.VerifiedActionContext)
	return v
}

// decode reads a JSON object body; an empty body is an empty object.
func decode(body []byte, v any) error {
	if len(bytes.TrimSpace(body)) == 0 {
		return nil
	}
	if json.Unmarshal(body, v) != nil {
		return errBadBody
	}
	return nil
}

type idBody struct {
	ID string `json:"id"`
}

var ok = map[string]bool{"ok": true}

var gitHandlers = map[string]handler{
	actionSetGitCredential: func(r *Runtime, ctx context.Context, ws string, body []byte) (any, error) {
		var in connection.GitCredentialInput
		if err := decode(body, &in); err != nil {
			return nil, err
		}
		return r.service.SetGitCredential(ctx, ws, in)
	},
	actionReposInspect: func(r *Runtime, ctx context.Context, ws string, body []byte) (any, error) {
		var in struct {
			URL string `json:"url"`
		}
		if err := decode(body, &in); err != nil {
			return nil, err
		}
		d, err := r.git.Inspect(ctx, ws, in.URL)
		if err != nil || d == nil {
			return map[string]bool{"matched": false}, err
		}
		return map[string]any{"repository": d}, nil
	},
	actionReposBranches: func(r *Runtime, ctx context.Context, ws string, body []byte) (any, error) {
		var in struct {
			Repository git.Descriptor `json:"repository"`
		}
		if err := decode(body, &in); err != nil {
			return nil, err
		}
		b, err := r.git.Branches(ctx, ws, in.Repository)
		return map[string]any{"branches": b}, err
	},
	actionReposList: func(r *Runtime, ctx context.Context, ws string, body []byte) (any, error) {
		var in struct {
			Query  string `json:"query"`
			Cursor string `json:"cursor"`
		}
		if err := decode(body, &in); err != nil {
			return nil, err
		}
		return r.git.ListRepositories(ctx, ws, in.Query, in.Cursor)
	},
	actionPRLink: func(r *Runtime, ctx context.Context, ws string, body []byte) (any, error) {
		var in struct {
			Reference string `json:"reference"`
		}
		if err := decode(body, &in); err != nil {
			return nil, err
		}
		v := verified(ctx)
		return r.git.Link(ctx, ws, v.TaskID, in.Reference, v.RepositoryID)
	},
	actionPRUnlink: func(r *Runtime, ctx context.Context, ws string, body []byte) (any, error) {
		var in struct {
			ReviewKey string `json:"reviewKey"`
		}
		if err := decode(body, &in); err != nil {
			return nil, err
		}
		return ok, r.git.Unlink(ctx, ws, verified(ctx).TaskID, in.ReviewKey)
	},
	actionPRCreate: func(r *Runtime, ctx context.Context, ws string, body []byte) (any, error) {
		var in struct {
			Title      string `json:"title"`
			Body       string `json:"body"`
			BaseBranch string `json:"baseBranch"`
		}
		if err := decode(body, &in); err != nil {
			return nil, err
		}
		v := verified(ctx)
		return r.git.CreatePR(ctx, ws, git.CreateInput{TaskID: v.TaskID, RepositoryID: v.RepositoryID, HeadBranch: v.HeadBranch,
			Title: in.Title, Body: in.Body, BaseBranch: in.BaseBranch})
	},
	actionPRStatus: func(r *Runtime, ctx context.Context, ws string, _ []byte) (any, error) {
		s, err := r.git.Status(ctx, ws, verified(ctx).TaskID)
		return map[string]any{"summaries": s}, err
	},
	actionLinksList: func(r *Runtime, ctx context.Context, ws string, _ []byte) (any, error) {
		a, err := r.git.Associations(ctx, ws)
		return map[string]any{"associations": a}, err
	},
	actionImpact: func(r *Runtime, ctx context.Context, ws string, body []byte) (any, error) {
		var in struct {
			ProjectKeys []string `json:"projectKeys"`
		}
		if err := decode(body, &in); err != nil {
			return nil, err
		}
		return r.git.Impact(ctx, ws, in.ProjectKeys)
	},
	actionWatchesList: func(r *Runtime, ctx context.Context, ws string, _ []byte) (any, error) {
		w, err := r.git.ListWatches(ctx, ws)
		return map[string]any{"watches": w}, err
	},
	actionWatchesSave: func(r *Runtime, ctx context.Context, ws string, body []byte) (any, error) {
		var in git.WatchInput
		if err := decode(body, &in); err != nil {
			return nil, err
		}
		return r.git.SaveWatch(ctx, ws, in)
	},
	actionWatchesDelete: withID(func(r *Runtime, ctx context.Context, ws, id string) (any, error) {
		return ok, r.git.DeleteWatch(ctx, ws, id)
	}),
	actionWatchesRun: withID(func(r *Runtime, ctx context.Context, ws, id string) (any, error) {
		return map[string]bool{"queued": true}, r.watcher.Run(ctx, ws, id)
	}),
	actionWatchesPause: withID(func(r *Runtime, ctx context.Context, ws, id string) (any, error) {
		return r.git.PauseWatch(ctx, ws, id)
	}),
	actionWatchesResume: withID(func(r *Runtime, ctx context.Context, ws, id string) (any, error) {
		return r.git.ResumeWatch(ctx, ws, id)
	}),
	actionQueriesList: func(r *Runtime, ctx context.Context, ws string, _ []byte) (any, error) {
		q, err := r.git.ListQueries(ctx, ws)
		return map[string]any{"queries": q}, err
	},
	actionQueriesSave: func(r *Runtime, ctx context.Context, ws string, body []byte) (any, error) {
		var in git.QueryInput
		if err := decode(body, &in); err != nil {
			return nil, err
		}
		return r.git.SaveQuery(ctx, ws, in)
	},
	actionQueriesDelete: withID(func(r *Runtime, ctx context.Context, ws, id string) (any, error) {
		return ok, r.git.DeleteQuery(ctx, ws, id)
	}),
	actionQueriesRun: withID(func(r *Runtime, ctx context.Context, ws, id string) (any, error) {
		rows, err := r.git.RunQuery(ctx, ws, id)
		return map[string]any{"rows": rows}, err
	}),
	actionPRList: func(r *Runtime, ctx context.Context, ws string, body []byte) (any, error) {
		var in git.PRListInput
		if err := decode(body, &in); err != nil {
			return nil, err
		}
		return r.git.ListPullRequests(ctx, ws, in)
	},
	actionQueriesDefault: withDefault(func(r *Runtime, ctx context.Context, ws string, in defaultBody) (any, error) {
		q, err := r.git.SetQueryDefault(ctx, ws, in.ID, in.IsDefault)
		return map[string]any{"queries": q}, err
	}),
}

// withID decodes {"id": ...} for the handlers that act on one watch or query.
func withID(fn func(r *Runtime, ctx context.Context, ws, id string) (any, error)) handler {
	return func(r *Runtime, ctx context.Context, ws string, body []byte) (any, error) {
		var in idBody
		if err := decode(body, &in); err != nil {
			return nil, err
		}
		return fn(r, ctx, ws, in.ID)
	}
}

// defaultBody is the {"id", "isDefault"} body of the *.queries.set_default actions.
type defaultBody struct {
	ID        string `json:"id"`
	IsDefault bool   `json:"isDefault"`
}

// withDefault decodes a defaultBody for the handlers that star a saved query.
func withDefault(fn func(r *Runtime, ctx context.Context, ws string, in defaultBody) (any, error)) handler {
	return func(r *Runtime, ctx context.Context, ws string, body []byte) (any, error) {
		var in defaultBody
		if err := decode(body, &in); err != nil {
			return nil, err
		}
		return fn(r, ctx, ws, in)
	}
}
