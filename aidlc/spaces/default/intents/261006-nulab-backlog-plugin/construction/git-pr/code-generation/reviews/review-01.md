## Review

**Verdict:** NOT-READY
**Reviewer:** aidlc-architecture-reviewer-agent
**Date:** 2026-10-06T12:59:11Z
**Iteration:** 1

### Findings

| ID | Severity | Location | Finding | Required action | Status |
|---|---|---|---|---|---|
| R-01 | Critical | internal/plugin/host_port.go > hostPort.FindTaskByMetadata (line 59); test fake internal/plugin/actions_u4_test.go:123 | Verified against Kandev v0.96.0. `pluginTaskMetadata` (apps/backend/internal/plugins/host_write.go:411-418) stores plugin metadata nested as `metadata["plugin:<pluginID>"] = {...}` next to `source`. `PublicTaskMetadata` (host_data_mappers.go:126) does not flatten it. The code reads `t.Metadata[key]` at the top level, so it never matches a real task. A reserved-but-unconfirmed ledger entry (crash between create and id store) is therefore recreated, giving duplicate tasks and breaking AC6.2.3 and AC6.2.4. The fake stores metadata flat, so no test catches it. | Look the key up under `t.Metadata["plugin:"+pluginID]` (a map), or scan every nested map. Change the fake to nest the metadata the way Kandev does. Add a test that reserves an entry, creates a task through the nested fake, and recovers it without a duplicate. | New |
| R-02 | Critical | internal/plugin/host_port.go > hostPort.Repository (line 85, `Name: r.Name`) | Verified. For provider repositories Kandev sets `Repository.Name = "<owner>/<providerName>"` (internal/task/service/service_resources.go:1256). The SDK `pluginsdk.Repository` exposes the real name as `ProviderName` (data_types.go:779, mapped at host_data_mappers.go:274). The adapter maps `Name` and drops `ProviderName`. Backlog repo paths for CreatePR, `repo#n` and `#42` short refs are then built from `owner/web-app`, so AC5.2.1, AC5.3.2 and AC5.3.4 fail on a real host. The fake uses `Name: "web-app"`, which hides it. | Map `ProviderName` as the Backlog repository name, falling back to `Name` only for non-provider sources. Make the fake return Kandev-shaped data (`Name: "owner/web-app"`, `ProviderName: "web-app"`) and add a test. | New |
| R-03 | Major | internal/git/watcher.go > cycleAll (111-138), cycleWatch (199-215), Run (90-103); connection.Service.SetEnabled | Verified. The PR watcher never calls `RequireEnabled` and `SetEnabled` emits no ConnectionChanged. With the Backlog switch off, every cycle still calls Backlog (5-minute tick or manual run) and still creates tasks. This breaks U1 BR7.3 and NFR3.9: the switch must stop calls and writes, and fail closed when it cannot be read. `credentials`/`Current` do not check the switch. | Call `conn.RequireEnabled` at the start of `cycleWatch` (skip, writing nothing, on `ErrIntegrationDisabled`; fail closed on a store error) and again before each task creation. Reject `Run` when disabled. Test: switch off, then no Backlog call and no task. | New |
| R-04 | Minor | internal/git/events.go > reconcile (59-68) / apply with restore=true | Verified. `OnConnectionChanged` guards against stale events with `lastEpoch`, but `reconcile` takes a snapshot and applies it with restore=true without re-checking the epoch. A disconnect landing between `Current` and `apply` lets the stale snapshot reactivate links and watches for up to one cycle. The effect is bounded because `cycleWatch` re-checks the connection and the epoch before writing. Medium downgraded to Minor. | Re-check the epoch (or hold the same lastEpoch guard) around the restore in `reconcile`, or document the one-cycle window as accepted. | New |
| R-05 | Minor | internal/plugin/runtime.go > Close (162-164) vs git_actions.go:173 (`r.watcher.Run`) | Verified. `Close` replaces `r.watcher` under `lifeMu`. The `watches.run` action reads `r.watcher` without the lock, which is a data race when Close overlaps a request. Shutdown only, but the team posture requires `-race` cleanliness. | Read the watcher under `lifeMu` (or keep one watcher and make it restartable). | New |
| R-06 | Minor | internal/git/service.go:464-468 (status mapping) | Verified. An unknown PR status id renders as "Open" instead of "Status unknown". PR statuses 2 and 3 are unverified on a real space, so a wrong guess is shown as fact. | Map an unknown id to state "unknown" with the label "Status unknown". | New |
| R-07 | Minor | internal/git/service.go:865-891 > RunQuery | Verified. The saved query's project is not checked against the current selection or space host. A deselected project can still be queried. | Return the Not connected result when the query's project or host is not covered by the snapshot. | New |

### Validation Tool Results

| Tool | Result | Interpretation |
|---|---|---|
| gofmt -l internal server cmd | PASS (no output) | Formatting is clean. |
| go vet ./... | PASS | No vet findings. |
| go test -race -cover ./internal/... ./server/... | PASS | Every package is above the 80% floor (git 90.7%, plugin 93.1%). R-01 and R-02 pass anyway because the fakes use shapes Kandev does not produce. |
| go run ./cmd/ci secrets -root . | PASS (secrets: OK) | No secrets in the repo. |
| git status --short before and after | Identical | The workspace was not modified. |

### Summary

The Git/PR unit has solid logic, tests and coverage, but two adapter bugs against the real Kandev host (nested plugin metadata, `Repository.Name` versus `ProviderName`) break crash-recovery de-duplication and every repo-path-dependent Git feature. The watcher also ignores the per-workspace Backlog switch (U1 BR7.3). The fakes mask all three. Fix R-01 to R-03 and make the fakes Kandev-shaped before this can be READY. Known open risks (not blocking): the Git HTTPS probe and PR statuses 2/3 are unverified on a real space, the base-branch heuristic, and a Kandev clone may lack the task/session id.
