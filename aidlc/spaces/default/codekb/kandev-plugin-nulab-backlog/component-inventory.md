# Component Inventory — kandev-plugin-nulab-backlog

Health ratings: healthy / at-risk / degraded. "At-risk" for UI components means at risk for intent 261007-github-parity-actions, not broken: all tests pass ([code-quality-assessment.md](code-quality-assessment.md#test-coverage-and-baselines)).

## Backend Components (Go)

### Server Entrypoint
- `server/main.go`. Calls `pluginsdk.Serve(plugin.NewRuntime())` only. Health: healthy. Not re-scanned this run.

### KandevAdapter
- `internal/plugin/`. Routes 48 actions through one `handlers` map, OAuth webhook, `task.deleted`, host port (`hostPort`, `issueHost`), `#` reference source; maps domain errors to `ActionError`. Depends on: Connection, Issues, Git, BacklogGateway, Redact, `pluginsdk`. Health: healthy (`host_port.go:119-125` `ponytail:` gRPC code read by text match). New actions need manifest entries plus `manifest_test.go` updates.

### BacklogGateway
- `internal/backlog/`. Only Backlog API v2 client; per-group rate limiter; OAuth exchange and refresh; stateless about credentials. Depends on: Redact. Health: at-risk (process-wide rate-limit queue, `client.go:346`). Not re-scanned this run.

### Connection
- `internal/connection/`. API key / OAuth connection, lifecycle, space change, project selection, record and secret store (incl. `ConnectedUserID`), Git credential. Depends on: BacklogGateway, Redact. Health: healthy. Not re-scanned this run.

### Issues
- `internal/issues/`. Issue list and filters (numeric assignee ids only), `CreateTask` with `NewTaskFor` (title = summary, description = body + link, priority mapped; no prompt, no agent launch), link/unlink, issue panel and comments, `#` suggestions, `Syncer`, issue watches and watcher. Depends on: BacklogGateway, Connection, Redact. Health: at-risk (store helpers copied from Git; `service.go:381` no relink when a link write fails after create; 893-line service).

### Git
- `internal/git/`. Repository provider, PR link/status/create, PR list (`git.prs.list`, one repository), PR watches (`Watcher`), saved PR queries (`git.queries`, max 50, repository required, no default flag). Depends on: BacklogGateway, Connection, Redact. Health: at-risk (unpruned watch ledger `store.go:153`; 925-line service).

### Redact
- `internal/redact/`. Masks secrets and Backlog URL query strings. Health: healthy.

### CI Tooling
- `internal/ci/`, `cmd/ci/`. Secret scan, release preflight, marketplace entry, workflow policy, packaged-host contract driver. Health: healthy. Skimmed.

### PackageVerify
- `internal/pkgverify/`, `cmd/verifypkg/`. Package verification; fails if `ui/bundle.js` references a Nulab/Backlog asset URL (`pkgverify.go:46,79`). Health: healthy.

### TestUtil
- `internal/testutil/`. Fake keys and tokens. Health: healthy.

## UI Components (TypeScript)

### UI Registration
- `ui/src/index.ts`. All extension points ([api-documentation.md](api-documentation.md#kandev-ui-extension-points)); one nav item and one route `/backlog`; task menu action is Unlink only. Health: healthy.

### UI Shared Kit
- `ui/src/layout.ts`, `ui/src/host-ui.ts`, `ui/src/icons.tsx`. Tailwind class constants (`STACK`, `FIELD`, `ROW`, `BUTTON`; no page padding constant), `hostUi(host)` loose-props accessor, inline icon factory. Health: healthy.

### UI Brand
- `ui/src/brand/backlog-logo.tsx`. `PLUGIN_ICON` is the single selection point (card, nav entry, topbar). Health: healthy. Not re-scanned this run.

### UI Page
- `ui/src/page/BacklogPage.tsx`. `/backlog` with host `Tabs` (Issues / Pull requests, `?scope=prs`); root `STACK` with no padding while the host `PageShell` adds none, so content runs edge to edge. Health: at-risk (layout alignment target).

### UI Settings
- `ui/src/settings/`. `SettingsScreen` stacks `SettingsSection`s (connection, PR watches, issue watches, saved queries, ...); `saved-queries-section` lists PR queries with rename/delete (new queries only from the PR list); sections save per dialog (no page-level Save/Discard). Health: at-risk (a quick-actions section would be added here; 441-line screen).

### UI Issues
- `ui/src/issues/`. `issues-page` (625 lines: 400 ms debounced search, Project/Status/Assignee `Select`s, 20 rows per page, desktop `Table` and phone cards, "..." `RowMenu` with Create task / Link to task, icon Refresh; no saved queries or presets), `issues-state`, `task-menu` (Unlink only), `i18n`, panel, badge, linking. Health: at-risk (large file; differs from PR list in rows and toolbar).

### UI Git
- `ui/src/git/`. `pr-list` (loads `git.queries.list`, "Query" preset `Select`, `START` with empty repo so the list opens empty, `ChangeRequestRow` without `action`), `pr-toolbar` (hand-made imitation of `IntegrationListToolbar`, no border, no padding, no search), `save-query-dialog` (plugin-built, name only), `git-state` (first 50 repositories only, `git-state.ts:89`), watch form, providers. Health: at-risk (all three intent items touch it).

### UI Switch
- `ui/src/switch/`. `host.ui.IntegrationEnabledControl` in the card action slot; `enabled-events.ts` channel because v0.96.0 cannot read the switch back. Health: healthy.

### UI Messages
- `ui/src/messages/en.ts`. English catalogue (316 lines). Health: healthy.

### UI Test Harness
- `ui/src/testing/harness.ts`. Fake host with fakes for the `host.ui` components in use; `ChangeRequestRow` fake ignores `action`. Health: at-risk (every newly used host component needs a fake).

## External Reference (Kandev v0.96.0, not in this repo)

Model for intent 261007-github-parity-actions only; not inventoried or scoped: `apps/web/components/github/my-github/{action-presets.ts,quick-task-launcher.tsx,search-bar.tsx,use-default-query-presets.ts,saved-preset-model.ts,issue-list.tsx}`, `apps/web/components/github/action-presets-section.tsx`, `apps/web/components/integrations/{integration-start-task-menu.tsx,integration-list-toolbar.tsx,presets-scope-bar-base.tsx,integration-save-query-dialog.tsx,saved-query-default-button.tsx,change-request-list.tsx}`, `apps/web/app/github/github-page-client.tsx`, `apps/web/lib/plugins/host-api.ts`. Mapping: [architecture.md](architecture.md#external-reference-kandev-github-integration-v0960).
