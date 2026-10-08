# Deployment Log - v0.5.2

Release scope chosen at Deployment Execution: **B - up to the GitHub Release**. The self-hosted install and the marketplace registry PR are left to you. All times are UTC, 2026-10-08.

| # | Step | Result |
|---|------|--------|
| 1 | Commit the code and the AI-DLC records on `feature/fix-link-task-ui-do5` | `29c2528` |
| 2 | Rebase onto `origin/main` `2182715` (#19, v0.5.1) | The code merged cleanly, including `ui/src/messages/en.ts` and `README.md`. Conflicts were in the 9 shared code knowledge base files and `intents/intents.json`. They were resolved by keeping both intents' facts and both registry rows. The scope block keeps fingerprint `992b7f48…`, with the #19 paths added as shallow. Result: `a123977`. |
| 3 | Bump `manifest.yaml` to `0.5.2`; add the README `### 0.5.2` upgrade note (text approved at Deployment Pipeline Q2) | `2ac7321` |
| 4 | Re-run checks on the rebased branch (Go 1.26.8 local, `../kandev` at `v0.96.0`) | `make check-sdk check-format vet lint test coverage check-secrets build package verify-package` all exit 0. 425 Vitest tests passed; coverage 92.9%; `verifypkg: OK dist/nulab-backlog-0.5.2.tar.gz`; `make contract-test` passed 10 out of 10 runs on Kandev v0.96.0. |
| 5 | Push; open PR [#21](https://github.com/khuongdo/kandev-plugin-nulab-backlog/pull/21) (account `khuongdo`) | done |
| 6 | `main` moved again: #20 (AI-DLC records for v0.5.1) made the PR conflicting | Rebased onto `8105779`. The only conflict was `intents.json`: `gh-cli-auth` is kept as complete and the `link-task-modal` row is kept. #20 changed records only, so step 4's checks still apply. Pushed with `--force-with-lease` (`716141c`). |
| 7 | PR CI | `changes` (32s), `checks` (2m18s), `packaged-host-contract` (57s) and `secret-scan` (28s) all pass |
| 8 | Squash-merge #21 | `main` at `abee1f4` "Link a Backlog issue from the task, GitHub-style Link Task dialog, version 0.5.2 (#21)" |
| 9 | Re-check before tagging (project rule) | The latest Release is still `v0.5.1`; there is no `v0.5.2` tag on the remote; `origin/main` `manifest.yaml` is version `0.5.2` |
| 10 | Tag `v0.5.2` (annotated) on `abee1f4` and push | done; `release.yml` run 37731187049 started 05:12:10Z |
| 11 | `release.yml` | Jobs `verify`, `contract` and `publish` all succeeded |
| 12 | GitHub Release [v0.5.2](https://github.com/khuongdo/kandev-plugin-nulab-backlog/releases/tag/v0.5.2) | Assets: `nulab-backlog-0.5.2.tar.gz` (23,514,655 bytes) and `checksums.txt`. Notes edited so the 0.5.2 upgrade note is on top and the generated PR list below (project rule). |

## Not Done Here (your steps)

- Install 0.5.2 on the self-hosted Kandev: Settings > Plugins > Install plugin > From URL with `https://github.com/khuongdo/kandev-plugin-nulab-backlog/releases/download/v0.5.2/nulab-backlog-0.5.2.tar.gz`. Then run the six smoke checks in `operation/deployment-pipeline/deployment-strategy.md`, including real keyboard Enter in the "Link to task" dialog (architecture review R-04).
- Open the marketplace registry PR for 0.5.2.
- Open the AI-DLC Operation records PR for this intent. The Deployment Execution records were written after #21.

## Database Migrations

None. The backend and stored data are unchanged.

## Rollback

Follow `operation/deployment-pipeline/rollback-runbook.md`: reinstall `v0.5.1`, mark `v0.5.2` as broken, and fix forward. Not needed so far.
