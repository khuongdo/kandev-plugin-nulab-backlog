# Smoke Test Results — v0.1.0

| # | Test | Environment | Result |
|---|------|-------------|--------|
| S1 | Packaged-host contract test (install and run the exact Release package) | throwaway Kandev v0.96.0 in `release.yml` job `contract` | pass |
| S2 | Manual check steps 1–9: install, API-key connect, reload keeps the connection with an empty key field, `/backlog` page, switch off/on, no key or `apiKey=` in logs | self-hosted Kandev v0.97, `khuongdo.backlog.com` (local build of the same tree) | pass |
| S3 | After installing the Release asset: Settings > Integrations > Backlog shows `Connected as … @ khuongdo.backlog.com`; **Test connection** succeeds | self-hosted Kandev v0.97 | pass |

## Not run

Manual steps 10–26 (OAuth, Git and pull requests, Backlog issues, keyboard / screen reader / 320 px, 20-row latency). AC5.4.2, AC5.6.2, AC8.1.2, AC8.2.1 and AC8.2.4 stay open and are carried to performance-validation and later manual checks.
