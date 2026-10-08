# API Documentation — kandev-plugin-nulab-backlog

## Kandev Host Contract (`manifest.yaml`)

- `version: "0.5.2"`, `api_version: 2`, `runtime.type: binary`, 4 executables (`linux-amd64`, `linux-arm64`, `darwin-amd64`, `darwin-arm64`), `min_kandev_version: "0.96.0"`.
- Capabilities: `state`, `secrets`, `api_read: [tasks, repositories]`, `api_write: [tasks]`, `events: [task.deleted]`. No capability governs starting processes.
- `repository_providers: [nulab-backlog]`, one `reference_sources` entry, `config_schema` (OAuth client, public base URL), `ui.bundle`.

## Plugin Actions

78 action keys (counted from `manifest.yaml` at v0.5.2):

| Prefix | Count |
|---|---|
| `connection.*` | 9 |
| `repositories.*` | 2 |
| `git.*` | 19 |
| `issues.*` | 25 |
| `scm.*` | 23 |

Each declares `scope` (workspace/task), `access` (authenticated/admin) and `max_body_bytes`. Keys must match `^[a-z0-9][a-z0-9._-]*$`, and every manifest key must have a runtime handler and vice versa (`internal/plugin/manifest_test.go`).

### SCM provider actions (intent area)

All `scope: workspace`, `max_body_bytes: 16384`, handlers in `internal/plugin/scm_actions.go`.

| Key | Access | Body | Service call | Returns |
|---|---|---|---|---|
| `scm.providers.list` | authenticated | — | `Providers` | `{providers: ProviderView[]}` |
| `scm.providers.set_token` | admin | `{provider, token, username?}` | `SetToken` | `ProviderView` |
| `scm.providers.use_cli` | admin | `{provider}` (github or gitlab) | `UseCLI` | `ProviderView` |
| `scm.providers.test` | admin | `{provider}` | `Test` (refusal recorded as `lastError`) | `ProviderView` |
| `scm.providers.remove` | admin | `{provider}` | `RemoveToken` | `ProviderView` |

Other `scm.*`: `repos.search`, `mappings.set`, `prs.*`, `links.list`, `task_prs.list`, `queries.*` (5), `watches.*` (6).

`ProviderView`: `provider`, `state` (`not_configured` / `connected` / `error`), `method?` (`token` / `cli`), `account?`, `lastError?` (`invalid_token`, `missing_scope`, `rate_limited`, `unreachable`, `cli_unavailable`), `mappings`. Never contains a token; readable by every member, so any new field (for example a chosen login) must stay non-secret. TS mirror: `ui/src/git/git-state.ts`.

SCM error mapping (`classifySCM`): 401/403 -> `reconnect_required`; 404 -> `not_found`; 429 -> `rate_limited` with `retryAfterSeconds`; host refused / other HTTP -> `unreachable`; `ErrConflict` -> `conflict`; `ErrNoToken` -> `validation` (field `token`); `ErrCLIUnavailable` -> its own `cli_unavailable` outcome.

## Plugin gRPC Services

- Git credential: `ResolveGitCredential`, `GetGitCredentialBinding` (`internal/plugin/credential.go`) — only for `nulab-backlog` repositories. It does not serve GitHub repositories, so it cannot steer `gh` in a GitHub worktree.

## Webhooks

`oauth-callback`: GET, public, 1 KiB body.

## Kandev Host API Used (pluginsdk v0.96.0)

- `Tasks().Create/List/Get`, `Repositories().List`, secrets (`GetSecret`, `SetSecret`, `DeleteSecret`), state, config, events.
- `CreateTaskInput`: workspace, workflow, step, title, description, priority, metadata, `Repositories`, `Launch{AgentProfileID, ExecutorProfileID, Prompt, PlanMode}`. **No environment field.** The plugin passes neither `Repositories` nor `Launch` today (`host_port.go:120-130`, `scm_actions.go:268-275`).
- `ExecutorProfiles()`: read-only.
- No method exposes Kandev's own GitHub credential or runs a process.

## UI Host API Used

`host.api.invokeAction`, `TaskCreateDialog`, `IntegrationStartTaskMenu`, `openTaskLinkDialog`, registry calls (`registerIntegrationSettings`, `registerNavItem`, `registerRoute`, `registerComponent`, `registerTaskAction`, `registerTaskMenuAction`, `registerTaskPanel`, `registerRepositoryProvider`, `registerReviewProvider`).

## Outbound APIs

- Backlog REST v2 (`internal/backlog`): users, projects, statuses, issues, comments, attachments, git repositories, pull requests, `oauth2/token`.
- GitHub REST (`https://api.github.com`, `Authorization: Bearer`, read-only): `GET /user`, `/user/repos`, `/repos/{o}/{r}`, `/repos/{o}/{r}/pulls[/{n}]`. Any token GitHub accepts works, including a per-account `gh auth token --user <login>`.
- GitLab / Bitbucket REST (read-only).

## Local CLI Contract (server host)

| Command | Used by | Notes |
|---|---|---|
| `gh auth token --hostname github.com` | `cliCommand` (GitHub) | active account only; `--user <login>` is supported by gh 2.97.0 on the scan host; older gh may lack it (detect via `gh auth token --help`) |
| `glab config get token --host gitlab.com` | `cliCommand` (GitLab) | single stored login |
| `gh auth status --json hosts` | not used yet | lists logins per host with an active flag (Kandev `ListGHAccounts` reference) |
