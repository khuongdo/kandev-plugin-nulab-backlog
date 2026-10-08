# Deployment Strategy: v0.6.0

## Strategy

Recreate-style plugin upgrade (the team's standard): a tagged GitHub Release produces the package; an admin installs it on the self-hosted Kandev server; the marketplace entry follows through a registry PR. No blue/green or canary: one plugin instance per Kandev server.

## Version

0.6.0 (minor), chosen by the user (Q1=A): new behaviour for existing workspaces and a new admin action `scm.active.set`.

## Steps

1. Re-check `gh release list` and `origin/main` (project rule). If a newer release landed, rebase and re-ask the version.
2. Bump `manifest.yaml` to 0.6.0 and rename the README section; add the Backlog Git worktree warning line.
3. `make check-format vet lint test coverage package verify-package` locally; open the PR with code + AI-DLC records; self-merge with squash when CI is green.
4. Re-check `gh release list` and `origin/main`; tag `v0.6.0` on the merge commit and push the tag.
5. Wait for `release.yml`; put the README 0.6.0 section at the top of the generated release notes (`gh release edit`).
6. Install the package on the self-hosted Kandev server; run the smoke check.
7. Open the marketplace registry PR.

## Smoke Check (after install)

- Plugin shows version 0.6.0 and starts (Settings > Integrations card loads).
- Settings > Source control shows the **Source control service** selector and a single framed card; an existing workspace with one connected service shows that service as active; a workspace with several shows the pick notice.
- Existing PR list for the active service still loads.

## Approvals

The deliberate creation of the `v0.6.0` tag is the production approval (team Deployment practice).
