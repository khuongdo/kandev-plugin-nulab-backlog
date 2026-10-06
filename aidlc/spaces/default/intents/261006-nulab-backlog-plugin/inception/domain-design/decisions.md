# Architecture Decisions (ADR) — Kandev Plugin for Nulab Backlog

Inputs:

- `requirements` (FR, NFR)
- `stories` (US…)
- `team-practices` (Code Style)
- `components.md`
- Answers Q1–Q5 in `domain-design-questions.md`

## ADR-001: Split the Plugin into Six Components

- **Context**: The plugin has an issue part (Must) and a Git/PR part (optional, per Q8 at the story step). There are two periodic jobs: issue status sync and PR watches. The team is a single person, working with AI. Boundaries must be clear enough to test independently, but not split too finely.
- **Decision**: Six components: Connection, BacklogGateway, IssueIntegration (including issue status sync), GitIntegration (including repositories, PRs, PR watches, queries), KandevAdapter, PluginUI. [Q1]
- **Consequences**:
  - (+) Fewer components means fewer boundaries to maintain. The optional Git/PR part is contained in one component, so dropping it or doing it later is easy.
  - (−) GitIntegration is fairly large, with four feature groups; it needs clear package splits inside.
  - (−) IssueIntegration both serves user actions and runs cycles, so data races must be avoided (a `-race` test requirement already exists).
  - Security: unchanged compared with the 8-component option, because secrets still live only in Connection and the secret-masking layer is still in BacklogGateway.
- **Alternatives Rejected**:
  - *8 components by feature* (separating Issue sync, Repository+PR, PR watch+query): finer boundaries, but three more boundaries for one person. Issue sync and issue linking always change together.
  - *A single business component*: the optional Git part cannot be separated, and independent testing is hard.

## ADR-002: BacklogGateway Is the Anti-Corruption Layer and Holds No Credentials

- **Context**: Every Backlog call must respect the per-user limit (C-T3), mask secrets (NFR3), have a time limit (NFR5), and not follow redirects to unknown hosts (US1.2). Connection needs to call Backlog to verify credentials, while Issue and Git need credentials from Connection.
- **Decision**: BacklogGateway is the only component that calls the Backlog API. It depends on no other component. Credentials are passed in by the caller on every call. The gateway converts HTTP errors into fixed error kinds and does not carry the response body.
- **Consequences**:
  - (+) The dependency graph has no cycles. All rate-limit and secret-masking rules live in one place, testable with a fake server.
  - (−) Callers must fetch credentials from Connection before every call.
  - Security: secrets are not cached in the Gateway, which reduces the places they can leak.
- **Alternatives Rejected**:
  - *Gateway reads credentials from Connection itself*: creates a Connection ↔ Gateway dependency cycle.
  - *Each component calls Backlog itself*: the API rate limit and secret masking are repeated in many places, easy to miss.

## ADR-003: Each Periodic Component Has Its Own Timer; Spacing Is Handled by the Gateway

- **Context**: IssueIntegration (status sync) and GitIntegration (PR watches) both run on cycles and share one account's API quota. Requirement US8.4 is that calls in the update group and the search group never run in parallel, and are at least 1 second apart.
- **Decision**: Each component has its own timer [Q2]. BacklogGateway has a queue per call group to guarantee ordering and spacing, no matter which component the call comes from.
- **Consequences**:
  - (+) Each component controls its own cycle. The Git part can be absent without affecting the issue part.
  - (−) The two timers can fire at the same time. Calls then have to queue at the Gateway, and a cycle may run longer than expected. The number of waits must be logged (US8.3).
  - Security: no impact.
- **Alternatives Rejected**:
  - *One shared timer component*: better scheduling, but adds a component and makes the issue part depend on the Git part's schedule.

## ADR-004: Links Are Stored in the Plugin's Own Data

- **Context**: We need to know which issue or PR is attached to which task, for: showing labels, status sync, marking "no longer connected" and restoring.
- **Decision**: IssueLink and PullRequestLink are stored in the plugin's own data store and are the primary source. The label on the task card is only a display copy, rebuilt from this data. [Q3]
- **Consequences**:
  - (+) Keeps information a label cannot hold, including the connection version marker, link state and space. This makes restore on reconnect possible.
  - (−) Must sync when a task is deleted in Kandev, otherwise there will be dangling links (assumption in `components.md`).
  - Security: link data contains no secrets. It only contains issue keys and task IDs.
- **Alternatives Rejected**:
  - *Store in labels or Kandev task metadata*: users can edit labels by hand, and there is not enough room for link state.

## ADR-005: ConnectionChanged Event with a Connection Version Marker

- **Context**: Disconnect, space change or project deselection must move links and watches to "no longer connected". Reconnecting to the same old space or reselecting the same old project must restore them automatically (Q4 at the Refined Mockups step). Late results from the old space must not overwrite new data (AC1.8.3).
- **Decision**: Connection increments `connectionEpoch` and emits a ConnectionChanged event, containing the space, projects, marker and restore flag. Each component updates the data it owns itself, and ignores every result carrying an old marker. [Q4]
- **Consequences**:
  - (+) Connection does not need to know other components' data, so there are no dependency cycles. A new component only needs to subscribe.
  - (−) Updates happen asynchronously: for a short time the UI may still show the old state.
  - (−) The version marker must be stored durably so restore works after a restart.
  - Security: on disconnect, secrets are deleted in Connection immediately, before the event is emitted. No component holds a copy of the secrets.
- **Alternatives Rejected**:
  - *Connection calls each component directly*: Connection would have to depend on Issue and Git, creating a dependency cycle, and would need changes every time a component is added.

## ADR-006: A Single UI Package

- **Context**: There are 12 screens, sharing the state component and the confirmation dialog.
- **Decision**: PluginUI is a single package, split by screen inside. [Q5]
- **Consequences**:
  - (+) One registration with Kandev, shared state and accessibility code. Simple packaging.
  - (−) The Git/PR part is in the same package, so it is loaded even when not used. Acceptable for a small plugin.
  - Security: the UI never receives secrets, only flags.
- **Alternatives Rejected**:
  - *Several packages by area*: more complex registration and packaging with no clear benefit.

## ADR-007: Only KandevAdapter Depends on the Kandev Toolkit; Business Components Use a Host Port

- **Context**: Per the Code Style convention, only `internal/plugin` may import the Kandev toolkit. Yet IssueIntegration and GitIntegration still need to create tasks, set labels and supply Git credentials through Kandev.
- **Decision**: KandevAdapter implements a host port (interface) and passes that port into the business components. In the component catalogue, the host port is recorded as an external dependency of the business components, so there is no dependency cycle with KandevAdapter.
- **Consequences**:
  - (+) Business components can be tested with a fake port, without Kandev. When Kandev changes its toolkit, only one component needs changes (risk R6).
  - (−) Adds an interface layer to maintain.
  - Security: error mapping and data masking at the boundary are centralised in one place.
- **Alternatives Rejected**:
  - *Every component imports the toolkit directly*: violates the Code Style convention and is hard to test.

## Sources

- [desc] Initial description: a Kandev plugin for Nulab Backlog.
- [scope] Workflow-selected scope: `feature`.
- [Q1]–[Q5]: answers in `domain-design-questions.md`.
- C-…: `ideation/feasibility/constraint-register.md`; US…, AC…: `inception/user-stories/stories.md`; `requirements.md`; `team-practices.md`.

## Assumptions & Open Questions

- [assumption] Kandev lets a plugin run timers in its own process (the developer confirmed this when contributing at the story step).
- How to durably store the connection version marker and the deduplication ledger will be settled at the infrastructure design step.
