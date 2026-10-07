# Deployment Log — 261007-github-parity-actions (release v0.2.0)

Date: 2026-10-07 (UTC). Plan: [cd-config.md](../deployment-pipeline/cd-config.md), [deployment-strategy.md](../deployment-pipeline/deployment-strategy.md). Decisions: Q1 = A (all the way to the Release and the registry), Q2 = B (skip the install, record the release as "released, not installed").

## Pre-deployment Checks

| Check | Result |
|---|---|
| Build and Test targets | All Met ([test-results.md](../../construction/build-and-test/test-results.md)) |
| Database migrations | None needed: new state documents only, plus an optional field on saved PR queries |
| Dependent services | None beyond Backlog and Kandev, and neither changes |
| Deployment window | Not needed (manual install on a single self-hosted server) |

## Steps

| Step | Command / action | Result |
|---|---|---|
| Account | `gh auth switch -u khuongdo` | Active account `khuongdo` |
| Commit and push | Commit `bb977c9` on `feature/add-default-queries-87j` (77 paths: code, tests, docs, AI-DLC records); `git push -u origin feature/add-default-queries-87j` | Pushed. The commit message lost its blank line after the subject. This has no effect on `main`, because the squash commit uses the PR title. |
| Pull request | https://github.com/khuongdo/kandev-plugin-nulab-backlog/pull/6 | CI `checks` pass (1m45s), `packaged-host-contract` pass (1m3s) |
| Squash merge | `gh pr merge 6 --squash` | Merged as `3d7e0d6` on `main`, with `manifest.yaml` `version: "0.2.0"` |
| Tag | Annotated tag `v0.2.0` on `3d7e0d6`; `git push origin v0.2.0` | Pushed |
| Release workflow | `release.yml` run 37585429654 | `verify` pass (2m3s, including `release-preflight`), `contract` pass (1m8s), `publish` pass (17s) |
| GitHub Release | https://github.com/khuongdo/kandev-plugin-nulab-backlog/releases/tag/v0.2.0 | Assets `nulab-backlog-0.2.0.tar.gz` and `checksums.txt`. The notes were edited so the README "Upgrade notes → 0.2.0" text comes first (Q2 of Deployment Pipeline) |
| Release verification | Downloaded the assets; `sha256sum -c checksums.txt`; `gh attestation verify nulab-backlog-0.2.0.tar.gz --repo khuongdo/kandev-plugin-nulab-backlog` | Both pass |
| Marketplace registry | Checked the open registry PR kdlbs/kandev#4284: its `plugin-registry/plugins.yaml` entry (`id: nulab-backlog`, `repo: khuongdo/kandev-plugin-nulab-backlog`, `categories: [integrations]`) carries no version and equals `make marketplace-entry` output | No change needed. The catalogue follows the repository's releases, so `0.2.0` is picked up once #4284 is merged. The PR is still awaiting Kandev maintainer review. |
| Account restored | `gh auth switch -u khuongdo-nicosys` | Active account `khuongdo-nicosys` |
| Install on self-hosted Kandev | Not done (Q2 = B) | Released, not installed |

## Rollback

No tag was deleted or moved. If a later install fails, follow [rollback-runbook.md](../deployment-pipeline/rollback-runbook.md): reinstall `v0.1.1` and fix forward with `v0.2.1`.
