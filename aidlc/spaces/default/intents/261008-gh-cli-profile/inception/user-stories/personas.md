# Personas — 261008-gh-cli-profile

## P1 — Multi-account developer (primary)

- **Role**: developer who runs Kandev locally with the Nulab Backlog plugin and keeps several GitHub accounts logged in to `gh` (for example a personal account and a client/company account).
- **Goals**: each Kandev workspace talks to GitHub as the right account; switching the active account in a terminal (`gh auth switch`) must not change what a workspace does.
- **Pain points**: today every workspace silently follows the active gh account; PR lists, the "mine" filter and PR links come from the wrong account after a switch, and the workspace's saved account is overwritten by Test.
- **Context**: uses "Use gh CLI login" instead of a personal access token; creates tasks from Backlog issues and lets agents work in task worktrees.

## P2 — Upgrading gh CLI user (secondary)

- **Role**: existing user of v0.5.1–v0.5.2 who already connected GitHub with "Use gh CLI login" in one or more workspaces.
- **Goals**: upgrade the plugin without reconnecting and without any workspace changing account.
- **Pain points**: fears that an upgrade resets settings or picks a different account.
- **Context**: may have only one gh account; the new picker should not get in the way.

## Relationships and Priority

- P1 is the reason for the change and drives the Must stories.
- P2 constrains the change (compatibility); its upgrade story is Must because a silent account change is a data-exposure risk.
