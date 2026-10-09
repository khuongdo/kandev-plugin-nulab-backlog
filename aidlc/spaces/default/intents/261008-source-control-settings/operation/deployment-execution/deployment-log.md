# Deployment Log - 261008-source-control-settings (release v0.6.0)

Scope (Q1 = B): up to the GitHub Release. The self-hosted Kandev install and the marketplace step are done by the user.

## Timeline (UTC, 2026-10-08)

| Step | Result |
|---|---|
| Version | `manifest.yaml` `0.5.3` -> `0.6.0`; README `### 0.6.0: one source control service per workspace` with the Backlog Git worktree warning line (Deployment Pipeline Q2=A). `make test` and `make package verify-package` re-run green with 0.6.0 |
| Commit | `304018b` on `feature/refactor-source-cont-c9o` (code, tests, README, AI-DLC records); base `origin/main` `6d43d69`, no rebase needed |
| Push + pull request | as `khuongdo` (gh active account switched to `khuongdo` and left active): PR #26 https://github.com/khuongdo/kandev-plugin-nulab-backlog/pull/26 |
| CI | `changes` pass, `checks` pass (2m9s), `packaged-host-contract` pass (1m6s), `secret-scan` pass |
| Merge | squash-merged -> `main` at `10bcdab` (~23:44) |
| Pre-tag re-check | latest release still `v0.5.3`; no `v0.6.0` tag on the remote; `main` carries `version: "0.6.0"` |
| Tag | annotated `v0.6.0` on `10bcdab`, pushed ~23:45 |
| `release.yml` run 37861147005 | `verify` success, `contract` success, `publish` success |
| GitHub Release | https://github.com/khuongdo/kandev-plugin-nulab-backlog/releases/tag/v0.6.0 - `nulab-backlog-0.6.0.tar.gz` (23,586,356 bytes), `checksums.txt`; notes start with the README 0.6.0 upgrade note, generated PR list below |
| Release package check | downloaded assets: `sha256sum -c` OK, `verifypkg` OK (`nulab-backlog@0.6.0`), `gh attestation verify` exit 0 |

## Database Migrations

None. The SCM settings document gains an optional `active` field (schema version 1 unchanged); values are derived on first read.

## Not Done (by choice, Q1 = B)

- Install on the self-hosted Kandev server and the smoke checks in `operation/deployment-pipeline/deployment-strategy.md` (the user installs).
- Marketplace registry step. Previous releases found the registry entry repo-based (no version field), so no new PR is expected; the user handles it.

## Rollback

Per `operation/deployment-pipeline/rollback-runbook.md` (reinstall `v0.5.3`; never move or delete the tag; fix forward with 0.6.1).
