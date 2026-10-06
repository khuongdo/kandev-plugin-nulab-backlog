# Units of Work — Kandev Plugin for Nulab Backlog

Inputs:

- `components` (6 components) and `decisions` (ADR-001 to ADR-007), from the Domain Design step.
- `requirements` (FR, NFR).
- `stories` (US1.1–US8.5).
- The answers to Q1–Q4 in `units-generation-questions.md`.

Each unit is a vertical feature slice [Q1]. The backend and the screens of a feature sit in the same unit. There are 5 units in total, split at a coarse level [Q2].

## Unit Table

| Unit ID | Directory | Name | Kind | Priority | Size | Deployment |
|---------|-----------|-----|------|---------|--------|------------|
| U1 | u1-walking-skeleton | walking-skeleton | service | Must | M | Embedded in the single plugin package |
| U2 | u2-connection | connection | service | Must | L | Embedded in the single plugin package |
| U3 | u3-issues | issues | service | Must | XL | Embedded in the single plugin package |
| U4 | u4-git-pr | git-pr | service | Should (optional) | XL | Embedded in the single plugin package |
| U5 | u5-ci-release | ci-release | packaging | Must | M | The repo's CI and release process |

The plugin is always deployed as **one single package**, made of the Go backend and a UI bundle. All four service units are embedded in that package; no unit is deployed separately.

## U1 — walking-skeleton

- **Description**: the smallest thin slice that runs end to end (`team-practices`, Walking Skeleton).
- **Boundary**:
  - The minimal part of KandevAdapter: `server/main.go`, settings page registration, the SDK dependency mechanism through `replace` and `.kandev-sdk-ref`.
  - Connection: only API key connection and space address validation.
  - BacklogGateway: can make one call (get the current user), with secret redaction and a time limit.
  - PluginUI: a reduced M1 screen.
  - Makefile: build, package, verify-package.
- **Responsibility**: prove that the chain package → install on self-hosted Kandev → connect → call real Backlog works.
- **Notes**: per `team-practices`, this unit is complete when the Construction verification command runs for real, the first manual check with a real space is done, and you approve the skeleton checkpoint. This unit has no OAuth yet.

## U2 — connection

- **Description**: complete the Connection component and BacklogGateway.
- **Boundary**:
  - OAuth and automatic token refresh.
  - Re-check and disconnect.
  - Replace credentials, change space.
  - Select and deselect projects.
  - The ConnectionChanged event with a version stamp, and restore after reconnect.
  - The full API call limit rules (429, backoff, per-group queues).
  - Structured logs.
  - The matching part of M1, plus the M12 restore notice.
- **Responsibility**: everything about connecting and calling Backlog safely. Later units rely on this part.
- **Notes**:
  - Some ACs of US1.5, US1.8 and US1.9 mention the Git password, PR links and PR watch. Those parts only exist if U4 is built.
  - In U2, these ACs are tested for the part about issues and connection. The Git part is tested in U4, through the ConnectionChanged event (per review finding R-03 at the story step).

## U3 — issues

- **Description**: the whole issue part.
- **Boundary**:
  - IssueIntegration.
  - Screens M2, M2m, M3, M8, plus the issue badge on the card (M6) and the `#` source (M10).
  - The issue status sync cycle.
  - List performance.
  - Accessibility and multi-language support for the plugin screens.
- **Responsibility**: the main value of the product (Must).
- **Notes**:
  - This is the largest unit. Inside, it should be split by story.
  - US8.2 and US8.5 apply to every plugin screen, but their owner is here because most screens belong to this unit. U4 must apply the same rules to its own screens.

## U4 — git-pr (optional)

- **Description**: the whole GitIntegration component.
- **Boundary**:
  - Git credential storage (US5.5; the secret storage lives in Connection, but is built together with this unit).
  - Repository source.
  - Link and create pull requests, PR badge.
  - PR watch and saved queries.
  - Screens M4, M5, M7, M9, M11.
- **Responsibility**: the Git/PR part. Dropping this unit or doing it later does not affect the Must units [Q4].
- **Notes**:
  - Has its own timer (ADR-003).
  - Receives the ConnectionChanged event from U2.

## U5 — ci-release

- **Description**: the CI and release process.
- **Boundary**:
  - CI workflow with format, vet, lint, test `-race`, coverage, `go mod tidy` check, packaging, package verification, and a check that test data contains no real secrets.
  - Contract test on the minimum Kandev version.
  - Release workflow with build provenance attestation.
  - Marketplace submission.
- **Responsibility**: a real quality gate, and a trustworthy release.
- **Notes**:
  - A packaging unit, so it needs no business model.
  - Releasing the first version needs U3 done and a second manual check. This ordering is decided by the Delivery Planning step.

## Sources

- [desc] Initial description: plugin Kandev cho Nulab Backlog.
- [scope] Workflow-selected scope: `feature`.
- [Q1]–[Q4]: the answers in `units-generation-questions.md`.
- `components.md`, `decisions.md` from the Domain Design step; `requirements.md`; `stories.md`; `team-practices.md`.

## Assumptions & Open Questions

- [assumption] The S/M/L/XL sizes are relative estimates, not yet based on data.
- [assumption] U3 is size XL. If it is too big for one pass, it can be split further at the Delivery Planning step without changing the dependency graph.
