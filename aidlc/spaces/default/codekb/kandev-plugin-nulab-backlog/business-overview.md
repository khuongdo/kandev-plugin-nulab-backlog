# Business Overview — kandev-plugin-nulab-backlog

## Business Context

A Kandev plugin (`id: nulab-backlog`, `version: 0.3.0`, `min_kandev_version: 0.96.0`, MIT) that connects a Kandev workspace to a Nulab Backlog space. Users are teams that hand work to AI agents in Kandev but keep issues, Git repositories and pull requests in Backlog. Released through GitHub Releases and the Kandev marketplace (public repo `khuongdo/kandev-plugin-nulab-backlog`); installation on the self-hosted Kandev server is manual.

## Key Functionality

| Area | What it does | Who uses it |
|---|---|---|
| Connection | Connect a space with an API key or OAuth (Sign in with Nulab), test, disconnect, change space, pick projects, turn Backlog on/off per workspace (opt-in since v0.3.0), store one Backlog Git credential | Workspace admin |
| Issues | Issue list on `/backlog` (Issues tab: search, filters, row menu), link/unlink task and issue, issue panel, card badge, `#` reference source, status sync loop, issue watches, quick actions (`issues.quick_actions.*`) and saved issue queries with a default (`issues.queries.*`) | Members; poll interval set by an admin |
| Git / PR | **Backlog Git only**: Backlog Git as the single Kandev repository provider (`nulab-backlog`), Git credential lease for clone/push, link a PR to a task, PR status, create PR, review provider, PR list on `/backlog`, saved PR queries with a default (`git.queries.set_default`), PR watches | Signed-in members; Git credential set by an admin |

"Source control" is not a separate concept today: Git means the Backlog Git of the connected Backlog space, reached with the Backlog connection's credentials and scoped to its selected projects ([architecture.md](architecture.md#source-control-coupling-intent-261007-source-control-agnostic)).

## Business Rules Locked by Code and Tests

- Only `https` space addresses under `backlog.com`, `backlog.jp`, `backlogtool.com` are accepted (project Mandated rule; `backlog.Client` pins every request to the space host).
- API keys, tokens and Git passwords live only in the Kandev secret store; never returned to the browser or logged (`internal/redact`).
- Backlog is **off after installation** until an admin turns it on (`connection.set_enabled`; no switch record = off, `internal/connection/store.go:212-229`). While off, every action except `connection.get` and `connection.set_enabled` is refused, and both Git credential RPCs fail closed (`internal/plugin/runtime.go` `guarded`). The settings card, switch and nav entry never depend on the enabled state.
- Connection-changing actions and `issues.set_poll_interval` are `admin` ([api-documentation.md](api-documentation.md)).
- Git items (repositories, PR links, watches, queries, credential leases) are allowed only for the connected Backlog space and its selected projects; a project deselect, space change or disconnect disables them (`internal/git/events.go`).
- The Git credential is valid only for the connected Backlog host and is deleted on disconnect and on a space change (`internal/connection/store.go:273,383,683-685`).
- A saved PR query needs a non-empty name (at most 100 runes), a repository in a selected project, at least one status, and assignee/creator `anyone|me`; at most 50 per workspace.
- The Backlog logo is used under Nulab brand terms that forbid modified or recoloured versions (`docs/brand/backlog-logo.md`).

## Current Intent (261007-source-control-agnostic, express, Minimal)

Verbatim: "Backlog integration có thể link với nhiều source control khác nhau, ví dụ: github, backlog, bitbucket.... Trong settings có thể setup được những điều kiện để liên kết với source control service tương ứng (auth, repo, space...)". In short: link Backlog work to several source-control services (GitHub, Backlog Git, Bitbucket, ...) and configure each service's linking conditions (auth, repository, space/owner) in Settings. What must change and the open decisions: [architecture.md](architecture.md#source-control-coupling-intent-261007-source-control-agnostic), risks in [code-quality-assessment.md](code-quality-assessment.md#intent-261007-source-control-agnostic-risks).

Earlier intents: `261007-opt-in-default` (v0.3.0, Backlog opt-in), `261007-github-parity-actions` (v0.2.0, quick actions, default queries, GitHub-style layout), `261007-uiux-github-style` (v0.1.1, single `/backlog` entry, watches and saved queries in Settings).
