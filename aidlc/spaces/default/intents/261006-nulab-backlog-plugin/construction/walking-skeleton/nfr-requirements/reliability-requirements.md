# Reliability Requirements — walking-skeleton (U1)

Inputs: `requirements` (NFR5, NFR6, NFR7), `functional-spec` and `rules` for U1, `contract-summary` (timeouts, C4, C8), `team-practices`, and answers Q2 and Q4 in `nfr-requirements-questions.md`.

The plugin runs inside a self-hosted Kandev server, so its availability is the server's availability. U1 sets no SLO. It defines how the plugin behaves when Backlog or Kandev's stores fail.

| ID | Requirement | Pass/fail criterion | Source |
|----|-------------|---------------------|--------|
| NFR5.2 | At most 1 MiB is read from any Backlog response. A larger body counts as `unreachable` | A test with a 2 MiB fake body returns `unreachable` without reading it fully | BR4.2 |
| NFR5.3 | Every failure path ends in a defined error code (`validation`, `unreachable`, `rate_limited`, `conflict`, `integration_disabled`, `internal`). No error is swallowed, and no Go error is returned to Kandev from the action handler, except that a cancelled action context returns `context.Canceled` or `context.DeadlineExceeded` unchanged (team Code Style) | A table-driven test covers every row of the WF3 outcome table and every store failure, and asserts the code and HTTP status | NFR5, BR2.4, contract-summary (error codes) |
| NFR5.4 | A failed record write after a secret write, or a secret write that fails in a way that may have stored the value, is rolled back on an independent 2-second context [Q4]. A failed rollback leaves a state that `connection.get` reports as `error` and that is never used for a Backlog call | Tests inject a record-write failure, then a rollback failure, and check the secret store and the view | BR2.8, BR2.11 |
| NFR5.8 | A rejected key, an unreachable Backlog, a rate limit or a store failure never changes an existing working connection | Tests start from a connected state, inject each failure, and see an unchanged record and secret | BR2.3, BR2.4, BR2.5 |
| NFR5.9 | Turning Backlog off or on never changes the connection or the key; turning it back on needs no reconnect. A failed switch save leaves the stored and published values unchanged | Tests start connected, turn the switch off and on, and see the same record and secret; a test with a failing state write sees the old value from `connection.get` | BR7.4, BR7.5 |
| NFR6.1 | The manifest declares `min_kandev_version: "0.96.0"`. The package installs and runs on that version | The U1 manual check runs on Kandev 0.96.0. The U5 contract test checks out exactly that version | NFR6, Q2 |
| NFR7.1 | The package runs on all five supported platforms | The package has the five executables of BR5.1, and `verify-package` passes | NFR7, BR5.1, BR5.2 |

## Graceful Degradation

| Failure | Effect on the user | Recovery |
|---------|--------------------|----------|
| Backlog unreachable | Connect shows "Could not reach Backlog" with Retry. An existing connection is kept | Retry |
| Backlog rate limit | Connect shows the wait time | Retry after the wait |
| State or secret store unavailable | The settings screen shows "Could not load" with Retry. Connect returns `internal` | Retry. Nothing partial is left (NFR5.4) |
| Switch save fails | Kandev keeps the change unsaved and reports the failed save | Save again |
| Plugin process restart | None. No connection data is held in memory | Automatic |

## Sources

- [desc] Initial description: Kandev plugin for Nulab Backlog.
- [scope] Workflow-selected scope: `feature`.
- [Q2], [Q4]: answers in `nfr-requirements-questions.md`.
- `functional-spec.md`, `rules.md` (U1); `requirements.md`; `contract-summary.md`; `team-practices.md`.

## Assumptions & Open Questions

None.
