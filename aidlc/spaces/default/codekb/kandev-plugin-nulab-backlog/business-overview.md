# Business Overview — kandev-plugin-nulab-backlog

## Business Context

A Kandev plugin (`id: nulab-backlog`, `version: 0.1.1`, `min_kandev_version: 0.96.0`, MIT) that connects a Kandev workspace to a Nulab Backlog space. Users are teams that hand work to AI agents in Kandev but keep issues, Git repositories and pull requests in Backlog. Released through GitHub Releases and the Kandev marketplace (public repo `khuongdo/kandev-plugin-nulab-backlog`); installation on the self-hosted Kandev server is manual.

## Key Functionality

| Area | What it does | Who uses it |
|---|---|---|
| Connection | Connect a space with an API key or OAuth (Sign in with Nulab), test, disconnect, change space, pick projects, enable/disable per workspace, store a Git credential | Workspace admin |
| Issues | Issue list on `/backlog` (Issues tab: search, Project/Status/Assignee filters, 20 rows per page, "..." row menu with Create task / Link to task), link/unlink task and issue, issue panel in a task, card badge, `#` reference source, status sync loop on a per-workspace poll interval (1–1440 min); issue watches that create at most one task per watch per run on a per-watch interval (default 5 min), managed in Settings | Members; poll interval set by an admin |
| Git / PR | Backlog Git as a Kandev repository provider; link a PR to a task, PR status, create PR, review provider; PR list on `/backlog` (Pull requests tab, `git.prs.list`); saved PR queries used as a "Query" preset in the PR list and listed in Settings (rename/delete); PR watches that create review tasks, managed in Settings | Signed-in members |

Tasks from issues are created by the plugin itself (`issues.create_task`: title = issue summary, description = issue description + Backlog link). There are no prompt templates / quick actions, no start-task action on PR rows, no saved issue queries and no default query: the PR list opens empty until a repository or saved query is chosen. Details: [code-quality-assessment.md](code-quality-assessment.md#intent-261007-github-parity-actions-risks).

## Business Rules Locked by Code and Tests

- Only `https` space addresses under `backlog.com`, `backlog.jp`, `backlogtool.com` are accepted.
- API keys, tokens and Git passwords live only in the Kandev secret store; never returned to the browser or logged (`internal/redact`).
- The integration is enabled by default; when disabled the plugin refuses Backlog calls but keeps the connection. The settings card, switch and nav entry never depend on the enabled state (BR5.4/BR7.6/BR7.8).
- Connection-changing actions and `issues.set_poll_interval` are `admin`; other issue and Git actions (including `git.queries.*`, `git.prs.list`, `issues.watches.*`) are `authenticated` ([api-documentation.md](api-documentation.md)).
- A saved PR query needs a non-empty name (at most 100 runes), a repository in a selected project, at least one status, and assignee/creator `anyone|me`; at most 50 per workspace (`internal/git/types.go:179-215`, `internal/git/store.go`).
- The Backlog logo is used under Nulab brand terms that forbid modified or recoloured versions (`docs/brand/backlog-logo.md`).

## Current Intent (261007-github-parity-actions, express, Minimal)

Match Kandev's first-party GitHub integration: (1) customizable quick actions (default Implement / Investigate prompts plus user-added ones) to start a task from an issue or PR; (2) a preinstalled default saved query for Issues and for PRs; (3) GitHub-aligned page margins and button positions on `/backlog`. External reference: [architecture.md](architecture.md#external-reference-kandev-github-integration-v0960). Risks: [code-quality-assessment.md](code-quality-assessment.md#intent-261007-github-parity-actions-risks).

Previous intent `261007-uiux-github-style` (refactor, released as v0.1.1) delivered the single `/backlog` entry with Issues/PRs tabs, watches and saved queries in Settings, issue watches and host-styled controls.
