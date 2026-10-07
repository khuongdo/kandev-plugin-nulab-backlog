## Review

**Verdict:** READY
**Reviewer:** aidlc-architecture-reviewer-agent
**Date:** 2026-10-06T22:56:41Z
**Iteration:** 1

### Findings

| ID | Severity | Location | Finding | Required action | Status |
|---|---|---|---|---|---|
| R-01 | Minor | internal/plugin/actions_timing_test.go; aidlc/spaces/default/intents/261006-nulab-backlog-plugin/construction/walking-skeleton/nfr-design/performance-design.md > NFR1.1 | connection.get does 4 store reads (record, secret, switch, Git secret), but performance-design says 3. This is a doc mismatch with a thin real margin. At 100 ms per op, p95 is 400 ms against the 500 ms target. NFR1.1 only holds while each store op is under about 125 ms (3 reads allowed about 166 ms). The requirement's own assumption is "well under 200 ms", so a slow store could breach it. The test is genuine, not trivial: it reads durations from the synctest clock, asserts p95 > 0 and asserts zero Backlog calls. | Update performance-design NFR1.1 to say 4 reads and state the 125 ms per-op ceiling. Check store latency in the first manual check on the real server. Optionally skip the Git secret read when the connection record shows no Git credential. | New |
| R-02 | Minor | internal/connection/lifecycle.go > test (line 177) | The Myself call in the post-connect recheck keeps the retry policy. This is correct: only Connect and OAuth sign-in were in scope for NFR2.1. The lack of retry-disabling there is intentional but not written down. | Add a one-line comment in test() saying the recheck keeps retries on purpose. | New |
| R-03 | Minor | internal/plugin (host allowlist check) | Earlier U1 Minor: the host allowlist is checked only at entry points. | Carry forward; no change in this loop-back. | Unresolved |
| R-04 | Minor | internal/plugin (runtime start) | Earlier U1 Minor: background workers start before Host injection. | Carry forward; no change in this loop-back. | Unresolved |

### Validation Tool Results

| Tool | Result | Interpretation |
|---|---|---|
| go test -race -cover ./internal/... ./server/... | PASS; coverage 88.0-97.4% per package (server/ main wiring 0%, allowed) | Floor of 80% holds and nothing regressed. |
| go test -race -count=5 ./internal/... | PASS, no failures | No flake or race in 5 repeats. The developer's count=40 run was not repeated. |
| go vet ./...; gofmt -l | clean | No issues. |
| go run ./cmd/ci secrets -root . | OK | The new alphabet does not trip the secret scanner. |
| git status --short before/after | identical | Workspace was not modified. |

### Verification notes

- Step 23: NoRetry is an unexported-key context value. It is applied only at 3 call sites: service.go verify (API key Connect), oauth.go signIn Myself, plus the definition. send() checks it after the attempt, so the 429 is returned after 1 request. Other callers (lifecycle recheck, issues, git) keep retries. The marker rides only on the derived context of that one call, so it cannot leak to unrelated calls. Cancellation is untouched: the early return skips pause(), and context errors flow as before. TestNoRetryReturns429AtOnce covers it.
- Step 24: Windows() trims both fixed prefixes and slides over the 32-character random part, so every 4-character window is checked (29 windows). The fixed prefix is never scanned. TestAssertNoLeak_Catches4CharWindow checks each window position. The 16-letter alphabet uses 4 random bits per character (the byte mask 0x0f gives no modulo bias), which gives 128 bits. The keys are printable ASCII with a length well inside 1-256, so validation tests are unaffected. The call-site edits drop only the `, 8` argument. The diff shows nothing else changed there.
- Step 25: latency comes from slowHost sleeping inside synctest. Latencies are 100 ms per op, half the "well under 200 ms" assumption in performance-requirements. 100 samples are taken and d[94] is used as p95.

### Summary

Steps 23-25 are correct, minimal and verified: tests, race, vet, gofmt, secrets scan and the coverage floor are all green. The only item worth acting on is the 4-read vs 3-read mismatch. It is a documentation fix with a thin latency margin (about 125 ms per store op), not a defect, so the severity is Minor and the verdict is READY.
