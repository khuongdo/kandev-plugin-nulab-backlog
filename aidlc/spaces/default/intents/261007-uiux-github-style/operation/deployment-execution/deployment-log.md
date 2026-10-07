# Deployment Log — 261007-uiux-github-style (release v0.1.1)

Date: 2026-10-07 (UTC). Plan: [cd-config.md](../deployment-pipeline/cd-config.md), [deployment-strategy.md](../deployment-pipeline/deployment-strategy.md). Decisions: Q1 = B (through to the Release and the registry PR), Q2 = A (act as `khuongdo`, switch back after), Q3 = A (maintainer installs and runs the smoke check later).

## Pre-deployment Checks

| Check | Result |
|---|---|
| Build and Test targets | All Met ([test-results.md](../../construction/build-and-test/test-results.md)) |
| Database migrations | None needed (new documents only; no schema change to existing ones) |
| Dependent services | None beyond Backlog and Kandev; no change to either |
| Deployment window | Not needed (manual install, single self-hosted server) |

## Steps

| Step | Command / action | Result |
|---|---|---|
| Version bump and upgrade notes | `manifest.yaml` `0.1.0` → `0.1.1`; README "Upgrade notes" section; `make check-format test package verify-package` | Pass; `verifypkg: OK dist/nulab-backlog-0.1.1.tar.gz (nulab-backlog@0.1.1)` |
| Account | `gh auth switch -u khuongdo` | Active account `khuongdo` |
| Commit and push | Commit `4c1d1bb` on `feature/refactor-uiux-2bi` (133 files: code, tests, docs, AI-DLC records); `git push -u origin feature/refactor-uiux-2bi` | Pushed |
| Pull request | https://github.com/khuongdo/kandev-plugin-nulab-backlog/pull/4 | CI `checks` pass (2m4s), `packaged-host-contract` pass (1m0s) |
| Squash merge | `gh pr merge 4 --squash` | Merged as `9a1c718` on `main` |
| Tag | Annotated tag `v0.1.1` on `9a1c718`; `git push origin v0.1.1` | Pushed |
| Release workflow | `release.yml` run 37573279932 | `verify` pass (1m34s, incl. `release-preflight`), `contract` pass (59s), `publish` pass (8s) |
| GitHub Release | https://github.com/khuongdo/kandev-plugin-nulab-backlog/releases/tag/v0.1.1 | Assets `nulab-backlog-0.1.1.tar.gz`, `checksums.txt`; notes edited to lead with the upgrade notes (Q2) |
| Release verification | Downloaded assets; `sha256sum -c checksums.txt`; `gh attestation verify nulab-backlog-0.1.1.tar.gz --repo khuongdo/kandev-plugin-nulab-backlog` | Both pass |
| Marketplace registry | Fork `khuongdo/kandev`, branch `add-nulab-backlog`, entry added to `plugin-registry/plugins.yaml` (checked with `make marketplace-entry REGISTRY=...`: not yet listed); PR https://github.com/kdlbs/kandev/pull/4284 | Open, awaiting Kandev maintainer review (first listing of `nulab-backlog`; the v0.1.0 registry PR had been postponed) |
| Account restored | `gh auth switch -u khuongdo-nicosys` | Active account `khuongdo-nicosys` |
| Install on self-hosted Kandev | Manual, by the maintainer (Q3 = A) | Smoke check skipped by maintainer decision on 2026-10-07 |

## Notes

- The registry PR follows the Kandev repository's PR template, which forbids tool attribution footers; none was added there.
- No tag was deleted or moved. Rollback path: [rollback-runbook.md](../deployment-pipeline/rollback-runbook.md).
