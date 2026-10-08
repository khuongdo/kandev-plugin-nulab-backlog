# Deployment Execution Questions - 261008-source-control-settings (release v0.6.0)

Pre-deployment facts (checked, not asked):

- **Build and Test passed every target:** Go `-race` all packages, coverage 92.9%; UI 446/446; package verified; contract test 10/10 on Kandev 0.96.0.
- **No data migration.** The settings document gains an optional `active` field (schema version 1 unchanged); existing workspaces derive their active service on first read, and workspaces with two or three external services show a pick notice.
- **Dependent services:** GitHub (CI, Releases) and the self-hosted Kandev.
- **Branch base is `origin/main` `6d43d69`** (latest release `v0.5.3`); no rebase needed unless `main` moves.
- **Your go-ahead is needed for every step that publishes outside this machine:** push, pull request, merge, tag, install, and the marketplace pull request.

## Question 1
How far should I carry out the release now?

A. All the way: bump `manifest.yaml` to 0.6.0 and finalize the README note, commit, push, open the pull request as `khuongdo`, wait for CI, squash-merge, re-check releases, tag `v0.6.0`, wait for `release.yml`, put the upgrade note at the top of the Release notes, install on the self-hosted Kandev and run the smoke checks, then check whether the marketplace registry needs a pull request
B. Up to the GitHub Release: everything in A except the self-hosted install and the marketplace step (you install yourself)
C. Up to the pull request only: bump, commit, push and open the pull request; you merge and tag yourself
X. Other (please specify)

[Answer]: B
