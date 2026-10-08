# Deployment Pipeline Questions

Context: the team Deployment practice already settles strategy, gates, approvals, rollback and feature flags (tag `vX.Y.Z` on `main` -> `release.yml` -> GitHub Release -> marketplace registry PR; manual install on the self-hosted Kandev server; fix-forward rollback). Only release-specific questions remain. Checked on 2026-10-09: latest GitHub release is v0.5.3, `origin/main` is at 6d43d69 (same as this branch's base), `manifest.yaml` is 0.5.3.

## Question 1
Which version should this change ship as? It changes behaviour for existing workspaces: only one source control service works at a time, and while GitHub/GitLab/Bitbucket is active, Backlog Git pull requests and Git access are off.

A. 0.6.0 (minor: new behaviour and a new admin action `scm.active.set`)
B. 0.5.4 (patch)
X. Other (please specify)

[Answer]: A

## Question 2
The README has an upgrade note under "Unreleased: one source control service per workspace" (selector, switching keeps data disabled, upgrade rules, service-named labels). How should it go into the release?

A. Rename the heading to the chosen version and use it as-is at the top of the GitHub Release notes; also add one line warning that task worktrees cannot fetch or push Backlog Git repositories while an external service is active
B. Rename the heading and use it as-is, without the extra warning line
X. Other (please specify)

[Answer]: A
