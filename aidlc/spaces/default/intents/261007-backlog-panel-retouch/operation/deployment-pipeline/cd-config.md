# CD Configuration — 261007-backlog-panel-retouch (release v0.4.1)

The delivery pipeline already exists and does not change for this intent. This document records how this UI refactor flows through it and the release inputs it adds (Q2, Q3, Q4 in `deployment-pipeline-questions.md`).

## Pipeline (unchanged)

| Stage | Where | What runs | Gate |
|---|---|---|---|
| Pull request | `.github/workflows/ci.yml` on `pull_request` | `make check-format vet lint test coverage check-secrets build package verify-package` (includes UI Prettier, `tsc`, ESLint, Vitest), `go mod tidy` check, packaged-host contract job on Kandev `min_kandev_version` | All required checks green; `main` is protected (no direct push) |
| Merge | GitHub | Self-merge with squash | CI green |
| Release | Deliberate tag `v0.4.1` on `main` → `.github/workflows/release.yml` | `verify` job: full make chain plus `release-preflight TAG=v0.4.1` (tag format, tag equals `manifest.yaml` version, tag on `main`, no existing Release); `contract` job: `verify-package contract-test` on Kandev 0.96.0 against the exact package bytes; `publish` job: build provenance attestation and `gh release create` with the package and `checksums.txt` | Creating the tag is the manual production approval |
| Marketplace | Pull request to the Kandev plugin registry | Updated entry from `make marketplace-entry` | Kandev maintainers review |
| Install | Self-hosted Kandev | Manual install of the released package | Manual |

Every action is pinned to a full commit SHA; workflows default to `permissions: contents: read`; only `publish` gets `id-token: write` and `attestations: write`.

## Release Inputs Added by This Intent

1. **Base.** The branch was rebased onto `v0.4.0` (`5724d88`, PR #9) during Loop-back 2. `origin/main` has since moved to `ad4adcf` (PR #10, AI-DLC records for v0.4.0 only, no code); rebase onto it before opening the pull request (expected conflict: `aidlc/spaces/default/intents/intents.json` only) and let CI re-run. Re-check `gh release list` and `origin/main` immediately before tagging (project rule).
2. **Version (Q3 = `0.4.1`)** — `manifest.yaml` already says `0.4.1` in this change set. `release-preflight` refuses `v0.4.1` until `main` carries that version.
3. **Release notes (Q2 = README + Release)** — README `## Upgrade notes` already has "### 0.4.1: GitHub-style lists"; use the same text as the GitHub Release body for `v0.4.1`, adding one line for code review R-07:
   - Issue search now runs when you press Enter (or leave the search box), not automatically while typing.
   - The issue list and every pull-request list (Backlog Git and the connected GitHub, GitLab and Bitbucket providers) use Kandev's GitHub-style toolbar: title and count, searchable dropdown filters, last updated and refresh. On phones the filters stack at full width; the "Filters (n)" button is gone.
   - Linked Kandev tasks show their title; several tasks collapse into a "Tasks (n)" menu; clicking a task opens it.
   - The Backlog badge on Kanban cards opens the Backlog issue in a new tab (the status detail moves to its tooltip).
   - Pull-request status is one "Status (n)" filter; at least one status always stays selected (now also on provider lists).
   - No data, setting or action changes; no reconnect needed.
4. **Manual real-host UI check before tagging** — NFR2-UI-IN-HOST was accepted as Unverified at Build and Test; run the checklist in `construction/build-and-test/integration-test-instructions.md` (issue list, Backlog and provider pull-request lists, Kanban badge, phone width, review findings R-01/R-04/R-05/R-06/R-07) on the self-hosted Kandev with the pull-request package before creating the tag (see `deployment-strategy.md`).

## Environment Promotion Matrix

| Environment | Artifact | Promotion trigger | Verification |
|---|---|---|---|
| CI (pull request) | Package built from the PR head | Every push to the PR | CI checks + contract test on Kandev 0.96.0 |
| Self-hosted Kandev (pre-release check) | Package from the PR (`make package`, `nulab-backlog-0.4.1.tar.gz`) | Manual install by the maintainer before tagging | Manual real-host UI checklist |
| GitHub Release | `nulab-backlog-0.4.1.tar.gz` + `checksums.txt` + provenance | Tag `v0.4.1` on `main` | `release.yml` verify/contract/publish; `gh attestation verify` by users |
| Self-hosted Kandev ("production") | The released package | Manual install by the maintainer | Smoke check in `deployment-strategy.md` |
| Marketplace | Registry entry for `0.4.1` | Registry pull request merged by Kandev maintainers | Maintainer review |

## Feature Flags

None. The team does not use feature flags; the UI change ships as a whole in `0.4.1`.
