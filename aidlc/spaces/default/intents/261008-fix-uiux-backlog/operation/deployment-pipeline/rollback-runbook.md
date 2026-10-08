# Rollback Runbook - v0.5.0

The plugin has no in-place deployment to roll back; rollback means reinstalling the previous release and fixing forward (team Deployment practice). A released tag is never deleted or overwritten (project Forbidden rule).

## Triggers

- The plugin fails to install or start on the self-hosted Kandev.
- Any smoke check in deployment-strategy.md fails in a way that blocks daily use (for example the badge breaks the task list, or the Backlog page cannot be reached while Backlog is ON).
- An error message or log shows Backlog issue content or a secret (NFR1 breach).

## Steps

1. Note the failure (screenshot, Kandev server log lines without secrets) for the fix-forward issue.
2. Mark the GitHub Release `v0.5.0` as broken: `gh release edit v0.5.0 --notes "<KNOWN ISSUE: ...> + original notes"` (do not delete the tag or the Release).
3. Reinstall `v0.4.2` on the self-hosted Kandev: Settings > Plugins > Install plugin > From URL with the `v0.4.2` Release package URL.
4. Verify: plugin 0.4.2 running; the Kanban badge, the Backlog page and the settings card work. Links saved by 0.5.0 keep working (0.4.2 ignores the extra `summary` field).
5. If the marketplace registry PR for 0.5.0 is merged, open a PR pointing the registry back to 0.4.2 or forward to the fix release.
6. Fix forward on a short-lived branch and release `v0.5.1` through the normal pipeline.

## Data

No migration to reverse. The only data change is the optional `summary` on stored links, which older versions ignore.
