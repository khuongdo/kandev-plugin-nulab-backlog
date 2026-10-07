# Code Generation Plan — Plugin install failed: 502

## Overview

Zero-Unit bugfix (scope `bugfix`, Minimal depth, brownfield). Source of work: [requirements.md](../../inception/requirements-analysis/requirements.md) and the code knowledge base at `aidlc/spaces/default/codekb/kandev-plugin-nulab-backlog/`.

Goal: shrink the release package from 5 to 4 server executables by dropping `windows-amd64` (FR1), and document install by release URL plus the 502 troubleshooting note (FR2). No Go runtime behaviour of the plugin changes; the only testable layer is the package verifier (`internal/pkgverify`, business logic of the build tooling).

## Blast Radius

| File | Change | Impact |
|---|---|---|
| `internal/pkgverify/pkgverify.go` | 4-entry `executables` map; reject any `server/` file that is not a listed executable; wording "four" instead of "5"/"five" | low (build tooling only; used by `cmd/verifypkg`, `make verify-package`, release workflow) |
| `internal/pkgverify/pkgverify_test.go` | fixtures without Windows; regression cases | low |
| `manifest.yaml` | `runtime.executables` without `windows-amd64` | medium (decides which Kandev hosts can install) |
| `Makefile` | `PLATFORMS` without `windows-amd64`; drop the `.exe` branch; clear `build/server/` before building so a stale binary is never packaged | low |
| `README.md` | install by URL (recommended), upload as alternative, 502 troubleshooting, supported platforms | none (docs) |

Not changed: `manifest.yaml` `version` (set at release in Deployment Pipeline), release workflow, CI workflow, UI, plugin Go code.

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

Applicable layers: only **business logic** (the package verifier). Data model, repository, API and frontend layers are not touched by this fix and are omitted.

## Steps

- [x] **Step 1 — Toolchain and test runner.** Make Go 1.26.x available and point `../kandev` (relative to the worktree) at the Kandev checkout at `.kandev-sdk-ref` (v0.96.0, e.g. a `git worktree` of `~/repo/kandev`). Verify the runner with the exact command in `unit-test-instructions.md` and record the baseline result (pass count) of `go test -race ./internal/pkgverify/...` before any change. (NFR2)
- [x] **Step 2 — Business logic, Red.** In `internal/pkgverify/pkgverify_test.go`: change the good fixture to the 4 Linux/macOS executables (manifest and files); add regression cases: (a) a manifest that still lists `windows-amd64` is rejected; (b) a package that contains `server/plugin-windows-amd64.exe` (checksummed) while the manifest lists only the 4 is rejected with `unexpected executable: server/plugin-windows-amd64.exe`; (c) a manifest missing one of the 4 is rejected; keep the existing "missing executable" case. Update expected error text to `manifest must list exactly the four executables`. Run the test command and record the failing output. (FR1.1, FR1.2, NFR3)
- [x] **Step 3 — Business logic, Green.** In `internal/pkgverify/pkgverify.go`: remove `windows-amd64` from `executables`; add a check that every archive file under `server/` is one of the listed executables (`unexpected executable: <path>`); change the manifest error to `manifest must list exactly the four executables of BR5.1`. Run the test command until green. (FR1.1, FR1.2, NFR3)
- [x] **Step 4 — Business logic, Refactor.** Update doc comments ("exactly the four executables", the `maxEntry` comment) and keep tests green; `gofmt` clean. (FR1.2)
- [x] **Step 5 — Build configuration.** `manifest.yaml`: remove `windows-amd64` from `runtime.executables`. `Makefile`: `PLATFORMS := linux-amd64 linux-arm64 darwin-amd64 darwin-arm64`; drop the `ext`/`.exe` branch; `build` removes `$(BUILD)/server` before building so a stale binary from an older build is never copied into the package. (FR1.1, FR1.2)
- [x] **Step 6 — Package check.** Run `make package verify-package` and record the package size from `ls -l dist/`; it must be at most 25 MB (NFR1). If it is larger, stop and report the measured size instead of changing the target. Run `go test -race ./...` (whole suite stays green, NFR2).
- [x] **Step 7 — Documentation.** `README.md` "Install on Kandev": (1) recommended: **Settings > Plugins**, install from URL with the GitHub Release asset URL `https://github.com/khuongdo/kandev-plugin-nulab-backlog/releases/download/v<version>/nulab-backlog-<version>.tar.gz` (Kandev downloads it itself); (2) alternative: upload the `.tar.gz`; (3) a "Troubleshooting" note: `Plugin install failed: 502` after about 30 s (server log: 400 `missing multipart field "package"`) means the upload took longer than the server's 30 s read limit — install from URL, or raise `KANDEV_SERVER_READTIMEOUT` (seconds) on the Kandev server; (4) supported server platforms: Linux and macOS on amd64 and arm64, no Windows. Remove any other "5 binaries"/Windows wording in README. (FR2.1, FR2.2, FR2.3)
- [x] **Step 8 — Traceability and summary.** Write `code-summary.md`, `source-manifest.json` and `traceability.json` for FR1.1, FR1.2, FR2.1, FR2.2, FR2.3, NFR1, NFR2, NFR3. FR1.3 (patch release) is traced to the Deployment Pipeline stage.

## Requirement Traceability

| Requirement | Plan step |
|---|---|
| FR1.1 four executables, no `.exe` | Steps 2, 3, 5 |
| FR1.2 manifest, Makefile, verifier agree; verifier rejects Windows and missing executables | Steps 2–5 |
| FR1.3 patch release | Deployment Pipeline (not this stage) |
| FR2.1 install by URL documented | Step 7 |
| FR2.2 502 troubleshooting note | Step 7 |
| FR2.3 supported platforms stated | Step 7 |
| NFR1 package at most 25 MB | Step 6 |
| NFR2 suite green, coverage floor | Steps 1, 6 (coverage in Build and Test) |
| NFR3 regression test | Steps 2, 3 |
