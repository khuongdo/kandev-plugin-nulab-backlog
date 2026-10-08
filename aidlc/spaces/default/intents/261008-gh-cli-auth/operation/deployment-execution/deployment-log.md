# Deployment Log — release v0.5.1

Scope chosen (Q1 = B): up to the GitHub Release. The self-hosted Kandev install and the marketplace registry pull request are left to the owner.

## Pre-deployment checks (2026-10-08)

- Build and Test: every target Met (see `construction/build-and-test/test-results.md`).
- Latest GitHub release `v0.5.0`; `origin/main` at `d3d17e5` = branch base, no rebase needed.
- No database migration (plugin-state field `source` is optional).

## Steps

| # | Step | Result |
|---|---|---|
| 1 | `manifest.yaml` → `version: "0.5.1"`; README `### 0.5.1` upgrade note (approved text) | done |
| 2 | `make check-format test package verify-package` after the bump | all exit 0; `verifypkg: OK dist/nulab-backlog-0.5.1.tar.gz (nulab-backlog@0.5.1)` |
| 3 | Commit on `feature/th-m-auth-method-cho-0pe`, push | done |
| 4 | Pull request [#19](https://github.com/khuongdo/kandev-plugin-nulab-backlog/pull/19) | opened |
| 5 | CI on #19: `changes`, `checks`, `packaged-host-contract`, `secret-scan` | all pass |
| 6 | Squash-merge #19 | `main` at `2182715` |
| 7 | Re-check releases before tagging (project rule) | latest still `v0.5.0`; `main` manifest `0.5.1` |
| 8 | On `main` (`2182715`): `make package verify-package` | `verifypkg: OK` |
| 9 | `git tag v0.5.1 && git push origin v0.5.1` | pushed |
| 10 | `release.yml` run for `v0.5.1`: jobs `verify`, `contract`, `publish` | all success |
| 11 | GitHub Release `v0.5.1` (not a pre-release) with `nulab-backlog-0.5.1.tar.gz`, `checksums.txt` | published |
| 12 | README upgrade note put at the top of the generated Release notes (`gh release edit`), PR list kept below (project rule) | done |
| 13 | `gh attestation verify nulab-backlog-0.5.1.tar.gz --repo khuongdo/kandev-plugin-nulab-backlog` | exit 0 (verified) |

## Not done (owner's steps)

- Install `v0.5.1` on the self-hosted Kandev (Settings > Plugins, From URL of the Release package) and run the smoke checks in `operation/deployment-pipeline/cd-config.md`.
- Marketplace registry pull request for 0.5.1.

## Rollback

See `operation/deployment-pipeline/rollback-runbook.md`.
