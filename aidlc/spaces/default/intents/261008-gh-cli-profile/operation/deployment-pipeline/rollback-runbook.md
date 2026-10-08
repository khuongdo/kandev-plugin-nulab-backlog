# Rollback Runbook - v0.5.3

The plugin has no in-place deployment to roll back. Rolling back means reinstalling the previous release and fixing forward, per the team Deployment practice. A released tag is never deleted or overwritten (project Forbidden rule).

## Triggers

- The plugin fails to install or start on the self-hosted Kandev.
- An existing gh CLI workspace changes account after the upgrade, or a workspace calls GitHub as an account other than its chosen one.
- An error message or log shows a token or raw gh output (NFR1 breach).
- Any smoke check in deployment-strategy.md fails in a way that blocks the GitHub connection.

## Steps

1. Note the failure for the fix-forward issue: a screenshot plus Kandev server log lines with no secrets.
2. Mark the GitHub Release `v0.5.3` as broken with `gh release edit v0.5.3 --notes "<KNOWN ISSUE: ...> + original notes"`. Do not delete the tag or the Release.
3. Reinstall `v0.5.2` on the self-hosted Kandev: Settings > Plugins > Install plugin > From URL, with the `v0.5.2` Release package URL.
4. Verify plugin 0.5.2 is running and GitHub connections work. Note: 0.5.2 uses gh's active account for every gh CLI workspace; tell users to `gh auth switch` to the needed account meanwhile.
5. If the marketplace registry PR for 0.5.3 is merged, open a PR pointing the registry back to 0.5.2, or forward to the fix release.
6. Fix forward on a short-lived branch and release `v0.5.4` through the normal pipeline.

## Data

No migration and no stored-data change (the chosen login is the existing `AccountID`), so there is nothing to reverse.
