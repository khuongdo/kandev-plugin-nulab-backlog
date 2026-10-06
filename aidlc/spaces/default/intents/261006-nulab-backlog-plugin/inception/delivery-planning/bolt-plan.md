# Bolt Plan — Kandev Plugin for Nulab Backlog

A **Bolt** is one build pass over part of the work (one or more units of work) that ends in something that runs. This plan orders the Bolts, and says for each one what "done" means and what shipping it proves.

Inputs: `unit-of-work` and `unit-of-work-dependency` (the five units and their dependency graph), `unit-of-work-story-map` (stories per unit), `requirements`, `stories`, `mockups`, `components`, `contract-summary`, and `team-practices`. Answers Q1–Q10 are in `delivery-planning-questions.md`.

## Bolt Sequence

| Order | Bolt | Unit(s) | Walking skeleton | WSJF score | Earliest start |
|-------|------|---------|------------------|------------|----------------|
| 1 | B1 walking-skeleton | U1 `walking-skeleton` | Yes | 6.2 | Immediately |
| 2 | B2 ci-release | U5 `ci-release` | No | 3.8 | After B1 approved |
| 3 | B3 connection | U2 `connection` | No | 3.6 | After B1 approved |
| 4 | B4 issues | U3 `issues` | No | 1.8 | After B3 |
| 5 | B5 git-pr | U4 `git-pr` | No | 1.2 | After B3 |
| — | Release `v0.1.0` | all | — | — | After B1–B5 and the second manual check [Q2, Q9] |

Each unit is one Bolt [Q3]. The order follows the dependency graph in `unit-of-work-dependency` exactly. The WSJF scores (see `risk-and-sequencing-rationale.md`) give the same order, so this plan does not deviate from the graph.

**Parallel work** [Q7]: once B1 is approved, units that do not depend on each other may have their code written in the same batch: B2 with B3, then B4 with B5. B2 is merged before B3, so that B3's pull request already passes through the real CI gates [Q1]. A **batch** is the set of units the build runs together; the dependency graph decides batches, while this plan decides merge order inside a batch.

## B1 — walking-skeleton (U1)

The **walking skeleton** is the smallest slice that runs end to end through every integration point (`team-practices`, Walking Skeleton).

- **Stories**: US7.1, US1.2, US1.1, US7.2 (`unit-of-work-story-map`).
- **Proves the architecture parts**:
  - The Go backend served through `pluginsdk.Serve`, with the SDK pinned via `replace ../kandev` and `.kandev-sdk-ref`.
  - The manifest and host contract C8.
  - The `backlog.Client` contract C1, with one call: `users/myself`.
  - The secrets vault.
  - The minimal settings screen M1 and the `connection.get` and `connection.connectApiKey` actions from C5 (`contract-summary`).
  - The `Makefile` targets from C4.
- **Definition of Done**:
  - `make check-format vet lint test coverage build package verify-package` all pass locally, with Go line coverage of at least 80% and `-race`.
  - The package installs on a self-hosted Kandev server at `min_kandev_version`.
  - API-key connect works against a real Backlog space: the settings page shows the Backlog user name and space URL.
  - A wrong key or a non-Backlog host is rejected with the documented message (FR1.2, FR1.3).
  - A test asserts that the API key never appears in logs, errors, or action responses (NFR3).
  - The first manual end-to-end check with a real space is recorded (`team-practices`, Testing Posture).
  - The Construction verification command passes, and you approve the skeleton checkpoint.
- **Confidence hypothesis**: a Go plugin built with the Kandev SDK installs on a self-hosted Kandev and reaches the real Backlog API, with the secret kept out of every output. If this fails, the plugin approach itself has to be rethought (risks R-A and R-E).
- **Expected demo**:
  1. Install the package on self-hosted Kandev.
  2. Open the plugin settings and paste a space URL plus an API key.
  3. See "Connected as …".
  4. Show that the logs contain no key.

## B2 — ci-release (U5)

- **Stories**: US7.3, US7.4, US7.5, US7.6.
- **Definition of Done**:
  - The CI workflow runs on `pull_request`. It calls the C4 targets and checks that `go mod tidy` leaves no diff, that test data holds no real secrets, and that the packaged-host contract test passes on `min_kandev_version`.
  - All actions are pinned by full SHA, with `permissions: contents: read` by default.
  - Branch protection on `main` requires these checks.
  - The release workflow (`release.yml`) is triggered by a `vX.Y.Z` tag on `main`. It refuses existing tags, reruns every check, and publishes the package plus `checksums.txt` with build provenance attestation.
  - The release workflow is proven on a pre-release tag. The real `v0.1.0` waits for the release milestone.
- **Confidence hypothesis**: the quality gates are real gates, so every later Bolt is merged only after CI is green, and a broken SDK upgrade on the Kandev side is caught by the contract test (risk R-E).
- **Expected demo**: a pull request with a failing check is blocked. A pre-release tag produces a GitHub Release whose attestation verifies with `gh attestation verify`.

## B3 — connection (U2)

- **Stories**: US8.4, US8.3, US1.5, US1.6, US1.7, US1.3, US1.4, US1.8, US1.9.
- **Definition of Done**:
  - OAuth sign-in and automatic token refresh work, including the single-refresh guarantee under concurrent calls (`-race`).
  - Re-test, disconnect, credential replacement, space change, and project selection all work.
  - `ConnectionChanged` is emitted with `connectionEpoch` and the restore flag (contract C3).
  - The full rate-limit rules are in place, and structured logging covers errors and waits.
  - The open contract findings owned by this unit are settled in its functional design before code [Q6]:
    - R-01: admin access for `startOAuth`, `setProjects` and `setPollInterval`.
    - R-02: where the OAuth client id and secret come from, and how the redirect URI is built.
    - R-06: the 403 and 400/409/422 error mapping.
    - R-08: the epoch rules per reason, and `workspaceId` in the OAuth state.
  - The open questions on PKCE and the `localhost` callback are answered.
- **Confidence hypothesis**: OAuth against Backlog works through a public Kandev webhook callback, and the plugin respects Backlog rate limits without blocking user actions past the 15-second Kandev action limit (risk R-B).
- **Expected demo**:
  1. Sign in with OAuth.
  2. Wait past token expiry, then make a call that still succeeds.
  3. Disconnect and reconnect, and see links restored.
  4. Feed a fake 429 and see the wait being logged.

## B4 — issues (U3)

- **Stories**: US2.1, US2.2, US3.1, US3.3, US2.3, US3.2, US3.4, US4.1, US4.2, US8.1, US8.2, US8.5, US3.5.
- **Definition of Done**:
  - Issue list, filters, search, create-task-from-issue, link/unlink, the sidebar panel (M8), and `#` references all work against the fake Backlog server and against the real space.
  - The status sync cycle runs. Deleted tasks are handled both by the `task.deleted` event and by the `GetTask` check in each cycle (Q6 of Contract Design).
  - p95 list time is ≤ 3 s (NFR1).
  - WCAG 2.1 AA basics hold, and screens translate (NFR9, NFR10).
  - The open contract findings owned by this unit are settled in its functional design before code [Q6]:
    - R-03: the time budget for multi-call actions and the priority of user calls in the queue.
    - R-04: full schemas for the U3 actions and `reference_sources`.
    - R-05: the internal `TaskDeleted` contract.
  - The open questions on writing real labels and on the `issues.createTask` time budget are answered.
- **Confidence hypothesis**: the core value works. A developer can find a Backlog issue, turn it into a Kandev task, and see its status follow Backlog, all within the performance and rate-limit budgets.
- **Expected demo**:
  1. Search `PROJ-12` and create a task from it.
  2. See the issue badge on the card.
  3. Change the status on Backlog and see the badge update within one cycle.
  4. Delete the task and see the link disappear.

## B5 — git-pr (U4)

- **Stories**: US5.5, US5.1, US5.6, US5.2, US5.4, US5.3, US6.1, US6.2, US6.3.
- **Definition of Done**:
  - Git credentials can be stored.
  - The repository source lets Kandev clone and push.
  - Pull requests can be linked, created, and shown as status badges.
  - PR watches create at most one task per PR, never recreate deleted tasks, and resume after restart.
  - Saved queries are available on the dashboard.
  - The U4 action schemas (`git.*`) are defined in its functional design.
  - The Git-related acceptance criteria of US1.5, US1.8, and US1.9 pass via `ConnectionChanged`.
- **Confidence hypothesis**: Backlog Git hosting works as a Kandev repository source end to end. This includes the assumption that pull requests exist on the test space's plan (assumption A1 in `requirements`).
- **Expected demo**:
  1. Create a task on a Backlog repository and let the agent push a branch.
  2. Create a PR from the task and see the PR badge.
  3. Let a watch create a task for a new PR.

## Release Milestone — `v0.1.0`

- **Contents**: B1–B5. The first release waits for U4, even if U4 slips [Q2, Q9].
- **Gate**:
  - The second manual end-to-end check with a real Backlog space is recorded (`team-practices`).
  - Package verification passes before tagging (`project.md`, Mandated).
  - You push the `v0.1.0` tag, which is the manual production approval.
- **After the release**: open a pull request to the Kandev marketplace registry (US7.6). From then on, the project relies on automated tests only.

## Sources

- [desc] Initial description: Kandev plugin for Nulab Backlog, modelled on the Bitbucket plugin.
- [scope] Workflow-selected scope: `feature`.
- [Q1]–[Q10]: answers in `delivery-planning-questions.md`.
- `unit-of-work.md`, `unit-of-work-dependency.md`, `unit-of-work-story-map.md` (Units Generation); `contract-summary.md` (Contract Design); `components.md` (Domain Design); `requirements.md`; `stories.md`; `mockups.md`; `team-practices.md`.

## Assumptions & Open Questions

- [assumption] WSJF inputs (value, urgency, risk reduction, size) are relative judgements on the Fibonacci scale, not measurements.
- [assumption] Inside a parallel batch, merging B2 before B3 is enough to honour "CI before connection" [Q1]. B3's code may be written while B2 is still open.
- U4 is optional in `unit-of-work`, but the first release now depends on it [Q9]. Dropping U4 later means changing this plan.
