# Code Generation Plan — 261009-no-workflow-error

Zero-Unit bugfix (scope `bugfix`, Minimal strategy, Brownfield). Source of scope: `inception/requirements-analysis/requirements.md` (FR1-FR3, NFR1-NFR6) and the code knowledge base `aidlc/spaces/default/codekb/kandev-plugin-nulab-backlog/` (architecture.md § Task Creation Context).

## Overview

- **Backend (Go, `internal/plugin`)**: a new workspace action `workflows.status` answers `{"hasWorkflow": bool}` from `Host.Workflows().List(ctx, ws, Page{Limit: 1})`; the manifest gains `api_read: ["tasks", "repositories", "workflows"]` and the action entry.
- **Frontend (`ui/src/page/start-task.tsx`)**: when `getTaskCreationContext` is `null`, call `workflows.status`; `hasWorkflow: true` or a failed check opens Kandev's `TaskCreateDialog` with `workflowId: null` (no steps); `hasWorkflow: false` shows `host.toast.error(messages.errorWorkflow)`. Link failures use `host.toast.error(messages.taskNotLinked)`. The inline notice span is removed.
- **Frontend error retouch (Backlog page)**: one shared inline alert (`role="alert"`, full width, wrapping, Kandev destructive tokens) for page-state and in-dialog errors; toasts for click-caused errors, per the mapping table produced in Step 13.

Blast radius: medium — `start-task.tsx` is used by issue rows and every PR row; the alert component touches every Backlog-page error surface listed in FR2.5. No data model, storage or Backlog API change.

## Decisions

- `workflows.status` counts every workflow the Host data API returns for the workspace (hidden ones included); the plugin does not filter. If Kandev's dialog then picks a hidden workflow, that is Kandev's choice — the real-host check in Build and Test (Step 21) verifies the dialog lists a usable workflow. (Review R-02)
- FR1.4 fail-open: a failed `workflows.status` call opens the dialog with `workflowId: null`, because the bug being fixed is a false "no workflow" error; Kandev's dialog still validates on submit. (Review R-01)
- NFR1 applies to plugin-rendered inline alerts only; Kandev toasts are out of the plugin's control. (Review R-04)
- FR2.4 shared alert component is design intent, not a user requirement. (Review R-06)
- The action name `workflows.status` matches Kandev's key pattern `^[a-z0-9][a-z0-9._-]*$`.

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
  "input_sha256": "sha256:d93834fe9ce9a29c231b958ad21f7e8cc8cb2b175797341155dc75acc12e179a",
  "contract_sha256": "sha256:dd18054e773370e0612b5481933c9e0620001f7476a4eff866598e9142d67dd0"
}
```

Layers not applicable: Data model / database behavior (no data model change). Repository / data access is the Host workflow read (`hostPort`). Business logic has no separate layer beyond the read (it is a boolean), so it is folded into the API/endpoint layer.

## Steps

- [x] **Step 1 — Project structure (no change).** Confirm no new package is needed: backend code stays in `internal/plugin` (only package allowed to import `pluginsdk`); UI code in `ui/src/page/` and a shared alert next to the existing shared UI helpers (`ui/src/host-ui.ts` neighbourhood). Traces: FR3.3, team Code Style.
- [x] **Step 2 — Verify runners.** Run `go test -race ./internal/plugin/ -run 'Workflow|Manifest'` and, in `ui/`, `npx vitest run src/page/start-task.test.tsx`; record both commands and current results in `unit-test-instructions.md` terms. Traces: NFR5.
- [x] **Step 3 — Repository / data access — Red.** In `internal/plugin/host_port_test.go` (or the existing host-fake test file), add table tests for `hostPort.HasWorkflow(ctx, ws)`: one workflow → true; zero workflows → false; host `List` error → wrapped error; host not ready within the bounded wait → error; asserts `Page.Limit == 1` and the workspace id passed. Record the failing output. Traces: FR3.2, FR3.3.
- [x] **Step 4 — Repository / data access — Green.** Add `func (p hostPort) HasWorkflow(ctx context.Context, ws string) (bool, error)` using `waitHost` and `h.Workflows().List(ctx, ws, pluginsdk.Page{Limit: 1})`, wrapping errors with `fmt.Errorf("list workflows: %w", err)`. Extend the host fake used by tests if it lacks `Workflows()`. Traces: FR3.2.
- [x] **Step 5 — Repository / data access — Refactor.** Update the `hostPort` doc comment to `api_read: [tasks, repositories, workflows]`; keep tests green.
- [x] **Step 6 — API / endpoint — Red.** Add tests: (a) the `workflows.status` handler returns `{"hasWorkflow": true}` / `false` for the workspace from the verified action context and maps a host error to the plugin's existing host-unavailable error code without host error text (NFR4); (b) `manifest_test.go` expects `api_read` to equal `["tasks", "repositories", "workflows"]` and a `workflows.status` action with `scope: workspace`, `access: authenticated`, `max_body_bytes: 8192`. Record the failing output. Traces: FR3.1, FR3.2, FR3.3, NFR4.
- [x] **Step 7 — API / endpoint — Green.** Register `workflows.status` in the handlers map (new small file `internal/plugin/workflow_actions.go` following `issue_actions.go`'s `init() { maps.Copy(handlers, ...) }` pattern); add the capability and action entry to `manifest.yaml`. Leave `internal/plugin/testdata/v030/manifest.yaml` unchanged (frozen snapshot). Traces: FR3.1, FR3.2.
- [x] **Step 8 — API / endpoint — Refactor.** Keep the handler a few lines; run `go test -race ./internal/plugin/` green.
- [x] **Step 9 — Frontend behavior ("+ Task") — Red.** In `ui/src/page/start-task.test.tsx`, replace the test that asserts the current inline error (`start-task.test.tsx:132-147`) and add: (1) **regression** — context `null`, `workflows.status` → `{hasWorkflow: true}`: selecting a quick action opens `TaskCreateDialog` with `workflowId: null`, no `steps`, prefilled title/description, and no error; (2) context `null`, `{hasWorkflow: false}`: no dialog, `host.toast.error` called with `messages.errorWorkflow`; (3) context `null`, action rejects: dialog opens with `workflowId: null` (FR1.4); (4) context present: dialog opens with context `workflowId`/`defaultStepId`/`steps` and `workflows.status` is not called (FR1.1, FR3.4); (5) created via the `workflowId: null` dialog, `issues.link` succeeds → `onLinked` called (FR1.5, review R-05); (6) link fails → `host.toast.error(messages.taskNotLinked)` and no inline notice element (FR2.1); (7) a PR row (`kind: "pr"`, `linkAction: "scm.prs.link"`) follows the same null-context path (FR1.6). Record the failing output. Traces: FR1.1-FR1.6, FR2.1, FR3.4.
- [x] **Step 10 — Frontend behavior ("+ Task") — Green.** Change `select` in `ui/src/page/start-task.tsx` to the FR1 flow (async; ignore re-entrant clicks while the check runs); open state holds an optional context; render `TaskCreateDialog` with `workflowId={ctx?.workflowId ?? null}` and omit `defaultStepId`/`steps` when there is no context; replace `setNotice` with `host.toast.error`; remove the inline notice span. Adjust the `TaskCreateDialog` test double / host-ui typing if `workflowId` must accept `null`. Traces: FR1, FR2.1.
- [x] **Step 11 — Frontend behavior ("+ Task") — Refactor.** Simplify while green; keep `data-testid`s on interactive elements.
- [x] **Step 12 — Frontend behavior (shared inline alert) — Red.** Add `ui/src/ui/error-alert.test.tsx` (or beside the existing shared UI helpers): renders the message with `role="alert"`, full-width and wrapping classes (`w-full`, `break-words`/`whitespace-normal`), destructive token classes, an optional action slot (retry/reconnect) kept inside the alert, and nothing when the message is empty. Record the failing output. Traces: FR2.2-FR2.4, NFR1, NFR3.
- [x] **Step 13 — Frontend behavior (shared inline alert) — Green.** Implement the alert component. Inventory every error display under the Backlog page (`ui/src/page/`, `ui/src/issues/`, `ui/src/git/`, `ui/src/switch/`; Settings excluded) and write the **error-to-style mapping table** (message key in `ui/src/messages/en.ts` → toast / page alert / dialog alert) into `code-summary.md`. Rule: click-caused errors outside dialogs → toast; page-state errors → page alert at the top of the section; errors inside a dialog → dialog alert above the dialog actions. Traces: FR2.1-FR2.5, review R-03.
- [x] **Step 14 — Frontend behavior (Backlog-page surfaces) — Red.** For each surface in the mapping table, update or add one test in its existing test file asserting the chosen style (toast call, or the shared alert's `role="alert"` within the section/dialog) — e.g. `issues-page`, `issue-panel`, `pr-list`, `scm-pr-list`, `link-task-dialog`, `save-query-dialog`, `watch-form`, `scm-watch-form`, `git-access`, `review-provider`, `issue-prs`, `poll-interval`, `integration-switch`. Record the failing output. Traces: FR2.1-FR2.5.
- [x] **Step 15 — Frontend behavior (Backlog-page surfaces) — Green.** Replace each surface's ad-hoc error markup with the shared alert or `host.toast.error` per the table; keep message texts from `messages/en.ts` (FR2.6) and remove row-level inline error text that squeezed layouts. Watch-form workflow logic is not changed (out of scope). Traces: FR2.
- [x] **Step 16 — Frontend behavior — Refactor.** Remove now-unused classes/state; `tsc --noEmit`, ESLint and Prettier clean; Vitest for touched files green.
- [x] **Step 17 — Environment/build configuration.** No build config change expected; run `make check-format vet lint` and `cd ui && npm run typecheck && npm run lint` (or the existing equivalents) to confirm. Traces: team Code Style.
- [x] **Step 18 — Documentation.** README: add a v-next upgrade note that the plugin now reads workflows (`api_read: workflows`) and the admin may be asked to re-approve permissions; mention the "+ Task" fix and error-display retouch. Traces: Constraints (Q7), FR3.1.
- [x] **Step 19 — Traceability.** Write `code-summary.md` (files, decisions, the mapping table, test results), `source-manifest.json`, and `traceability.json` covering FR1-FR3 sub-IDs and NFR1-NFR6.
- [x] **Step 20 — Full suites (existing green).** `go test -race ./...` and `cd ui && npx vitest run`; keep the Go 80% coverage floor (`make coverage`, coverage profile outside repo root).
- [ ] **Step 21 — Real-host check (handed to Build and Test, mandatory before release).** Packaged-host contract test on Kandev v0.96.0 (`make contract-test KANDEV_MIN_DIR=../kandev`, 10 runs) must pass with the new capability; plus a manual check on real Kandev v0.96.0: open `/backlog` directly (reload), click "+ Task" → Kandev's dialog opens with a usable workflow and the created task links to the issue; check error displays at 320/360/768/1280 px. Traces: NFR1, NFR2, NFR6, review R-02.
