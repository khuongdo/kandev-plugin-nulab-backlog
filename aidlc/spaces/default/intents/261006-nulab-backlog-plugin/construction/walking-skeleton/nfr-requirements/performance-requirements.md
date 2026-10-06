# Performance Requirements — walking-skeleton (U1)

Inputs:

- `requirements` (NFR1, NFR2, NFR5).
- `functional-spec` and `rules` for U1.
- `contract-summary`: timeouts and the 15-second Kandev action limit.
- Answers Q1 and Q4 in `nfr-requirements-questions.md`.

NFR1 (3 seconds at p95) applies to lists, which arrive in U3. U1 gets its own targets [Q1].

| ID | Metric | Target | Load condition | Measurement |
|----|--------|--------|----------------|-------------|
| NFR1.1 | `connection.get` response time, measured at the plugin | p95 ≤ 500 ms | One workspace, normal state-store and secret-store latency; no Backlog call is made | Test timing over 100 calls against fake stores, plus the first manual check on a real server |
| NFR1.2 | `connection.connect_api_key` response time, measured at the plugin | p95 ≤ 3 s when Backlog answers normally; hard ceiling 14 s in every case, inside Kandev's 15-second action limit [Q4] | One Connect at a time per workspace | Fake server with a fixed delay in tests; the first manual check against a real space |
| NFR2.1 | Waiting on a Backlog rate limit during Connect | 0 s: Connect returns `rate_limited` at once with `retryAfterSeconds` ≥ 1 | Fake server answering 429 | Test asserts no retry and a response time under 1 s |
| NFR1.3 | `connection.set_enabled` response time, measured at the plugin | p95 ≤ 500 ms | No Backlog call; one state write | Test timing over 100 calls against fake stores |
| NFR1.4 | Connect time budget split [Q4] | Connect sets one 12 s deadline at entry for everything except the rollback: the first switch read, the lock and validation take ≤ 1 s; the Myself call takes ≤ 10 s and never runs past the deadline minus 2 s; the store writes, including the second switch read, take ≤ 2 s. The rollback has its own 2 s on a fresh context. Total ≤ 14 s | Fake stores and a fake server, each made to stall | Tests with an injected clock assert each limit and the total |
| NFR5.1 | Time limit on one Backlog call | ≤ 10 s, or earlier when the caller cancels | A fake server that never answers | Test with an injected clock or a short fake timeout asserts `unreachable` |

The 14-second ceiling in NFR1.2 comes from Q4: a 12 s deadline that covers the pre-call steps, the Myself call (10 s at most) and the store writes (2 s), plus 2 s for a rollback, leaving 1 s of margin under Kandev's 15-second action limit. U1 does not wait on rate limits at all (NFR2.1). This replaces the earlier 13-second figure and the 5-second rollback in the functional design (BR2.8, WF3 step 8.4).

## Resource Constraints

- **Response size**: at most 1 MiB is read per response (NFR5.2 in `reliability-requirements.md`, BR4.2).
- **Memory**: the plugin keeps no connection data in memory between actions. The only in-memory state is the ConnectAttempt lock, one per workspace.

## Sources

- [desc] Initial description: Kandev plugin for Nulab Backlog.
- [scope] Workflow-selected scope: `feature`.
- [Q1], [Q4]: answers in `nfr-requirements-questions.md`.
- `functional-spec.md`, `rules.md` (U1); `requirements.md`; `contract-summary.md`.

## Assumptions & Open Questions

- [assumption] Kandev's state store and secret store each answer in well under 200 ms on a self-hosted server. The 500 ms target for NFR1.1 depends on this.
