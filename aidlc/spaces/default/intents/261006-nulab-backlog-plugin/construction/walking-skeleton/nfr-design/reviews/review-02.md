## Review

**Verdict:** READY
**Reviewer:** aidlc-architecture-reviewer-agent
**Date:** 2026-10-06T07:27:47Z
**Iteration:** 1

### Findings

| ID | Severity | Location | Finding | Required action | Status |
|---|---|---|---|---|---|
| R-01 | Major | aidlc/spaces/default/intents/261006-nulab-backlog-plugin/construction/walking-skeleton/nfr-design/security-design.md > NFR3.9 (and reliability-design.md > Failure Handling Summary) | The disabled-switch guard reads the IntegrationSwitch "first", but no design says what happens when that read fails (state store error, timeout, corrupt record). Fail-open would let Connect run (secret read, Backlog call) while Backlog may be off, breaking BR7.3; fail-closed needs an error code and the NFR5.3 mapping table has no row for it. The Connect second read inside the 2 s store budget has the same gap. A developer must guess a security-relevant default. | State that a failed or undecodable switch read is fail-closed: the action returns `internal`, with no secret read, no Backlog call and no write. Add the row to the NFR5.3 mapping table and a test case to NFR3.9. A missing record still means on (BR7.1). | New |
| R-02 | Minor | aidlc/spaces/default/intents/261006-nulab-backlog-plugin/construction/walking-skeleton/nfr-design/performance-design.md > NFR1.1 | The design says `connection.get` does "two store reads (record, secret)". Since C5 `ConnectionView` gained `enabled` (functional-spec WF2 step 4 reads the IntegrationSwitch), `get` now does three reads. A test that counts reads, or the p95 argument, would be off by one. | Update NFR1.1 to three reads (record, secret, switch) and re-state the 500 ms assumption for three round trips. | New |
| R-03 | Minor | aidlc/spaces/default/intents/261006-nulab-backlog-plugin/construction/walking-skeleton/nfr-design/observability-design.md > Events (NFR11.2, NFR11.6) | NFR11.2 requires exactly one outcome event per Connect. A Connect refused by the guard logs `action_refused_disabled`; a Connect refused at the second switch read (WF3 step 8) is not assigned any event. It is unclear whether `connect_failed` (errorCode `integration_disabled`), `action_refused_disabled`, or both are logged. | State which event each of the two refusal points emits, so that one Connect yields exactly one outcome event. | New |
| R-04 | Minor | aidlc/spaces/default/intents/261006-nulab-backlog-plugin/construction/walking-skeleton/nfr-design/security-design.md > NFR3.9 | The second switch read narrows the race but is not atomic with SetSecret: `set_enabled(false)` can still land between the read and the store. The design does not say this window is accepted, and the stores offer no compare-and-set. | Record the residual window as an accepted limit, bounded by the 2 s store budget, and make clear the test covers only the switch turned off during the Myself call. | New |
| R-05 | Minor | aidlc/spaces/default/intents/261006-nulab-backlog-plugin/construction/walking-skeleton/nfr-design/security-design.md > NFR3.10 | The logo design covers the inline SVG and bundle check but not the compliance requirement in security-requirements (Q6): `docs/brand/backlog-logo.md` with source and terms, the pre-release statement of Nulab brand approval, and the fallback to a host built-in icon. No component or test owns it. | Add the brand record file and the fallback rule to NFR3.10 or the logical-components table (Manual checks row). | New |

### Validation Tool Results

| Tool | Result | Interpretation |
|---|---|---|
| required-sections | Manual check: all five design files have `## Sources` and `## Assumptions & Open Questions` | No gap found |
| traceability | Manual check: `traceability.json` lists 38 upstream IDs and 38 coverage rows, sets equal, all OK, including NFR1.3, NFR1.4, NFR3.8-3.10, NFR5.9, NFR5.10, NFR11.6 | Every new NFR ID has a design row. Judgement is in R-01 to R-05 |
| upstream-coverage, linter, type-check | Not run as separate commands; no engine runner available in the reviewer shell | Findings come from reading |

Kandev v0.96.0 checks: the 15 s action timeout (`pluginActionTimeout`) matches the 14 s ceiling with 1 s margin. `writeActionResponse` accepts any status in 200-599, so 409 for `integration_disabled` is implementable. `access: admin` and `scope: workspace` are supported manifest fields. `IntegrationEnabledControl` with a `persist` callback and `setIntegrationEnabled` exist in the host API. The Connect budget arithmetic is consistent: 1 s pre-call, Myself ending by deadline minus 2 s, 2 s store, then 2 s rollback gives 14 s.

### Summary

The design covers every NFR ID, including the new ones, and is implementable on Kandev 0.96.0. The one real gap is the unspecified fail-closed behaviour of the switch guard when the state read fails (R-01). The rest are small consistency fixes.
