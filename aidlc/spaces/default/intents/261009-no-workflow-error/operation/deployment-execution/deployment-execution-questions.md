# Deployment Execution Questions — 261009-no-workflow-error (release v0.6.1)

Pre-deployment facts: Build and Test passed (all targets Met, contract test 10/10, real-Kandev check passed); no database or stored-data migration; no dependent service other than the user's self-hosted Kandev; GitHub account `khuongdo` is active in `gh`.

## Question 1
How far should the release go in this step? (Pushing, merging and tagging are outward-facing and cannot be undone for the tag.)

A. Up to the GitHub Release: commit, push, open the PR, wait for CI, squash-merge, re-check releases/main, tag `v0.6.1`, wait for `release.yml`, put the README 0.6.1 note at the top of the Release notes and verify the released package; you install on the self-hosted Kandev (same as v0.6.0)
B. Only commit, push and open the PR; you merge and tag yourself
C. Everything in A, and I also handle the marketplace registry step
X. Other (please specify)

[Answer]: A

## Question 2
When should it run?

A. Now
B. Later (please specify)
X. Other (please specify)

[Answer]: A

## Consolidated Summary Confirmation

- Q1=A: release up to the GitHub Release (commit, push, PR, CI, squash-merge, pre-tag re-check, tag `v0.6.1`, `release.yml`, Release notes with the README 0.6.1 note on top, released-package verification); the user installs on the self-hosted Kandev and handles the marketplace step.
- Q2=A: run now.

Does this all look correct before I run the release and generate the deployment artifacts?

Looks correct / Request changes

[Answer]: Looks correct
