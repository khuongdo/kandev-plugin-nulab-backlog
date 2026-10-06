## Review

**Verdict:** READY
**Reviewer:** aidlc-architecture-reviewer-agent
**Date:** 2026-10-06T13:08:15Z
**Iteration:** 2

### Findings

| ID | Severity | Location | Finding | Required action | Status |
|---|---|---|---|---|---|
| R-01 | Critical | internal/plugin/host_port.go > FindTaskByMetadata, metadataNamespace | Verified against Kandev v0.96.0. `pluginTaskMetadata` (host_write.go) stores plugin data at `metadata["plugin:<id>"]`, with `plugin:` + `h.pluginID` equal to the manifest id `nulab-backlog`. The List mapper uses `PublicTaskMetadata`, which strips only host keys, so the nested namespace is returned. The code now reads only `t.Metadata["plugin:nulab-backlog"][key]`. The test fake `kandevMetadata` nests the same way. `TestU4_HostPort_ReadsNestedMetadata` and `TestU4_Watcher_CrashAfterCreateNoDuplicate` would fail on the old top-level read, so they guard the regression. Flat metadata and another plugin's namespace are correctly ignored. | None. | Resolved |
| R-02 | Critical | internal/plugin/host_port.go > repoName, Repository | Verified. `repositoryModelToDTO` sets Name to the provider-qualified name and `ProviderName` to the bare repository name. The code now uses ProviderName and falls back to the last path segment of the clone URL, stripping `.git`. For nulab-backlog repositories it then checks the result against `git.ValidRepoName` and fails closed with ErrRepositoryNotFound. Other providers are not validated (`TestU4_HostPort_OtherProviderKeepsWorking`). The fake `backlogRepo` has the real shape (Name `PROJ/web-app`, ProviderName `web-app`). The Backlog calls are asserted to use `PROJ/web-app` (`pathLog`), and the short-ref test covers `42`, `#42` and `web-app#42`. The old Name-based code would have failed these. No secret path is added: the RemoteURL is credential-free from Kandev and only its last segment is used. | None. | Resolved |
| R-03 | Major | internal/git/watcher.go > cycleWatch, createOne; internal/git/service.go > Connection.RequireEnabled | Verified. `cycleWatch` calls RequireEnabled first. Switch off returns 0 with no error and makes no Backlog call. A switch-read error counts as an error with no call (fail closed). `createOne` re-checks before the ledger reservation, so a mid-cycle switch-off leaves no ledger entry and no task. `ErrIntegrationDisabled` is dropped from `failed` so a disabled workspace is not counted as an error. `Watcher.Run` relies on the plugin guard (runtime.go:213) plus the `cycleWatch` check, and both layers are covered. The tests `SwitchOffSkipsWorkspace`, `SwitchReadFailsClosed`, `SwitchOffMidCycleStops` and `Runtime_WatcherObeysTheSwitch` assert no Backlog call, no task and no error, and that watching resumes when the switch is turned back on. All fail on the old code. | None. | Resolved |
| R-04 | Minor | internal/git/events.go > reconcile (59-68) / apply with restore=true | Deferred to Build and Test by the user. Not re-opened in this pass. | Re-check the epoch around the restore in `reconcile`, or document the one-cycle window as accepted. | Unresolved |
| R-05 | Minor | internal/plugin/runtime.go > Close (162-164) vs git_actions.go:173 | Deferred to Build and Test by the user. Not re-opened in this pass. | Read the watcher under `lifeMu`, or keep one restartable watcher. | Unresolved |
| R-06 | Minor | internal/git/service.go:464-468 (status mapping) | Deferred to Build and Test by the user. Not re-opened in this pass. | Map an unknown PR status id to state "unknown" with the label "Status unknown". | Unresolved |
| R-07 | Minor | internal/git/service.go:865-891 > RunQuery | Deferred to Build and Test by the user. Not re-opened in this pass. | Return the Not connected result when the query's project or host is not covered by the snapshot. | Unresolved |

### Validation Tool Results

| Tool | Result | Interpretation |
|---|---|---|
| gofmt -l internal server cmd | PASS (no output) | Formatting is clean. |
| go vet ./... | PASS | No findings. |
| go test -race -cover ./internal/... ./server/... | PASS | All packages are ok. The git package has 90.8% coverage and the plugin package 93.3%, above the 80% floor. The server package is wiring only. |
| go run ./cmd/ci secrets -root . | OK | No credentials in the tree. |
| git status --short before and after | Identical | The workspace was not modified. |

### Summary

All three blocking findings are repaired and checked against Kandev v0.96.0: nested `plugin:nulab-backlog` metadata, the repository name from ProviderName, and the per-workspace switch enforced in the watcher with fail-closed behaviour. The fakes now match the real host shapes, and the new tests fail on the old behaviour. I found no regression in crash/restart idempotency, other providers or secret handling. R-04 to R-07 remain as Minor items deferred to Build and Test.
