## Review

**Verdict:** READY
**Reviewer:** aidlc-architecture-reviewer-agent
**Date:** 2026-10-06T02:52:43Z
**Iteration:** 1

### Findings

| ID | Severity | Location | Finding | Required action | Status |
|---|---|---|---|---|---|
| R-01 | Major | aidlc/spaces/default/intents/261006-nulab-backlog-plugin/inception/domain-design/components.md > component Connection (external_dependencies, behaviour) vs ADR-007 | Connection stores secrets through Kandev's secret store (SetSecret) and must durably store SpaceConnection, ProjectSelection, connectionEpoch, OAuthPendingState. But only KandevAdapter may import pluginsdk (ADR-007, Code Style). Connection declares no host port and no own data store (only IssueIntegration and GitIntegration have one). A developer cannot tell by which path Connection touches secrets/durable storage without breaking the convention. | Declare in Connection a host port (secret store, data store) implemented by KandevAdapter, and add the own data store to Connection's external_dependencies; update ADR-007 to match. | New |
| R-02 | Major | aidlc/spaces/default/intents/261006-nulab-backlog-plugin/inception/domain-design/decisions.md > ADR-005; components.md > Connection behaviour | ConnectionChanged is an asynchronous event, but it does not say whether delivery is durable. If the plugin restarts between incrementing connectionEpoch and the components receiving the event, links and watches get stuck in the old state. The restore flag also needs to know the previous space/projects, but no Connection entity stores that history (SpaceConnection is dropped on disconnect). The event bus mechanism (synchronous in-process, ordering, subscribers) is not stated either. | State clearly: the event is only a hint and each component must compare its durably stored connectionEpoch with the current epoch at startup and before each cycle; state where the previous space/projects are stored to compute the restore flag; state the in-process event bus. | New |
| R-03 | Minor | aidlc/spaces/default/intents/261006-nulab-backlog-plugin/inception/domain-design/components.md > IssueIntegration/GitIntegration external_dependencies "Kandev server (via the host port provided by KandevAdapter)" | At runtime there is a reverse edge IssueIntegration/GitIntegration -> KandevAdapter, while KandevAdapter -> these two components. Only the compile-time cycle is avoided thanks to dependency inversion; the yaml does not declare the host port as a component or interface, so the name KandevAdapter appears in a string rather than as a declared entry. | Record the host port as a named entry (for example HostPort, with the create task/set label/check task/supply Git credentials operations) and state the initialisation order; or add a note in the Component Diagram. | New |
| R-04 | Minor | aidlc/spaces/default/intents/261006-nulab-backlog-plugin/inception/domain-design/components.md > BacklogGateway behaviour | The sequential update/search group queue does not distinguish background calls (sync, PR watch) from interactive user calls (issue search US2.2). Waiting on 429 for up to 60 seconds several times can push interactive search past the 3-second target (US8.1). The value of "N attempts" is not defined. RateLimitWindow is keyed by category, not by space/user, so old remaining quota can carry over to a new space. | Give interactive calls priority or a separate wait cap; record the value of N; key RateLimitWindow by (spaceHost, category) and clear it on epoch change. | New |
| R-05 | Minor | aidlc/spaces/default/intents/261006-nulab-backlog-plugin/inception/domain-design/components.md > KandevAdapter behaviour (mount point list); Connection responsibilities US1.3 | OAuth needs a plugin webhook callback (developer assumption A2) but KandevAdapter does not list webhook registration, and there is also no clear owner for the task of deleting the Git secret on disconnect/space change. | Add the OAuth webhook callback to KandevAdapter's mount point list, routed to Connection; state clearly that Connection deletes the Git secret too (F2) before emitting ConnectionChanged. | New |

### Validation Tool Results

| Tool | Result | Interpretation |
|---|---|---|
| Manual check of the YAML catalogue | Unique names; every referenced component is declared; no self-dependency; depends_on and dependents are symmetric; each entity has one owner and an identifier; all references match the declared owner; the declared graph has no cycles | Passes the well-formedness rules; the issues are in undeclared dependencies (R-01, R-03). |
| traceability.json | US1.1–US8.5 are all covered (US7.4/7.5 Deferred, US7.6 N/A with a reason) | Acceptable. |

### Summary

The component catalogue is structurally correct and has no declared cycles; the ADRs have Context, Decision, Consequences, Alternatives Rejected and a security item. Two Major points the approver should consider: Connection uses secrets/durable storage without a host port (conflicting with ADR-007), and the ConnectionChanged mechanism does not yet guard against lost events/restore after a restart.
