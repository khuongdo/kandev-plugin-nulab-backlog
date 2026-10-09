# Health Check Report - v0.6.0

## Release Health

| Signal | Status |
|---|---|
| `main` CI on PR #26 | Green (changes, checks, packaged-host-contract, secret-scan) |
| Release workflow | Success (verify, contract, publish) |
| Release assets | Present: package (23,586,356 bytes) and `checksums.txt`; checksums and package verification pass; attestation verified |
| Release notes | README 0.6.0 note at the top, generated PR list below |

## Runtime Health

Not observed in this stage: the self-hosted install is done by the user (Q1 = B). The plugin has no separate health endpoint; health after install is the smoke checks in `smoke-test-results.md` (plugin active at 0.6.0, Source control section renders, PR list loads).

## Verdict

Release is healthy and ready to install. Runtime health pending the user's install.
