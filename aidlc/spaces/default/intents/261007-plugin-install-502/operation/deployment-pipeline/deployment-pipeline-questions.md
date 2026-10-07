# Deployment Pipeline Questions — 261007-plugin-install-502

Context: the release pipeline already exists and the team's Deployment practice settles strategy, gates, approvals and rollback (pull request to `main` with CI green, squash merge, deliberate `vX.Y.Z` tag on `main` → `release.yml` re-runs every check and package verification → GitHub Release with checksums and provenance → marketplace registry pull request; manual install on the self-hosted Kandev; rollback = reinstall the previous version and fix forward; no feature flags). So only release-specific questions are asked.

Checked before asking (2026-10-07): the latest GitHub release is `v0.4.1` (2026-10-07T22:30Z); `origin/main` is at `bf20039` ("AI-DLC Operation records for v0.4.1 (#12)"); this branch is based on `1819cc3` (v0.4.1), one commit behind `origin/main`, so it must be rebased before the pull request. `manifest.yaml` says `0.4.1`. Answer by writing the letter after the answer tag.

## Question 1
Which version should this fix be released as? It changes no plugin behaviour, but it drops the Windows server executable, so a Kandev server running on Windows can no longer install or upgrade the plugin (requirements FR1.3 asked for a patch; the code review flagged the platform drop as breaking).

A. `0.4.2` — patch, as the requirements say; the platform drop is called out in the release notes
B. `0.5.0` — minor, to signal the dropped Windows support through the version number
X. Other (please specify)

[Answer]: A

## Question 2
What should the release notes say?

A. A short entry in the README `## Upgrade notes` and the same text in the GitHub Release notes: smaller package (about 23 MB) to avoid the 30-second upload limit; install From URL recommended; Windows servers no longer supported
B. GitHub Release notes only; no README upgrade-notes entry
X. Other (please specify)

[Answer]: A

## Consolidated Summary Confirmation

- Version (Q1): release as `0.4.2` (patch); `manifest.yaml` moves from `0.4.1` to `0.4.2` and the tag is `v0.4.2` (latest release is `v0.4.1`). Rebase onto `origin/main` (`bf20039`) and re-check `gh release list` again right before tagging.
- Release notes (Q2): a short "0.4.2" entry under the README `## Upgrade notes` and the same text in the GitHub Release notes: smaller package (about 23 MB, 4 executables) so browser upload fits Kandev's 30-second read limit more often; install From URL is the recommended path; Windows-hosted Kandev servers are no longer supported (they cannot install or upgrade to 0.4.2).
- Everything else follows the team's Deployment practice unchanged: pull request to `main` with CI green, squash merge, deliberate `v0.4.2` tag, `release.yml` re-runs all checks and package verification, GitHub Release with checksums and provenance, marketplace registry pull request, manual install on the self-hosted Kandev (From URL), rollback by reinstalling `v0.4.1` and fixing forward; no feature flags.

Does this all look correct before I generate the deployment pipeline artifacts?

- Looks correct
- Request changes

[Answer]: Looks correct
