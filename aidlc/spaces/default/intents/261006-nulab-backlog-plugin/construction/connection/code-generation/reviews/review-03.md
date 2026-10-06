## Review

**Verdict:** READY
**Reviewer:** aidlc-architecture-reviewer-agent
**Date:** 2026-10-06T22:18:12Z
**Iteration:** 1

### Findings

| ID | Severity | Location | Finding | Required action | Status |
|---|---|---|---|---|---|
| R-01 | Major | internal/connection OAuth start/callback (state binding) | OAuth state is not bound to the browser that started the flow. Code unchanged since the last review (git diff shows no change under internal/connection). | Bind state to the starting browser, or record an accepted risk, in Build and Test. | Unresolved |
| R-02 | Major | ui/src/settings/project-picker.tsx | ProjectPicker is not reloaded after an API-key replace to another space. U3 changed only the impact-count call in this file. | Reload the picker after a replace that changes space. | Unresolved |
| R-03 | Minor | ui/src/settings/project-picker.tsx selection handling | A reordered selection bumps the epoch. | Compare selections as sets. | Unresolved |
| R-04 | Minor | internal/connection viewOf | viewOf never sets HasGitCredential. | Set it from the stored Git credential. | Unresolved |
| R-05 | Minor | internal/connection space-change path | deleteGit runs before the write on a space change. | Write first, or make the pair recoverable. | Unresolved |
| R-06 | Minor | ui/src/settings/connected-panel.tsx | No Disconnect action in the sign_in_again or error states. | Offer Disconnect in those states. | Unresolved |
| R-07 | Minor | internal/connection refresh | refresh runs on the caller ctx. | Run refresh on a detached context with a timeout. | Unresolved |
| R-08 | Minor | internal/connection OAuth callback | The callback holds the workspace lock through the code exchange. | Release the lock around the network call. | Unresolved |
| R-09 | Minor | internal/backlog rate-limit parsing | X-RateLimit-Reset is unbounded, and an HTTP-date Retry-After is ignored. | Cap the wait and parse the HTTP-date form. | Unresolved |
| R-10 | Minor | internal/connection refresh error mapping | ErrOAuthNotConfigured during refresh is reported as validation. | Map it to its own outcome. | Unresolved |
| R-11 | Minor | ui/src/issues/issues-state.ts loadImpactText (used by ui/src/settings/connected-panel.tsx, SettingsScreen.tsx, project-picker.tsx) | Each failed impact call (issues.impact or git.impact) counts as 0 and is swallowed. If both fail, the Disconnect, change-space and project-removal dialogs show no impact warning, and the user cannot tell that from "nothing affected". The confirm dialogs still disconnect, replace and change space correctly. | Show an "impact unknown" line when either call fails. | New |

### Validation Tool Results

| Tool | Result | Interpretation |
|---|---|---|
| git diff c3ca400 (settings UI, connection, client.go, runtime.go) | internal/connection has no changes. U3 edits are limited to the impact helper swap, a PollInterval mount, a children slot in confirm-dialog and the added "a[href]" focus-trap selector, and the removal of the unused Client.Issue. | U2 behaviour is preserved. The three-choice dialog keeps its handlers. The impact text is only informational. |
| go vet ./... | PASS (no output) | Clean. |
| go test -race -cover ./internal/... ./server/... | PASS. connection 94.5%, backlog 96.0%, plugin 93.6%. server is wiring only (0%). | Above the 80% floor. |
| git status --short before and after | Identical | Workspace untouched. |

### Summary

U3's edits do not break U2. internal/connection is unchanged, the dialogs still act correctly, and vet and race tests pass. The two prior Major findings (R-01, R-02) remain open and were deferred to Build and Test. The only new item is a Minor one: impact-load failures are silently shown as zero.
