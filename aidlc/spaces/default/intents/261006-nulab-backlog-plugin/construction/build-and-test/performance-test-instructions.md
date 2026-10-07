# Performance Test Instructions

The test strategy is Standard, so no load-test suite is generated. Performance NFRs exist (NFR1, NFR1.1–NFR1.4, NFR2, NFR5.1), so this file records how each is checked. All automated checks run on virtual time (`testing/synctest`) against fakes with injected latency; real-network timing is a manual check on a real Backlog space and a deployed Kandev.

## Automated (local, part of `make test`)

| Target | Check | Command |
|--------|-------|---------|
| AC8.1.1: 20-row issue list ready ≤ 2.5 s with 300 ms per Backlog call | `TestU3_List_MeetsTheLatencyBudget` | `go test -race ./internal/issues/ -run TestU3_List_MeetsTheLatencyBudget` |
| NFR1.2: Connect p95 < 3 s with a 2 s Backlog delay | `TestConnectBudget` | `go test -race ./internal/connection/ -run TestConnectBudget` |
| NFR1.4: one 12 s Connect deadline, total ≤ 14 s | `TestNewServiceUsesTheDesignBudget`, `TestGuardRunsInsideTheConnectDeadline` | `go test -race ./internal/connection/ ./internal/plugin/ -run 'TestNewServiceUsesTheDesignBudget\|TestGuardRunsInsideTheConnectDeadline'` |
| NFR5.1 / AC8.1.3: every Backlog call times out at ≤ 10 s; others still served | `TestMyselfTimeoutIsUnreachable`, `TestU3_Issue_TimeoutDoesNotBlockOthers` | `go test -race ./internal/backlog/ -run 'TestMyselfTimeoutIsUnreachable\|TestU3_Issue_TimeoutDoesNotBlockOthers'` |
| NFR2 / AC8.4.1–AC8.4.3: 429 waits, max 3 retries, Search/Update queues 1 s apart | limiter tests | `go test -race ./internal/backlog/ -run '^Test(RateLimit\|Queue)\|TestU3_Issues_SearchCallsAreQueued\|TestU4_CreatePR_Queued'` |
| NFR1.1/1.3 structure: `set_enabled` does 1 state read, 1 write, 0 secret access | `TestSetEnabledDoesOneStateReadOneStateWriteAndNoSecretAccess` | `go test -race ./internal/plugin/ -run TestSetEnabledDoesOneStateRead` |

## Gaps

- NFR1.1/NFR1.3 ask for `connection.get` and `connection.set_enabled` p95 ≤ 500 ms over 100 calls. No automated timing test over 100 calls exists; only the call-count structure is tested.

## Manual (real Backlog space + deployed Kandev)

- AC8.1.2: open a real 20-row issue list 20 times; at least 19 opens within 3 s (`docs/manual-checks/TEMPLATE.md` step 25).
- NFR1.1/NFR1.3 timing: time 100 `connection.get` / `set_enabled` calls on the self-hosted server.

Owning later stage for real-environment performance: `performance-validation` (4.6).
