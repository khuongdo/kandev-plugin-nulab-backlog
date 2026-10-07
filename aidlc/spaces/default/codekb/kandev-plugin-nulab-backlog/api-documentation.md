# API Documentation — kandev-plugin-nulab-backlog

## Kandev Plugin Actions

Declared in `manifest.yaml`, routed in `internal/plugin/` (`runtime.go` + `git_actions.go` + `issue_actions.go`). 41 actions, all workspace-scoped (some task-scoped); 7 `admin`, 34 `authenticated`. Keys must match `^[a-z0-9][a-z0-9._-]*$`.

| Group | Actions | Access |
|---|---|---|
| `connection.*` (9) | `get`, `test`, `list_projects`; `connect_api_key`, `set_enabled`, `start_oauth`, `disconnect`, `set_projects`, `set_git_credential` | first three `authenticated`; the rest `admin` |
| `repositories.*` (2) | `inspect`, `branches` | `authenticated` |
| `git.*` (17) | `repositories.list`; `prs.link`, `prs.unlink`, `prs.create`, `prs.status`; `links.list`; `impact`; `watches.list`, `watches.save`, `watches.delete`, `watches.run`, `watches.pause`, `watches.resume`; `queries.list`, `queries.save`, `queries.delete`, `queries.run` | `authenticated` |
| `issues.*` (13) | `list`, `filters`, `create_task`, `tasks.search`, `links.list`, `refresh`, `impact`, `settings.get`, `link`, `unlink`, `get`, `comments`; `set_poll_interval` | `authenticated`; `set_poll_interval` `admin` |

Gaps relevant to intent 261007:
- No action lists a repository's PRs without a saved query: `git.queries.run` needs a saved query id and returns at most 20 PRs of one repository (`internal/git/service.go:868`), no paging.
- No issue-watch action (auto-create tasks from new issues).

## Webhook, Events, Manifest Surfaces

- Webhook `oauth-callback` — GET, `public`, 1024 bytes; redirect `<public_base_url>/api/plugins/nulab-backlog/webhooks/oauth-callback`.
- Events: `task.deleted`.
- `repository_providers: ["nulab-backlog"]`; `reference_sources`: Backlog issues on `#`.
- `capabilities`: `state`, `secrets`, `api_read: [tasks, repositories]`, `api_write: [tasks]`.
- `config_schema`: OAuth client id, OAuth client secret (secret), public base URL. `ui.bundle`.

## Kandev UI Extension Points

Registered in `ui/src/index.ts`:

| Extension point | Use today | Lock |
|---|---|---|
| `registerIntegrationSettings` | Card `nulab-backlog`, `Component` = `SettingsScreen`, `action` = switch | Independent of enabled state (BR5.4/BR7.6/BR7.8, `index.ts:48-49`); `settingsHref()` = `/settings/workspaces/{ws}/integrations/nulab-backlog` |
| `registerNavItem` x3, `section: "integrations"` | `backlog` → `/backlog` (`index.ts:58-64`); `backlog-watches` → `/backlog/watches`, `backlog-dashboard` → `/backlog/dashboard` (`index.ts:72-83`) | `index.test.ts:104-105,120-122` asserts 3 calls and both extra paths |
| `registerRoute` x3 | Same paths, `topbar: { title, icon }` | same test |
| `registerRepositoryProvider`, `registerTaskAction` (PR link), `registerReviewProvider` | Git / PR | `review-provider.tsx` uses `host.ui.ChangeRequestDetail` |
| `registerComponent("task-card-tags")`, `registerTaskMenuAction`, `registerTaskPanel` | Issue badge, menu, panel | |
| `registerTranslations` | `messages/en.ts` | |

Host API used: `api.invokeAction`, `context.*` (active workspace, task-creation context), `navigate`, `setIntegrationEnabled`, `useResponsiveBreakpoint`, `utils.formatRelativeTime`, `i18n`.

`host.ui` used in non-test code: `Button` (13 files), `Input` (8), `Label` (8), `Skeleton` (1), `IntegrationEnabledControl` (1), `ChangeRequestDetail` (1). Everything else SDK v0.96.0 offers is unused; the offer and its limits are in [architecture.md](architecture.md#external-reference-kandev-github-integration-ui).

## Backlog API v2 (outbound)

`internal/backlog` — 16 `Client` methods: users/myself, projects, project statuses, project users, issues, issue count, issue, issue comments, issue attachments, Git repositories, pull requests list/get/create, Git access check, OAuth token exchange and refresh. Credentials passed per call; HTTP errors become a typed error with `Status` and `Retry-After`; responses capped with `io.LimitReader`.
