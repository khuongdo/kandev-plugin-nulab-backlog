## Review

**Verdict:** READY
**Reviewer:** aidlc-architecture-reviewer-agent
**Date:** 2026-10-06T07:12:24Z
**Iteration:** 2

### Findings

| ID | Severity | Location | Finding | Required action | Status |
|---|---|---|---|---|---|
| R-01 | Major | entities.md > ConnectionView; functional-spec.md > WF2 step 6 and Decisions That Change Upstream Wording | ConnectionView now has one definition. entities.md lists the same U1 fields as WF2 step 6, and the upstream table has a C5 row adding `enabled`. | None. | Resolved |
| R-02 | Major | functional-spec.md > WF6; rules.md > BR7.5 | WF6 and BR7.5 now follow the host's drafted switch. A flip marks the page dirty, Save calls persist, and persist calls `connection.set_enabled`. A failed persist keeps the change unsaved. `setIntegrationEnabled` is published only after success. This matches `DraftedIntegrationEnabledControl` in the Kandev v0.96.0 checkout (`persist` prop, dirty flag). | None. | Resolved |
| R-03 | Minor | functional-spec.md > Decisions That Change Upstream Wording | A row now records the unit-of-work (U1, U3), requirements and stories changes and their origin (Q6-Q9). | None. | Resolved |
| R-04 | Minor | functional-spec.md > WF6; rules.md > BR7.2 | The workspace now comes from the verified action context and the body is `{enabled}`. This matches plugins-authoring.md, where the plugin receives host-verified `workspaceID`. | None. | Resolved |
| R-05 | Minor | functional-spec.md > WF6 step 1, WF6 step 5 and P1 | The integration id `nulab-backlog` is named for the switch, `setIntegrationEnabled` and the settings link. It equals the manifest `id`. The host prefixes the control id itself (`createPluginUIApi`), so passing the bare id is consistent. | None. | Resolved |
| R-06 | Minor | functional-spec.md > WF3 step 8; rules.md > BR7.3 | Connect re-reads the switch just before storing and returns `integration_disabled` with no write if it is now off. | None. | Resolved |
| R-07 | Minor | functional-spec.md > WF3, WF6; rules.md > BR2.1, BR7.2 | The wording is now "Kandev admin". It matches the host's instance-wide `admin` access, which returns 403 for members. | None. | Resolved |
| R-08 | Minor | functional-spec.md > WF3 steps 6-8 | Kandev gives each action 15 seconds (plugins-authoring.md). Myself can take 10 s, and the failure-path rollback adds up to 5 s plus store operations, so the worst case can pass 15 s. The host then abandons the request while the plugin may still be writing or rolling back. The design does not say that the rollback context is independent of the host's cancellation. BR2.8 step 4 already uses a fresh context, so the risk is small. | State the 15 s host budget in WF3, or lower the Myself limit or the rollback limit so the sum fits. | New |

### Validation Tool Results

| Tool | Result | Interpretation |
|---|---|---|
| Manual cross-check against the Kandev v0.96.0 checkout (`host-api.ts`, `drafted-integration-enabled-control.tsx`, plugins-authoring.md) | Consistent | The drafted switch, the id prefixing, admin access and the verified workspace context all match the design. |
| required-sections, upstream-coverage, linter, type-check, traceability | Not run by this reviewer | The brief names them as stage tools, but I only inspected the artifacts and the SDK directly. |

### Summary

All seven iteration-1 findings are resolved, and the fixes introduced no new Critical or Major defects. One Minor point remains: the 15 s host action budget against the 10 s call plus the 5 s rollback.
