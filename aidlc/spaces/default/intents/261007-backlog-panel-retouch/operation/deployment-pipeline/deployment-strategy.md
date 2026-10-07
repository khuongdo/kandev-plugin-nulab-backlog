# Deployment Strategy — 261007-backlog-panel-retouch (release v0.4.1)

## Strategy

**Recreate on manual install.** A Kandev plugin is a package installed into one Kandev server; there is no fleet, load balancer or traffic split, so blue/green, canary and rolling strategies do not apply. Installing `v0.4.1` replaces `v0.4.0` on the self-hosted server; Kandev reloads the plugin UI bundle and restarts the plugin process.

Risk is controlled before install instead of during it:

- Full suite green on the rebased code (Vitest 373/373, Go `-race`, coverage 92.8%) and the packaged-host contract test 10/10 on Kandev 0.96.0 with `nulab-backlog-0.4.1.tar.gz` (Build and Test).
- **Pre-release real-host UI check** (closes NFR2-UI-IN-HOST): install the pull-request package on the self-hosted Kandev and run the checklist in `construction/build-and-test/integration-test-instructions.md` before the tag is created, including the provider (GitHub/GitLab/Bitbucket) pull-request lists and the open review Minors: touch input on the filters (R-01), toolbar total after a failed load or workspace switch (R-04), visual cue on the last checked status (R-05), "All repositories" on a provider list (R-06).
- The release tag is created deliberately by the maintainer (the production approval), after re-checking `gh release list` and `origin/main`.
- The package is verified (`verify-package`) and attested before it is published.

## Versioning

Semver tag `v0.4.1` (Q3, patch after `v0.4.0`; the human chose a patch although the UI behaviour changes). `manifest.yaml` must say `0.4.1` on `main` before the tag; `release-preflight` enforces it. Released tags are never deleted or overwritten.

## Data and Compatibility

- No data migration, no schema, setting or action change; `min_kandev_version` stays `0.96.0`.
- Behaviour changes for users: issue search runs on Enter; filters and task display look like the GitHub list on every list; the Kanban badge opens the Backlog issue; provider pull-request lists keep at least one status selected. Documented in the README and Release notes.
- Saved queries (including provider saved queries from v0.4.0), issue links, watches and connections are unaffected.

## Post-install Smoke Check (self-hosted Kandev)

1. The plugin shows version `0.4.1` and starts without errors in the Kandev log.
2. `/backlog` issue list: toolbar on one row (desktop); typing does not reload; Enter reloads; Project/Status/Assignee dropdowns filter the list.
3. An issue with a linked task shows the task title; clicking opens the task. An issue title opens the Backlog issue in a new tab.
4. Pull requests: the provider selector is first; Backlog Git and one connected provider each show the same toolbar layout without search, a working "Status (n)" filter and task titles.
5. A Kanban card's Backlog badge opens the Backlog issue in a new tab without opening the card.

The team's manual end-to-end check against a real Backlog space is not repeated as a separate ceremony (team Testing Posture: exactly twice, both done); the pre-release UI check above is this release's verification of NFR2.
