# Deployment Execution Questions - 261008-fix-uiux-backlog (release v0.5.0)

Pre-deployment facts (checked, not asked): Build and Test passed every target (709 Go + 387 UI tests, coverage 92.8%, package verified, contract test 10/10 on Kandev 0.96.0). There are no database migrations (the new link `summary` field is optional and backward compatible). Dependent services are GitHub (CI, Releases) and the self-hosted Kandev. This branch is three commits behind `origin/main` (#14, #15, #16); a rebase is needed first, with likely conflicts only in the shared code knowledge base files. Steps that publish outside this machine (push, pull request, merge, tag, install, marketplace pull request) need your go-ahead.

## Question 1
How far should I carry out the release now?

A. All the way: bump to 0.5.0 and add the README upgrade note, commit, rebase onto `origin/main`, push, open the pull request, wait for CI, squash-merge, re-check releases, tag `v0.5.0`, wait for `release.yml`, put the upgrade note at the top of the Release notes, install From URL on the self-hosted Kandev and run the smoke checks, then open the marketplace registry pull request
B. Up to the GitHub Release: everything in A except the self-hosted install and the marketplace pull request (you install yourself)
C. Up to the pull request only: bump, commit, rebase, push and open the pull request; you merge and tag yourself
X. Other (please specify)

[Answer]: B
