# Code Summary — github-parity-actions

Zero-Unit express run, one implementation iteration, TDD per the Testing Contract. Red outputs for every layer are in [progress-notes.md](progress-notes.md).

## Host Component Citations (Step 1)

All components the plan relies on are exposed on `host.ui` at Kandev `v0.96.0` (`../kandev`, tag `v0.96.0`), so nothing was rebuilt in the plugin:

| Component | Exposed at | Props used |
|---|---|---|
| `TaskCreateDialog` | `apps/web/lib/plugins/host-api.ts:353` | `open, onOpenChange, workspaceId, workflowId, defaultStepId, steps, initialValues{title, description}, onSuccess(task)` (`components/task-create-dialog-types.ts:62-106, 198-221`) |
| `ChangeRequestRow` | `host-api.ts:355` | `stateIcon, title, href, metadata, taskIndicator, action, testId` (`components/integrations/change-request-list.tsx:38-47`) |
| `IntegrationStartTaskMenu` | `host-api.ts:357` | `presets{id,label,hint,iconName}, onSelect, triggerLabel, triggerAriaLabel, triggerTestId, itemTestId` (`components/integrations/integration-start-task-menu.tsx:26-34`) |
| `IntegrationScopeBar` | `host-api.ts:360` | `testId, savedMenuTestId, kinds, selected, onSelect, onKindChange, presetsByKind, savedPresets{isDefault}, onDeleteSaved, canSaveCurrent, onSaveCurrent, onToggleSavedDefault, defaultMutationPendingId` (`components/integrations/presets-scope-bar-base.tsx:375-395`) |
| `host.context.getTaskCreationContext` | `lib/plugins/plugin-context-api.ts:6-36, 79` | workflow, first step and steps for the dialog |

## Files

Created:
- `internal/issues/quick_actions.go` (+ `_test.go`): `QuickAction`, `QuickActions`, the six Backlog-worded defaults, `Resolved` (empty kind → defaults), `QuickActionsInput.Validate`, store on `issues.quick_actions`, `Service.QuickActions` / `SaveQuickActions`.
- `internal/issues/queries.go` (+ `_test.go`): `IssueQuery` with `Validate`, store on `issues.queries` (max 50), `Service.ListQueries` / `SaveQuery` / `DeleteQuery` / `SetQueryDefault`.
- `ui/src/page/quick-actions.ts` (+ test): types, `interpolate`, `taskTitle`, `loadQuickActions`.
- `ui/src/page/start-task.tsx` (+ test): the "+ Task" menu → `TaskCreateDialog` → `issues.link` / `git.prs.link`.
- `ui/src/settings/quick-actions-section.tsx`: Settings "Quick actions" section.

Modified:
- Go: `internal/issues/types.go` and `service.go` (`Query.Assignee` "me", resolved with one `Myself` call), `internal/git/types.go` and `service.go` (`QueryInput.IsDefault`, `SetQueryDefault`, save keeps the star), `internal/plugin/issue_actions.go` and `git_actions.go` (7 new actions, `withDefault` decoder), `manifest.yaml` (7 keys, version `0.2.0`), plugin tests (`actions_u3_test.go`, `actions_u4_test.go`, `actions_watch_test.go`, `manifest_test.go`), `internal/git/queries_test.go`, `internal/issues/list_test.go`.
- UI: `page/BacklogPage.tsx` (scope bar, saved queries, default-on-open, quick actions loaded once), `issues/issues-page.tsx` (ChangeRequestRow rows, "+ Task", filters with "Not closed"/"Me", selection, Save query, toolbar), `git/pr-list.tsx` (selection with the preset, row "+ Task", in-toolbar preset Select removed), `git/pr-toolbar.tsx` (one-row GitHub toolbar, reused by issues), `git/save-query-dialog.tsx` (`action`, `description`), `settings/saved-queries-section.tsx` (issue table, star), `settings/SettingsScreen.tsx`, `issues/issues-state.ts`, `layout.ts`, `icons.tsx`, `messages/en.ts`, `testing/harness.ts`, and the affected tests.
- `README.md`: features and 0.2.0 upgrade notes.

The full list is in [source-manifest.json](source-manifest.json).

## Key Decisions

- Defaults live only in Go; the UI reads them resolved from `issues.quick_actions.get`. Reset saves an empty list, so there is no reset action.
- New actions are `authenticated` and workspace-scoped like `git.queries.*`. `issues.quick_actions.save` declares `max_body_bytes: 262144` (Kandev's cap is 1 MiB): 20 prompts of 4,000 characters do not fit in the usual 16 KiB.
- Saving a query never sets the star; only `*.queries.set_default` does, clearing every other star in the same read-modify-write (at most one default per kind). A missing `isDefault` reads false, so stored data needs no migration.
- `issues.list` accepts only `assignee: ""|"me"`; "me" replaces `assigneeIds` and is resolved on the server. Saved issue queries store `""|"me"|<id>`, and the UI maps a numeric id to `assigneeIds`.
- The issue preset is `assignee: "me"` plus every status from `issues.filters` except Backlog's Closed (id 4), which settles the requirements' open question. The issue list waits for `issues.filters` before applying it, and the PR list waits for the repositories (first one), so each list makes exactly one call on open (NFR4).
- The scope bar's Saved menu replaces the old saved-query Select in the PR toolbar. A saved query whose project is no longer selected is labelled and cannot be applied or used as the default.
- A link failure after `TaskCreateDialog` keeps the task and shows "The task was created but not linked".

## Test Coverage Summary

- Go: `make coverage` → **92.8%** over `./internal/...` and `./server/...` (floor 80%, only `server/main.go` excluded), all with `-race`. Per package: backlog 96.1%, ci 91.0%, connection 94.6%, git 90.9%, issues 91.5%, pkgverify 93.2%, plugin 93.8%, redact 97.4%, testutil 88.0%. New functions: `quick_actions.go` 70-100% per function, `queries.go` 83-100%, both `SetQueryDefault` 100%.
- New Go tests: 16 (quick actions 5, issue queries 5, `List` assignee "me" 1, git queries 2, plugin actions 3), plus the updated manifest test and the leak test extended to the new actions.
- UI (Vitest): 317 tests in 31 files pass (baseline before this change: 286 in 29 files). New: `quick-actions.test.ts` (6), `start-task.test.tsx` (7), Quick actions section (4), saved issue queries in Settings (1), default queries and scope bar (10), layout (1), issue row "+ Task" (1), `openStatusIds`/`issueQueryFilters` (1). Removed with their code: the Create task flow tests (2), the table/title-clamp assertions, `rowLabel`, and the PR preset-Select test (replaced by the Saved menu test).
- `tsc --noEmit`, ESLint, Prettier, gofmt, go vet, golangci-lint: clean. `make package verify-package`: OK for `nulab-backlog@0.2.0`.

## Deviations from Plan

- Task titles are cut at Kandev's 60 characters with "…" (`truncateRemoteTaskTitle`), not 100, because FR1.2 requires "truncated the same way as Kandev" and the dialog clamps titles at 60 anyway.
- `issues.Service.SetQueryDefault` was implemented in Step 4 instead of Step 7, because the Step 3 test for "set_default leaves exactly one" uses it.
- Step 19's "scope bar wrapper has `px-4 py-2 sm:px-6`": the host `IntegrationScopeBar` already carries those classes, so the page adds no padded wrapper. The test checks that the bar is the page's first child and that the page root has no padding.
- Hint text is limited to 100 characters (not in FR2.4) so every user-supplied field has a bound.
- Settings "Saved PR queries" became "Saved queries" with an Issues table and a Pull requests table. Quick actions sit after it as an eighth section.
- The plan named a `net/http/httptest` harness for the "me" test. The issues package uses its existing in-memory fake gateway, which counts the `myself` call and records the `assigneeId[]` sent.
