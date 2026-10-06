## Review

**Verdict:** READY
**Reviewer:** aidlc-architecture-reviewer-agent
**Date:** 2026-10-06T22:20:03Z
**Iteration:** 1

### Findings

| ID | Severity | Location | Finding | Required action | Status |
|---|---|---|---|---|---|
| R-01 | Critical | internal/plugin/host_port.go > FindTaskByMetadata | Nested task metadata fix is intact. The refactor into eachTask keeps the filter (WorkspaceIDs plus IncludeArchived) and the plugin-namespace nesting lookup (metadataNamespace, then key). CreateTask still passes in.Metadata unchanged. | None. | Resolved |
| R-02 | Critical | internal/plugin/host_port.go, internal/git | Repo name from ProviderName is unaffected: the diff does not touch the Repository adapter or git code. | None. | Resolved |
| R-03 | Major | internal/git/watcher.go, internal/plugin/runtime.go | The watcher still obeys the per-workspace switch (RequireEnabled in internal/git at the service calls). runtime.go only adds the issues service and syncer beside the watcher; Start and Close keep the watcher lifecycle. | None. | Resolved |
| R-04 | Minor | internal/git (reconcile) | Reconcile still has no epoch guard. Deferred to Build and Test. | Add an epoch guard in Build and Test. | Unresolved |
| R-05 | Minor | internal/plugin/runtime.go > Close vs watcher | Close still swaps r.watcher (and now r.syncer) under lifeMu while other goroutines may read it, so the Close vs watches.run race remains. The new syncer swap follows the same pattern. | Guard or snapshot r.watcher and r.syncer reads in Build and Test. | Unresolved |
| R-06 | Minor | internal/backlog/pullrequests.go, PR status rendering | An unknown PR status still renders as "Open". | Render an unknown status explicitly. | Unresolved |
| R-07 | Minor | internal/git (RunQuery) | RunQuery still does not check the selected project. | Check the selected project in RunQuery. | Unresolved |

### Validation Tool Results

| Tool | Result | Interpretation |
|---|---|---|
| git diff c3ca400 (5 U4-touched files) | Only U3 additive changes (issueHost adapter, eachTask with TaskFilter, Issue moved to issues.go, runtime wiring, classify cases). | U4 behaviour intact. The narrowed TestU4_IssueParse test still asserts ID, IssueKey and Summary, the 3 fields U4 uses, and the parse error on a missing id. The fake Create's added Identifier is harmless. |
| go vet ./... | PASS (no output) | Clean. |
| go test -race -cover ./internal/... ./server/... | PASS, all packages 87-97% (git 90.8%, plugin 93.6%); server has no statements (wiring only) | The 80% floor holds. |
| git status --short before and after | Identical; no coverage.out | Workspace unchanged. |

### Summary

U3's edits to the U4-touched files are additive and keep U4 behaviour, including the R-01..R-03 fixes. Vet and race tests are green with coverage above the floor. Only the four known Minor items remain, deferred to Build and Test.
