## Review

**Verdict:** READY
**Reviewer:** aidlc-architecture-reviewer-agent
**Date:** 2026-10-06T23:26:14Z
**Iteration:** 2

### Findings

| ID | Severity | Location | Finding | Required action | Status |
|---|---|---|---|---|---|
| R-01 | Critical | internal/connection/oauth.go > StartOAuth, completeOAuth; internal/connection/store.go > TakePending | The empty-verifier bypass is closed by three independent layers: start_oauth rejects verifierHash == sha256("") (field verifierHash, validation); completeOAuth returns bad_state for an empty verifier before the lock and any comparison, and keeps the record; TakePending returns no match for an empty verifier. The hash pattern is lower-case hex only, so a case variant of the empty hash is rejected at start by the pattern. A whitespace-only or empty-quoted cookie value is either skipped by net/http Cookie parsing or yields "" and is refused. Duplicate cookies cannot satisfy an attacker hash, because the attacker cannot place a value in the victim's browser. A parent-path cookie would need a same-site cookie-tossing position, which is outside this threat model. TestU2_EmptyVerifierNeverMatches covers start refusal, no-cookie, lock-busy (proves the up-front guard) and the end-to-end attack; TestOAuthPendingRefusesAnEmptyVerifier covers the store guard; the webhook case covers an empty cookie. Each guard has a test that fails without it by reading the code; I did not re-run mutations, since I was not allowed to modify the workspace. | none | Resolved |
| R-02 | Major | ui/src/settings/SettingsScreen.tsx > ProjectPicker key | Resolved in iteration 1. | none | Resolved |
| R-03 | Minor | internal/connection/store.go > nextConnection | A reordered selection still bumps the epoch. | Carry forward. | Unresolved |
| R-04 | Minor | internal/connection/store.go > viewOf | HasGitCredential is still missing from the view. | Carry forward. | Unresolved |
| R-05 | Minor | internal/connection/store.go > saveConnection | deleteGit runs before the write, so a failed write loses the Git password. | Carry forward. | Unresolved |
| R-06 | Minor | ui/src/settings/SettingsScreen.tsx > sign_in_again and error views | No Disconnect action in sign_in_again or error. | Carry forward. | Unresolved |
| R-07 | Minor | internal/connection refresh path | Refresh runs on the caller ctx. | Carry forward. | Unresolved |
| R-08 | Minor | internal/connection/oauth.go > completeOAuth | The callback holds the workspace lock through the code exchange and Myself call. | Carry forward. | Unresolved |
| R-09 | Minor | internal/backlog rate-limit handling | X-RateLimit-Reset is unbounded and an HTTP-date Retry-After is ignored. | Carry forward. | Unresolved |
| R-10 | Minor | internal/connection ErrOAuthNotConfigured mapping | ErrOAuthNotConfigured is reported as a validation error. | Carry forward. | Unresolved |
| R-11 | Minor | U2 impact errors in the project picker | Impact errors are shown as 0. | Carry forward. | Unresolved |
| R-12 | Minor | ui/src/settings/SettingsScreen.tsx > ProjectPicker key | The key is now spaceHost:connectedUserName, which fixes the stale picker for a same-host replace to a differently named account. Two accounts with the same display name on one space still keep a stale list. The server keeps the selection, so the checkbox state is correct, and the refresh button and a reload fix the list. The cause is that the view has no user id. | Accept the display-name keying as a known limitation. Add a user id to the view if U3 needs one. | Accepted risk |
| R-13 | Minor | internal/plugin/webhook.go; ui/src/settings/oauth.ts > CALLBACK_PATH; README.md | Kandev under a path prefix makes the cookie Path miss the callback. Sign-in fails closed (bad_state) and nothing is stored. README now documents it as unsupported and recommends an API key. This is an availability limit only. | none | Accepted risk |

### Validation Tool Results

| Tool | Result | Interpretation |
|---|---|---|
| gofmt -l internal server cmd | clean | No formatting drift. |
| go vet ./... | clean | None. |
| go test -race -cover ./internal/... ./server/... | PASS; connection 94.6%, plugin 93.7%, backlog 96.0% | No race regression; the 80% floor holds. |
| git status --short before and after | identical | Workspace unmodified. |

### Summary

The empty-verifier bypass (R-01) is closed in depth at start, callback and store, with tests that exercise each layer. R-12 and R-13 are acceptable documented limitations, and R-03..R-11 stay as non-blocking Minor items. Not verified: whether Kandev v0.96.0 forwards the Cookie header, which is still left to the contract test.
