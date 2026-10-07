# Deployment Execution Questions — 261007-github-parity-actions

Context: the pre-deployment checks pass. In Build and Test every target was Met, and the contract test passed 10/10 on Kandev 0.96.0. This release needs no database migration, touches no dependent service, and has no deployment window to agree. The changes are uncommitted on branch `feature/add-default-queries-87j` of `khuongdo/kandev-plugin-nulab-backlog`. Release steps from `cd-config.md`: commit → pull request → CI green → squash merge → tag `v0.2.0` → `release.yml` → marketplace registry → manual install on the self-hosted Kandev → smoke check. Pushing, merging, tagging and publishing are outward-facing and hard to undo, so they need your go-ahead. Pushes and pull requests are made as `khuongdo`, the repository owner, and `gh` is switched back to `khuongdo-nicosys` afterwards, as in the v0.1.1 release. The registry pull request kdlbs/kandev#4284, which adds `nulab-backlog` at `0.1.1`, is still open.

## Q1. How far should I take the release now?

A. All the way: commit code and AI-DLC records, push, open a pull request to `main`, wait for CI, squash-merge, tag `v0.2.0`, watch `release.yml` until the GitHub Release exists with the upgrade notes, then update the open registry pull request #4284 to `0.2.0`
B. Prepare and open the pull request only; stop there, and you merge and tag
C. Only commit locally; you push and do the rest
X. Other (please specify)

[Answer]: A

## Q2. Who installs the release on the self-hosted Kandev and runs the smoke check?

The smoke check in `deployment-strategy.md` needs your self-hosted Kandev and a connected Backlog space.

A. You install and run the smoke check later and tell me the result; I record it as pending until then
B. Skip the install for now; record the deployment as "released, not installed"
X. Other (please specify)

[Answer]: B
