# Component Inventory — kandev-plugin-nulab-backlog

Health ratings: healthy / at-risk / degraded. "At-risk" for UI components means at risk for intent 261007, not broken: all tests pass ([code-quality-assessment.md](code-quality-assessment.md#test-coverage-and-baselines)).

## Backend Components (Go)

### Server Entrypoint
- `server/main.go`. Calls `pluginsdk.Serve(plugin.NewRuntime())` only. Health: healthy.

### KandevAdapter
- `internal/plugin/`. Routes 41 actions, OAuth webhook, `task.deleted`, host port, `#` reference source; maps domain errors to `ActionError`. Depends on: Connection, Issues, Git, BacklogGateway, Redact, `pluginsdk`. Health: healthy.

### BacklogGateway
- `internal/backlog/`. Only Backlog API v2 client; per-group rate limiter; OAuth exchange and refresh; stateless about credentials. Depends on: Redact. Health: at-risk (process-wide rate-limit queue, `client.go:346`). Skimmed this run.

### Connection
- `internal/connection/`. API key / OAuth connection, lifecycle, space change, project selection, record and secret store, Git credential. Depends on: BacklogGateway, Redact. Health: healthy. Skimmed this run.

### Issues
- `internal/issues/`. Issue list and filters, create/link task, issue panel and comments, `#` suggestions, `Syncer` (1-minute tick, per-workspace interval 1–1440 min). Depends on: BacklogGateway, Connection, Redact. Health: at-risk (one GET per linked issue per cycle; store and loop code copied from Git). Skimmed this run.

### Git
- `internal/git/`. Repository provider, PR link/status/create, PR watches (`Watcher`, fixed 5-minute ticker, states `open|closed|merged`), saved queries (`RunQuery` max 20 PRs of one repository). Depends on: BacklogGateway, Connection, Redact. Health: at-risk (unpruned watch ledger `store.go:153`; no unfiltered PR list).

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
- `ui/src/index.ts`. All extension points ([api-documentation.md](api-documentation.md#kandev-ui-extension-points)); registers 3 nav items and 3 routes. Health: at-risk (intent item 2 changes it and its test).

### UI Brand
- `ui/src/brand/backlog-logo.tsx`. Filled official Backlog icon (`#42CE9F` tile, `white` mark); `PLUGIN_ICON` is the single selection point (card, 3 nav entries, 3 topbars); `backlog-logo.test.tsx:55` locks the fills. Health: at-risk (Nulab brand terms; fallback comment suggests non-curated `"plug"`).

### UI Page
- `ui/src/page/BacklogPage.tsx`. `/backlog`: status line, "Open settings" text link, `IssuesPage` once connected. Target for the merged Issue / PR list. Health: healthy.

### UI Settings
- `ui/src/settings/`. `SettingsScreen` (method radio, connect form, replace/change-space confirm, notices; member view when an admin action returns 403), `connected-panel` (Test, Disconnect, `PollInterval`, `GitAccess`), `project-picker`, `confirm-dialog` (inline `role="alertdialog"`, own focus trap), `oauth.ts`, `state.ts`. `SettingsScreen.tsx:321` navigates to `/backlog/watches`. Health: at-risk (raw controls, no framed sections).

### UI Issues
- `ui/src/issues/`. `issues-page` (589 lines: 400 ms debounced search, 3 `<select>` filters, 20 rows per page, desktop `<table>` and phone cards, `<details>` row menu, Refresh), `poll-interval`, `issue-panel`, `issue-badge`, `link-task-dialog`, `links-store`, `issues-state`, `task-menu`, `i18n`. Health: at-risk (raw controls, large file).

### UI Git
- `ui/src/git/`. `watches-page` (`/backlog/watches`), inline `watch-form`, `dashboard-page` (`/backlog/dashboard`, saved query `<select>` + `<table>`), `repository-provider`, `review-provider`, `pr-link`, `create-pr`, `git-access` (in `<details>`), `git-state` (first 50 repositories only, `git-state.ts:89`). Health: at-risk (separate routes, raw controls).

### UI Switch
- `ui/src/switch/`. `host.ui.IntegrationEnabledControl` in the card action slot; `enabled-events.ts` channel because v0.96.0 cannot read the switch back. Health: healthy.

### UI Messages
- `ui/src/messages/en.ts`. English catalogue (272 lines). Health: healthy.

### UI Test Harness
- `ui/src/testing/harness.ts`. Fake host; `host.ui` stubs only `Button`, `Input`, `Label` (plain DOM tags), `Skeleton`, `ChangeRequestDetail`. Health: at-risk (every newly used host component needs a stub).

## External Reference (Kandev v0.96.0, not in this repo)

Model for intent 261007 only; not inventoried or scoped: `apps/web/components/github/github-settings.tsx`, `apps/web/src/plugin-integration-settings-route.tsx`, `apps/web/components/app-sidebar/sections/integrations-section.tsx`, `apps/web/app/github/github-page-client.tsx`, `apps/packages/ui/src/button.tsx`, `apps/packages/ui/src/input.tsx`, `apps/web/lib/plugins/icons.ts`, `apps/packages/plugin-sdk/src/index.ts`. Mapping: [architecture.md](architecture.md#external-reference-kandev-github-integration-ui).
