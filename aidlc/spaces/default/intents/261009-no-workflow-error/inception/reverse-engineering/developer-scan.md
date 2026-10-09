## Developer Code Scan Results

Focused bugfix scan for intent `261009-no-workflow-error`: clicking a "+ Task" quick action shows "Kandev has no workflow for this workspace yet." even though the workspace has a workflow, and the error display needs a mobile/desktop UI/UX retouch. No source code was modified.

### Scan Coverage
- **Analyzed deeply**:
  - ui/src/page/ (start-task.tsx, start-task.test.tsx, BacklogPage.tsx workspace resolution)
  - ui/src/issues/ (issues-page.tsx row/actions/notice layout, link-task-dialog.tsx error, task-menu.ts toast usage)
  - ui/src/messages/ (en.ts `errorWorkflow`)
  - internal/issues/ (service.go create-task validation and `ErrWorkflowMissing`, watch.go workflow fields, watcher.go error mapping)
  - internal/plugin/ (host_port.go CreateTask, issue_actions.go action names)
- **Skimmed only**:
  - ui/src/settings/ (issue-watch-dialog.tsx: same `getTaskCreationContext` read)
  - ui/src/git/ (watch-form.tsx, scm-watch-form.tsx, pr-list.tsx, scm-pr-list.tsx: sibling callers)
  - ui/src/layout.ts, ui/src/host-ui.ts, ui/src/testing/harness.ts
  - Kandev host checkout ~/repo/kandev at v0.96.0 (f099a46, equal to `.kandev-sdk-ref`), read-only: apps/packages/plugin-sdk/src/index.ts, apps/web/lib/plugins/plugin-context-api.ts, apps/web/lib/plugins/host-runtime-resources.ts, apps/web/src/spa-routes.tsx, apps/web/app/github/page.tsx, apps/web/hooks/use-workflows.ts, apps/web/components/task-create-dialog-types.ts, task-create-dialog-computed.ts, task-create-dialog-effects.ts, apps/web/components/integrations/change-request-list.tsx

### Packages Found
- `github.com/khuongdo/kandev-plugin-nulab-backlog` — Go module (plugin backend) — Go 1.26 — Kandev plugin for Nulab Backlog
  - `internal/issues` — library — Go — issue list, links, create-task, watches
  - `internal/plugin` — library — Go — `pluginsdk` adapter, action routing, HostPort
- `kandev-plugin-nulab-backlog-ui` (ui/) — UI bundle — TypeScript/React (host-provided React via `host.React`/`host.jsx`) — Backlog page, rows, dialogs

### Build System
- **Type**: Go modules + Makefile; npm + esbuild for ui/
- **Config Files**: go.mod (`replace github.com/kandev/kandev => ../kandev/apps/backend`), .kandev-sdk-ref, Makefile, manifest.yaml (`min_kandev_version: "0.96.0"`), ui/package.json, ui/package-lock.json, ui/tsconfig.json (path alias `@kandev/plugin-sdk` -> `../../kandev/apps/packages/plugin-sdk/src/index.ts`), ui/vitest.config.ts
- **Build Dependencies**: ui bundle -> `@kandev/plugin-sdk` types from the sibling ../kandev checkout; Go backend -> `pluginsdk` from ../kandev. Both require `../kandev` next to the repo (it was absent in this worktree; a symlink to ~/repo/kandev was created outside the repo for the baseline).

### APIs Discovered
- Plugin UI -> host context API (SDK `PluginContextApi`, plugin-sdk/src/index.ts:166-180): `getActiveWorkspaceId`, `subscribeActiveWorkspace`, `getTaskCreationContext(workspaceId)`, `subscribeTaskCreationContext(workspaceId, listener)`. The plugin has no host API to list workflows or steps.
- Plugin UI -> host UI kit: `IntegrationStartTaskMenu`, `TaskCreateDialog`, `ChangeRequestRow`, `host.toast.success/error`.
- Plugin UI -> plugin backend actions used by the create-task flow: `issues.link`, `git.prs.link`, `scm.prs.link` (called only after Kandev created the task). Backend also has `issues.create_task` (service.go:369-429, requires `workflowId`) and `issues.watches.save` (watch.go:109, requires `workflowId`).

### Frameworks & Libraries
- Kandev plugin SDK — v0.96.0 checkout — host APIs (Go `pluginsdk`, TS `@kandev/plugin-sdk`)
- React — ^19.3.0 (dev/test only; runtime React comes from the host)
- Vitest — ^5.0.3 with jsdom ^30.1.2, axe-core ^4.14.0 — UI tests
- Go `testing` + `github.com/stretchr/testify/require` — backend tests
- Tailwind utility classes from the host stylesheet (no plugin CSS files)

### Test Coverage
- **Test Directories**: internal/issues/*_test.go, internal/plugin/*_test.go, ui/src/**/*.test.ts(x)
- **Test Frameworks**: Go testing + testify (run with `-race`), Vitest + jsdom
- **Coverage Config**: present (Makefile `coverage` target, 80% Go floor per team rule)
- **Relevant tests**: ui/src/page/start-task.test.tsx:132-147 "explains a missing workflow instead of opening the dialog" pins the current behaviour (null context -> inline notice with `en.errorWorkflow`, dialog not opened). The default harness stub (ui/src/testing/harness.ts:872) always returns a context, so no test exercises "workflow exists but steps not hydrated".
- **Baseline (2026-10-09)**:
  - `go test -race ./internal/issues/... ./internal/plugin/...` (Go 1.26.8 at ~/.local/go/bin): PASS (issues 6.9s, plugin 6.7s). No coverage profile written.
  - `npx vitest run src/page src/issues` (after `npm ci` in ui/): PASS, 14 files, 166 tests.

### Code Quality Indicators
- **Linting**: golangci-lint + gosec, go vet, gofmt (Makefile); ESLint (ui/eslint.config.js), Prettier, `tsc --noEmit` strict
- **CI/CD**: .github/workflows (CI on pull_request, release.yml on tag) per team rules
- **Documentation**: doc comments on exported Go names; TSX components carry FR/BR references in comments

### Technical Debt Signals
- **Root cause (primary): the plugin reads a store-derived context that the plugin route never hydrates.**
  - ui/src/page/start-task.tsx:54-56 calls `host.context.getTaskCreationContext?.(workspaceId)` once, synchronously, at click time; a `null` result sets `messages.errorWorkflow` (ui/src/messages/en.ts:157) and the dialog is never opened.
  - Kandev host (apps/web/lib/plugins/plugin-context-api.ts:6-36) returns `null` in two cases: no workflow in `state.workflows.items` for the workspace (line 12), OR no steps for the chosen workflow (line 25). Steps come only from `state.kanban.steps` when `state.kanban.workflowId === workflow.id`, otherwise from `state.kanbanMulti.snapshots[workflow.id]` (lines 14-17).
  - Workflows are loaded globally by the layout (apps/web/hooks/use-workflows.ts `useEnsureWorkspaceWorkflows`), but steps/snapshots are loaded only by the kanban board (`components/kanban-board.tsx:320 useAllWorkflowSnapshots`), the task page, and first-party integration pages (`app/github/github-page-client.tsx`, `app/gitlab/gitlab-page-client.tsx`, which also SSR-hydrate steps in `app/github/page.tsx`). The plugin route renderer (apps/web/src/spa-routes.tsx:369-380 `PluginRoute`) hydrates nothing.
  - Consequence: opening `/backlog` directly (reload, bookmark, mobile) or reaching it without having opened the board leaves steps empty, so `getTaskCreationContext` returns `null` although the workspace has a workflow. This matches the report exactly. It also explains intermittency: it works after visiting the board in the same SPA session.
- **Contributing factors (secondary)**:
  - Host fallback picks the first workflow of the workspace when `workflows.activeId` belongs to another workspace (plugin-context-api.ts:7-11); workflows are fetched with `includeHidden: true` (use-workflows.ts), so the fallback can land on a hidden workflow whose snapshot is never loaded -> `null`.
  - `subscribeTaskCreationContext` exists but is unused by the plugin; subscribing alone would not fix the bug because nothing loads the steps on the plugin route.
  - The workspace id is not the cause: BacklogPage.tsx:109,130 uses `getActiveWorkspaceId`/`subscribeActiveWorkspace`, the same id the host keys workflows by.
  - The plugin backend is not involved before the error: the notice is raised client-side; `internal/issues` only sees `workflowId` for `issues.create_task` (service.go:394) and watches (watch.go:109), and watch.go:74-75 notes "the plugin cannot read workflows".
- **Fix lever found in the host (needs confirmation in Functional Design)**: Kandev's `TaskCreateDialog` resolves its own workflow when `workflowId` is null (task-create-dialog-computed.ts:62-93: manual selection -> last used per workspace -> context -> workspace workflows) and fetches steps itself whenever the effective workflow differs from the `workflowId` prop (task-create-dialog-effects.ts:32-57 `listWorkflowSteps`). The props type allows `workflowId: string | null`, `defaultStepId: string | null`, `steps: []` (task-create-dialog-types.ts:60-74). So start-task can open the dialog without a context instead of erroring.
- **Sibling callers with the same defect** (same `getTaskCreationContext` read, same `errorWorkflow` notice): ui/src/settings/issue-watch-dialog.tsx:118-121, ui/src/git/watch-form.tsx:168-171, ui/src/git/scm-watch-form.tsx:119-122. These need a concrete `workflowId` + `workflowStepId` to persist, so the dialog fallback does not apply to them; they run on the host settings route, which may or may not hydrate snapshots (not verified).
- **Error display UI/UX (desktop and mobile)**:
  - ui/src/page/start-task.tsx:79-92 renders the notice as an inline `<span role="alert" className="text-xs text-destructive">` next to the "+ Task" trigger inside `<span className="flex items-center gap-2">`.
  - That span sits in the row's action slot: issues-page.tsx:393-418 (`ROW = "flex flex-wrap items-center gap-2"`, ui/src/layout.ts:8) passed as `action` to host `ChangeRequestRow`, which wraps it in `<div className="shrink-0">` (change-request-list.tsx:84). A sentence-long message in a non-shrinking column squeezes the title column (`min-w-0 flex-1`) on narrow screens and pushes the row wider than the viewport; on desktop it shifts the trigger and menu sideways per row. Same placement in git/pr-list.tsx:334 and git/scm-pr-list.tsx:283.
  - The notice is never cleared except by the next click, has no dismiss, and gives no next step (for example a link to the board or workflow settings).
  - Other notices are inconsistent: issues-page.tsx:582-586 and BacklogPage.tsx:357 use unstyled `<p role="alert">`; link-task-dialog.tsx:152-156 uses `text-xs text-destructive`. The plugin already uses `host.toast.error` (ui/src/issues/task-menu.ts:34), which is responsive on both form factors.

## Handoff Summary
- **Intent-relevant finding**: The error is raised client-side at ui/src/page/start-task.tsx:54-55 when Kandev's `getTaskCreationContext` returns null. In Kandev v0.96.0 that happens whenever the workflow's steps are not in the app store (apps/web/lib/plugins/plugin-context-api.ts:14-25), and the plugin route (`PluginRoute`, apps/web/src/spa-routes.tsx:369-380) never loads them; only the kanban board and first-party integration pages do (`useAllWorkflowSnapshots`). So `/backlog` opened directly shows "no workflow" although one exists. Secondary: the fallback may pick a hidden workflow. The host `TaskCreateDialog` can resolve the workflow and fetch steps itself when given `workflowId: null` (task-create-dialog-computed.ts:62-93, task-create-dialog-effects.ts:32-57), which is the smallest root-cause fix for the start-task path.
- **Risks / follow-up**:
  - The same null-context defect affects three watch dialogs (settings/issue-watch-dialog.tsx:118, git/watch-form.tsx:168, git/scm-watch-form.tsx:119), which must persist a real `workflowId`/`workflowStepId`; the requirements stage should decide whether they are in scope and how they obtain a workflow (for example `subscribeTaskCreationContext`, or a clear notice with a link to the board).
  - The `workflowId: null` dialog path must be verified against the real host (packaged-host contract test or a manual check), because the plugin's Vitest harness fakes `TaskCreateDialog`; a genuine "workspace has zero workflows" case must still give a clear message.
  - start-task.test.tsx:132-147 pins the current behaviour and will need to change; add a regression test for "context null -> dialog still opens" and responsive notice tests.
  - Error-display retouch: move the row notice out of the `shrink-0` action slot (toast via `host.toast.error`, or a wrapping notice that does not widen the row) and align notice styling across issues-page, BacklogPage, link-task-dialog and the watch dialogs.
  - Environment: `../kandev` is required by go.mod and ui/tsconfig.json; for this scan a symlink `../kandev -> ~/repo/kandev` (v0.96.0, f099a46) was created outside the repo, and `ui/node_modules` was installed with `npm ci` (ignored by git).
