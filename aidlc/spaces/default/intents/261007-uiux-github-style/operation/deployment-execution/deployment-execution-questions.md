# Deployment Execution Questions — 261007-uiux-github-style

Context: pre-deployment checks pass (Build and Test: every target Met, contract test 10/10 on Kandev 0.96.0). No database migrations and no dependent services are involved; there is no deployment window to agree. The changes are uncommitted on branch `feature/refactor-uiux-2bi` of `khuongdo/kandev-plugin-nulab-backlog`. Release steps from `cd-config.md`: version bump to `0.1.1` + upgrade notes → pull request → CI green → squash merge → tag `v0.1.1` → `release.yml` → marketplace registry pull request → manual install on the self-hosted Kandev → smoke check. Pushing, merging, tagging and publishing are outward-facing and hard to undo, so they need your go-ahead. Answer each by writing the letter after the answer tag.

## Q1. How far should I take the release now?

A. Prepare and open the pull request: bump `manifest.yaml` to `0.1.1`, add the README upgrade notes, commit everything (code and AI-DLC records), push the branch and open a pull request to `main`; stop there — you merge and tag
B. Option A, then wait for CI, squash-merge, tag `v0.1.1`, watch `release.yml` until the GitHub Release exists, and open the marketplace registry pull request
C. Only commit locally; you push and do the rest
X. Other (please specify)

[Answer]: B

## Q2. Which GitHub account should push and open pull requests?

The local `gh` has two accounts: `khuongdo-nicosys` (active) and `khuongdo` (the repository owner).

A. Switch to `khuongdo` for this release, then switch back afterwards
B. Keep the active account `khuongdo-nicosys`
X. Other (please specify)

[Answer]: A

## Q3. Who installs the release on the self-hosted Kandev and runs the smoke check?

The smoke check (deployment-strategy.md) needs your self-hosted Kandev and, for step 4, a connected Backlog space.

A. You install and run the smoke check later, and tell me the result; I record it as pending until then
B. You install and run it now, before I close this stage
C. Skip the install for now; record the deployment as "released, not installed"
X. Other (please specify)

[Answer]: A

## Consolidated Summary Confirmation

- Scope (Q1 = B): bump `manifest.yaml` to `0.1.1`, add README upgrade notes, commit code and AI-DLC records on `feature/refactor-uiux-2bi`, push, open a pull request to `main`, wait for CI green, squash-merge, tag `v0.1.1` on `main`, watch `release.yml` until the GitHub Release (package, checksums, provenance) exists with the upgrade notes, then open the marketplace registry pull request.
- Account (Q2 = A): switch `gh` to `khuongdo` for these steps and switch back to `khuongdo-nicosys` afterwards.
- Install and smoke check (Q3 = A): you install `v0.1.1` on the self-hosted Kandev and run the smoke check later; I record it as pending.
- If CI or `release.yml` fails, I stop and report instead of retrying blindly; the tag is never deleted or moved.

Does this all look correct before I run the deployment?

Looks correct
Request changes

[Answer]: Looks correct
