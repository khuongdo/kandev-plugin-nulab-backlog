# Security Test Instructions — CI path filter

## Scope

Test Strategy is Minimal; no separate security suite is generated. The security-relevant checks for this change run inside the standard targets:

| Check | Command | Covers |
|-------|---------|--------|
| Workflow policy (actions pinned to full SHAs, top-level `permissions: contents: read`, no `pull_request_target`, write only in release `publish`) | `make lint` (actionlint + `go run ./cmd/ci workflows -dir .github/workflows`) | NFR1 |
| Event values reach shell only through `env:` | Review of `.github/workflows/ci.yml` job `changes` | Script-injection hardening |
| `changes` refuses non-SHA base/head before calling git | `TestChangesCommand` | Argument-injection hardening |
| Credential scan | `make check-secrets` (and `make -o check-sdk check-secrets` as in `secret-scan`) | NFR5 |
| No credentials in test data | `make check-secrets` | NFR5 |

## Known Limitation

The credential scanner's scope (`inScanScope` in `internal/ci/secrets.go`) does not read `aidlc/` or `docs/` outside `docs/manual-checks/`. `secret-scan` runs on every change, but it does not catch a credential added only under `aidlc/`. Accepted at the Code Generation gate; widening the scope is follow-up work.
