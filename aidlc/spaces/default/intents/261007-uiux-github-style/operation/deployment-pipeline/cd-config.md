# CD Configuration — 261007-uiux-github-style (release v0.1.1)

The delivery pipeline already exists and does not change for this intent. This document records how this change flows through it and the two release inputs this intent adds (Q1, Q2).

## Pipeline (unchanged)

| Stage | Where | What runs | Gate |
|---|---|---|---|
| Pull request | `.github/workflows/ci.yml` on `pull_request` | `make check-format vet lint test coverage check-secrets build package verify-package`, `go mod tidy` check, packaged-host contract job on Kandev `min_kandev_version` | All required checks green; `main` is protected (no direct push) |
| Merge | GitHub | Self-merge with squash | CI green |
| Release | Deliberate tag `v0.1.1` on `main` → `.github/workflows/release.yml` | `verify` job: the full make chain plus `release-preflight TAG=v0.1.1` (tag format, tag equals `manifest.yaml` version, tag on `main`, no existing Release); `contract` job: `verify-package contract-test` on Kandev 0.96.0 against the exact package bytes; `publish` job: build provenance attestation and `gh release create` with the package and `checksums.txt` | Creating the tag is the manual production approval |
| Marketplace | Pull request to the Kandev plugin registry | Updated entry from `make marketplace-entry` | Kandev maintainers review |
| Install | Self-hosted Kandev | Manual install of the released package | Manual |

Every action is pinned to a full commit SHA; workflows default to `permissions: contents: read`; only `publish` gets `id-token: write` and `attestations: write`.

## Release Inputs Added by This Intent

1. **Version bump (Q1 = `0.1.1`)** — before tagging, a pull request changes `manifest.yaml` `version: "0.1.0"` → `"0.1.1"`. `release-preflight` refuses `v0.1.1` until `main` carries that version.
2. **Upgrade notes (Q2 = Release notes + README)** — the same pull request adds an "Upgrade notes" section to `README.md`, and the GitHub Release body for `v0.1.1` carries the same text:
   - The Integrations menu now has one Backlog entry; `/backlog` has Issues and Pull requests tabs.
   - `/backlog/watches` and `/backlog/dashboard` were removed; bookmarks to them no longer work.
   - PR watches, the new issue watches, and saved PR queries are in Settings > Integrations > Backlog; saved queries are created from the Pull requests tab.
   - The plugin icon is a new outline icon; the Nulab logo is no longer shipped.
   - Existing PR watches, saved queries and issue links are kept; no data migration is needed.

## Environment Promotion Matrix

| Environment | Artifact | Promotion trigger | Verification |
|---|---|---|---|
| CI (pull request) | Package built from the PR head | Every push to the PR | CI checks + contract test on Kandev 0.96.0 |
| GitHub Release | `nulab-backlog-0.1.1.tar.gz` + `checksums.txt` + provenance | Tag `v0.1.1` on `main` | `release.yml` verify/contract/publish; `gh attestation verify` by users |
| Self-hosted Kandev ("production") | The released package | Manual install by the maintainer | Smoke check in `rollback-runbook.md` |
| Marketplace | Registry entry for `0.1.1` | Registry pull request merged by Kandev maintainers | Maintainer review |

## Feature Flags

None. The team does not use feature flags; the integration's own enable switch per workspace is product behaviour, not a release flag.
