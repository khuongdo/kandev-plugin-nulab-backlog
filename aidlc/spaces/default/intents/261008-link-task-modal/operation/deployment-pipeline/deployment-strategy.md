# Deployment Strategy - v0.5.2

## Strategy

Tag-based release of a plugin package, installed manually on the self-hosted Kandev, per the team Deployment practice. There is no traffic shifting: Kandev runs one installed version at a time, so blue/green, canary and rolling do not apply.

No feature flag is added. The new task-side item follows the existing per-workspace Backlog switch, the same way "Link Backlog pull request" does.

## Environment Promotion

| Step | Environment | Gate |
|------|-------------|------|
| 1 | PR CI (GitHub Actions) | All required checks green, including the contract test on Kandev 0.96.0 |
| 2 | `main` | Squash merge (self-merge, protected branch) |
| 3 | GitHub Release `v0.5.2` | Deliberately creating the tag is the production approval; `release.yml` green; provenance attestation |
| 4 | Self-hosted Kandev | Manual install From URL; smoke checks below |
| 5 | Kandev marketplace | Registry PR, reviewed by the Kandev maintainers |

## Compatibility

- `min_kandev_version` stays `0.96.0`. No new host API: `registerTaskAction` with `placement: "link"`, `visible`, `openTaskLinkDialog` and `DialogDescription` all exist in 0.96.0, and the contract test passes on 0.96.0.
- The backend and stored data are unchanged (`issues.link` still takes `{issueKey}`), so upgrading and downgrading are safe.
- No setting, secret or permission changes.

## Smoke Checks After Install (self-hosted Kandev)

1. Settings > Plugins shows nulab-backlog 0.5.2 with status running.
2. On an unlinked task, the Link menu (Kanban card "..." > Link, and the task switcher) lists "Link Backlog issue" next to "GitHub Issue" and "Link Backlog pull request".
3. In that dialog, `PROJ-123` (a real key in a selected project) saves, a success toast appears, the dialog closes, and the task's issue badge shows the key without a page reload. Repeat with the issue's full `https://<space>.backlog.com/view/...` link on another task.
4. In the dialog, an unknown key shows "Issue … was not found." inline, and `http://…` or a non-Backlog link shows the "not a Backlog issue key or link" message without closing.
5. On a linked task, "Link Backlog issue" is hidden and "Unlink Backlog issue" is in the task menu. After unlinking, "Link Backlog issue" is back.
6. On the Backlog issues page, issue row "..." > "Link to task" opens the restyled dialog: a description line, a wider dialog, Save instead of Link. Choosing a task and pressing **Enter on the keyboard** saves it (architecture review R-04). A toast appears and the badge updates at once.

Abort condition: any smoke check fails in a way that blocks linking, or the plugin does not start. Then follow rollback-runbook.md.
