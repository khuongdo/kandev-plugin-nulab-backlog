# Health Check Report — release v0.4.2

## Release Pipeline Health

| Component | Status | Evidence |
|---|---|---|
| `main` branch | healthy | `3a983ab`, CI green on PR #13 (`checks`, `packaged-host-contract`) |
| GitHub Release `v0.4.2` | published | release assets present; checksum and attestation verified (see `smoke-test-results.md`) |
| Release asset download (From URL path) | healthy | HTTP 200, full 23,337,444 bytes |

## Self-Hosted Kandev

| Component | Status | Evidence |
|---|---|---|
| Kandev v0.97.0 service | running | listening on `:38429`, behind `tailscale serve` `https://webfrontier.tail152aaa.ts.net` |
| `nulab-backlog` plugin | not installed | uninstalled 2026-10-07T22:31Z; 0.4.2 install left to the maintainer (Q1 = B) |

## Root-Cause Status

- Upload path: the package is now 23.3 MB; browser upload still fails on links slower than about 0.78 MB/s because Kandev's `server.readTimeout` (30 s) is unchanged (a Kandev setting, out of scope).
- Reliable path: install From URL; the server downloads the package itself (60 s, 100 MiB limit), verified reachable above.

## Open Items

1. Install 0.4.2 From URL on the self-hosted Kandev and run the smoke check.
2. Marketplace registry pull request for 0.4.2.
