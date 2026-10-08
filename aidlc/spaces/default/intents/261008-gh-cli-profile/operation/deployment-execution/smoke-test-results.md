# Smoke Test Results — v0.5.3 on the self-hosted Kandev

Server: local Kandev at `http://localhost:38429` (behind `https://webfrontier.tail152aaa.ts.net`). gh on the server has two github.com accounts (`khuongdo`, `khuongdo-nicosys`; `khuongdo-nicosys` active).

| # | Check (deployment-strategy.md) | Result | Evidence |
|---|---|---|---|
| 1 | Plugin installed and running | **Pass** | `GET /api/plugins` → `nulab-backlog 0.5.3 active`; install response HTTP 201, `min_kandev_version 0.96.0` |
| — | New action reachable and gated | **Pass** | `POST /api/plugins/nulab-backlog/actions/scm.providers.cli_accounts` (workspace `kandev`) → HTTP 409 `integration_disabled`, the plugin's normal answer while the per-workspace switch is off |
| — | Existing provider state unchanged after upgrade | **Pass** | `scm.providers.list` (workspace `kandev`) → github/gitlab/bitbucket `not_configured`, same as before |
| 2 | Existing gh CLI workspace keeps its account | Not run | no workspace on this server had a gh CLI connection |
| 3 | Picker lists both accounts; pick B | Not run | Backlog integration is off in all three workspaces (`kandev`, `contract`, `monitoring`) |
| 4 | `gh auth switch` does not change workspace accounts | Not run | same |
| 5 | Change account / Cancel | Not run | same |
| 6 | Logged-out account message | Not run | same |
| 7 | Worktree note on the card | Not run | same |

## Verdict

Install verified; the interactive checks 2-7 are pending the owner (they need the integration switched on in a workspace and a GitHub gh CLI connection). The same behaviours are covered by automated tests (AC1.1.1-AC4.1.1) that passed in Build and Test and in release CI.
