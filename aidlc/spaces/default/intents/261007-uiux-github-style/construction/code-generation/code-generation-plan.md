# Code Generation Plan — 261007-uiux-github-style

Zero-Unit refactor intent: one implementation iteration over the whole change. Inputs: [requirements.md](../../inception/requirements-analysis/requirements.md), [functional-spec.md](../functional-design/functional-spec.md), [rules.md](../functional-design/rules.md), [entities.md](../functional-design/entities.md), [frontend-components.md](../functional-design/frontend-components.md), and the code knowledge base `aidlc/spaces/default/codekb/kandev-plugin-nulab-backlog/`.

## Scope and Baseline

- Brownfield: modify files in place; no duplicates.
- Baseline before any change (recorded in RE, re-run in Step 2): Go `go test -race` all packages pass, 92.9% coverage; UI Vitest 229/229, `tsc`, ESLint, Prettier clean.
- Environment: Go 1.26.x on `PATH` (`$HOME/.local/go/bin`), `../kandev` → v0.96.0 (`f099a46`), Node 22 per `.nvmrc`. Coverage profiles go under `build/`, never the repo root.
- Out of scope: PR watch loop and its 10-task cap (unchanged), new manifest capabilities.

## Design Decisions Closing Review Findings R-09 and R-10

The functional design was approved with two open reviewer findings. This plan resolves them inside BR3.12 without changing any other rule:

- **R-09 (resume offset overshoot)**: a run starts at `max(0, dayOffset - 5)`; if the first page read does not contain an issue at or before the cursor (so the start may have overshot), the run restarts once from offset 0 of the cursor date. Test: issues before the cursor were closed so the list shrank by more than 5 — no matching issue is skipped.
- **R-10 (createdSince timezone)**: `createdSince` is the cursor's created time in UTC minus one day (`yyyy-MM-dd`); the client-side `(created, id) > cursor` filter makes the extra day harmless. Test: a cursor at 00:30 UTC still sees issues created later the same UTC day.

## Testing Contract

```json
{
  "version": 1,
  "methodology": "tdd",
  "source": "team",
  "ordering": "For each testable layer, we write a failing test first, then write the minimum code to make it pass, then refactor while the test stays green, before moving to the next behaviour or layer.",
  "scope": "refactor",
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
      "text": "- Keep the coverage profile out of the repository root during AI-DLC stages (delete coverage.out after local make coverage runs, or write it under build/); reviewers never use -coverprofile, because an unclaimed coverage.out blocks the Code Generation gate (learned 2026-10-07) \n\n- Install the Go toolchain (Go 1.26.x) and link ../kandev to the pinned v0.96.0 checkout before Reverse Engineering, so the scan records a Go test baseline instead of skipping it (learned 2026-10-07)"
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
  "input_sha256": "sha256:cd0b9356dfd3a1d4e31665e1e9719a457a2eae09499a297fa2bcd4433ff424e7",
  "contract_sha256": "sha256:f482164fba027f02e5cb78452acda04f5091d3f5ec9a0217ffb0cde178cabfa0"
}
```

## Steps

Every Red step records the exact failing command and its output in `code-summary.md` before the matching Green step. Each step lists the requirements and rules it implements.

### Step 1: Project structure and production configuration skeleton
- [x] Confirm package layout: issue watch in `internal/issues` (new files `watch.go`, `watch_store.go`, `watcher.go`), PR list in `internal/git` (`prs.go`), gateway additions in `internal/backlog` (`issues.go`, `pullrequests.go`), actions in `internal/plugin/issue_actions.go` and `git_actions.go`. No new packages, no new dependencies, no new manifest capabilities. (FR3, FR4)

### Step 2: Verify the test runner and record the exact commands
- [x] Run the Go and UI commands from `unit-test-instructions.md` on the unchanged tree; record pass counts as the baseline in `code-summary.md`.

### Step 3: Data model — Red
- [x] Table tests for `IssueWatchInput.Validate`: name 0/1/100/101 chars, unselected project, 0/1/20/21 status ids, duplicate status ids removed, assignee/creator only `anyone|me`, empty workflow, interval 0/1/1440/1441/non-integer, default interval 5 when absent (BR3.1, BR3.2; FR3.1, FR3.6).
- [x] Tests for `IssueCursor` ordering `(created, id)` and `Before/After` helpers (BR3.5).
- [x] Test that `QueryInput` accepts optional `creator` and treats absent as `anyone` (SavedPRQuery change; FR2.6).
- [x] Run, record failing output.

### Step 4: Data model — Green
- [x] Implement `IssueWatch`, `IssueWatchInput`, `IssueCursor`, `IssueWatchLedgerEntry` types and validation; add `Creator` to `QueryInput`.

### Step 5: Data model — Refactor
- [x] Share the `anyone|me` validation with `internal/git` only if it removes duplication without a new import cycle; otherwise keep local. Tests stay green.

### Step 6: Repository / data access — Red
- [x] `internal/issues` store tests: watches document round-trip, cap 50 watches, ledger per watch with uniqueness `(watchId, key)`, cap 5000 per watch, delete watch removes its ledger (BR3.10, BR3.13), schemaVersion handling.
- [x] `internal/backlog` gateway tests against the `httptest` fake: `IssueQuery` encodes `createdUserId[]`, `createdSince`, `sort=created&order=asc` when set and keeps `sort=updated&order=desc` by default; `PullRequestCount` calls `.../pullRequests/count` with the same filters; 401 and 429 (`Retry-After`) map to the existing error type (BR3.5, BR4.1; NFR3).
- [x] Run, record failing output.

### Step 7: Repository / data access — Green
- [x] Implement the store documents and the gateway additions (fixtures in `internal/backlog/testdata/`).

### Step 8: Repository / data access — Refactor
- [x] Keep the copied document helpers as today (`ponytail` note already in `internal/issues/store.go`); tidy names; tests green.

### Step 9: Business logic — Red
- [x] Issue watch service tests: save (new/edit), edit with filter change clears cursor (BR3.8), pause/resume, run-now refused when not active (BR3.9), connection lost stores `stateBeforeDisconnect` and reconnect restores it (SM1, BR3.9), delete keeps tasks/links (BR3.10).
- [x] Watcher run tests with injected clock and fake gateway/host: due by per-watch interval (BR3.3); at most one task per run (BR3.4); existing issues picked oldest first (BR3.5); linked issue skipped and cursor advanced (BR3.6); reserve→create→link→mark, failed create removes reservation, leftover reservation resolved next run (BR3.7); at most 5 pages per run and cursor advances past skipped issues (BR3.12); R-09 overshoot fallback; R-10 createdSince UTC minus one day; ledger full → `ledger_full` (BR3.13); removed workflow → `workflow_missing` and watch stays active (BR3.14); 401 → `unauthorized`, 429 → `rate_limited`, no task (BR3.11); bounded wait for the host before the first tick (NFR6).
- [x] PR list service tests: 20 per page, offset `(page-1)*20`, total from count, `hasNext`, unselected project rejected, page < 1 rejected, linked task joined by `spaceHost|repositoryId|number` (BR4.1, BR4.2; FR4.1, FR4.2).
- [x] Run, record failing output.

### Step 10: Business logic — Green
- [x] Implement `issues.WatchService` methods, `issues.Watcher` (1-minute tick, run queue like the PR watcher), and `git.Service.ListPullRequests`. Task title/description/priority reuse `NewTaskFor`; links reuse `putLink`.

### Step 11: Business logic — Refactor
- [x] Remove duplication between `CreateTask` and the watcher's create path; keep `-race` green.

### Step 12: API / endpoint — Red
- [x] Adapter tests for `issues.watches.list|save|delete|run|pause|resume` and `git.prs.list`: request decoding, field errors mapped to validation, conflict mapping, access `authenticated`; manifest test lists the new keys with the key pattern (FR4.3, BR1.3); `git.queries.save` accepts `creator`.
- [x] Extend the secret-leak test to the new actions and watcher log lines (NFR2).
- [x] Runtime test: the issue watcher starts and stops with the plugin and waits for the host (NFR6).
- [x] Run, record failing output.

### Step 13: API / endpoint — Green
- [x] Add handlers to `internal/plugin/issue_actions.go` and `git_actions.go`, register the watcher in `runtime.go`, declare the actions in `manifest.yaml`.

### Step 14: API / endpoint — Refactor
- [x] Align handler shapes with existing ones; tests green.

### Step 15: Frontend behavior — Red
- [x] Harness: stub the newly used host components (`SettingsSection`, `Card*`, `Table*`, `Dialog*`, `DropdownMenu*`, `Select*`, `Checkbox`, `Alert*`, `Empty*`, `Pagination*`, `Tabs*`, `IntegrationRepositoryFilter`, `ChangeRequestList`, `ChangeRequestRow`, `IntegrationChangeRequestStatus`) with `variant`/`size` passed through as data attributes.
- [x] `index.test.ts`: exactly one integrations nav item and one route `/backlog`; no `/backlog/watches` or `/backlog/dashboard`; independent of enabled state (BR2.1; FR2.1, FR2.2).
- [x] Settings tests: seven sections in order when connected, only Connection when not (BR1.1); watch/query sections visible in member view (BR1.2); sign-in method is a `Select` (BR1.4); restore notice scrolls to PR watches (BR1.5); PR watches table + dialog; Issue watches table + `IssueWatchDialog` with field errors and default interval 5 (FR1.2, FR1.3, BR3.1); Saved PR queries table with edit/delete and confirm dialog (Q3, BR2.5).
- [x] `/backlog` tests: alert when not connected or disabled (BR2.3); Tabs Issues/Pull requests with `?scope=prs` (BR2.2); Issues list keeps filters/search/paging (FR2.4); PR list: choose-a-repository empty state, rows with linked task, paging via `hasNext`/`total`, saved query preset applies filters and resets page, unselected-project preset disabled, "Save query" opens dialog with current filters (FR2.5, FR2.6, BR2.4, BR2.5, BR4.1); empty and error states including reconnect message (BR7.1, BR7.2).
- [x] Control style tests: no raw `select`, `table`, checkbox/radio `input`, `details`, or `button` elements rendered by plugin components; button variant/size per role (BR5.1, BR5.2; FR5); icon-only buttons have `aria-label` (NFR4); axe checks on changed screens.
- [x] Icon test: outline SVG with `stroke="currentColor"`, `fill="none"`, size from `className`, used by nav, card and topbar; no Nulab colours (BR6.1; FR6).
- [x] Run, record failing output.

### Step 16: Frontend behavior — Green
- [x] Rebuild `SettingsScreen` as stacked sections (new `settings/pr-watches-section.tsx`, `settings/issue-watches-section.tsx`, `settings/issue-watch-dialog.tsx`, `settings/saved-queries-section.tsx`; reuse and restyle `watch-form.tsx` as the PR watch dialog), move the poll interval into `IssueSyncSection`.
- [x] Rebuild `BacklogPage` with Tabs, keep `issues-page.tsx` behaviour with host controls, add `git/pr-list.tsx` and `git/pr-toolbar.tsx`.
- [x] Replace raw controls across `settings/`, `issues/`, `git/` with host components and the button style table.
- [x] Replace `brand/backlog-logo.tsx` content with the original outline icon behind `PLUGIN_ICON`.
- [x] Update `index.ts` registrations; delete `git/watches-page.tsx`, `git/dashboard-page.tsx` and their tests (their behaviour moves to Settings and the PR list).
- [x] Add message strings to `messages/en.ts`.

### Step 17: Frontend behavior — Refactor
- [x] Replace the 12 copies of `STACK`/`FIELD`/`ROW` with one `ui/src/layout.ts` module; `tsc`, ESLint, Prettier clean; tests green.

### Step 18: Environment / build configuration
- [x] `manifest.yaml` actions added; `make ui-build` gets `--jsx-fragment=Fragment` to match `npm run build`; `make coverage` writes its profile under `build/` instead of the repo root (Makefile only; the 80% floor and exclusions are unchanged).
- [x] Run `make check-format vet lint test coverage build package verify-package` with `../kandev` linked.

### Step 19: Documentation and traceability
- [x] Update `docs/brand/backlog-logo.md` (plugin no longer ships the Nulab mark) and the README sections for Settings, `/backlog`, issue watches and the PR list.
- [x] Write `code-summary.md`, `source-manifest.json` (every created/modified/deleted path) and `traceability.json` (every FR/NFR/BR mapped to an implementation or test file).

## Requirement-to-Step Map

| Requirement | Steps |
|---|---|
| FR1 Settings sections, watches in Settings | 15, 16 |
| FR2 One entry, Issue/PR lists | 9, 10, 15, 16 |
| FR3 Issue watch | 3–14, 15, 16 |
| FR4 PR list action | 6, 7, 9, 10, 12, 13 |
| FR5 Host controls | 15, 16, 17 |
| FR6 Outline icon | 15, 16, 19 |
| NFR1 TDD, coverage | every Red/Green pair, 18 |
| NFR2 No secret leaks | 12 |
| NFR3 Rate limits | 6, 9 |
| NFR4 Accessibility | 15 |
| NFR5 SDK v0.96.0 compatibility | 15, 16, 18 |
| NFR6 Host wait | 9, 12 |
