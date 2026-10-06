## Review

**Verdict:** READY
**Reviewer:** aidlc-architecture-reviewer-agent
**Date:** 2026-10-06T08:44:31Z
**Iteration:** 2

### Findings

| ID | Severity | Location | Finding | Required action | Status |
|---|---|---|---|---|---|
| R-01 | Major | internal/connection/service.go > Service.SetEnabled | Fixed. SetEnabled does one LoadSwitch (a failed read only logs "unknown"), one SaveSwitch and no secret or record access, then returns SwitchView{Enabled} with nil error once the write succeeded. Tests cover the secret-read-failure case (200) and the exact store-call sequence. | None. | Resolved |
| R-02 | Major | ui/src/settings/SettingsScreen.tsx; ui/src/switch/enabled-events.ts | Fixed. publishEnabled calls host.setIntegrationEnabled and then notifies local listeners. SettingsScreen subscribes per workspace and dispatches enabledChanged, which patches only view.enabled (BR7.4), so no reload is needed. A failed save publishes nothing. A mounted-together test covers Off to On and On to Off. | None. | Resolved |
| R-03 | Minor | ui/src/page/BacklogPage.tsx > reload | Fixed. A sequence counter ignores late replies and late failures. The effect cleanup bumps the counter, which also covers a Retry that overlaps a load. Out-of-order tests exist. | None. | Resolved |
| R-04 | Minor | internal/plugin/runtime.go guard; internal/connection/service.go preCall/connect | Connect still reads the switch three times (guard, preCall, second read before the write). The time budget still holds. Deferred by user choice. | Remove the preCall read or record the third read in the design. | Unresolved |
| R-05 | Minor | code-generation-plan.md Step 1, Step 13 | Plan steps still mention connection.connectApiKey and plugin-package-verify. Deferred by user choice. | Mark superseded or correct. | Unresolved |
| R-06 | Minor | code-generation-plan.md Step 13; .github/workflows/ci.yml | The contract test on min_kandev_version is neither in CI nor assigned to a unit. Deferred by user choice. | Assign an owner unit or record it as deferred. | Unresolved |
| R-07 | Minor | functional-design/functional-spec.md WF6; functional-design/entities.md ConnectionView; contract C5 | The upstream contracts still say set_enabled returns the full ConnectionView, while the code returns {enabled}. | Amend WF6, entities.md and C5 to say set_enabled returns {enabled}. | New |

### Validation Tool Results

| Tool | Result | Interpretation |
|---|---|---|
| gofmt -l . | clean | No formatting issues. |
| go vet ./... | clean | No findings. |
| go test -race -cover ./internal/... ./server/... | all ok (connection 97.5%, plugin 91.4%) | The floor of 80% holds. |
| tsc --noEmit | clean | Strict types pass. |
| eslint . | clean | No violations. |
| prettier --check . | clean | Formatting passes. |

### Spec note: set_enabled returns {enabled}

This is the right resolution. A full ConnectionView needs the record and the secret (BR2.11), which NFR1.3 forbids. A post-write read can also fail after the switch was saved, which contradicts WF6 step 6. The UI is consistent with it: the card switch ignores the reply body and uses its own value, and the settings screen and the page learn the value from the published event or from connection.get. The only remaining work is to amend the contracts (R-07). It is not blocking, because no consumer in U1 reads the extra fields.

### Residual notes

A settings-screen load that was issued before a publish and resolves after it could overwrite the new value with an older one. The window is negligible in practice, because a load happens at mount and a save needs user action.

### Summary

All three fixes are verified in the code and tests, and no regression was found. The Go and UI checks are clean. The three Minor findings the user deferred stay Unresolved, and one new Minor finding records the spec amendment for the new set_enabled reply.
