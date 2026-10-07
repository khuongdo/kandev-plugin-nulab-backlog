# Smoke Test Results — 261007-uiux-github-style (release v0.1.1)

The post-install smoke check (deployment-strategy.md) was planned for the maintainer's self-hosted Kandev after installing `v0.1.1` (Q3 = A). On 2026-10-07 the maintainer decided to **skip the smoke check** and close the intent; every step is recorded as **Skipped**. The automated evidence below is the release verification of record.

| # | Check | Expected | Result |
|---|---|---|---|
| 1 | Plugin version and start | Kandev shows `nulab-backlog` `0.1.1`; no start-up error in the Kandev log | Skipped (maintainer decision) |
| 2 | Integrations menu and page | Exactly one Backlog entry; `/backlog` opens with Issues and Pull requests tabs | Skipped (maintainer decision) |
| 3 | Settings | Settings > Integrations > Backlog shows Connection and, when connected, PR watches, Issue watches and Saved PR queries | Skipped (maintainer decision) |
| 4 | Connected workspace | Issues tab lists issues; Pull requests tab lists PRs of a chosen repository | Skipped (maintainer decision) |

## Pre-install Evidence Already Collected

These automated checks ran against the exact release bytes and passed:

- `release.yml` `contract` job: the published package installed and ran on Kandev v0.96.0 (run 37573279932).
- Local packaged-host contract test on Kandev v0.96.0: 10/10 runs.
- `sha256sum -c checksums.txt` and `gh attestation verify` on the downloaded release assets.

