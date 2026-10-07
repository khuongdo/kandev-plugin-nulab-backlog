## Review

**Verdict:** READY
**Reviewer:** aidlc-architecture-reviewer-agent
**Iteration:** 1

Confirmation review of unit `walking-skeleton` (U1) after a stage restart that only reset the review budget. No code changed since the previous attempt's final U1 review (READY, iteration 2).

### Validation

- `guarded()` in `internal/plugin/runtime.go` is `action != actionGet && action != actionSetEnabled`; only `connection.get` and `connection.set_enabled` are exempt.
- `internal/backlog/client.go`: the `httpsOnly` RoundTripper rejects non-https requests; `CheckRedirect` returns `http.ErrUseLastResponse`; `maxBody` is `1 << 20` via `io.LimitReader`.
- `NoRetry` has two production callers: Connect (`internal/connection/service.go:242`) and the OAuth sign-in probe (`internal/connection/oauth.go:263`), both one-shot connect checks.
- `go vet ./...` clean; `go test -race -cover ./internal/... ./server/...` passes (backlog 96.0%, ci 91.0%, connection 94.6%, git 90.8%, issues 91.3%, pkgverify 93.2%, plugin 93.7%, redact 97.4%, testutil 88.0%; `server` 0.0% is the permitted main-wiring exclusion).
- All commands were read-only.

### Findings

| ID | Severity | Location | Finding | Required action | Status |
|----|----------|----------|---------|-----------------|--------|
| R-01 | Minor | `internal/connection` `connection.get` read path vs U1 performance-design | `connection.get` does 4 store reads; performance-design says 3. | Align the design doc with the code or fold the reads. | Unresolved |
| R-02 | Minor | `internal/connection` recheck path | No comment explains why the recheck keeps rate-limit retries. | Add a one-line comment. | Unresolved |
| R-03 | Minor | `internal/backlog/client.go` / entry points | The host allowlist is checked only at the entry points; https-only and no-redirect limit the risk. | Re-check the host in the client or record it as accepted. | Unresolved |
| R-04 | Minor | `internal/plugin/runtime.go` | Background workers start before Host injection; a nil Host fails safe. | Start after Host injection or document the ordering. | Unresolved |
