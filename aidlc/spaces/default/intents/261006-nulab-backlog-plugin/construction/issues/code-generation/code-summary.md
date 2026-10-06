# Code Summary — issues (U3)

All 13 plan steps were carried out in order under the TDD Testing Contract (`contract_sha256 sha256:376593af89ee0d9cddcd078826773d6d5ca3215ec737b9af952712bee2395bcc`). Each layer went Red, then Green, then Refactor; the Red output of every layer is below. From `make clean`, `make check-format vet lint test coverage check-secrets build package verify-package` passes. Go line coverage is **92.9%** (floor 80%, only `server/main.go` excluded); `internal/issues` is **91.3%** (target 85%), `internal/backlog` **96.0%** and `internal/plugin` **93.6%** (U4 levels: 95.8% and 93.3%). `go mod tidy` leaves no diff. `make contract-test KANDEV_MIN_DIR=../kandev-min` passes on Kandev v0.96.0.

U3 adds one Go package (`internal/issues`) and no Go module or npm dependency. Only `internal/plugin` imports `pluginsdk`; `internal/issues` imports neither `pluginsdk` nor `internal/git`. `Makefile`, `.github/workflows/*`, `.golangci.yml`, `internal/ci`, `cmd/`, `internal/connection/*`, `internal/git/*`, `go.mod`, `go.sum` and `ui/package.json` were not touched. `README.md` only got a section appended at the end.

## Files

`source-manifest.json` lists every path U3 created or changed (78 paths). Shared files (also used by other units) are marked *shared*.

| Area | Files |
|------|-------|
| BacklogGateway | `internal/backlog/issues.go` (new: `Issue` moved here from `pullrequests.go` and widened with `ProjectID`, `Description`, `StatusID/Name`, `PriorityID/Name`, `AssigneeID/Name`, `DueDate`, `Updated`; `Comment`, `Attachment`, `Status`, `ProjectUser`; `IssueQuery.Values`, `CommentQuery.Values`, `ValidIssueRef`; calls `Issues`, `IssueCount`, `Issue`, `IssueComments`, `IssueAttachments`, `ProjectStatuses`, `ProjectUsers` on the one `send` path); *shared* `client.go` (the old `Issue` method removed), *shared* `pullrequests.go` (`Issue` and `parseIssue` removed); tests `issues_types_test.go`, `issues_client_test.go`, *shared* `pullrequests_types_test.go` (deviation 1); fixtures `testdata/issues_ok.json`, `issues_count_ok.json`, `issue_detail_ok.json`, `comments_ok.json`, `attachments_ok.json`, `statuses_ok.json`, `project_users_ok.json` |
| IssueIntegration | `internal/issues/doc.go`, `types.go` (`ParseIssueKey`, `Query`/`ValidateQuery`, `ShowingRange`, `PriorityFor`, `NewTask`/`NewTaskFor`, `IssueURL`, `ValidatePollMinutes`, `Link` with `Stale`, `Matches`, `SuggestMatch`), `store.go` (workspace documents `issues.links` and `issues.settings`, instance key `issues.index`, `schemaVersion: 1`, 1,000-link cap, per-workspace mutex, 1 s calls), `service.go` (`Gateway`, `Connection`, `HostPort`; list, filters, create, task search, link, unlink, links, detail, comments, suggest, authorize, impact, settings), `sync.go` (`Syncer`: one worker, ticks and `Refresh`, `issue_sync_cycle` log), `events.go` (`Listen`, `OnConnectionChanged`, `ReconcileAll`, `OnTaskDeleted`); tests `types_test.go`, `store_test.go`, `list_test.go`, `filters_test.go`, `create_test.go`, `links_test.go`, `detail_test.go`, `suggest_test.go`, `sync_test.go`, `events_test.go`, `impact_test.go`, `settings_test.go`, `leak_test.go`, helpers `fakes_test.go`, `harness_test.go` |
| KandevAdapter | `internal/plugin/issue_actions.go` (13 handlers merged into `handlers`), `references.go` (`SearchEntityReferences`, `AuthorizeEntityReference`), `events.go` (`OnEvent` for `task.deleted`); *shared* `host_port.go` (U4's `CreateTask` and `FindTaskByMetadata` now use the shared `create`/`eachTask` helpers; new `issueHost` adapter); *shared* `runtime.go` (issues service and syncer wiring, `Start`/`Close`, three `classify` cases, `gateway` embeds `issues.Gateway`); tests `actions_u3_test.go`, `references_test.go`, `events_test.go`, `host_port_u3_test.go`, `runtime_u3_test.go`, *shared* `manifest_test.go` (U3 cases; capability count 4 → 5), *shared* `actions_u4_test.go` (deviation 6) |
| Manifest | *shared* `manifest.yaml`: 13 `issues.*` actions (`max_body_bytes: 16384`), `capabilities.events: ["task.deleted"]`, `reference_sources` (`nulab-backlog-issues` / `nulab-backlog` / `issue`, "Backlog issues" / "Issue") |
| UI | `ui/src/issues/issues-state.ts`, `issues-page.tsx`, `link-task-dialog.tsx`, `links-store.ts`, `issue-badge.tsx`, `task-menu.ts`, `issue-panel.tsx`, `poll-interval.tsx`, `i18n.ts` and eight test files; *shared* `ui/src/index.ts` (catalogue first, `messagesFor(host)` passed to every screen, `task-card-tags` badge, Unlink menu action, task panel), `messages/en.ts` (U3 keys), `page/BacklogPage.tsx` (`/backlog` renders the issue list once connected), `settings/connected-panel.tsx` (sync interval block, issue counts), `settings/project-picker.tsx` and `settings/SettingsScreen.tsx` (issue counts), `settings/confirm-dialog.tsx` (deviation 7), `testing/harness.ts`; tests *shared* `index.test.ts`, `page/backlog-page.test.tsx`, `settings/connected-panel.test.tsx`, `settings/project-picker.test.tsx` |
| Docs | *shared* `README.md` (appended "Backlog issues"), *shared* `docs/manual-checks/TEMPLATE.md` (B4 demo steps 21–26 and a U3 result table) |

## Action schemas (R-04)

Bodies are JSON objects. Task-scoped actions take the task from the verified action context, never from the body.

| Action | Scope / access | Body | Reply |
|--------|----------------|------|-------|
| `issues.list` | workspace / authenticated | `{page?, pageSize?, projectKeys?, statusIds?, assigneeIds?, keyword?}` | `{items: [{issueKey, summary, status, statusId, assignee?, updatedAt, url, linkedTasks: [{taskId, taskKey?}]}], total, page, pageSize, refreshedAt, connectionEpoch}` |
| `issues.filters` | workspace / authenticated | — | `{projects: [{key, name}], statuses: [{id, name}], assignees: [{id, name}]}` |
| `issues.create_task` | workspace / authenticated | `{issueKey, workflowId, workflowStepId?, force?}` | `{taskId, taskKey, issueKey}` |
| `issues.tasks.search` | workspace / authenticated | `{query?}` | `{tasks: [{taskId, taskKey?, title, linkedIssueKey?}]}` (≤ 20) |
| `issues.links.list` | workspace / authenticated | — | `{links: [{taskId, taskKey?, issueKey, spaceHost, state, status?, statusUpdatedAt?, stale, unavailable, url}]}` |
| `issues.refresh` | workspace / authenticated | — | `{updatedCount, refreshedAt}` |
| `issues.impact` | workspace / authenticated | `{projectKeys?}` | `{issueLinks}` |
| `issues.settings.get` | workspace / authenticated | — | `{pollMinutes, lastCycleAt?}` |
| `issues.set_poll_interval` | workspace / **admin** | `{minutes}` (whole number 1–1440) | `{pollMinutes, lastCycleAt?}` |
| `issues.link` | task / authenticated | `{issueKey}` | the link |
| `issues.unlink` | task / authenticated | — | `{ok: true}` |
| `issues.get` | task / authenticated | — | `{issueKey, linkState, statusUpdatedAt?, spaceHost? (not connected only), issue?: {key, summary, status, assignee?, priority?, dueDate?, url}, attachments?: [{name, size, tooLarge}], attachmentsError?}` |
| `issues.comments` | task / authenticated | `{maxId?}` | `{comments: [{id, author, content, created}], nextMaxId?}` (20 per page, newest first) |

Error codes: `validation` (with `field`: `page`, `pageSize`, `projectKeys`, `statusIds`, `assigneeIds`, `keyword`, `issueKey`, `taskId`, `workflowId`, `minutes`, `limit`), `not_found` (unavailable issue, missing link), `conflict` (create while another create of the issue runs, an issue that already has a task without `force`, a task linked to another issue, a connection change during the action), `integration_disabled` (every U3 action while Backlog is off), `reconnect_required`, `rate_limited` (with `Retry-After`), `unreachable`, `internal`.

## Key Decisions

- **Reads only.** The U3 gateway interface has only GET calls; a test wrapper fails any non-GET request to the fake Backlog (AC3.2.3, FR4.3). Bad issue references and project keys are refused before a path is built.
- **Rate-limit groups.** The issue list and count are Search (queued per host, 1 s apart); everything else is Read. The list is Interactive (3 s budget) and the sync cycle Background, so a 429 in the cycle never blocks the list. With a 300 ms Backlog the list reply is ready in 1.9 virtual seconds (AC8.1.1).
- **Project ids cached** per (workspace, connection epoch) from one `Projects` call; a query without a resolved project id is refused rather than sent (Backlog would search every project).
- **Sync schedule in IssueIntegration's own state** (`issues.settings`), so changing the interval never raises the connection epoch. One worker ticks every minute and runs each workspace whose `lastCycleAt + pollMinutes` has passed; a restart continues the schedule. `Refresh` runs on the same worker and waits at most 10 s, so polls never overlap.
- **One cycle per workspace:** switch check (off → skip; unreadable → error, fail closed), `Current()` (not connected → skip), one task list (drops links of deleted tasks, refreshes task keys), one `Issue` GET per linked issue, then an epoch check that drops late results. 404/403 → `unavailable`, other failures count up to "may be out of date" after 3. Exactly one `issue_sync_cycle` line per cycle that ran a workspace or hit an error.
- **Links follow the connection** with U4's rules: disconnect, another host or a deselected project → `not_connected`; a restore or a reconciliation with `Current()` brings covered links back; older events are ignored. `task.deleted` removes links and the cycle re-checks, since delivery is best effort.
- **`#` references.** Suggestions coalesce per workspace (a newer query within 250 ms wins), stay inside 1.2 s, and every failure is an empty list. Authorization allows only this source/provider/kind and a live `Issue` 200 in a selected project, and answers denials with one fixed reason.
- **UI.** The issue list, the Link to task dialog, the card badge, the task menu action and the task panel use the host `Button`, `Input`, `Label` and `Skeleton`, native `table`, `select`, `details` and buttons. Every interactive element has a `data-testid`; axe reports no violations; every string comes from the catalogue, which is registered with `registerTranslations` and read through `host.i18n.t` with the English text as the default.

## Red evidence

Each Red run happened before the production code of its layer existed. Key lines are trimmed.

**Step 1, runner readiness.** `go test -race ./internal/backlog/... ./internal/plugin/... -run '^TestU3_'` → `ok … [no tests to run]` for both; with `./internal/issues/...` added (after `doc.go`): `[no test files]`. The eight-file Vitest command with `--passWithNoTests` ran clean. Regression: Go `backlog`, `connection`, `git`, `plugin` ok; Vitest 19 files, 161 tests passed.

**Step 2, data model.** `go test -race ./internal/backlog/... ./internal/issues/... -run '^TestU3_'`

```
internal/issues/types_test.go:38:23: undefined: ParseIssueKey
internal/issues/types_test.go:55:9: undefined: Query
internal/issues/types_test.go:62:47: undefined: FieldPage
internal/issues/types_test.go:63:57: undefined: FieldPageSize
internal/backlog/issues_types_test.go:12:17: undefined: parseIssues
internal/backlog/issues_types_test.go:15:35: unknown field ProjectID in struct literal of type Issue
internal/backlog/issues_types_test.go:16:61: unknown field StatusID in struct literal of type Issue
FAIL	.../internal/backlog [build failed]
FAIL	.../internal/issues [build failed]
```

**Step 4, gateway and data access.** Same command.

```
internal/issues/store_test.go:19:7: undefined: NewStore
internal/issues/store_test.go:29:59: undefined: Settings
internal/issues/store_test.go:88:87: undefined: maxLinks
internal/backlog/issues_client_test.go:37:15: c.Issues undefined (type *Client has no field or method Issues)
internal/backlog/issues_client_test.go:41:15: c.IssueCount undefined (type *Client has no field or method IssueCount)
internal/backlog/issues_client_test.go:57:15: c.ProjectStatuses undefined (type *Client has no field or method ProjectStatuses)
FAIL	.../internal/backlog [build failed]
FAIL	.../internal/issues [build failed]
```

During Step 6 the panel needed the priority name: `PriorityName` was added to the expected `Issue` in `issues_types_test.go` first (`unknown field PriorityName in struct literal of type Issue`, build failed), then to `backlog.Issue`.

**Step 6, business logic.** `go test -race ./internal/issues/... -run '^TestU3_'`

```
internal/issues/harness_test.go:293:14: undefined: TaskInfo
internal/issues/harness_test.go:301:63: undefined: TaskRef
internal/issues/harness_test.go:351:9: undefined: Service
internal/issues/create_test.go:13:27: undefined: CreateInput
internal/issues/create_test.go:45:33: undefined: ErrConflict
internal/issues/suggest_test.go:48:16: undefined: Candidate
internal/issues/sync_test.go:19:36: undefined: Syncer
FAIL	.../internal/issues [build failed]
```

**Step 8, API / endpoint.** `go test -race ./internal/plugin/...`

```
internal/plugin/actions_u3_test.go:290:2: undefined: actionIssuesList
internal/plugin/actions_u3_test.go:290:20: undefined: actionIssuesFilters
internal/plugin/actions_u3_test.go:290:41: undefined: actionIssuesCreate
internal/plugin/actions_u3_test.go:291:44: undefined: actionIssuesLink
FAIL	.../internal/plugin [build failed]
```

**Step 10, frontend behaviour.** From `ui/`: the eight-file U3 Vitest command, then the shared files U3 changes.

```
 FAIL  src/issues/i18n.test.ts
Error: Failed to resolve import "./i18n" from "src/issues/i18n.test.ts". Does the file exist?
 FAIL  src/issues/issues-page.test.tsx
Error: Failed to resolve import "./issues-page" from "src/issues/issues-page.test.tsx". Does the file exist?
 …(the same for the other six files)
 Test Files  8 failed (8)

$ npx vitest run src/settings/connected-panel.test.tsx src/settings/project-picker.test.tsx src/page/backlog-page.test.tsx src/index.test.ts
     × registers the U3 catalogue first, the card badge, the task menu and the task panel
     × says how many issue links, PR links and watches of the unselected projects are turned off (AC1.9.1)
     × lists the issues once connected, and not otherwise (U3)
     × says how many issue links, PR links and watches a disconnect turns off (AC1.8.1)
     × says how many issue links, PR links and watches a space change turns off
     × mounts the sync interval block (M1, US4.2)
 Test Files  4 failed (4)
      Tests  6 failed | 43 passed (49)
```

Each Green step then made its layer pass. Refactor kept it green, for example: `load` without an unused `found` result, `slices.ContainsFunc` instead of a local helper, the U4 host adapter rebuilt on the shared `create`/`eachTask` helpers, and `issueNotice` split from the list-state mapping. Green also fixed three test-side mistakes found on the first run (a schema test counting writes of the other document, test writes to fakes racing the worker under `-race`, and a disconnect test whose `Current()` still said connected).

## Test, coverage and contract results

- **Go:** 120 new `TestU3_` tests (backlog 17, issues 80, plugin 23), all under `-race`; `-count=3` over `internal/issues` is stable. Every underscore part of a test name is under 32 characters. Timer tests (sync, refresh, suggestion coalescing and deadline, list budget, store limit, queue spacing, 429 and timeout) run in `testing/synctest` with no real sleep.
- **UI:** 56 tests in the eight `ui/src/issues` files (issues-state 7, issues-page 12, link-task-dialog 6, issue-badge 6, task-menu 5, issue-panel 10, poll-interval 5, i18n 5) plus 3 new tests in `index.test.ts`, `backlog-page.test.tsx` and `connected-panel.test.tsx` (three U2/U4 tests were updated for the issue counts). The full Vitest run is 27 files, 220 tests, all green. `tsc --noEmit`, ESLint and Prettier are clean.
- **Gate (from `make clean`):** `make check-format vet lint test coverage check-secrets build package verify-package` → exit 0. golangci-lint (with gosec): `0 issues.` `coverage: 92.9% (floor 80%, excluded: server/main.go)`; per package: backlog 96.0%, ci 91.0%, connection 94.5%, git 90.8%, issues 91.3%, pkgverify 93.2%, plugin 93.6%, redact 97.4%, testutil 87.0%. `ci secrets: OK`. `verifypkg: OK dist/nulab-backlog-0.0.1.tar.gz (nulab-backlog@0.0.1)`. The packaged `manifest.yaml` carries the 13 `issues.*` actions, `events: ["task.deleted"]` and `reference_sources`.
- **`go mod tidy`:** no diff.
- **Contract:** `make contract-test KANDEV_MIN_DIR=../kandev-min` → `ci contract: OK nulab-backlog on Kandev v0.96.0`; Kandev v0.96.0 accepted the manifest with the reference source and the `task.deleted` subscription and started the plugin. `../kandev` and `../kandev-min` have no tracked changes.

## Deviations

1. **`TestU4_IssueParse_DecodesTheIssue` changed.** It compared the whole `backlog.Issue`, so widening the struct broke it; it now compares the three fields U4 uses (`ID`, `IssueKey`, `Summary`). The plan expected U4's tests to stay green unchanged.
2. **`backlog.Issue.PriorityName`** was added (the panel shows Backlog's priority name, M8). `Issue` signatures are unchanged.
3. **`issues.Store` takes no `Now`**: it stores what the service stamps. The service has `Now` and `Wait`; the syncer has `Tick`.
4. **`registerMessages(registry)`** takes no host; `messagesFor(host)` falls back to English on a host without `i18n`.
5. **Link stores the task key** found with one paged task list (`ponytail` note); the cycle refreshes it.
6. **`internal/plugin/actions_u4_test.go`** (not on the shared list): the fake `Tasks().Create` now also returns `Identifier` `T-<n>`, as Kandev does. U4's tests are unchanged otherwise.
7. **`ui/src/settings/confirm-dialog.tsx`** (not on the shared list): an optional `children` slot under the body and links in the Tab trap, for the three-choice dialog's "Open T-17" link. U2's tests are unchanged.
8. **The task panel is offered only for linked tasks** (`visible` from the shared links store), so `not_found` in the panel means an unavailable issue.
9. **`issues.unlink` of a task without a link** answers `not_found`. **`issues.comments`** of a not-connected link answers `reconnect_required`.
10. **The cycle log line** is written for cycles that ran at least one workspace or hit an error, not for every idle 1-minute tick.
11. **`IssueCount`** sends only the filters (no `sort`, `order`, `offset`, `count`).
12. **UI filters are single-choice** per project, status and assignee (sent as one-element arrays); the action accepts up to 50 ids.
13. **The dialog impact line** is "N issue links, M PR links and K PR watches will be turned off." whenever any count is above 0, so it can say "0 issue links". It lives in `ui/src/issues/issues-state.ts` (`loadImpactText`); U4's `loadImpact` in `ui/src/git/git-state.ts` is no longer called by the settings screens but was left in place (U4 file).
14. **A body of the wrong JSON types** (for example `{"page":"one"}`) is `validation` without a `field`, as in U4.
15. **`ui/src/issues/issues-page.tsx` imports `settingsHref`** from `ui/src/page/BacklogPage.tsx`, which imports the issue page: an import cycle used only at run time.

## Upstream amendments (for the contract and functional docs)

- **C1:** `Issues(ctx, creds, class, IssueQuery)`, `IssueCount(ctx, creds, class, IssueQuery)`, `Issue(ctx, creds, class, ref)` (wider type with `PriorityName`; invalid refs refused), `IssueComments(ctx, creds, class, ref, CommentQuery{MaxID, Count})`, `IssueAttachments(ctx, creds, class, ref)`, `ProjectStatuses(ctx, creds, projectKey)`, `ProjectUsers(ctx, creds, projectKey)`.
- **C2:** `HostPort` = `CreateTask(ctx, NewTask) (TaskRef{ID, Key}, error)` and `ListTasks(ctx, ws) ([]TaskInfo{ID, Key, Title}, error)`; `SetTaskLabels` and `TaskExists` removed (labels are the plugin's `task-card-tags` badge; deleted tasks come from `task.deleted` and the cycle's task list).
- **C3:** `PollMinutes` leaves `Snapshot`; the interval is IssueIntegration's `issues.settings`.
- **C5:** the action table above; `connection.setPollInterval` → `issues.set_poll_interval` (admin).
- **C8:** `capabilities.events: ["task.deleted"]`, `reference_sources` (`nulab-backlog-issues`), the `task-card-tags` slot, `registerTaskPanel` (`backlog-issue`), `registerTaskMenuAction` (`backlog-unlink-issue`), `registerTranslations`.

## Open items

- AC8.1.2, AC8.2.1 and AC8.2.4 are `[manual]` and `Deferred` to the B4 demo (`docs/manual-checks/TEMPLATE.md` steps 25–26).
- Unverified against a real space: that Backlog's `keyword` matches issue keys (the exact-key GET covers full keys), and that Kandev's `task.deleted` payload carries `task_id` with the workspace on the event (read from v0.96.0 source, not observed live).
- English only for the first release (NFR10); other locales fall back to English.
- The links store refreshes every 60 s while a badge is mounted and on window focus (see "Review iteration 1 repairs"). The sync makes one `Issue` GET per linked issue per cycle (`ponytail`: batch with `id[]` past about 200 links).
- R-01 (OAuth webhook lock growth) and R-02 (`SetProjects` epoch race) from the U2 review are unchanged; scheduled for Build and Test.

## Review iteration 1 repairs

Review 1 (`reviews/review-01.md`) returned NOT-READY. R-01, R-02 and R-03 are repaired here, each with the failing test written first. R-04, R-05, R-07 and the server-side part of R-06 are deferred to Build and Test as directed. The R-06 retry-on-render part is covered by the R-03 backoff.

### R-01 and R-02: the sync cycle no longer deletes links

- **Not-found detection.** Kandev v0.96.0 reports a missing task from `Tasks().Get` only as a gRPC `codes.NotFound` status error (`internal/plugins/host.go` `taskNotFound`, and `pkg/pluginsdk/host.go` `GetTask`). The SDK has no sentinel or helper for it. `google.golang.org/grpc` is only an `// indirect` requirement, and a probe import of `grpc/status` in `internal/plugin` made `go mod tidy` move it to the direct block, which is a `go.mod` diff. So the per-link `Get` confirmation could not be built within the rules. As the brief allows, the cycle now never deletes a link. `task.deleted` (`internal/plugin/events.go` > `OnTaskDeleted`) is the only removal.
- **Cycle behaviour** (`internal/issues/sync.go` > `syncWorkspace`). The cycle still lists tasks once to refresh task keys. A link whose task is missing from that list is kept and is not polled this cycle, so a deleted task's issue is not read. It is counted in a new `missingTasks` field of the single `issue_sync_cycle` line, which logs it once per cycle. Because nothing is deleted, a link written while the cycle waits on Backlog survives (R-01).
- **Ephemeral tasks** (`internal/plugin/host_port.go` > `issueHost.ListTasks`). `TaskFilter.IncludeEphemeral` is now `true`, so links to ephemeral tasks get task keys and are polled. `eachTask` now takes the `TaskFilter`. U4's `FindTaskByMetadata` passes its unchanged filter (`IncludeArchived` only).
- **Tests.** `TestU3_Sync_KeepsLinksMadeMidCycle` (synctest: the task and link are created while the cycle is blocked on a Backlog GET, after the task list was read) and `TestU3_Sync_KeepsLinksOfMissingTasks` (replaces `TestU3_Sync_PrunesDeletedTasks`: both links kept, the key refreshed, the missing task's issue not read, `missingTasks: 1` in each of two cycle lines) in `internal/issues/sync_test.go`. `TestU3_HostPort_ListsEphemeralTasks` is in `internal/plugin/host_port_u3_test.go`. The brief's `Get` cases (present on Get, definite not-found, Get error) do not apply because the cycle makes no `Get` call.
- **Red.** `KeepsLinksOfMissingTasks`: `should have 2 item(s), but has 1`. `KeepsLinksMadeMidCycle`: `should have 3 item(s), but has 2`. `ListsEphemeralTasks`: `IncludeEphemeral: true` expected, `false` actual.

### R-03: the card badge refreshes

- **`ui/src/issues/links-store.ts`.** While a workspace has at least one subscriber (a mounted badge), the store reloads `issues.links.list` every **`LINKS_REFRESH_MS` = 60 s**, a fixed interval. It is the server sync's 1-minute tick and minimum poll interval, so a synced status shows within about one minute of the cycle without an extra `issues.settings.get` call. It also reloads on window `focus` and on `visibilitychange` to visible. The timer and the window listeners stop when the last subscriber leaves. A request in flight is shared by `load`, the timer and focus. `refresh` (after the plugin's own link or unlink) still forces a new request.
- **Backoff (also R-06, UI part).** After a failure, `load`, the timer and focus make no request until the backoff passes: 60 s, then 120 s, doubling up to 10 minutes, reset on success. The task menu's `visible()` and every badge render no longer start a request each time. `load` also reloads data older than 60 s, so the menu does not show very old links when no badge is mounted.
- **Tests.** `ui/src/issues/links-store.test.ts` (new, fake timers, 4 tests): a status change on the fake shows after 60 s and not at 59.999 s; no request after unmount, even on focus; focus and visibility changes refresh and share one request in flight; a failing load is called once for ten `visible()` calls, one render and a focus, then retried at 60 s, then waits 120 s.
- **Red.** 3 of the 4 tests failed on the old store, for example `expected 2 to be 1` for the failure case. The unmount test passed on the old store too; it guards against a timer that outlives its badges.
- Not done: refreshing the store when `issues.refresh` returns on the issue list page. The 60 s timer and focus cover it.

### Files changed in this repair

- U3-owned: `internal/issues/sync.go`, `internal/issues/sync_test.go`, `internal/issues/doc.go` (only `gofmt`: a stray leading blank line that `make check-format` reported; I did not make that change), `internal/plugin/events.go` (doc comment only), `internal/plugin/host_port_u3_test.go`, `ui/src/issues/links-store.ts`, `ui/src/issues/links-store.test.ts` (new; added to `source-manifest.json`).
- *Shared*: `internal/plugin/host_port.go` (the `eachTask` signature, the U4 `FindTaskByMetadata` call site with the same filter as before, and `issueHost.ListTasks`).
- Not touched: `go.mod`, `go.sum`, `manifest.yaml`, `Makefile`, the plan, `code-generation-questions.md`, `.aidlc-engine/`. `traceability.json` is unchanged because no IDs changed.

### Results

- `make check-format vet lint test coverage check-secrets` → exit 0. golangci-lint `0 issues.`; Vitest 28 files, 224 tests passed; `coverage: 92.9% (floor 80%, excluded: server/main.go)`; `internal/issues` 91.3%, `internal/plugin` 93.6%, `internal/backlog` 96.0%; `ci secrets: OK`.
- `make build package verify-package` → `verifypkg: OK dist/nulab-backlog-0.0.1.tar.gz (nulab-backlog@0.0.1)`.
- `go mod tidy` → no diff in `go.mod` or `go.sum`. `go test -race -count=3 ./internal/issues/ -run TestU3_Sync` passes.
