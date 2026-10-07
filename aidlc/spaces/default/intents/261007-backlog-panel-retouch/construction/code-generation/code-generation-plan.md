# Code Generation Plan — 261007-backlog-panel-retouch

Zero-Unit refactor (no Unit DAG). Scope: refactor · Depth: Minimal · Test strategy: Minimal · Brownfield. Inputs: `construction/functional-design/` (functional-spec.md, rules.md BR1.1–BR6.2, entities.md, frontend-components.md, reviews/review-01.md R-01..R-08) and `inception/requirements-analysis/requirements.md` (FR1–FR5, NFR1–NFR5).

## Blast Radius

| File | Change | Impact |
|---|---|---|
| `ui/src/testing/harness.ts` | Add fakes `TaskRowIndicator`, `IntegrationListToolbar`, `Popover`/`PopoverTrigger`/`PopoverContent` (if missing) | Medium — shared by all UI tests |
| `ui/src/issues/issues-page.tsx` | Host toolbar, draft/committed query, `IntegrationRepositoryFilter` filters, `TaskRowIndicator`, remove Filters toggle and 400 ms timer | High — main /backlog list |
| `ui/src/git/pr-toolbar.tsx` | Restyle to `IntegrationListToolbar` layout (stacked mobile status row) | Medium — PR list only after this change |
| `ui/src/git/pr-list.tsx` | `TaskRowIndicator`, label-less dropdown filters, `StatusMultiFilter` popover | Medium |
| `ui/src/issues/issue-badge.tsx` | Anchor when openable, accessible detail | Medium — every Kanban card with a link |
| `ui/src/issues/issues-state.ts` | Small pure helpers (task link mapping, badge openability, https check) | Low |
| `ui/src/messages/en.ts` | New/renamed strings (All projects/statuses/assignees, Status (n), query placeholder, query label) | Low |
| Tests: `ui/src/issues/issues-page.test.tsx`, `ui/src/issues/issue-badge.test.tsx`, `ui/src/page/backlog-lists.test.tsx`, `ui/src/page/backlog-page.test.tsx`, `ui/src/controls.test.ts` and any test using removed test ids | Update, not delete | Medium |

No Go, manifest or action change (BR6.2). Test baseline from the code knowledge base: Vitest 31 files / 322 tests passing; `tsc`, ESLint, Prettier clean; Go `go test -race ./internal/... ./server/...` all passing (`internal/issues` 91.5%, `internal/plugin` 93.8%).

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
  "input_sha256": "sha256:def693f2644e2e00a25ba7a5c1561d606d5e781b2b967d6ac9958de98b537a56",
  "contract_sha256": "sha256:c832d7302104c2278f4539d9d4082d5cff074338c2d095de61b5f5091ea1e44a"
}
```

## Layer Applicability

- Data model / database, Repository / data access, API / endpoint: **not applicable** — no backend, storage or action change (BR6.2, functional-spec D7). The Go suite must stay green unchanged.
- Business logic: small pure UI helpers in `ui/src/issues/issues-state.ts` (TDD).
- Frontend behavior: all component changes (TDD per behaviour slice).

## Steps

### Setup

- [x] **Step 1 — Environment.** Link `../kandev` → `/home/k_do_webfrontier/repo/kandev` (tag `v0.96.0`, commit `f099a46`, equal to `.kandev-sdk-ref`); use Go at `~/go-sdk/bin/go`; run `npm ci` in `ui/` if `node_modules` is missing. No production configuration change.
- [x] **Step 2 — Runner readiness.** Run the scoped Vitest command from `unit-test-instructions.md` against the current tests and confirm it passes (baseline) before the first Red.

### Business logic (pure helpers, `issues-state.ts`)

- [x] **Step 3 — Red.** Tests in `ui/src/issues/issues-state.test.ts` for: `taskRowLinks(linkedTasks)` → `{id, taskId, fallbackTitle: taskKey ?? taskId}` (BR1.2); `prTaskRowLinks(ids)` → `fallbackTitle = id` (BR1.2); `badgeHref(link)` → url only when available, connected and the URL is `https://` (BR2.2, BR2.3, R-07). Record failing output.
- [x] **Step 4 — Green.** Implement the three helpers.
- [x] **Step 5 — Refactor** while green.

### Frontend behavior

- [x] **Step 6 — Harness fakes (test support).** Add to `ui/src/testing/harness.ts` fakes mirroring host v0.96.0 props and test ids: `TaskRowIndicator` (nothing for empty; `${prefix}-single` showing `fallbackTitle`; `${prefix}-multi` with count and items; click records host navigation to the task), `IntegrationListToolbar` (title `titleTestId`, count, query input `queryTestId` calling `onCustomQueryChange`, Enter and blur-if-dirty calling `onCommitCustomQuery`, `filter` node, refresh `refreshTestId`), and `Popover`/`PopoverTrigger`/`PopoverContent` if not present. `IntegrationRepositoryFilter` fake gains `triggerClassName`/`className` pass-through.
- [x] **Step 7 — Red: linked tasks on issue rows (FR1, BR1.1–BR1.5).** In `issues-page.test.tsx`: one task → `backlog-issue-task-<key>-single` with the fallback title; three tasks → `-multi` with count 3; empty → nothing; choosing a task navigates to it. Update the old anchor-id tests to the new ids. Record failing output.
- [x] **Step 8 — Green.** Render `host.ui.TaskRowIndicator` in the issue row `taskIndicator` slot with `taskRowLinks`, no `emptyLabel`.
- [x] **Step 9 — Red: issue toolbar and query commit (FR4.1, FR4.2, BR4.1–BR4.3, R-03, R-04).** Tests: toolbar is `IntegrationListToolbar` with title/count/refresh test ids; typing does not call `issues.list`; Enter commits trimmed query from page 1; blank commit clears keyword; re-committing the same value does not reload; the query input has an accessible name (R-03: `aria-label` passed via the toolbar's placeholder is not enough — wrap the toolbar so the input has `aria-label="Search issues"`, or set it through the fake-verified prop if the host forwards it; if the host does not forward it, document the limitation and assert the placeholder plus a visually hidden label); dropdown change while a draft is dirty commits the draft once and reloads once (R-04: one combined reload, not two). Record failing output.
- [x] **Step 10 — Green.** Replace `createPrToolbar` + `Input` + 400 ms timer in `issues-page.tsx` with `host.ui.IntegrationListToolbar` (`customQuery=draftQuery`, `committedQuery`, `onCommitCustomQuery`), `lastFetchedAt` as `Date | null`; move pagination focus target to the results region (host title takes no ref). Coalesce draft commit with a filter change into one load (R-04).
- [x] **Step 11 — Red: issue filters (FR4.3–FR4.5, BR4.4–BR4.7, BR5.4, R-02, R-06).** Tests: Project/Status/Assignee are `IntegrationRepositoryFilter` with `ariaLabel`, `allLabel` "All projects/statuses/assignees", Status offers "Not closed", Assignee offers "Me"; "" maps to All (R-06); change reloads from page 1; saved/default query sets dropdowns, draft and committed query with one reload; Save query stores the committed query and is the last item in the filter slot; each filter gets GitHub's `triggerClassName` (`w-full … md:w-[220px]`-style, full width on phones, fixed width on desktop) and a `className` for the popover width (R-02). Record failing output.
- [x] **Step 12 — Green.** Render the three filters + Save query in the toolbar `filter` node, wrapped in a container that is `flex w-full flex-col gap-2 md:w-auto md:flex-row md:items-center` (R-02); remove field `Label`s, the `isMobile` "Filters (n)" toggle and its state (BR4.8).
- [x] **Step 13 — Red: PR list tasks and toolbar (FR5, BR1.1–BR1.4, BR5.1–BR5.4, R-05, R-06).** In `ui/src/page/backlog-lists.test.tsx` (or the PR list test file): PR row with linked ids → `backlog-pr-task-<number>-single`/`-multi` with id fallback; toolbar has no query box; Repository/Assignee/Creator are label-less `IntegrationRepositoryFilter` (Assignee/Creator: "Anyone" = All maps to `anyone`, "Me" = `me`, R-06); Status trigger reads "Status (n)", popover shows three checkboxes, toggling reloads from page 1, the last checked status cannot be cleared — its checkbox is `disabled` and stays checked (R-05, one behaviour used everywhere). Update old checkbox/select test ids. Record failing output.
- [x] **Step 14 — Green.** `pr-list.tsx`: `TaskRowIndicator` with `prTaskRowLinks`; new `StatusMultiFilter` (outline trigger "Status (n)", `aria-haspopup="dialog"`, `Popover` + `Checkbox` + `Label`); dropdowns as above with GitHub classes; `pr-toolbar.tsx` restyled to the `IntegrationListToolbar` layout (desktop one row; phones: title row, full-width filters, bottom row with count, last-updated, refresh).
- [x] **Step 15 — Red: Kanban badge (FR2, BR2.1–BR2.3, R-01, R-07, R-08).** In `issue-badge.test.tsx`: openable → `<a>` with `href`, `target="_blank"`, `rel="noopener noreferrer"`, text = key · status; click and pointer-down do not reach the card (R-08); detail reachable by keyboard/touch: the link has `aria-describedby` pointing to a visually hidden detail element and a `title` (R-01); not openable (unavailable / not connected / no or non-https url) → focusable `<span tabIndex=0 role="note">` whose detail (including the reconnect hint) shows on focus/tap as today (R-01, R-07). Record failing output.
- [x] **Step 16 — Green.** Implement in `issue-badge.tsx` using `badgeHref`; remove the open toggle only for the openable state.
- [x] **Step 17 — Red: phone layout and raw-controls guard (FR4.6, BR4.8, BR6.1).** Tests: no `backlog-issues-filters-toggle`; all filters rendered at mobile width; `controls.test.ts` still finds no raw HTML controls and no plugin CSS import. Record failing output (if already green, record that the guard holds and why).
- [x] **Step 18 — Green.** Any remaining fixes; i18n strings added to `ui/src/messages/en.ts` for every new label (no hard-coded text).
- [x] **Step 19 — Refactor (frontend).** Remove dead code (`SEARCH_WAIT`, unused `FIELD`/`Label` imports, old toolbar props), split helpers out of `issues-page.tsx` only where it reduces its size without changing behaviour; keep all tests green.

### Quality gates and build

- [x] **Step 20 — UI gates.** `cd ui && npm run typecheck && npm run lint && npm run format:check && npm test` — all green (NFR4).
- [x] **Step 21 — Go and package.** `~/go-sdk/bin/go test -race ./internal/... ./server/...` unchanged and green; `make package` and `make verify-package` pass (NFR1, BR6.1); delete any `coverage.out` in the repo root.
- [x] **Step 22 — Host contract.** `make contract-test KANDEV_MIN_DIR=../kandev` with the checkout at `v0.96.0`, 10 runs, to prove host components render inside the plugin route (NFR2). If the environment cannot run it, record that in `code-summary.md` for Build and Test.

### Documentation and traceability

- [x] **Step 23 — Records.** Write `code-summary.md`, `source-manifest.json` (every changed source path) and `traceability.json` (FR1–FR5, NFR1–NFR5, BR1.1–BR6.2 → implementation/test files).

### Loop-back 1 fixes (from Build and Test, code review review-01 Minor findings)

- [x] **Step 24 — Red: harness fidelity for blur-before-select (code review R-01, BR4.2, BR4.5).** Make the `IntegrationListToolbar` / `IntegrationRepositoryFilter` fakes reproduce the host order: opening a dropdown blurs the query input first (blur-if-dirty commits), then the selection fires. Rewrite the "dirty draft + dropdown change" test to that order and assert `issues.list` is called exactly once with both the committed keyword and the new filter. Record failing output.
- [x] **Step 25 — Green.** In `issues-page.tsx`, coalesce a commit followed by a filter change in the same interaction into one load (e.g. schedule the load once per render cycle / microtask from a single query state), so the real host order also yields one reload; keep the test title truthful.
- [x] **Step 26 — Red → Green: pagination does not blank the toolbar (code review R-04, FR4.1).** Test that changing page keeps the toolbar count (no "…") and keeps refresh enabled; pass `loading` to `IntegrationListToolbar` only for initial load and refresh, not for page changes.
- [x] **Step 27 — Red → Green: last status stays focusable (code review R-05, BR5.3, NFR3).** Test that the last checked status checkbox is focusable, has `aria-disabled="true"`, stays checked and its toggle is ignored; replace `disabled` with `aria-disabled` plus a guarded handler in `status-multi-filter.tsx`.
- [x] **Step 28 — Gates and records.** Re-run the scoped command and the UI gates (typecheck, lint, format:check, full Vitest); update `code-summary.md`, `source-manifest.json` and `traceability.json` for the changed files.

### Loop-back 2 (rebased onto v0.4.0; SCM pull-request list; release 0.4.1)

Context: the branch now sits on `origin/main` `5724d88` (v0.4.0, PR #9). Merge state: `pr-list.tsx` conflicts merged (provider selector + SCM early return from v0.4.0, this intent's dropdown filters and Status popover); type-check fails in `scm-pr-list.tsx` (removed `taskHref`, `messages.taskLink`, `createPrToolbar` `idPrefix`) and in `pr-list.tsx` (provider selector uses `Label`/`Select*` no longer destructured). Baseline after the rebase: 21 Vitest failures in `src/git/pr-list.test.tsx` (v0.4.0's SCM tests) and `src/page/backlog-lists.test.tsx`.

- [x] **Step 29 — Make the merge compile (Green on existing tests first).** Restore `createPrToolbar` support for an `idPrefix` (default `backlog-prs`) so both PR lists keep their test-id namespaces; make the provider selector in `pr-list.tsx` compile. Run the scoped and full Vitest; record which failures remain and why (expected: tests asserting the old SCM task anchors/labels).
- [x] **Step 30 — Red: SCM PR list like GitHub (FR5.1, FR5.2, BR1.1–BR1.4, BR5.1–BR5.4, NFR3).** In `src/git/pr-list.test.tsx`: SCM rows render host `TaskRowIndicator` (`backlog-scm-pr-task-<number>-single`/`-multi`, fallback = task id); the SCM toolbar uses the same lookalike layout (no query box); repository, saved-query and author filters are label-less searchable dropdowns (`IntegrationRepositoryFilter`, GitHub classes, accessible names; "" = All); SCM status is the same "Status (n)" multi-select popover with the last checked status `aria-disabled`; the provider selector is a label-less dropdown with an accessible name shown first in the filter slot of both lists. Record failing output.
- [x] **Step 31 — Green.** Implement in `scm-pr-list.tsx` (reuse `prTaskRowLinks`, `createStatusMultiFilter`, `FILTER_TRIGGER`/`FILTER_POPOVER`/`FILTERS`) and the provider selector in `pr-list.tsx`; update v0.4.0's existing SCM tests to the new test ids (update, not delete).
- [x] **Step 32 — Refactor** shared dropdown helper between `pr-list.tsx` and `scm-pr-list.tsx` only if it removes duplication without changing behaviour; tests stay green.
- [x] **Step 33 — Version 0.4.1 and release notes.** `manifest.yaml` `version: "0.4.0"` → `"0.4.1"`; README `## Upgrade notes` gains "### 0.4.1: GitHub-style lists" (search on Enter; GitHub-style toolbar and searchable dropdown filters on the issue list and every pull-request list; linked tasks with title and "Tasks (n)" menu; Kanban badge opens the Backlog issue; no data or setting change). Update any test or CI check that pins the manifest version.
- [x] **Step 34 — Gates and records.** Scoped command plus `src/git/pr-list.test.tsx`; UI gates (typecheck, lint, format:check, full Vitest); `make check-format vet lint coverage package verify-package`; contract test ×10 on Kandev v0.96.0; update `code-summary.md` ("Loop-back 2"), `source-manifest.json` (add `scm-pr-list.tsx`, `pr-list.test.tsx`, `manifest.yaml`, `README.md` and any other changed path), `traceability.json`.

## Traceability (plan step → requirement)

| Requirement / rule | Steps |
|---|---|
| FR1 / BR1.1–BR1.5 | 3–8 |
| FR2 / BR2.1–BR2.3 | 3–5, 15–16 |
| FR3 / BR3.1 | 7–8 (existing title-link test kept green) |
| FR4 / BR4.1–BR4.8 | 9–12, 17–18 |
| FR5 / BR5.1–BR5.4 | 13–14 |
| NFR1 / BR6.1 | 17, 21 |
| NFR2 | 22 |
| NFR3 | 9, 11, 13, 15 |
| NFR4 | 20–21 |
| NFR5 / BR6.2 | 21 |
| Review R-01..R-08 | R-01: 15–16 · R-02: 11–12 · R-03: 9 · R-04: 9–10 · R-05: 13 · R-06: 11, 13 · R-07: 3, 15 · R-08: 15 |
| Code review (code-generation review-01) R-01, R-04, R-05 | R-01: 24–25 · R-04: 26 · R-05: 27 · gates: 28 |
| Loop-back 2: v0.4.0 merge, SCM PR list (FR5 extended), release 0.4.1 | merge: 29 · SCM list: 30–32 · version/notes: 33 · gates: 34 |
