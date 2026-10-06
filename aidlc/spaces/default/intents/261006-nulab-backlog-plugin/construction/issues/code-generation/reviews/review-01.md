## Review

**Verdict:** NOT-READY
**Reviewer:** aidlc-architecture-reviewer-agent
**Date:** 2026-10-06T14:20:57Z
**Iteration:** 1

### Findings

| ID | Severity | Location | Finding | Required action | Status |
|---|---|---|---|---|---|
| R-01 | Critical | internal/issues/sync.go > syncWorkspace (`exists` snapshot, `gone`, final `UpdateLinks` callback) | The task list is read once at the start of the cycle, but `gone(l)` is applied inside the final `UpdateLinks` callback to links re-read under the lock. A link written by `issues.create_task` or `issues.link` while the cycle runs (the cycle makes one Backlog GET per link, so it can last many seconds) refers to a task that is missing from the stale `exists` map and is silently deleted. The user's new link vanishes, and the plan accepts no automatic relink. No test covers a link created during a cycle (sync_test.go has only PrunesDeletedTasks). | Only prune links whose `CreatedAt` is before the time the task list was read, or take the list under the same guard as the link write. Add a regression test that creates a task and link while a cycle is blocked on a Backlog call. | New |
| R-02 | Major | internal/plugin/host_port.go > issueHost.ListTasks / eachTask; internal/issues/sync.go (pruning on a missing task) | Pruning by absence from `Tasks().List` is destructive but the list is not a reliable inventory. Verified in ../kandev: the host sets `excludeConfig` true and `IncludeEphemeral` defaults false, so ephemeral and config tasks never appear. Paging uses an offset cursor over a list rebuilt on every call (host_data.go paginate), so a task deleted between two pages shifts the list and skips an entry. A skipped or hidden task has its link deleted permanently. The plan chose List over `Tasks().Get` NotFound only to keep go.mod unchanged. | Prune a link only after it is missing in two consecutive cycles, or confirm with a per-task read before deleting. Pass `IncludeEphemeral` or document that ephemeral tasks cannot be linked. Add a test where a page boundary shifts. | New |
| R-03 | Major | ui/src/issues/links-store.ts and issue-badge.tsx (loaded once per workspace, no polling) | The card badge reads `issues.links.list` once and refreshes only on the plugin's own unlink or link actions. The server sync updates `LastKnownStatus`, but an open Kanban board never shows it until a page reload. That defeats US4.1 (status on the card) and the purpose of the sync worker. The plan's `ponytail:` note covers other users' links, not status changes. | Re-fetch on a timer of about the poll interval, or when the tab regains focus, and when `issues.refresh` returns. Add a test with fake timers. | New |
| R-04 | Minor | internal/issues/events.go (`reconcile`, no epoch check); sync.go (`polled`); service.go (`Detail`, `Comments`) | `reconcile` applies a snapshot read earlier without comparing it to `lastEpoch`. A disconnect that lands between the read and the apply can set links back to active until the next tick. `polled`, `Detail` and `Comments` trust the stored link state and do not recheck host and selected project against the current snapshot. The first reconcile happens one minute after start, not at startup as the plan says. Credentials are always the current connection's, so the window is short, but an issue key from the old space can be fetched in it. | Skip a reconcile whose epoch is below `lastEpoch`. Check host and selected project in `Detail`, `Comments` and `polled`. Reconcile once when the syncer starts. | New |
| R-05 | Minor | internal/issues/service.go > Comments; internal/backlog/issues.go > CommentQuery | The next page uses the 21st comment's id as `maxId`, which assumes Backlog treats `maxId` as inclusive. The fake test only asserts the query string, so the boundary is unverified and a comment could be skipped or repeated. Comments with null content (status-change-only) render as blank rows. | Check the boundary in the manual Backlog check, or drop the one-extra-row trick. Skip comments with empty content. | New |
| R-06 | Minor | ui/src/issues/links-store.ts (`load` after a failure); internal/issues/service.go (`taskKey`, `SearchTasks`) | A failed `issues.links.list` (for example while the switch is off) leaves the store empty, so every menu `visible()` call and every badge mount issues a new request. `Link` and every `issues.tasks.search` keystroke run a full host task scan, and the host rebuilds the whole list on each page. | Cache the failure for a short time. Add a debounce, or cache the task keys. | New |
| R-07 | Minor | internal/issues/sync.go (one Issue GET per link) with maxLinks=1000 and a 1-minute minimum interval | The worst case is 1000 Read-group GETs per minute, above Backlog's documented per-minute limit. Background retries are unbounded, and the interactive `#` Authorize has a 1.2 s budget, so it can be denied during a cycle. The `ponytail:` threshold of about 200 links is not enforced anywhere. | Use `id[]` batches, enforce a per-cycle cap, or raise the minimum interval for large link sets. | New |

### Validation Tool Results

| Tool | Result | Interpretation |
|---|---|---|
| gofmt -l internal server cmd | clean | No formatting issues. |
| go vet ./... | clean | No findings. |
| go test -race -cover ./internal/... ./server/... | PASS, internal/issues 91.3%, internal/plugin 93.6%, backlog 96.0% | Green, but the tests do not cover the R-01 and R-02 races. |
| go run ./cmd/ci secrets -root . | OK | No committed secrets. |
| git status --short before and after | identical | The workspace was not modified. |

**Verified against Kandev v0.96.0 (no finding):**
- The `task.deleted` payload carries `task_id` and `workspace_id`, and the delivered `Event.WorkspaceID` is taken from `workspace_id`.
- `Task.Identifier` is assigned on create. `TaskFilter.IncludeArchived` exists. The `PageInfo` cursor is an offset string.
- The `EntityReferenceHandler` signatures match. The host builds the reference with `key` and `provider`, and its authorize timeout is 1.5 s.
- The `reference_sources` identifiers satisfy the host's validation pattern. `capabilities.events` accepts `task.deleted`.
- The `task-card-tags` slot props are `taskId`, `workspaceId` and `workflowStepId`. `registerTaskPanel` and `registerTaskMenuAction` match the SDK types.
- The Backlog testdata matches the API v2 shapes.
- No non-GET call is made. Issue refs are regex-validated before they reach a path. All issue actions go through the switch guard, and `Suggest`, `Authorize` and the syncer check the switch themselves. The create lock is per (workspace, issue key) and covers the link write.

**Developer deviations:**
- Items 1-4 and 6-9 are acceptable. The U4 test narrowing keeps the checks for the fields U4 uses, and `PriorityName` is additive.
- Item 5 (the `a[href]` tab order) is correct.
- Item 7 (the cycle log written only when work happened) is acceptable.
- Item 9 (U4 `loadImpact` left unused) is acceptable, but it is dead code. Remove it or lint will flag it.

### Summary

The wiring to Kandev matches the real surfaces and the switch, selected-project, non-GET and path-injection checks hold. The blocker is the sync worker's destructive pruning. R-01 deletes freshly created links in normal use, and R-02 deletes links from an unreliable task list. R-03 (the board never shows polled status) is the other significant gap.
