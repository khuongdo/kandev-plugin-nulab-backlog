# Health Check Report — 261007-uiux-github-style (release v0.1.1)

## Status

| Area | Status | Evidence |
|---|---|---|
| Source on `main` | Healthy | PR #4 squash-merged as `9a1c718`; CI `checks` and `packaged-host-contract` green |
| Release artifact | Healthy | GitHub Release `v0.1.1` with package and `checksums.txt`; checksum and provenance attestation verified |
| Release pipeline | Healthy | `release.yml` run 37573279932: verify, contract, publish all green |
| Upgrade notes | Done | README "Upgrade notes"; Release notes lead with the same text |
| Marketplace listing | Pending review | https://github.com/kdlbs/kandev/pull/4284 |
| Self-hosted install | Not verified | Smoke check skipped by maintainer decision (2026-10-07); see smoke-test-results.md |

## Runtime Health Signals to Watch After Install

The plugin has no metrics endpoint; health is read from Kandev and the plugin's own surfaces:

- Kandev log: the plugin starts once, with no repeated restarts or Host-wait timeouts.
- Issue watcher log line per cycle (`watches`, `created`, `skipped`, `errors` counts, no secrets): `errors` stays 0 on a healthy connection.
- Settings > Integrations > Backlog: no watch shows a `lastError` badge (`unauthorized`, `rate_limited`, `workflow_missing`, `ledger_full`).
- Expected task creation rate: at most one task per issue watch per its interval.

## Rollback Readiness

`v0.1.0` assets remain on its GitHub Release; the rollback steps are in [rollback-runbook.md](../deployment-pipeline/rollback-runbook.md). Fast containment without reinstall: pause issue watches or switch the integration off for the workspace.

## Open Items

1. Follow the marketplace registry PR #4284 until a Kandev maintainer merges it.
2. Commit this intent's Operation records in a follow-up pull request (as done for `v0.1.0` in PR #3).
