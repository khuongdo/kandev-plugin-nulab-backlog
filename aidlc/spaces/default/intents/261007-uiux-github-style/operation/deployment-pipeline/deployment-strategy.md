# Deployment Strategy — 261007-uiux-github-style (release v0.1.1)

## Strategy

**Recreate on manual install.** A Kandev plugin is a package installed into one Kandev server; there is no fleet, load balancer or traffic split, so blue/green, canary and rolling strategies do not apply. Installing `v0.1.1` replaces `v0.1.0` on the self-hosted server; Kandev restarts the plugin process.

Risk is controlled before install instead of during it:

- CI and `release.yml` run every check plus the packaged-host contract test on the minimum Kandev version (locally passed 10/10 for this change).
- The release tag is created deliberately by the maintainer (the production approval).
- The package is verified (`verify-package`) and attested before it is published.

## Versioning

Semver tag `v0.1.1` (Q1). `manifest.yaml` must say `0.1.1` on `main` before the tag; `release-preflight` enforces it. Released tags are never deleted or overwritten.

## Data and Compatibility

- No data migration: stored PR watches, saved queries, issue links and settings keep their schema; the saved-query `creator` field is optional and absent means "anyone"; issue watches and their ledger are new documents.
- Six new actions (`issues.watches.*`) and `git.prs.list` are additive; no existing action key or access level changes.
- Two UI routes are removed (`/backlog/watches`, `/backlog/dashboard`); users are told through the upgrade notes (Q2).
- `min_kandev_version` stays `0.96.0`.

## Post-install Smoke Check (self-hosted Kandev)

1. The plugin shows version `0.1.1` and starts without errors in the Kandev log.
2. Home > Integrations shows exactly one Backlog entry; `/backlog` opens with Issues and Pull requests tabs.
3. Settings > Integrations > Backlog shows the Connection section and, when connected, the PR watches, Issue watches and Saved PR queries sections.
4. On a connected workspace: the Issues tab lists issues; the Pull requests tab lists PRs of a chosen repository.

The team's manual end-to-end check against a real Backlog space is not repeated for this release (team Testing Posture: exactly twice, both already done); this smoke check is the release verification required by the Operation phase rules.
