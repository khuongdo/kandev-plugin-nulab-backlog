# Health Check Report — v0.5.3

Checked 2026-10-08 ~07:15 UTC.

| Component | Status | Evidence |
|---|---|---|
| GitHub `main` | healthy | `81608d4` (#23), all required checks green |
| `release.yml` run 37741700665 | success | verify, contract, publish all success |
| GitHub Release `v0.5.3` | published | package 23,544,506 bytes + `checksums.txt`; build provenance attestation by the publish job |
| Kandev service (self-hosted) | running | `http://localhost:38429` answers `GET /api/plugins` (HTTP 200) |
| Plugin `nulab-backlog` | active, 0.5.3 | `GET /api/plugins`; `restart_count` 0 at install |
| Other plugin `kandev-provider-usage` | active, 0.10.0 | unchanged |
| Plugin action path | healthy | `scm.providers.list` HTTP 200; `scm.providers.cli_accounts` HTTP 409 `integration_disabled` (expected while the switch is off) |

## Observations

- No errors returned by the install or action calls; no secrets in any response.
- The plugin has no hosted runtime of its own beyond the Kandev-managed process; there are no metrics or alerts to watch (Observability Setup has nothing extra to configure).

## Follow-ups

- Owner: run smoke checks 2-7 (`smoke-test-results.md`) after turning the Backlog integration on in a workspace and connecting GitHub with "Use gh CLI login".
- Marketplace: PR kdlbs/kandev#4284 still awaits Kandev maintainer review.
