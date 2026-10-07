# CD Configuration — 261007-source-control-agnostic (release v0.4.0)

The delivery pipeline already exists and does not change for this intent: no workflow file, Makefile target or manifest capability is added. This document records how the feature flows through the pipeline and the release inputs it adds (Q1, Q2).

## Pipeline (unchanged)

| Stage | Where | What runs | Gate |
|---|---|---|---|
| Pull request | `.github/workflows/ci.yml` on `pull_request` | `make check-format vet lint test coverage check-secrets build package verify-package`, `go mod tidy` check, packaged-host contract job on Kandev `min_kandev_version` (0.96.0) | All required checks green; `main` is protected |
| Merge | GitHub | Self-merge with squash | CI green |
| Release | Deliberate tag `v0.4.0` on `main` → `.github/workflows/release.yml` | `verify` job: full make chain plus `release-preflight TAG=v0.4.0` (tag format, tag equals `manifest.yaml` version, tag on `main`, no existing Release); `contract` job on the exact package bytes; `publish` job: provenance attestation and `gh release create` with the package and `checksums.txt` | Creating the tag is the manual production approval |
| Marketplace | Pull request to the Kandev plugin registry | Updated entry from `make marketplace-entry` | Kandev maintainers review |
| Install | Self-hosted Kandev | Manual install of the released package | Manual |

Every action stays pinned to a full commit SHA; workflows default to `permissions: contents: read`; only `publish` gets `id-token: write` and `attestations: write`.

## Release Inputs Added by This Intent

1. **Base is current.** Checked before release planning: latest release `v0.3.0`; `origin/main` = `5bf88b9` (`version: "0.3.0"`); this branch already contains `origin/main`, so no rebase is needed. Re-check `gh release list` and `origin/main` right before opening the pull request; if `main` moved, rebase and re-run `make test coverage contract-test`.
2. **Version bump (Q1 = `0.4.0`)** — the same pull request changes `manifest.yaml` `version: "0.3.0"` → `"0.4.0"`. `release-preflight` refuses `v0.4.0` until `main` carries that version.
3. **Upgrade notes (Q2 = A)** — README "Upgrade notes" gets a `### 0.4.0` entry above `0.3.0`, and the GitHub Release body for `v0.4.0` carries the same text:
   - New: pull requests from GitHub, GitLab and Bitbucket (cloud only) next to Backlog Git — PR list, saved queries, PR watches, PR status, and linking to tasks / Backlog issues, including automatic links when a Backlog issue key is in the branch name or PR title.
   - Setup: an admin adds one read-only token per service and maps Backlog projects to repositories in Settings > Integrations > Backlog > Source control. Required token scopes are shown there.
   - Moved: the Backlog Git "Git access" form is now inside the "Source control" section.
   - Nothing to do after upgrading: existing Backlog Git links, watches, saved queries and credentials keep working unchanged.
4. **No new manifest capability, permission or config field.** New `scm.*` actions only; tokens use the existing `capabilities.secrets`. `min_kandev_version` stays `0.96.0`.

## Environment Promotion Matrix

| Environment | Artifact | Promotion trigger | Verification |
|---|---|---|---|
| CI (pull request) | Package built from the PR head | Every push to the PR | CI checks + contract test on Kandev 0.96.0 |
| GitHub Release | `nulab-backlog-0.4.0.tar.gz` + `checksums.txt` + provenance | Tag `v0.4.0` on `main` | `release.yml` verify/contract/publish; `gh attestation verify` by users |
| Self-hosted Kandev ("production") | The released package | Manual install by the maintainer | Smoke check in `deployment-strategy.md` |
| Marketplace | Registry entry for `0.4.0` | Registry pull request merged by Kandev maintainers | Maintainer review |

## Feature Flags

None. The team does not use feature flags. A provider is inactive until an admin adds its token, and all provider features follow the existing per-workspace Backlog switch; both are product behaviour, not release flags.
