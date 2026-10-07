## Review

**Verdict:** READY
**Reviewer:** aidlc-architecture-reviewer-agent
**Date:** 2026-10-07T00:00:57Z
**Iteration:** 1

### Findings

| ID | Severity | Location | Finding | Required action | Status |
|---|---|---|---|---|---|
| R-01 | Critical | internal/connection/oauth.go > StartOAuth, completeOAuth | OAuth sign-in is bound to the browser verifier. Re-checked: StartOAuth rejects non-hex hashes and sha256("") (emptyVerifierHash) before SavePending. completeOAuth returns bad_state on an empty verifier before taking the lock, so the pending record is kept. | None | Resolved |
| R-02 | Major | internal/connection (space picker) | The picker is keyed on spaceHost:connectedUserName. Code is unchanged since the previous review. | None | Resolved |
| R-03 | Minor | internal/connection selection | A reordered selection bumps the epoch. | Optional follow-up | Unresolved |
| R-04 | Minor | internal/connection | HasGitCredential is missing. | Optional follow-up | Unresolved |
| R-05 | Minor | internal/connection | deleteGit runs before the write. | Optional follow-up | Unresolved |
| R-06 | Minor | internal/connection UI states | sign_in_again and error states have no Disconnect action. | Optional follow-up | Unresolved |
| R-07 | Minor | internal/connection refresh | Refresh runs on the caller ctx. | Optional follow-up | Unresolved |
| R-08 | Minor | internal/connection oauth.go | The workspace lock is held through the token exchange. | Optional follow-up | Unresolved |
| R-09 | Minor | internal/backlog rate limit | X-RateLimit-Reset is unbounded and an HTTP-date Retry-After is ignored. | Optional follow-up | Unresolved |
| R-10 | Minor | internal/connection | ErrOAuthNotConfigured is reported as a validation error. | Optional follow-up | Unresolved |
| R-11 | Minor | internal/connection impact | Impact errors are shown as 0. | Optional follow-up | Unresolved |
| R-12 | Minor | internal/connection | Two accounts with the same display name are indistinguishable. | None | Accepted risk |
| R-13 | Minor | internal/connection address validation | A path prefix in the space address is unsupported. | None | Accepted risk |

### Validation Tool Results

| Tool | Result | Interpretation |
|---|---|---|
| go test -race ./internal/connection/... ./internal/plugin/... | PASS for both packages (results cached, so inputs are unchanged since the last passing run) | No regression. |

### Summary

The U2 code is unchanged since the last READY review, and the empty-verifier guards were re-confirmed in StartOAuth and completeOAuth. Race tests pass. Only Minor items and two accepted risks remain, so the unit is READY.
