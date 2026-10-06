## Review

**Verdict:** READY
**Reviewer:** aidlc-architecture-reviewer-agent
**Date:** 2026-10-06T12:56:51Z
**Iteration:** 1

### Findings

| ID | Severity | Location | Finding | Required action | Status |
|---|---|---|---|---|---|
| R-01 | Major | internal/connection/oauth.go StartOAuth, CompleteOAuth, decodeState; store.go SavePending/TakePending; manifest webhook oauth-callback access: public | Verified (A-01). The OAuth state is a nonce plus workspaceId checked only against the server-side pending record; nothing binds it to the browser that started sign-in. A workspace admin can call start_oauth for a space where a victim has an account, send the authorizeUrl to the victim, and if the victim clicks Allow the victim's token is stored in the attacker's workspace. The plan assumption claims this gives the same guarantee as a signed state, which overstates it. Note the plugin cannot set the cookie itself, because Kandev drops Set-Cookie on relayed webhook responses (handlers.go:636). Kandev v0.96.0 does forward non-session cookies to public webhooks (flattenHeaders, handlers.go:702). | Bind the sign-in to the starting browser. The settings UI (same Kandev origin) sets a random verifier cookie scoped to the callback path before opening authorizeUrl. The pending record stores its hash and the callback compares it in constant time. If the team will not do this, record it as an Accepted risk with the attacker model and correct the assumption text. | New |
| R-02 | Major | ui/src/settings/SettingsScreen.tsx ProjectPicker element (about lines 243-245); project-picker.tsx load deps [workspaceId] | Verified (A-02). ProjectPicker loads only when workspaceId changes and has no key. After an API key replace to a different space (or a restore) onView updates the view, but the picker stays mounted and keeps the old space's projects and checked keys. This breaks AC1.8.2, which says the selection is cleared on a space change. Save then sends the old keys. The server validates them against the new space, so the result is a validation error, or a wrongly kept selection when keys overlap, not corruption. The OAuth path reloads the page, so only the API-key path is affected. | Reload the picker when connectionEpoch or spaceHost changes, for example key={view.spaceHost + view.connectionEpoch}. Add a Vitest case: replace to another space, then expect the new list with nothing selected. | New |
| R-03 | Minor | internal/connection/store.go:408 saveProjects (slices.Equal); projects.go ValidateProjectKeys | Verified (A-03). The comparison is order-sensitive, so the same selection in a different order bumps the epoch and emits a spurious projects_changed event. | Compare as sets, or sort in ValidateProjectKeys. | New |
| R-04 | Minor | internal/connection/store.go viewOf (about 181-186), saveConnection return, saveProjects return | Verified (A-04, U4 regression). viewOf never sets HasGitCredential. Only Store.Load does. The set_projects and same-host replace replies therefore show Git access as missing even when the Git secret is kept. The UI applies these views through onView. | Set HasGitCredential in the replies (read the Git secret) or have the UI refetch after these actions. Add a test. | New |
| R-05 | Minor | internal/connection/store.go saveConnection (about 267-276) deleteGit before write | Verified (A-05, U4 regression). On a space change the old Git password is deleted before the new connection is written. If the write fails, the rollback restores the connection secret but not the Git secret, so the old Git password is lost while the old connection stays. | Delete the Git secret after a successful write, or keep it for a rollback. | New |
| R-06 | Minor | ui/src/settings/SettingsScreen.tsx (about 208-245); connected-panel.tsx | Verified (A-06). ConnectedPanel (with Disconnect) renders only when status is set. In the sign_in_again and error states the user can sign in again or replace credentials but cannot disconnect. | Show a Disconnect control in those states. | New |
| R-07 | Minor | internal/connection/lifecycle.go:96-102 refresh, UpdateTokens | Still open (old R-03). Refresh runs on the caller's ctx and Backlog rotates the refresh token on every refresh. If ctx ends, or the store write fails, after Backlog accepted the refresh, the new token is lost and the old one is dead, so the user must sign in again. The window is small and the result is recoverable, which is why this is Minor. | Run the exchange and the token write under context.WithoutCancel plus a short timeout. | Unresolved |
| R-08 | Minor | internal/connection/oauth.go:197-242 completeOAuth | Still open (old R-05). The callback holds the workspace lock through the code exchange and Myself, so refresh and other writes for that workspace wait for up to the Backlog timeouts. It is bounded and rare. | Take the lock only for the store write, or document the bound. | Unresolved |
| R-09 | Minor | internal/backlog/client.go:487-495 rateLimitWait | Still open (old R-06). X-RateLimit-Reset is not bounded, so a far-future value yields a huge RetryAfter. An HTTP-date Retry-After is ignored and falls back to 60 s. | Cap the wait (for example 1 hour) and parse HTTP-date Retry-After. | Unresolved |
| R-10 | Minor | internal/connection/service.go:312-313 Classify | Still open (old R-07). ErrOAuthNotConfigured raised during a token refresh is classified as a validation error on field oauth, even for an unrelated action such as list_projects. | Map it to reconnect_required or internal when it arises outside start_oauth. | Unresolved |

### Validation Tool Results

| Tool | Result | Interpretation |
|---|---|---|
| gofmt -l internal server cmd | clean | No formatting issues |
| go vet ./... | clean | No findings |
| go test -race -cover ./internal/... ./server/... | PASS; connection 94.5%, backlog 95.8%, plugin 93.1% | Above the 80% floor; server/ is 0% (wiring only) |
| go run ./cmd/ci secrets -root . | OK | No secrets found |
| git status --short before and after | identical | The workspace was not modified |

### Summary

Two Major findings (browser binding of the OAuth state, and a stale ProjectPicker after a space change) and no Critical findings, so the verdict is READY under the rules. R-01 is a real login-CSRF style gap that should be fixed or explicitly accepted before the first release. The rest are Minor, and the carried-over items are re-confirmed.
