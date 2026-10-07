# Deployment Strategy — 261007-plugin-install-502 (release v0.4.2)

## Strategy

**Recreate (single instance), manual.** The plugin runs as one process per Kandev server; installing a new version stops the old process and starts the new one. There is no traffic shifting, canary or blue/green for a self-hosted plugin. This follows the team Deployment practice.

## Steps

1. Rebase the branch onto `origin/main`, push, open a pull request to `main`; wait for all CI checks (including the packaged-host contract job).
2. Self-merge with squash.
3. Re-check `gh release list` and `origin/main`; confirm the latest release is still `v0.4.1` and `main` carries `version: "0.4.2"`.
4. Create and push the tag `v0.4.2` on `main` (manual production approval). Watch `release.yml` until the GitHub Release `v0.4.2` exists with `nulab-backlog-0.4.2.tar.gz` and `checksums.txt`.
5. Install on the self-hosted Kandev (v0.97.0, `https://webfrontier.tail152aaa.ts.net`): **Settings > Plugins > Install plugin > From URL** with `https://github.com/khuongdo/kandev-plugin-nulab-backlog/releases/download/v0.4.2/nulab-backlog-0.4.2.tar.gz`.
6. Send the marketplace registry pull request (`make marketplace-entry`).

## Smoke Check (after step 5)

- The install request returns success (no `Plugin install failed`); the backend log shows `POST /api/plugins/install` with a 2xx status and no 400 `missing multipart field "package"`.
- **Settings > Plugins** lists `nulab-backlog` `0.4.2` as active; the process `~/.kandev/plugins/nulab-backlog/0.4.2/server/plugin-linux-amd64` runs.
- **Settings > Integrations > Backlog** opens; after turning Backlog on, the issue list loads for the connected space.
- Optional evidence for the root cause fix: a browser **Upload file** install of the 23.4 MB package from a fast link also succeeds (not required; slow links are documented to use From URL).

## Abort Conditions

- CI or `release.yml` fails: stop, fix forward on a new pull request; never move or delete a tag.
- From URL install fails on the self-hosted Kandev: read `~/.kandev/logs/backend-logs.log` for the `/api/plugins/install` line and follow `rollback-runbook.md`.

## Security Implications

No IAM, network or encryption change. The package carries fewer executables and the verifier rejects any unexpected file under `server/`, which narrows what a release can contain. The release keeps its provenance attestation.
