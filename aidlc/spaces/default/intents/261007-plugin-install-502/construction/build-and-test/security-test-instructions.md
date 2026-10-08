# Security Test Instructions — Plugin install failed: 502

## Scope at Minimal Strategy

No new security test suite is generated. The change touches only packaging, the offline package verifier and README; it adds no input surface, credential handling or network call. The existing security checks stay mandatory.

## Checks to Run

```bash
export PATH=$HOME/.local/go/bin:$PATH
make lint            # golangci-lint default set + gosec, tsc, eslint, actionlint, workflow policy
make check-secrets   # no credential-shaped string in test data or artifacts
make verify-package  # checksums, no unlisted or unexpected files, no Nulab asset URLs in the bundle
```

## Security Review of the Change (STRIDE, brief)

- **Tampering**: the verifier now also rejects any `server/` file that is not one of the 4 listed executables, even when it is checksummed. This narrows what a tampered or stale package can carry.
- **Information disclosure**: README troubleshooting text names only public settings (`KANDEV_SERVER_READTIMEOUT`) and a public log message; no secrets.
- No dependency vulnerability scan is added (team decision, Testing Posture).
