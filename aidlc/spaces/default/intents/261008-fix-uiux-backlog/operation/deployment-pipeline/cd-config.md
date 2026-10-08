# CD Configuration - v0.5.0 (Backlog issue on task rows and the task top bar)

No pipeline change. Express scope skipped CI Pipeline and Infrastructure Design; this release uses the existing workflows in the repository, as the team Deployment practice requires.

## Existing Pipeline (unchanged)

| Workflow | Trigger | What it does |
|----------|---------|--------------|
| `.github/workflows/ci.yml` | `pull_request` to `main` (on `origin/main` it runs the app checks only when app files change, #14) | `make check-format vet lint test coverage`, package + verify, packaged-host contract test on `min_kandev_version` |
| `.github/workflows/secrets.yml` (on `origin/main`, #14) | `pull_request` | credential scan |
| `.github/workflows/release.yml` | tag `vX.Y.Z` on `main` | refuses an existing tag, re-runs all checks and package verification, publishes the GitHub Release with the package and `checksums.txt`, build provenance attestation |

All actions are pinned to full commit SHAs; workflows default to `permissions: contents: read`.

## Release-Specific Settings

| Item | Value | Source |
|------|-------|--------|
| Version | `0.5.0` | [Q1] |
| Version file | `manifest.yaml` `version: "0.5.0"` (the only version reference; `min_kandev_version` stays `0.96.0`) | code scan |
| Tag | `v0.5.0` on `main`, after the PR is squash-merged | team Deployment |
| README | new `### 0.5.0: Backlog issue on task rows and the task top bar` at the top of `## Upgrade notes` (text in deployment-pipeline-questions.md, Q2) | [Q2] |
| GitHub Release notes | the same upgrade-note text at the top, generated PR list below (`gh release edit`) | project Deployment rule |
| Marketplace | PR to the Kandev marketplace registry after the Release | team Deployment |
| Install | manual on the self-hosted Kandev (Settings > Plugins > Install plugin, From URL with the Release package URL) | team Deployment |

## Promotion Gates

1. PR from `feature/fix-uiux-i41` to `main`: every required CI check green (protected `main`), then self-merge with squash.
2. Re-check `gh release list` and `origin/main` right before tagging (project rule); stop if `v0.5.0` exists or `main` has moved the version.
3. Tag `v0.5.0` on the squash commit; `release.yml` must pass, including package verification (project Mandated rule).
4. Manual install on the self-hosted Kandev and the smoke checks in deployment-strategy.md.

## Before the PR

- Rebase onto `origin/main` (three commits ahead: #14, #15, #16). Expected conflicts: the shared code knowledge base files under `aidlc/spaces/default/codekb/kandev-plugin-nulab-backlog/` (changed by #15/#16 and republished by this intent's Reverse Engineering); keep this intent's republished store, which already includes the upstream facts it scanned, and re-check `README.md` (changed by #14).
- Re-run `make check-format vet lint test coverage package verify-package` after the rebase.
