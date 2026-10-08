# Developer Code Scan - 261008-fix-uiux-backlog

Focused scan for the intent: task badge on Home > Tasks rows, hiding the Home > Integrations entry while Backlog is OFF, and three Backlog settings fixes (Projects order, issue-watch empty message, duplicate Add watch button).

## Developer Code Scan Results

### Scan Coverage
- **Analyzed deeply**:
  - ui/src/index.ts
  - ui/src/index.test.ts
  - ui/src/issues/issue-badge.tsx
  - ui/src/issues/issues-state.ts
  - ui/src/issues/links-store.ts
  - ui/src/issues/i18n.ts
  - ui/src/switch/enabled-events.ts
  - ui/src/switch/integration-switch.tsx
  - ui/src/settings/SettingsScreen.tsx
  - ui/src/settings/issue-watches-section.tsx
  - ui/src/settings/section-parts.tsx
  - ui/src/settings/sections.test.tsx
  - ui/src/messages/en.ts
  - ui/src/host-ui.ts
  - ui/package.json, ui/vitest.config.ts, ui/tsconfig.json
  - internal/issues/service.go (Links, LinkView, newLink, Link, Detail)
  - internal/issues/types.go (Link, IssueURL)
  - internal/issues/sync.go (status refresh loop)
  - internal/plugin/issue_actions.go (issues.links.list handler)
  - manifest.yaml
- **Skimmed only**:
  - ui/src/settings/pr-watches-section.tsx (empty state and header action only)
  - ui/src/settings/source-control-section.tsx, ui/src/issues/issue-badge.test.tsx, ui/src/testing/harness.ts (structure only)
  - internal/plugin/runtime.go, internal/connection/service.go (opt-in gating only)
  - Makefile, go.mod, .kandev-sdk-ref, .github/workflows/
  - Kandev v0.96.0 reference (read-only, ~/repo/kandev, HEAD f099a46 = .kandev-sdk-ref): apps/web/app/tasks/rich-task-list-row.tsx, apps/web/components/task/task-row-plugin-slots.tsx, apps/web/components/github/pr-task-icon*.tsx, apps/web/components/integrations/integrations-menu.tsx, apps/web/lib/navigation/{core-destinations,plugin-destinations,resolve-destinations}.ts, apps/web/hooks/use-app-destinations.ts, apps/web/lib/plugins/{types,registry,host,host-api}.ts

### Packages Found
- ui (kandev-plugin-nulab-backlog-ui) - browser bundle - TypeScript/TSX (host React through `host.jsx`, no bundled React) - every Kandev screen of the plugin
- internal/plugin - Go - action routing, the only package importing pluginsdk
- internal/issues - Go - issue list, links, sync, watches, queries, quick actions
- (others under internal/ not in scope: backlog, connection, git, scm, github, gitlab, bitbucket, redact, pkgverify, ci, testutil)

### Build System
- **Type**: Go modules + Make; npm + esbuild for the UI
- **Config Files**: go.mod (replace `github.com/kandev/kandev => ../kandev/apps/backend`), .kandev-sdk-ref (f099a46...), Makefile (`check-sdk`, `test` = `go test -race` + `npx vitest run`, `lint` = golangci-lint + `tsc --noEmit` + eslint), ui/package.json (`build`, `test`, `typecheck`, `lint`, `format:check`), ui/tsconfig.json (path alias `@kandev/plugin-sdk` -> `../../kandev/apps/packages/plugin-sdk/src/index.ts`)
- **Build Dependencies**: server -> internal/plugin -> internal/issues -> internal/backlog, internal/connection; ui imports `@kandev/plugin-sdk` as types only (no runtime import, verified by grep)

### APIs Discovered
- Kandev plugin actions - manifest.yaml - 90+ workspace/task actions; relevant: `issues.links.list` (workspace, authenticated) returns `{ links: LinkView[] }`, `connection.get`, `connection.set_enabled`, `issues.watches.*`
- UI registrations - ui/src/index.ts - `registerIntegrationSettings` (Settings > Integrations card + switch), `registerNavItem({ id: "backlog", path: "/backlog", section: "integrations" })` (Home > Integrations entry), `registerRoute("/backlog")`, `registerComponent("task-card-tags", IssueBadge)` (Kanban card only), task menu action, task panel, repository/review providers

### Frameworks & Libraries
- Kandev plugin SDK - pinned v0.96.0 (min_kandev_version "0.96.0") - host API, registry, UI kit
- vitest ^5.0.3 + jsdom ^30.1.2 - UI tests
- typescript ~6.0.3, eslint ^10, prettier ^3.9, esbuild ^0.28.2
- Go stdlib + testify (per team practice)

### Test Coverage
- **Test Directories**: ui/src/**/*.test.{ts,tsx} (33 files, about 283 `it` cases), internal/**/*_test.go (about 704 `func Test`)
- **Test Frameworks**: Vitest (jsdom, `h` JSX factory) with the shared fake host in ui/src/testing/harness.ts; Go `testing` + testify, httptest fakes
- **Coverage Config**: present (Makefile `coverage` target, 80% floor on ./internal/... and ./server/..., profile under build/)
- **Baseline**: NOT RUN in this scan. Go is not installed in this environment (`go: command not found`), `../kandev` is absent next to the worktree (Makefile `check-sdk` and the go.mod replace need it; ~/repo/kandev is at the right ref f099a46), ui/node_modules is absent in the worktree, and the delegated-agent guard blocked running Vitest from a scratch copy. The main session must record the baseline.

### Code Quality Indicators
- **Linting**: golangci-lint + gosec (Makefile), `tsc --noEmit` strict, ESLint (ui/eslint.config.js), Prettier
- **CI/CD**: .github/workflows/ci.yml, .github/workflows/release.yml
- **Documentation**: doc comments on every exported Go symbol and every UI factory; comments cite BR/FR/AC ids. No TODO/FIXME in the scanned paths.

### Technical Debt Signals
- `watchesEmpty` ("No PR watches yet", ui/src/messages/en.ts:126) is shared by PR watches (pr-watches-section.tsx:177) and issue watches (issue-watches-section.tsx:156), which causes the wrong issue-watch empty message.
- The same `add(...)` button renders twice when a watch list is empty: inside the empty state (`backlog-issue-watches-empty-add`, issue-watches-section.tsx:157; `backlog-pr-watches-empty-add`, pr-watches-section.tsx:178) and as the SettingsSection header action (`...-add`, issue-watches-section.tsx:186, pr-watches-section.tsx:205).
- `Link` (internal/issues/types.go:177) stores no issue summary, so the badge has nothing to show on hover. Only `Detail` (live call) knows `Summary`.
- Kandev v0.96.0 has no way to read or observe the integration switch, so the plugin keeps its own bus (ui/src/switch/enabled-events.ts).

## Feature-by-feature findings

### 1. Backlog issue on Home > Tasks rows (like GitHub)
- The plugin registers its badge only in the `task-card-tags` slot (ui/src/index.ts, `registry.registerComponent("task-card-tags", ...)`). Kandev renders that slot only on the Kanban card (`kanban-card-content.tsx`) and the graph2 pipeline row.
- Home > Tasks (`/tasks`, apps/web/app/tasks/rich-task-list-row.tsx) renders first-party `PRTaskIcon` / `MRTaskIcon` / `RegisteredChangeRequestTaskIcon` and then `<TaskRowMetadata surface="task-list">`, which mounts the plugin slot `task-row-metadata` (apps/web/components/task/task-row-plugin-slots.tsx). The sidebar task list uses the same slot with `surface: "sidebar"` (components/task/task-item.tsx).
- Slot props: `TaskRowMetadataSlotProps = { taskId, workspaceId, workflowStepId, surface: "sidebar" | "task-list" }` (lib/plugins/types.ts:181), the same task/workspace fields the current badge reads. The existing `IssueBadge` component and the shared `LinksStore` (one `issues.links.list` per workspace) can be registered for `task-row-metadata` as well.
- GitHub reference: `PRTaskIcon` is a small glyph with a hover/focus `Tooltip` (`PRTaskIconTooltip`, pr-task-icon-disclosure.tsx:123) and stops pointer propagation so the row does not open. The host UI kit exposes `Tooltip`, `TooltipContent`, `TooltipTrigger`, `TooltipProvider` and `Popover*` to plugins (lib/plugins/host-api.ts:81-109).
- Today the badge already opens `https://<space>/view/<KEY>` in a new tab when openable (`badgeHref`, issues-state.ts:171, with the backlog.com/.jp/backlogtool.com https re-check), and its hover text is only `title={detail}` (updated time, stale hint), not the summary.
- To show the issue summary on hover, the backend must carry it: add `Summary` to `Link` (set in `newLink`, service.go:455, refresh in the sync loop, sync.go:~300 where `LastKnownStatus` is updated) and to `LinkView` (service.go:612, `Links` at :626), then to the TS `LinkView` (issues-state.ts:34). Older stored links have no summary until the next sync; the UI must fall back to the key.

### 2. Hide Backlog in Home > Integrations when OFF
- The entry is `registry.registerNavItem({ id: "backlog", section: "integrations", path: "/backlog" })`, registered unconditionally in `initialize` (ui/src/index.ts). The comment and tests state this on purpose: "These registrations never depend on the switch" (BR5.4, BR7.6, BR7.8); ui/src/index.test.ts asserts "registers the entry and the route even when Backlog is off everywhere".
- Kandev v0.96.0 cannot hide a plugin nav item:
  - `NavItem` has only `id, label, path, icon, section` (lib/plugins/types.ts:35); no `requires`/visibility field.
  - Plugin destinations are never availability-gated: `pluginDestinations` sets no `requires` (lib/navigation/plugin-destinations.ts), and `resolveDestinations` only filters on `requires` (resolve-destinations.ts:28). First-party GitHub/GitLab/Jira entries use `requires: "<key>"` (core-destinations.ts).
  - There is no `unregisterNavItem`; removal is only `unregisterPlugin`.
  - Late registration does not work: `stagedGenerationRegistry` (lib/plugins/host.ts:129) only queues calls while the generation is open, which ends when `initialize` returns.
  - `setIntegrationEnabled` / `isIntegrationEnabled` exist in the registry but are used only for the Settings sidebar badge (components/app-sidebar/sections/settings/integration-enabled.tsx:75), not for nav destinations.
  - Kandev v0.97.0 has no change in apps/web/lib/navigation for this.
- Options for the architect: (a) await `connection.get` for the workspaces inside `initialize` and register the nav item only if Backlog is ON in at least one workspace (global, not per workspace; a later toggle needs a page reload; adds startup latency inside the host's initialize timeout); (b) drop the nav item and reach /backlog from the Settings card; (c) an upstream Kandev change (gate plugin nav items whose id matches an owned integration on `isIntegrationEnabled` for the active workspace), which would raise `min_kandev_version`. The /backlog page already handles OFF state through `integration_disabled` errors.

### 3. Backlog settings: Projects right below the sign-in method
- SettingsScreen.tsx renders sections in this order (lines 388-439): connection, pr-watches, issue-watches, saved-queries, quick-actions, issue-sync, source-control, projects. The sign-in method dropdown (`backlog-method`) is inside the Connection section form; the project picker is the last section (`backlog-section-projects`, admin only, keyed by space host and user name).
- Moving the `projects` block to right after `connection` is a JSX reorder. ui/src/settings/sections.test.tsx:24 `SECTIONS` asserts the exact order and must change with it. Projects only shows when connected and not a member; the form (method dropdown) shows only when not connected or replacing, so "below the method" means below the Connection section.

### 4. Issue watch empty message
- issue-watches-section.tsx:156 uses `messages.watchesEmpty` = "No PR watches yet". Needs a new catalogue key (for example `issueWatchesEmpty`) in ui/src/messages/en.ts; `messagesFor` falls back to English for every key, so no other catalogue exists. sections.test.tsx:374 asserts `en.watchesEmpty` for the issue list and must change.

### 5. Remove the per-list Add watch button
- The "per-watcher" button is the empty-state child `add("backlog-issue-watches-empty-add")` (issue-watches-section.tsx:157). PR watches has the same duplicate (`backlog-pr-watches-empty-add`, pr-watches-section.tsx:178). `ListEmpty` already renders without children (section-parts.tsx:64). No test references the `-empty-add` test ids.

## Handoff Summary
- **Intent-relevant finding**: Home > Tasks renders plugins only through the `task-row-metadata` slot (Kandev apps/web/components/task/task-row-plugin-slots.tsx), while the plugin registers its issue badge only for `task-card-tags` (ui/src/index.ts); registering the existing `IssueBadge`/`LinksStore` for `task-row-metadata` is the main change, and the hover summary needs `Summary` added to `Link`/`LinkView` (internal/issues/types.go:177, service.go:612) because links store no summary today. Settings fixes are local: reorder JSX in SettingsScreen.tsx:388-439, new empty-message key for issue watches (issue-watches-section.tsx:156, en.ts:126), drop the empty-state add buttons (issue-watches-section.tsx:157, pr-watches-section.tsx:178).
- **Risks / follow-up**:
  - Kandev v0.96.0 offers no way for a plugin to hide or gate its Home > Integrations nav item per workspace (no `requires`, no unregister, no late registration). The requirement needs a decision: register-only-if-ON at load (needs reload after a toggle), remove the entry, or an upstream Kandev change plus a `min_kandev_version` bump.
  - Hiding the entry contradicts existing business rules BR7.6/BR5.4 and the test "registers the entry and the route even when Backlog is off everywhere" (ui/src/index.test.ts); requirements must supersede them explicitly. The Settings > Integrations card must stay registered so Backlog can be turned back on.
  - `task-row-metadata` is also shown in the sidebar task list (`surface: "sidebar"`); decide whether the badge shows there too or only for `surface === "task-list"`.
  - Adding `Summary` to the stored `Link` touches the persisted link JSON (backward compatible with `omitempty`), the sync loop, and leak tests (internal/issues/leak_test.go) that check no Backlog content leaks into errors.
  - Test baseline was not recorded: Go is not installed, `../kandev` is missing (use ~/repo/kandev, already at f099a46), ui/node_modules is missing in the worktree, and the guard blocks delegated test runs.
