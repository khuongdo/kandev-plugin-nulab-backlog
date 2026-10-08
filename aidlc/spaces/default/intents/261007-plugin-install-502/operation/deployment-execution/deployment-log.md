# Deployment Log — 261007-plugin-install-502 (release v0.4.2)

Scope (Q1 = B): up to the GitHub Release. Installing on the self-hosted Kandev and the marketplace registry pull request are left to the maintainer.

## Timeline (UTC, 2026-10-07)

| Step | Result |
|---|---|
| Version bump | `manifest.yaml` `0.4.1` → `0.4.2`; README `## Upgrade notes` gained "0.4.2: smaller package, install From URL" |
| Commit | `c702331` on `feature/plugin-install-faile-9qm` (code, README, AI-DLC records) |
| Rebase onto `origin/main` (`bf20039`) | one conflict in `aidlc/spaces/default/intents/intents.json` (both sides edited the intent list); resolved by keeping `backlog-panel-retouch` as `complete` and adding `plugin-install-502`; rebased commit `f5a7529` |
| Push + pull request | PR #13 https://github.com/khuongdo/kandev-plugin-nulab-backlog/pull/13 |
| CI | `checks` pass (2m24s), `packaged-host-contract` pass (57s) |
| Merge | squash-merged → `main` at `3a983ab` |
| Pre-tag re-check | latest release still `v0.4.1`; no `v0.4.2` tag; `main` carries `version: "0.4.2"` |
| Tag | annotated `v0.4.2` on `3a983ab`, pushed ~23:49 |
| `release.yml` run 37704397347 | `verify` success, `contract` success, `publish` success |
| GitHub Release | https://github.com/khuongdo/kandev-plugin-nulab-backlog/releases/tag/v0.4.2 — `nulab-backlog-0.4.2.tar.gz` (23,337,444 bytes), `checksums.txt`; release notes set to the 0.4.2 upgrade notes |

## Database Migrations

None.

## Not Done (by choice, Q1 = B)

1. Install on the self-hosted Kandev (v0.97.0): **Settings > Plugins > Install plugin > From URL** with `https://github.com/khuongdo/kandev-plugin-nulab-backlog/releases/download/v0.4.2/nulab-backlog-0.4.2.tar.gz`, then run the smoke check in `operation/deployment-pipeline/deployment-strategy.md`. Note: `nulab-backlog` is currently not installed on that server (it was uninstalled at 2026-10-07T22:31Z before the failed upgrade attempts).
2. Marketplace registry pull request (`make marketplace-entry`).

## Rollback

Per `operation/deployment-pipeline/rollback-runbook.md` (reinstall `v0.4.1` From URL; never move or delete the tag).
