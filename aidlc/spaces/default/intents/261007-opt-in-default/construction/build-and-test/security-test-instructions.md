# Security Test Instructions — Opt-in by default

## Security Assessment of the Change (STRIDE, brief)

- **Elevation of privilege / secure by default**: the change makes the plugin *more* secure by default — nothing calls Backlog, syncs, or serves Git credentials until an admin explicitly opts in. `connection.set_enabled` stays admin-only (manifest `access: admin`, checked by `internal/plugin/manifest_test.go`) and touches only the switch record.
- **Fail closed**: an unreadable switch record is still an error, never "on" (NFR3.9 of intent 261006, kept by FR1.1; covered in `internal/connection/store_test.go`).
- **Information disclosure**: no new logs, errors or responses carry secrets; existing redaction tests remain green.
- **Denial of service**: workers stay started but idle for off workspaces; no new loops or network calls.

## How to Run

```bash
export PATH=$HOME/.local/go/bin:$HOME/go/bin:$PATH
make lint            # golangci-lint with gosec (SAST)
make check-secrets   # no credential-shaped strings in test data or artifacts
go test -race ./internal/connection/... ./internal/plugin/...   # fail-closed switch, guard refusals, redaction tests
```

## Expected Result

`0 issues.`, `ci secrets: OK`, all packages `ok`. Dependency vulnerability scanning is intentionally not run (team decision recorded in Testing Posture).
