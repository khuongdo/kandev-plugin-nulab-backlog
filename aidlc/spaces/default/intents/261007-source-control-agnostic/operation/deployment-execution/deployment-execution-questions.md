# Deployment Execution Questions — 261007-source-control-agnostic

Context: pre-deployment checks pass (Build and Test: every target Met; contract test 10/10 on Kandev 0.96.0). No data migration, no dependent service, no deployment window. The changes are uncommitted on branch `feature/source-control-agnos-2jr` of `khuongdo/kandev-plugin-nulab-backlog`, which already contains `origin/main` (`v0.3.0`). Release steps from `cd-config.md`: version bump to `0.4.0` + README `0.4.0` upgrade note → commit → push → pull request → CI green → squash merge → tag `v0.4.0` → `release.yml` → marketplace registry pull request → manual install → smoke check. Pushing, merging, tagging and publishing are outward-facing and hard to undo, so they need your go-ahead. GitHub account: `khuongdo` (the repository owner). Answer each by writing the letter after the answer tag.

## Q1. How far should I take the release now?

A. Prepare and open the pull request: bump `manifest.yaml` to `0.4.0`, add the README `0.4.0` upgrade note, re-run `make test coverage verify-package`, commit code and AI-DLC records, push and open a pull request to `main`; stop there — you merge and tag
B. Option A, then wait for CI, squash-merge, tag `v0.4.0`, watch `release.yml` until the GitHub Release exists, and open the marketplace registry pull request
C. Only commit locally; you push and do the rest
X. Other (please specify)

[Answer]: X. Other: release va khong mở PR marketplace (option B without the marketplace registry pull request)

## Q2. Who installs the release on the self-hosted Kandev and runs the smoke check?

A. You install and run the smoke check later and tell me the result; I record it as pending
B. You install and run it now, before I close this stage
C. Skip the install for now; record the deployment as "not installed"
X. Other (please specify)

[Answer]: A
