# API Documentation — kandev-plugin-nulab-backlog

## Kandev Plugin Actions

Declared in `manifest.yaml`, routed in `internal/plugin/` (`runtime.go` `handlers` map, merged from `git_actions.go` and `issue_actions.go` in `init()`). 48 actions, all workspace-scoped (some task-scoped); 7 `admin`, 41 `authenticated`. Keys must match `^[a-z0-9][a-z0-9._-]*$`; `internal/plugin/manifest_test.go` asserts the action list and access.

| Group | Actions | Access |
|---|---|---|
| `connection.*` (9) | `get`, `test`, `list_projects`; `connect_api_key`, `set_enabled`, `start_oauth`, `disconnect`, `set_projects`, `set_git_credential` | first three `authenticated`; the rest `admin` |
| `repositories.*` (2) | `inspect`, `branches` | `authenticated` |
| `git.*` (18) | `repositories.list`; `prs.link`, `prs.unlink`, `prs.create`, `prs.status`, `prs.list`; `links.list`; `impact`; `watches.list`, `watches.save`, `watches.delete`, `watches.run`, `watches.pause`, `watches.resume`; `queries.list`, `queries.save`, `queries.delete`, `queries.run` | `authenticated` |
| `issues.*` (19) | `list`, `filters`, `create_task`, `tasks.search`, `links.list`, `refresh`, `impact`, `settings.get`, `link`, `unlink`, `get`, `comments`; `watches.list`, `watches.save`, `watches.delete`, `watches.run`, `watches.pause`, `watches.resume`; `set_poll_interval` | `authenticated`; `set_poll_interval` `admin` |

Contracts relevant to intent 261007-github-parity-actions:
- `git.queries.save` takes `QueryInput{ID, Name, ProjectKey, RepoName, Statuses, Assignee, Creator}`; no id = create (`newID()`), id = replace. Validation in [business-overview.md](business-overview.md#business-rules-locked-by-code-and-tests). No default flag, no ordering field.
- `git.prs.list` lists PRs of ONE repository with filters (Backlog's PR API is per repository).
- `issues.list` filters by project, status, keyword and numeric assignee ids (`issues.Query.AssigneeIDs`); no "me" value. `issues.watches.*` does have `assignee: anyone|me`.
- `issues.create_task {issueKey, workflowId, workflowStepId, force?}` builds the task with `NewTaskFor` (`internal/issues/types.go:151-160`); no prompt or preset field.
- Missing: any action for quick actions (task prompt presets), saved issue queries, or a default query.

## Webhook, Events, Manifest Surfaces

- Webhook `oauth-callback` — GET, `public`, 1024 bytes; redirect `<public_base_url>/api/plugins/nulab-backlog/webhooks/oauth-callback`.
- Events: `task.deleted`.
- `repository_providers: ["nulab-backlog"]`; `reference_sources`: Backlog issues on `#`.
- `capabilities`: `state`, `secrets`, `api_read: [tasks, repositories]`, `api_write: [tasks]`.
- `config_schema`: OAuth client id, OAuth client secret (secret), public base URL. `ui.bundle`.

## Kandev Host Data API (Go, used through the host port)

`Tasks().Create` with `pluginsdk.CreateTaskInput{Title, Description, Priority, Metadata, WorkflowStepID}` and `Tasks().List`. At v0.96.0 `CreateTaskInput` also offers `StartAgent bool` and `Launch *PluginTaskLaunchOptions{AgentProfileID, ExecutorProfileID, Prompt, PlanMode}` (`apps/backend/pkg/pluginsdk/data_types.go:1120-1136, 1264-1269` in the Kandev checkout); the plugin uses neither.

## Kandev UI Extension Points

Registered in `ui/src/index.ts` (10 calls):

| Extension point | Use today |
|---|---|
| `registerTranslations` | `messages/en.ts` (via `issues/i18n.ts`) |
| `registerIntegrationSettings` | Card `nulab-backlog`, `Component` = `SettingsScreen`, `action` = switch; independent of enabled state (BR5.4/BR7.6/BR7.8); `settingsHref()` = `/settings/workspaces/{ws}/integrations/nulab-backlog` |
| `registerNavItem` x1 | `backlog` → `/backlog`, section `integrations` |
| `registerRoute` x1 | `/backlog` (Issues / Pull requests tabs, `?scope=prs`), `topbar: { title, icon }` |
| `registerRepositoryProvider`, `registerTaskAction` (PR link), `registerReviewProvider` | Git / PR |
| `registerComponent("task-card-tags")`, `registerTaskMenuAction` (Unlink only, `group: "primary"`), `registerTaskPanel` | Issue badge, menu, panel |

Host API used: `api.invokeAction`, `context.getActiveWorkspaceId/subscribeActiveWorkspace/getWorkspaceIds/subscribeWorkspaces/getTaskCreationContext`, `navigate`, `toast`, `i18n.t`, `setIntegrationEnabled`, `useResponsiveBreakpoint`, `utils.formatRelativeTime`.

`host.ui` used: Tabs, Table, Select, Dialog, DropdownMenu, Pagination, Empty, Alert, Skeleton, Checkbox, Label, Input, Button, Card, SettingsSection, ChangeRequestList, ChangeRequestRow (no `action`), ChangeRequestDetail, IntegrationIcon, IntegrationRepositoryFilter, IntegrationEnabledControl. Offered at v0.96.0 but unused: `IntegrationStartTaskMenu`, `IntegrationListToolbar`, `IntegrationScopeBar`, `IntegrationSaveQueryDialog`, `TaskCreateDialog`, `IntegrationCursorPagination`, `TaskRowIndicator`, `Combobox`, `Textarea`, `PageTopbar`. How GitHub uses them (external reference): [architecture.md](architecture.md#external-reference-kandev-github-integration-v0960).

## Backlog API v2 (outbound)

`internal/backlog` — 16 `Client` methods: users/myself, projects, project statuses, project users, issues, issue count, issue, issue comments, issue attachments, Git repositories, pull requests list/get/create, Git access check, OAuth token exchange and refresh. Credentials passed per call; HTTP errors become a typed error with `Status` and `Retry-After`; responses capped with `io.LimitReader`.
