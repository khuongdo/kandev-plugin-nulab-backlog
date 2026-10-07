# Deployment Pipeline Questions — 261007-source-control-agnostic

Context: the release pipeline already exists and the team's Deployment practice settles strategy, gates, approvals and rollback (tag `vX.Y.Z` on `main` → `release.yml` re-runs every check and package verification → GitHub Release with checksums and provenance → marketplace registry pull request; manual install on the self-hosted Kandev; rollback = reinstall the previous version and fix forward; no feature flags). Checked before asking (project rule): the latest GitHub release is `v0.3.0` (2026-10-07), `origin/main` is `5bf88b9` with `manifest.yaml` `version: "0.3.0"`, and this branch already contains `origin/main`, so no rebase is needed. Answer by writing the letter after the answer tag.

## Q1. Which version should this feature be released as?

The change adds a new feature (GitHub, GitLab and Bitbucket pull requests, new Settings section, 22 new actions) and keeps every existing action and stored data unchanged. One visible change for existing users: the Backlog Git "Git access" form moved into the new "Source control" Settings section.

A. `0.4.0` — minor: new, backward-compatible feature
B. `0.3.1` — patch
X. Other (please specify)

[Answer]: A

## Q2. What should the upgrade notes (README "Upgrade notes" and the GitHub Release body) say?

A. A short `0.4.0` entry: new GitHub/GitLab/Bitbucket support (cloud only, read-and-link, admin token per workspace, project → repository mapping, auto-link by issue key); "Git access" for Backlog Git moved into Settings > Source control; nothing to do after upgrading, existing Backlog Git data unchanged
B. No upgrade note; only the Release body describes the feature
X. Other (please specify)

[Answer]: A
