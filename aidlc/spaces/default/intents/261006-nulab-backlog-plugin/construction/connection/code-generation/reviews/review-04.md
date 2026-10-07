## Review

**Verdict:** NOT-READY
**Reviewer:** aidlc-architecture-reviewer-agent
**Date:** 2026-10-06T23:19:38Z
**Iteration:** 1

### Findings

| ID | Severity | Location | Finding | Required action | Status |
|---|---|---|---|---|---|
| R-01 | Critical | internal/connection/oauth.go > StartOAuth, verifierHashPattern (line 172); internal/connection/store.go > TakePending (line 458); internal/plugin/webhook.go > verifierCookie | The cross-user attack is NOT closed. verifierHash is attacker-chosen at start_oauth and is only checked for 64 lowercase hex. verifierCookie returns "" when the cookie is absent, and TakePending then compares sha256("") with the stored hash with no empty-verifier guard. An attacker who starts a sign-in with verifierHash = e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855 (sha256 of the empty string) and sends the authorize URL to a victim gets a match in the victim's browser, because the victim has no cookie. The attack the verifier was meant to stop therefore still works, and no test covers it. | Reject an empty verifier in CompleteOAuth/TakePending, before any comparison, as a failed bad_state that keeps the record. Optionally also reject verifierHash equal to sha256("") at start. Add regression tests: callback with no cookie against a record whose hash is sha256(""), and an empty cookie value. | Unresolved |
| R-02 | Major | ui/src/settings/SettingsScreen.tsx > ProjectPicker key={state.view?.spaceHost} | Original R-02 is fixed for the cases it named. An API-key replace to another host changes the key and remounts the picker. A restore goes through disconnected, which unmounts the picker, so it remounts. OAuth returns by full navigation. The epoch is rightly not in the key. | none | Resolved |
| R-03 | Minor | internal/connection/store.go > nextConnection | A reordered selection still bumps the epoch. | Carry forward from earlier review. | Unresolved |
| R-04 | Minor | internal/connection/store.go > viewOf | HasGitCredential is still missing from the view. | Carry forward from earlier review. | Unresolved |
| R-05 | Minor | internal/connection/store.go > saveConnection (lines 267-273) | deleteGit still runs before write, so a failed write loses the Git password. | Carry forward from earlier review. | Unresolved |
| R-06 | Minor | ui/src/settings/SettingsScreen.tsx > sign_in_again and error views | There is still no Disconnect action in sign_in_again or error. | Carry forward from earlier review. | Unresolved |
| R-07 | Minor | internal/connection refresh path | Refresh still runs on the caller ctx. | Carry forward from earlier review. | Unresolved |
| R-08 | Minor | internal/connection/oauth.go > completeOAuth | The callback still holds the workspace lock through the code exchange and Myself call. | Carry forward from earlier review. | Unresolved |
| R-09 | Minor | internal/backlog rate-limit handling | X-RateLimit-Reset is unbounded and an HTTP-date Retry-After is ignored. | Carry forward from earlier review. | Unresolved |
| R-10 | Minor | internal/connection ErrOAuthNotConfigured mapping | ErrOAuthNotConfigured is still reported as a validation error. | Carry forward from earlier review. | Unresolved |
| R-11 | Minor | U2 impact errors in the project picker | Impact errors are still shown as 0. | Carry forward from earlier review. | Unresolved |
| R-12 | Minor | ui/src/settings/SettingsScreen.tsx > ProjectPicker key | A same-host API-key replace to a different account does not remount the picker. The server keeps the selection (nextConnection same-host branch), so the checkbox state stays correct. But the visible project list was loaded for the old account and stays stale (it may show projects the new account cannot see, or miss new ones) until reload. This is a narrower gap than the old R-02 and is not a leak of checkbox state. | Optionally key on spaceHost plus connectedUserId, which does not change on set_projects. | New |
| R-13 | Minor | ui/src/settings/oauth.ts > CALLBACK_PATH; internal/plugin/webhook.go | A public_base_url with a path prefix makes the cookie Path miss the callback. The sign-in then fails closed (bad_state) and the pending record stays until the 10-minute TTL. This is an availability limitation, not a security flaw. It is acceptable only if documented: put it in the unit-test or deployment notes as an unsupported configuration. | Document the limitation. No code change is needed for release. | New |

### Validation Tool Results

| Tool | Result | Interpretation |
|---|---|---|
| gofmt -l internal server cmd | clean | no formatting drift |
| go vet ./... | clean | none |
| go test -race -cover ./internal/... ./server/... | PASS; connection 94.5%, plugin 93.7%, backlog 96.0%, all above the 80% floor | no race regression; coverage floor holds. The tests do not exercise the empty-verifier bypass in R-01. |
| git status --short before and after | identical | workspace unmodified |

**Verified OK**

- The comparison is constant time (sha256 plus subtle.ConstantTimeCompare in hashMatches), and it runs in TakePending before ExchangeOAuthCode.
- A mismatch keeps the record. An expired record is deleted.
- Only hashes are stored.
- The verifier is added to redact.WithSecrets and never reaches the Location, events or UI replies (verified by reading the code).
- The verifier has 256 bits of entropy (crypto.getRandomValues, 32 bytes).
- The cookie attributes are sensible: Secure, SameSite=Lax, a Path scoped to the callback, Max-Age=600.
- The 64-hex pattern is correct as a format check, but it does not bound the value to hashes of real verifiers (see R-01).

**Not verified:** the reviewer-scope hook blocked reading ../kandev (handlers.go). The claims that Kandev forwards Cookie and drops Set-Cookie are taken from the developer's report and the code comment in webhook.go. They were not checked against v0.96.0 and should be confirmed by the contract test.

### Summary

The verifier design is sound except for one hole. Because the attacker picks verifierHash and a missing cookie reads as an empty string, an attacker can set the hash to sha256("") and bypass the binding for any victim. This leaves the original cross-user finding exploitable, so the verdict is NOT-READY until empty verifiers are rejected and tested. The old R-02 is resolved, and the other carried Minors are unchanged.
