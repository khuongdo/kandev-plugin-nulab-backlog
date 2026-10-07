# Deployment Strategy — 261007-github-parity-actions (release v0.2.0)

## Strategy

**Recreate on manual install.** A Kandev plugin is one package installed into one Kandev server. There is no fleet, load balancer or traffic split, so blue/green, canary and rolling strategies do not apply. Installing `v0.2.0` replaces `v0.1.1` on the self-hosted server, and Kandev restarts the plugin process.

Risk is controlled before the install rather than during it:

- CI and `release.yml` run every check plus the packaged-host contract test on the minimum Kandev version. For this change it passed 10/10 locally (`construction/build-and-test/test-results.md`).
- The maintainer creates the release tag deliberately; that is the production approval.
- The package is verified (`verify-package`) and attested before it is published.

## Versioning

Semver tag `v0.2.0` (Q1, a minor bump for new features). `manifest.yaml` already says `0.2.0`, and `release-preflight` enforces the match. Released tags are never deleted or overwritten.

## Data and Compatibility

- No data migration is needed:
  - The new documents `issues.quick_actions` and `issues.queries` start empty, and quick actions fall back to the defaults.
  - Saved PR queries gain an optional `isDefault` field, and an absent field reads as false.
- The 7 new actions are additive (`issues.quick_actions.get/save`, `issues.queries.list/save/delete/set_default`, `git.queries.set_default`). No existing action key or access level changes.
- UI change: the issue row "Create task" item is removed and the tabs become a scope bar. Users are told through the upgrade notes (Q2).
- `min_kandev_version` stays `0.96.0`.

## Post-install Smoke Check (self-hosted Kandev)

1. The plugin shows version `0.2.0` and starts without errors in the Kandev log.
2. `/backlog` opens with the scope bar (Issues / Pull requests). The issue list shows open issues assigned to the connected user, with no filter chosen by hand.
3. An issue row's **+ Task** menu lists Implement / Investigate / Reproduce. Picking one opens the Kandev create-task dialog with the prefilled title and prompt. Creating the task links it to the issue.
4. Switching to Pull requests lists open PRs assigned to the connected user in the first repository.
5. Settings > Integrations > Backlog shows the Quick actions section, and the saved queries section with PR and issue queries.

Abort condition: if any step fails, follow `rollback-runbook.md`. The team's manual end-to-end check against a real Backlog space is not repeated for this release, because the team Testing Posture requires it exactly twice and both runs are done. This smoke check is the release health check required by the Operation phase rules.
