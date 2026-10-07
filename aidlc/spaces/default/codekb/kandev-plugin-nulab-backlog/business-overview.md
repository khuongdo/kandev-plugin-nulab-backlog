# Business Overview — kandev-plugin-nulab-backlog

## Business Context

A Kandev plugin (`id: nulab-backlog`, `version: 0.1.0`, `min_kandev_version: 0.96.0`, MIT) that connects a Kandev workspace to a Nulab Backlog space. Users are teams that hand work to AI agents in Kandev but keep issues, Git repositories and pull requests in Backlog. Released through GitHub Releases and the Kandev marketplace (public repo `khuongdo/kandev-plugin-nulab-backlog`); installation on the self-hosted Kandev server is manual.

## Key Functionality

| Area | What it does | Who uses it |
|---|---|---|
| Connection | Connect a space with an API key or OAuth (Sign in with Nulab), test, disconnect, change space, pick projects, enable/disable per workspace, store a Git credential | Workspace admin |
| Issues | Issue list (search, 3 filters, 20 rows per page), create a task from an issue, link/unlink task and issue, issue panel in a task (comments, attachments), card badge, `#` reference source, status sync loop on a per-workspace poll interval (1–1440 min) | Members; poll interval set by an admin |
| Git / PR | Backlog Git as a Kandev repository provider; link a PR to a task, PR status, create PR, review provider; PR watches that create review tasks (fixed 5-minute timer); saved PR queries shown on a dashboard (max 20 PRs of one repository) | Signed-in members |

There is no "issue watch" that creates tasks from new Backlog issues (GitHub has one); see [code-quality-assessment.md](code-quality-assessment.md#intent-261007-risks).

## Business Rules Locked by Code and Tests

- Only `https` space addresses under `backlog.com`, `backlog.jp`, `backlogtool.com` are accepted.
- API keys, tokens and Git passwords live only in the Kandev secret store; never returned to the browser or logged (`internal/redact`).
- The integration is enabled by default; when disabled the plugin refuses Backlog calls but keeps the connection. The settings card, switch and nav entry never depend on the enabled state (BR5.4/BR7.6/BR7.8).
- Connection-changing actions and `issues.set_poll_interval` are `admin`; other issue and Git actions are `authenticated` ([api-documentation.md](api-documentation.md)).
- The Backlog logo is used under Nulab brand terms that forbid modified or recoloured versions (`docs/brand/backlog-logo.md`).

## Current Intent (261007-uiux-github-style, refactor, Minimal)

Restyle the UI after Kandev's first-party GitHub integration: (1) PR and issue watcher settings inside Settings > Integrations > Backlog; (2) a single Backlog entry under Home > Integrations holding an Issue list / PR list; (3) textboxes and buttons in the GitHub style; (4) an outline-only logo matching the existing icons. Mapping to the GitHub reference: [architecture.md](architecture.md#external-reference-kandev-github-integration-ui). Risks: [code-quality-assessment.md](code-quality-assessment.md#intent-261007-risks).
