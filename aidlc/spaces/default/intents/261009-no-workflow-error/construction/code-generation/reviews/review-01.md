## Review

**Verdict:** READY
**Reviewer:** aidlc-architecture-reviewer-agent
**Date:** 2026-10-09T10:39:11Z
**Iteration:** 1

### Findings

| ID | Severity | Location | Finding | Required action | Status |
|---|---|---|---|---|---|
| R-01 | Minor | aidlc/spaces/default/intents/261009-no-workflow-error/construction/code-generation/code-generation-plan.md > Step 6(a) vs internal/plugin/workflow_actions.go and workflow_actions_test.go > TestWorkflowsStatus_HostErrorLeaksNoText | The plan says a host error maps to "the plugin's existing host-unavailable error code". The implementation returns the raw error and the runtime dispatcher turns it into a generic 500 `{"error":{"code":"internal"}}` (runtime.go:381). The test asserts `internal`. NFR4 holds (no host text leaks), but the deviation from the plan is not listed under "Deviations from the Plan" in code-summary.md. | Record the deviation in code-summary.md (generic `internal` code, which the UI treats as fail-open per FR1.4), or map the error to the existing unavailable code. | New |
| R-02 | Minor | ui/src/page/start-task.tsx > select / checking ref | While `workflows.status` is pending (NFR2 allows up to 1 s), the menu gives no feedback, and re-entrant picks are silently dropped. `setOpen` also runs after the await, so it can fire on an unmounted row. Neither is a correctness bug, since the dialog still opens or the toast still shows. | Optionally disable the trigger or show a pending state during the check. Otherwise accept the risk. | New |
| R-03 | Minor | aidlc/spaces/default/intents/261009-no-workflow-error/construction/code-generation/code-generation-plan.md > Step 21 | The plan leaves Step 21 open. It covers the real-host check that `TaskCreateDialog` with `workflowId: null` selects a usable workflow (and may pick a hidden workflow), plus the 320/360/768/1280 px check. This is the only evidence that the root-cause fix works on real Kandev, because the unit tests use a test double for the dialog. code-summary.md reports that the contract test x10 passed. The manual check remains. | Build and Test must run the manual real-Kandev v0.96.0 check before release. Do not tag without it. | New |

### Validation Tool Results

| Tool | Result | Interpretation |
|---|---|---|
| go test -race ./internal/plugin/ | PASS (cached) | Includes HasWorkflow, action, manifest and host-error tests. |
| vitest run (ui) | PASS, 462/462 | The "+ Task" regression and the alert tests are included. |
| tsc --noEmit (ui) | no errors printed | Typing of `workflowId: null` is accepted. |
| coverage.out in repo root | absent | Complies with the project rule on coverage profiles. |

### Summary

The root cause is fixed where it occurs. When `getTaskCreationContext` is `null`, `start-task.tsx` calls the new read-only `workflows.status` action (`HasWorkflow`, page size 1, bounded host wait). It then opens `TaskCreateDialog` with `workflowId: null`, shows a toast when there is no workflow, and fails open on a check error (FR1.2-FR1.4). The manifest adds `api_read: workflows` and the action with `access: authenticated`. The tests assert the regression path and a boolean-only response, and the manifest and error-leak tests carry real assertions (NFR4/NFR5). The recorded deviations are sound: watch-form, git-access and poll-interval are only imported by Settings, which FR2.5 excludes. Only minor items remain, chiefly the open real-host check.

READY
