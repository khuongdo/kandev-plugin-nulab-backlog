# Code Structure — kandev-plugin-nulab-backlog

## Repository Layout

| Path | Kind | Notes |
|---|---|---|
| `server/main.go` | entrypoint | Only `pluginsdk.Serve(plugin.NewRuntime())` |
| `internal/plugin/` | Go adapter | `runtime.go` (Runtime struct, `handlers` map), `git_actions.go`, `issue_actions.go` (merged in `init()`), `host_port.go` (`hostPort`, `issueHost`), webhook, events, references, credential, config |
| `internal/backlog/` | Go gateway | Backlog API v2 client, per-group rate limiter, OAuth; fake JSON fixtures in `testdata/` |
| `internal/connection/` | Go domain | Address check, API key, OAuth, lifecycle, space change, projects, Git credential, store |
| `internal/issues/` | Go domain (U3) | `service.go` (list, `CreateTask`, `createLinkedTask`), `types.go` (`Query`, `NewTask`, `NewTaskFor`), `store.go` (`issues.links`, `issues.settings`, `issues.index`), `sync.go` (`Syncer`), `watch.go` / `watch_store.go` / `watcher.go` (issue watches), events |
| `internal/git/` | Go domain (U4) | `service.go` (PR link/status/create, `ListQueries`/`SaveQuery`/`DeleteQuery`/`RunQuery`, PR list), `types.go` (`QueryInput` + validation), `store.go` (state keys and limits, `git.queries` max 50), `watcher.go` (`Watcher`), `prs.go`, host, events |
| `internal/redact/` | Go utility | Masks secrets and Backlog URL query strings |
| `internal/ci/`, `cmd/ci/` | tooling | Secret scan, release preflight, marketplace entry, workflow policy, packaged-host contract driver |
| `internal/pkgverify/`, `cmd/verifypkg/` | tooling | Package verification |
| `internal/testutil/` | test helper | Fake keys and tokens |
| `ui/` | TypeScript UI | See module map |
| `manifest.yaml` | plugin manifest | 48 actions, webhook, capabilities, config schema, UI bundle |
| `docs/brand/`, `docs/manual-checks/` | docs | Logo source and terms; manual check records |
| `aidlc/` | AI-DLC records | Not product code |

## UI Module Map (`ui/src/`)

| Directory | Main files | Role |
|---|---|---|
| (root) | `index.ts`, `jsx.d.ts`, `layout.ts` (`STACK`, `FIELD`, `ROW`, `BUTTON` class strings; no CSS shipped), `host-ui.ts` (`hostUi(host)` loose-props accessor), `icons.tsx` (`icon(host, name, className)` inline SVGs) | Registrations and shared kit |
| `brand/` | `backlog-logo.tsx` | Backlog icon; `PLUGIN_ICON` |
| `page/` | `BacklogPage.tsx` | `/backlog`: host `Tabs` Issues / Pull requests (`?scope=prs`, each tab mounted once opened); root `STACK`, no padding; `settingsHref()` |
| `issues/` | `issues-page.tsx` (625 lines: search, filters, `Table`/cards, `RowMenu`), `issues-state.ts`, `task-menu.ts`, `i18n.ts`, `issue-panel.tsx`, `issue-badge.tsx`, `link-task-dialog.tsx`, `links-store.ts`, `poll-interval.tsx` | Issue list, task menu, panel, badge, linking |
| `git/` | `pr-list.tsx` (386 lines), `pr-toolbar.tsx`, `save-query-dialog.tsx`, `git-state.ts`, `watch-form.tsx`, `repository-provider.ts`, `review-provider.tsx`, `pr-link.ts`, `create-pr.ts`, `git-access.tsx` | PR list, saved query dialog, providers |
| `settings/` | `SettingsScreen.tsx` (441 lines; stacked `SettingsSection`s), `saved-queries-section.tsx`, `pr-watches-section.tsx`, `issue-watches-section.tsx`, `issue-watch-dialog.tsx`, `section-parts.tsx`, `use-list.ts`, `connected-panel.tsx`, `project-picker.tsx`, `confirm-dialog.tsx`, `oauth.ts`, `state.ts` | Integration settings card content |
| `switch/` | `integration-switch.tsx`, `enabled-events.ts` | Enable switch and in-bundle event channel |
| `messages/` | `en.ts` (316 lines) | English catalogue |
| `testing/` | `harness.ts` (579 lines) | Fake host and `host.ui` fakes for Vitest |

## Code Patterns

- **Go**: tests beside code, table-driven with `t.Run`, Backlog faked with `httptest`, injected clock and wait, errors wrapped with `%w`, package doc comments everywhere. Domain state is a workspace-scoped `{schemaVersion, items}` document per key; `internal/issues/store.go` copies the helpers from `internal/git/store.go` (`ponytail:` extract when a third package needs them).
- **UI**: `createXxx(host, messages, ...)` factories return components; JSX through `h = host.jsx`; host components through `hostUi(host)`; layout through `layout.ts` constants; each module has a sibling test; doc comments cite requirement IDs (US/AC/BR/FR).
- **Naming**: Go `snake_case.go`; TS kebab-case with two historical PascalCase files (`SettingsScreen.tsx`, `BacklogPage.tsx`).
