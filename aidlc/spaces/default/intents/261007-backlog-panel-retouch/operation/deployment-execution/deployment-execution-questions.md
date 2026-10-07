# Deployment Execution Questions — 261007-backlog-panel-retouch

Context: pre-deployment checks pass (Build and Test: every command green — Vitest 373/373, Go `-race`, coverage 92.8%, `verify-package` OK for `nulab-backlog-0.4.1.tar.gz`, contract test 10/10 on Kandev 0.96.0; NFR2-UI-IN-HOST accepted as Unverified, to be closed by the manual real-host UI check). No data migration, no dependent service, no deployment window. The changes are uncommitted on branch `feature/refactor-backlog-pan-03d` of `khuongdo/kandev-plugin-nulab-backlog`, based on `v0.4.0` (`5724d88`); `origin/main` is at `ad4adcf` (records only), so a rebase comes first. `manifest.yaml` and the README already say `0.4.1`. Release steps from `cd-config.md`: rebase → commit (code + AI-DLC records) → push → pull request → CI green → manual real-host UI check with the PR package → squash merge → tag `v0.4.1` → `release.yml` → (marketplace registry pull request) → manual install → smoke check. Pushing, merging, tagging and publishing are outward-facing and hard to undo, so they need your go-ahead. GitHub account: `khuongdo` (the repository owner). Answer each by writing the letter after the answer tag.

## Question 1
How far should I take the release now?

A. Prepare and open the pull request: rebase onto `origin/main`, commit code and AI-DLC records, push and open a pull request to `main`; stop there — you run the UI check, merge and tag
B. Option A, then wait for CI; after you confirm the manual UI check passed, squash-merge, tag `v0.4.1` and watch `release.yml` until the GitHub Release exists (no marketplace registry pull request, as in v0.4.0)
C. Only commit locally; you push and do the rest
X. Other (please specify)

[Answer]:

## Question 2
The manual real-host UI check (checklist in `construction/build-and-test/integration-test-instructions.md`) must pass before the tag. Who runs it, and when?

A. You install the PR package (`dist/nulab-backlog-0.4.1.tar.gz`, built locally) on the self-hosted Kandev and run it now; you tell me pass/fail per item and I record it
B. You run it later; I record it as pending and do not tag until you report a pass
C. Skip it for this release and record NFR2-UI-IN-HOST as an accepted risk
X. Other (please specify)

[Answer]: B

## Consolidated Summary Confirmation

- Release extent (Q1 = B): rebase onto `origin/main` (`ad4adcf`), commit code and AI-DLC records on `feature/refactor-backlog-pan-03d`, push, open a pull request to `main` and wait for CI. Merge and tag only after you report that the manual UI check passed; then squash-merge, tag `v0.4.1`, and watch `release.yml` until the GitHub Release exists. No marketplace registry pull request.
- Manual UI check (Q2 = B): you run the checklist later with `dist/nulab-backlog-0.4.1.tar.gz` (or the CI artifact of the pull request); I record it as pending and do not merge or tag until you report a pass.
- Before tagging I re-check `gh release list` and `origin/main` again (project rule).
- Post-install smoke check after the release: you install `v0.4.1` on the self-hosted Kandev and report the result; recorded as pending until then.

Does this all look correct before I generate the deployment execution artifacts?

Looks correct
Request changes

[Answer]: Looks correct
