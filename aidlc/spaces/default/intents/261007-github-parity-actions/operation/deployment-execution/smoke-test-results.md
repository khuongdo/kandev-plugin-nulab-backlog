# Smoke Test Results — 261007-github-parity-actions (release v0.2.0)

## Release-level Checks (done)

| Check | Evidence | Result |
|---|---|---|
| Package installs and runs on the minimum Kandev (0.96.0) | `release.yml` `contract` job (run 37585429654) on the exact released bytes; local `make contract-test` 10/10 | Pass |
| Published package integrity | `sha256sum -c checksums.txt` | Pass |
| Provenance | `gh attestation verify nulab-backlog-0.2.0.tar.gz --repo khuongdo/kandev-plugin-nulab-backlog` | Pass |
| Version | Release asset `nulab-backlog-0.2.0.tar.gz`; `release-preflight` matched the tag to `manifest.yaml` | Pass |

## Self-hosted Smoke Check (not run)

The 5-step smoke check in [deployment-strategy.md](../deployment-pipeline/deployment-strategy.md) needs `0.2.0` installed on the self-hosted Kandev. The maintainer chose to skip the install for now (Deployment Execution Q2 = B), so these steps were not run:

1. Plugin shows `0.2.0` and starts cleanly.
2. `/backlog` opens with the scope bar on the "Assigned to me, open" issue list.
3. "+ Task" → Investigate opens the prefilled create-task dialog, and the created task is linked.
4. Pull requests opens on "Open, assigned to me" for the first repository.
5. Settings shows Quick actions and the saved queries section with PR and issue queries.

Status: **not run, released but not installed**. Run these steps after a later install, and roll back per `rollback-runbook.md` if any step fails.
