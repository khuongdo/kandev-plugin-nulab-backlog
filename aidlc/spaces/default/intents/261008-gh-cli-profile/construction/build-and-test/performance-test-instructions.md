# Performance Test Instructions — 261008-gh-cli-profile

## Scope

No load or benchmark tests: the change adds one CLI call path and a per-login token cache; the plugin serves a single user's browser actions.

## Targets and How They Are Checked

- NFR2: gh calls bounded by the existing 10 s timeout; tokens cached 5 min per provider+login. Checked by unit tests with an injected clock in `internal/scm/cli_token_test.go` (cache hit within TTL, refresh after TTL, no cross-login reuse).

```bash
export PATH=$HOME/.local/go/bin:$PATH
go test -race -count=1 -run 'CLI|Cache|TTL' ./internal/scm/
```

## Expected Result

All pass; no real sleeps.
