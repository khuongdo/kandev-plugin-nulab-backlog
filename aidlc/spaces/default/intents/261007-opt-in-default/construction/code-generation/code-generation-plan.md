# Code Generation Plan — Opt-in by default after installation

## Scope

Zero-Unit bugfix (no Units Generation). Inputs: `inception/requirements-analysis/requirements.md` (FR1–FR6, NFR1–NFR3) and the code KB `aidlc/spaces/default/codekb/kandev-plugin-nulab-backlog/` (change points D1–D6 in `code-quality-assessment.md`). Brownfield: every file is modified in place.

Blast radius: one default in `internal/connection/store.go` (high fan-out: every `RequireEnabled` caller, i.e. guarded actions, RPCs, webhook and workers), two UI fallbacks, the contract test, README, and test fixtures that implicitly relied on the ON default (about 112 Go tests in `internal/connection` and `internal/plugin`, measured by the developer scan). Test baseline before the change: Go `go test -race ./...` all pass (1159 tests), Vitest 286/286.

## Testing Contract

```json
{
  "version": 1,
  "methodology": "tdd",
  "source": "team",
  "ordering": "For each testable layer, we write a failing test first, then write the minimum code to make it pass, then refactor while the test stays green, before moving to the next behaviour or layer.",
  "scope": "bugfix",
  "test_strategy": "minimal",
  "project_type": "brownfield",
  "applicable_notes": [
    {
      "layer": "org",
      "text": "We treat tests as a first-class deliverable in every Bolt. The specific\nmethodology (TDD, BDD, ATDD, or classic test-after) is affirmed at\npractices-discovery and recorded in `team.md` under this heading with explicit\n`Methodology` and `Ordering` fields; Code Generation resolves those fields\nindependently from coverage, tooling, and scope notes.\n\nWhen no posture has been affirmed, our default per scope is:\n- **Methodology**: test-after\n- **Ordering**: implement each applicable testable layer, then write and run\n  that layer's tests.\n- `mvp`, `enterprise`, `feature`, `infra`, `classic` add an 80% line-coverage\n  floor and CI execution before merge.\n- `bugfix`, `security-patch` add a targeted regression for the specific\n  bug/vulnerability and require the existing suite to remain green.\n- `express` uses the Minimal strategy: requirement-driven unit tests (one per\n  requirement, with a happy-path floor per component); existing tests remain\n  green.\n- `poc`, `refactor`, `workshop` add no extra new-test floor and require the\n  existing suite to remain green.\n\nThe active `Test Strategy` still applies in every scope and determines test\nvolume/types. Scope floors are additive; they never reduce or replace the\nselected strategy.\n\nBuild and Test verifies defined coverage floors and affirmed quality targets;\nthey may not be weakened to make a step pass.\n\nAffirm a stricter posture in `team.md` if the team commits to one."
    },
    {
      "layer": "team",
      "text": "- We treat tests as a deliverable of every Bolt, not extra work.\n- **Methodology**: tdd\n- **Ordering**: For each testable layer, we write a failing test first, then write the minimum code to make it pass, then refactor while the test stays green, before moving to the next behaviour or layer.\n- A floor of **80% line coverage for the Go code**, measured with `go test -coverprofile` and enforced by a `make coverage` target; CI blocks the merge when it is lower. The measured scope is `./internal/...` and `./server/...`; only the minimal wiring in `main` is excluded, and every exclusion must be listed explicitly in the `Makefile`. We do not lower the floor or add exclusions to make it pass.\n- Go tests always run with **`go test -race`** (needs CGO; works on the `ubuntu-latest` runner). This is an addition to the Bitbucket template, because token refresh and API rate limiting are two places prone to data races.\n- **Automated contract test with the real package on the minimum Kandev version**, like the template's `packaged-host-contract` job: CI checks out Kandev at exactly the `min_kandev_version` in `manifest.yaml`, builds a throwaway server, installs the packaged plugin, and runs it.\n- **Manual end-to-end check against a real Backlog space exactly twice**: when the walking skeleton is done, and before the first release. Each result is recorded. After the first release, we rely only on automated tests.\n- Unit and integration tests do not call real Backlog: they use a **fake Backlog server built with `net/http/httptest`**, with JSON sample data in `internal/<package>/testdata/`, including the 401, 429 (with `Retry-After`), and expired-token error cases.\n- Tools: `testing` + `github.com/stretchr/testify/require`, table-driven tests with `t.Run`, test names that describe behaviour. The clock and wait functions are injected into the code so rate limits and token expiry can be tested without a real `time.Sleep`. Tests are independent of each other and use `t.TempDir()` and `t.Setenv()`. The TypeScript UI (if any) is tested with Vitest.\n- There is a test asserting that API keys and tokens do not appear in logs, error messages, or responses sent to the UI.\n- We **do not add a dependency vulnerability scan** (`govulncheck`, `npm audit`, Dependabot) — this is your decision in Q4."
    },
    {
      "layer": "project",
      "text": "- Keep the coverage profile out of the repository root during AI-DLC stages (delete coverage.out after local make coverage runs, or write it under build/); reviewers never use -coverprofile, because an unclaimed coverage.out blocks the Code Generation gate (learned 2026-10-07) \n\n- Install the Go toolchain (Go 1.26.x) and link ../kandev to the pinned v0.96.0 checkout before Reverse Engineering, so the scan records a Go test baseline instead of skipping it (learned 2026-10-07) \n\n- Run the packaged-host contract test locally with make contract-test KANDEV_MIN_DIR=../kandev while the SDK checkout is at the minimum version tag, and run it 10 times to catch host startup races (learned 2026-10-07)"
    }
  ],
  "obligations": {
    "strategy": "minimal",
    "strategy_volume": [
      "One verifiable test per requirement at the narrowest effective level.",
      "At least one happy-path unit test per component.",
      "Unit tests are the default; a bugfix/security scope floor may require an integration or E2E regression when that is the narrowest level that reproduces the defect."
    ],
    "scope_floor": [
      "Include a targeted regression for the bug or vulnerability.",
      "Keep the existing test suite green."
    ],
    "combination_rule": "Apply every selected-strategy obligation and every scope-floor obligation; neither replaces the other, and a targeted scope regression may add the narrowest necessary test type beyond the strategy default."
  },
  "plan_profile": {
    "methodology": "tdd",
    "runner_step": "Verify the existing test runner/configuration and record the exact unit-scoped command.",
    "runner_ready_before_first_test": true,
    "testable_layers": [
      "Data model / database behavior",
      "Repository / data access",
      "Business logic",
      "API / endpoint",
      "Frontend behavior"
    ],
    "steps": [
      "Project structure and production configuration skeleton.",
      "Verify the existing test runner/configuration and record the exact unit-scoped command.",
      "Data model / database behavior - Red: write the failing tests and record the failing command output.",
      "Data model / database behavior - Green: implement only enough behavior to pass.",
      "Data model / database behavior - Refactor: improve the implementation while tests stay green.",
      "Repository / data access - Red: write the failing tests and record the failing command output.",
      "Repository / data access - Green: implement only enough behavior to pass.",
      "Repository / data access - Refactor: improve the implementation while tests stay green.",
      "Business logic - Red: write the failing tests and record the failing command output.",
      "Business logic - Green: implement only enough behavior to pass.",
      "Business logic - Refactor: improve the implementation while tests stay green.",
      "API / endpoint - Red: write the failing tests and record the failing command output.",
      "API / endpoint - Green: implement only enough behavior to pass.",
      "API / endpoint - Refactor: improve the implementation while tests stay green.",
      "Frontend behavior - Red: write the failing tests and record the failing command output.",
      "Frontend behavior - Green: implement only enough behavior to pass.",
      "Frontend behavior - Refactor: improve the implementation while tests stay green.",
      "Environment/build configuration.",
      "Documentation and traceability."
    ]
  },
  "input_sha256": "sha256:bfdca19d433a15267e965589c8fd047f16efa297557bc07dae30e96a563d49e8",
  "contract_sha256": "sha256:85ae4f86a7c2b09e82a94d61b48a4bfaeb7bbd2734f54bb954cf3fa6c6e1bdf6"
}
```

Inapplicable layers: no database schema or migration (the switch record format is unchanged; only the meaning of a missing record changes), no new production configuration.

## Steps

- [x] **Step 1 — Runner check (structure/config skeleton: none needed).** Confirm `../kandev` points at the pinned v0.96.0 checkout and run the unit-scoped commands from `unit-test-instructions.md` once to confirm a green baseline before any change. (NFR2)
- [x] **Step 2 — Test helper first (keeps existing tests meaningful).** Add one shared "switch on" step to the existing test harnesses that rely on the ON default — `newTestStore`/`newHarness` users in `internal/connection` and `newRig` in `internal/plugin` (plus `internal/issues`/`internal/git` rigs only if they read the real store) — so tests that exercise ON behaviour turn the switch on explicitly through `SaveSwitch`/`set_enabled`. Off-state tests keep their explicit OFF setup. No assertion is weakened or removed. Fix `TestSecondConnectInTheSameWorkspaceIsAConflict` so it fails fast instead of hanging if it breaks. Suite must still be green with the old default. (NFR2)
- [x] **Step 3 — Repository / data access, Red.** Invert `TestSwitchWithNoRecordIsEnabled` (`internal/connection/store_test.go:297`) into a "no record means off" regression test; add a `connection.get` view test asserting `enabled: false` for a workspace with no record. Run and record the failing output. (FR1, FR1.1, NFR1)
- [x] **Step 4 — Repository / data access, Green.** In `Store.LoadSwitch` (`internal/connection/store.go:212-227`) return `false, nil` when no record exists; keep the undecodable-record error (NFR3.9). (FR1.1)
- [x] **Step 5 — Repository / data access, Refactor.** Update the `LoadSwitch` doc comment: no record means off (opt-in); BR7.1 of intent 261006 is superseded by this intent. Tests stay green. (FR1.1)
- [x] **Step 6 — Business logic / API, Red.** Add targeted regressions for a freshly installed workspace with no switch record: (a) a guarded action (`connection.connect_api_key`) is refused with `409 integration_disabled` and no Backlog request is made; (b) the background workers (issue sync/watch, Git watcher) and the Git credential RPC do no work / refuse for that workspace; (c) after an admin turns the switch on, connect succeeds; (d) a non-admin cannot turn it on; (e) an explicit ON record stays ON. Reuse the existing off-state test patterns (`plugin/actions_test.go`, `credential_test.go`, `git/watcher_test.go`, `issues/sync_test.go`). Record which fail before Step 4 lands (if Step 4 is already in, record that they would fail by temporarily checking against the old default in the Red log). (FR1.2, FR2, NFR1)
- [x] **Step 7 — Business logic / API, Green.** No production change is expected beyond Step 4 (all paths already go through `RequireEnabled`); if any regression from Step 6 fails, fix the path so it uses the guard. (FR1.2, FR2)
- [x] **Step 8 — Business logic / API, Refactor.** Remove any duplication introduced in the helpers; suite green with `-race`. (NFR2)
- [x] **Step 9 — Frontend behavior, Red.** In `ui/src/switch/switch.test.tsx` and `ui/src/index.test.ts`, add tests that a settings view without an `enabled` field renders the Off state (switch off, page shows the existing Off copy). Run and record the failing output. (FR3, FR4)
- [x] **Step 10 — Frontend behavior, Green.** Change the fallbacks to `view.enabled === true` in `ui/src/index.ts:32` and `ui/src/switch/integration-switch.tsx:45`. Existing Off wording untouched. (FR3, FR4)
- [x] **Step 11 — Frontend behavior, Refactor.** Tidy; `tsc --noEmit`, ESLint, Prettier and Vitest green. (NFR2)
- [x] **Step 12 — Contract test (environment/build).** Update `internal/ci/contract.go:195-207` and its fixtures in `internal/ci/contract_test.go` to: assert the fresh install reports `enabled: false` and `connect_api_key` is refused with `409 integration_disabled`; then turn the switch on (`connection.set_enabled`) and run the existing connect validation (`400 validation spaceUrl`), so the real connect path is still exercised. Run `make contract-test KANDEV_MIN_DIR=../kandev` (10 runs) if the environment allows; otherwise record that CI runs it. (FR6)
- [x] **Step 13 — Documentation.** README (`README.md:38`): Backlog is off by default after installation; an admin turns it on, then connects. Add an "Upgrading from v0.1.x" note: workspaces that never touched the switch turn off after upgrading; the saved connection is kept; an admin turns the switch on to resume (no reconnect needed). The release notes for the next release reuse this note (written in the Deployment Pipeline stage). (FR5, FR5.1, FR5.2)
- [x] **Step 14 — Full verification.** `go test -race ./...`, `make coverage` (≥ 80%, delete `coverage.out` afterwards or keep it under `build/`), `make lint`, `make vet`, `make check-format`, UI checks. No secrets in logs/tests (NFR3). (NFR2, NFR3)
- [x] **Step 15 — Traceability.** Write `code-summary.md`, `source-manifest.json` and `traceability.json` (FR1–FR6, NFR1–NFR3 → implementation/test files).

## Requirement → Step Map

| Requirement | Steps |
|---|---|
| FR1 / FR1.1 | 3, 4, 5 |
| FR1.2 | 6, 7 |
| FR2 | 6, 7 |
| FR3 | 9, 10 |
| FR4 | 9, 10 |
| FR5 / FR5.1 / FR5.2 | 13 |
| FR6 | 3, 12 |
| NFR1 | 3, 6 |
| NFR2 | 1, 2, 8, 11, 14 |
| NFR3 | 14 |
