# Deployment Execution Questions — 261007-plugin-install-502 (release v0.4.2)

Pre-deployment facts (checked, not asked): Build and Test passed every target (1325 Go + 373 UI tests, coverage 92.8%, package 23.4 MB, contract test 10/10 on Kandev 0.96.0). There are no database migrations. The dependent services are GitHub (CI, Releases) and the self-hosted Kandev v0.97.0 at `https://webfrontier.tail152aaa.ts.net`, which is running. Steps that publish outside this machine (push, pull request, merge, tag, install, marketplace pull request) need your go-ahead.

## Question 1
How far should I carry out the release now?

A. All the way: bump to 0.4.2 and add the release notes, commit, rebase onto `origin/main`, push, open the pull request, wait for CI, squash-merge, re-check releases, tag `v0.4.2`, wait for `release.yml`, install From URL on the self-hosted Kandev and run the smoke check, then open the marketplace registry pull request
B. Up to the GitHub Release: everything in A except the self-hosted install and the marketplace pull request (you install yourself)
C. Up to the pull request only: commit, rebase, push and open the pull request; you merge and tag yourself
X. Other (please specify)

[Answer]: B

## Consolidated Summary Confirmation

- Release scope (Q1): B — bump `manifest.yaml` to `0.4.2` and add the 0.4.2 README upgrade notes, commit, rebase onto `origin/main`, push, open the pull request, wait for CI, squash-merge, re-check `gh release list` and `origin/main`, tag `v0.4.2` on `main`, and wait for `release.yml` to publish the GitHub Release with the package, `checksums.txt` and provenance.
- Not done by me: installing 0.4.2 on the self-hosted Kandev (you install it From URL) and the marketplace registry pull request; both stay as next steps in the deployment log.
- No database migrations; no feature flags; rollback per `operation/deployment-pipeline/rollback-runbook.md`.

Does this all look correct before I run the release and write the deployment artifacts?

- Looks correct
- Request changes

[Answer]: Looks correct
