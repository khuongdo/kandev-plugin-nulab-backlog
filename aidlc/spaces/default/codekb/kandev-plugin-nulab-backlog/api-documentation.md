# API Documentation — kandev-plugin-nulab-backlog

## Kandev Host Contract (`manifest.yaml`)

- `version: "0.6.0"` (action counts and SCM tables below were taken at v0.5.3 and not recounted), `api_version: 2`, `runtime.type: binary`, 4 executables (`linux-amd64`, `linux-arm64`, `darwin-amd64`, `darwin-arm64`), `min_kandev_version: "0.96.0"`.
- Capabilities: `state`, `secrets`, `api_read: [tasks, repositories]`, `api_write: [tasks]`, `events: [task.deleted]`.
- `repository_providers: [nulab-backlog]`, one `reference_sources` entry, `config_schema` (OAuth client, public base URL), `ui.bundle`.
- The frozen snapshot `internal/plugin/testdata/v030/manifest.yaml` must stay unchanged.

## Plugin Actions

79 action keys (counted from `manifest.yaml` at v0.5.3):

| Prefix | Count |
|---|---|
| `connection.*` | 9 |
| `repositories.*` | 2 |
| `git.*` | 19 |
| `issues.*` | 25 |
| `scm.*` | 24 |

Each declares `scope` (workspace/task), `access` (authenticated/admin) and `max_body_bytes`. Keys must match `^[a-z0-9][a-z0-9._-]*$`; every manifest key must have a runtime handler and vice versa (`internal/plugin/manifest_test.go`, `TestSCM_Manifest_Actions`). Admin-only is enforced by the manifest `access: admin`.

### SCM actions used by the Source Control settings page (intent area)

All `scope: workspace`, `max_body_bytes: 16384`, handlers in `internal/plugin/scm_actions.go`.

| Key | Access | Body | Service call | Returns |
|---|---|---|---|---|
| `scm.providers.list` | authenticated (allowed by `guarded()` while Backlog is off) | — | `Providers` | `{providers: ProviderView[]}`, always all three |
| `scm.providers.set_token` | admin | `{provider, token, username?}` | `SetToken` | `ProviderView` |
| `scm.providers.use_cli` | admin | `{provider, login?}` (github or gitlab) | `UseCLI` | `ProviderView` |
| `scm.providers.cli_accounts` | admin | — | gh account list | `{accounts: CliAccount[]}` |
| `scm.providers.test` | admin | `{provider}` | `Test` (refusal recorded as `lastError`) | `ProviderView` |
| `scm.providers.remove` | admin | `{provider}` | `RemoveToken` (data kept, disabled) | `ProviderView` |
| `scm.repos.search` | admin | `{provider, query}` | `SearchRepos` | `{repos: {fullName, url}[]}` |
| `scm.mappings.set` | admin | `{provider, projectKey, repos[]}` | `SetMapping` (max 20; each repo checked; empty list removes) | `ProviderView` |

Other `scm.*` (`prs.list`, `prs.link`/`unlink`, `links.list`, `task_prs.list`, `queries.*`, `watches.*`) are used by the PR list, watch forms and task/issue panels; each item carries its own `provider`. Every `scm.*` except `providers.list` is refused while Backlog is off (`TestSCM_Actions_GuardRefusesAllButProvidersList`).

There is no action that sets or reads an "active" provider. A new one (e.g. `scm.providers.set_active`) touches `manifest.yaml`, `scmHandlers`, `TestSCM_Manifest_Actions` and likely the `guarded()` allow-list.

`ProviderView`: `provider`, `state` (`not_configured` / `connected` / `error`), `method?` (`token` / `cli`), `account?`, `lastError?` (`invalid_token`, `missing_scope`, `rate_limited`, `unreachable`, `cli_unavailable`), `mappings`. Never contains a token; readable by every member, so any new field must stay non-secret. TS mirror: `ui/src/git/git-state.ts`.

SCM error mapping (`classifySCM`): 401/403 -> `reconnect_required`; 404 -> `not_found`; 429 -> `rate_limited` with `retryAfterSeconds`; host refused / other HTTP -> `unreachable`; `ErrConflict` -> `conflict`; `ErrNoToken` -> `validation` (field `token`); `ErrCLIUnavailable` -> `cli_unavailable`.

## Host State and Secrets Used by SCM

- Workspace state keys: `scm.settings`, `scm.links`, `scm.dismissed`, `scm.queries`, `scm.watches`, `scm.ledger`; instance key `scm.index`. Each document is `{schemaVersion: 1, items}`.
- Secret key: `backlog.scm.<provider>.<workspace>`.

## Plugin gRPC Services

- Git credential: `ResolveGitCredential`, `GetGitCredentialBinding` (`internal/plugin/credential.go`) — only for `nulab-backlog` repositories.

## Webhooks

`oauth-callback`: GET, public, 1 KiB body.

## Kandev Host API Used (pluginsdk v0.96.0)

- `Tasks().Create/List/Get`, `Repositories().List`, secrets (`GetSecret`, `SetSecret`, `DeleteSecret`), state, config, events.
- `CreateTaskInput` has no environment field; `ExecutorProfiles()` is read-only (from the previous run, not re-verified).

## UI Host API Used

`host.api.invokeAction`, `TaskCreateDialog`, `IntegrationStartTaskMenu`, `ChangeRequestRow`, `openTaskLinkDialog`, `host.toast.success/error`, host UI kit (`SettingsSection`, `Badge`, `Button`, `Input`, `Label`, `Select*`, `Alert`), registry calls (`registerIntegrationSettings`, `registerNavItem`, `registerRoute`, `registerComponent`, `registerTaskAction`, `registerTaskMenuAction`, `registerTaskPanel`, `registerRepositoryProvider`, `registerReviewProvider`).

### Host context API (`host.context`, SDK `PluginContextApi`, `plugin-sdk/src/index.ts:166-180`)

| Method | Returns | Used by |
|---|---|---|
| `getActiveWorkspaceId()` / `subscribeActiveWorkspace(listener)` | workspace id | `BacklogPage.tsx:109,130` |
| `getTaskCreationContext(workspaceId)` | `{workspaceId, workflowId, defaultStepId, steps[], repositories[]}` or `null` (no workflow **or** steps not loaded in the web store) | `start-task.tsx:54`, three watch dialogs |
| `subscribeTaskCreationContext(workspaceId, listener)` | unsubscribe | unused |

There is no host API to list workflows or steps. `TaskCreateDialog` props: `workspaceId`, `workflowId: string | null`, `defaultStepId: string | null`, `steps`, `initialValues`, `onSuccess`, `onOpenChange`; with `workflowId: null` the host resolves the workflow and fetches steps itself. Details: [architecture.md](architecture.md#task-creation-context-host-derived-kandev-v0960).

## Outbound APIs

- Backlog REST v2 (`internal/backlog`): users, projects, statuses, issues, comments, attachments, git repositories, pull requests, `oauth2/token`.
- GitHub (`https://api.github.com`), GitLab (`https://gitlab.com/api/v4`), Bitbucket (`https://api.bitbucket.org/2.0`), read-only, 5 methods each: `CurrentUser`, `SearchRepos`, `GetRepo`, `ListPRs`, `GetPR`.

## Local CLI Contract (server host)

| Command | Used by | Notes |
|---|---|---|
| `gh auth token --hostname github.com --user <login>` | GitHub CLI method | falls back to the active account on gh without `--user` (< 2.40) |
| `gh auth status --json hosts` | `scm.providers.cli_accounts` | gh >= 2.81.0 |
| `glab config get token --host gitlab.com` | GitLab CLI method | single stored login |
