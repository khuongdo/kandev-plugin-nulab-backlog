# Deployment Strategy - v0.5.3

## Strategy

Tag-based release of a plugin package, installed manually on the self-hosted Kandev, per the team Deployment practice. Kandev runs one installed version at a time, so blue/green, canary and rolling do not apply.

No feature flag. Existing gh CLI connections keep their stored GitHub login (`AccountID`), so the change is invisible until a user picks or changes an account.

## Environment Promotion

| Step | Environment | Gate |
|------|-------------|------|
| 1 | PR CI (GitHub Actions) | All required checks green, including the contract test on Kandev 0.96.0 |
| 2 | `main` | Squash merge (self-merge, protected branch) |
| 3 | GitHub Release `v0.5.3` | Creating the tag is the production approval; `release.yml` green; provenance attestation |
| 4 | Self-hosted Kandev | Manual install From URL; smoke checks below |
| 5 | Kandev marketplace | Registry PR, reviewed by the Kandev maintainers |

## Compatibility

- `min_kandev_version` stays `0.96.0`; no new host API. The new action `scm.providers.cli_accounts` is declared in `manifest.yaml`; the contract test passes on 0.96.0.
- Stored settings format unchanged (the chosen login is the existing `AccountID`), so upgrade and downgrade are safe. After a downgrade to 0.5.2 a workspace follows gh's active account again.
- Runtime needs gh ≥ 2.40 on the Kandev server for `--user`; gh ≥ 2.81.0 for the account list. Older gh: only the active account is offered/used.
- Token-mode, GitLab and Bitbucket connections are unchanged.

## Smoke Checks After Install (self-hosted Kandev)

Prerequisite: two GitHub accounts logged in to gh on the Kandev server (`gh auth login` twice), A active.

1. Settings > Plugins shows nulab-backlog 0.5.3 running.
2. An existing gh CLI workspace (from 0.5.2) still shows its old account as "Connected via gh CLI as @login" without reconnecting.
3. On another workspace's GitHub card, "Use gh CLI login" lists A (active in gh) and B; pick B and Connect → card shows @B.
4. Run `gh auth switch` on the server; press Test on both workspaces → each still shows its own login; the PR list "mine" filter differs per workspace.
5. "Change account" on the B workspace lists accounts with B preselected; Cancel changes nothing.
6. `gh auth logout --user B`, then Test on the B workspace → alert "B is not logged in to gh on the Kandev server — log in again or pick another account"; Change account offers A only. Log B back in afterwards.
7. The card in gh CLI mode shows the worktree note naming the chosen login.

Abort condition: the plugin does not start, an existing connection changes account on upgrade, or any workspace uses an account other than its chosen one. Then follow rollback-runbook.md.
