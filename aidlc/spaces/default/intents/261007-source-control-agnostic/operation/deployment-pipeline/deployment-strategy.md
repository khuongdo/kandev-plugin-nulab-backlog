# Deployment Strategy — 261007-source-control-agnostic (release v0.4.0)

## Strategy

**Recreate on manual install.** A Kandev plugin is a package installed into one Kandev server; there is no fleet or traffic split, so blue/green, canary and rolling strategies do not apply. Installing `v0.4.0` replaces `v0.3.0`; Kandev restarts the plugin process.

Risk is controlled before install:

- Full suite (1324 Go tests with `-race`, 364 Vitest tests), 92.8% coverage, and the packaged-host contract test passed 10/10 on Kandev 0.96.0 (Build and Test).
- The v0.3.0 regression test proves existing `git.*` data and actions read unchanged.
- The release tag is created deliberately by the maintainer (the production approval); the package is verified and attested before publishing.

## Versioning

Semver tag `v0.4.0` (Q1, minor: new backward-compatible feature). `manifest.yaml` must say `0.4.0` on `main` before the tag; `release-preflight` enforces it. Released tags are never deleted or overwritten.

## Data and Compatibility

- No migration and no change to existing documents or secrets. New data lives in new `scm.*` documents and new `backlog.scm.<provider>.<ws>` secrets, created only when an admin configures a provider.
- No existing action key, input or access level changes; `min_kandev_version` stays `0.96.0`.
- Visible UI change: "Git access" moved into the new "Source control" section; members now see that section read-only.
- Outbound calls go only to `api.github.com`, `gitlab.com` and `api.bitbucket.org`, and only after an admin adds a token. Self-hosted Kandev servers behind a restrictive egress firewall must allow those hosts to use the feature.

## Post-install Smoke Check (self-hosted Kandev)

1. The plugin shows version `0.4.0` and starts without errors in the Kandev log.
2. Existing Backlog features work as before: issue list, a linked Backlog Git PR shows its status, an existing saved PR query runs.
3. Settings > Integrations > Backlog shows "Source control" with Backlog Git, GitHub, GitLab and Bitbucket; "Git access" is inside it; a non-admin sees it read-only.
4. Optional, when the maintainer has a test account: add a read-only token for one provider, press Test (account name shown), map one Backlog project to one repository, and open the PR list for that provider. This is the first real-account check of the new clients (Build and Test used fake servers only); record the result.

The team's manual end-to-end check against a real Backlog space is not repeated (team Testing Posture: exactly twice, already done).
