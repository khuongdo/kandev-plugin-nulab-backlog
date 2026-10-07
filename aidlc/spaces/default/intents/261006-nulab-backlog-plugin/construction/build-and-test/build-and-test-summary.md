# Build and Test Summary

Scope `feature`, test strategy **Standard**, depth Minimal. Units: walking-skeleton (U1), ci-release (U5), connection (U2), git-pr (U4), issues (U3). Code: `main` at `4b0a090`.

## Build status and prerequisites

- Build: **success** from `make clean` (format, vet, lint+gosec, tests with `-race`, coverage, secret scan, 5-platform build, package, package verification). See `build-instructions.md`.
- Packaged-host contract test on Kandev v0.96.0: **pass**.
- Prerequisites: Go 1.26 (`GOTOOLCHAIN=local`), Node 20+, gcc, `../kandev` at v0.96.0, `../kandev-min` at v0.96.0 for the contract test.

## Test type inventory

| Type | Generated / location | Status |
|------|----------------------|--------|
| Unit (Go, Vitest) | per unit, `unit-test-instructions.md` | 529 Go + 509 subtests, 224 UI tests pass |
| Integration (key boundaries) | `integration-test-instructions.md` | pass |
| Contract (packaged plugin on real Kandev v0.96.0) | `make contract-test` | pass |
| Performance (virtual-time budgets) | `performance-test-instructions.md` | automated budgets pass; real-network timing manual |
| Security (gosec, secret scan, leak, transport, authz) | `security-test-instructions.md` | automated pass; 1 requirement/test discrepancy |
| Accessibility (axe, focus) | Vitest | pass (axe `color-contrast` disabled in jsdom; manual contrast check pending) |
| Manual end-to-end on a real Backlog space | `docs/manual-checks/TEMPLATE.md` | **not done yet** (no skeleton record exists) |

## Coverage expectations per unit

| Unit | Packages | Expected | Actual |
|------|----------|----------|--------|
| U1 walking-skeleton | backlog, connection, plugin, redact, pkgverify | ≥ 80% | 93–97% |
| U5 ci-release | ci | ≥ 80% | 91.0% |
| U2 connection | backlog, connection, plugin | ≥ 80% | 94–96% |
| U4 git-pr | git (+ shared) | ≥ 80%, aim 85% | 90.8% |
| U3 issues | issues (+ shared) | ≥ 80%, aim 85% | 91.3% |
| Total | `./internal/...`, `./server/...` | ≥ 80% | 92.9% |

## Target Verification Matrix

Sources: NR = `construction/walking-skeleton/nfr-requirements/`, ND = `construction/walking-skeleton/nfr-design/`, REQ = `inception/requirements-analysis/requirements.md`, ST = `inception/user-stories/stories.md`, TM = team.md Testing Posture / Deployment / Code Style, TC = Testing Contract in every unit's `code-generation-plan.md`. Evidence logs: `test-results.md`.

| Target ID | Source | Expected | Actual | Evidence | Owning Stage | Verdict |
|-----------|--------|----------|--------|----------|--------------|---------|
| T-PERF-01 | REQ#NFR1; ST#AC8.1.1 | 20-row list ≤ 2.5 s with 300 ms fake latency | within budget | `TestU3_List_MeetsTheLatencyBudget` pass | build-and-test | Met |
| T-PERF-02 | REQ#NFR1; ST#AC8.1.2 | p95 ≤ 3 s on a real space (19/20 opens) | not measured | needs real space + deployed Kandev; TEMPLATE.md step 25 | performance-validation | Unverified |
| T-PERF-03 | NR/performance#NFR1.1, NFR1.3 | get/set_enabled p95 ≤ 500 ms over 100 calls | virtual-time p95: get 400 ms, set_enabled 200 ms (100 ms/store op); 4 store reads vs 3 designed | TestConnectionGetP95Under500ms, TestSetEnabledP95Under500ms pass (Run 2) | build-and-test | Met |
| T-PERF-04 | NR/performance#NFR1.2 | Connect p95 < 3 s at 2 s Backlog delay | within budget | `TestConnectBudget` pass | build-and-test | Met |
| T-PERF-05 | NR/performance#NFR1.4; ND/performance | 12 s deadline, total ≤ 14 s | as designed | `TestNewServiceUsesTheDesignBudget`, `TestConnectBudget`, `TestGuardRunsInsideTheConnectDeadline` pass | build-and-test | Met |
| T-PERF-06 | NR#NFR5.1; ST#AC8.1.3 | timeout ≤ 10 s as `unreachable`, others served | as specified | `TestMyselfTimeoutIsUnreachable`, `TestU3_Issue_TimeoutDoesNotBlockOthers` pass | build-and-test | Met |
| T-RATE-01 | NR/performance#NFR2.1; ND/performance | Connect on 429 returns `rate_limited` at once (< 1 s, no retry) | Connect returns rate_limited after 1 request (<1 s); other callers keep retries | TestNoRetryReturns429AtOnce, TestConnectAsksForNoRetryOn429 pass (Run 2) | build-and-test | Met |
| T-RATE-02 | REQ#NFR2; ST#AC8.4.1 | wait for reset / Retry-After / 60 s, then retry | as specified | limiter tests pass | build-and-test | Met |
| T-RATE-03 | ST#AC8.4.2 | max 3 retries; cancel returns `context.Canceled` | as specified | `TestRateLimitStopsAfterThreeRetries`, `…CancelDuringTheWait…` pass | build-and-test | Met |
| T-RATE-04 | REQ#NFR2; ST#AC8.4.3 | Search/Update never parallel, ≥ 1 s apart | as specified under `-race` | queue tests pass | build-and-test | Met |
| T-RATE-05 | ST#AC8.4.4 | "Retrying in N seconds", announced once | as specified | `issues-state.test.ts` pass | build-and-test | Met |
| T-SEC-01 | NR/security#NFR3.1 | key only in secret store | 0 copies in state | store/connect tests pass | build-and-test | Met |
| T-SEC-02 | NR/security#NFR3.2, NFR3.3; ND#NFR4.1; ST#AC7.3.2 | no **4-character** substring of the key in UI/logs/errors | 4-character windows over the 32-char random part | TestAssertNoLeak_Catches4CharWindow + all leak tests pass (Run 2) | build-and-test | Met |
| T-SEC-03 | NR#NFR3.3; ST#AC7.3.3; NR#NFR11.5 | no query/key/name in logs or errors | as specified | redact + client log tests pass | build-and-test | Met |
| T-SEC-04 | NR#NFR3.4; REQ#NFR3 | https only, header auth, no redirects, bare host | as specified | client tests pass | build-and-test | Met |
| T-SEC-05 | REQ#FR1.2; project.md Mandated; NR#NFR3.7 | only https `backlog.com`/`.jp`/`backlogtool.com`; key 1–256 printable | as specified | address/key tests pass | build-and-test | Met |
| T-SEC-06 | NR#NFR3.5, NFR3.8 | admin/authenticated access; non-admin gets 403 live | manifest + verified-workspace tests pass; live non-admin 403 not exercised | manifest tests; contract test does not use a non-admin user | deployment-execution | Unverified |
| T-SEC-07 | NR#NFR3.9 | switch off: 0 secret reads, 0 Backlog calls, fail closed | as specified | switch tests (plugin, connection, git, issues) pass | build-and-test | Met |
| T-SEC-08 | NR#NFR3.6 | no Backlog error body in responses/logs | as specified | `ErrorsNeverEchoTheBody` pass | build-and-test | Met |
| T-SEC-09 | NR#NFR3.10 | inline safe SVG; no runtime Nulab asset | as specified | brand test, `verifypkg` pass | build-and-test | Met |
| T-SEC-10 | REQ#NFR4; ND#NFR4.1; ST#AC7.3.4; project.md Forbidden | 0 credential-shaped strings | `ci secrets: OK` | `make check-secrets`, `TestScanSecrets` pass (scanner gaps: U5 R-02) | build-and-test | Met |
| T-SEC-11 | NR#NFR4.2 | manual record holds only space domain | no record yet | `docs/manual-checks/` has only TEMPLATE.md | deployment-execution | Unverified |
| T-REL-01 | NR#NFR5.2 | ≤ 1 MiB read; 2 MiB → `unreachable` | as specified | body-limit tests pass | build-and-test | Met |
| T-REL-02 | NR#NFR5.3 | every failure → defined code; cancel unchanged; panic → internal | as specified | mapping tests pass | build-and-test | Met |
| T-REL-03 | NR#NFR5.4; NR#NFR11.4 | rollback on fresh 2 s ctx; inconsistency logged once | as specified | rollback tests pass | build-and-test | Met |
| T-REL-04 | NR#NFR5.8, NFR5.9 | failures never change existing connection | as specified | replace/recheck/switch tests pass | build-and-test | Met |
| T-REL-05 | REQ#FR1.5 | 5 concurrent callers → 1 refresh; stale epoch writes nothing | as specified | token tests pass | build-and-test | Met |
| T-REL-06 | REQ#FR4.2 | poll default 5 min, min 1 | as specified | `TestU3_Settings_*`, `ValidatePollMinutes` pass | build-and-test | Met |
| T-REL-07 | REQ#FR6.2 | one task per PR; no duplicates across restarts | as specified (Kandev-shaped fakes) | `TestU4_Watcher_RestartCreatesNoDuplicates`, `CrashAfterCreateNoDuplicate` pass | build-and-test | Met |
| T-REL-08 | REQ#FR7.2; TM#Deployment; ST#AC7.5.1–7.5.3 | tag rules, Release with package/checksums/attestation | preflight logic passes; no Release yet | `TestParseTag`, `TestCheckRelease`, `TestRunPreflight` pass | deployment-execution | Unverified |
| T-COMPAT-01 | REQ#NFR6; NR#NFR6.1; ST#AC7.4.1–7.4.2 | `min_kandev_version` 0.96.0; runs on exactly v0.96.0 | min_kandev_version 0.96.0; packaged plugin installs and runs on exactly v0.96.0; contract test 10/10 after the Host-readiness fix | test-results.md Run 3 (10 consecutive `make contract-test` passes) | build-and-test | Met |
| T-PLAT-01 | REQ#NFR7, FR7.1; NR#NFR7.1; ST#AC7.1.1, AC7.1.3 | 5 executables; verify passes; tampered checksum fails | as specified | build log, pkgverify tests, `verifypkg: OK` | build-and-test | Met |
| T-SCAL-01 | NR#NFR5.5–5.7, NFR5.10 | workspaces independent; parallel under `-race` | as specified | workspace tests pass | build-and-test | Met |
| T-OBS-01 | NR#NFR11.1 | JSON log lines with required fields | as specified | `TestLogLinesCarryTheRequiredFields` pass | build-and-test | Met |
| T-OBS-02 | NR#NFR11.2, 11.3, 11.6; ST#AC8.3.1–8.3.2 | exactly one outcome event per action/cycle | as specified | log-count tests (connection, plugin, issues, git) pass | build-and-test | Met |
| T-OBS-03 | NR#NFR11 (collection) | Kandev collects plugin stderr logs | not observed | needs deployed Kandev | observability-setup | Unverified |
| T-TEST-01 | TM; REQ#NFR8; NR#NFR8.1; ST#AC7.3.1 | ≥ 80% line coverage, only `server/main.go` excluded | 92.9% | `make coverage` | build-and-test | Met |
| T-TEST-02 | TM | all Go tests with `-race`; all UI tests pass | 529+509 Go, 224 UI pass | `make test`, fresh `-count=1` run | build-and-test | Met |
| T-TEST-03 | TC#obligations | 5–8 tests per component; integration at key boundaries | ≥ 5 per component; integration per `integration-test-instructions.md` | test counts per package | build-and-test | Met |
| T-CODE-01 | TM#Code Style; ST#AC7.3.1 | all linters clean; `go mod tidy` no diff | clean | `make check-format vet lint`; tidy no diff | build-and-test | Met |
| T-CI-01 | TM#Deployment | SHA-pinned actions, least privilege, no `pull_request_target` | as specified | `ci workflows` pass | build-and-test | Met |
| T-ACC-01 | REQ#NFR9; NR#NFR9.1; ST#AC8.2.2–8.2.3 | 0 axe A/AA violations; focus trap; Esc returns focus | as specified (contrast rule off in jsdom) | Vitest axe/focus tests pass | build-and-test | Met |
| T-ACC-02 | ST#AC8.2.1, AC8.2.4, AC5.4.2 | keyboard-only; contrast ≥ 4.5:1; 320 px | not checked | manual (TEMPLATE.md steps 17, 26) | deployment-execution | Unverified |
| T-L10N-01 | REQ#NFR10; NR#NFR10.1; ST#AC8.5.1–8.5.2 | no literal strings; unknown locale → English | as specified | settings/page/i18n tests pass | build-and-test | Met |
| T-MKT-01 | REQ#FR7.3; ST#AC7.6.1–7.6.2 | catalogue entry valid; maintainers accept | entry fields valid; not submitted | marketplace tests pass | deployment-execution | Unverified |
| T-MAN-01 | TM#Testing Posture, Walking Skeleton; ST#AC7.1.2, AC7.2.1 | two manual real-space records (skeleton, pre-release) | **none recorded** | `docs/manual-checks/` has only TEMPLATE.md | deployment-execution | Unverified |

Totals (46 targets, after Run 3): **38 Met, 0 Not Met, 8 Unverified** (all eight need a real Backlog space, a deployed Kandev, GitHub, or a manual check).

## Readiness assessment

- **Build-ready:** yes.
- **Test-ready:** yes — all automated suites pass with 92.9% coverage.
- **Deployment-ready:** **no** — 8 targets Unverified (manual real-space checks, release on GitHub, marketplace, Kandev plugin log collection). Every automated target is Met; the Run 1 and Run 2 failures are fixed.

## Known limitations and outstanding items

1. Run 1 Not Met T-RATE-01 and T-SEC-02 were fixed in Loop-back 1. Run 2 found T-COMPAT-01 Not Met: actions arriving before Kandev injects the Host answer 500.
2. The skeleton's manual check against a real Backlog space (AC7.2.1) and the manual install (AC7.1.2) have not been recorded, although the team's Walking Skeleton rule requires them.
3. Open Code Generation review findings: U2 R-01 (OAuth not bound to the browser, Major), U2 R-02 (stale project picker, Major); Minors U1 R-01/R-02, U5 R-01–R-05, U2 R-03–R-11, U4 R-04–R-07, U3 R-04–R-07; U3 R-02 accepted risk.
4. Unverified risks against real services: Backlog accepting the `Backlog-API-Key` header, Git HTTPS probe and PR statuses 2/3, keyword matching issue keys, real `task.deleted` payload.
5. Contract amendments C1/C2/C3/C5/C8 still to apply upstream.
6. `make coverage` leaves `coverage.out` in the repo root; inside an AI-DLC Code Generation attempt it blocks the stage gate. Consider writing it outside the repo.
7. Branch protection on `main` should require `checks` and `packaged-host-contract`, and block force-push.
