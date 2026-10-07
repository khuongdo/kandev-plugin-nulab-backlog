## Developer Code Scan Results

Intent: `261007-backlog-panel-retouch` (refactor, Minimal depth). Focused scan, 2026-10-07, branch
`feature/refactor-backlog-pan-03d` at `5bf88b9` (v0.3.0). Snapshot set: `ui/src`, `internal/plugin`,
`internal/issues`, `manifest.yaml`.

### Scan Coverage
- **Analyzed deeply**:
  - ui/src/issues/issues-page.tsx
  - ui/src/issues/issues-state.ts
  - ui/src/issues/links-store.ts
  - ui/src/issues/issue-badge.tsx
  - ui/src/page/BacklogPage.tsx
  - ui/src/git/pr-toolbar.tsx
  - ui/src/host-ui.ts
  - ui/src/index.ts
  - ui/src/layout.ts
  - ui/src/testing/harness.ts (host UI stubs: ChangeRequestRow, ChangeRequestList, Select, IntegrationRepositoryFilter, navigate mock)
  - internal/issues/service.go (List, TaskLink, IssueItem, Link/Unlink, Links/LinkView, taskKey)
  - internal/issues/types.go (Link, IssueURL, NewTaskFor)
  - manifest.yaml (version, min_kandev_version, action keys)
- **Skimmed only**:
  - ui/src/issues/ (issue-panel.tsx, link-task-dialog.tsx, task-menu.ts, i18n.ts, poll-interval.tsx, tests)
  - ui/src/page/ (start-task.tsx, quick-actions.ts, tests)
  - ui/src/git/ (pr-list.tsx row/task-link rendering, save-query-dialog.tsx, others)
  - ui/src/settings/, ui/src/switch/, ui/src/brand/, ui/src/messages/en.ts
  - ui/package.json, ui/tsconfig.json, ui/vitest.config.ts
  - internal/plugin/ (issue_actions.go action keys, host_port.go ListTasks mapping)
  - internal/issues/ (store.go, sync.go, watch.go, watcher.go, queries.go, quick_actions.go, tests)
  - go.mod, Makefile (coverage target), .kandev-sdk-ref
  - Out of snapshot, reference only (not deep-analyzed): `~/repo/kandev` (Kandev v0.96.0 host source, see Handoff)

### Packages Found
- `ui` (kandev-plugin-nulab-backlog-ui) — plugin UI bundle — TypeScript/JSX (host React via `host.jsx`) — Backlog page, issue/PR lists, settings, card badge, task panel
- `internal/issues` — domain service — Go — issue list/filters, task links, sync, watches, saved queries, quick actions
- `internal/plugin` — adapter — Go — only package importing `pluginsdk`; action dispatch, HostPort implementation
- Others (shallow): `internal/backlog`, `internal/connection`, `internal/git`, `internal/redact`, `internal/pkgverify`, `internal/ci`, `internal/testutil`, `server`

### Build System
- **Type**: Go modules + Makefile; npm + esbuild for the UI bundle
- **Config Files**: go.mod (`replace github.com/kandev/kandev => ../kandev/apps/backend`), .kandev-sdk-ref, Makefile, ui/package.json, ui/tsconfig.json (path alias `@kandev/plugin-sdk` → `../../kandev/apps/packages/plugin-sdk/src/index.ts`), ui/vitest.config.ts, manifest.yaml (v0.3.0, `min_kandev_version: "0.96.0"`)
- **Build Dependencies**: `ui` → `@kandev/plugin-sdk` (types only, from the sibling kandev checkout); `server` → `internal/plugin` → `internal/issues` → `internal/backlog`, `internal/connection`

### APIs Discovered
- Plugin actions (manifest.yaml + internal/plugin/issue_actions.go) — 26 `issues.*` keys, 23 `git.*`/`repositories.*`, 9 `connection.*`. Relevant here: `issues.list` (rows with `linkedTasks`), `issues.filters` (projects/statuses/assignees), `issues.links.list`, `issues.link`, `issues.queries.*`.
- `issues.list` row shape (internal/issues/service.go:200-215): `IssueItem{issueKey, summary, status, statusId, assignee, updatedAt, url, linkedTasks[]}`; `TaskLink{taskId, taskKey?}` — **no task title**.
- Host APIs used by the issue list: `host.navigate(href)` (task open, `taskHref` = `/t/<id>` in issues-state.ts:149), `host.ui.ChangeRequestRow` / `ChangeRequestList`, `IntegrationScopeBar`, `Select*`, `Input`, `Pagination*`, `host.useResponsiveBreakpoint`, `host.utils.formatRelativeTime`.
- Host UI kit exposed at v0.96.0 but **not yet used** by the plugin: `TaskRowIndicator`, `IntegrationListToolbar`, `Combobox`, `Popover*`, `Drawer*`, `Sheet*` (SDK `PluginUIShape`, plugin-sdk/src/index.ts:546-695; web host-api.ts:358-367).

### Frameworks & Libraries
- Kandev plugin SDK — pinned `f099a46d` = tag v0.96.0 — host runtime, React, UI kit
- React (host-provided; devDep `react` ^19.3.0 for tests) — UI
- esbuild ^0.28.2 — bundle; vitest ^5.0.3 + jsdom ^30.1.2 — UI tests; axe-core ^4.14.0 — a11y checks
- TypeScript ~6.0.3, ESLint ^10.12.0 + typescript-eslint, Prettier ^3.9.9
- Go 1.26.0, testify v1.12.1, yaml.v3

### Test Coverage
- **Test Directories**: co-located `*_test.go` in internal/*; co-located `ui/src/**/*.test.ts(x)`; host stubs in ui/src/testing/harness.ts
- **Test Frameworks**: Go `testing` + testify (`-race`); Vitest (jsdom) with the fake host harness
- **Coverage Config**: present — `make coverage` (floor 80%, excludes `server/main.go`)
- **Baseline (2026-10-07, this worktree)**:
  - `go test -race ./internal/... ./server/...` — all 9 packages **ok** (run with a scratch `-modfile` whose `replace` points to `~/repo/kandev/apps/backend`, because `../kandev` is absent here). Coverage: `internal/issues` 91.5%, `internal/plugin` 93.8%.
  - `npx vitest run` (ui) — **31 files, 322 tests passed** (after `npm ci`; node_modules is gitignored).
  - `tsc --noEmit` — clean (via a scratch tsconfig pointing the SDK alias to `~/repo/kandev`); `eslint .` clean; `prettier --check .` clean.

### Code Quality Indicators
- **Linting**: golangci-lint + gosec (.golangci.yml), go vet, gofmt; ESLint (ui/eslint.config.js), Prettier, tsc strict
- **CI/CD**: .github/workflows/ci.yml, .github/workflows/release.yml
- **Documentation**: README present; doc comments on every exported Go name and every UI factory, with requirement IDs (FR/BR/US/AC)

### Technical Debt Signals
- The plugin re-implements host widgets instead of using them: `createPrToolbar` (ui/src/git/pr-toolbar.tsx) copies `IntegrationListToolbar`'s layout classes (layout.ts `TOOLBAR`), and both lists render linked tasks as raw `<a>` elements (issues-page.tsx:353-369, pr-list.tsx:292-302) instead of the host `TaskRowIndicator`.
- Linked task text is `taskKey ?? taskId` (issues-page.tsx:365) and `Task {id}` in the PR list — no title, no icon, no tooltip; easy to miss in the metadata line.
- Filter panel (issues-page.tsx:289-351): three labelled `Select`s (Project/Status/Assignee) plus a labelled search `Input` and a "Save query" button, all inside the toolbar; labels above controls make the toolbar tall compared with the GitHub toolbar (single row: title+count, one compact filter, query input, refresh). Mobile uses an inline "Filters (n)" toggle, not a Drawer/Sheet. Search fires after 400 ms instead of Enter-to-commit like the host toolbar. No visible "clear filters" control except on the empty state.
- `issues-page.tsx` is 637 lines with load, filters, rows, pagination, and dialogs in one closure — the largest UI file.
- Toolchain coupling: go.mod `replace` and tsconfig `paths` both assume a sibling `../kandev` checkout; in a Kandev task worktree it is missing, so `make` targets and `tsc` fail without a workaround.

## Handoff Summary
- **Intent-relevant finding**:
  1. *Clearly show linked tasks on the issue card / click opens task*: rows are host `ChangeRequestRow`s (issues-page.tsx:400-416) with `taskIndicator={tasksOf(item)}` rendering plain anchors labelled `taskKey ?? taskId` that call `host.navigate('/t/<id>')`. Kandev's GitHub issue list (`~/repo/kandev/apps/web/components/github/my-github/issue-list.tsx:79-87`) passes the host `TaskRowIndicator` (`tasks: {id, taskId, fallbackTitle}[]`, `testIdPrefix`, `emptyLabel?`), which shows the task title + workflow step with a checklist icon and tooltip, a "Tasks (n)" dropdown for several tasks, and navigates via `setActiveTask` + router. It is exposed to plugins as `host.ui.TaskRowIndicator` at v0.96.0 (= `min_kandev_version`), so no backend change is strictly required; the host resolves the title from its task store by id and falls back to `fallbackTitle` (only `taskKey` is available from `issues.list`; adding a `title` to `TaskLink` in internal/issues/service.go:200 is optional).
  2. *Click issue title opens the Backlog issue in the browser*: already wired — `ChangeRequestRow` is given `href={item.url}` (`https://<space>/view/<KEY>`, types.go:153) and the host renders it as an `<a target="_blank" rel="noopener noreferrer">` with an external-link icon (`components/integrations/change-request-list.tsx:66-77`); the harness stub mirrors this and issues-page.test.tsx:191 asserts it. The refactor may only need to verify this in the real host (e.g. desktop app opening external links) rather than change code.
  3. *Retouch the filter panel*: filter state is local to `IssuesPage` (`Filters{projectKey, statusIds, assignee}`, options from `issues.filters`); the GitHub reference is `IntegrationListToolbar` (`components/integrations/integration-list-toolbar.tsx`) with a single compact filter (`RepoFilterCombobox`/`IntegrationRepositoryFilter`) and an Enter-to-commit query input. `host.ui.IntegrationListToolbar` is available but takes `count:number`, `lastFetchedAt: Date|null`, `customQuery/committedQuery` and a single `filter` node, so adopting it changes the search behaviour (debounce → Enter) and the toolbar test ids.
- **Risks / follow-up**:
  - The test harness (ui/src/testing/harness.ts) has no stubs for `TaskRowIndicator`, `IntegrationListToolbar` or `Combobox`; any adoption needs new stubs, and existing tests keyed on `backlog-issue-task-<key>-<id>` (issues-page.test.tsx:206, 234), `backlog-issues-search`, `backlog-issues-project/status/assignee` and `rawControls` (BR5.1) checks will change. `TaskRowIndicator` only emits `${testIdPrefix}-single` / `-multi` ids.
  - `TaskRowIndicator` reads the host app store (`useTaskById`, `setActiveTask`); confirm in the packaged-host contract test / real host that it renders inside a plugin route, since the jsdom harness cannot prove it.
  - "Card issue" is ambiguous: it may also mean the Kanban card badge (issue-badge.tsx, `task-card-tags` slot), whose click currently toggles a status detail and does not open the issue URL. Requirements Analysis should confirm which card is meant.
  - The PR list (pr-list.tsx:292) has the same raw task-link pattern; keep both lists consistent or explicitly scope it out.
  - Environment: Go is at `~/go-sdk/bin/go` (not on PATH); `../kandev` is NOT present in this worktree. A Kandev checkout at the pinned SDK commit `f099a46dc7aab16f6ff5806cd29b2b480296303f` (tag v0.96.0) exists at `/home/k_do_webfrontier/repo/kandev` (also `/home/k_do_webfrontier/repo/kandev-min`). GitHub integration UI reference (read-only): `~/repo/kandev/apps/web/components/github/my-github/` and `~/repo/kandev/apps/web/components/integrations/`. Code Generation needs `../kandev` linked (or the same scratch `-modfile`/tsconfig workaround) for `make` targets and the contract test.
