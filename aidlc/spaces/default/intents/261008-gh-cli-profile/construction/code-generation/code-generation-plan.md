# Code Generation Plan — 261008-gh-cli-profile (zero-Unit, express)

## Scope

Per-workspace gh CLI account for the GitHub source control connection. Inputs: `../../inception/requirements-analysis/requirements.md` (FR1-FR6, NFR1-NFR4), `../../inception/user-stories/stories.md` (US1.1-US4.1, AC IDs), codekb `aidlc/spaces/default/codekb/kandev-plugin-nulab-backlog/`. No Units, no design stages (express).

Design decisions carried from the stories (mob notes):
- The chosen login is the existing `Settings.AccountID` (already the GitHub login and the "mine" filter key). No new stored field, no migration (US3.2).
- `Service.credential()` is the single token source; it passes the workspace's login to the CLI token read (US2.1).
- The CLI token cache is keyed by `{provider, login}`; `forgetCLI(p)` drops every login of `p`.
- `runCLI` sets `cmd.Env` to the parent environment minus `GH_TOKEN`, `GITHUB_TOKEN`, `GH_ENTERPRISE_TOKEN`, `GITHUB_ENTERPRISE_TOKEN` (new work, AC1.1.6, AC2.1.8).
- New error code `cli_account_missing` (manifest/action error mapping, UI message built from code + stored login, AC3.1.1).
- Token read for login L: `gh auth token --hostname github.com --user L`; on failure, list accounts: list fails → `cli_unavailable`; L not listed → `cli_account_missing`; L listed (old gh without `--user`) → plain `gh auth token --hostname github.com`, accepted only when GitHub `/user` login equals L (case-insensitive), else `cli_account_missing` (FR3.4, AC2.1.4, AC2.1.6, AC3.1.5).
- Account list: `gh auth status --json hosts --hostname github.com`; decode stdout even on non-zero exit; keep only entries with `state == "success"`; validate each login; return only `login` + `active`. Output cap raised for this call only (32 KiB); a cut-off or undecodable output → `cli_unavailable`. If gh has no `auth status --json` (decode fails / unknown flag), fall back to the active account only (plain token + `/user`) (AC1.1.8, AC1.1.9). The exact gh version that added `--json` is confirmed and recorded in code-summary.
- Login rule `^[A-Za-z0-9](?:-?[A-Za-z0-9])*$`, length ≤ 39, applied to browser input and gh output (AC1.1.5).
- `use_cli` body gains optional `login`. GitHub: empty = the active account (resolved through the list, so the stored `AccountID` is the active login). GitLab: non-empty login rejected; empty behaves as before (AC3.2.3).
- Test in CLI mode keeps `AccountID`; it may refresh `Account` (display name) only when `/user` returns the same login (FR5.2, AC3.1.2).
- Provider view gains `login` (from `AccountID`) for the card.

## Testing Contract

```json
{
  "version": 1,
  "methodology": "tdd",
  "source": "team",
  "ordering": "For each testable layer, we write a failing test first, then write the minimum code to make it pass, then refactor while the test stays green, before moving to the next behaviour or layer.",
  "scope": "express",
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
      "text": "- Keep the coverage profile out of the repository root during AI-DLC stages (delete coverage.out after local make coverage runs, or write it under build/); reviewers never use -coverprofile, because an unclaimed coverage.out blocks the Code Generation gate (learned 2026-10-07) \n\n- Install the Go toolchain (Go 1.26.x) and link ../kandev to the pinned v0.96.0 checkout before Reverse Engineering, so the scan records a Go test baseline instead of skipping it (learned 2026-10-07) \n\n- Run the packaged-host contract test locally with make contract-test KANDEV_MIN_DIR=../kandev while the SDK checkout is at the minimum version tag, and run it 10 times to catch host startup races (learned 2026-10-07) \n\n- When the worktree has no Go, install Go 1.26.x to ~/.local/go (checksum-verified from go.dev) and run make coverage, make package/verify-package and 10x make contract-test KANDEV_MIN_DIR=../kandev locally, so no target stays Unverified (learned 2026-10-08)"
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
      "Keep the existing test suite green.",
      "This scope adds no extra new-test floor beyond the selected test strategy."
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
  "input_sha256": "sha256:ca371ed31439c4d37a250d5e63d50b77aeaddf0755a2b7cfb8af1ee9221f1c2e",
  "contract_sha256": "sha256:dc2c3668efe5d53a084b8cac79c7c6d886c9148655498e84845031714e7646b7"
}
```

## Steps

Layers not applicable: data model / database (state is a JSON settings record; the only change is reading `AccountID` as the chosen login, covered under business logic) and repository / data access (no store change; the frozen v0.5.2 fixture test lives with business logic).

- [x] **Step 1 — Workspace setup.** Link `../kandev` to the pinned v0.96.0 checkout (`~/repo/kandev`, commit matches `.kandev-sdk-ref`) so `go.mod`'s `replace` resolves; confirm Go 1.26.x is on PATH; run `npm ci` in `ui/`. No repo file changes. (all stories)
- [x] **Step 2 — Runner check.** Run the exact unit-scoped commands in `unit-test-instructions.md` and confirm the existing tests are green before any Red step (baseline 1383 Go test results). (NFR4)
- [x] **Step 3 — CLI layer Red** (`internal/scm/cli_token_test.go`): fake runner recording argv + env with an argv allowlist guard (FR3.3). Failing tests for: login validation (AC1.1.5); env stripping of `GH_TOKEN`/`GITHUB_TOKEN` (AC1.1.6, AC2.1.8); account list parse — active marking, error-state entries dropped, non-zero exit with decodable stdout, cut-off/undecodable → unavailable, response fields only login+active (AC1.1.1, AC1.1.6, AC1.1.8); `--user` token read and cache per `{provider, login}` with injected clock (AC1.2.2, AC2.1.7); old-gh fallback accept/refuse and the unavailable vs account-missing classification (AC2.1.4, AC2.1.6, AC3.1.5); no-`--json` fallback to active only (AC1.1.9). Record the failing output. (US1.1, US1.2, US2.1, US3.1)
- [x] **Step 4 — CLI layer Green**: in `internal/scm/cli_token.go` add the login rule, `ListCLIAccounts`, login-aware `cliToken(ctx, p, login)`, `{provider, login}` cache key, env filtering in `runCLI`, per-call output cap, `ErrCLIAccountMissing`. Minimum code to pass.
- [x] **Step 5 — CLI layer Refactor** while green; keep the existing `ponytail:` lock note accurate.
- [x] **Step 6 — Service layer Red** (`internal/scm/service_test.go` or a new `service_cli_account_test.go`): `UseCLI` with login saves `AccountID` = login, empty login uses the active account, unknown login → account-missing and nothing saved, GitLab non-empty login rejected (AC1.1.2, AC1.1.7, AC1.2.1, AC1.2.3, AC3.2.3); `credential()` passes the workspace login for every call site — table over Test, SearchRepos, PR list, PR link, "mine" filter, background PR-watch poll, with an httptest GitHub server recording `Authorization` per workspace, concurrent W1/W2 under `-race` (AC2.1.1, AC2.1.2, AC2.1.3, AC2.1.5); Test keeps `AccountID` after `gh auth switch` and records `cli_account_missing` when the login is gone (AC3.1.1, AC3.1.2); frozen v0.5.2 settings fixture in `internal/scm/testdata/` keeps working as its `AccountID` and a record with no `AccountID` fails as account-missing (AC3.2.1, AC3.2.4); redaction table — no token sentinel or raw gh output in errors, logs or views (AC3.1.3). Record the failing output.
- [x] **Step 7 — Service layer Green**: `UseCLI(ctx, ws, p, login)`, `credential()` login pass-through, Test guard, `errorCode` mapping for `ErrorCLIAccountMissing`, `ProviderView.Login`.
- [x] **Step 8 — Service layer Refactor** while green.
- [x] **Step 9 — Action layer Red** (`internal/plugin/scm_actions_test.go`, `internal/plugin/manifest_test.go`): new `scm.providers.cli_accounts` action returns `{accounts:[{login, active}]}` and maps failures to `cli_unavailable`; `scm.providers.use_cli` accepts `login` and rejects an invalid one as a field error; `cli_account_missing` reaches the browser as a code; the manifest declares the new action key with the same access as `use_cli`. (AC1.1.1, AC1.1.5, AC1.1.7, AC3.1.1)
- [x] **Step 10 — Action layer Green**: `internal/plugin/scm_actions.go` handler + body field, `manifest.yaml` action entry.
- [x] **Step 11 — Action layer Refactor** while green.
- [x] **Step 12 — UI Red** (`ui/src/settings/source-control-section.test.tsx`): "Use gh CLI login" loads accounts (button disabled, "Loading gh accounts…"); one account connects at once (AC1.1.4); two or more show the labelled select "GitHub account (gh)" with `alice (active in gh)` preselected, focus on the select, Connect/Cancel (AC1.1.1, AC1.1.2); connected card reads "Connected via gh CLI as @bob" (AC1.1.2, AC3.2.2); Change account opens the picker with the current login preselected, Cancel saves nothing (AC1.2.1, AC1.2.4); account-missing alert text exact and Change account available (AC3.1.1, AC3.1.4); failures use `role="alert"` (AC1.1.3); no-`--json` case shows only the active account (AC1.1.9); worktree note present in gh CLI mode with the login filled in, absent in token mode, read-only cards show note but no picker (AC4.1.1); `data-testid` on new controls. Record the failing output.
- [x] **Step 13 — UI Green**: `ui/src/settings/source-control-section.tsx` (and its API/types module and `ui/src/messages/en.ts` strings) using the host `Select` already used by `scm-watch-form.tsx`.
- [x] **Step 14 — UI Refactor** while green; run `tsc --noEmit`, ESLint and Prettier on `ui/`.
- [x] **Step 15 — Build configuration.** No new dependency. Bump plugin version in `manifest.yaml` (and `ui/package.json` if versioned together) to the next patch/minor per repo convention. Run `gofmt`, `go vet`, golangci-lint if available.
- [x] **Step 16 — Documentation and traceability.** README source control section: per-workspace gh account, gh version note, worktree note (AC4.1.2). Update doc comments that cite requirement IDs. Write `code-summary.md`, `source-manifest.json`, `traceability.json` (every AC from stories.md + NFR1-NFR4).

## Story Traceability

| Story | Steps |
|---|---|
| US1.1 | 3, 4, 6, 7, 9, 10, 12, 13 |
| US1.2 | 3, 4, 6, 7, 12, 13 |
| US2.1 | 3, 4, 6, 7 |
| US3.1 | 3, 4, 6, 7, 9, 10, 12, 13 |
| US3.2 | 6, 7, 12, 13 |
| US4.1 | 12, 13, 16 |
