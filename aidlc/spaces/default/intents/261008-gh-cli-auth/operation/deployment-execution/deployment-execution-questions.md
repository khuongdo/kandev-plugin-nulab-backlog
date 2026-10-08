# Deployment Execution Questions - 261008-gh-cli-auth (release v0.5.1)

Pre-deployment facts (checked, not asked): Build and Test passed every target (Go -race all packages, 391 UI tests, coverage 92.9%, package verified, contract test 10/10 on Kandev 0.96.0). There are no database migrations (the new `source` field in plugin state is optional and backward compatible). Dependent services are GitHub (CI, Releases) and the self-hosted Kandev. Re-checked on 2026-10-08: latest GitHub release is `v0.5.0`; `origin/main` is at `d3d17e5`, the base of this branch, so no rebase is needed. The only version string to bump is `manifest.yaml` (the README heading for 0.5.0 and a test comment stay). Steps that publish outside this machine (push, pull request, merge, tag, install, marketplace pull request) need your go-ahead.

## Question 1
How far should I carry out the release now?

A. All the way: bump to 0.5.1 and add the README upgrade note, commit, push, open the pull request, wait for CI, squash-merge, re-check releases, tag `v0.5.1`, wait for `release.yml`, put the upgrade note at the top of the Release notes, install From URL on the self-hosted Kandev and run the smoke checks, then open the marketplace registry pull request
B. Up to the GitHub Release: everything in A except the self-hosted install and the marketplace pull request (you install yourself)
C. Up to the pull request only: bump, commit, push and open the pull request; you merge and tag yourself
X. Other (please specify)

[Answer]: B
