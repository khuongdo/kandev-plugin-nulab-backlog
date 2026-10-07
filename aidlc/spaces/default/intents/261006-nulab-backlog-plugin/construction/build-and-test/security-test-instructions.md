# Security Test Instructions

Security NFRs exist (NFR3, NFR3.1–NFR3.10, NFR4, NFR4.1–NFR4.2, project Mandated/Forbidden rules). The team decided not to add a dependency vulnerability scan (team.md, Q4). SAST is `gosec` inside golangci-lint.

## Automated checks (local)

| Area | Check | Command |
|------|-------|---------|
| SAST | golangci-lint with `gosec`, `go vet`, `tsc --noEmit` strict, ESLint | `make vet lint` |
| Secret scanning | repository scan; fake-prefix test secrets | `make check-secrets`; `go test -race ./internal/ci/ -run TestScanSecrets`; `go test -race ./internal/testutil/` |
| No secret leakage (logs, errors, UI replies, events, task descriptions, candidates) | leak tests per unit | `go test -race ./internal/plugin/ -run 'TestResponsesAndLogsNeverContainTheKey\|FailureLogHasNoSecret'`; `go test -race ./internal/connection/ -run 'NeverLeaks\|NoSecretInLogs\|PasswordNeverLeaks'`; `go test -race ./internal/git/ -run TestU4_Leak_`; `go test -race ./internal/issues/ -run Leak` |
| Redaction / no URL query in logs | redact package and client log tests | `go test -race ./internal/redact/`; `go test -race ./internal/backlog/ -run 'HideTheURLAndKey\|WithoutQuery\|HideEverySecret'` |
| Transport: https-only, header auth, no redirects, bare host, 1 MiB limit, no echoed error bodies | client tests | `go test -race ./internal/backlog/ -run 'HTTPS\|Redirect\|BareHost\|OneMiB\|NeverEchoTheBody'` |
| Space address allowlist (`https` + `backlog.com`/`backlog.jp`/`backlogtool.com`) | address tests | `go test -race ./internal/connection/ -run 'TestParseSpaceAddress\|TestValidateAPIKey'` |
| Authorization: action access levels, verified workspace/task only | manifest and action tests | `go test -race ./internal/plugin/ -run 'TestManifestActionsAndAccess\|Manifest\|VerifiedWorkspace'` |
| Fail-closed switch | switch tests | `go test -race ./internal/plugin/ ./internal/connection/ ./internal/git/ -run 'IntegrationDisabled\|Switch'` |
| Path injection via issue refs | `ValidIssueRef` tests | `go test -race ./internal/backlog/ -run TestU3_` |
| Logo / bundle: no external assets | brand test, package verifier | `cd ui && npx vitest run src/brand`; `make package verify-package` |
| CI hardening (SHA-pinned actions, least privilege, no `pull_request_target`) | workflow policy | `go run ./cmd/ci workflows -dir .github/workflows` |

## Known findings carried from Code Generation reviews (not fixed yet)

| Ref | Severity | Finding |
|-----|----------|---------|
| U2 R-01 | Major | OAuth state not bound to the browser that started sign-in (cross-user token capture by a workspace admin). |
| U2 R-02 | Major | Project picker keeps the old space's selection after an API-key replace to another space. |
| U5 R-01 | Minor | Main-only release rule lives in the workflow at the tagged commit; tag creation not restricted. |
| U5 R-02 | Minor | Secret scanner misses unquoted `password:`, `api_key=`/`apikey=`, and testutil/UI fixtures. |
| U5 R-03 | Minor | Tag name reaches Make shell recipes unvalidated. |
| U5 R-05 | Minor | `actions/checkout` keeps `persist-credentials: true`. |
| U1 R-01 | Minor | Host allowlist checked only at entry points, not inside `backlog.Client.send`. |

## Requirement/test discrepancy

- NFR3.2 requires scanning for any **4-character** substring of the key; `testutil.AssertNoLeak` checks **8-character** windows (NFR4.1 design). Either the requirement or the tests must change.

## Manual

- NFR4.2: manual-check records contain only the space domain (no key, no URL query).
- Live 403 for non-admin callers on a deployed Kandev (the contract test exercises the packaged host but not a non-admin user).
