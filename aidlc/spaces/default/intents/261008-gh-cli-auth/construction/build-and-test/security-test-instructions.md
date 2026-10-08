# Security Test Instructions

## Threat notes for this change (STRIDE, short)

- **Information disclosure**: the CLI token could leak through logs, error messages, the provider list sent to the UI, or plugin state. Controls: the token is never stored; it is registered with `redact.WithSecrets`; CLI stderr is discarded; every CLI failure is the fixed `ErrCLIUnavailable`; `ProviderView` carries only `method` and account. Checked by the leak tests in `internal/scm/service_test.go` and `internal/scm/cli_token_test.go`.
- **Tampering / command injection**: the CLI is run without a shell, with fixed arguments from `cliCommand` (no user input in arguments), 10-second timeout, 4 KiB stdout cap. Checked by `cli_token_test.go` (exact name/args) and gosec (G204 suppressed only with a documented reason).
- **Elevation of privilege**: `scm.providers.use_cli` lets a workspace use the Kandev server user's CLI login, so it is `access: admin` in `manifest.yaml`, like `set_token`; members only see the method and account. Checked by the manifest parity test and the UI read-only test.
- **Accepted risk**: every workspace admin on the same Kandev server can connect with the server's single gh/glab login. This is inherent to the requested feature and documented in the README.

## How to run

```bash
export PATH="$HOME/.local/go/bin:$PATH"
make lint            # golangci-lint with gosec
make check-secrets   # no credentials committed
go test -race ./internal/scm/ -run 'CLI|Leak|Secret'
```

No dependency vulnerability scan is run (team decision).

## Expected result

All commands pass; no token text in any test output.
