## Review

**Verdict:** READY
**Reviewer:** aidlc-architecture-reviewer-agent
**Date:** 2026-10-06T05:01:00Z
**Iteration:** 2

### Findings

| ID | Severity | Location | Finding | Required action | Status |
|---|---|---|---|---|---|
| R-01 | Major | aidlc/spaces/default/intents/261006-nulab-backlog-plugin/construction/walking-skeleton/functional-design/functional-spec.md > WF3 step 6 table; rules.md > BR2.3, BR2.4, BR2.9, BR2.10 | The result table now covers 200-valid, 200-unparsable, 401, 403, 404, 429, 400/422, 409, other 4xx, 3xx, 5xx, timeout, oversize and cancellation. U1's 429 behaviour (return at once, no retry) is stated. Residual: BR2.10 gives no rule for X-RateLimit-Reset being absent or in the past, and C5 schema requires retryAfterSeconds minimum 1. | Resolved. Optionally clamp retryAfterSeconds to at least 1 in BR2.10. | Resolved |
| R-02 | Major | functional-spec.md > WF3 step 7; rules.md > BR2.8, BR2.11; entities.md > ApiKeySecret | The order is now read-old, write-new-secret, write-record, rollback on a fresh 5-second context, with a rollback-failure event and a reconciliation rule in connection.get. This matches the SDK signature GetSecret returning (value, found, err). Residual gap: failure of step 7.1 (GetSecret error) and step 7.2 (SetSecret error) has no explicit outcome. The Error Outcomes row "Store write failed" implies internal with nothing changed. | Resolved for the blocking concern. Add one sentence that a failure at 7.1 or 7.2 returns internal with nothing written. | Resolved |
| R-03 | Major | entities.md > Contract Types Owned by U1 (C1); functional-spec.md > Decisions That Change Upstream Wording; contract-summary C2 | Credentials, User, Error with the status-to-kind map, and CallClass are now designed. HostPort is deferred to U3 with a reason that is consistent with the unit-of-work U1 boundary, which does not list it. Residual: contract-summary C2 still names walking-skeleton as owner and the decisions table lists no upstream change for it ("—"). | Resolved. Add C2 ownership move (walking-skeleton to U3) to the tracked upstream changes. | Resolved |
| R-04 | Major | rules.md > BR5.1, BR5.2; entities.md > PluginPackage, PlatformExecutable; functional-spec.md > WF4 | The five targets are named, PlatformExecutable is defined, and the keys match the Kandev manifest doc (key is goos-goarch, Windows value ends in .exe). BR5.2 now checks contents, manifest and the executables list. Residual: the in-package per-file checksum file is not named. plugin-pack generates checksums.txt inside the archive and rejects a pre-supplied one, while dist/checksums.txt is a different file with the same name. | Resolved. Clarify that the per-file checksums are plugin-pack's in-archive checksums.txt and dist/checksums.txt is the archive hash. | Resolved |
| R-05 | Major | traceability.json > coverage AC1.1.7; functional-spec.md > Decisions That Change Upstream Wording; stories.md AC1.1.7 | The spec now records the AC1.1.7, NFR3, components.md and C7 amendments as tracked upstream changes, and records the FR1.1 confirmation deviation as closed by U2 (US1.6). Residual: traceability.json still marks AC1.1.7 as plain OK with no amended note, although the AC text says header and BR3.4 says query parameter. | Partially resolved. Add an amended marker or note to the AC1.1.7 coverage entry that points to the decisions table. | Unresolved |
| R-06 | Minor | functional-spec.md > Scope, WF4; rules.md > BR5.5, BR5.6 | SDK pin mechanism with failure message is in BR5.5. All C4 targets are assigned to U1 and the coverage behaviour is in BR5.6. C4 is owned by walking-skeleton in contract-summary, which supports delivering all targets although the unit-of-work boundary lists only three. | Resolved | Resolved |
| R-07 | Minor | rules.md > BR1.1, BR1.2; functional-spec.md > Address examples | Scheme detection, trimming, ASCII-only, one label of 1-63 characters and checking on the parsed host are defined. The case table covers all nine AC1.2.2 inputs plus extras. | Resolved | Resolved |
| R-08 | Minor | rules.md > BR3.4, BR3.5, BR4.2 | Trimming, 1-256 printable ASCII, query encoding and a 1 MiB limit are stated as numbers. | Resolved | Resolved |
| R-09 | Minor | functional-spec.md > M1 Settings Screen States, WF2 | A state table now covers loading, load failed, member and admin variants, error, connecting, field error, unreachable, rate limited and busy. The ConnectionView fields include connected and align with C5 for the U1 subset. | Resolved | Resolved |
| R-10 | Minor | functional-spec.md > Entity Relationships; entities.md; rules.md > BR5.3 | The ER diagram text fallback now aligns with entities.md, hasApiKey is stated as derived, the state key is named, and the manual-check template path is given. | Resolved | Resolved |
| R-11 | Major | aidlc/spaces/default/intents/261006-nulab-backlog-plugin/construction/walking-skeleton/functional-design/ > missing frontend-components.md; stage definition .claude/aidlc-common/stages/construction/functional-design.md Step 4 | The stage definition makes frontend-components.md (component hierarchy, props and state, interaction flows, form validation, API integration points) conditional on the unit having UI. U1 includes the PluginUI M1 screen, yet the folder has no such file. The M1 state table and BR6.1 to BR6.4 cover states and accessibility but not component breakdown, props or which action each control calls. | Add frontend-components.md for M1, or record in functional-spec.md that the component design is deferred to code generation, with the reason. The state table can stay the source for states. | New |

### Validation Tool Results

| Tool | Result | Interpretation |
|---|---|---|
| Manual cross-check of traceability.json against rules.md IDs | PASS | Every OK target (BR1.1 to BR1.4, BR2.2 to BR2.4, BR2.6, BR2.7, BR3.1 to BR3.4, BR3.2, BR5.1 to BR5.4, BR6.1, BR6.3) exists in rules.md. |
| Kandev docs spot check (plugins-manifest.md, plugins-authoring.md) | PASS | Executable keys, the .exe rule, the secret APIs and the .tar.gz format match the design. |

### Summary

The revision closes the blocking gaps from iteration 1: exhaustive Backlog result mapping, an implementable secret and record write order with rollback and reconciliation, the C1 types, and a buildable and verifiable package definition. Remaining items are one Major (R-05 traceability note, partially resolved) and one new Major (R-11 missing frontend-components.md). That is two Majors and no Critical, so the verdict is READY, with both items worth fixing before approval.
