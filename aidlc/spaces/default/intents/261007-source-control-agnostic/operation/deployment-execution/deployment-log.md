# Deployment Log — 261007-source-control-agnostic (v0.4.0)

## Scope Executed

Per the answers (Q1 = Other: "release va khong mở PR marketplace" — full release without the marketplace registry pull request; Q2 = A: the maintainer installs and runs the smoke check later).

## Steps

| # | Step | Result |
|---|---|---|
| 1 | Pre-checks | Latest release `v0.3.0`; branch `feature/source-control-agnos-2jr` already contained `origin/main` (`5bf88b9`), no rebase needed |
| 2 | Version bump | `manifest.yaml` `0.3.0` → `0.4.0` |
| 3 | Upgrade note | README `### 0.4.0: GitHub, GitLab and Bitbucket pull requests` above `0.3.0` |
| 4 | Re-verify | `make check-format test coverage package verify-package`: all OK, coverage 92.8%, `verifypkg: OK dist/nulab-backlog-0.4.0.tar.gz (nulab-backlog@0.4.0)` |
| 5 | Commit | `471aa91` "GitHub, GitLab and Bitbucket pull requests, version 0.4.0" (code + AI-DLC records up to Deployment Pipeline) |
| 6 | Push + pull request | https://github.com/khuongdo/kandev-plugin-nulab-backlog/pull/9 (GitHub account `khuongdo`) |
| 7 | CI | `checks` pass (2m17s), `packaged-host-contract` pass (1m1s) |
| 8 | Merge | Squash merge → `main` `5724d88abaa0940c956db561a04b0e962bfce122` |
| 9 | Tag | Annotated tag `v0.4.0` on `5724d88` (same form as `v0.3.0`), pushed |
| 10 | Release workflow | `release.yml` run 37628246786: `verify` success, `contract` success, `publish` success |
| 11 | GitHub Release | https://github.com/khuongdo/kandev-plugin-nulab-backlog/releases/tag/v0.4.0 with `nulab-backlog-0.4.0.tar.gz` and `checksums.txt` (provenance attested by `publish`) |
| 12 | Release notes | Upgrade note prepended to the generated notes, same format as `v0.3.0` |

## Not Done (by decision)

- Marketplace registry pull request: not opened (Q1).
- Install on the self-hosted Kandev and smoke check: maintainer, later (Q2 = A). Status: **released, install pending**.
- The AI-DLC records for Deployment Execution and later stages are not in the release commit; they go in a follow-up records pull request, as for earlier releases.

## Rollback

Follow `operation/deployment-pipeline/rollback-runbook.md` (reinstall `v0.3.0`; never delete or move the `v0.4.0` tag).
