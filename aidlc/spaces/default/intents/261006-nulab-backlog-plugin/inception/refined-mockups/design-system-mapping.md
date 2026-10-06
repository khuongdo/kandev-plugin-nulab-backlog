# Design System Mapping — Kandev Plugin for Nulab Backlog

Inputs: `mockups.md`, `interaction-spec.md`, and `wireframes` from the rough-mockups step. Per the decision at the rough-mockups step, the UI uses Kandev's existing components and style, with no separate identity (`rough-mockups` Q4). Additional reference sources: `stories`, `requirements`, `user-flow`, `team-practices`.

## Kandev Mount Points

| Screen | Mount point in Kandev | Who renders | Notes |
|----------|-----------------------|-------|---------|
| M1 | Integrations settings page | Plugin | Same as the Bitbucket plugin |
| M2, M4, M5 | Entry under "Integrations" in the sidebar | Plugin | Three subpages: Issues, PR watches, Dashboard |
| M6 | Labels on the task card (`task-card-tags`) and task menu | Kandev renders, plugin provides data | PR status is refreshed by Kandev itself |
| M8 | Task detail sidebar (`task-sidebar`) | Plugin | Collapsible block |
| M9 | Repository source (`repository_providers`) | Kandev | Plugin provides the list and Git info |
| M10 | `#` reference source (`reference_sources`) | Kandev | Plugin provides search results |
| M11 | Kandev's PR creation flow | Kandev | Plugin performs the PR creation on Backlog |

The mount point names come from the developer's contribution at the story step (`contributions/aidlc-developer-agent.md`), which was checked against Kandev's plugin documentation and source code.

## Component Mapping

| Plugin component | Reused Kandev component | Notes |
|-------------------|---------------------------|---------|
| BacklogListState (loading) | Skeleton rows | No full-page covering spinner |
| BacklogListState (empty, error) | Empty state, inline alert | Errors have a Retry button |
| ConfirmDialog | Alert dialog | Default focus on Cancel |
| ConnectionForm | Form, text input, password input, radio group, checkbox list | Secret fields are never filled back in |
| IssueRow | Data table row | The "…" menu uses a menu button |
| IssueCard | Card | Used on phones |
| PullRequestLinkDialog | Dialog, text input, inline status | — |
| BacklogSidebarPanel | Collapsible section, definition list | Read-only |
| WatchFormDialog, SaveQueryDialog | Dialog, fieldset, select, checkbox group | — |
| Issue/PR status labels | Badge / tag | Always has text, never relies on color alone |
| Notifications | Toast | Success hides itself; errors must be closed by hand |

## Design Tokens

The plugin defines no colors, font sizes or spacing of its own. Every value uses Kandev's tokens. This way the UI follows Kandev's light/dark mode and theme automatically.

| Purpose | Kandev token (by meaning) |
|----------|---------------------------|
| Open / In Progress status | Info color |
| Resolved / Merged status | Success color |
| Closed status | Neutral color |
| "not connected", "may be out of date" status | Warning color, with text |
| Error | Error color, with icon and text |

## Breakpoints

| Breakpoint | Behaviour |
|-----------|---------|
| < 768px | Cards stacked vertically; filters collapse into a drawer; touch targets at least 44x44px; check at 320px |
| 768–1024px | Condensed table |
| > 1024px | Full table |

## Sources

- [desc] Initial description: a Kandev plugin for Nulab Backlog, similar to the Bitbucket plugin.
- [scope] Workflow-selected scope: `feature`.
- [Q3], [Q6]: answers in `refined-mockups-questions.md`.
- `contributions/aidlc-developer-agent.md` at the story step: the Kandev mount points.

## Assumptions & Open Questions

- [assumption] The Kandev component names in the table above are by meaning. The real names in Kandev's component library will be checked when writing the UI code.
