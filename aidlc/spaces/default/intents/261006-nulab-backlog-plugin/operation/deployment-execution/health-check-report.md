# Health Check Report — v0.1.0

| Check | How | Result |
|-------|-----|--------|
| Release integrity | `sha256sum -c checksums.txt` on the downloaded asset | OK |
| Provenance | `gh attestation verify nulab-backlog-0.1.0.tar.gz -R khuongdo/kandev-plugin-nulab-backlog` | OK (signer `release.yml@refs/tags/v0.1.0`) |
| Plugin running on the self-hosted Kandev | plugin installed and enabled; Backlog card shows the connection | OK |
| Backlog API reachable with stored credentials | **Test connection** | OK |
| No secret in logs | manual step 9 (key and `apiKey=` search) | OK |

## Verdict

Healthy (2026-10-07). No rollback needed.

## Open items

- Marketplace registry PR (postponed; draft in the appendix of `deployment-log.md`).
- Target-matrix items still Unverified from Build and Test: T-PERF-02, T-SEC-06, T-SEC-11, T-OBS-03, T-ACC-02, T-REL-08, T-MKT-01, T-MAN-01.
- Plugin stderr is not visible in the Kandev backend logs (observability-setup).
