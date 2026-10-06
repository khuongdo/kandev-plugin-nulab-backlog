## Review

**Verdict:** READY
**Reviewer:** aidlc-architecture-reviewer-agent
**Date:** 2026-10-06T22:00:20Z
**Iteration:** 2

### Findings

| ID | Severity | Location | Finding | Required action | Status |
|---|---|---|---|---|---|
| R-01 | Critical | internal/issues/sync.go > syncWorkspace | No sync path deletes links any more. UpdateLinks in sync.go only edits fields (TaskKey, status, FailCount, Unavailable) and appends every link back. The only deleters are OnTaskDeleted (events.go) and the explicit unlink in service.go. A link written mid-cycle therefore survives. The tests TestU3_Sync_KeepsLinksMadeMidCycle and TestU3_Sync_KeepsLinksOfMissingTasks cover this, and the suite passes under -race. | None. | Resolved |
| R-02 | Major | internal/issues/sync.go (missing, polled); internal/plugin/host_port.go > ListTasks | Pruning by absence is gone. A link whose task is not in the list is kept, not polled this cycle, and counted in missingTasks. ListTasks now sets IncludeEphemeral: true (TaskFilter.IncludeEphemeral exists in ../kandev pkg/pluginsdk/data_types.go:426). The not-found claim holds: taskNotFound in internal/plugins/host.go:139 returns status.Errorf(codes.NotFound), and pluginsdk exports no sentinel or helper, so the SDK gives no way to read it without importing grpc/status. Residual: if a task.deleted event is missed (plugin stopped or event dropped), its link stays. It is never polled and never shown, because the badge only renders for an existing task. It counts toward the 1,000-link cap and is visible only in the missingTasks log field. AC2.3.2 is met on the best-effort event path only. That is a safe failure (no data loss, bounded by the cap) and a better trade than deleting live links, so the residual is recorded as Accepted risk. | Record the missed-event leak as an accepted risk in code-summary.md. Optionally add an explicit not-found check or a manual purge in Build and Test. | Accepted risk |
| R-03 | Major | ui/src/issues/links-store.ts | The store now refreshes every 60 s while at least one listener is subscribed, and on focus or visibility. In-flight requests are shared. A failure backs off from 60 s, doubling, to 10 min. Timers are cleared when the last listener leaves, and the finally handler does not re-arm without listeners. The focus and visibility listeners are removed once no listener is active. No leak found. links-store.test.ts covers the behaviour. | None. | Resolved |
| R-04 | Minor | internal/issues/events.go (reconcile, no epoch check); sync.go (polled); service.go (Detail, Comments) | Not changed; the user deferred it to Build and Test. | Skip a reconcile whose epoch is below lastEpoch. Check host and selected project in Detail, Comments and polled. Reconcile once when the syncer starts. | Unresolved |
| R-05 | Minor | internal/issues/service.go > Comments; internal/backlog/issues.go > CommentQuery | Not changed; the user deferred it to Build and Test. The maxId boundary is still unverified. | Check the boundary in the manual Backlog check, or drop the one-extra-row trick. Skip comments with empty content. | Unresolved |
| R-06 | Minor | internal/issues/service.go (taskKey, SearchTasks) | The failed-load retry on render is fixed by the new backoff in links-store.ts. The part about full host task scans on every Link and every issues.tasks.search keystroke is not fixed. | Add a debounce or cache the task keys. | Unresolved |
| R-07 | Minor | internal/issues/sync.go (one Issue GET per link), maxLinks=1000 | Not changed; the user deferred it to Build and Test. | Use id[] batches, enforce a per-cycle cap, or raise the minimum interval for large link sets. | Unresolved |

### Validation Tool Results

| Tool | Result | Interpretation |
|---|---|---|
| gofmt -l internal server cmd | clean | No formatting issues. |
| go vet ./... | clean | No findings. |
| go test -race -cover ./internal/... ./server/... | PASS; issues 91.3%, plugin 93.6% | No regression. Both coverage figures are above the 80% floor. Some packages were served from the test cache. |
| go run ./cmd/ci secrets -root . | OK | No credentials in the repo. |
| git status --short before and after | identical | The review did not modify the workspace. |

### Summary

The destructive sync paths from iteration 1 are removed, and the badge store now refreshes without leaking timers. U4's FindTaskByMetadata keeps the same filter (Archived only, workspace-scoped) and behaves as before. The only open architectural cost is that links of deleted tasks are removed solely by the best-effort task.deleted event. I judged that acceptable and recorded it as an accepted risk. The four Minor findings stay open for Build and Test.
