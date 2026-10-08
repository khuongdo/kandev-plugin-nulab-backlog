# Build and Test Summary — 261007-backlog-panel-retouch

## Overall Status

Build: **success**. All tests: **pass** (Vitest 373/373, scoped 147/147, Go `-race` all ok, contract test 10/10 on Kandev v0.96.0). Coverage: Go 92.8% (floor 80%). Package `nulab-backlog-0.4.1.tar.gz` verified. Re-run after Loop-back 2 (rebased onto v0.4.0; provider pull-request list brought to the GitHub-style toolbar and task display).

Prerequisites: complete Go 1.26 toolchain with `covdata` and `GOTOOLCHAIN=local`; `../kandev` at `v0.96.0` (see build-instructions.md).

## Test Type Inventory

| Type | Generated / run | File |
|---|---|---|
| Unit (UI, Vitest) | Run (from Code Generation) | ../code-generation/unit-test-instructions.md |
| Go regression (`-race`, coverage) | Run | build-instructions.md |
| Packaged-host contract (integration) | Run ×10 | integration-test-instructions.md |
| Manual real-host UI check | Defined; owned by Deployment Execution | integration-test-instructions.md |
| Performance | Not applicable (no target) | performance-test-instructions.md |
| Security regression | Run (lint/gosec, redaction, badge link) | security-test-instructions.md |

## Coverage Expectations

Zero-Unit refactor: Go floor 80% (actual 92.8%); UI suite green with a test per requirement (cross-unit-traceability.md: PASS).

## Target Verification Matrix

| Target ID | Source | Expected | Actual | Evidence | Owning Stage | Verdict |
|---|---|---|---|---|---|---|
| TC-SUITE-GREEN | Testing Contract | Existing suite green | 373/373; Go ok | test-results.md | build-and-test | Met |
| TC-MIN-PER-REQ | Testing Contract | ≥1 test per requirement | all FR/NFR mapped | cross-unit-traceability.md | build-and-test | Met |
| TC-GO-COVERAGE-80 | Testing Contract | ≥ 80% | 92.8% | test-results.md | build-and-test | Met |
| TC-GO-RACE | Testing Contract | `-race` pass | pass | test-results.md | build-and-test | Met |
| TC-CONTRACT-10 | Testing Contract | 10/10 on v0.96.0 | 10/10 | test-results.md | build-and-test | Met |
| TC-NO-ROOT-COVERAGE | Testing Contract | no root `coverage.out` | absent | test-results.md | build-and-test | Met |
| NFR1 | requirements.md | host components, no plugin CSS | pass | test-results.md | build-and-test | Met |
| NFR2-MIN-VERSION | requirements.md | runs on v0.96.0 | 10/10 | test-results.md | build-and-test | Met |
| NFR2-UI-IN-HOST | requirements.md | UI renders in real host | not observed | integration-test-instructions.md | deployment-execution | Unverified |
| NFR3 | requirements.md | a11y and safe links | pass | test-results.md | build-and-test | Met |
| NFR4 | requirements.md | quality gates green | pass | test-results.md | build-and-test | Met |
| NFR5 | requirements.md | no leakage; locked behaviour | pass | test-results.md | build-and-test | Met |

## Readiness Assessment

- Build-ready: **yes**.
- Test-ready: **yes** (all automated checks green).
- Deployment-ready: **conditional** — NFR2-UI-IN-HOST is `Unverified` until the manual real-host UI check runs during Deployment Execution.

## Known Limitations and Outstanding Items

- Manual real-host UI check (NFR2-UI-IN-HOST) pending; checklist in integration-test-instructions.md.
- Local Go toolchain: the auto-switched Go 1.26.0 lacks `covdata`; fix the local toolchain or keep using a complete one with `GOTOOLCHAIN=local`.
- Code review after Loop-back 2 (code-generation review-03, READY): adds R-06 (inherited from v0.4.0: choosing "All repositories" in the provider list sends no request and leaves the previous page) and R-07 (README 0.4.1 note omits the at-least-one-status rule on the provider list). Earlier Minor findings (code-generation review-02, all Unresolved Minor): R-01 touch input may still reload twice (final result correct); R-02 fake TaskRowIndicator test ids invented; R-03 plan overstates the contract test as UI proof; R-04 toolbar keeps the previous total after a failed load or workspace switch; R-05 `aria-disabled` checkbox may lack a visual disabled cue. Check these in the manual real-host UI check.
