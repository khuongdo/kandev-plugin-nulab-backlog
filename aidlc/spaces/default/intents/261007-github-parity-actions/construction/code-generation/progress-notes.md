# Progress Notes — Code Generation (github-parity-actions)

## Step 1 — Host components at Kandev v0.96.0

All four components exist in `../kandev/apps/web/lib/plugins/host-api.ts` (tag `v0.96.0`), exported on `host.ui` at lines 353-361: `TaskCreateDialog`, `ChangeRequestRow` (props incl. `action`, `components/integrations/change-request-list.tsx:38-47`), `IntegrationStartTaskMenu` (`components/integrations/integration-start-task-menu.tsx:26-34`), `IntegrationScopeBar` (`components/integrations/presets-scope-bar-base.tsx:375-395`). `host.context.getTaskCreationContext` is in `lib/plugins/plugin-context-api.ts:6-36`. No replacement needed.

## Step 2 — Runner readiness (unchanged tree)

- `go test -race ./internal/issues/ -run 'QuickAction|IssueQuer|SaveQuery|SetQueryDefault|List'` → ok
- `go test -race ./internal/git/ -run 'Quer'` → ok
- `go test -race ./internal/plugin/ -run 'QuickAction|IssueQuer|QueryDefault|Manifest|Leak|Redact'` → ok
- `npx vitest run src/settings/sections.test.tsx src/page/backlog-lists.test.tsx src/page/backlog-page.test.tsx src/issues/issues-page.test.tsx` → 4 files, 52 tests passed

## Step 3 — Red (repository / state documents)

New tests: `internal/issues/quick_actions_test.go`, `internal/issues/queries_test.go`, additions to `internal/git/queries_test.go`.

```
$ go test -race ./internal/issues/ -run 'QuickAction|IssueQuer|SaveQuery|SetQueryDefault|List'
internal/issues/queries_test.go:10:46: undefined: IssueQuery
internal/issues/queries_test.go:12:29: r.store.UpdateQueries undefined (type *Store has no field or method UpdateQueries)
...
FAIL	github.com/khuongdo/kandev-plugin-nulab-backlog/internal/issues [build failed]
$ go test -race ./internal/git/ -run 'Quer'
internal/git/queries_test.go:91:64: in.IsDefault undefined (type *QueryInput has no field or method IsDefault)
internal/git/queries_test.go:101:21: r.svc.SetQueryDefault undefined (type *Service has no field or method SetQueryDefault)
FAIL	github.com/khuongdo/kandev-plugin-nulab-backlog/internal/git [build failed]
```

## Step 4 — Green

`internal/issues/quick_actions.go` (types, defaults, `Resolved`, store load/update on `issues.quick_actions`), `internal/issues/queries.go` (`IssueQuery`, store on `issues.queries`, max 50, `Service.SetQueryDefault`), `internal/git/types.go` (`QueryInput.IsDefault`), `internal/git/service.go` (`SetQueryDefault`; `SaveQuery` keeps the stored star and never sets it). Scoped Go commands: ok.

Note: `issues.Service.SetQueryDefault` was pulled forward from Step 7 because the Step 3 test for "set_default leaves exactly one" exercises it.

## Step 5 — Refactor

Star logic rewritten as "clear all when starring, then set this row" for readability; new store methods reuse `load`/`put`/`lock`. `go test -race ./internal/issues/ ./internal/git/` ok; gofmt clean.

## Step 6 — Red (business logic)

Tests: `TestQuickActionsInput_Validate`, `TestQuickActions_SaveReplacesOneKindAndEmptyResets`, `TestIssueQuery_Validate`, `TestIssueQueries_SaveListDeleteKeepTheStar`, `TestList_AssigneeMeIsResolvedOnTheServer`.

```
$ go test -race ./internal/issues/ -run 'QuickAction|IssueQuer|SaveQuery|SetQueryDefault|List'
internal/issues/list_test.go:167:44: unknown field Assignee in struct literal of type Query
internal/issues/queries_test.go:90:45: c.q.Validate undefined (type IssueQuery has no field or method Validate)
internal/issues/queries_test.go:103:21: r.svc.ListQueries undefined (type *Service has no field or method ListQueries)
internal/issues/queries_test.go:108:18: r.svc.SaveQuery undefined (type *Service has no field or method SaveQuery)
...
FAIL	github.com/khuongdo/kandev-plugin-nulab-backlog/internal/issues [build failed]
```

## Step 7 — Green

`QuickActionsInput.Validate`, `Service.QuickActions` / `SaveQuickActions` (quick_actions.go); `IssueQuery.Validate`, `Service.ListQueries` / `SaveQuery` / `DeleteQuery` (queries.go); `Query.Assignee` with `ValidateQuery` check and one `Myself` call in `Service.List` (types.go, service.go). Scoped command: ok.

## Step 8 — Refactor

Named `maxQueryName` instead of reusing the watch limit. `go test -race ./internal/issues/ ./internal/git/` ok.

## Step 9 — Red (API / endpoint)

Tests: `TestQuickActionActions_GetSaveAndValidate`, `TestIssueQueryActions_CRUDAndDefault` (actions_u3_test.go), `TestQueryDefaultAction_MovesTheStar` (actions_u4_test.go), manifest expectations (u3 list 25 keys, `issues.quick_actions.save` max body 262144), leak test extended to the new actions (actions_watch_test.go).

```
$ go test -race ./internal/plugin/ -run 'QuickAction|IssueQuer|QueryDefault|Manifest|Leak|Redact'
internal/plugin/actions_u3_test.go:298:2: undefined: actionQuickActionsGet
internal/plugin/actions_u3_test.go:299:26: undefined: actionIssueQueriesSave
internal/plugin/actions_u4_test.go:428:2: undefined: actionQueriesDefault
...
FAIL	github.com/khuongdo/kandev-plugin-nulab-backlog/internal/plugin [build failed]
```

## Step 10 — Green

Handlers for the 7 new keys in `internal/plugin/issue_actions.go` and `internal/plugin/git_actions.go` (shared `withDefault` decoder for `{id, isDefault}`), entries in `manifest.yaml` (`issues.quick_actions.save` gets `max_body_bytes: 262144`, below Kandev's 1 MiB cap, because 20 prompts of 4,000 characters do not fit in 16 KiB). Scoped plugin command: ok.

## Step 11 — Refactor + full Go run

No further refactor needed. `go test -race ./internal/... ./server/...` → all 9 packages ok.

## Step 12 — Harness

`ui/src/testing/harness.ts`: fakes for `IntegrationStartTaskMenu` (trigger + one button per preset with `itemTestId` and `data-preset-id`), `TaskCreateDialog` (renders when open, shows the initial title and description, Create calls `onSuccess({id: "t-1"})`), `IntegrationScopeBar` (kind buttons, preset pills, Saved items with star and delete, Save current), `Textarea`; the `ChangeRequestRow` fake renders `action`. Full Vitest suite unchanged: 29 files, 286 tests passed.

## Step 13 — Red (quick action launch)

```
$ npx vitest run src/page/quick-actions.test.ts src/page/start-task.test.tsx
 FAIL  src/page/quick-actions.test.ts
Error: Failed to resolve import "./quick-actions" from "src/page/quick-actions.test.ts". Does the file exist?
 FAIL  src/page/start-task.test.tsx
Error: Failed to resolve import "./start-task" from "src/page/start-task.test.tsx". Does the file exist?
 Test Files  2 failed (2)
```

## Step 14 — Green

`ui/src/page/quick-actions.ts`, `ui/src/page/start-task.tsx`; issue rows now use `ChangeRequestRow` with the "+ Task" menu and the "Link to task" row menu (the prompt-less "Create task" item, its linked-task confirm dialog and the table/cards markup were removed); PR rows get the menu in the `action` slot; `BacklogPage` loads `issues.quick_actions.get` once per connected workspace and passes each kind down. Messages `startTask`, `startTaskLabel`, `taskNotLinked`. Existing issues-page tests that asserted the removed table, cards, title clamp and Create task flow were replaced by row/menu tests. Full Vitest: 31 files, 300 tests passed.

Deviation: titles are truncated to Kandev's 60 characters with "…" (`truncateRemoteTaskTitle`), as FR1.2 requires "the same way as Kandev", instead of the plan's 100.

## Step 15 — Red (quick-actions settings)

```
$ npx vitest run src/settings/sections.test.tsx
 × stacks the seven sections in order when connected
 × shows the Issues tab with the defaults and saves edits, additions and deletions
   TypeError: Cannot read properties of null (reading 'getAttribute')
 × resets a kind by saving an empty list and shows the defaults again
 × refuses an empty or over-long label with a message and saves nothing
 × shows a server refusal without saving and an error when loading fails
 Tests  7 failed
```

## Step 16 — Green

New `ui/src/settings/quick-actions-section.tsx` (tabs Issues / Pull requests, editor per action, Add, Delete, Save, Reset = save empty list, client label check, server refusal notice, load error with Retry), mounted in `SettingsScreen.tsx` after Saved queries (visible to members: the actions are `authenticated`). Full Vitest green.

## Step 17 — Red (default queries, scope bar, saved issue queries)

New describe "Default queries and scope bar" in `ui/src/page/backlog-lists.test.tsx` (10 tests) and "Saved issue queries in Settings" in `ui/src/settings/sections.test.tsx` (1 test).

```
$ npx vitest run src/page/backlog-lists.test.tsx src/settings/sections.test.tsx
 × replaces the tabs with the host scope bar holding the kinds, a preset and the saved menu
 × opens Issues on 'Assigned to me, open' without a starred issue query, with one list call
 × opens Issues on the starred issue query instead
 × opens Pull requests on 'Open, assigned to me' in the first repository without a starred PR query
 × opens Pull requests on the starred PR query instead
 × shows the repository guidance when the projects have no repository (FR3.4)
 × stars and un-stars saved queries through set_default
 × applies a saved query from the menu and ignores one of an unselected project (BR2.4)
 × saves the issue filters as a saved issue query and lists it in the menu (FR4.3)
 × deletes a saved query from the menu and falls back to the preset
 × lists issue queries next to PR queries, renames, stars and deletes them
 Tests  11 failed | 25 passed (36)
```

## Step 18 — Green

- `ui/src/page/BacklogPage.tsx`: host `IntegrationScopeBar` (kinds, one built-in preset per kind, Saved menu with star, delete and Save current) replaces the `Tabs`; reads `issues.queries.list`, `git.queries.list` and quick actions once per connected workspace, then mounts the list on the starred usable query or the preset; star → `*.queries.set_default`, delete → `*.queries.delete` (falls back to the preset when the open query is deleted).
- `ui/src/issues/issues-page.tsx`: filters shaped like a saved issue query (status "Not closed" and assignee "Me" choices), `selection` applied once (the preset waits for `issues.filters`, so one `issues.list` call), "Save query" toolbar button + dialog (`issues.queries.save`).
- `ui/src/git/pr-list.tsx`: `selection` applied once (the preset waits for the repositories and uses the first one); the in-toolbar saved-query Select and its own `git.queries.list` call were removed (the scope bar's Saved menu replaces them).
- `ui/src/git/save-query-dialog.tsx`: `action` and `description` props so it saves either kind.
- `ui/src/settings/saved-queries-section.tsx`: an Issues table next to the Pull requests table, each with Rename, Make/Remove default and Delete.
- `ui/src/issues/issues-state.ts`: `IssueQuery`, `CLOSED_STATUS_ID` (4), `openStatusIds`, `issueQueryFilters`.
- Old page tests that asserted the tabs, the empty "choose a repository" start and the in-toolbar preset Select were updated to the scope bar and preset behaviour. Full Vitest: 31 files, 316 tests passed.

## Step 19 — Red then Green (layout)

Red: new test "puts the host scope bar first, a bordered toolbar with refresh last, and padded results" (`ui/src/page/backlog-lists.test.tsx`):
```
 × puts the host scope bar first, a bordered toolbar with refresh last, and padded results
TypeError: Cannot read properties of null (reading 'classList')
 Tests  1 failed | 19 skipped (20)
```
Green: `ui/src/layout.ts` gains `TOOLBAR` (`flex shrink-0 flex-col gap-2 border-b px-4 py-2.5 sm:px-6 md:flex-row md:flex-wrap md:items-center md:gap-3`) and `RESULTS` (`px-3 py-4 md:px-6`); `ui/src/git/pr-toolbar.tsx` is one bordered row (title + count, filters, last fetched + ghost refresh last) with `idPrefix`, `refreshLabel` and `headingRef`, and is reused by the issue list; both lists wrap their results in `RESULTS`. The scope bar's own padding (`px-4 py-2 sm:px-6`) comes from the host `IntegrationScopeBar`, so the page adds no wrapper around it (the test checks the bar is the page's first child and the page root has no padding). Full Vitest: 317 passed.

## Step 20 — Refactor + full UI checks

Deleted code made dead by the switch: `rowLabel`/`joinList` (issues-state.ts) and their test, the messages `queryLabel`, `chooseQuery`, `colTitle`, `colKey`, `colUpdated`, `createTask`, `creatingTask`, `createdTask`, `rowSr*`, `listAnd`, `listComma`, `alreadyLinked*`, `openTask`, `createAnother`, `scopeLabel` (the i18n and registration tests now use `linkToTask` as their sample key). Added a unit test for `openStatusIds`/`issueQueryFilters`. The non-connected page states get `px-4 py-4 sm:px-6` (plugin routes have no host padding). Settings card, switch, nav entry and `settingsHref()` code and tests are untouched (FR5.4).

`npm run typecheck` clean, `npm run lint` clean, `npm run format:check` clean, `npm test` → 31 files, 317 tests passed.

## Step 21 — Build configuration

No new Go or npm dependency; `go mod tidy` leaves `go.mod`/`go.sum` unchanged. `manifest.yaml` version `0.1.1` → `0.2.0`.

## Step 22 — Full verification

- `make check-format vet lint test coverage` → exit 0; golangci-lint `0 issues`; Go `-race` all 9 packages ok; Vitest 31 files / 317 tests passed; coverage **92.8%** (floor 80%, excluded `server/main.go`). The profile is written to `build/` (git-ignored) and was deleted afterwards; no `coverage.out` in the repository root.
- `ui/`: `npm run typecheck`, `npm run lint`, `npm run format:check` clean; `npm test` 317 passed.
- `make package verify-package` → `verifypkg: OK dist/nulab-backlog-0.2.0.tar.gz (nulab-backlog@0.2.0)`.

## Step 23 — Documentation and traceability

`README.md` (What it does: quick actions, default queries, scope bar; Upgrade notes 0.2.0), `code-summary.md`, `source-manifest.json`, `traceability.json`.
