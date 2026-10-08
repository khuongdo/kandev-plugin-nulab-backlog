# Security Test Instructions

## Scope

Security-relevant requirements: NFR1 (no secrets in logs, errors or UI responses) and FR2.2 / FR1.4 (only admins switch the service; actions for a non-active service are refused). No dependency vulnerability scan is added (team decision).

## Checks

- `gosec` via `make lint` (golangci-lint) - must report 0 issues.
- Secret-leak test in `internal/plugin/actions_scm_test.go`, extended to `scm.active.set`: no token value appears in any action response or error.
- Admin-only: `scm.active.set` is in the admin action list in `actions_scm_test.go` and declared `access: admin` in `manifest.yaml`.
- Refusal path: non-active provider actions return 409 `service_inactive`; the Backlog Git credential is refused while an external service is active (`ResolveGitCredential`).

## How to Run

```bash
make lint
go test -race -count=1 ./internal/plugin/ ./internal/scm/
```

## STRIDE Notes (security engineer)

- Elevation of privilege: switching the service is admin-only; members get a read-only view and the host enforces `access: admin`.
- Information disclosure: `ProviderView` stays non-secret; error bodies carry only the service name (`activeService`).
- Tampering: `service` input is validated against the four known values; unknown or empty values return a validation error.
- Denial of service: request body limited by `max_body_bytes: 16384` like other `scm.*` actions.
