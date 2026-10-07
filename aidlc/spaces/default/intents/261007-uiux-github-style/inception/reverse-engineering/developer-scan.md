## Developer Code Scan Results

Full rescan of the repo (snapshot `./`) at commit `86ae473` on branch `feature/refactor-uiux-2bi`. The scan was read-only for tracked files: `ui/node_modules/` (from `npm ci`) and a temporary `build/coverage.out` (deleted after use, `build/` removed) were the only on-disk changes, both gitignored. The UI (`ui/src/**`) was read in depth because the intent is a UI/UX refactor.

### Scan Coverage
- **Analyzed deeply**:
  - `manifest.yaml`
  - `Makefile`
  - `go.mod`
  - `.kandev-sdk-ref`
  - `.nvmrc`
  - `.gitignore`
  - `server/main.go`
  - `internal/plugin/` (`runtime.go`, `git_actions.go`, `issue_actions.go`, `manifest_test.go` test list)
  - `internal/git/` (`types.go`, `service.go` watch/query paths, `watcher.go` timer)
  - `ui/package.json`
  - `ui/tsconfig.json`
  - `ui/vitest.config.ts`
  - `ui/eslint.config.js`
  - `ui/.prettierrc`
  - `ui/src/index.ts`
  - `ui/src/index.test.ts`
  - `ui/src/jsx.d.ts`
  - `ui/src/brand/`
  - `ui/src/page/`
  - `ui/src/settings/`
  - `ui/src/git/`
  - `ui/src/issues/`
  - `ui/src/switch/`
  - `ui/src/testing/harness.ts`
  - `ui/src/messages/en.ts`
  - `docs/brand/backlog-logo.md`
- **Skimmed only**:
  - `internal/backlog/` (doc, method list, `ponytail:` markers, `testdata/` names)
  - `internal/connection/` (doc, imports, file sizes, `ponytail:` markers)
  - `internal/issues/` (doc, method list, `sync.go` header, `ponytail:` markers)
  - `internal/redact/`, `internal/testutil/`
  - `internal/ci/`, `cmd/ci/`
  - `internal/pkgverify/`, `cmd/verifypkg/` (asset-URL rule only)
  - `.github/workflows/` (job names, targets, toolchain pins)
  - `.golangci.yml`
  - `README.md`, `docs/manual-checks/`
- **External reference (outside the snapshot, not part of the analyzed paths)**: Kandev checkout `/home/k_do_webfrontier/repo/kandev`, tag `v0.96.0`, commit `f099a46dc7aab16f6ff5806cd29b2b480296303f` (equal to `.kandev-sdk-ref`). Read only to describe the GitHub integration UI and what SDK v0.96.0 offers.

### Packages Found
- `server` — binary entrypoint — Go — only `pluginsdk.Serve(plugin.NewRuntime())`.
- `internal/plugin` — adapter (KandevAdapter) — Go — the only package besides `server/` that imports `pluginsdk`. Routes the 41 actions (`runtime.go` plus `git_actions.go` and `issue_actions.go`, each merged into one `handlers` map in `init()`), the OAuth callback webhook, error mapping to `ActionError`, host port, `#` reference source.
- `internal/backlog` — gateway — Go — the only Backlog API v2 client. Credentials are passed per call. Per-group rate limiter; OAuth token exchange and refresh.
- `internal/connection` — domain service — Go — API-key/OAuth connection, lifecycle, space change, record and secret store, Git credential.
- `internal/issues` — domain service (U3) — Go — issue list and filters, create/link task, issue panel and comments, `#` suggestions, status sync loop (`Syncer`, 1-minute tick, per-workspace poll interval 1–1440 min).
- `internal/git` — domain service (U4) — Go — repository provider, PR link/status/create, **PR watches** (`Watcher`, fixed 5-minute timer, PR states `open|closed|merged`), saved PR queries (`RunQuery` returns at most 20 PRs of one repository).
- `internal/redact` — utility — Go — masks secrets and Backlog URL query strings.
- `internal/ci`, `cmd/ci` — tooling — Go — secret scan, release preflight, marketplace entry, workflow policy, packaged-host contract driver.
- `internal/pkgverify`, `cmd/verifypkg` — tooling — Go — package verification; fails if `ui/bundle.js` references a Nulab/Backlog asset URL (`pkgverify.go:46,79`).
- `internal/testutil` — test helper — Go.
- `ui` (`kandev-plugin-nulab-backlog-ui`) — frontend bundle — TypeScript + JSX through the host factory `h = host.jsx`; React comes from the host and is never bundled (`make ui-build` fails if it is):
  - `src/index.ts` — every registration with Kandev (see APIs).
  - `src/brand/backlog-logo.tsx` — Nulab's official Backlog icon as a filled inline SVG (`#42CE9F` tile and `white` mark). `PLUGIN_ICON` is the single place the icon is chosen and is used by the settings card, all 3 nav entries and all 3 route topbars.
  - `src/page/BacklogPage.tsx` — `/backlog`: connection status line, an "Open settings" text link, and `IssuesPage` once connected. `settingsHref()` builds `/settings/workspaces/{ws}/integrations/nulab-backlog`.
  - `src/settings/` — `SettingsScreen.tsx` (342 lines: API key / OAuth radio, connect form, replace/change-space confirm, notices), `connected-panel.tsx` (Test, Disconnect, then `PollInterval` and `GitAccess`), `project-picker.tsx`, `confirm-dialog.tsx` (hand-made inline `role="alertdialog"`, not a host Dialog), `oauth.ts`, `state.ts` (pure reducer).
  - `src/issues/` — `issues-page.tsx` (589 lines: search with 400 ms debounce, 3 filter `<select>`, 20 rows per page, `<table>` on desktop and cards on phones, row actions in a `<details>` menu, Refresh), `poll-interval.tsx` (issue sync interval field), `issue-panel.tsx`, `issue-badge.tsx`, `link-task-dialog.tsx`, `links-store.ts`, `issues-state.ts`, `task-menu.ts`, `i18n.ts`.
  - `src/git/` — `watches-page.tsx` (`/backlog/watches`, PR watch list with Run/Pause/Resume/Edit/Delete), `watch-form.tsx` (inline form, not a dialog), `dashboard-page.tsx` (`/backlog/dashboard`: choose a saved query in a `<select>`, then a `<table>` of up to 20 PRs), `repository-provider.ts`, `review-provider.tsx`, `pr-link.ts`, `create-pr.ts`, `git-access.tsx` (Git username/password inside `<details>`), `git-state.ts`.
  - `src/switch/` — `integration-switch.tsx` (`host.ui.IntegrationEnabledControl` in the card action slot), `enabled-events.ts` (in-bundle channel, because v0.96.0 cannot read the switch back).
  - `src/messages/en.ts` — English catalogue (272 lines), registered with `registerTranslations`.
  - `src/testing/harness.ts` — fake host for tests; `host.ui` only stubs `Button`, `Input`, `Label` (as plain DOM tags), `Skeleton`, `ChangeRequestDetail`.

Internal Go dependencies (acyclic): `plugin → {backlog, connection, git, issues, redact}`; `git → {backlog, connection, redact}`; `issues → {backlog, connection, redact}`; `connection → {backlog, redact}`; `backlog → redact`.

### Build System
- **Type**: Go modules (`go 1.26.0`) + npm (Node 22 per `.nvmrc`) + esbuild; the `Makefile` is the entry point and CI calls exactly its targets.
- **Config Files**: `go.mod`, `go.sum`, `Makefile`, `.kandev-sdk-ref`, `.nvmrc`, `.golangci.yml`, `manifest.yaml`, `ui/package.json`, `ui/package-lock.json`, `ui/tsconfig.json`, `ui/vitest.config.ts`, `ui/eslint.config.js`, `ui/.prettierrc` (`printWidth: 110`), `ui/.prettierignore`.
- **Build Dependencies**:
  - `go.mod`: `replace github.com/kandev/kandev => ../kandev/apps/backend`. `ui/tsconfig.json` maps `@kandev/plugin-sdk` to `../../kandev/apps/packages/plugin-sdk/src/index.ts`. Both need the sibling `../kandev` at `.kandev-sdk-ref` (`make check-sdk`). In this worktree `../kandev` is a symlink to `/home/k_do_webfrontier/repo/kandev`, which is at `v0.96.0` = `f099a46`.
  - `make package` = `build` (5 platforms, `CGO_ENABLED=0`) + `ui-build` (esbuild to `build/ui/bundle.js`) + Kandev's `cmd/plugin-pack`, giving `dist/nulab-backlog-<ver>.tar.gz` and `checksums.txt`.
  - `make coverage` writes `coverage.out` at the **repo root** (`Makefile` coverage target), which the project rule says to keep out of the root. `make clean` deletes it.
  - Small mismatch: `ui/package.json` `build` passes `--jsx-fragment=Fragment`, while `make ui-build` does not. Nothing uses `Fragment` today; adding `<>…</>` later would build under one and fail under the other.

### APIs Discovered
- **Kandev plugin actions** — `manifest.yaml`, `internal/plugin/*.go` — 41 actions, all workspace-scoped: 7 `admin` (`connection.connect_api_key|set_enabled|start_oauth|disconnect|set_projects|set_git_credential`, `issues.set_poll_interval`) and 34 `authenticated`. Groups: `connection.*` 9, `repositories.*` 2, `git.*` 17 (`repositories.list`, `prs.link|unlink|create|status`, `links.list`, `impact`, `watches.list|save|delete|run|pause|resume`, `queries.list|save|delete|run`), `issues.*` 13. There is **no action that lists a repository's PRs without a saved query**: `git.queries.run` needs a saved query id and returns at most 20 rows of one repository.
- **Webhook** — 1: `oauth-callback` (GET, `public`, 1024 bytes).
- **Other manifest surfaces** — `repository_providers: ["nulab-backlog"]`, `reference_sources` (`#` Backlog issues), `events: ["task.deleted"]`, `capabilities` (`state`, `secrets`, `api_read: tasks, repositories`, `api_write: tasks`), `config_schema` (OAuth client id/secret, public base URL), `ui.bundle`.
- **Kandev UI extension points used** — `ui/src/index.ts`:
  - `registerIntegrationSettings` (card `nulab-backlog`, `Component` = `SettingsScreen`, `action` = switch).
  - `registerNavItem` ×3 with `section: "integrations"`: `backlog` → `/backlog` (`index.ts:58-64`), `backlog-watches` → `/backlog/watches`, `backlog-dashboard` → `/backlog/dashboard` (`index.ts:72-83`).
  - `registerRoute` ×3 with `topbar: { title, icon }`.
  - `registerRepositoryProvider`, `registerTaskAction` (PR link), `registerReviewProvider`, `registerComponent("task-card-tags")`, `registerTaskMenuAction`, `registerTaskPanel`, `registerTranslations`.
  - Host API: `api.invokeAction`, `context.*` (active workspace, task-creation context), `navigate`, `setIntegrationEnabled`, `useResponsiveBreakpoint`, `utils.formatRelativeTime`, `i18n`.
  - `host.ui` in use (non-test code): `Button` 13 files, `Input` 8, `Label` 8, `Skeleton` 1, `IntegrationEnabledControl` 1, `ChangeRequestDetail` 1. Nothing else from `host.ui` is used.
- **Backlog API v2 (outbound)** — `internal/backlog/*.go` — 16 `Client` methods (users/myself, projects, statuses, users, issues, count, issue, comments, attachments, git repositories, pull requests list/get/create, Git access check, OAuth token exchange/refresh).

### Reference: Kandev GitHub integration UI (v0.96.0, read-only)
- **Settings > Integrations > GitHub** (`apps/web/components/github/github-settings.tsx`): `GitHubIntegrationPage` wraps everything in `WorkspaceScopedSection`, then `PerWorkspaceSection` stacks framed `SettingsSection` blocks with `space-y-8`: Connection (icon + `GitHubEnabledControl` action), **Review watches** (PR), **Issue watches**, repo scope, action presets, PR analytics, default queries. Each watch section has a header `action` with `Button size="sm" variant="outline"` (Clean up, with `IconTrashX h-4 w-4 mr-1`) and `Button size="sm"` (`IconPlus h-4 w-4 mr-1` + "Add watch"); the body is `Card > CardContent p-0` holding a table; create/edit opens a **dialog** (`ReviewWatchDialog`, `IssueWatchDialog`), never a separate page. A GitHub **issue watch** auto-creates tasks for new issues matching a filter/labels/custom query on its own poll interval.
- **How a plugin's settings are mounted** (`apps/web/src/plugin-integration-settings-route.tsx`): the host wraps the plugin `Component` in an **unframed** `SettingsSection` with the plugin label, description, icon (`h-5 w-5`) and `action` (switch). So the plugin's `SettingsScreen` can stack its own framed `host.ui.SettingsSection` blocks (Connection, PR watches, Issue sync, Git access, Projects) exactly like `PerWorkspaceSection`.
- **Home sidebar Integrations** (`components/app-sidebar/sections/integrations-section.tsx`): one `IntegrationRow` per destination (first-party and every plugin `registerNavItem({section: "integrations"})`), icon `h-4 w-4`, label `truncate`. A row is active for its `href` and any sub-path. GitHub has one destination, `/github`.
- **`/github` page** (`app/github/github-page-client.tsx`): `PageShell` + `PresetsScopeBar` (kind switch **PR / Issue** + preset pills + saved queries) + `ListToolbar` (title, count, query box, last-fetched time, refresh `Button variant="ghost" size="icon"` with `IconRefresh`) + `PRList` / `IssueList` + `ResultsPagination`. When not configured it shows an `Alert` with a link to settings.
- **Buttons / inputs** (`apps/packages/ui/src/button.tsx`, `input.tsx`): shadcn `Button` with variants `default|outline|secondary|ghost|destructive|link` and sizes `default|xs|sm|lg|icon|icon-xs|icon-sm|icon-lg`; GitHub uses `size="sm"` for header actions, `variant="outline"` for secondary actions, `variant="ghost" size="icon"` for icon-only actions, and `className="cursor-pointer"`. `Input` has `controlSize` (default `standard`). `Select*`, `Checkbox`, `Switch`, `Table*`, `Tabs*`, `Card*`, `Dialog*`, `Badge`, `Alert*`, `Empty*`, `Tooltip*`, `Pagination*` are all host components.
- **Icons**: Tabler outline icons (`stroke="currentColor"`, no fill) sized by class. A plugin `PluginIcon` is either a curated name (`lib/plugins/icons.ts`: `bell, bolt, book, bug, calendar, chart, checklist, cloud, database, flask, globe, message, puzzle, robot, rocket, settings, ticket, users`; unknown names fall back to `IconPuzzle`) or a plugin component taking `className`. `host.ui.IntegrationIcon` only offers `filter | merged | pull-request | pull-request-closed`. The plugin cannot import `@tabler/icons-react` (it would bundle React), so other outline icons (plus, refresh, trash) must be hand-drawn inline SVGs or replaced by text buttons.
- **What SDK v0.96.0 lets the plugin reproduce** (`apps/packages/plugin-sdk/src/index.ts`, `PluginUIShape` and `PluginRegistry`):
  - Can: a single nav entry; one `/backlog` page using `host.ui.IntegrationScopeBar` (kinds Issue/PR), `IntegrationListToolbar`, `ChangeRequestList`/`ChangeRequestRow`, `Table*`, `IntegrationCursorPagination`/`Pagination*`, `IntegrationSaveQueryDialog`, `Empty*`, `Alert*`; settings sections with `SettingsSection`, `SettingsCard`, `Card*`, `Dialog*`, `Select*`, `Checkbox`, `Switch`, `Tabs*`, `Badge`; any plugin-drawn icon component.
  - Cannot: GitHub's own `PageShell`, `PresetsScopeBar`, `ListToolbar`, watch tables/dialogs, `GitHubEnabledControl`; sidebar header shortcut icons (first-party only); nested/child nav items (`registerNavItem` is flat); sub-routes under the plugin's integration settings page (only a separate `registerSettingsRoute`); an `AlertDialog` (only `Dialog`).

### Frameworks & Libraries
- Go `1.26.0` (toolchain used for the baseline: go1.26.8) — backend.
- `github.com/kandev/kandev` (`pkg/pluginsdk`) — replaced by `../kandev/apps/backend` @ `f099a46` (`v0.96.0`) — plugin SDK; indirect `hashicorp/go-plugin v1.8.0`, `google.golang.org/grpc v1.83.1`.
- `github.com/stretchr/testify v1.12.1` — Go tests.
- `gopkg.in/yaml.v3 v3.0.1` — manifest/workflow reading in CI tooling.
- `@kandev/plugin-sdk` — path-mapped from the Kandev checkout — TS types only.
- `typescript ~6.0.3`, `esbuild ^0.28.2`, `vitest ^5.0.3`, `jsdom ^30.1.2`, `eslint ^10.12.0` + `typescript-eslint ^8.71.1` + `@eslint/js ^10.0.1`, `prettier ^3.9.9`, `axe-core ^4.14.0` (a11y checks in tests), `react`/`react-dom ^19.3.0` + `@types/*` (tests only; runtime uses `host.React`).
- `golangci-lint v2.14.0` (+ `gosec`) and `actionlint v1.7.12`, run with `go run` in `make lint`.

### Test Coverage
- **Test Directories**: Go tests next to the code in `internal/*/` (86 `_test.go` files) with fixtures in `internal/backlog/testdata/` (fake JSON only); UI tests in `ui/src/**/*.test.{ts,tsx}` (28 files) with `ui/src/testing/harness.ts`.
- **Test Frameworks**: Go `testing` + testify, always `-race`; Vitest + jsdom + axe-core for the UI.
- **Coverage Config**: present for Go — `make coverage`, floor 80% over `./internal/... ./server/...`, only exclusion `server/main.go`. No UI coverage config.
- **Baselines measured during this scan** (no tracked file changed):
  - Go: `go test -race -coverprofile=build/coverage.out ./internal/... ./server/...` — **all 9 test packages pass** (exit 0). Per package: backlog 96.0%, ci 91.0%, connection 94.6%, git 90.8%, issues 91.3%, pkgverify 93.2%, plugin 93.6%, redact 97.4%, testutil 88.0%, server 0.0% (no tests; the only `make coverage` exclusion). **Total 92.9%** from the raw profile (it still includes `server/main.go`; `make coverage` filters it, so the gated figure is the same or slightly higher). Matches the previous 92.9%. `build/coverage.out` was deleted and `build/` removed.
  - UI (in `ui/`, after `npm ci`): `vitest run` **229/229 pass** in 28 files; `tsc --noEmit` clean; `eslint .` clean; `prettier --check .` clean.
  - Environment note: local Node is v25.2.1, while `.nvmrc` and CI use Node 22.

### Code Quality Indicators
- **Linting**: Go — `gofmt`, `go vet`, `.golangci.yml` (default + `gosec`; exceptions as `//nolint:<rule> // reason`), `go mod tidy` check in CI. UI — `ui/eslint.config.js` (recommended + typescript-eslint; `h` allowed as unused), `ui/.prettierrc`, `tsc --strict`. Workflows — actionlint + `go run ./cmd/ci workflows`. No `eslint-disable` or `TODO/FIXME` in non-test UI code; `as unknown as` appears only in tests.
- **CI/CD**: `.github/workflows/ci.yml` (`checks` job: `make check-format vet lint test coverage check-secrets build package verify-package`; `packaged-host-contract` job on Kandev min version), `.github/workflows/release.yml` (`verify`, `contract`, `publish` with provenance).
- **Documentation**: `README.md`, `docs/brand/backlog-logo.md` (logo source, hash, Nulab terms, pre-release check), `docs/manual-checks/`. Every Go package has a doc comment; TS components carry doc comments that cite requirement IDs (US/AC/BR/M/WF).

### Technical Debt Signals
- No Button styling at all: zero `variant=`/`size=` props in non-test UI code, so every one of the 42 `<Button>` usages (Retry, Cancel, Delete, Prev/Next, Edit…) renders as the primary filled button. GitHub uses `sm`, `outline`, `ghost`, `icon`.
- Raw HTML controls instead of host components: `<select>` in `issues/issues-page.tsx:261`, `git/watch-form.tsx:147,175`, `git/dashboard-page.tsx:130,174`; `<input type="radio">` in `settings/SettingsScreen.tsx:220`; `<input type="checkbox">` in `settings/project-picker.tsx:176`, `git/watch-form.tsx:199`; `<table>` in `issues/issues-page.tsx:448`, `git/dashboard-page.tsx:214`; `<details>` menus in `issues/issues-page.tsx:311`, `git/git-access.tsx:111`; plain `<button>` in `issues/issue-badge.tsx:47`, `issues/issue-panel.tsx:123,180`, `issues/link-task-dialog.tsx:122`. They get no host styling (Select/Checkbox/Table/DropdownMenu/Collapsible exist in `host.ui`).
- Layout constants `STACK`/`FIELD`/`ROW` are redefined in 12 component files (`settings/{SettingsScreen,connected-panel,project-picker,confirm-dialog}.tsx`, `issues/{issues-page,issue-panel,link-task-dialog,poll-interval}.tsx`, `git/{watches-page,watch-form,dashboard-page,git-access}.tsx`), with two different `ROW` values (`flex gap-2` vs `flex flex-wrap gap-2`).
- `ConfirmDialog` is an inline `role="alertdialog"` block rendered in the page flow, with its own focus trap, not an overlay; `host.ui.Dialog*` exists.
- Settings are spread over three places: connection + poll interval + Git access in `SettingsScreen`; PR watches on `/backlog/watches`; saved queries on `/backlog/dashboard`. `SettingsScreen.tsx:321` hard-codes `host.navigate("/backlog/watches")` for "Review watches".
- `ui/src/brand/backlog-logo.tsx:31-35` suggests the fallback `"plug"`, which is not a curated host icon name and would silently render `IconPuzzle`.
- `make coverage` writes `coverage.out` at the repo root (see Build System).
- Go `ponytail:` markers already record known ceilings (e.g. `internal/backlog/client.go:346` process-wide rate-limit queue, `internal/issues/store.go:46` and `internal/issues/sync.go:191` code copied from `internal/git`, `internal/git/store.go:153` unpruned ledger, `internal/issues/sync.go:26` one GET per linked issue); UI marker `ui/src/git/git-state.ts:89` (first 50 repositories only, which also limits the repo choices in the watch and query forms).

## Handoff Summary
- **Intent-relevant finding**:
  - (1) Watcher settings into Settings > Integrations > Backlog: PR watches live in `ui/src/git/watches-page.tsx` (route `/backlog/watches`) with an inline `watch-form.tsx`; the issue sync interval is already in Settings (`settings/connected-panel.tsx:435` → `issues/poll-interval.tsx`). The host already wraps `SettingsScreen` in an unframed `SettingsSection` with icon and switch (`kandev/apps/web/src/plugin-integration-settings-route.tsx`), so the screen can stack framed `host.ui.SettingsSection` blocks (Connection, PR watches as `Card` + `Table` + `Dialog`, Issue sync, Git access, Projects) like GitHub's `PerWorkspaceSection`. The backend has **only PR watches** (`internal/git/watcher.go`, `git.watches.*`); there is **no issue watcher** that creates tasks from new Backlog issues. "Issue watcher" can map only to the existing issue status sync interval (`issues.set_poll_interval`) unless a new backend feature is added.
  - (2) One Integrations entry: `ui/src/index.ts:58-64,72-83` registers 3 nav items and 3 routes; `ui/src/index.test.ts:104-105,120-122` asserts `toHaveBeenCalledTimes(3)` and the two extra paths. A single `/backlog` page can switch Issue / PR with `host.ui.IntegrationScopeBar` (or `Tabs`) and use `IntegrationListToolbar`. The PR side today is query-driven only: `git.queries.run` needs a saved query and returns at most 20 PRs of **one repository** (`internal/git/service.go:868`), with no paging. A GitHub-like PR list (all repos, paging, search) needs either saved queries used as presets or a new `git.prs.list`-style action (manifest + adapter + service change).
  - (3) Inputs/buttons: no `variant`/`size` anywhere; raw `<select>`, checkbox, radio, `<table>`, `<details>`, `<button>` listed in Technical Debt; `confirm-dialog.tsx` is inline, not a `Dialog`.
  - (4) Outlined logo: `ui/src/brand/backlog-logo.tsx` is the filled official Nulab icon used for the card, nav and topbars; `brand/backlog-logo.test.tsx:55` locks the fills `["#42CE9F", "white"]`. `docs/brand/backlog-logo.md` quotes Nulab terms that forbid modified or recoloured versions of the logo, so an outline redraw of the official paths is a brand-terms risk. Safer options: an original stroke-only `currentColor` icon that is not the Nulab mark, or a curated name (e.g. `ticket`), switched at the one `PLUGIN_ICON` line.
- **Risks / follow-up**:
  - Tests that will need updating: `index.test.ts` (nav/route counts and paths), `connected-panel.test.tsx:148-149` (navigate to `/backlog/watches`), `backlog-logo.test.tsx` (fills), `watches-page.test.tsx`, `dashboard-page.test.tsx`, and any test that clicks a raw control by tag. The harness `host.ui` stubs only `Button`/`Input`/`Label` as DOM tags plus `Skeleton`/`ChangeRequestDetail`; every newly used host component (`SettingsSection`, `Card*`, `Select*`, `Checkbox`, `Table*`, `Dialog*`, `Tabs*`, `IntegrationScopeBar`, `IntegrationListToolbar`…) needs a stub, and `variant`/`size` props will pass through to plain `<button>` in tests.
  - Keep behaviour locked by business rules: the settings card, switch and nav entry must not depend on the enabled switch (BR5.4/BR7.6/BR7.8, `index.ts:48-49`); `settingsHref()` format; the restore notice "Review watches" (`SettingsScreen.tsx:315-325`) must point to the new location if `/backlog/watches` goes away.
  - Permissions: `git.watches.*` and `git.queries.*` are `authenticated`, while Settings is usually an admin surface; `SettingsScreen` switches to a member view (`state.isMember`, set only after an admin action returns 403, `settings/state.ts:165-166`) that hides the connect form and the connected panel. Moving watches into Settings needs a decision on who can see and edit them; changing `access` is a manifest and contract change.
  - Do not rename action keys (Kandev only accepts `^[a-z0-9][a-z0-9._-]*$`); a new action must be added to `manifest.yaml` and `TestU*_Manifest*` tests.
  - Removing `/backlog/watches` and `/backlog/dashboard` breaks existing bookmarks; a kept route without a nav item is possible (`registerRoute` without `registerNavItem`).
  - `make coverage` still writes `coverage.out` at the repo root; reviewers should use the `build/` command form used here.
