# Reliability Design — walking-skeleton (U1)

Inputs:

- `reliability-requirements`: NFR5.2–NFR5.4, NFR5.8, NFR5.9, NFR6.1, NFR7.1.
- `functional-spec` and `rules` (U1).
- `tech-stack-decisions`.
- `contract-summary`.
- `nfr-design-questions.md` (Q1, superseded by NFR requirements Q4).

## Storage Consistency (BR2.8, BR2.11)

The secret write is the commit point. The record is written after it. An outcome is valid only when the epoch and host match.

| Situation | What is stored | What `connection.get` reports | Effect |
|-----------|----------------|-------------------------------|--------|
| Normal | Secret and record, same epoch and host | `connected` | — |
| Crash or failure after the secret write, first Connect, rollback succeeded | Nothing | `not_connected` | None |
| Same, rollback failed or the process died | Secret only, no record | `not_connected` | Harmless: the secret is never read without a record. The next Connect overwrites it, and U2's Disconnect deletes it |
| Crash or failure after the secret write, replace, rollback succeeded | Old secret and old record | `connected` (old) | None |
| Same, rollback failed or the process died | New secret, old record (epoch mismatch) | `error` ("connect again") | The mismatched secret is never used for a call (BR2.11). The next Connect repairs it |

There is no transaction across the two stores, so this is the strongest guarantee available. Every leftover state is either invisible and unused, or visible as `error` and unused.

| ID | Design |
|----|--------|
| NFR5.2 | 1 MiB limited reader. Over the limit, the call returns `Unreachable` (see `performance-design.md`) |
| NFR5.3 | The action handler maps every `backlog.Error.Kind`, store error and `connection.ErrDisabled` to an action error code through one table, and returns a `PluginActionResponse` with a status. The only Go errors returned are `context.Canceled` and `context.DeadlineExceeded` from a cancelled action context, unchanged. A table-driven test covers every row of the WF3 outcome table, plus store failures |
| NFR5.4 | The rollback runs with a fresh 2-second context. It runs when the record write fails, and also when SetSecret returns an error after which the value may have been stored (a timeout or a transport error), because the plugin cannot tell whether that write landed. Its failure is logged as `connection_inconsistent`. Reconciliation is by epoch comparison on every read (table above) |
| NFR5.8 | Nothing is written before a 200 with a valid `User`. The old record and secret are only replaced after that, in the order above |
| NFR5.3a | Switch read failures | A failed or undecodable switch read maps to `internal` in the same table; the guard fails closed (security-design, NFR3.9) |
| NFR5.9 | The switch lives in its own state key (`integration`). `set_enabled` writes only that key, so no switch change can touch the connection record or the secret. A failed write returns `internal`; the drafted control keeps the change unsaved and nothing is published |
| NFR6.1 | `manifest.yaml` declares `min_kandev_version: "0.96.0"`. A unit test reads the manifest and asserts the value. The first manual check installs on Kandev 0.96.0 |
| NFR7.1 | `make build` cross-compiles the five platform keys from BR5.1 with `CGO_ENABLED=0`. `-race` tests run separately with CGO on the CI host. `make verify-package` checks all five |

## Failure Handling Summary

| Failure | Detection | Response |
|---------|-----------|----------|
| Backlog slow | Myself sub-deadline (10 s, within the 12 s Connect deadline) | `unreachable`, nothing written |
| Backlog errors | Status mapping (functional-spec WF3) | Defined code per status |
| Store slow or failing | 2 s shared store budget | `internal`; rollback when the secret may have been written |
| Backlog turned off during a Connect | Second switch read before storing | `integration_disabled`, nothing written |
| Plugin crash | — | Restart is safe (`scalability-design.md`); leftover states as in the table above |

There are no retries in U1. Connect is a user action, and the user decides whether to press Retry.

## Sources

- [desc] Initial description: Kandev plugin for Nulab Backlog.
- [scope] Workflow-selected scope: `feature`.
- [Q1]: answer, superseded by NFR requirements Q4, in `nfr-design-questions.md`.
- `reliability-requirements.md`, `tech-stack-decisions.md` (U1 NFR requirements); `functional-spec.md`, `rules.md` (U1); `contract-summary.md`.

## Assumptions & Open Questions

None.
