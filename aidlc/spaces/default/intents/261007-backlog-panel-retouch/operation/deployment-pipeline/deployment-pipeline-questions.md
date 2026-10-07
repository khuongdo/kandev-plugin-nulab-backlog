# Deployment Pipeline Questions — 261007-backlog-panel-retouch

Context: the release pipeline already exists and the team's Deployment practice settles strategy, gates, approvals and rollback (pull request to `main` with CI green, squash merge, deliberate `vX.Y.Z` tag on `main` → `release.yml` re-runs every check and package verification → GitHub Release with checksums and provenance → marketplace registry pull request; manual install on the self-hosted Kandev; rollback = reinstall the previous version and fix forward; no feature flags). Checked before asking: the latest GitHub release is `v0.3.0` (2026-10-07), `origin/main` is at `5bf88b9` (v0.3.0) and this branch is based on it; `manifest.yaml` says `0.3.0`. Answer by writing the letter after the answer tag.

## Question 1
Which version should this UI refactor be released as? It adds no action or data, but users will notice changed behaviour: issue search now runs on Enter instead of automatically, the filter panel and task display look different, and the Kanban badge now opens the Backlog issue.

A. `0.4.0` — minor: visible UI/behaviour changes for existing users
B. `0.3.1` — patch: no API or data change
X. Other (please specify)

[Answer]: A

## Question 2
What should the release notes say about the changed behaviour?

A. A short "What changed" list in the GitHub Release notes and the README changelog: search on Enter, GitHub-style filters and task display, Kanban badge opens the issue, PR list matched
B. GitHub Release notes only; no README change
X. Other (please specify)

[Answer]: A

## Earlier Summary (superseded: written before v0.4.0 was found released)

- Version (Q1): release as `0.4.0`; `manifest.yaml` moves from `0.3.0` to `0.4.0` and the release tag is `v0.4.0` (latest release is `v0.3.0`; branch is on `origin/main`).
- Release notes (Q2): a short "What changed in 0.4.0" list in the README (under `## Upgrade notes`) and the same text in the GitHub Release notes for `v0.4.0`: issue search runs on Enter; GitHub-style filter toolbar with searchable dropdowns; linked tasks shown with title and a "Tasks (n)" menu; Kanban badge opens the Backlog issue in a new tab; the PR list matches.
- The manual real-host UI check (NFR2-UI-IN-HOST, accepted as Unverified at Build and Test) runs during Deployment Execution on the self-hosted Kandev before the tag is created.
- Everything else follows the team's Deployment practice unchanged: pull request to `main` with CI green, squash merge, deliberate `v0.4.0` tag, `release.yml` re-runs all checks and package verification, GitHub Release with checksums and provenance, marketplace registry pull request, manual install on the self-hosted Kandev, rollback by reinstalling `v0.3.0` and fixing forward; no feature flags.

Does this all look correct before I generate the deployment pipeline artifacts?

Looks correct
Request changes

[Answer]: Looks correct

## Requested Changes Feedback

What should change?

[Answer]: v0.4.0 đã đc deploy rồi kiểm tra lại

## Revision (after the gate's Request Changes)

Re-checked: `v0.4.0` was released at 2026-10-07T13:28Z from PR #9 "GitHub, GitLab and Bitbucket pull requests" (`origin/main` = `5724d88`, `manifest.yaml` 0.4.0). This branch was rebased onto it; conflicts in `ui/src/git/pr-list.tsx`, `intents.json` and the shared code KB were resolved (code KB taken from `main`). PR #9 also added a GitHub/GitLab/Bitbucket PR list (`ui/src/git/scm-pr-list.tsx`) that still uses the old task links and toolbar.

## Question 3
Which version should this UI refactor be released as, now that `v0.4.0` is taken?

A. `0.5.0` — minor
B. `0.4.1` — patch
X. Other (please specify)

[Answer]: B (0.4.1)

## Question 4
Should the new GitHub/GitLab/Bitbucket PR list get the same GitHub-style toolbar and linked-task display?

A. Not now; a later intent
B. Yes, now (loop back to Code Generation)
X. Other (please specify)

[Answer]: B

## Consolidated Summary Confirmation

- Version (Q3): release as `0.4.1`; `manifest.yaml` already says `0.4.1` (Code Generation Step 33) and the release tag is `v0.4.1`. Latest release is `v0.4.0`; `origin/main` is now `ad4adcf` (PR #10, AI-DLC records only, no code), so the branch is rebased onto it before the pull request.
- Release notes (Q2 = README + Release): README already has "### 0.4.1: GitHub-style lists"; the same text is the GitHub Release body for `v0.4.1`, and it should also mention that the provider pull-request list keeps at least one status selected (code review R-07).
- Scope (Q4): the provider (GitHub/GitLab/Bitbucket) pull-request list now uses the same GitHub-style toolbar and task display; it is part of 0.4.1.
- The manual real-host UI check (NFR2-UI-IN-HOST, accepted as Unverified twice at Build and Test) runs on the self-hosted Kandev with the pull-request package before the `v0.4.1` tag, including the provider lists and review findings R-06/R-07.
- Everything else follows the team's Deployment practice unchanged: pull request to `main` with CI green, squash merge, deliberate `v0.4.1` tag, `release.yml` re-runs all checks and package verification, GitHub Release with checksums and provenance, marketplace registry pull request, manual install on the self-hosted Kandev, rollback by reinstalling `v0.4.0` and fixing forward; no feature flags.

Does this all look correct before I generate the deployment pipeline artifacts?

Looks correct
Request changes

[Answer]: Looks correct
