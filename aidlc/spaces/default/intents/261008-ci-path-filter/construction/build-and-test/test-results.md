# Test Results — CI path filter

Run 2026-10-08, branch `feature/plugin-install-faile-9qm`, Go 1.26.8, `../kandev` at `v0.96.0` (= `.kandev-sdk-ref`).

## Build

`make check-format vet lint test coverage check-secrets build package verify-package` — **exit 0** (log: `build/bt-make.log`, local only).

- golangci-lint: `0 issues`; actionlint clean; `ci workflows: OK`
- `ci secrets: OK`
- `verifypkg: OK dist/nulab-backlog-0.4.2.tar.gz (nulab-backlog@0.4.2)`

## Unit Tests

Scoped command from `construction/code-generation/unit-test-instructions.md` (run once):

```bash
go test -race ./internal/ci/ -run 'TestClassifyChanges|TestChangesCommand|TestCheckWorkflows|TestRepositoryWorkflows'
```

8 passed, 0 failed, 0 skipped: `TestClassifyChanges`, `TestChangesCommand`, `TestCheckWorkflowsPassesACompliantPair`, `TestCheckWorkflowsFlagsViolations`, `TestCheckWorkflowsFailsOnUnreadableInput`, `TestRepositoryWorkflowsFollowThePolicy`, `TestRepositoryWorkflowsSkipAppJobsOnlyForNonAppChanges`, `TestRepositoryWorkflowsScanEveryChangeForSecrets`.

Full suite (`make test`): all Go packages `ok` with `-race`; Vitest 373 passed (373).

## Integration (Packaged-Host Contract Test)

`make contract-test KANDEV_MIN_DIR=../kandev` — **exit 0**: `ci contract: OK nulab-backlog on Kandev v0.96.0` (log: `build/bt-contract.log`, local only). Run once; this change does not touch plugin runtime code.

## Coverage

Total **92.8%** (floor 80%, exclusion `server/main.go` only, unchanged). `internal/ci` 91.2%.

## Target Verification Matrix

| Target ID | Source | Expected | Actual | Evidence | Owning Stage | Verdict |
|-----------|--------|----------|--------|----------|--------------|---------|
| TC-COV | code-generation-plan.md § Testing Contract (team Testing Posture) | Go line coverage ≥ 80% over `./internal/...` `./server/...`, exclusions only `server/main.go` | 92.8%, exclusions unchanged | `make coverage` output | build-and-test | Met |
| TC-RACE | Testing Contract (team) | Go tests run with `-race` | `go test -race` in `make test`/`make coverage` and the scoped command | Makefile `test`/`coverage` targets; run output | build-and-test | Met |
| TC-GREEN | Testing Contract (scope floor) | Existing suite stays green | All Go packages ok; Vitest 373/373 | `make test` output | build-and-test | Met |
| TC-REQ | Testing Contract (Minimal strategy) | ≥1 verifiable test per requirement, happy-path per component | Tests for FR1–FR4, NFR1, NFR2, NFR4, NFR5; FR5.1 is a manual ruleset step (Deferred) | `code-generation/traceability.json`, scoped test run | build-and-test | Met |
| TC-CONTRACT | Testing Contract (team) | Packaged plugin installs and runs on Kandev `min_kandev_version` | `ci contract: OK … v0.96.0` | `make contract-test` output | build-and-test | Met |
| TC-LINT | Team Code Style | Lint/format/vet clean, workflow policy OK | `0 issues`, `ci workflows: OK`, format clean | `make check-format vet lint` output | build-and-test | Met |

## Known Limitation (not a quality target)

FR3 acceptance case "a records-only pull request adding a credential under `aidlc/` fails `secret-scan`" does not hold: the scanner does not read `aidlc/`. Accepted at the Code Generation gate.
