# External Dependency Map — Kandev Plugin for Nulab Backlog

This file lists the things outside the build itself that a Bolt waits for. A **Bolt** is one build pass over part of the work that ends in something that runs (see `bolt-plan.md`).

Inputs: `requirements` (constraints and assumptions A1–A4), `contract-summary` (open questions on C6, C7 and C8), `team-practices`, `unit-of-work`, `unit-of-work-dependency`, `unit-of-work-story-map`, `components`, `stories`, `mockups`, and the Q2 answer at Approval & Handoff (create the test space and the OAuth app during the build, when needed).

## Dependencies

| # | Dependency | Owner | Lead time | Blocks | If it slips |
|---|------------|-------|-----------|--------|-------------|
| X1 | Backlog test space with an API key (Free plan or trial) | You | Under 1 hour | B1: first manual check; the skeleton checkpoint | The skeleton cannot be approved, because a real-space check is part of its Definition of Done (`team-practices`) |
| X2 | Self-hosted Kandev server at `min_kandev_version` | You | Under 1 day | B1: install check | Same as X1 |
| X3 | Local `../kandev` checkout at the commit pinned in `.kandev-sdk-ref` | You (already present) | None | B1: build | — |
| X4 | Public GitHub repository with branch protection on `main` and GitHub Actions enabled | You | Under 1 hour | B2 | B3 onward cannot merge through real gates; work continues, but merges wait |
| X5 | OAuth application registered with Nulab for the test space, with the Kandev webhook redirect URI | You (Nulab developer settings) | 1 day or less, assumed; not verified | B3: OAuth stories US1.3 and US1.4 | Build the rest of B3 with API key; OAuth stories wait (contract open questions on PKCE and `localhost`) |
| X6 | Pull requests and Git hosting available on the test space's plan (assumption A1) | Nulab plan; you check | Under 1 hour to check | B5 | Upgrade the test space plan or use a trial; B5 waits |
| X7 | Workspace admin grant for `host.v2.write:tasks`, if real task labels are chosen | You as Kandev admin | Minutes | B4 (open contract question on labels) | Show labels through the UI slot only |
| X8 | Kandev marketplace maintainers accept the registry pull request | Kandev maintainers | Unknown | Listing after `v0.1.0` (US7.6); does not block the release | The plugin stays installable from the GitHub Release |

The rest of the build needs nothing outside this team.

## Sources

- [desc] Initial description: Kandev plugin for Nulab Backlog.
- [scope] Workflow-selected scope: `feature`.
- [Q2], [Q9]: answers in `delivery-planning-questions.md`.
- `requirements.md` (A1, A2, C-T2); `contract-summary.md`; `team-practices.md`; `unit-of-work.md`; `unit-of-work-dependency.md`; `unit-of-work-story-map.md`; `components.md`; `stories.md`; `mockups.md`; `bolt-plan.md`.

## Assumptions & Open Questions

- [assumption] Lead times are rough estimates, not commitments from Nulab or the Kandev maintainers.
