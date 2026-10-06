# Story Map by Unit — Kandev Plugin for Nulab Backlog

Inputs:

- `stories`: US1.1–US8.5.
- `unit-of-work.md`.
- `components` and `decisions` from the Domain Design step.
- `requirements`.

## Stories by Unit

| Story | Priority | Unit ID | Directory | Order within unit |
|-------|---------|---------|-----------|---------------------|
| US7.1 | Must | U1 | u1-walking-skeleton | 1 |
| US1.2 | Must | U1 | u1-walking-skeleton | 2 |
| US1.1 | Must | U1 | u1-walking-skeleton | 3 |
| US7.2 | Must | U1 | u1-walking-skeleton | 4 |
| US8.4 | Must | U2 | u2-connection | 1 |
| US8.3 | Could | U2 | u2-connection | 2 |
| US1.5 | Must | U2 | u2-connection | 3 |
| US1.6 | Must | U2 | u2-connection | 4 |
| US1.7 | Must | U2 | u2-connection | 5 |
| US1.3 | Must | U2 | u2-connection | 6 |
| US1.4 | Must | U2 | u2-connection | 7 |
| US1.8 | Must | U2 | u2-connection | 8 |
| US1.9 | Must | U2 | u2-connection | 9 |
| US2.1 | Must | U3 | u3-issues | 1 |
| US2.2 | Must | U3 | u3-issues | 2 |
| US3.1 | Must | U3 | u3-issues | 3 |
| US3.3 | Must | U3 | u3-issues | 4 |
| US2.3 | Must | U3 | u3-issues | 5 |
| US3.2 | Must | U3 | u3-issues | 6 |
| US3.4 | Must | U3 | u3-issues | 7 |
| US4.1 | Must | U3 | u3-issues | 8 |
| US4.2 | Must | U3 | u3-issues | 9 |
| US8.1 | Must | U3 | u3-issues | 10 |
| US8.2 | Must | U3 | u3-issues | 11 |
| US8.5 | Should | U3 | u3-issues | 12 |
| US3.5 | Should | U3 | u3-issues | 13 |
| US5.5 | Should | U4 | u4-git-pr | 1 |
| US5.1 | Should | U4 | u4-git-pr | 2 |
| US5.6 | Should | U4 | u4-git-pr | 3 |
| US5.2 | Should | U4 | u4-git-pr | 4 |
| US5.4 | Should | U4 | u4-git-pr | 5 |
| US5.3 | Should | U4 | u4-git-pr | 6 |
| US6.1 | Should | U4 | u4-git-pr | 7 |
| US6.2 | Should | U4 | u4-git-pr | 8 |
| US6.3 | Could | U4 | u4-git-pr | 9 |
| US7.3 | Must | U5 | u5-ci-release | 1 |
| US7.4 | Should | U5 | u5-ci-release | 2 |
| US7.5 | Should | U5 | u5-ci-release | 3 |
| US7.6 | Should | U5 | u5-ci-release | 4 |

The order within each unit follows the story dependencies recorded in `stories.md`. For example, in U2 the API call limit (US8.4) comes first because every later call needs it. The order between units is decided by the Delivery Planning step.

## Stories That Touch Several Units

| Story | Owning unit | Related units | Notes |
|-------|-----------|------------------|---------|
| US1.5, US1.8, US1.9 | U2 | U4 | The AC parts about the Git password, PR links and PR watch are tested in U4, through the ConnectionChanged event |
| US8.2, US8.5 | U3 | U4 | The accessibility and multi-language rules also apply to the U4 screens |
| US7.3 | U5 | U1–U4 | The CI gate applies to the code of every unit |
| US5.5 | U4 | U2 | Secret storage lives in the Connection component but is built together with U4 |

## Coverage Check

- Stories in `stories.md`: 39. Stories assigned to a unit: 39. No story is left unassigned.
- Every unit has stories: U1 has 4, U2 has 9, U3 has 13, U4 has 9, U5 has 4.
- No Must unit (U1, U2, U3, U5) depends on U4.

## Sources

- [desc] Initial description: plugin Kandev cho Nulab Backlog.
- [scope] Workflow-selected scope: `feature`.
- `stories.md`, `unit-of-work.md`, `components.md`, `decisions.md`, `requirements.md`.

## Assumptions & Open Questions

None.
