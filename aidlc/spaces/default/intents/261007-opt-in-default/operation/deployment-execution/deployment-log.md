# Deployment Log — 261007-opt-in-default (v0.3.0)

## Scope Executed

Per the confirmed answers (Q1 = A, Q2 = C): prepare and open the pull request, then stop. Merge, tag `v0.3.0`, `release.yml`, marketplace registry pull request and installation are left to the maintainer.

## Steps

| # | Step | Result |
|---|---|---|
| 1 | Commit code and AI-DLC records on `feature/change-integration-t-4cn` | `ac485d0` |
| 2 | Rebase onto `origin/main` (`130093c`, v0.2.0) | Conflicts in README, the shared code KB (9 files) and `intents.json`; code files merged cleanly |
| 3 | Resolve conflicts | README: kept the v0.2.0 notes and added the `0.3.0` note "Backlog is now off by default (upgrading from v0.2.0 or earlier)". `intents.json`: kept both intents. Shared code KB: kept `main`'s v0.2.0 store (built by intent `261007-github-parity-actions`); the opt-in default is recorded in this intent's records and will be picked up by the next Reverse Engineering scan |
| 4 | Version bump | `manifest.yaml` `0.2.0` → `0.3.0` |
| 5 | Check v0.2.0 additions against the switch | v0.2.0 added only actions (quick actions, queries, git actions); `guarded()` in `internal/plugin/runtime.go` guards every action except `connection.get` and `connection.set_enabled`, so they are opt-in too. No new workers, RPCs or webhooks |
| 6 | Re-verify on the rebased code | See `health-check-report.md`: all green |
| 7 | Push | `feature/change-integration-t-4cn` → origin, commit `9d6d312` (with GitHub account `khuongdo`) |
| 8 | Open pull request | https://github.com/khuongdo/kandev-plugin-nulab-backlog/pull/8 — CI started (`ci.yml`) |

## Not Done (by decision)

- Merge, tag `v0.3.0`, GitHub Release with the `0.3.0` upgrade note (FR5.2), marketplace registry pull request: maintainer, after CI is green.
- Install on the self-hosted Kandev: not done (Q2 = C). Status: **pull request opened, not released, not installed**.

## Rollback

Nothing is released, so nothing to roll back. Before merge the pull request can simply be closed. After release, follow `operation/deployment-pipeline/rollback-runbook.md`.
