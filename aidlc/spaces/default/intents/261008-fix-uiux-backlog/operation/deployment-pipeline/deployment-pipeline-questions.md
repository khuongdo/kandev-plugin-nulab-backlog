# Deployment Pipeline Questions - 261008-fix-uiux-backlog

Context: the team Deployment practice already settles strategy, gates, approvals, rollback and feature flags (tag `vX.Y.Z` on `main` -> `release.yml` -> GitHub Release with provenance; manual install; fix forward). Only release-specific questions remain. Checked on 2026-10-08: latest GitHub release is `v0.4.2` (2026-10-07T23:52:26Z); `origin/main` is at `f3ca715` (three commits ahead of this branch: #14 CI path filter, #15 and #16 AI-DLC records), `manifest.yaml` there is still `0.4.2`.

## Question 1
Which version should this release be? It adds user-visible features (badge on Home > Tasks rows, sidebar and the task top bar with a hover card; issue summary stored on links) and changes behaviour (the Home > Integrations entry is hidden when Backlog is OFF everywhere at load).

A. 0.5.0 (minor: new features and a visible behaviour change)
B. 0.4.3 (patch)
X. Other (please specify)

[Answer]: A

## Question 2
README upgrade note for this version (also put at the top of the GitHub Release notes). Proposed text:

> ### 0.5.0: Backlog issue on task rows and the task top bar
> - The Backlog issue badge now shows on Home > Tasks rows, in the sidebar task list and in the task top bar (right of the workflow steps), not only on Kanban cards. Hover or focus it to see the issue key, summary and status; click it to open the issue.
> - The issue summary is filled in for existing links at the next issue sync; until then the badge shows the key and status.
> - Home > Integrations shows Backlog only when Backlog is on in at least one workspace. After turning Backlog on or off, reload the page. The Settings > Integrations card is always there.
> - Backlog settings: Projects now comes right after the connection; the empty issue watch list says "No issue watches yet"; the extra "Add watch" button inside empty watch lists is gone.
> - Nothing to do after upgrading: no setting or permission changes.

A. Use the proposed text
X. Other (please specify)

[Answer]: A
