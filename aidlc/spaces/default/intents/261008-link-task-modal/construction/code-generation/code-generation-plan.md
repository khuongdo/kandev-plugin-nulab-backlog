# Code Generation Plan — Link Task modal, GitHub-style

Intent `261008-link-task-modal`, scope bugfix, zero-Unit (stage-level). Source of requirements: `inception/requirements-analysis/requirements.md` (FR1-FR4, NFR1-NFR4). Code map: `aidlc/spaces/default/codekb/kandev-plugin-nulab-backlog/`.

## Scope

UI-only change in `ui/src/`. No Go or backend contract change: `issues.link` still takes `{issueKey}`.

| Area | Files |
|---|---|
| Task-side "Link Backlog issue" action (FR1-FR3) | new `ui/src/issues/issue-link.ts`, new `ui/src/issues/issue-link.test.ts`, `ui/src/index.ts` (register), `ui/src/index.test.ts` (registration assertion, if it lists task actions) |
| Issue-side dialog restyle (FR4) | `ui/src/issues/link-task-dialog.tsx`, `ui/src/issues/link-task-dialog.test.tsx`, `ui/src/issues/issues-page.tsx` (only if needed to pass the links store / toast) |
| Copy (NFR3) | `ui/src/messages/en.ts` |
| Test harness (only if a host member is missing) | `ui/src/testing/harness.ts`, `ui/src/host-ui.ts` (e.g. `DialogDescription`) |

## Decisions on the requirements review findings

- **R-01 (visibility before the store loads)**: same as the Unlink item. `visible` starts `store.load(workspaceId)` and returns `false` until the store has the workspace. Once loaded, it returns `true` only when the store has no link for the task. In practice the store is already loaded, because the issue badge on every task card and row loads it. No extra integration-on or connected check, same as "Link Backlog pull request". A not-connected workspace gets the inline "connect Backlog first" message from FR1.5.
- **R-02 (lowercase key)**: the UI trims and upper-cases the key (`proj-123` becomes `PROJ-123`). It then validates against `^[A-Z][A-Z0-9_]*-[1-9][0-9]*$`. Anything else, including a non-matching URL, shows the inline "not a Backlog issue key or link" message without calling the backend.
- **R-03 (error mapping)**: client-side format check first, as in R-02. Backend failure `not_found` maps to "Issue {key} was not found."; `conflict` maps to "This task is already linked to another Backlog issue. Unlink it first."; `validation` with `field=issueKey` maps to "{project} is not one of the projects selected for this workspace."; anything else goes through the existing `issueNotice` / `noticeText` mapping (not connected, reconnect, rate limit). New catalogue messages are added to `en.ts`, and no response body is shown (NFR1).
- **R-04 (refresh placement)**: `onSubmit` awaits `issues.link`, then calls `store.refresh(workspaceId)` inside a `try/catch` that swallows refresh errors, so a successful link never becomes an inline error. The host shows the success toast and closes.
- **R-05**: `singleTaskOnly: true`, same as the PR action.
- **R-06**: tests that select "Link"/"Linking..." are updated to "Save"/"Saving..." in the same Red step that asserts the new labels.

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

Layers not applicable to this change: data model / database, repository / data access, and API / endpoint (the backend is unchanged). The testable layers are **business logic** (the pure reference parser) and **frontend behavior** (the task action and the dialog).

## Steps

- [x] **Step 1: Project structure.** No new configuration. The new module `ui/src/issues/issue-link.ts` sits next to `task-menu.ts`, following the `ui/src/git/pr-link.ts` pattern.
- [x] **Step 2: Test runner readiness.** Link `../kandev` to the pinned v0.96.0 checkout (`ln -s ~/repo/kandev ../kandev`, only if absent and only if `~/repo/kandev` is at `v0.96.0`), run `npm ci` in `ui/`, then run the existing link tests with `cd ui && npx vitest run src/issues/link-task-dialog.test.tsx src/issues/task-menu.test.ts src/git/pr-link.test.ts` and record the baseline (pass / fail counts). Then run the full `cd ui && npx vitest run` once and record the baseline count.
- [x] **Step 3: Business logic, Red (FR2, R-02).** In `issue-link.test.ts`, write failing tests for `parseIssueReference(raw)`: key `PROJ-123`, lowercase `proj-123` becomes `PROJ-123`, padded whitespace, URL `https://acme.backlog.com/view/PROJ-123` (also `.backlog.jp`, `.backlogtool.com`, and trailing `#comment-1` / `?x=1`), rejected `http://` URL, rejected foreign host `https://example.com/view/PROJ-123`, rejected garbage, rejected empty. Run and record the failing output.
- [x] **Step 4: Business logic, Green.** Implement `parseIssueReference` in `issue-link.ts`. It returns the key or `undefined`; the URL check uses `new URL()`, an `https:` protocol and a host suffix in the allowed set (project Mandated rule).
- [x] **Step 5: Business logic, Refactor.** Tidy names and doc comments while green.
- [x] **Step 6: Frontend (task action), Red (FR1, FR3, R-01, R-03, R-04, R-05).** In `issue-link.test.ts`, using the shared fake host (`ui/src/testing/harness.ts`), write failing tests modelled on `pr-link.test.ts`:
  - action shape: id, label "Link Backlog issue", `placement: "link"`, `singleTaskOnly: true`;
  - `run` opens `openTaskLinkDialog` with the catalogue copy and test ids `backlog-link-issue-input` / `-error` / `-submit`;
  - `onSubmit("https://acme.backlog.com/view/PROJ-123", signal)` calls `issues.link` with `{workspaceId, taskId, body: {issueKey: "PROJ-123"}}` and the signal, then refreshes the store;
  - a store refresh failure after a successful link does not throw;
  - an invalid reference throws the "not a Backlog issue key or link" message and makes no backend call;
  - `not_found`, `conflict` and `validation` (`field=issueKey`) map to their messages; other failures use `issueNotice`;
  - `visible`: false and starts a load when the store has no data for the workspace, false when the task is linked, true when the task is unlinked.

  Run and record the failing output.
- [x] **Step 7: Frontend (task action), Green.** Implement `createIssueLinkAction(host, store, messages)` in `issue-link.ts`, add its messages to `en.ts`, and register it in `ui/src/index.ts` next to `createPRLinkAction`, passing the shared `links` store. Update `index.test.ts` if it asserts the registered task actions.
- [x] **Step 8: Frontend (task action), Refactor.** While green, share the error-mapping shape with `pr-link.ts` only if it removes duplication without a new abstraction.
- [x] **Step 9: Frontend (issue-side dialog), Red (FR4, R-06).** Update `link-task-dialog.test.tsx` with failing assertions for:
  - the description line, and `DialogContent` class `w-[calc(100vw-2rem)] sm:max-w-lg`;
  - submit labelled "Save", and "Saving..." while pending (old "Link"/"Linking..." selectors updated);
  - Enter in the search field submits when a task is chosen, and does nothing when none is chosen;
  - an inline error with `role="alert"` and class `text-xs text-destructive`;
  - on success, a success toast, a links-store refresh for the workspace, `onLinked`, then `onClose`.

  Keep the existing cases (AC3.3.1-AC3.3.4, AC8.2.3, axe, catalogue-only) and the `backlog-link-task-*` test ids. Run and record the failing output.
- [x] **Step 10: Frontend (issue-side dialog), Green.** Restyle `link-task-dialog.tsx`:
  - wrap the content in a `<form>` (submit = link, `preventDefault`) and add the `DialogDescription`;
  - set the width class on `DialogContent` and style the error inline;
  - rename the submit to Save/Saving...;
  - call `host.toast.success` and refresh the links store (passed in from `issues-page.tsx` or the factory, whichever already has it).

  Add `DialogDescription` to `host-ui.ts` and the harness if they are missing. Add the new messages to `en.ts`.
- [x] **Step 11: Frontend (issue-side dialog), Refactor.** Remove now-dead code, such as the old error paragraph styling, while green.
- [x] **Step 12: Regression and full suite.** Run the unit-scoped command from `unit-test-instructions.md`, then the full `cd ui && npx vitest run`, `npx tsc --noEmit`, `npx eslint .` and `npx prettier --check .` in `ui/`. All must be green; compare against the Step 2 baseline.
- [x] **Step 13: Environment/build configuration.** None expected. `cd ui && npm run build` must still bundle. Do not commit `build/` output or `coverage.out`.
- [x] **Step 14: Documentation and traceability.** Update doc comments: `link-task-dialog.tsx`'s comment about "Kandev has no task picker", and a factory doc on `createIssueLinkAction` citing FR1-FR3. Add a README note only if the README lists task-menu actions.

## Requirement-to-step traceability

| Requirement | Steps |
|---|---|
| FR1.1-FR1.5 | 6, 7, 8 |
| FR2.1-FR2.3 | 3, 4, 5 |
| FR3.1-FR3.2 | 6, 7 |
| FR4.1-FR4.6 | 9, 10, 11 |
| NFR1 (no secrets, https-only hosts) | 4, 6 |
| NFR2 (accessibility, axe) | 9 |
| NFR3 (catalogue text) | 7, 10 |
| NFR4 (no new polling) | 6, 7 |
