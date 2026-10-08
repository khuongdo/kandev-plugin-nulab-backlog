# Deployment Strategy — 261007-opt-in-default (release v0.3.0)

## Strategy

**Recreate on manual install.** A Kandev plugin is a package installed into one Kandev server; there is no fleet, load balancer or traffic split, so blue/green, canary and rolling strategies do not apply. Installing `v0.3.0` replaces `v0.2.0` on the self-hosted server; Kandev restarts the plugin process.

Risk is controlled before install instead of during it:

- The branch is rebased onto `v0.2.0` and the full suite plus the packaged-host contract test re-run on the rebased code (the contract test asserts a fresh install is off and refuses connect until the switch is on).
- The release tag is created deliberately by the maintainer (the production approval).
- The package is verified (`verify-package`) and attested before it is published.

## Versioning

Semver tag `v0.3.0` (Q1). `manifest.yaml` must say `0.3.0` on `main` before the tag; `release-preflight` enforces it. Released tags are never deleted or overwritten.

## Data and Compatibility

- No data migration and no schema change: the switch record format is unchanged; only the meaning of a *missing* record changes (now off).
- Behaviour change for upgraders: workspaces with no switch record turn off after the upgrade (accepted, requirements FR5; documented in the upgrade notes). Saved connections, watches, saved queries and issue links are kept and resume when the admin turns the switch on.
- No action key or access level changes; `connection.set_enabled` stays admin-only and unguarded. `min_kandev_version` stays `0.96.0`.

## Post-install Smoke Check (self-hosted Kandev)

1. The plugin shows version `0.3.0` and starts without errors in the Kandev log.
2. On a workspace that was connected but never touched the switch: Settings > Integrations > Backlog shows the Off state with the existing wording; the saved connection is still listed.
3. As an admin, turn the switch on and save: the Backlog page lists issues again without reconnecting; watches resume.
4. As a non-admin member, the switch cannot be changed.

The team's manual end-to-end check against a real Backlog space is not repeated for this release (team Testing Posture: exactly twice, both already done); this smoke check is the release verification required by the Operation phase rules.
