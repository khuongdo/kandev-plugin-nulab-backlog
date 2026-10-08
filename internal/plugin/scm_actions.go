package plugin

import (
	"context"
	"errors"
	"maps"
	"math"
	"slices"
	"strings"

	"github.com/kandev/kandev/pkg/pluginsdk"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/bitbucket"
	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/connection"
	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/github"
	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/gitlab"
	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/issues"
	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/scm"
)

// Source control action keys declared in manifest.yaml (intent
// 261007-source-control-agnostic). They are new keys, so every v0.3.0 key
// and input stays as it was (FR7.2).
const (
	actionSCMProviders      = "scm.providers.list"
	actionSCMSetToken       = "scm.providers.set_token" //nolint:gosec // G101: an action key, not a credential
	actionSCMTest           = "scm.providers.test"
	actionSCMRemove         = "scm.providers.remove"
	actionSCMRepos          = "scm.repos.search"
	actionSCMMapping        = "scm.mappings.set"
	actionSCMPRList         = "scm.prs.list"
	actionSCMLink           = "scm.prs.link"
	actionSCMUnlink         = "scm.prs.unlink"
	actionSCMLinks          = "scm.links.list"
	actionSCMTaskPRs        = "scm.task_prs.list"
	actionSCMQueriesList    = "scm.queries.list"
	actionSCMQueriesSave    = "scm.queries.save"
	actionSCMQueriesDelete  = "scm.queries.delete"
	actionSCMQueriesRun     = "scm.queries.run"
	actionSCMQueriesDefault = "scm.queries.set_default"
	actionSCMWatchesList    = "scm.watches.list"
	actionSCMWatchesSave    = "scm.watches.save"
	actionSCMWatchesDelete  = "scm.watches.delete"
	actionSCMWatchesRun     = "scm.watches.run"
	actionSCMWatchesPause   = "scm.watches.pause"
	actionSCMWatchesResume  = "scm.watches.resume"
	// actionSCMUseCLI connects GitHub or GitLab with the gh / glab login of
	// the Kandev server (intent 261008-gh-cli-auth, FR1.2, FR2.1).
	actionSCMUseCLI = "scm.providers.use_cli"
	// actionSCMCLIAccounts lists gh's github.com logins (intent
	// 261008-gh-cli-profile, FR1.1).
	actionSCMCLIAccounts = "scm.providers.cli_accounts"
	// actionSCMActiveSet makes one service the workspace's source control
	// service (intent 261008-source-control-settings, FR2.1).
	actionSCMActiveSet = "scm.active.set"
)

const (
	// codeCLIUnavailable: gh / glab is missing or not logged in on the Kandev server (FR4.1).
	codeCLIUnavailable = "cli_unavailable"
	// codeCLIAccountMissing: gh has no login for the workspace's chosen account (FR5.1).
	codeCLIAccountMissing = "cli_account_missing"
	// codeServiceInactive: the action's service is not the workspace's active
	// one; the error names the active service (FR1.4).
	codeServiceInactive = "service_inactive"
)

func init() { maps.Copy(handlers, scmHandlers) }

// scmClients are the production provider clients (FR1.2).
func scmClients() map[scm.Provider]scm.Client {
	return map[scm.Provider]scm.Client{scm.GitHub: github.New(), scm.GitLab: gitlab.New(), scm.Bitbucket: bitbucket.New()}
}

// wireSCM builds the source control service and its watcher on clients.
func (r *Runtime) wireSCM(clients map[scm.Provider]scm.Client) {
	r.scm = scm.NewService(clients, r.service, r.stores, scmHost{r.ports}, scm.NewStore(r.stores))
	r.scmWatcher = scm.NewWatcher(r.scm, r.log)
}

type providerBody struct {
	Provider scm.Provider `json:"provider"`
	Query    string       `json:"query"`
	Login    string       `json:"login"` // use_cli only: the gh account (FR2.1)
}

// withProvider decodes {"provider", "query"} for the provider actions.
func withProvider(fn func(r *Runtime, ctx context.Context, ws string, in providerBody) (any, error)) handler {
	return func(r *Runtime, ctx context.Context, ws string, body []byte) (any, error) {
		var in providerBody
		if err := decode(body, &in); err != nil {
			return nil, err
		}
		return fn(r, ctx, ws, in)
	}
}

// withBody decodes a T body for the actions that take one input type.
func withBody[T any](fn func(r *Runtime, ctx context.Context, ws string, in T) (any, error)) handler {
	return func(r *Runtime, ctx context.Context, ws string, body []byte) (any, error) {
		var in T
		if err := decode(body, &in); err != nil {
			return nil, err
		}
		return fn(r, ctx, ws, in)
	}
}

var scmHandlers = map[string]handler{
	actionSCMProviders: func(r *Runtime, ctx context.Context, ws string, _ []byte) (any, error) {
		return r.scmSettings(ctx, ws)
	},
	actionSCMActiveSet: withBody(func(r *Runtime, ctx context.Context, ws string, in struct {
		Service scm.Provider `json:"service"`
	}) (any, error) {
		if err := r.scm.SetActive(ctx, ws, in.Service); err != nil {
			return nil, err
		}
		return r.scmSettings(ctx, ws)
	}),
	actionSCMSetToken: withBody(func(r *Runtime, ctx context.Context, ws string, in scm.TokenInput) (any, error) {
		return r.scm.SetToken(ctx, ws, in)
	}),
	actionSCMTest: withProvider(func(r *Runtime, ctx context.Context, ws string, in providerBody) (any, error) {
		return r.scm.Test(ctx, ws, in.Provider)
	}),
	actionSCMRemove: withProvider(func(r *Runtime, ctx context.Context, ws string, in providerBody) (any, error) {
		return r.scm.RemoveToken(ctx, ws, in.Provider)
	}),
	actionSCMUseCLI: withProvider(func(r *Runtime, ctx context.Context, ws string, in providerBody) (any, error) {
		return r.scm.UseCLI(ctx, ws, in.Provider, in.Login)
	}),
	actionSCMCLIAccounts: func(r *Runtime, ctx context.Context, _ string, _ []byte) (any, error) {
		a, err := r.scm.ListCLIAccounts(ctx)
		return map[string]any{"accounts": a}, err
	},
	actionSCMRepos: withProvider(func(r *Runtime, ctx context.Context, ws string, in providerBody) (any, error) {
		repos, err := r.scm.SearchRepos(ctx, ws, in.Provider, in.Query)
		return map[string]any{"repos": repos}, err
	}),
	actionSCMMapping: withBody(func(r *Runtime, ctx context.Context, ws string, in scm.MappingInput) (any, error) {
		return r.scm.SetMapping(ctx, ws, in)
	}),
	actionSCMPRList: withBody(func(r *Runtime, ctx context.Context, ws string, in scm.PRListInput) (any, error) {
		return r.scm.ListPRs(ctx, ws, in)
	}),
	actionSCMLink: withBody(func(r *Runtime, ctx context.Context, ws string, in struct {
		URL string `json:"url"`
	}) (any, error) {
		return r.scm.Link(ctx, ws, verified(ctx).TaskID, in.URL)
	}),
	actionSCMUnlink: withBody(func(r *Runtime, ctx context.Context, ws string, in struct {
		Key      string `json:"key"`
		IssueKey string `json:"issueKey"`
	}) (any, error) {
		return ok, r.scm.Unlink(ctx, ws, verified(ctx).TaskID, in.Key, in.IssueKey)
	}),
	actionSCMLinks: withBody(func(r *Runtime, ctx context.Context, ws string, in struct {
		TaskID   string `json:"taskId"`
		IssueKey string `json:"issueKey"`
	}) (any, error) {
		l, err := r.scm.Links(ctx, ws, in.TaskID, in.IssueKey)
		return map[string]any{"links": l}, err
	}),
	actionSCMTaskPRs: func(r *Runtime, ctx context.Context, ws string, _ []byte) (any, error) {
		prs, err := r.taskPRs(ctx, ws, verified(ctx).TaskID)
		return map[string]any{"pullRequests": prs}, err
	},
	actionSCMQueriesList: func(r *Runtime, ctx context.Context, ws string, _ []byte) (any, error) {
		q, err := r.scm.ListQueries(ctx, ws)
		return map[string]any{"queries": q}, err
	},
	actionSCMQueriesSave: withBody(func(r *Runtime, ctx context.Context, ws string, in scm.QueryInput) (any, error) {
		return r.scm.SaveQuery(ctx, ws, in)
	}),
	actionSCMQueriesDelete: withID(func(r *Runtime, ctx context.Context, ws, id string) (any, error) {
		return ok, r.scm.DeleteQuery(ctx, ws, id)
	}),
	actionSCMQueriesRun: withID(func(r *Runtime, ctx context.Context, ws, id string) (any, error) {
		return r.scm.RunQuery(ctx, ws, id)
	}),
	actionSCMQueriesDefault: withDefault(func(r *Runtime, ctx context.Context, ws string, in defaultBody) (any, error) {
		q, err := r.scm.SetQueryDefault(ctx, ws, in.ID, in.IsDefault)
		return map[string]any{"queries": q}, err
	}),
	actionSCMWatchesList: func(r *Runtime, ctx context.Context, ws string, _ []byte) (any, error) {
		w, err := r.scm.ListWatches(ctx, ws)
		return map[string]any{"watches": w}, err
	},
	actionSCMWatchesSave: withBody(func(r *Runtime, ctx context.Context, ws string, in scm.WatchInput) (any, error) {
		return r.scm.SaveWatch(ctx, ws, in)
	}),
	actionSCMWatchesDelete: withID(func(r *Runtime, ctx context.Context, ws, id string) (any, error) {
		return ok, r.scm.DeleteWatch(ctx, ws, id)
	}),
	// One run creates at most one task, so it runs inside the action (FR4.3).
	actionSCMWatchesRun: withID(func(r *Runtime, ctx context.Context, ws, id string) (any, error) {
		n, err := r.scm.RunWatch(ctx, ws, id)
		return map[string]int{"created": n}, err
	}),
	actionSCMWatchesPause: withID(func(r *Runtime, ctx context.Context, ws, id string) (any, error) {
		return r.scm.PauseWatch(ctx, ws, id)
	}),
	actionSCMWatchesResume: withID(func(r *Runtime, ctx context.Context, ws, id string) (any, error) {
		return r.scm.ResumeWatch(ctx, ws, id)
	}),
}

// scmSettings is the Source control section: the providers and the active
// service ("" while an upgraded workspace waits for a pick, FR3.3).
func (r *Runtime) scmSettings(ctx context.Context, ws string) (any, error) {
	v, err := r.scm.Providers(ctx, ws)
	if err != nil {
		return nil, err
	}
	a, err := r.scm.Active(ctx, ws)
	return map[string]any{"providers": v, "active": a}, err
}

// classifySCM maps the scm errors to action codes (NFR5): 401 and 403 are
// reconnect_required, 404 not_found, 429 rate_limited with its wait, and
// anything else from a provider unreachable. ok is false for other errors.
func classifySCM(err error) (out outcome, ok bool) {
	var he *scm.HTTPError
	var inactive *scm.InactiveError
	code := func(c string) (outcome, bool) { return outcome{Outcome: connection.Outcome{Code: c}}, true }
	switch {
	case errors.As(err, &inactive):
		return outcome{Outcome: connection.Outcome{Code: codeServiceInactive}, ActiveService: string(inactive.Active)}, true
	case errors.As(err, &he) && (he.Status == 401 || he.Status == 403):
		return code(connection.CodeReconnectRequired)
	case errors.As(err, &he) && he.Status == 404, errors.Is(err, scm.ErrNotFound):
		return code(codeNotFound)
	case errors.As(err, &he) && he.Status == 429:
		secs := max(int(math.Ceil(he.RetryAfter.Seconds())), 1)
		return outcome{Outcome: connection.Outcome{Code: connection.CodeRateLimited, RetryAfterSeconds: secs}}, true
	case errors.As(err, &he), errors.Is(err, scm.ErrHostRefused):
		return code(connection.CodeUnreachable)
	case errors.Is(err, scm.ErrConflict):
		return code(connection.CodeConflict)
	case errors.Is(err, scm.ErrCLIUnavailable):
		return code(codeCLIUnavailable)
	case errors.Is(err, scm.ErrCLIAccountMissing):
		return code(codeCLIAccountMissing)
	case errors.Is(err, scm.ErrNoToken):
		return outcome{Outcome: connection.Outcome{Code: connection.CodeValidation, Field: scm.FieldToken}}, true
	}
	return outcome{}, false
}

// maxIssueTasks caps the tasks of one issue whose Kandev PRs are read.
const maxIssueTasks = 20

// taskPR is one pull request Kandev attached to a task (FR5.4).
type taskPR struct {
	TaskID     string `json:"taskId"`
	Number     int64  `json:"number"`
	URL        string `json:"url"`
	Title      string `json:"title"`
	State      string `json:"state"`
	IsDraft    bool   `json:"isDraft,omitempty"`
	Provider   string `json:"provider"`
	HeadBranch string `json:"headBranch"`
	BaseBranch string `json:"baseBranch"`
}

// taskPRs returns the GitHub and GitLab pull requests Kandev itself attached
// to every task linked to the same Backlog issue as taskID (FR5.4), of the
// active service only (FR1.5). It uses Kandev's data only, never a plugin
// token. A deleted task is skipped.
func (r *Runtime) taskPRs(ctx context.Context, ws, taskID string) ([]taskPR, error) {
	active, err := r.scm.Active(ctx, ws)
	if err != nil {
		return nil, err
	}
	links, err := r.issues.Links(ctx, ws)
	if err != nil {
		return nil, err
	}
	out := []taskPR{}
	i := slices.IndexFunc(links, func(l issues.LinkView) bool { return l.TaskID == taskID })
	if taskID == "" || i < 0 {
		return out, nil
	}
	var ids []string
	for _, l := range links {
		if l.IssueKey == links[i].IssueKey && len(ids) < maxIssueTasks && !slices.Contains(ids, l.TaskID) {
			ids = append(ids, l.TaskID)
		}
	}
	for _, id := range ids {
		prs, err := r.ports.taskPullRequests(ctx, id)
		// ponytail: the gRPC code read from the text, like workflowRefused.
		if err != nil && strings.Contains(err.Error(), "code = NotFound") {
			continue
		}
		if err != nil {
			return nil, err
		}
		for _, p := range prs {
			if (p.Provider == "github" || p.Provider == "gitlab") && (active == "" || string(active) == p.Provider) {
				out = append(out, taskPR{TaskID: id, Number: p.Number, URL: p.URL, Title: p.Title, State: p.State,
					IsDraft: p.IsDraft, Provider: p.Provider, HeadBranch: p.HeadBranch, BaseBranch: p.BaseBranch})
			}
		}
	}
	return out, nil
}

// scmHost adapts the Host data API to scm.Tasks: a watch's task.
type scmHost struct{ hostPort }

// CreateTask creates a watch task; an empty step means the workflow's default.
func (p scmHost) CreateTask(ctx context.Context, in scm.NewTask) (string, error) {
	t, err := p.create(ctx, pluginsdk.CreateTaskInput{WorkspaceID: in.WorkspaceID, WorkflowID: in.WorkflowID,
		WorkflowStepID: optional(in.WorkflowStepID), Title: in.Title, Description: in.Description, Metadata: in.Metadata})
	if err != nil {
		return "", err
	}
	return t.ID, nil
}
