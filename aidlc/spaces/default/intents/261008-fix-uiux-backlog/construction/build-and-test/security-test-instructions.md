# Security Test Instructions - Fix UIUX

## Scope at Minimal Strategy

No new security suite is generated. The change adds one piece of Backlog content (the issue summary) to stored links and to the UI, so the checks focus on information disclosure (STRIDE "I"):

- NFR1: the summary never appears in error messages, error responses or logs (`internal/issues/leak_test.go`, `TestNFR1_Leak_SummaryNeverInErrorsOrLogs`). Existing leak tests keep covering API keys and tokens (project Mandated rule).
- The badge link keeps the https and Backlog-host check (`badgeHref`: `backlog.com`, `backlog.jp`, `backlogtool.com`) and opens with `rel="noopener noreferrer"` (FR1.5, project Mandated rule on space addresses).
- The summary is rendered as text through host React (no HTML injection).
- Static analysis: `golangci-lint` with `gosec`, `go vet`, ESLint, `tsc` strict.
- Secrets: `make check-secrets`.

## How to Run

```bash
export PATH=$HOME/.local/go/bin:$PATH
go test -race ./internal/issues/ -run 'Leak'
make lint check-secrets
```

## Expected Result

All pass; golangci-lint reports 0 issues. No dependency vulnerability scan is run (team decision).
