# Deployment Pipeline Questions - 261008-link-task-modal

The team Deployment practice already settles strategy, gates, approvals, rollback and feature flags: tag `vX.Y.Z` on `main` → `release.yml` → GitHub Release with provenance; manual install; fix forward. Only release-specific questions remain.

State checked on 2026-10-08:
- The latest GitHub release is `v0.5.1` (2026-10-08T04:08:52Z), published while this intent ran.
- `origin/main` is at `2182715` (#19, "Connect GitHub or GitLab with the gh / glab CLI login, version 0.5.1"), one commit ahead of this branch's base `d3d17e5`.
- `manifest.yaml` on `origin/main` is `0.5.1`. #19 also touched `ui/src/messages/en.ts` and `README.md`, which this change edits too.

## Question 1
Which version should this release be? It adds a user-visible feature: "Link Backlog issue" in the task's Link menu, with an issue key or link. It also changes the look and behaviour of the "Link to task" dialog: description line, Save button, Enter to save, a toast, and the badge updating at once.

A. 0.6.0 (minor: a new task-side action plus a visible dialog change)
B. 0.5.2 (patch: treat it as a UI fix)
X. Other (please specify)

[Answer]: B

## Question 2
README upgrade note for this version (also put at the top of the GitHub Release notes). Proposed text, with `<version>` from Question 1:

> ### <version>: link a Backlog issue from the task, GitHub-style
> - The task's Link menu (Kanban card menu, task switcher) has a new "Link Backlog issue" item, next to "GitHub Issue" and "Link Backlog pull request". Type an issue key such as `PROJ-123` or paste the issue link (`https://<space>.backlog.com/view/PROJ-123`, also `.backlog.jp` and `.backlogtool.com`) and press Save or Enter. The item is hidden while the task already has a Backlog issue; use "Unlink Backlog issue" first to change it.
> - The "Link to task" dialog on the Backlog issues page now looks like Kandev's GitHub link dialog: a short description, a Save button, Enter saves the chosen task, errors show in red under the list, and a "linked" message appears.
> - After linking from either place, the issue badge on the task updates at once instead of at the next refresh.
> - Nothing to do after upgrading: no setting or permission changes.

A. Use the proposed text
X. Other (please specify)

[Answer]: A

## Consolidated Summary Confirmation

- Q1 (B): the release is `0.5.2`. `manifest.yaml` goes from `0.5.1` (on `origin/main`) to `0.5.2`, tagged `v0.5.2` on `main` after the squash merge. The branch is first rebased onto `origin/main` (#19).
- Q2 (A): the proposed upgrade note goes into README `## Upgrade notes` as `### 0.5.2: link a Backlog issue from the task, GitHub-style`, above `0.5.1`. The same text goes at the top of the GitHub Release notes.

Does this all look correct before I generate the deployment pipeline artifacts?

- Looks correct
- Request changes

[Answer]: Looks correct
