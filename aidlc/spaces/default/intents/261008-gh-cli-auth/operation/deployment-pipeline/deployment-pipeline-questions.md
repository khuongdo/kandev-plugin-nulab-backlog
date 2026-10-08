# Deployment Pipeline Questions - 261008-gh-cli-auth

Context: the team Deployment practice already settles strategy, gates, approvals, rollback and feature flags (tag `vX.Y.Z` on `main` -> `release.yml` -> GitHub Release with provenance; manual install; fix forward). Only release-specific questions remain. Checked on 2026-10-08: latest GitHub release is `v0.5.0` (2026-10-08T02:44:37Z); `origin/main` is at `d3d17e5`, the base of this branch (no newer commits); `manifest.yaml` there is `0.5.0`.

## Question 1
Which version should this release be? It adds a user-visible feature (connect GitHub / GitLab with the gh / glab CLI login of the Kandev server, new admin action `scm.providers.use_cli`) and changes how GitLab tokens are sent (`Authorization: Bearer` instead of `PRIVATE-TOKEN`; personal access tokens keep working).

A. 0.6.0 (minor: new feature)
B. 0.5.1 (patch)
X. Other (please specify)

[Answer]: B

## Question 2
README upgrade note for this version (also put at the top of the GitHub Release notes). Proposed text:

> ### 0.6.0: connect GitHub or GitLab with the gh / glab CLI login
> - Settings > Source control: the GitHub card has a "Use gh CLI login" button and the GitLab card a "Use glab CLI login" button, next to the token field. The plugin then reads the token from the CLI that is logged in on the Kandev server (`gh auth token`, `glab config get token`), checks it, and never stores it. It follows `gh auth login` / `refresh` / `switch` within 5 minutes.
> - The CLI runs on the machine that runs Kandev, as the Kandev server user. It needs `gh` 2.17 or later / `glab`, logged in for that user; it does not work when Kandev runs where the CLI is not installed (for example a Docker image without it). If the CLI stops working, the card shows "gh CLI is not available or not logged in on the Kandev server"; there is no fallback to a typed token.
> - Saving a typed token switches back to the token method; Remove clears either method.
> - GitLab requests now send the token as `Authorization: Bearer`. Existing GitLab personal access tokens keep working; nothing to do.
> - Nothing to do after upgrading: existing connections stay on the token method.

A. Use the proposed text
X. Other (please specify)

[Answer]: A

Interpretation: with Q1 = B the heading of the note reads `### 0.5.1: connect GitHub or GitLab with the gh / glab CLI login`; the rest of the text is used as proposed.
