# API Documentation — kandev-plugin-nulab-backlog

## Kandev Plugin Actions

Declared in `manifest.yaml`, routed in `internal/plugin/` (`handlers` map). 55 actions, all workspace-scoped (some task-scoped); 7 `admin`, 48 `authenticated`. Keys must match `^[a-z0-9][a-z0-9._-]*$`; `internal/plugin/manifest_test.go` asserts the list and access. Every action except `connection.get` and `connection.set_enabled` is refused while Backlog is off (`runtime.go` `guarded`).

| Group | Actions | Access |
|---|---|---|
| `connection.*` (9) | `get`, `test`, `list_projects`; `connect_api_key`, `set_enabled`, `start_oauth`, `disconnect`, `set_projects`, `set_git_credential` | first three `authenticated`; the rest `admin` |
| `repositories.*` (2) | `inspect`, `branches` — fixed names Kandev calls for this plugin's repository provider; descriptor carries `provider_id` | `authenticated` |
| `git.*` (19) | `repositories.list`; `prs.link`, `prs.unlink`, `prs.create`, `prs.status`, `prs.list`; `links.list`; `impact`; `watches.list/save/delete/run/pause/resume`; `queries.list/save/delete/run/set_default` | `authenticated` |
| `issues.*` (25) | `list`, `filters`, `create_task`, `tasks.search`, `links.list`, `refresh`, `impact`, `settings.get`, `link`, `unlink`, `get`, `comments`, `set_poll_interval`; `watches.*` (6); `quick_actions.get/save`; `queries.list/save/delete/set_default` | `authenticated`; `set_poll_interval` `admin` |

Source-control relevant contracts (intent 261007-source-control-agnostic):
- No `git.*` or `repositories.*` action takes a provider argument; all assume Backlog Git of the connected space. A second provider needs payload-level dispatch on `provider_id` or new action keys (existing keys must stay unchanged).
- `connection.set_git_credential` (admin) stores `{username, password}` for the connected Backlog host only; there is no per-provider credential, owner/workspace or repository allowlist setting.
- `git.repositories.list` and the repository picker list repositories of the Backlog selected projects only.

## Git Credential Extension (gRPC)

`internal/plugin/credential.go` implements `pluginsdk.GitCredentialHandler`:
- `ResolveGitCredential(ProviderID, Host, Path, TaskID, workspace)` → username/password lease and binding; refused unless Backlog is on, `ProviderID == "nulab-backlog"`, host = connected space host, path `/git/<PROJ>/<repo>` in a selected project.
- `GetGitCredentialBinding` → `<connectionEpoch>.<revision>` only.

## Webhook, Events, Manifest Surfaces

- Webhook `oauth-callback` — GET, `public`; redirect `<public_base_url>/api/plugins/nulab-backlog/webhooks/oauth-callback`.
- Events: `task.deleted`.
- `repository_providers: ["nulab-backlog"]` (one entry, `manifest.yaml:31`); `reference_sources`: Backlog issues on `#`.
- `capabilities`: `state`, `secrets`, `api_read: [tasks, repositories]`, `api_write: [tasks]`.
- `config_schema` (operator-level): `oauth_client_id`, `oauth_client_secret` (secret), `public_base_url`. New providers' app credentials would go here; per-workspace secrets need no manifest change.

## Kandev Host Data API (Go, through the host port)

`Tasks().Create` / `Tasks().List`, task metadata (`nulab_backlog_pr`), and `Repositories` (`host_port.go`). `CreateTaskInput.StartAgent` / `Launch` exist at v0.96.0.

## Kandev UI Extension Points

Registered in `ui/src/index.ts`: `registerTranslations`, `registerIntegrationSettings` (card `nulab-backlog`, `SettingsScreen`, switch action), `registerNavItem` / `registerRoute` (`/backlog`), `registerRepositoryProvider` and `registerReviewProvider` (id `PLUGIN_ID = "nulab-backlog"`, URL matcher `BACKLOG_GIT_URL` for `*.backlog.com|backlog.jp|backlogtool.com/git/`), `registerTaskAction` (PR link), issue badge/menu/panel. At v0.96.0 a plugin may register one repository and one review provider per declared `repository_providers` id (`apps/web/lib/plugins/registry.ts:481,504`, external).

## Backlog Outbound APIs

`internal/backlog` (skimmed this run): REST v2 users/myself, projects, statuses, users, issues, comments, attachments, `/projects/{key}/git/repositories[/{repo}/pullRequests[/count|/{n}]]`, PR create, OAuth exchange/refresh; Git smart-HTTP probe `/git/{PROJ}/{repo}.git/info/refs?service=git-upload-pack` (Basic auth). Credentials passed per call; HTTP errors become `*backlog.Error` with `Status`/`Retry-After`; responses capped with `io.LimitReader`. No GitHub, Bitbucket or other vendor client exists.
