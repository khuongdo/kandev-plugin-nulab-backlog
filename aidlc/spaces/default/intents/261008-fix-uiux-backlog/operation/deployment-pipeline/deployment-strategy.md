# Deployment Strategy - v0.5.0

## Strategy

Tag-based release of a plugin package, installed manually on the self-hosted Kandev (team Deployment practice). There is no traffic shifting: Kandev runs one installed version at a time, so blue/green, canary and rolling do not apply. No feature flag is added: the Backlog on/off switch per workspace already gates the plugin, and the new Integrations-entry rule follows that switch.

## Environment Promotion

| Step | Environment | Gate |
|------|-------------|------|
| 1 | PR CI (GitHub Actions) | all required checks green, including the contract test on Kandev 0.96.0 |
| 2 | `main` | squash merge (self-merge, protected branch) |
| 3 | GitHub Release `v0.5.0` | deliberate tag = production approval; `release.yml` green; provenance attestation |
| 4 | Self-hosted Kandev | manual install From URL; smoke checks below |
| 5 | Kandev marketplace | registry PR reviewed by the Kandev maintainers |

## Compatibility

- `min_kandev_version` stays `0.96.0` (no new host API; `chat-top-bar`, `task-row-metadata` and the host Tooltip exist in 0.96.0).
- Stored links gain an optional `summary` field; links saved by 0.4.x load unchanged and get a summary at the next issue sync. A downgrade to 0.4.2 ignores the field.
- No setting, secret or permission changes.

## Smoke Checks After Install (self-hosted Kandev)

1. Settings > Plugins shows nulab-backlog 0.5.0, status running.
2. Home > Tasks: a task linked to a Backlog issue shows the badge; hover shows key, summary (after a sync) and status; click opens the issue in a new tab and does not open the task.
3. Sidebar task list shows the same badge; the task page top bar (right of the workflow steps) shows the issue button; linked Backlog PRs show in the same area (one button, or a dropdown for two or more).
4. With Backlog OFF in every workspace and the page reloaded, Home > Integrations has no Backlog entry; Settings > Integrations still shows the Backlog card. Turn it ON, reload, and the entry is back.
5. Backlog settings: Projects is right after the connection; an empty issue watch list says "No issue watches yet" with one Add watch button (top right); the same for an empty PR watch list.

Abort condition: any smoke check fails, or the plugin does not start -> follow rollback-runbook.md.
