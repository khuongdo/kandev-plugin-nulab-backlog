# Deployment Log - v0.5.0

Release scope chosen at Deployment Execution: **B - up to the GitHub Release** (the self-hosted install and the marketplace registry PR are left to you). All times UTC, 2026-10-08.

| # | Step | Result |
|---|------|--------|
| 1 | Bump `manifest.yaml` to `0.5.0`; add the README `### 0.5.0` upgrade note (text approved at Deployment Pipeline Q2) | done |
| 2 | Commit `cf71b4e` on `feature/fix-uiux-i41` (code + AI-DLC records) | done |
| 3 | Rebase onto `origin/main` `f3ca715` (#14, #15, #16) | conflicts in the 9 shared code knowledge base files and `intents/intents.json`; resolved by keeping both intents' facts (ci-path-filter run and this run) and both registry rows; scope block keeps fingerprint `bf70f7c…` so the next intent rescans. Result `fec15ec` |
| 4 | Re-run checks on the rebased branch | `make check-sdk check-format vet lint test coverage check-secrets build package verify-package` all exit 0; 387 Vitest tests; coverage 92.8%; `verifypkg: OK dist/nulab-backlog-0.5.0.tar.gz`; `make contract-test` OK on Kandev v0.96.0 |
| 5 | Push `feature/fix-uiux-i41`; open PR [#17](https://github.com/khuongdo/kandev-plugin-nulab-backlog/pull/17) (account `khuongdo`) | done |
| 6 | PR CI | `changes`, `checks` (2m5s), `packaged-host-contract` (54s), `secret-scan` all pass |
| 7 | Squash-merge #17 | `main` at `7682d69` "Backlog issue on task rows and the task top bar, settings fixes, version 0.5.0 (#17)" |
| 8 | Re-check before tagging (project rule) | latest Release still `v0.4.2`; no `v0.5.0` tag on the remote; `origin/main` `manifest.yaml` version `0.5.0` |
| 9 | Tag `v0.5.0` (annotated) on `7682d69`, push | done; `release.yml` run 37719084025 started 02:41:14Z |
| 10 | `release.yml` | jobs `verify`, `contract`, `publish` all success |
| 11 | GitHub Release [v0.5.0](https://github.com/khuongdo/kandev-plugin-nulab-backlog/releases/tag/v0.5.0) | assets `nulab-backlog-0.5.0.tar.gz` (23,337,685 bytes) and `checksums.txt`; notes edited so the 0.5.0 upgrade note is on top and the generated PR list below (project rule) |

## Not Done Here (your steps)

- Install 0.5.0 on the self-hosted Kandev: Settings > Plugins > Install plugin > From URL with `https://github.com/khuongdo/kandev-plugin-nulab-backlog/releases/download/v0.5.0/nulab-backlog-0.5.0.tar.gz`, then the five smoke checks in `operation/deployment-pipeline/deployment-strategy.md`.
- Marketplace registry PR for 0.5.0.
- AI-DLC Operation records PR for this intent (Deployment Execution and Observability Setup records written after #17).

## Database Migrations

None. Stored links gain an optional `summary` field; older versions ignore it.

## Rollback

Per `operation/deployment-pipeline/rollback-runbook.md` (reinstall `v0.4.2`, mark `v0.5.0`, fix forward). Not needed.
