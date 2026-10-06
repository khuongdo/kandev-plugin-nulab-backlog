# Team Allocation — Kandev Plugin for Nulab Backlog

A **Bolt** is one build pass over part of the work that ends in something that runs (see `bolt-plan.md`). A **mob** is the group that builds a Bolt together. This file says who builds and who approves each Bolt.

Inputs: `unit-of-work`, `unit-of-work-dependency`, `unit-of-work-story-map`, `requirements`, `stories`, `mockups`, `components`, `contract-summary`, `team-practices`, and answers Q7–Q8 in `delivery-planning-questions.md`.

## Staffing Model

- Team Formation was skipped for this workflow: one person works with AI (C-O1 in `requirements`).
- All Bolts are executed in this session by the AI developer (`aidlc-developer-agent`), with the architect, quality and security perspectives applied in their stages [Q8].
- You are the only approver: plan approval before code, checkpoint approvals, the skeleton checkpoint, and the release tag.
- There is one team, so there is no cross-team Program Board.

## Bolt-to-Mob Assignment

| Bolt | Unit | Builder | Approver | Parallel with |
|------|------|---------|----------|---------------|
| B1 walking-skeleton | U1 | AI developer (this session) | You, including the skeleton checkpoint and the first manual check with a real space | — (runs alone, first) |
| B2 ci-release | U5 | AI developer (this session) | You | B3; B2 merges first [Q1] |
| B3 connection | U2 | AI developer (this session) | You | B2 |
| B4 issues | U3 | AI developer (this session) | You | B5 |
| B5 git-pr | U4 | AI developer (this session) | You, plus the second manual check before `v0.1.0` | B4 |

## Working Agreements

- Every Bolt goes through a pull request to the protected `main` and is squash-merged when CI is green (`team-practices`, Way of Working).
- Tests are written first (TDD, `team-practices`, Testing Posture), with at least 80% Go line coverage and `-race`.
- Parallel batches are built in separate worktrees. Your approval stays one decision at a time.

## Sources

- [desc] Initial description: Kandev plugin for Nulab Backlog.
- [scope] Workflow-selected scope: `feature`.
- [Q7], [Q8]: answers in `delivery-planning-questions.md`.
- `bolt-plan.md`; `unit-of-work.md`; `unit-of-work-dependency.md`; `unit-of-work-story-map.md`; `contract-summary.md`; `components.md`; `requirements.md`; `stories.md`; `mockups.md`; `team-practices.md`.

## Assumptions & Open Questions

None.
