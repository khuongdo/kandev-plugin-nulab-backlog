# Smoke Test Results - v0.6.0

## Executed

| Check | Result | Evidence |
|---|---|---|
| Release workflow re-runs all checks on the tag | Pass | `release.yml` run 37861147005: verify, contract, publish all success |
| Packaged plugin runs on Kandev 0.96.0 | Pass | `contract` job in the release run; PR CI `packaged-host-contract`; local 10/10 in Build and Test |
| Released package integrity | Pass | `sha256sum -c checksums.txt` OK; `verifypkg` OK `nulab-backlog@0.6.0` |
| Build provenance | Pass | `gh attestation verify nulab-backlog-0.6.0.tar.gz --repo khuongdo/kandev-plugin-nulab-backlog` exit 0 |

## Not Executed (user installs, Q1 = B)

The post-install checks from `deployment-strategy.md` are left to the user after installing v0.6.0 on the self-hosted Kandev server:

1. Plugin shows version 0.6.0 and the Settings > Integrations card loads.
2. Settings > Source control shows the **Source control service** selector and one framed card; a workspace with one connected service shows it as active; a workspace with several shows the pick notice.
3. The PR list for the active service still loads.
