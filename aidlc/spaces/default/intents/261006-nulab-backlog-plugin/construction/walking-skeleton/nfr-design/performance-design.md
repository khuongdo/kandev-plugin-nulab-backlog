# Performance Design — walking-skeleton (U1)

Inputs:

- `performance-requirements`: NFR1.1–NFR1.4, NFR2.1, NFR5.1.
- `reliability-requirements`: NFR5.2.
- `functional-spec` (U1).
- `contract-summary`: the 15-second Kandev action limit.
- `nfr-design-questions.md`: Q1, superseded by NFR requirements Q4 (see its change section).

## Connect Time Budget (NFR1.4)

The Connect handler sets one 12-second deadline when it starts. Every step except the rollback runs under it, each with its own smaller limit, also bounded by the time left:

| Step | Limit | On timeout |
|------|-------|------------|
| Read the switch, validate input, acquire lock | 1 s together | `internal`, nothing written |
| Backlog `Myself` call | 10 s, and never later than the deadline minus 2 s | `unreachable` |
| Read the switch again, GetSecret (previous), SetSecret (new), write record | 2 s together | `internal`; rollback when SetSecret may have stored the value or the record write failed |
| Rollback (restore or delete secret) | 2 s, on a fresh context not derived from the action context | Logged `connection_inconsistent`, `internal` |

The worst case is 12 + 2 = 14 s, 1 s under Kandev's 15-second action limit. This replaces the earlier 13-second budget with its 8-second call limit (Q1, superseded).

| ID | Design |
|----|--------|
| NFR1.1 | `connection.get` does three store reads (record, secret, switch) and no network call. There is no cache: each read is one store round trip, and nothing else runs |
| NFR1.2 | The budget above. A test with a fake Backlog that delays 2 s confirms p95 under 3 s. A test with a Backlog that never answers confirms a response at about 10 s |
| NFR1.3 | `connection.set_enabled` does one state read and one state write, with no network call and no secret access |
| NFR1.4 | The budget table above, driven by an injected clock in tests: each stalled step hits its own limit, and the total never exceeds 14 s |
| NFR2.1 | On 429, the gateway reads `X-RateLimit-Reset` (epoch seconds), then `Retry-After`, else 60. It computes the wait as at least 1 s and returns it at once, with no sleep |
| NFR5.1 | Each gateway call takes a context. The client default is 10 s; Connect passes its own sub-deadline, which is never longer |
| NFR5.2 | The response body is read through a 1 MiB limited reader. Reading past the limit gives `Unreachable` |

## Resource Use

- One shared `http.Client` per process, with connection reuse. U1 makes so few calls that no pool tuning is needed.
- No caching. U1 has nothing worth caching.

## Sources

- [desc] Initial description: Kandev plugin for Nulab Backlog.
- [scope] Workflow-selected scope: `feature`.
- [Q1]: answer in `nfr-design-questions.md`, superseded by NFR requirements Q4.
- `performance-requirements.md`, `reliability-requirements.md` (U1 NFR requirements); `functional-spec.md` (U1); `contract-summary.md`.

## Assumptions & Open Questions

None.
