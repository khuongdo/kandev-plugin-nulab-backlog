# Code Summary — connection (U2)

All 16 plan steps were carried out in order under the TDD Testing Contract (`contract_sha256 sha256:376593af89ee0d9cddcd078826773d6d5ca3215ec737b9af952712bee2395bcc`). Each layer went Red, then Green, then Refactor; the Red output of every layer is below. From `make clean`, `make check-format vet lint test coverage check-secrets build package verify-package` passes. Go line coverage is **93.9%** (floor 80%, only `server/main.go` excluded). `go mod tidy` leaves no diff. `make contract-test KANDEV_MIN_DIR=../kandev-min` passes on Kandev v0.96.0.

U2 adds no Go package, no Go module and no npm dependency. Only `internal/plugin` imports `pluginsdk`. The uncommitted U5 work (`Makefile`, `.github/workflows/*`, `.golangci.yml`, `internal/ci`, `cmd/`) was not touched; `README.md` only got a section appended at the end.

## Files

`source-manifest.json` lists every path U2 created or changed.

| Area | Files |
|------|-------|
| BacklogGateway | `internal/backlog/client.go` (one `send` path: auth header, per-(host, group) queue, 1 s spacing, 429 retry loop, Interactive 3 s budget, `backlog_wait` log), `group.go` (rate-limit groups), `oauth.go` (`OAuthClient`, `TokenSet`, `ExchangeOAuthCode` and `RefreshToken` are in `client.go`), `projects.go` (`Project`, `parseProjects`; `Projects` is in `client.go`); tests `client_test.go` (U2 cases), `limiter_test.go`, `oauth_test.go`, `oauth_types_test.go`, `projects_test.go`, `projects_types_test.go`, `group_test.go`; fixtures `testdata/token_ok.json`, `token_refreshed.json`, `token_invalid_grant.json`, `projects_ok.json`, `projects_empty.json`, `error_500_bait.json` |
| Connection | `internal/connection/oauth.go` (`OAuthConfig`, `ParseOAuthConfig`, state token, `StartOAuth`, `CompleteOAuth`), `projects.go` (`ValidateProjectKeys`), `change.go` (`Reason`, `ConnectionChanged`, `changeFor`, `Subscribe` and the per-subscriber queues), `lifecycle.go` (`Current`, `Credentials` and the refresh, `Test`, `Disconnect`, `ListProjects`, `SetProjects`), `store.go` (one write path, `SaveOAuth`, `UpdateTokens`, `MarkSignInAgain`, pending state, `Disconnect`, `SaveProjects`, space change and restore), `service.go` (wider `Gateway`, `ConfigReader`, per-workspace write lock, `Classify`); tests `oauth_config_test.go`, `state_token_test.go`, `project_key_test.go`, `change_test.go`, `record_compat_test.go`, `store_test.go` (U2 cases), `credentials_test.go`, `oauth_flow_test.go`, `lifecycle_test.go`, `projects_service_test.go`, `events_test.go`, `u2_helpers_test.go`; fakes extended in `fakes_test.go` and `service_test.go` |
| KandevAdapter | `internal/plugin/runtime.go` (5 handlers, `reconnect_required` → 401, `action_failed` log, config wiring), `webhook.go` (`HandleWebhook`), `config.go` (`GetConfig` pass-through on the host adapter); tests `actions_u2_test.go`, `webhook_test.go`, `config_test.go`, `manifest_test.go` (U2 cases), `actions_test.go` (fakes extended) |
| Manifest | `manifest.yaml`: 5 actions, the public `oauth-callback` webhook, the optional `config_schema` |
| Test helper | `internal/testutil/testutil.go`: `Token()` (`test-token-` + 32 random hex characters), `Windows` trims both prefixes |
| UI | `ui/src/settings/oauth.ts`, `confirm-dialog.tsx`, `connected-panel.tsx`, `project-picker.tsx`, `state.ts`, `SettingsScreen.tsx`, `ui/src/messages/en.ts`; tests `state.test.ts`, `oauth.test.tsx`, `connected-panel.test.tsx`, `project-picker.test.tsx`, `confirm-dialog.test.tsx`; `settings.test.tsx` (U1) and `ui/src/testing/harness.ts` adjusted |
| Docs | `README.md` (appended "Sign in with OAuth"), `docs/manual-checks/TEMPLATE.md` (U2 steps 10–14) |

## Key Decisions

- **One send path in the gateway.** `Myself`, `Projects`, `ExchangeOAuthCode` and `RefreshToken` all go through `Client.send`. An access token goes in `Authorization: Bearer`, otherwise the API key goes in `Backlog-API-Key`; a request never carries both, and nothing is put in the URL. The token endpoint sends a form body and no auth header.
- **Queues are a one-slot channel per (host, group)** with the last start time guarded by holding the slot. A channel (not `sync.Mutex`) lets an Interactive call stop waiting when its 3 s budget runs out. Read calls are not queued. Queue time, spacing and 429 waits all come out of the same Interactive budget; Background calls wait as long as Backlog asks. `Client.Wait` is injected; the default honours the context.
- **Timing tests run in memory.** `limiter_test.go` uses an `http.RoundTripper` that calls the handler directly, so the tests run inside a `testing/synctest` bubble with no socket and no real sleep.
- **The record and the secret move together.** `Store.write` is the one BR2.8 writer (secret first, then record, U1 rollback). Connect, OAuth sign-in and project changes use it; each bumps the epoch. A token refresh rewrites only the secret and keeps the epoch (`ErrStale` if the epoch moved). Disconnect deletes the secret first, then writes a disconnect record with epoch + 1 that remembers the host and its projects.
- **Space change and restore live in one function** (`nextConnection`): same host keeps the selection; another host clears it and remembers the old host and projects; the remembered host puts its projects back. `changeFor` then names the reason and the `Restore` flag.
- **Single refresh per workspace.** `Credentials` checks the token without a lock. When fewer than 5 minutes are left, it takes the workspace write lock, reads the secret again and refreshes only if it is still old. This uses the standard library only. The same lock serialises Connect's store step, the OAuth callback, Disconnect and project saves, and events are sent under it, so they arrive in epoch order.
- **Events are in-process.** Each subscriber has its own unbounded FIFO queue and goroutine, and panics are recovered and logged as `connection_subscriber_panic`. Every sent event writes one `connection_changed` line. Tests use that line to check that no event was sent.
- **OAuth state.** The state is base64url(32-byte nonce ‖ workspaceId). The workspace's `oauth_pending` state holds only the SHA-256 of the nonce, the host and the expiry. A matching record is deleted before use. A wrong nonce leaves the record in place, so a forged callback cannot cancel a real sign-in. An expired record is deleted.
- **Callback redirect.** The callback returns a relative `302` to `/settings/workspaces/<ws>/integrations/nulab-backlog?oauth=<outcome>[&restored=1]`, with the workspace path-escaped. The decoded workspace id must match `^[A-Za-z0-9_-][A-Za-z0-9._-]*$`. The `Location` never holds the code, a token, the state or Backlog text.
- **UI dialog.** Kandev's UI kit has no AlertDialog, so `confirm-dialog.tsx` is a plain `role="alertdialog"` element. It moves focus to Cancel, traps Tab, closes on Esc and gives the focus back to the element that opened it.

## Red evidence

Each Red run happened before the production code of its layer existed. Key lines are trimmed.

**Step 1, test helper.** `go test -race ./internal/testutil/...`

```
internal/testutil/testutil_test.go:26:10: undefined: Token
internal/testutil/testutil_test.go:32:15: undefined: TokenPrefix
FAIL	.../internal/testutil [build failed]
```

**Step 2, runner readiness.** The U2-scoped `go test … -run '^Test(OAuth|…|U2)'` reported `ok … [no tests to run]` for `backlog`, `connection`, `plugin` and `testutil`. The U2 Vitest command with `--passWithNoTests` ran clean. The U1 regression (`npx vitest run src/settings/settings.test.tsx src/switch src/page`) was green: 3 files, 54 tests.

**Step 3, data model.** `go test -race ./internal/backlog/... ./internal/connection/...`

```
internal/backlog/group_test.go:13:16: undefined: Group
internal/backlog/group_test.go:15:44: undefined: GroupRead
internal/connection/change_test.go:10:80: unknown field SelectedProjects in struct literal of type record
internal/connection/change_test.go:15:11: undefined: Reason
FAIL	.../internal/backlog [build failed]
FAIL	.../internal/connection [build failed]
```

**Step 5, repository / data access.** `go test -race ./internal/connection/...`

```
internal/connection/store_test.go:421:21: store.SaveOAuth undefined (type *Store has no field or method SaveOAuth)
internal/connection/store_test.go:458:27: store.UpdateTokens undefined (type *Store has no field or method UpdateTokens)
internal/connection/store_test.go:477:26: undefined: ErrStale
FAIL	.../internal/connection [build failed]
```

**Step 7, gateway.** `go test -race ./internal/backlog/...`

```
internal/backlog/limiter_test.go:44:36: undefined: request
internal/backlog/limiter_test.go:72:17: c.send undefined (type *Client has no field or method send)
internal/backlog/limiter_test.go:178:4: c.Wait undefined (type *Client has no field or method Wait)
FAIL	.../internal/backlog [build failed]
```

**Step 9, connection service.** `go test -race ./internal/connection/...`

```
internal/connection/credentials_test.go:18:18: u.svc.Current undefined (type *Service has no field or method Current)
internal/connection/credentials_test.go:26:19: undefined: Snapshot
internal/connection/credentials_test.go:52:19: undefined: CodeReconnectRequired
FAIL	.../internal/connection [build failed]
```

**Step 11, API / endpoint.** `go test -race ./internal/plugin/...`. The first run failed to build (`hostStores{…}.GetConfig undefined`). With only the 8-line config adapter added, the behaviour tests then failed:

```
--- FAIL: TestRecheckActionSucceedsAndRejectedKeyIsReconnectRequired
--- FAIL: TestProjectsActions
--- FAIL: TestOAuthStartActionReadsTheConfigOnEveryCall
--- FAIL: TestActionFailureLogOnABacklog500
--- FAIL: TestManifestActionsAndAccess
--- FAIL: TestU2_ManifestWebhookAndConfigSchema
--- FAIL: TestWebhookRedirectsWithTheOutcome
FAIL	.../internal/plugin
```

**Step 13, frontend.** `npx vitest run src/settings/state.test.ts src/settings/oauth.test.tsx src/settings/connected-panel.test.tsx src/settings/project-picker.test.tsx src/settings/confirm-dialog.test.tsx`

```
FAIL  src/settings/confirm-dialog.test.tsx  Error: Failed to resolve import "./confirm-dialog"
FAIL  src/settings/oauth.test.tsx  Error: Failed to resolve import "./oauth"
FAIL  src/settings/project-picker.test.tsx  Error: Failed to resolve import "./project-picker"
FAIL  connected-panel.test.tsx > tests the connection and announces the user name  TypeError: Cannot read properties of null (reading 'click')
FAIL  state.test.ts > maps reconnect_required to the revoked-key notice  TypeError: failureNotice is not a function
Failed Tests 19
```

## Results

| Check | Result |
|-------|--------|
| `make clean` then `make check-format vet lint test coverage check-secrets build package verify-package` | exit 0 |
| golangci-lint (default + gosec, v2.14.0) | `0 issues.` |
| Go tests (`-race`, whole module) | 266 passing test functions, 0 failures |
| U2-scoped Go command (`-run '^Test(OAuth|…|U2)'`) | 119 passing test functions |
| Go coverage (`make coverage`) | `coverage: 93.9% (floor 80%, excluded: server/main.go)`; backlog 94.7%, connection 95.3%, plugin 95.9%, testutil 87.0% |
| Vitest (all) | 10 files, 102 tests passed; U2 files: 5 files, 37 tests |
| `tsc --noEmit`, ESLint, Prettier | clean |
| `make check-secrets` | `ci secrets: OK` |
| `make verify-package` | `verifypkg: OK dist/nulab-backlog-0.0.1.tar.gz (nulab-backlog@0.0.1)`; the packaged `manifest.yaml` has the webhook and `config_schema` |
| `go mod tidy` | no diff in `go.mod` / `go.sum` |
| `make contract-test KANDEV_MIN_DIR=../kandev-min` | `ci contract: OK nulab-backlog on Kandev v0.96.0` (ports 38529/39529; the user's instance on 38429/39429 was not touched) |

## Deviations from the plan

- **Test names with `U2` use `TestU2_…`.** U5's `check-secrets` treats a run of 32 or more letters and digits that mixes cases and digits as a credential, and `TestU2StoreCallsKeepTheOneSecondLimit` matched it. The underscore splits the run. The `-run '^Test(…|U2)'` filter still selects these tests.
- **File names.** The service OAuth tests are in `oauth_flow_test.go` and the service project tests are in `projects_service_test.go`, because `oauth_config_test.go` and `project_key_test.go` already cover the plan's data-model files. The OAuth and project gateway calls live in `client.go`, next to `send`; `oauth.go` and `projects.go` hold the types and parsers. The rate-limit groups are in `group.go`. The shared U2 fakes are in `u2_helpers_test.go`.
- **`hostConfig` adapter.** `GetConfig` is a method on the existing `hostStores` adapter, in `config.go`, instead of a new type, because the same adapter already resolves the Host.
- **A wrong nonce does not delete the pending sign-in.** Only a matching or expired record is deleted. Otherwise anyone who knows a workspace id could cancel a user's sign-in.
- **`Credentials` returns `(backlog.Credentials, epoch, error)`.** Late results are dropped by comparing the returned epoch with a later `Current` (AC1.8.3).
- **Two U1 behaviours change, as the plan intends.** The API key moves from the `apiKey` query parameter to the `Backlog-API-Key` header. The U1 test that checked the query parameter is replaced. In the U1 `TestMyselfRateLimitWait` table, the case that clamps a past reset to 1 s now retries (4 attempts), because a 1 s wait fits the Interactive budget. Waits over 3 s still return at once.
- **Native checkbox and radio inputs.** The project list uses `<input type="checkbox">` and the sign-in method uses `<input type="radio">`, each with a label, because the host `Checkbox` has a different (Radix) API. Text inputs and buttons still come from the host UI kit, and the spacing uses the U1 utility classes plus `flex gap-2` for button rows.
- **U1 settings tests updated only for the connected layout:** the connect call is found by key, because the project picker now loads after connecting; the member test confirms the new replace dialog; the layout test counts `label[for]`, because the radio labels wrap their inputs.
- **Members (status 403 after a connect) do not see the connected panel or the project picker**, matching U1's member view.
- **Plan checkbox timing.** The Step 15 checkboxes were ticked slightly before Step 15 ran (with the same edit that ticked Step 14). Step 15 then ran in full, and its results are above.

## Upstream amendments needed

- **C1 (BacklogGateway):** `ExchangeOAuthCode(ctx, spaceHost, OAuthClient, code, redirectURI) (TokenSet, error)` and `RefreshToken(ctx, spaceHost, OAuthClient, refreshToken) (TokenSet, error)`: no `codeVerifier` (Backlog documents no PKCE), and the client credentials are passed in. Add `Projects(ctx, Credentials) ([]Project, error)` and `Project{ID, Key, Name, Archived}`. The API key goes in the `Backlog-API-Key` header. Calls without a `CallClass` are Interactive (R-07).
- **C3 (ConnectionReader / ConnectionChanged):** the methods take a `workspaceID`. `Current(ctx, ws) (Snapshot{SpaceHost, AuthMethod, ConnectionEpoch, SelectedProjects}, error)`, `Credentials(ctx, ws) (backlog.Credentials, epoch int, error)` and `Subscribe(func(ConnectionChanged)) (unsubscribe func())`. `ConnectionChanged{WorkspaceID, Reason, ConnectionEpoch, SpaceHost, SelectedProjects, Restore}` with reasons `connected`, `credentials_replaced`, `space_changed`, `disconnected` and `projects_changed`. The epoch rises on every connection write except a token refresh (R-08). `PollMinutes` is left to U3.
- **C5 (actions):** keys are snake_case (`connection.connect_api_key`, `connection.set_enabled`, `connection.start_oauth`, `connection.test`, `connection.disconnect`, `connection.list_projects`, `connection.set_projects`). New codes: `reconnect_required` (401) and `integration_disabled` (409). Validation fields `oauth` and `projectKeys`. `ConnectionView` adds `selectedProjects`, `restored` and the state `sign_in_again`. `list_projects` replies `{projects: [{projectKey, projectId, projectName, selected}]}`. The OAuth state is a single-use server-side nonce, not a signed state (C6).

## Open items

- The manual B3 demo (OAuth sign-in, refresh after expiry, disconnect and restore, a 429 wait line) needs X5, a registered Nulab OAuth app. The steps are in `docs/manual-checks/TEMPLATE.md`. The automated tests do not need it. The contract test checks that Kandev v0.96.0 accepts the new manifest (public webhook, `config_schema`) and runs the plugin. The relay of the `302` through Kandev's webhook route has only been checked against the SDK source, not on a running server.
- Assumptions still unverified: one Nulab OAuth app serves any space; Nulab accepts an `http://localhost` callback (A2); Backlog accepts the `Backlog-API-Key` header.
- Deferred to later units: the Git-password parts of AC1.5.4, AC1.8.2 and AC1.9.2 (U4, through `ConnectionChanged`); link and watch counts in the dialogs (AC1.8.1, AC1.9.1) and the M12 counts and "Review watches" (U3/U4); the "No project selected" Issues page (AC1.7.2) and the polling-cycle log (AC8.3.2), both U3; the link to the settings page from other Backlog pages in the Sign-in-again state (AC1.4.3, U3 pages).
- `ConnectionChanged` has no production subscriber yet; U3 and U4 subscribe.
