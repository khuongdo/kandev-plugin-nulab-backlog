# Persona — Kandev Plugin for Nulab Backlog

Inputs: `requirements.md` (FR1–FR7, NFR1–NFR11), `team-practices.md`, and the answers in `user-stories-questions.md`.

## Primary Persona

As you chose, there is one shared persona (Q1). This persona is also used for the release stories (Q7).

### Minh — a Kandev user who uses Backlog

| Attribute | Description |
|-----------|-------|
| Role | A development team member who uses Nulab Backlog daily to manage issues and source code, and uses Kandev to give tasks to AI agents. When needed, Minh also installs the plugin, connects the space, and releases new versions of the plugin |
| Goals | Pick a Backlog issue and give it to an agent to work on in Kandev without copying by hand; see issue and pull request status right on the task board; install new plugin versions reliably |
| Current pain points | Kandev does not support Backlog yet, so they must switch between two tools; issue titles and descriptions must be copied by hand; they cannot tell whether a task's pull request has been merged without opening Backlog |
| Tech familiarity | High: familiar with Git, pull requests, and working with agents |
| Frequency of use | Daily |
| Context | A self-hosted Kandev workspace, with one Backlog connection shared by the whole workspace (`requirements.md` FR1.1). So when Minh connects, changes the space or disconnects, **everyone in the workspace** is affected. The space is on one of the domains `backlog.com`, `backlog.jp`, `backlogtool.com` |
| Devices | Minh works mainly on a computer, but often checks the task board on a phone when away from the desk |
| Data | Issue titles may be in Japanese, Vietnamese, or very long |
| Source code | Source code may live on Backlog's Git, or elsewhere such as GitHub. So the plugin's Git part is optional (Q8) |

## Priority Order

There is only one persona, so no ranking is needed. Stories are grouped by feature, not by role (Q2). Stories for administration (connecting, choosing projects) or for releasing state that context right in the story sentence.

## Sources

- [desc] Initial description: a Kandev plugin for Nulab Backlog.
- [scope] Workflow-selected scope: `feature`.
- [Q1], [Q7]: answers in `user-stories-questions.md`.
- `requirements.md`, `stakeholder-map.md`.

## Assumptions & Open Questions

- [assumption] The name "Minh" is only illustrative and does not represent a specific person.
