# Deployment Execution Questions - 261008-gh-cli-profile (release v0.5.3)

Pre-deployment facts (checked, not asked):

- **Build and Test passed every target:** Go 1433/1433 with `-race`, coverage 92.9%; UI 435/435; package `0.5.3` verified; contract test 10/10 on Kandev 0.96.0.
- **No data migration.** The chosen gh login is the existing `AccountID`.
- **Dependent services:** GitHub (CI, Releases) and the self-hosted Kandev.
- **Branch base is `origin/main` `ca8146c`** (latest release `v0.5.2`); no rebase needed unless `main` moves.
- **Your go-ahead is needed for every step that publishes outside this machine:** push, pull request, merge, tag, install, and the marketplace pull request.

## Question 1
How far should I carry out the release now?

A. All the way: commit, push, open the pull request as `khuongdo`, wait for CI, squash-merge, re-check releases, tag `v0.5.3`, wait for `release.yml`, put the upgrade note at the top of the Release notes, install From URL on the self-hosted Kandev and run the smoke checks, then open the marketplace registry pull request
B. Up to the GitHub Release: everything in A except the self-hosted install and the marketplace pull request (you install yourself)
C. Up to the pull request only: commit, push and open the pull request; you merge and tag yourself
X. Other (please specify)

[Answer]: A
