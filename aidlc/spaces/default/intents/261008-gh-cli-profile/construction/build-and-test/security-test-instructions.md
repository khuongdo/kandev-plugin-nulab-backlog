# Security Test Instructions — 261008-gh-cli-profile

## Threats Considered (STRIDE, brief)

- **Tampering / injection**: a browser-supplied login becomes a gh argument. Mitigation: login rule `^[A-Za-z0-9](?:-?[A-Za-z0-9])*$`, ≤ 39 chars, checked before any CLI run; no shell; fixed argv. Test: AC1.1.5 table (`-x`, `a;b`, whitespace, 40 chars → field error, zero CLI calls).
- **Spoofing / elevation**: an env `GH_TOKEN` on the Kandev server would make every workspace act as that token. Mitigation: `runCLI` strips `GH_TOKEN`, `GITHUB_TOKEN`, `GH_ENTERPRISE_TOKEN`, `GITHUB_ENTERPRISE_TOKEN`. Test: AC1.1.6 / AC2.1.8 (real `runCLI` with a helper subprocess printing its env).
- **Spoofing (wrong account)**: silent fallback to the active account. Mitigation: `--user` read; old-gh fallback accepted only when `/user` login matches; otherwise `cli_account_missing`. Tests: AC2.1.4, AC2.1.6, AC3.1.2, AC3.2.4.
- **Information disclosure**: token or raw gh output in errors/logs/views. Mitigation: stderr discarded, outputs mapped to error codes, `redact`. Test: AC3.1.3 redaction table; `make check-secrets`.
- **Denial of service**: huge `gh auth status` output. Mitigation: 32 KiB cap; cut-off → `cli_unavailable`. Test: AC1.1.8.
- **Authorization**: new action `scm.providers.cli_accounts` declared with the same admin access as `use_cli` (manifest test).

## How to Run

```bash
export PATH=$HOME/.local/go/bin:$PATH
make lint            # golangci-lint incl. gosec
make check-secrets   # repo secret scan
go test -race -count=1 ./internal/scm/ ./internal/plugin/
```

## Expected Result

Lint 0 issues, secret scan clean, all security-related tests pass. No dependency vulnerability scan (team decision).
