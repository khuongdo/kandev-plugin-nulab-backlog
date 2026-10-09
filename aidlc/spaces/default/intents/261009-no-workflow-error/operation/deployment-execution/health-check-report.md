# Health Check Report - v0.6.1

## Release Health

| Item | Status |
|---|---|
| GitHub Release `v0.6.1` published, marked Latest | Healthy |
| Assets present (`nulab-backlog-0.6.1.tar.gz`, `checksums.txt`) | Healthy |
| Integrity (checksum, package verification, attestation) | Healthy |
| `main` at `6f152d8`, version `0.6.1`, CI green | Healthy |

## Runtime Health

The plugin runs inside each self-hosted Kandev; there is no hosted service to probe. Health after install is the Kandev plugin page showing `nulab-backlog@0.6.1` enabled, plus the functional smoke check (smoke-test-results.md). The packaged-host contract test (install + start on Kandev 0.96.0) passed in the release pipeline.

## Open Items

- User installs 0.6.1 on the self-hosted Kandev, approves the `workflows` read permission, and re-runs the smoke check.
- Marketplace registry step (user).
- Follow-ups from reviews: pending indicator on "+ Task" while the workflow check runs (R-02); workflow lookup in the three watch dialogs (Q2=A).
