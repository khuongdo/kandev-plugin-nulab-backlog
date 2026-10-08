# Deployment Execution Questions - 261008-link-task-modal (release v0.5.2)

Pre-deployment facts (checked, not asked):

- **Build and Test passed every target:**
  - UI: 421 tests.
  - Go: 13 of 13 packages pass with `-race`, coverage 92.9%.
  - Package verified, and the contract test passed 10 of 10 runs on Kandev 0.96.0.
- **No database migrations.** The backend and stored data are unchanged.
- **Dependent services:** GitHub (CI, Releases) and the self-hosted Kandev.
- **This branch is one commit behind `origin/main`** (#19, v0.5.1), so it needs a rebase first. Likely overlaps: `manifest.yaml`, `README.md`, `ui/src/messages/en.ts`, and possibly the shared code knowledge base.
- **Your go-ahead is needed for every step that publishes outside this machine:** push, pull request, merge, tag, install, and the marketplace pull request.

## Question 1
How far should I carry out the release now?

A. All the way: bump to 0.5.2 and add the README upgrade note, commit, rebase onto `origin/main`, re-run checks, push, open the pull request, wait for CI, squash-merge, re-check releases, tag `v0.5.2`, wait for `release.yml`, put the upgrade note at the top of the Release notes, install From URL on the self-hosted Kandev and run the smoke checks, then open the marketplace registry pull request
B. Up to the GitHub Release: everything in A except the self-hosted install and the marketplace pull request (you install yourself)
C. Up to the pull request only: bump, commit, rebase, re-check, push and open the pull request; you merge and tag yourself
X. Other (please specify)

[Answer]: B

## Consolidated Summary Confirmation

- Q1 (B): carry the release up to the GitHub Release:
  - bump `manifest.yaml` to `0.5.2` and add the README `### 0.5.2` upgrade note;
  - commit, rebase onto `origin/main` (#19), and re-run the checks;
  - push, open the pull request as `khuongdo`, wait for CI, and squash-merge;
  - re-check releases, tag `v0.5.2`, wait for `release.yml`, and put the upgrade note at the top of the Release notes.
- Not done here: the self-hosted install, its smoke checks, and the marketplace pull request. You do those.

Does this all look correct before I generate the deployment execution artifacts?

- Looks correct
- Request changes

[Answer]: Looks correct
