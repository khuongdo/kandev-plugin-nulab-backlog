# Deployment Pipeline Questions - 261008-gh-cli-profile

The team Deployment practice already settles strategy, gates, approvals, rollback and feature flags: tag `vX.Y.Z` on `main` → `release.yml` → GitHub Release with provenance; manual install; fix forward. Only release-specific questions remain.

State checked on 2026-10-08:
- The latest GitHub release is `v0.5.2` (2026-10-08T05:15:19Z).
- `origin/main` is at `ca8146c` (#22, AI-DLC Operation records for v0.5.2), which is this branch's base; no rebase needed.
- `manifest.yaml` on `origin/main` is `0.5.2`; this branch sets `0.5.3`.

## Question 1
Which version should this release be? It adds a user-visible choice: a gh account picker and a "Change account" button on the GitHub card, a new browser action (`scm.providers.cli_accounts`), a new error code, and the plugin no longer follows `gh auth switch`.

A. 0.5.3 (patch: improvement of the existing gh CLI login, like 0.5.2 was treated)
B. 0.6.0 (minor: a new per-workspace setting and a new action)
X. Other (please specify)

[Answer]: A

## Question 2
README upgrade note for this version (also put at the top of the GitHub Release notes). The proposed text is already in README `## Upgrade notes` as `### 0.5.3: choose the gh account per workspace` (three bullets: picker and Change account; `--user`, never `gh auth switch`; nothing to do after upgrading, logged-out message). Add a fourth bullet about task worktrees and gh versions?

A. Use the current text plus a bullet: "Agents in task worktrees get their GitHub login from Kandev's own GitHub integration (or the executor profile), not from this plugin; set it to the same account. The account list needs gh 2.81.0 or later; older gh shows only the active account (`--user` needs gh 2.40)."
B. Use the current text as is
X. Other (please specify)

[Answer]: A
