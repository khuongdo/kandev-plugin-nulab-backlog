# Security Test Instructions — github-parity-actions

## Scope (security engineer view)

The new attack surface consists of 7 `authenticated` actions that store user text: prompt templates and saved queries. These are added to a per-workspace plugin state.

| Threat (STRIDE) | Control | Check |
|---|---|---|
| Tampering / DoS: oversized input | Label and hint at most 100 runes, prompt at most 4000, at most 20 actions per kind, at most 50 saved queries, `max_body_bytes` 256 KiB on `issues.quick_actions.save` (16 KiB on the others) | `internal/issues/quick_actions_test.go`, `queries_test.go` |
| Information disclosure: secrets in responses or logs | The redaction and leak tests cover the new actions | `internal/plugin/actions_watch_test.go`, `internal/*/leak_test.go` |
| Elevation of privilege | The access level of every new action is asserted (`authenticated`, the same as `git.queries.*`); link actions take the task from the verified context | `internal/plugin/manifest_test.go` |
| Injection: `{{url}}` / `{{title}}` interpolation | Plain string replacement into the Kandev create-task dialog; no HTML is rendered from prompts | `ui/src/page/quick-actions.test.ts` |
| Static analysis | golangci-lint with gosec in `make lint` | `make lint` -> `0 issues.` |

A dependency vulnerability scan is excluded by team decision (`team.md`, Testing Posture).

## How to Run

```bash
export PATH=$HOME/.local/go/bin:$PATH
make lint
go test -race ./internal/plugin/ -run 'Manifest|Leak|Redact|Watch'
go test -race ./internal/issues/ -run 'QuickAction|IssueQuer|Leak'
```
