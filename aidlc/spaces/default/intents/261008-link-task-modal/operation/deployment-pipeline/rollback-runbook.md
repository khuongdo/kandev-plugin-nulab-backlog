# Rollback Runbook - v0.5.2

The plugin has no in-place deployment to roll back. Rolling back means reinstalling the previous release and fixing forward, per the team Deployment practice. A released tag is never deleted or overwritten (project Forbidden rule).

## Triggers

- The plugin fails to install or start on the self-hosted Kandev.
- Any smoke check in deployment-strategy.md fails in a way that blocks linking. Examples: the task Link menu breaks, or the "Link to task" dialog cannot save.
- An error message shows Backlog response content or a secret (NFR1 breach).

## Steps

1. Note the failure for the fix-forward issue: a screenshot, plus Kandev server log lines with no secrets.
2. Mark the GitHub Release `v0.5.2` as broken with `gh release edit v0.5.2 --notes "<KNOWN ISSUE: ...> + original notes"`. Do not delete the tag or the Release.
3. Reinstall `v0.5.1` on the self-hosted Kandev: Settings > Plugins > Install plugin > From URL, with the `v0.5.1` Release package URL.
4. Verify that plugin 0.5.1 is running and that linking still works from the issue row "Link to task". Links saved by 0.5.2 keep working, because the stored format is unchanged.
5. If the marketplace registry PR for 0.5.2 is merged, open a PR that points the registry back to 0.5.1, or forward to the fix release.
6. Fix forward on a short-lived branch and release `v0.5.3` through the normal pipeline.

## Data

No migration and no stored-data change, so there is nothing to reverse.
