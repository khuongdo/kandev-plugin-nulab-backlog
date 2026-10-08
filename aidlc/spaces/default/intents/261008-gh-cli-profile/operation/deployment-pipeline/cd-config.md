# CD Configuration - v0.5.3 (choose the gh account per workspace)

No pipeline change. The express scope skipped CI Pipeline and Infrastructure Design, so this release uses the existing workflows in the repository, as the team Deployment practice requires.

## Existing Pipeline (unchanged)

| Workflow | Trigger | What it does |
|----------|---------|--------------|
| `.github/workflows/ci.yml` | `pull_request` to `main`; app checks run only when app files change | `make check-format vet lint test coverage`, package + verify, packaged-host contract test on `min_kandev_version` |
| `.github/workflows/secrets.yml` | `pull_request` | credential scan |
| `.github/workflows/release.yml` | tag `vX.Y.Z` on `main` | Refuses an existing tag, re-runs all checks and package verification, publishes the GitHub Release with the package and `checksums.txt`, attaches build provenance |

All actions are pinned to full commit SHAs. Workflows default to `permissions: contents: read`.

## Release-Specific Settings

| Item | Value | Source |
|------|-------|--------|
| Version | `0.5.3` | [Q1] |
| Version file | `manifest.yaml`: `0.5.2` (on `origin/main`) → `0.5.3` (already set on this branch). `min_kandev_version` stays `0.96.0`. | code |
| Tag | `v0.5.3` on `main`, after the PR is squash-merged | team Deployment |
| README | `### 0.5.3: choose the gh account per workspace` at the top of `## Upgrade notes`, above `0.5.2`, with the extra worktree / gh version bullet | [Q2] |
| GitHub Release notes | The same upgrade-note text at the top, generated PR list below (`gh release edit`) | project Deployment rule |
| Marketplace | PR to the Kandev marketplace registry after the Release | team Deployment |
| Install | Manual, on the self-hosted Kandev: Settings > Plugins > Install plugin > From URL, with the Release package URL | team Deployment |

## Promotion Gates

1. PR from `feature/gh-cli-profile-scope-q1o` to `main`: every required CI check green (protected `main`), then self-merge with squash.
2. Right before tagging, re-check `gh release list` and `origin/main` (project rule). Stop if `v0.5.3` exists or `main` has moved the version.
3. Tag `v0.5.3` on the squash commit. `release.yml` must pass, including package verification (project Mandated rule).
4. Install manually on the self-hosted Kandev and run the smoke checks in deployment-strategy.md.

## Before the PR

- The branch base is `origin/main` `ca8146c` (latest release `v0.5.2`); no rebase needed unless `main` moves.
- Local checks already green in Build and Test (`make check-format vet lint check-secrets coverage package verify-package`, Vitest/tsc/ESLint/Prettier, contract test 10/10). Re-run if anything changes before the PR.
