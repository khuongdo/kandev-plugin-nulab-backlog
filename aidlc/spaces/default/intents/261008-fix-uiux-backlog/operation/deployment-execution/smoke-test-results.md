# Smoke Test Results - v0.5.0

## Release Artifact Checks (run here)

| Check | Command | Result |
|-------|---------|--------|
| Release exists with both assets | `gh release view v0.5.0` | pass: `nulab-backlog-0.5.0.tar.gz` (23,337,685 bytes), `checksums.txt` |
| Checksum matches | `sha256sum` of the downloaded package vs `checksums.txt` | pass: `0369931314ded2dcc2030164f3db48670447728c72f35089922f14b0e971f7ac` both |
| Build provenance | `gh attestation verify nulab-backlog-0.5.0.tar.gz --repo khuongdo/kandev-plugin-nulab-backlog` | pass (exit 0) |
| Package verification in the release workflow | `release.yml` job `verify` | pass |
| Installed and run on the minimum Kandev | `release.yml` job `contract` (Kandev v0.96.0) | pass |

## Self-Hosted Kandev Smoke Checks (your step, release scope B)

| # | Check (deployment-strategy.md) | Result |
|---|-------------------------------|--------|
| 1 | Plugin 0.5.0 installed and running | Pending - you install From URL |
| 2 | Badge and hover card on Home > Tasks rows; click opens the issue, not the task | Pending |
| 3 | Badge in the sidebar; issue button and Backlog PRs in the task top bar | Pending |
| 4 | Integrations entry hidden when Backlog is OFF everywhere (after reload); Settings card still there | Pending |
| 5 | Settings: Projects after Connection; "No issue watches yet"; one Add watch button | Pending |
