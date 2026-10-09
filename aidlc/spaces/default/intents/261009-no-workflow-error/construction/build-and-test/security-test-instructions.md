# Security Test Instructions — 261009-no-workflow-error

## Scope

Security review of the change (security engineer view):

- **New capability `api_read: workflows`** — read-only, least privilege for FR3; the action returns only a boolean, no workflow names or ids (no information disclosure beyond "has a workflow").
- **New action `workflows.status`** — `access: authenticated`, `scope: workspace`, `max_body_bytes: 8192`; the workspace comes from the verified action context, not the body.
- **Error text** — host errors map to the generic `internal` code; no host or Backlog text and no secrets reach the UI (NFR4, project rule "ALWAYS redact API keys and tokens").
- No new dependency (team rule: no dependency scan is added).

## How to Run

```bash
make lint                                                    # golangci-lint with gosec
go test -race ./internal/plugin/ -run 'WorkflowsStatus|Redact|Leak'   # host-error leak test and existing redaction tests
go test -race ./internal/redact/
```

Expected: gosec 0 issues; leak and redaction tests pass.
