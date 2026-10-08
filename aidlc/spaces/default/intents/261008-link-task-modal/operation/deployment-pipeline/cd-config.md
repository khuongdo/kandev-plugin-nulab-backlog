# CD Configuration - v0.5.2 (link a Backlog issue from the task, GitHub-style)

No pipeline change. The bugfix scope skipped CI Pipeline and Infrastructure Design, so this release uses the existing workflows in the repository, as the team Deployment practice requires.

## Existing Pipeline (unchanged)

| Workflow | Trigger | What it does |
|----------|---------|--------------|
| `.github/workflows/ci.yml` | `pull_request` to `main`. App checks run only when app files change (#14); `ui/` and `README.md` changes count. | `make check-format vet lint test coverage`, package + verify, packaged-host contract test on `min_kandev_version` |
| `.github/workflows/secrets.yml` | `pull_request` | credential scan |
| `.github/workflows/release.yml` | tag `vX.Y.Z` on `main` | Refuses an existing tag, re-runs all checks and package verification, publishes the GitHub Release with the package and `checksums.txt`, and attaches build provenance |

All actions are pinned to full commit SHAs. Workflows default to `permissions: contents: read`.

## Release-Specific Settings

| Item | Value | Source |
|------|-------|--------|
| Version | `0.5.2` | [Q1] |
| Version file | `manifest.yaml`: `version: "0.5.1"` (on `origin/main`) becomes `"0.5.2"`. This is the only version reference; `min_kandev_version` stays `0.96.0`. | code scan |
| Tag | `v0.5.2` on `main`, after the PR is squash-merged | team Deployment |
| README | New `### 0.5.2: link a Backlog issue from the task, GitHub-style` at the top of `## Upgrade notes`, above `0.5.1`. The text is in deployment-pipeline-questions.md, Q2. | [Q2] |
| GitHub Release notes | The same upgrade-note text at the top, with the generated PR list below (`gh release edit`) | project Deployment rule |
| Marketplace | PR to the Kandev marketplace registry after the Release | team Deployment |
| Install | Manual, on the self-hosted Kandev: Settings > Plugins > Install plugin > From URL, with the Release package URL | team Deployment |

## Promotion Gates

1. PR from `feature/fix-link-task-ui-do5` to `main`: every required CI check is green (protected `main`), then self-merge with squash.
2. Right before tagging, re-check `gh release list` and `origin/main` (project rule). Stop if `v0.5.2` exists or `main` has moved the version.
3. Tag `v0.5.2` on the squash commit. `release.yml` must pass, including package verification (project Mandated rule).
4. Install manually on the self-hosted Kandev and run the smoke checks in deployment-strategy.md.

## Before the PR

- **Rebase onto `origin/main`**, which is one commit ahead with #19 (v0.5.1, gh/glab CLI login). Expected overlaps:
  - `manifest.yaml`: take `0.5.1`, then bump it to `0.5.2`.
  - `README.md`: keep `0.5.1`'s upgrade note and add `0.5.2` above it.
  - `ui/src/messages/en.ts`: keep both sets of new messages.
  - The shared code knowledge base under `aidlc/spaces/default/codekb/kandev-plugin-nulab-backlog/`, if #19 touched it.
- **Re-run checks after the rebase:** `make check-format vet lint test coverage package verify-package`, and `make contract-test KANDEV_MIN_DIR=../kandev` (10 runs).
- **Optional:** architecture review finding R-02 (`link-task-dialog.tsx`: keep the toast and `onLinked` out of the `issues.link` `try`). It was not required by the approved Code Generation gate.
