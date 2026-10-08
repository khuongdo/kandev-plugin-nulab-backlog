# Build and Test Summary — CI path filter

## Build Status

**Success.** All standard targets pass locally (`make check-format vet lint test coverage check-secrets build package verify-package`), and the packaged-host contract test passes on Kandev v0.96.0. Prerequisites: Go 1.26.x, Node from `.nvmrc`, `../kandev` at `.kandev-sdk-ref` (see `build-instructions.md`).

## Test Type Inventory

- Unit tests (Code Generation, Minimal strategy): `TestClassifyChanges`, `TestChangesCommand`, two new workflow tests over the real workflow files.
- Existing suite: all Go packages with `-race`, Vitest 373 tests.
- Integration: existing packaged-host contract test, run once.
- No separate integration, performance or security suites (Minimal strategy); the instruction files explain the scope.

## Coverage Expectations

Stage-level (zero-Unit): 80% floor over `./internal/...` and `./server/...`; actual 92.8%; `internal/ci` 91.2%.

## Target Verification Matrix

| Target ID | Source | Expected | Actual | Evidence | Owning Stage | Verdict |
|-----------|--------|----------|--------|----------|--------------|---------|
| TC-COV | Testing Contract (team) | ≥ 80% line coverage, exclusions only `server/main.go` | 92.8% | `make coverage` | build-and-test | Met |
| TC-RACE | Testing Contract (team) | Go tests with `-race` | yes | Makefile + run output | build-and-test | Met |
| TC-GREEN | Testing Contract (scope floor) | Existing suite green | all pass | `make test` | build-and-test | Met |
| TC-REQ | Testing Contract (Minimal) | ≥1 test per requirement | met; FR5.1 manual | `traceability.json` | build-and-test | Met |
| TC-CONTRACT | Testing Contract (team) | Contract test on min Kandev version passes | `ci contract: OK … v0.96.0` | `make contract-test` | build-and-test | Met |
| TC-LINT | Team Code Style | Lint, format, vet, workflow policy clean | clean | `make check-format vet lint` | build-and-test | Met |

## Readiness

- Build-ready: yes.
- Test-ready: yes.
- Deployment-ready: yes. The real proof is the pull request run on GitHub Actions: this change is app, so all three checks should run; after merge, a records-only pull request should skip `checks`/`packaged-host-contract` and still be mergeable.

## Known Limitations / Outstanding Items

- FR5.1 (manual): add `secret-scan` to ruleset 24580280 after it has reported once (`gh api` command in `code-generation/code-summary.md`).
- FR3.1 limitation: the credential scanner does not read `aidlc/` or most of `docs/` (accepted at the Code Generation gate; follow-up work).
- Release workflow unchanged by design (FR4.1).
