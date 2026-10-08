# Deployment Log — 261008-gh-cli-profile (release v0.5.3)

Scope (Q1 = A): all the way — PR, merge, tag, GitHub Release, Release notes, self-hosted install with smoke checks, marketplace registry.

## Timeline (UTC, 2026-10-08)

| Step | Result |
|---|---|
| Version | `manifest.yaml` `0.5.2` → `0.5.3` (set in Code Generation); README `### 0.5.3: choose the gh account per workspace` with the extra worktree / gh-version bullet (Q2 of Deployment Pipeline) |
| Commit | `72b6ade` on `feature/gh-cli-profile-scope-q1o` (code, tests, README, AI-DLC records); base `origin/main` `ca8146c`, no rebase needed |
| Push + pull request | as `khuongdo` (token read with `gh auth token --user khuongdo`, active gh account unchanged): PR #23 https://github.com/khuongdo/kandev-plugin-nulab-backlog/pull/23 |
| CI | `changes` pass, `checks` pass (1m40s), `packaged-host-contract` pass (51s), `secret-scan` pass |
| Merge | squash-merged → `main` at `81608d4` |
| Pre-tag re-check | latest release still `v0.5.2`; no `v0.5.3` tag on the remote; `main` carries `version: "0.5.3"` |
| Tag | annotated `v0.5.3` on `81608d4`, pushed ~07:09 |
| `release.yml` run 37741700665 | `verify` success, `contract` success, `publish` success |
| GitHub Release | https://github.com/khuongdo/kandev-plugin-nulab-backlog/releases/tag/v0.5.3 — `nulab-backlog-0.5.3.tar.gz` (23,544,506 bytes), `checksums.txt`; notes start with the README 0.5.3 upgrade note, generated PR list below |
| Self-hosted install | `POST http://localhost:38429/api/plugins/install` with the Release package URL → HTTP 201, `nulab-backlog` 0.5.3 `active` at 07:13:24 (replaced 0.5.2) |
| Marketplace registry | No new PR needed: the registry entry is repo-based (`id`, `repo`, `categories`, no version). The existing PR kdlbs/kandev#4284 (adds `nulab-backlog`) is still open and covers 0.5.3 once merged. |

## Database Migrations

None (the chosen gh login is the existing `AccountID`).

## Not Done

- Interactive smoke checks 2-7 of `operation/deployment-pipeline/deployment-strategy.md` (picker, Change account, `gh auth switch`, logged-out message, worktree note): every workspace on this Kandev server has the Backlog integration switched off, and turning it on or connecting GitHub in the user's workspaces was not done without asking. See `smoke-test-results.md`.

## Rollback

Per `operation/deployment-pipeline/rollback-runbook.md` (reinstall `v0.5.2` From URL; never move or delete the tag).
