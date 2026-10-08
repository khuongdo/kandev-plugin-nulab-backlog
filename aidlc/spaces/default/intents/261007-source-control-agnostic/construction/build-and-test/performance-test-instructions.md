# Performance Test Instructions — Multi-provider source control

## Scope

The only measurable performance target is NFR8: the first page of a provider's PR list appears within 3 seconds when the provider answers within 1 second, measured against a fake provider. Load testing is out of scope (Minimal strategy; no Performance Validation stage in the express plan).

## How to Run

```bash
export PATH=$HOME/.local/go/bin:$PATH
go test -race -count=1 -run 'NFR8|FirstPage' -v ./internal/scm/...
```

The test in `internal/scm/service_test.go` (marked `// NFR8`) uses a fake client that answers within the 1 s budget and asserts the first page returns within 3 s.

## Related Limits (NFR3, NFR4)

Rate-limit handling and response-size limits are verified by `internal/scm/httpx_test.go` (429 / `Retry-After`, GitHub `X-RateLimit-*`, `io.LimitReader`) with an injected clock, so no real waiting happens.
