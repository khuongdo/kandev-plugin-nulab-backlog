# User Story Need Assessment

## Decision

**Execute**: write user stories.

## Reasons

- The plugin has many features users interact with directly: connecting a space, browsing issues, creating and linking tasks, creating pull requests, watching PRs. The scope settled in `requirements.md` covers FR1 to FR7.
- There are at least two user groups with different goals. One side installs and connects; the other works on tasks every day. There is also the side that operates and releases the plugin.
- There are multi-step flows and business rules that need clear acceptance criteria: linking one issue to many tasks, one-way sync, PR watches creating tasks automatically, API rate limits.

## Factors Considered

| Factor | Assessment |
|--------|-----------|
| Project type | Brand-new (greenfield), new feature |
| User-facing scope | Has UI inside Kandev (W1–W10 in `wireframes.md`) |
| Complexity | Standard: one external system, two sign-in methods, periodic polling |

## Where User Stories Add the Most Value

- Given/When/Then acceptance criteria for writing tests first, because the team follows TDD (`team-practices`).
- Error and edge cases that the requirements step left open: connecting a second time, a PR watch's first run, deleted issues, empty lists.
- Assigning priorities and marking which parts belong to the thin slice done first.

## Sources

- [desc] Initial description: a Kandev plugin for Nulab Backlog.
- [scope] Workflow-selected scope: `feature`.
- `requirements.md`, `team-practices.md`, `wireframes.md`.

## Assumptions & Open Questions

None.
