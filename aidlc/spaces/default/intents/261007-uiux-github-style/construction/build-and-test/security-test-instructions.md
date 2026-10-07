# Security Test Instructions — 261007-uiux-github-style

## Scope

Security requirements in force (requirements NFR2, project Mandated/Forbidden rules):

- API keys and tokens never appear in logs, error messages or UI responses — covered by the existing secret-leak test, extended to the new actions and watcher log lines.
- No real credentials in the repository or test data — `make check-secrets`.
- Static analysis with `gosec` through `golangci-lint` — `make lint`.
- New actions use `authenticated` access; connection changes and `issues.set_poll_interval` stay `admin` — manifest tests.
- No dependency vulnerability scan (team decision, Testing Posture).

## How to Run

```bash
make lint check-secrets
PATH="$HOME/.local/go/bin:$PATH" go test -race ./internal/plugin/... -run 'Leak|Secret|Redact|Manifest'
```

## Expected Result

All commands exit 0; no finding from `gosec`; the leak tests pass.

## Test Data

Synthetic keys from `internal/testutil` only.
