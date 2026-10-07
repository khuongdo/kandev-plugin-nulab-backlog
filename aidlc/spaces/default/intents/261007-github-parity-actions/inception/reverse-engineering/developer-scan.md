# Developer Code Scan — 261007-github-parity-actions

- Scan date: 2026-10-07
- Commit: `2b4325f` (branch `feature/add-default-queries-87j`)
- Scan type: FOCUSED (snapshot paths `ui/src/`, `internal/plugin/`, `internal/git/`, `internal/issues/`, `manifest.yaml`)
- Prior code knowledge base: `aidlc/spaces/default/codekb/kandev-plugin-nulab-backlog/` recorded at commit `86ae473` (intent `261007-uiux-github-style`). It is STALE for this focus area: commit `9a1c718` (PR #4, "GitHub-style UI, issue watches and PR list, version 0.1.1") changed 71 files in the focus area (+5842 / -1419). See "Changes since the prior code knowledge base" below.

## Developer Code Scan Results

### Scan Coverage
- **Analyzed deeply**:
  - `manifest.yaml`
  - `ui/src/index.ts`
  - `ui/src/layout.ts`
  - `ui/src/host-ui.ts`
  - `ui/src/icons.tsx`
  - `ui/src/jsx.d.ts`
  - `ui/src/page/BacklogPage.tsx`
  - `ui/src/issues/issues-page.tsx`
  - `ui/src/issues/issues-state.ts`
  - `ui/src/issues/task-menu.ts`
  - `ui/src/issues/i18n.ts`
  - `ui/src/git/pr-list.tsx`
  - `ui/src/git/pr-toolbar.tsx`
  - `ui/src/git/save-query-dialog.tsx`
  - `ui/src/git/git-state.ts`
  - `ui/src/settings/saved-queries-section.tsx`
  - `ui/src/settings/section-parts.tsx`
  - `ui/src/settings/use-list.ts`
  - `ui/src/settings/SettingsScreen.tsx` (section order and render tree)
  - `internal/plugin/runtime.go` (action table, Runtime struct)
  - `internal/plugin/issue_actions.go`
  - `internal/plugin/git_actions.go` (action keys)
  - `internal/plugin/host_port.go`
  - `internal/issues/types.go` (Query, NewTask, NewTaskFor)
  - `internal/issues/service.go` (CreateTask, createLinkedTask)
  - `internal/issues/store.go`
  - `internal/issues/watch.go` (IssueWatchInput, constants)
  - `internal/git/types.go` (QueryInput and validation)
  - `internal/git/service.go` (ListQueries, SaveQuery, DeleteQuery, RunQuery)
  - `internal/git/store.go` (state keys and limits)
  - `internal/git/watcher.go` (review task creation)
- **Skimmed only**:
  - `ui/src/settings/` (other sections: connected-panel, issue-watch-dialog, issue-watches-section, pr-watches-section, project-picker, confirm-dialog, oauth, state)
  - `ui/src/git/` (watch-form, review-provider, repository-provider, pr-link, create-pr, git-access)
  - `ui/src/issues/` (issue-panel, issue-badge, link-task-dialog, links-store, poll-interval)
  - `ui/src/messages/en.ts`, `ui/src/testing/harness.ts`, `ui/src/switch/`, `ui/src/brand/`
  - `internal/issues/` (sync.go, watcher.go, watch_store.go, events.go)
  - `internal/git/` (prs.go, events.go, host.go)
  - `internal/plugin/` (credential.go, webhook.go, references.go, events.go, config.go)
  - `internal/backlog/`, `internal/connection/`, `Makefile`, `.github/workflows/` (outside the snapshot; read only for facts)

### Packages Found
- `internal/plugin` — library (the only Go package that imports `pluginsdk`) — Go — Kandev adapter: action table (`handlers` map, merged from `issueHandlers` and git handlers via `init()`), host port adapters (`hostPort`, `issueHost`), webhook, events, credential.
- `internal/issues` — library — Go — Backlog issues: list with filters, create task from issue, links, sync, issue watches and watcher. Own state store (`issues.links`, `issues.settings`, `issues.index`, plus watch documents).
- `internal/git` — library — Go — Backlog Git: repository provider, PR links, PR create, PR watches and watcher, saved PR queries (`git.queries`), PR list (`git.prs.list`).
- `ui` (`kandev-plugin-nulab-backlog-ui`) — application bundle — TypeScript/JSX (no React bundled; host `host.jsx` factory and `host.React` hooks) — registers the integration settings, the `/backlog` nav item and route, repository provider, review provider, task action, task-card badge, task menu action and task panel.
  - `ui/src/page/` — the `/backlog` route (Issues and Pull requests tabs).
  - `ui/src/issues/` — issue list, task menu (Unlink), issue panel, badge, links store, i18n.
  - `ui/src/git/` — PR list, PR toolbar, save query dialog, watch form, providers.
  - `ui/src/settings/` — the integration settings screen and its sections.
  - `ui/src/messages/en.ts` — the single English message catalogue (316 lines).
  - `ui/src/testing/harness.ts` — fake host for Vitest (host UI kit fakes).

### Build System
- **Type**: Go modules (Go 1.26.x) + npm (esbuild bundle, Node per `.nvmrc`), orchestrated by `Makefile`.
- **Config Files**: `go.mod` (`replace github.com/kandev/kandev => ../kandev/apps/backend`), `.kandev-sdk-ref` (`f099a46dc7aab16f6ff5806cd29b2b480296303f` = Kandev `v0.96.0`), `Makefile` (targets `check-sdk check-format vet lint test coverage check-secrets ui-build build package verify-package contract-test release-preflight marketplace-entry clean`), `ui/package.json`, `ui/tsconfig.json`, `ui/vitest.config.ts`, `ui/eslint.config.js`, `ui/.prettierrc`, `.golangci.yml`, `manifest.yaml`.
- **Build Dependencies**: `server/main.go` → `internal/plugin` → `internal/{issues,git,connection,backlog,redact}`; `internal/issues` and `internal/git` → `internal/connection`, `internal/backlog`. UI: `ui/src/index.ts` → `page/BacklogPage.tsx` → `issues/issues-page.tsx`, `git/pr-list.tsx` → `git/pr-toolbar.tsx`, `git/save-query-dialog.tsx`; `settings/SettingsScreen.tsx` → `settings/saved-queries-section.tsx` → `git/save-query-dialog.tsx`. Bundle: `esbuild src/index.ts --bundle --format=esm --jsx-factory=h --outfile=../build/ui/bundle.js`.

### APIs Discovered
- Plugin actions (Go, declared in `manifest.yaml`, dispatched by `internal/plugin/runtime.go` `handlers`) — 48 actions in total. Relevant to this intent:
  - Saved PR queries: `git.queries.list`, `git.queries.save`, `git.queries.delete`, `git.queries.run` (workspace scope, `authenticated`).
  - PR list: `git.prs.list` (workspace scope).
  - Issues: `issues.list`, `issues.filters`, `issues.create_task`, `issues.link`, `issues.unlink`, `issues.links.list`, `issues.refresh`.
  - There is NO action for saved issue queries, NO action for quick actions (task prompt presets), and NO "default query" flag on saved queries.
- UI extension points (host registry in `ui/src/index.ts`) — 9 registrations: `registerTranslations`, `registerIntegrationSettings`, `registerNavItem` (`/backlog`, section `integrations`), `registerRoute("/backlog", …, { topbar: { title, icon } })`, `registerRepositoryProvider`, `registerTaskAction` (PR link), `registerReviewProvider`, `registerComponent("task-card-tags", …)`, `registerTaskMenuAction` (Unlink only), `registerTaskPanel`.
- Host APIs used by the UI: `host.api.invokeAction`, `host.context.getActiveWorkspaceId/subscribeActiveWorkspace/getWorkspaceIds/subscribeWorkspaces/getTaskCreationContext`, `host.navigate`, `host.toast`, `host.i18n.t`, `host.useResponsiveBreakpoint`, `host.utils.formatRelativeTime`, `host.ui.*` (Tabs, Table, Select, Dialog, DropdownMenu, Pagination, Empty, Alert, Skeleton, Checkbox, Label, Input, Button, Card, SettingsSection, ChangeRequestList, ChangeRequestRow, IntegrationIcon, IntegrationRepositoryFilter).
- Host UI kit components available at the minimum Kandev version (`v0.96.0`, `apps/web/lib/plugins/host-api.ts`) but NOT used by the plugin today: `IntegrationStartTaskMenu`, `IntegrationListToolbar`, `IntegrationScopeBar`, `IntegrationSaveQueryDialog`, `TaskCreateDialog`, `IntegrationCursorPagination`, `TaskRowIndicator`, `Combobox`, `Textarea`, `PageTopbar`.
- Host Go data API used: `Tasks().Create` (`pluginsdk.CreateTaskInput` with `Title`, `Description`, `Priority`, `Metadata`, `WorkflowStepID`) and `Tasks().List`. `CreateTaskInput` at v0.96.0 also offers `StartAgent bool` and `Launch *PluginTaskLaunchOptions{AgentProfileID, ExecutorProfileID, Prompt, PlanMode}` (`apps/backend/pkg/pluginsdk/data_types.go:1120-1136, 1264-1269`); the plugin uses neither.

#### How saved queries are stored and rendered today
- Storage (Go): `internal/git/store.go` key `git.queries` (workspace scope, `{schemaVersion, items}` document, max 50). `Query = QueryInput{ID, Name, ProjectKey, RepoName, Statuses, Assignee, Creator}` (`internal/git/types.go:179-215`). `Validate` requires a non-empty name (≤100 runes), a repo in a SELECTED project, at least one status, assignee/creator in `anyone|me`. `SaveQuery` creates (no id, `newID()`) or replaces; no default flag, no ordering field.
- Rendering (UI): `ui/src/git/pr-list.tsx` loads `git.queries.list` once per workspace and shows the queries as a "Query" `Select` preset among the filter controls (`applyPreset` copies repo/statuses/assignee/creator into filters; queries of a non-selected project are disabled). The list starts from `START = { repo: "", statuses: ["open"], assignee: "anyone", creator: "anyone" }`, so the PR list is EMPTY ("choose a repository") until the user picks a repository or a query; nothing is applied automatically on open. "Save query" opens `ui/src/git/save-query-dialog.tsx` (plugin-built Dialog, name only, `git.queries.save`). Settings lists the same queries (`ui/src/settings/saved-queries-section.tsx`) with Edit (rename) and Delete; new queries can only be saved from the PR list.
- Issues: `ui/src/issues/issues-page.tsx` has search (400 ms debounce), Project/Status/Assignee `Select` filters (`issues.filters`), 20 rows per page, Refresh; no saved queries and no presets. Filter state is in component memory only. Assignee filter takes numeric ids (`issues.Query.AssigneeIDs`); there is no "me" value for `issues.list` (the issue watch does have `assignee: anyone|me`).

#### How tasks are created from issues and PRs today
- Issues: row menu (`RowMenu`, "..." icon) → "Create task" → `issues.create_task {issueKey, workflowId, workflowStepId, force?}` using `host.context.getTaskCreationContext(workspaceId)`. Go `NewTaskFor` (`internal/issues/types.go:151-160`) builds Title = issue summary, Description = issue description + "Backlog: <url>", Priority from Backlog priority. No prompt template, no agent launch (`StartAgent` unset), no user choice of preset. A duplicate create triggers the "already linked" confirm dialog. "Link to task" opens `link-task-dialog.tsx`.
- PRs: the PR list rows (`ChangeRequestRow`) have NO start-task action (`action` prop unused). Tasks for PRs come only from PR watches (`internal/git/watcher.go:316-320`, Title `Review PR #N: <summary>`, Description = PR URL) and from linking via the task action.
- Task menu (`ui/src/issues/task-menu.ts`): only "Unlink Backlog issue" (`group: "primary"`). No prompt templates anywhere in the UI or Go code (grep for `prompt` finds none).

#### Page layout and toolbar code
- `ui/src/layout.ts`: the plugin ships no CSS; four class constants: `STACK = "flex flex-col gap-4"`, `FIELD = "flex flex-col gap-2"`, `ROW = "flex flex-wrap items-center gap-2"`, `BUTTON = "cursor-pointer"`.
- `ui/src/page/BacklogPage.tsx`: root `<div className={STACK}>` with NO horizontal or vertical padding. Kandev renders a plugin route inside `PageShell` with `contentClassName="bg-background"` and no padding (`apps/web/components/plugins/plugin-page.tsx:68-78`), so the Backlog content runs edge to edge. Scope switch is a host `Tabs` (`TabsList` with Issues / Pull requests triggers), URL `?scope=prs`, each scope stays mounted once opened.
- `ui/src/git/pr-toolbar.tsx`: plugin-built imitation of `IntegrationListToolbar` (title `text-lg` + count left; last-updated + ghost refresh right; filters in a `ROW` under it), no border, no `px-4 sm:px-6`, no search input.
- `ui/src/issues/issues-page.tsx`: toolbar is a `ROW` with a labelled Search input, a mobile "Filters (n)" button and an icon Refresh; filters are labelled `Select`s in a second `ROW`; a separate `<h2>` "Issues" heading; data in a host `Table` (desktop) or card list (mobile); per-row actions in a "..." `RowMenu` in the last column; pagination as Previous/Next buttons.

### Frameworks & Libraries
- Go 1.26.8 (toolchain at `~/.local/go/bin`), `github.com/kandev/kandev/pkg/pluginsdk` at v0.96.0 (local `replace`), `github.com/stretchr/testify` (tests).
- Kandev plugin SDK (TypeScript types `@kandev/plugin-sdk`, `apps/packages/plugin-sdk/src/index.ts` in the Kandev checkout) — v0.96.0 — host registry, `PluginHostApi`, `PluginUIShape`.
- Host-provided React 19 (via `host.React`, not bundled); `@types/react` ^19.3.0, `react`/`react-dom` ^19.3.0 dev-only for tests.
- esbuild ^0.28.2 (bundle), TypeScript ~6.0.3 (`strict`), Vitest ^5.0.3 + jsdom ^30.1.2, axe-core ^4.14.0 (accessibility checks in tests), ESLint ^10.12.0 + typescript-eslint ^8.71.1, Prettier ^3.9.9.

### Test Coverage
- **Test Directories**: Go tests beside sources (`internal/*/…_test.go`) with `testdata/` JSON for `internal/backlog`; UI tests beside sources (`ui/src/**/*.test.ts(x)`), fake host in `ui/src/testing/harness.ts`.
- **Test Frameworks**: Go `testing` + testify `require`, `-race`; Vitest + jsdom + axe-core.
- **Coverage Config**: present — `make coverage` (writes `build/coverage.out`, floor 80%, exclusions listed in `Makefile`).
- **Baseline (this scan, 2026-10-07)**: `go test -race -cover ./internal/... ./server/...` all 9 packages pass — backlog 96.1%, ci 91.0%, connection 94.6%, git 90.8%, issues 91.3%, pkgverify 93.2%, plugin 93.7%, redact 97.4%, testutil 88.0%; total 92.7% (profile written to the scratchpad, not the repo root). UI: Vitest 29 files / 286 tests pass; `tsc --noEmit`, ESLint and Prettier clean. (`../kandev` was linked to `/home/k_do_webfrontier/repo/kandev` at `v0.96.0` for this run.)
- Focus-area tests to extend: `ui/src/page/backlog-lists.test.tsx` (PR list, presets, save query), `ui/src/issues/issues-page.test.tsx`, `ui/src/settings/sections.test.tsx` (saved queries section), `ui/src/index.test.ts` (registrations), `internal/git/queries_test.go`, `internal/issues/create_test.go`, `internal/plugin/actions_u3_test.go` / `actions_u4_test.go`, `internal/plugin/manifest_test.go` (asserts action keys and access).
- The fake host (`ui/src/testing/harness.ts`) fakes only the host components the plugin uses today (e.g. `IntegrationRepositoryFilter`, `ChangeRequestList`, `ChangeRequestRow` without `action`, `IntegrationIcon`); any newly used host component (`IntegrationStartTaskMenu`, `IntegrationScopeBar`, `IntegrationListToolbar`, `TaskCreateDialog`, `Textarea`) needs a fake there.

### Code Quality Indicators
- **Linting**: Go `gofmt`, `go vet`, `golangci-lint` + `gosec` (`.golangci.yml`); UI `tsc --noEmit` strict, ESLint (`ui/eslint.config.js`), Prettier (`ui/.prettierrc`). All clean at this commit.
- **CI/CD**: `.github/workflows/ci.yml`, `.github/workflows/release.yml` (calls the Makefile targets; packaged-host contract test on the minimum Kandev version).
- **Documentation**: every exported Go name and most UI factories carry doc comments that cite requirement ids (BR/FR/AC/US). `README.md` present. Comments explain host quirks (e.g. metadata namespace, async host injection).

### Technical Debt Signals
- `ui/src/git/pr-toolbar.tsx` re-implements the host `IntegrationListToolbar` layout instead of using it (host version has `border-b px-4 py-2.5 sm:px-6`, search input, mobile status row). The comment says it omits the search because Backlog's PR API has none; the host component requires `customQuery` props, so reuse would need a hidden or keyword-only query.
- `ui/src/git/save-query-dialog.tsx` is a plugin-built dialog parallel to the host `IntegrationSaveQueryDialog`.
- `ui/src/page/BacklogPage.tsx` uses plain `Tabs` for the Issues/PRs switch, while the GitHub page uses the scope bar (`IntegrationScopeBar`: kind segment + preset pills + Saved menu). No page padding (see layout).
- Issue list and PR list use different row components (`Table` vs `ChangeRequestRow`) and different toolbars (`issues-page` own `ROW` vs `PrToolbar`): inconsistent with each other and with GitHub, which uses `ChangeRequestRow` + `IntegrationListToolbar` for both kinds.
- `internal/issues/store.go` copies the document helpers from `internal/git/store.go` (`ponytail:` note: extract when a third package needs them) — a new store for quick actions or issue queries would be the third user.
- `ui/src/git/git-state.ts:89` `ponytail:` repository options load only the first page (50 repositories).
- `internal/plugin/host_port.go:119-125` `ponytail:` gRPC code read by text match.
- `internal/issues/service.go:381` `ponytail:` no automatic relink when a link write fails after create.
- Large files: `ui/src/issues/issues-page.tsx` (625 lines), `ui/src/settings/SettingsScreen.tsx` (441), `ui/src/git/pr-list.tsx` (386), `internal/git/service.go` (925), `internal/issues/service.go` (893).
- No TODO/FIXME/HACK markers and no `eslint-disable` in `ui/src`.

### Changes since the prior code knowledge base (commit `86ae473` → `2b4325f`)
- Removed: `ui/src/git/dashboard-page.tsx`, `ui/src/git/watches-page.tsx` (and tests).
- Added: `ui/src/git/pr-list.tsx`, `pr-toolbar.tsx`, `save-query-dialog.tsx`, `ui/src/icons.tsx`, `ui/src/layout.ts`, `ui/src/host-ui.ts`, `ui/src/settings/{issue-watch-dialog,issue-watches-section,pr-watches-section,saved-queries-section,section-parts,use-list}.ts(x)`, `internal/issues/{watch,watch_store,watcher}.go`; actions `git.prs.list` and `issues.watches.*` (6) in `manifest.yaml`.
- Changed: `BacklogPage.tsx` (Issues/PRs tabs on `/backlog`), `issues-page.tsx` (row menu, icon refresh), `SettingsScreen.tsx` (sections), `index.ts` (one nav entry and route), `messages/en.ts`, `testing/harness.ts` (+335 lines of host fakes), `internal/plugin/{runtime,issue_actions,git_actions,host_port}.go`.

## Reference: Kandev GitHub integration

Read-only reference: `/home/k_do_webfrontier/repo/kandev`, tag `v0.96.0` (commit `f099a46dc`), which is also this plugin's `min_kandev_version`. Paths below are under `apps/web/` unless stated.

### Quick actions (task prompt presets)
- Defaults: `components/github/my-github/action-presets.ts`.
  - `DEFAULT_ISSUE_PRESETS`: `implement` "Implement" (hint "Build and open a PR", icon `code`, prompt `Implement the changes described in the GitHub issue at {{url}} (title: "{{title}}"). Open a pull request when complete.`), `investigate` "Investigate" (hint "Find the root cause", icon `search`, prompt `Investigate the GitHub issue at {{url}} (title: "{{title}}"). Identify root cause and summarize findings.`), `reproduce` "Reproduce" (hint "Document repro steps", icon `bug`, prompt `Reproduce the bug described in the GitHub issue at {{url}} (title: "{{title}}"). Document the reproduction steps.`).
  - `DEFAULT_PR_PRESETS`: `review` "Review" (icon `eye`), `address_feedback` "Address feedback" (icon `message`), `fix_ci` "Fix CI" (icon `tool`), each with a `{{url}}` prompt.
  - Type `GitHubActionPreset = { id, label, hint, icon, prompt_template }` (`lib/types/github.ts:715-723`). Icons: `eye message tool code search bug sparkle check` (`components/integrations/integration-preset-icons.ts`).
  - `interpolatePromptTemplate` replaces `{{url}}`/`{{title}}` (also legacy `{url}`/`{title}`); unknown placeholders stay. Stored presets fall back to the defaults when the stored list is empty (`resolveIssuePresets`, `resolvePRPresets`). Defaults are deliberately NOT translated (persisted seed).
- Customization UI: `components/github/action-presets-section.tsx` — Settings section "Quick actions" with a Reset button in the section action, a card with Tabs "Pull requests" / "Issues", a `PresetEditor` per kind (rows: icon `Select`, label, hint, prompt template; expand/collapse; remove; "Add PR action" / "Add issue action"; `newPreset()` = label "New action", icon `sparkle`, empty prompt). Saved through `useSettingsSaveContributor` (the settings page Save/Discard bar), stored in GitHub workspace settings.
- Launch flow: row action `IntegrationStartTaskMenu` (`components/integrations/integration-start-task-menu.tsx`): outline Button "+ Task ▾" (`aria-label`, `triggerTestId`), `DropdownMenuContent align="end" className="w-52"`, one item per preset with icon, label and hint. Selecting a preset opens `QuickTaskLauncher` (`components/github/my-github/quick-task-launcher.tsx`), which opens the native `TaskCreateDialog` prefilled with title `"<preset label>: <item title>"` (truncated) and description = interpolated prompt (plus repository/branch hints), then links the created task to the issue/PR. The plugin kit exposes `IntegrationStartTaskMenu` (props `presets, onSelect, triggerLabel?, triggerAriaLabel?, triggerTestId?, itemTestId?`; preset `{id, label, hint, iconName?}`) and `TaskCreateDialog` (props `open, onOpenChange, workspaceId, workflowId, defaultStepId, steps, initialValues{title, description, …}, onSuccess(task)`) — `lib/plugins/host-api.ts:350-361`, types in `components/task-create-dialog-types.ts:62-106, 198-221`.
- Row placement: `ChangeRequestRow` `action` slot (`components/integrations/change-request-list.tsx:38-58`), used by `components/github/my-github/issue-list.tsx:89-95` (and the PR list).

### Default queries (issues and PRs)
- Built-in presets shown as pills (`components/github/my-github/search-bar.tsx:31-99`): PRs — "Review requested" (`review-requested:@me is:open`, group `inbox`), "Mentions", "Open" (`author:@me is:open`, group `created`), "Drafts", "Recently merged"; Issues — "Assigned" (`assignee:@me is:open`, `inbox`), "Mentions", "Open" (`author:@me is:open`, `created`), "Recently closed". The first preset of a kind is selected when the kind changes (`IntegrationScopeBar` default `onKindChange`), so a page always opens on a working query.
- Customizable defaults: `components/github/my-github/use-default-query-presets.ts` stores `{pr, issue}` preset lists per workspace (or per user) and falls back to the built-ins; `reset` restores them.
- Saved queries with a default: `components/github/my-github/saved-preset-model.ts` (`SavedPreset{id, kind, label, customQuery, repoFilter, createdAt, isDefault}`; at most one `isDefault` per kind; `findDefaultSavedPreset`, `setSavedPresetDefault`). The star toggle is `components/integrations/saved-query-default-button.tsx`; the Saved menu in `IntegrationScopeBar` takes `onToggleSavedDefault` and `defaultMutationPendingId`. The default saved query is opened on page load instead of the first built-in preset.
- Save dialog: `components/integrations/integration-save-query-dialog.tsx` (`open, onOpenChange, title?, description, suggestedLabel, query?, repositoryId?, repositoryOptions?, onSave(label, repositoryId)`), exposed to plugins as `IntegrationSaveQueryDialog`.

### Page layout (toolbar, buttons, margins)
- Page: `app/github/github-page-client.tsx:371-440` — column `flex-1 flex flex-col min-w-0 overflow-hidden`:
  1. Scope bar `PresetsScopeBar` → `IntegrationScopeBar` (`components/integrations/presets-scope-bar-base.tsx:375-460`): `flex items-center gap-1.5 overflow-x-auto px-4 py-2 sm:px-6`; kind segment (Pull requests / Issues), divider, inbox pills, divider, created pills, and the Saved menu pushed right (`ml-auto shrink-0 pl-2`). Hidden on mobile (`hidden md:flex`), replaced by `MobileViewsPicker` in the toolbar title.
  2. Toolbar `ListToolbar` → `IntegrationListToolbar` (`components/integrations/integration-list-toolbar.tsx`): `flex shrink-0 flex-col gap-2 border-b px-4 py-2.5 sm:px-6 md:flex-row md:flex-wrap md:items-center md:gap-3`; title `h2 text-sm font-semibold` + count; repository filter (`md:w-[220px]`); query input `md:min-w-[240px] md:flex-1` with "Press Enter" hint; last-updated text + ghost icon refresh at the right; a mobile status row.
  3. Results: `flex-1 overflow-auto px-3 py-4 md:px-6` with `ChangeRequestRow` rows (`flex items-start gap-3 px-4 py-3 hover:bg-muted/40`, start-task menu in the `action` slot at the row end).
  4. Pagination footer `ResultsPagination`.
- Plugin routes render inside `PageShell` with no content padding (`components/plugins/plugin-page.tsx:68-78`), so a plugin must add the same `px-4 sm:px-6` / `px-3 md:px-6` itself.

## Handoff Summary
- **Intent-relevant finding**: All three intent items have no existing implementation, but the host already exposes the GitHub building blocks at the minimum Kandev version. (1) Quick actions: no prompt templates exist; issue tasks are created server-side by `issues.create_task` with `NewTaskFor` (title = summary, description = Backlog description + link; `internal/issues/types.go:151-160`), from a "..." row menu (`ui/src/issues/issues-page.tsx` `actionsOf`); PR rows have no start-task action (`ui/src/git/pr-list.tsx` `row`, `ChangeRequestRow` `action` unused). `host.ui.IntegrationStartTaskMenu` and `host.ui.TaskCreateDialog` exist, and Go `CreateTaskInput.StartAgent`/`Launch.Prompt` exist, unused. (2) Default queries: only PR saved queries exist (`git.queries`, `internal/git/types.go:179-215`; no `isDefault`, repo REQUIRED and must be in a selected project); the PR list opens empty until a repo/query is chosen (`START` in `ui/src/git/pr-list.tsx`); issues have no saved queries at all. (3) Layout: `/backlog` content has no padding (`ui/src/page/BacklogPage.tsx` root `STACK`, host `PageShell` adds none), uses `Tabs` instead of the scope bar, and a hand-made `PrToolbar` instead of `IntegrationListToolbar` (`px-4 py-2.5 sm:px-6 border-b`).
- **Risks / follow-up**:
  - A preinstalled default PR query cannot be a stored `git.queries` row as-is: `QueryInput.Validate` requires a repository of a selected project, which is unknown at install time. Options to decide in Requirements: a built-in (unstored) preset such as "Open, assigned to me" over the first/any repository, or relaxing the repo requirement (`git.prs.list` takes one repo per call; Backlog's PR API is per repository).
  - Issue saved queries need a new store key and actions (`issues.queries.*` or a shared key) plus manifest entries; `internal/plugin/manifest_test.go` asserts the action list and access, and `issues.list` has no "me" assignee (numeric ids only), while a GitHub-like "Assigned to me" default needs one (`connection` stores `ConnectedUserID`).
  - Quick action launch choice: the GitHub flow opens the native `TaskCreateDialog` (user can edit, task created by Kandev, then linked), while the plugin today creates tasks itself via `issues.create_task` (link written atomically, duplicate-create conflict handled). Using `TaskCreateDialog` needs a follow-up `issues.link` with the created task id; keeping `issues.create_task` needs a prompt/preset field and Go-side `{{url}}`/`{{title}}` interpolation. Either way, preset prompts are user data stored in plugin state; keep the 100-char-name style validation and secret redaction rules.
  - Quick action customization in GitHub uses `useSettingsSaveContributor` (page-level Save/Discard); the plugin's settings sections save per dialog today. Default preset texts should stay untranslated (persisted seed), like Kandev's.
  - Any newly used host component must be added to `ui/src/testing/harness.ts`; `ChangeRequestRow`'s fake ignores `action` today.
  - Keep `coverage.out` out of the repo root (project rule); the Go baseline above was produced with `../kandev` symlinked to the v0.96.0 checkout, which does not exist by default in this worktree.
