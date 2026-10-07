# Smoke Test Results — 261007-opt-in-default (v0.3.0)

## Status

**Not run.** The release is not published or installed on the self-hosted Kandev yet (Q2 = C: record as not installed).

## Smoke Check to Run After Install

From `operation/deployment-pipeline/deployment-strategy.md`:

| # | Step | Result |
|---|---|---|
| 1 | Plugin shows version `0.3.0` and starts without errors in the Kandev log | Not run |
| 2 | A workspace connected on an older version but never switched shows the Off state with the existing wording; its saved connection is still listed | Not run |
| 3 | As an admin, turn the switch on and save: the Backlog page lists issues again without reconnecting; watches resume | Not run |
| 4 | As a non-admin member, the switch cannot be changed | Not run |

## Automated Substitute Evidence

The packaged-host contract test installs the real `0.3.0` package on Kandev v0.96.0 and checks the fresh-install behaviour (off by default, connect refused with `409 integration_disabled`, switch on, connect validation runs): passed 3/3 (see `health-check-report.md`). This does not replace steps 2–4 on a real upgraded workspace.
