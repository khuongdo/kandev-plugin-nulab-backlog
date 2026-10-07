# Code Summary — 261007-uiux-github-style

Zero-Unit refactor intent, one implementation iteration. Methodology: TDD (Testing Contract `sha256:f482164fba027f02e5cb78452acda04f5091d3f5ec9a0217ffb0cde178cabfa0`).

## Step 1 — Package layout

- Issue watch: `internal/issues` (`watch.go` types and validation, `watch_store.go` documents, `watcher.go` loop and run).
- PR list: `internal/git/prs.go`; gateway additions in `internal/backlog/issues.go` and `internal/backlog/pullrequests.go`.
- Actions: `internal/plugin/issue_actions.go`, `internal/plugin/git_actions.go`; manifest entries in `manifest.yaml`.
- No new package, no new dependency, no new manifest capability.

## Step 2 — Baseline (unchanged tree)

Environment: go1.26.8 linux/amd64; `../kandev` → `/home/k_do_webfrontier/repo/kandev` at `f099a46dc7aab16f6ff5806cd29b2b480296303f` (equals `.kandev-sdk-ref`); Node v25.2.1 locally (`.nvmrc` asks for 22; CI uses 22).

```text
$ PATH="$HOME/.local/go/bin:$PATH" go test -race ./internal/backlog/... ./internal/issues/... ./internal/git/... ./internal/plugin/...
ok  	.../internal/backlog	3.757s
ok  	.../internal/issues	2.422s
ok  	.../internal/git	1.223s
ok  	.../internal/plugin	6.287s

$ cd ui && npx vitest run
 Test Files  28 passed (28)
      Tests  229 passed (229)

$ cd ui && npx tsc --noEmit && npx eslint . && npx prettier --check .
All matched files use Prettier code style!
```

## Red Outputs

### Step 3 — Data model (Red)

```text
$ go test -race ./internal/issues/... ./internal/git/... -run 'TestIssueWatch|TestIssueCursor|TestQuery_|TestU4_Query'
internal/git/types_test.go:146:41: q.Creator undefined (type *QueryInput has no field or method Creator)
internal/git/types_test.go:159:32: q.Creator undefined (type QueryInput has no field or method Creator)
internal/issues/watch_test.go:11:24: undefined: IssueWatchInput
internal/issues/watch_test.go:30:64: undefined: FieldName
internal/issues/watch_test.go:32:83: undefined: FieldProjectKey
FAIL	.../internal/issues [build failed]
FAIL	.../internal/git [build failed]
```

### Step 6 — Repository / data access (Red)

```text
$ go test -race ./internal/issues/... ./internal/backlog/... -run 'TestWatchStore|TestIssueQuery_|TestIssues_Return|TestPullRequestCount|TestU4_PRParse'
internal/issues/watch_store_test.go:22:23: s.UpdateWatches undefined (type *Store has no field or method UpdateWatches)
internal/issues/watch_store_test.go:27:16: s.WatchIndex undefined (type *Store has no field or method WatchIndex)
internal/issues/watch_store_test.go:39:29: undefined: maxIssueWatches
internal/issues/watch_store_test.go:61:51: undefined: errDuplicateEntry
FAIL	.../internal/issues [build failed]
internal/backlog/pullrequests_types_test.go:19:36: unknown field AuthorName in struct literal of type PullRequest
internal/backlog/watch_queries_test.go:16:71: unknown field CreatedUserIDs in struct literal of type IssueQuery
internal/backlog/watch_queries_test.go:17:3: unknown field CreatedSince in struct literal of type IssueQuery
internal/backlog/watch_queries_test.go:28:53: issues[0].Created undefined (type Issue has no field or method Created)
internal/backlog/watch_queries_test.go:40:14: c.PullRequestCount undefined (type *Client has no field or method PullRequestCount)
FAIL	.../internal/backlog [build failed]
```

### Step 9 — Business logic (Red)

```text
$ go test -race ./internal/issues/... ./internal/git/...
git/prs_test.go:11:28: undefined: PRListInput
git/prs_test.go:24:21: r.svc.ListPullRequests undefined (type *Service has no field or method ListPullRequests)
git/prs_test.go:28:20: undefined: PullRequestPage
git/prs_test.go:67:3: undefined: FieldPage
issues/watcher_test.go:37:18: r.svc.SaveWatch undefined (type *Service has no field or method SaveWatch)
issues/watcher_test.go:90:21: r.svc.ListWatches undefined (type *Service has no field or method ListWatches)
issues/watcher_test.go:136:7: undefined: NewWatcher
issues/watcher_test.go:137:23: r.svc.PauseWatch undefined (type *Service has no field or method PauseWatch)
FAIL	.../internal/issues [build failed]
FAIL	.../internal/git [build failed]
```

### Step 12 — API / endpoint (Red)

```text
$ go test -race ./internal/plugin/...
internal/plugin/actions_u3_test.go:295:2: undefined: actionIssueWatchesList
internal/plugin/actions_u3_test.go:295:26: undefined: actionIssueWatchesSave
internal/plugin/actions_u3_test.go:296:27: undefined: actionIssueWatchesResume
internal/plugin/actions_u4_test.go:427:2: undefined: actionPRList
internal/plugin/actions_watch_test.go:48:25: undefined: actionIssueWatchesSave
internal/plugin/actions_watch_test.go:58:36: undefined: actionIssueWatchesPause
FAIL	.../internal/plugin [build failed]
```

### Step 15 — Frontend behavior (Red)

```text
$ cd ui && npx vitest run src/index.test.ts src/brand src/page src/settings src/issues src/git src/switch src/controls.test.ts
 × shows an alert with the settings link and no lists when not connected
 × opens Issues by default and switches to Pull requests in the URL
 × asks for a repository first, then lists 20 rows with the linked task and pages
 × applies a saved query as a preset; a query of an unselected project is disabled (BR2.4)
 × stacks the seven sections in order when connected
 × keeps the watch and query sections in the member view (BR1.2)
 × chooses the sign-in method with a dropdown (BR1.4)
 × points the restore notice at the PR watches section (BR1.5)
 × adds a watch in a dialog with the default interval 5 and field errors (BR3.1)
 × has exactly one Integrations entry and one route, /backlog (BR2.1, FR2.1, FR2.2)
 × is an outline drawing in the current text colour
 × ./issues/issues-page.tsx draws no raw select, table, details, button, checkbox or radio
 × ./settings/SettingsScreen.tsx gives every host Button the GitHub cursor (BR5.2)
⎯⎯⎯⎯⎯⎯ Failed Tests 51 ⎯⎯⎯⎯⎯⎯⎯
```

## Green and Refactor Notes

- **Step 4 (data model)**: `internal/issues/watch.go` — `IssueWatchInput.Validate` (BR3.1, BR3.2), `IssueCursor.After` / `CreatedSince` (BR3.5, R-10), `IssueWatch`, `IssueWatchLedgerEntry`; `git.QueryInput.Creator` (absent = anyone).
- **Step 5**: the `anyone|me` check stays local in `internal/issues` (it may not import `internal/git`; marked `ponytail:`); validation order made deterministic.
- **Step 7 (data access)**: `internal/issues/watch_store.go` — `issues.watches` (cap 50, own instance index `issues.watch_index`), one ledger document per watch `issues.watch_ledger.<id>` (unique key, cap 5000 → `ErrLedgerFull`, emptied on delete); `internal/backlog` — `IssueQuery` gains `CreatedUserIDs`, `CreatedSince`, `Sort`, `Order` (defaults unchanged), `Issue.Created`, `PullRequest.AuthorName` / `Updated`, new `Client.PullRequestCount` (`.../pullRequests/count`, Read group, project key checked before any request); fixture `testdata/pullrequests_count_ok.json`.
- **Step 8**: copied document helpers kept as they are (existing `ponytail:` note).
- **Step 10 (business logic)**: `Service.ListWatches/SaveWatch/DeleteWatch/PauseWatch/ResumeWatch`, `applyWatches` (SM1 with `stateBeforeDisconnect`, wired into `OnConnectionChanged` and every watcher cycle), `issues.Watcher` (1-minute tick, run queue, `issue_watch_cycle` log line), the run (`watchRun`: resolve reservations → at most 5 pages of 100 from `max(0, dayOffset-5)` → skip ledger/linked issues → reserve → create → link → mark → stop after one task); `git.Service.ListPullRequests` (20 per page, total from the count call, linked tasks).
- **Step 11**: `Service.createLinkedTask` is now the one create-and-link path of `issues.create_task` and the watcher.
- **Step 13 (actions)**: `issues.watches.list|save|delete|run|pause|resume` and `git.prs.list` (all `workspace` / `authenticated`, 16384 bytes) in `manifest.yaml` and the handlers; `issues.Watcher` started and stopped with the runtime; `issueHost.CreateTask` maps a Kandev `NotFound` / `InvalidArgument` refusal to `issues.ErrWorkflowMissing` (BR3.14).
- **Step 14**: handlers use the existing `decode` / `withID` shapes; nothing else to align.
- **Step 16 (UI)**: Settings is seven framed `SettingsSection`s (`SettingsScreen.tsx`, `pr-watches-section.tsx`, `issue-watches-section.tsx`, `issue-watch-dialog.tsx`, `saved-queries-section.tsx`, shared `section-parts.tsx` and `use-list.ts`); the PR watch form (`git/watch-form.tsx`) is restyled and shown in a host `Dialog`; the confirm and link-task dialogs use the host `Dialog`; `/backlog` shows an `Alert` when not connected or off, otherwise host `Tabs` Issues / Pull requests (`?scope=prs`), with `git/pr-list.tsx`, `git/pr-toolbar.tsx`, `git/save-query-dialog.tsx`; every raw `select`, `table`, `details`, checkbox/radio `input` and `button` replaced by host components with the GitHub variant/size table; original outline icon behind `PLUGIN_ICON`; `/backlog/watches` and `/backlog/dashboard` removed with their pages and tests.
- **Step 17**: one `ui/src/layout.ts` (`STACK`, `FIELD`, `ROW`, `BUTTON`) replaces the per-file copies; `ui/src/host-ui.ts` gives the host kit loose prop types once; `ui/src/icons.tsx` draws the three stroke icons of icon-only buttons; 12 unused message keys of the removed pages deleted.
- **Step 18**: `make coverage` writes `build/coverage.out` (floor 80% and the single exclusion `server/main.go` unchanged); `make ui-build` passes `--jsx-fragment=Fragment` like `npm run build`.

## Files

Created (application source and tests):
`internal/issues/watch.go`, `internal/issues/watch_store.go`, `internal/issues/watcher.go`, `internal/git/prs.go`, `internal/issues/watch_test.go`, `internal/issues/watch_store_test.go`, `internal/issues/watcher_test.go`, `internal/git/prs_test.go`, `internal/backlog/watch_queries_test.go`, `internal/backlog/testdata/pullrequests_count_ok.json`, `internal/plugin/actions_watch_test.go`, `ui/src/layout.ts`, `ui/src/host-ui.ts`, `ui/src/icons.tsx`, `ui/src/controls.test.ts`, `ui/src/git/pr-list.tsx`, `ui/src/git/pr-toolbar.tsx`, `ui/src/git/save-query-dialog.tsx`, `ui/src/page/backlog-lists.test.tsx`, `ui/src/settings/issue-watch-dialog.tsx`, `ui/src/settings/issue-watches-section.tsx`, `ui/src/settings/pr-watches-section.tsx`, `ui/src/settings/saved-queries-section.tsx`, `ui/src/settings/section-parts.tsx`, `ui/src/settings/use-list.ts`, `ui/src/settings/sections.test.tsx`.

Modified: `Makefile`, `manifest.yaml`, `README.md`, `docs/brand/backlog-logo.md`; Go `internal/backlog/{client,issues,pullrequests}.go`, `internal/git/{service,types}.go`, `internal/issues/{events,service}.go`, `internal/plugin/{git_actions,issue_actions,host_port,runtime}.go` and tests `internal/backlog/{issues_types,pullrequests_types}_test.go`, `internal/git/{harness,types}_test.go`, `internal/issues/harness_test.go`, `internal/plugin/{actions_u3,actions_u4,manifest}_test.go`; UI `ui/src/index.ts`, `brand/backlog-logo.tsx`, `git/{git-access,watch-form}.tsx`, `issues/{issue-badge,issue-panel,issues-page,link-task-dialog,poll-interval}.tsx`, `messages/en.ts`, `page/BacklogPage.tsx`, `settings/{SettingsScreen,confirm-dialog,connected-panel,project-picker}.tsx`, `testing/harness.ts` and tests `index.test.ts`, `brand/backlog-logo.test.tsx`, `git/{git-access,watch-form}.test.tsx`, `issues/issues-page.test.tsx`, `page/backlog-page.test.tsx`, `settings/{confirm-dialog,connected-panel,oauth,project-picker,settings}.test.tsx`, `switch/switch.test.tsx`.

Deleted: `ui/src/git/watches-page.tsx`, `ui/src/git/watches-page.test.tsx`, `ui/src/git/dashboard-page.tsx`, `ui/src/git/dashboard-page.test.tsx` (their behaviour is re-asserted in `settings/sections.test.tsx` and `page/backlog-lists.test.tsx`).

The full list is `source-manifest.json` (81 paths).

## Key Decisions

- R-09: the first page of a run that starts past the cursor (empty, or its first issue already after the cursor) restarts once from offset 0 (`TestIssueWatchRun_OvershootRestartsFromZero`). R-10: `createdSince` = cursor UTC date minus one day (`TestIssueCursor_CreatedSinceIsTheUTCDayBefore`, `TestIssueWatchRun_CreatedSinceIsTheDayBefore`).
- `dayOffset` is exact when the new cursor keeps the run's `createdSince`; otherwise it is counted over the issues seen in the run, which can only undercount (a later run starts a little earlier, never past an issue).
- The ledger is one state document per watch (`issues.watch_ledger.<id>`), so deleting a watch empties exactly its ledger and the 5000 cap is per watch.
- Issue watches use their own instance index, so a watch alone does not make the 1-minute status sync list a workspace's tasks.
- The issue watcher waits for the Kandev host through the same bounded store wait as the other workers; a failed tick is retried at the next one (NFR6, `TestIssueWatcher_StartsWithThePluginAndWaitsForTheHost`).
- UI tests stub every host component as plain DOM marked `data-host`, with `variant` / `size` as `data-variant` / `data-size`; `rawControls()` and the source scan in `controls.test.ts` enforce BR5.1, and the button table of BR5.2 is asserted per screen.
- Saved-query edit in Settings renames the query and keeps its filters (filters come from the Pull requests list, BR2.5).

## Test and Coverage Results

```text
$ PATH="$HOME/.local/go/bin:$PATH" go test -race -count=1 -coverprofile=build/coverage.out ./internal/... ./server/...
backlog 96.1%  ci 91.0%  connection 94.6%  git 90.8%  issues 91.3%  pkgverify 93.2%
plugin 93.7%  redact 97.4%  testutil 88.0%  server 0.0% (excluded)
total (raw profile): 92.7%      (profile deleted afterwards; no coverage.out at the repo root)

$ make check-format vet lint test coverage check-secrets build package verify-package
check-format OK · vet OK · lint: golangci-lint v2.14.0 "0 issues", tsc, eslint, actionlint, ci workflows OK
test OK (Go -race all packages; Vitest 29 files / 286 tests)
coverage: 92.8% (floor 80%, excluded: server/main.go) · ci secrets: OK
build OK (5 platforms) · package OK (no React in the bundle) · verifypkg: OK dist/nulab-backlog-0.1.0.tar.gz

$ cd ui && npx vitest run src/index.test.ts src/brand src/page src/settings src/issues src/git src/switch
 Test Files  28 passed (28)      Tests  237 passed (237)
$ cd ui && npx vitest run      (adds src/controls.test.ts)
 Test Files  29 passed (29)      Tests  286 passed (286)
$ cd ui && npx tsc --noEmit && npx eslint . && npx prettier --check .
All matched files use Prettier code style!
```

Baseline → now: Go coverage 92.9% → 92.7% raw (92.8% gated); UI 229 → 286 tests (4 test files removed with their pages, 5 added).

## Deviations

- **BR1.4 (partial, Deferred in traceability)**: the sign-in method is a dropdown, but it still offers both choices when OAuth is not configured: `connection.get` has no "OAuth configured" flag, and adding one is a backend contract change. Choosing OAuth without the server setup shows the existing `oauthNotConfigured` message.
- **Issue list toolbar**: the Issues scope keeps the plugin's own toolbar (search with the 400 ms wait, refresh as a ghost icon button) instead of `IntegrationListToolbar`, whose search commits on Enter and would change the existing search behaviour; it matches the Pull requests toolbar.
- **PR rows**: the state icon is `host.ui.IntegrationIcon` (`pull-request`, `merged`, `pull-request-closed`) with the state as text; `IntegrationChangeRequestStatus` is the topbar/composer status widget (pipeline rows, review), not a row icon.
- **Pagination**: Previous / Next are host `Button`s (outline, sm) inside `Pagination` / `PaginationContent` / `PaginationItem`; `PaginationPrevious` / `PaginationNext` are links that need an `href`.
- **Issue watch dialog**: statuses come from `issues.filters` (the union of the selected projects' statuses; the action has no per-project parameter), and the workflow and step come from Kandev's task creation context, like the PR watch form (there is no host workflow picker in SDK v0.96.0's `host.ui`).
- **BR3.11 (429)**: the gateway's Background class already waits `Retry-After` (up to 3 retries) inside the run; after that the run records `rate_limited` and the next run follows the watch interval. No separate "next allowed run" time is stored.
- **Replace heading**: "Replace connection" became bold text instead of a heading, so the heading order stays valid when a dialog opens inside the Connection section.
- **Not run**: `make contract-test` (packaged-host contract on Kandev 0.96.0) needs a `../kandev-min` checkout, which does not exist here; CI's `packaged-host-contract` job covers it. Node is v25.2.1 locally (`.nvmrc` 22).
- **Unit-test command**: the new cross-cutting `ui/src/controls.test.ts` lives at `ui/src/` and runs with `npx vitest run` (and `make test`), not with the folder list of `unit-test-instructions.md`.
