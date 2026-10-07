# Manual check: first-release

> **Never paste an API key, a token, or any URL with a query string (`?...`) into this file.**
> Record the space domain only, for example `myteam.backlog.com`.

| Field | Value |
|-------|-------|
| Date | 2026-10-07 |
| Check name | first-release |
| Kandev version | v0.97 |
| Plugin commit | 0545212 (package `nulab-backlog-0.1.0.tar.gz` built from an identical tree, `verifypkg` OK) |
| Space domain | not recorded |
| Result | pass |
| Connect duration (seconds) | not measured |
| Logs checked for the key | yes |

## Steps

Steps 1–9 (install, API-key connect, reload, `/backlog` page, switch off and on, log search for the key and `apiKey=`): **pass**, run by the maintainer on the self-hosted Kandev against a real Backlog space.

Steps 10–26 (OAuth, Git and pull requests, Backlog issues, keyboard / screen reader / 320 px, 20-row latency): **not run** in this check. AC5.4.2, AC5.6.2, AC8.1.2, AC8.2.1 and AC8.2.4 stay open.

## Notes

This record also serves as the walking-skeleton real-space check (steps 1–9 are the skeleton steps). The check ran on Kandev v0.97; the minimum supported version v0.96.0 is covered by the automated `packaged-host-contract` job.
