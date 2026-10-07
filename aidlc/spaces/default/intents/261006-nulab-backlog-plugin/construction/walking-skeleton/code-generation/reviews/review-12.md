## Review

**Verdict:** READY
**Reviewer:** aidlc-architecture-reviewer-agent
**Date:** 2026-10-07T00:50:00Z
**Iteration:** 1

### Findings

| ID | Severity | Location | Finding | Required action | Status |
|---|---|---|---|---|---|
| R-01 | Minor | aidlc/spaces/default/intents/261006-nulab-backlog-plugin/construction/walking-skeleton/code-generation/code-summary.md > store reads per Connect | Earlier U1 Minor: 4 store reads in the Connect path versus the 3 documented. Not touched by this repair. | Align the documented read count with the code, or cut the redundant read. | Unresolved |
| R-02 | Minor | internal/plugin (recheck comment, earlier U1 Minor) | Earlier U1 Minor: the recheck comment is still not corrected. Not touched by this repair. | Fix the comment to match the actual behaviour. | Unresolved |
| R-03 | Minor | internal/plugin and internal/backlog > host allowlist | Earlier U1 Minor: the host allowlist (https; backlog.com, backlog.jp, backlogtool.com) is enforced only at the entry points, not in the client. Unchanged by this repair. | Enforce it in the client constructor, or record it as an accepted risk. | Unresolved |
| R-04 | Minor | internal/plugin/runtime.go > SetHost, waitHost, workers | Earlier U1 Minor: workers could tick before Host injection. Now covered: store calls wait on `hostReady` inside their 1 s CallTimeout, and TestHost_WorkersWaitWithoutStorm shows no ERROR line and a clean first cycle when the Host arrives within 1 s. If the Host arrives later than 1 s, one cycle logs its error count once, which is bounded and not a storm. | None. | Resolved |
| R-05 | Minor | internal/plugin/runtime.go > waitHost callers: HandleWebhook (webhook.go), ResolveGitCredential and GetGitCredentialBinding (credential.go), OnEvent (events.go), SearchEntityReferences and AuthorizeEntityReference (references.go) | The up-front wait of at most 5 s exists only in HandleAction. Every other path first reaches the Host inside a store call that has a 1 s CallTimeout (issues/git/connection store.go), so it waits at most 1 s and then fails closed. The race is therefore narrowed to 1 s for those paths, not closed. Classification: webhook gives `oauth=failed` (user can retry); credential RPCs are refused (fail closed, Kandev retries); OnEvent returns an error (Kandev retries at 5, 15 and 45 s); `#` search returns an empty list; Authorize denies. All are safe and none leaks. The window only matters if Kandev routes traffic to the plugin more than 1 s before SetHost. Also, an unguarded action has no deadline, so it can wait up to 5 s on top of its own work. That is acceptable but not documented. | Document the 1 s versus 5 s split in code-summary.md. Optionally call waitHost once at the start of HandleWebhook and the RPCs. | New |
| R-06 | Minor | internal/plugin/runtime.go > waitHost (expired or canceled ctx returns errNoHost) | Deviates from the team rule to return `context.Canceled` and `DeadlineExceeded` unchanged. It applies only while no Host is set. Effect: a canceled or expired call is classified `internal` (ERROR) instead of a timeout or cancellation. In particular, `Close()` while the Host is missing is logged as an error. The deviation is bounded, contained and explained in a comment, so it is not blocking. | Return `errors.Join(errNoHost, ctx.Err())` (or wrap with `%w` for both) so `errors.Is` sees the context error, and add one test. Otherwise record it as an Accepted risk in the plan. | New |

### Validation Tool Results

| Tool | Result | Interpretation |
|---|---|---|
| go vet ./... | PASS, no output | No vet findings. |
| go test -race -count=3 ./internal/plugin/... | PASS (ok, 16.7 s) | No data race and no flake in 3 runs, including the concurrent SetHost test. |
| Manual check: SetHost and Host in the SDK (apps/backend/pkg/pluginsdk/plugin.go) | Both are guarded by an RWMutex | The `hostReady` close happens after the Host is stored, so a woken waiter always sees non-nil. `sync.Once` makes a double close impossible. |

**Verified claims.** There is no deadlock: `waitHost` selects on `ready`, `ctx.Done()` and the timer, takes no lock, and `defer t.Stop()` releases the timer. There is no goroutine leak: no goroutine is spawned. SetHost(nil) does not open the gate. The tests use `synctest.Test` with `synctest.Wait` and fake time properly, and the cap test asserts exactly `hostWait`. The Connect budget holds: the 12 s deadline starts before the Host wait and the guard, and the 5 s cap leaves at least 7 s for the Backlog call. In the late-Host case the effective Backlog time shrinks by the wait, and the 2 s rollback has its own timeout.

### Summary

The repair is sound: no deadlock, no leak, race-clean, and the early-action race is closed for HandleAction and workers. The webhook, RPC and OnEvent paths still fail fast after 1 s, but each fails closed or is retried (R-05). The `errNoHost` versus ctx.Err deviation is Minor (R-06). There are no Critical or Major findings, and R-01 to R-03 remain open Minors.
