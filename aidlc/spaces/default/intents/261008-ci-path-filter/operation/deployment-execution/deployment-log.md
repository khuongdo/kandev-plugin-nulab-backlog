# Deployment Log — CI path filter

Executed 2026-10-08 per `operation/deployment-pipeline/cd-config.md` (Q1 = A: this intent's records included; Q2 = A: run all including merge). No plugin release; no tag.

| # | Time (UTC, approx.) | Step | Result |
|---|---------------------|------|--------|
| 1 | 01:40 | New branch `feature/ci-path-filter` from `origin/main` (`3a983ab`); the old task branch tip `f5a7529` had the same tree as `3a983ab` (pre-squash copy of #13), so no rebase was needed | OK |
| 2 | 01:41 | Commit `05e463d`: workflows, `internal/ci`, README, this intent's records, code knowledge base, and only this intent's row in `intents.json`; the `261007` records left uncommitted for their own pull request | OK |
| 3 | 01:42 | Push, open pull request [#14](https://github.com/khuongdo/kandev-plugin-nulab-backlog/pull/14) | OK |
| 4 | 01:42–01:46 | PR CI: `changes` pass, `checks` pass (2m7s), `packaged-host-contract` pass (1m0s), `secret-scan` pass (29s) | Green |
| 5 | 01:47 | Ruleset 24580280 updated with the `gh api` command from `code-summary.md`; read-back required checks: `checks`, `packaged-host-contract`, `secret-scan`; other rules unchanged (`deletion`, `non_fast_forward`, `pull_request`, `required_status_checks`) | OK |
| 6 | 01:47 | Squash-merge #14 → `main` at `dd89cc2` (merge state CLEAN) | OK |
| 7 | 01:47–01:51 | `main` push CI on `dd89cc2`: `ci` success (`changes`, `checks`, `packaged-host-contract` all success), `secrets` success | Green |
| 8 | 01:52 | Branch `records/plugin-install-502-operation` from `origin/main`; commit with the `261007-plugin-install-502` Deployment Execution records and its `complete` status in `intents.json`; pull request [#15](https://github.com/khuongdo/kandev-plugin-nulab-backlog/pull/15) | OK |
| 9 | 01:52–01:53 | PR #15 CI: `changes` pass, `checks` skipped, `packaged-host-contract` skipped, `secret-scan` pass; merge state CLEAN, mergeable without bypass | Green |
| 10 | 01:54 | Squash-merge #15 → `main` at `a23845f` | OK |
| 11 | 01:54–01:55 | `main` push CI on `a23845f`: `ci` success (`changes` success, `checks` skipped, `packaged-host-contract` skipped), `secrets` success | Green |

## Not Done (by decision)

- No version bump, tag, GitHub Release or marketplace pull request (Q1 = A in Deployment Pipeline).
- `release.yml` unchanged.

## Rollback

Not needed. Procedure in `operation/deployment-pipeline/rollback-runbook.md`.
