# Code Structure — kandev-plugin-nulab-backlog

## Repository Layout

| Path | Kind | Notes |
|---|---|---|
| `server/main.go` | entrypoint | Only `pluginsdk.Serve(plugin.NewRuntime())` |
| `internal/plugin/` | Go adapter | `runtime.go`, `git_actions.go`, `issue_actions.go` (each merged into one `handlers` map in `init()`), webhook, events, host port, references, config |
| `internal/backlog/` | Go gateway | Backlog API v2 client, per-group rate limiter, OAuth; fake JSON fixtures in `testdata/` |
| `internal/connection/` | Go domain | Address check, API key, OAuth, lifecycle, space change, projects, Git credential, store |
| `internal/issues/` | Go domain (U3) | Service, store, `sync.go` (`Syncer`), events, types |
| `internal/git/` | Go domain (U4) | `service.go` (PR link/status/create, queries), `watcher.go` (`Watcher`), store, host, events, `types.go` |
| `internal/redact/` | Go utility | Masks secrets and Backlog URL query strings |
| `internal/ci/`, `cmd/ci/` | tooling | Secret scan, release preflight, marketplace entry, workflow policy, packaged-host contract driver |
| `internal/pkgverify/`, `cmd/verifypkg/` | tooling | Package verification |
| `internal/testutil/` | test helper | Fake keys and tokens |
| `ui/` | TypeScript UI | See module map |
| `manifest.yaml` | plugin manifest | Actions, webhook, capabilities, config schema, UI bundle |
| `docs/brand/`, `docs/manual-checks/` | docs | Logo source and terms; manual check records |
| `aidlc/` | AI-DLC records | Not product code |

## UI Module Map (`ui/src/`)

| Directory | Main files | Role |
|---|---|---|
| (root) | `index.ts`, `index.test.ts`, `jsx.d.ts` | All Kandev registrations |
| `brand/` | `backlog-logo.tsx` | Filled inline SVG of the official Backlog icon; `PLUGIN_ICON` |
| `page/` | `BacklogPage.tsx` | `/backlog`: status line, "Open settings" link, `IssuesPage` when connected; `settingsHref()` |
| `settings/` | `SettingsScreen.tsx` (342 lines), `connected-panel.tsx`, `project-picker.tsx`, `confirm-dialog.tsx`, `oauth.ts`, `state.ts` | Integration settings card content |
| `issues/` | `issues-page.tsx` (589 lines), `poll-interval.tsx`, `issue-panel.tsx`, `issue-badge.tsx`, `link-task-dialog.tsx`, `links-store.ts`, `issues-state.ts`, `task-menu.ts`, `i18n.ts` | Issue list, panel, badge, linking |
| `git/` | `watches-page.tsx`, `watch-form.tsx`, `dashboard-page.tsx`, `repository-provider.ts`, `review-provider.tsx`, `pr-link.ts`, `create-pr.ts`, `git-access.tsx`, `git-state.ts` | PR watches, dashboard, providers |
| `switch/` | `integration-switch.tsx`, `enabled-events.ts` | Enable switch and in-bundle event channel |
| `messages/` | `en.ts` (272 lines) | English catalogue |
| `testing/` | `harness.ts` | Fake host for tests |

## Code Patterns

- **Go**: tests beside code, table-driven with `t.Run`, Backlog faked with `httptest`, injected clock and wait, errors wrapped with `%w`, package doc comments everywhere.
- **UI**: `createXxx(host, messages, ...)` factories return components; JSX through `h = host.jsx`; each `.tsx` has a sibling test; doc comments cite requirement IDs (US/AC/BR/M/WF).
- **Layout constants**: `STACK`/`FIELD`/`ROW` class strings redeclared in 12 files (debt: [code-quality-assessment.md](code-quality-assessment.md#technical-debt)).
- **Naming**: Go `snake_case.go`; TS kebab-case with two historical PascalCase files (`SettingsScreen.tsx`, `BacklogPage.tsx`).
