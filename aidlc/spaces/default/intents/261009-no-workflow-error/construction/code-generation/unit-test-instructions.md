# Unit Test Instructions — 261009-no-workflow-error

## Framework Setup

- **Go**: `testing` + `github.com/stretchr/testify/require`, table-driven with `t.Run`, always `-race` (needs CGO). The Kandev SDK resolves through the `replace` to `../kandev` (pinned v0.96.0). If Go is missing, install Go 1.26.x to `~/.local/go` (checksum-verified) and put `~/.local/go/bin` on `PATH`.
- **UI**: Vitest in `ui/` (existing config). Run `npm ci` in `ui/` once if `node_modules` is missing. Host APIs (`host.api.invokeAction`, `host.toast`, `host.context`, `TaskCreateDialog`) are test doubles from the existing test helpers.

## How to Run This Change's Tests

Runnable before the first Red step (brownfield runners already exist):

```bash
# Backend: host workflow read, workflows.status action, manifest
go test -race ./internal/plugin/ -run 'HasWorkflow|WorkflowsStatus|Manifest'

# Frontend: "+ Task" flow and the shared alert
cd ui && npx vitest run src/page/start-task.test.tsx src/ui/error-alert.test.tsx
```

Backlog-page surfaces retouched in Steps 14-15 (run the exact files changed; list them in code-summary.md):

```bash
cd ui && npx vitest run src/issues/issues-page.test.tsx src/issues/issue-panel.test.tsx src/issues/link-task-dialog.test.tsx src/issues/issue-prs.test.tsx src/issues/poll-interval.test.tsx src/git/pr-list.test.tsx src/git/scm-pr-list.test.tsx src/git/save-query-dialog.test.tsx src/git/watch-form.test.tsx src/git/scm-watch-form.test.tsx src/git/git-access.test.tsx src/git/review-provider.test.tsx src/switch/integration-switch.test.tsx
```

Only files that exist are passed; a surface without a test file gets a new one at the same path.

## Expected Tests (Minimal strategy + bugfix regression, ~15)

| ID | Test | Requirement |
|---|---|---|
| T1 | **Regression**: context `null`, workspace has a workflow → dialog opens with `workflowId: null`, no error | FR1.2 |
| T2 | context `null`, no workflow → no dialog, `toast.error(errorWorkflow)` | FR1.3 |
| T3 | context `null`, check fails → dialog opens with `workflowId: null` | FR1.4 |
| T4 | context present → dialog uses context; `workflows.status` not called | FR1.1, FR3.4 |
| T5 | task from null-workflow dialog → link succeeds, `onLinked` called | FR1.5 |
| T6 | link fails → `toast.error(taskNotLinked)`, no inline notice | FR1.5, FR2.1 |
| T7 | PR row follows the same null-context path | FR1.6 |
| T8 | `HasWorkflow` true / false / host error / host not ready, `Limit == 1` | FR3.2, FR3.3 |
| T9 | `workflows.status` handler result and error mapping without host text | FR3.2, FR3.3, NFR4 |
| T10 | manifest declares `api_read` workflows and the `workflows.status` action | FR3.1 |
| T11 | shared alert: `role="alert"`, full-width wrapping classes, action slot, empty → nothing | FR2.2-FR2.4, NFR1, NFR3 |
| T12+ | one test per retouched Backlog-page surface asserting its mapped style | FR2.1-FR2.5 |

## Coverage Targets

- Go: the team floor of 80% line coverage for `./internal/...` and `./server/...` stays met (`make coverage`); the new code is fully covered by T8-T10. Delete `coverage.out` after a local run or write it under `build/`.
- UI: every new branch in `start-task.tsx` and the alert component is exercised by T1-T7 and T11.

## Mocking / Stubbing

- Go: extend the existing fake `pluginsdk.Host` used by `internal/plugin` tests with a `Workflows()` reader returning configured workflows or an error; inject host readiness the same way existing `waitHost` tests do (no real sleeps).
- UI: stub `host.api.invokeAction` per action key (`workflows.status`, `issues.link`, `scm.prs.link`); stub `host.context.getTaskCreationContext` to return `null` or a context; spy on `host.toast.error`; the `TaskCreateDialog` double records its props (`workflowId`, `steps`, `defaultStepId`, `initialValues`).

## Test Data

- Inline fixtures only (workspace id, one workflow `{id: "wf-1"}`, one step). No real credentials or Backlog data (project rule).
