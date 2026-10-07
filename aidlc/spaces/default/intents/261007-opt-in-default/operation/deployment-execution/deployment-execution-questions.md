# Deployment Execution Questions — 261007-opt-in-default

Context: pre-deployment checks pass on this branch (Build and Test: every target Met, contract test OK on Kandev 0.96.0). No database migrations and no dependent services are involved; there is no deployment window to agree. The changes are uncommitted on branch `feature/change-integration-t-4cn` of `khuongdo/kandev-plugin-nulab-backlog`, which was cut from `v0.1.1`; `v0.2.0` is already on `main`. Release steps from `cd-config.md`: rebase onto `origin/main` and re-verify → version bump to `0.3.0` + upgrade notes → pull request → CI green → squash merge → tag `v0.3.0` → `release.yml` → marketplace registry pull request → manual install on the self-hosted Kandev → smoke check. Pushing, merging, tagging and publishing are outward-facing and hard to undo, so they need your go-ahead. The GitHub account is `khuongdo` (the repository owner, already active). Answer each by writing the letter after the answer tag.

## Q1. How far should I take the release now?

A. Prepare and open the pull request: commit, rebase onto `origin/main` (resolve conflicts, check v0.2.0's new actions/workers go through the switch, re-run `make test coverage contract-test`), bump `manifest.yaml` to `0.3.0`, update the README upgrade note, push and open a pull request to `main`; stop there — you merge and tag
B. Option A, then wait for CI, squash-merge, tag `v0.3.0`, watch `release.yml` until the GitHub Release exists, and open the marketplace registry pull request
C. Only commit and rebase locally; you push and do the rest
X. Other (please specify)

[Answer]: A

## Q2. Who installs the release on the self-hosted Kandev and runs the smoke check?

The smoke check (deployment-strategy.md) needs your self-hosted Kandev and a workspace that was connected on an older version.

A. You install and run the smoke check later, and tell me the result; I record it as pending until then
B. You install and run it now, before I close this stage
C. Skip the install for now; record the deployment as "released, not installed"
X. Other (please specify)

[Answer]: C

## Consolidated Summary Confirmation

- Scope (Q1 = A): commit code and AI-DLC records on `feature/change-integration-t-4cn`; rebase onto `origin/main` (v0.2.0), resolve conflicts (README, shared code KB), check that v0.2.0's new actions and workers go through the switch guard, and re-run `make test coverage contract-test`; bump `manifest.yaml` to `0.3.0`; change the README note to "Upgrading from v0.2.0 or earlier" under a `0.3.0` heading; push with the `khuongdo` account and open a pull request to `main`. Stop there: you merge, tag `v0.3.0`, and the release follows `release.yml`.
- Install (Q2 = C): no install now; record the deployment as "pull request opened, not released, not installed"; the smoke check stays not run.
- If the rebase re-verification fails, I stop and report instead of pushing.

Does this all look correct before I run the deployment?

- Looks correct
- Request changes

[Answer]: Looks correct
