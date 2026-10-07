# CD Configuration — 261007-opt-in-default (release v0.3.0)

The delivery pipeline already exists and does not change for this intent. This document records how this fix flows through it and the release inputs it adds (Q1, FR5.1/FR5.2).

## Pipeline (unchanged)

| Stage | Where | What runs | Gate |
|---|---|---|---|
| Pull request | `.github/workflows/ci.yml` on `pull_request` | `make check-format vet lint test coverage check-secrets build package verify-package`, `go mod tidy` check, packaged-host contract job on Kandev `min_kandev_version` | All required checks green; `main` is protected (no direct push) |
| Merge | GitHub | Self-merge with squash | CI green |
| Release | Deliberate tag `v0.3.0` on `main` → `.github/workflows/release.yml` | `verify` job: full make chain plus `release-preflight TAG=v0.3.0` (tag format, tag equals `manifest.yaml` version, tag on `main`, no existing Release); `contract` job: `verify-package contract-test` on Kandev 0.96.0 against the exact package bytes; `publish` job: build provenance attestation and `gh release create` with the package and `checksums.txt` | Creating the tag is the manual production approval |
| Marketplace | Pull request to the Kandev plugin registry | Updated entry from `make marketplace-entry` | Kandev maintainers review |
| Install | Self-hosted Kandev | Manual install of the released package | Manual |

Every action is pinned to a full commit SHA; workflows default to `permissions: contents: read`; only `publish` gets `id-token: write` and `attestations: write`.

## Release Inputs Added by This Intent

1. **Rebase onto `origin/main` (v0.2.0) first.** This branch was cut from `v0.1.1` (`2b4325f`); `v0.2.0` (quick actions, default queries, GitHub-style layout) landed on `main` meanwhile. Rebase, resolve conflicts (README, shared code KB), and re-verify: every new v0.2.0 action and worker must go through `RequireEnabled` so the opt-in default covers it; run `make test coverage contract-test` again.
2. **Version bump (Q1 = `0.3.0`)** — the same pull request changes `manifest.yaml` `version: "0.2.0"` → `"0.3.0"`. `release-preflight` refuses `v0.3.0` until `main` carries that version.
3. **Upgrade notes (FR5.1/FR5.2)** — the README note becomes "Upgrading from v0.2.0 or earlier: Backlog is now off by default" under a `0.3.0` heading, and the GitHub Release body for `v0.3.0` carries the same text:
   - Backlog is now opt-in: after installation it is off in every workspace until an admin turns it on.
   - A workspace that never touched the switch on v0.1.0–v0.2.0 turns **off** after the upgrade; its saved connection is kept, but issue sync, issue and PR watches and Git credentials pause.
   - To resume, a Kandev admin turns the switch on (Settings > Integrations > Backlog) and saves; no reconnect is needed. Workspaces where the switch was saved (on or off) keep their setting.

## Environment Promotion Matrix

| Environment | Artifact | Promotion trigger | Verification |
|---|---|---|---|
| CI (pull request) | Package built from the PR head | Every push to the PR | CI checks + contract test on Kandev 0.96.0 |
| GitHub Release | `nulab-backlog-0.3.0.tar.gz` + `checksums.txt` + provenance | Tag `v0.3.0` on `main` | `release.yml` verify/contract/publish; `gh attestation verify` by users |
| Self-hosted Kandev ("production") | The released package | Manual install by the maintainer | Smoke check in `deployment-strategy.md` |
| Marketplace | Registry entry for `0.3.0` | Registry pull request merged by Kandev maintainers | Maintainer review |

## Feature Flags

None. The team does not use feature flags; the per-workspace integration switch is product behaviour (and is exactly what this fix changes the default of), not a release flag.
