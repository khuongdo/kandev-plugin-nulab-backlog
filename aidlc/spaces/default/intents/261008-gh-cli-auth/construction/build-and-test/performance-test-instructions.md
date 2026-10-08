# Performance Test Instructions

## Scope

No load or benchmark test is generated (Minimal strategy, no performance-validation stage in the express plan). The one load-related requirement is checked by a unit test:

- NFR4 (at most one CLI process per provider per 5 minutes in steady state): `internal/scm/cli_token_test.go` asserts that the cache returns the same token for 5 minutes on an injected clock and asks the runner again only after expiry.

## How to run

```bash
export PATH="$HOME/.local/go/bin:$PATH"
go test -race ./internal/scm/ -run 'CLI'
```

## Expected result

The CLI cache tests pass.
