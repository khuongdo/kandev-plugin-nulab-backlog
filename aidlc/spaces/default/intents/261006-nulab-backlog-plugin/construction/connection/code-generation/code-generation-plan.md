# Code Generation Plan — connection (U2)

Inputs:

- `unit-of-work` (U2 boundary), `unit-of-work-story-map` (U2 stories and their order), `unit-of-work-dependency`.
- `stories` (US8.4, US8.3, US1.5, US1.6, US1.7, US1.3, US1.4, US1.8, US1.9).
- `requirements` (FR1.1, FR1.4–FR1.7, NFR2, NFR3, NFR5, NFR9, NFR11).
- `components` (Connection, BacklogGateway, KandevAdapter, PluginUI) and `decisions` (ADR-002, ADR-003, ADR-005, ADR-007).
- `contract-summary`: C1, C3, C5, C6, C7, C8. Also the contract-design review findings R-01, R-02, R-06, R-07 and R-08, which the bolt plan gives to U2.
- `mockups` (M1, M12), `interaction-spec` (ConnectionForm, ConfirmDialog), `accessibility-checklist`.
- `bolt-plan` (B3 Definition of Done) and `external-dependency-map` (X5, the OAuth app).
- U1 design (`functional-spec`, `rules`, `entities`, NFR design) and U1 `code-summary.md`. U1's code is on `main` at `99760f8`; U5 (ci-release) changes are in the working tree.
- `team.md`, `project.md`, `phases/construction.md`.

U2 has no functional, NFR or infrastructure design. The user chose to go straight to code generation. Every design choice this plan had to make is listed under Assumptions & Open Questions.

Stories in U2, in build order: US8.4, US8.3, US1.5, US1.6, US1.7, US1.3, US1.4, US1.8, US1.9.

## Kandev SDK Facts Used by This Plan

These were read from the `../kandev` checkout at v0.96.0 (the commit in `.kandev-sdk-ref`):

- **Webhooks: an OAuth callback is possible.**
  - `pluginsdk.Plugin.HandleWebhook(ctx, *WebhookRequest) (*WebhookResponse, error)`. `UnimplementedPlugin` returns 404 by default.
  - Kandev relays both `GET` and `POST /api/plugins/{id}/webhooks/{key}` (`apps/backend/internal/plugins/handlers.go`).
  - `WebhookRequest{WebhookKey, Method, Path, Query, Headers, Body}`: the raw query string reaches the plugin.
  - The manifest declares `webhooks[]` with `key`, `method` (informational only, not enforced), `access` and `max_body_bytes`. For `api_version: 2`, `access` defaults to `authenticated`, so the callback must say `access: public` explicitly.
  - The plugin's status (100–599), headers and body are relayed as they are, except `Set-Cookie`, which is dropped. So `302` with a `Location` header works.
- **No OAuth support and no base URL from the host.** Kandev has no OAuth route for plugins. An action request carries only `VerifiedActionContext` (workspace, task, repository) and no request URL, so the plugin cannot work out Kandev's public address on its own.
- **Operator config.** The manifest has a `config_schema`, and the plugin reads it with `Host.GetConfig(ctx)`. That call needs no capability. A property with `secret: true` is kept in Kandev's vault and masked as `********` in Settings > Plugins; the plugin process receives the real value. Kandev restarts the plugin when its config changes.
- **State.** `GetState`, `SetState`, `DeleteState` and `ListState` are gated by `capabilities.state`. Scopes are `instance`, `workspace`, `task` and `agent`. A missing key gives `found == false` with no error.
- **Secrets.** `GetSecret`, `SetSecret` and `DeleteSecret` are gated by `capabilities.secrets`. Keys must match `[a-zA-Z0-9][a-zA-Z0-9._-]{0,127}`, and Kandev namespaces them per plugin. Deleting a missing key is not an error.
- **Events.** `Host.EmitEvent(ctx, name, payload)` publishes `plugin.<id>.<name>` on Kandev's bus and needs no capability. U2 does not use it: per contract C3, `ConnectionChanged` is an in-process Go event.
- **Background timers.** Kandev has no timer service for plugins. A plugin may run its own goroutines (C8, ADR-003). U2 needs no timer: token refresh happens when a call needs it, and rate-limit waits use an injected wait function.
- **Action rules.** Keys must match `^[a-z0-9][a-z0-9._-]*$`, so the new keys are snake_case. `access: admin` needs `min_kandev_version` 0.91.1 or later; the manifest already says 0.96.0. Each action gets at most 15 s. Only the `Content-Type`, `Cache-Control`, `ETag` and `Retry-After` response headers reach the browser.
- **UI.** `host.ui` has `Dialog`, `DialogContent`, `DialogTitle`, `DialogDescription`, `DialogFooter`, `Checkbox`, `Input`, `Label`, `Button` and `Spinner`. It has no `AlertDialog`, so the plugin sets `role="alertdialog"` itself. The settings route is `/settings/workspaces/<workspaceId>/integrations/nulab-backlog`. U1 already uses it.
- **Dependencies.** The Go standard library covers everything: `net/http`, `crypto/rand`, `crypto/subtle`, `encoding/base64`, `sync` and `testing/synctest`. No module is added, so `go mod tidy` stays clean.

### Backlog API v2 facts (developer.nulab.com, read 2026-10-06)

- **Authorization page:** `GET https://<space>/OAuth2AccessRequest.action?response_type=code&client_id=…&redirect_uri=…&state=…`.
- **Token endpoint:** `POST /api/v2/oauth2/token`, sent as `application/x-www-form-urlencoded`. `grant_type=authorization_code` takes `code`, `redirect_uri`, `client_id` and `client_secret`. `grant_type=refresh_token` takes `refresh_token`, `client_id` and `client_secret`. The response has `access_token`, `token_type` (`Bearer`), `expires_in` (3600) and `refresh_token`. A new refresh token is issued on every refresh.
- **Using credentials:** an access token goes in `Authorization: Bearer <token>`. An API key can go in the `Backlog-API-Key` header or the `apiKey` query parameter.
- **PKCE** is not documented.
- **Rate limits:** calls fall into four groups: Read (GET), Update (POST, PATCH, DELETE), Search (issue list, issue count, wiki list, wiki count) and Icon. The headers are `X-RateLimit-Limit`, `X-RateLimit-Remaining` and `X-RateLimit-Reset` (UTC epoch seconds). Going over the limit gives 429. Limits apply per user, and Backlog recommends sending requests one at a time.

## Testing Contract

```json
{
  "version": 1,
  "methodology": "tdd",
  "source": "team",
  "ordering": "For each testable layer, we write a failing test first, then write the minimum code to make it pass, then refactor while the test stays green, before moving to the next behaviour or layer.",
  "scope": "feature",
  "test_strategy": "standard",
  "project_type": "greenfield",
  "applicable_notes": [
    {
      "layer": "org",
      "text": "We treat tests as a first-class deliverable in every Bolt. The specific\nmethodology (TDD, BDD, ATDD, or classic test-after) is affirmed at\npractices-discovery and recorded in `team.md` under this heading with explicit\n`Methodology` and `Ordering` fields; Code Generation resolves those fields\nindependently from coverage, tooling, and scope notes.\n\nWhen no posture has been affirmed, our default per scope is:\n- **Methodology**: test-after\n- **Ordering**: implement each applicable testable layer, then write and run\n  that layer's tests.\n- `mvp`, `enterprise`, `feature`, `infra`, `classic` add an 80% line-coverage\n  floor and CI execution before merge.\n- `bugfix`, `security-patch` add a targeted regression for the specific\n  bug/vulnerability and require the existing suite to remain green.\n- `express` uses the Minimal strategy: requirement-driven unit tests (one per\n  requirement, with a happy-path floor per component); existing tests remain\n  green.\n- `poc`, `refactor`, `workshop` add no extra new-test floor and require the\n  existing suite to remain green.\n\nThe active `Test Strategy` still applies in every scope and determines test\nvolume/types. Scope floors are additive; they never reduce or replace the\nselected strategy.\n\nBuild and Test verifies defined coverage floors and affirmed quality targets;\nthey may not be weakened to make a step pass.\n\nAffirm a stricter posture in `team.md` if the team commits to one."
    },
    {
      "layer": "team",
      "text": "- We treat tests as a deliverable of every Bolt, not extra work.\n- **Methodology**: tdd\n- **Ordering**: For each testable layer, we write a failing test first, then write the minimum code to make it pass, then refactor while the test stays green, before moving to the next behaviour or layer.\n- A floor of **80% line coverage for the Go code**, measured with `go test -coverprofile` and enforced by a `make coverage` target; CI blocks the merge when it is lower. The measured scope is `./internal/...` and `./server/...`; only the minimal wiring in `main` is excluded, and every exclusion must be listed explicitly in the `Makefile`. We do not lower the floor or add exclusions to make it pass.\n- Go tests always run with **`go test -race`** (needs CGO; works on the `ubuntu-latest` runner). This is an addition to the Bitbucket template, because token refresh and API rate limiting are two places prone to data races.\n- **Automated contract test with the real package on the minimum Kandev version**, like the template's `packaged-host-contract` job: CI checks out Kandev at exactly the `min_kandev_version` in `manifest.yaml`, builds a throwaway server, installs the packaged plugin, and runs it.\n- **Manual end-to-end check against a real Backlog space exactly twice**: when the walking skeleton is done, and before the first release. Each result is recorded. After the first release, we rely only on automated tests.\n- Unit and integration tests do not call real Backlog: they use a **fake Backlog server built with `net/http/httptest`**, with JSON sample data in `internal/<package>/testdata/`, including the 401, 429 (with `Retry-After`), and expired-token error cases.\n- Tools: `testing` + `github.com/stretchr/testify/require`, table-driven tests with `t.Run`, test names that describe behaviour. The clock and wait functions are injected into the code so rate limits and token expiry can be tested without a real `time.Sleep`. Tests are independent of each other and use `t.TempDir()` and `t.Setenv()`. The TypeScript UI (if any) is tested with Vitest.\n- There is a test asserting that API keys and tokens do not appear in logs, error messages, or responses sent to the UI.\n- We **do not add a dependency vulnerability scan** (`govulncheck`, `npm audit`, Dependabot) — this is your decision in Q4."
    }
  ],
  "obligations": {
    "strategy": "standard",
    "strategy_volume": [
      "Five to eight tests per component.",
      "Unit tests plus integration tests for key boundaries.",
      "Add E2E, performance, or security tests when requirements demand them."
    ],
    "scope_floor": [
      "Meet an 80% line-coverage floor.",
      "Run the selected tests in CI before merge."
    ],
    "combination_rule": "Apply every selected-strategy obligation and every scope-floor obligation; neither replaces the other, and a targeted scope regression may add the narrowest necessary test type beyond the strategy default."
  },
  "plan_profile": {
    "methodology": "tdd",
    "runner_step": "Bootstrap the minimal test runner/configuration and record the exact unit-scoped command.",
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
      "Bootstrap the minimal test runner/configuration and record the exact unit-scoped command.",
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
  "input_sha256": "sha256:c08268ec87b4a805bbc352db69f8ea8730f7297e5e63201ea588835bf2850378",
  "contract_sha256": "sha256:376593af89ee0d9cddcd078826773d6d5ca3215ec737b9af952712bee2395bcc"
}
```

## Plan Steps

Each TDD layer goes Red → Green → Refactor. Each Red step records its failing command output in `construction/connection/code-generation/code-summary.md`.

**Test naming rule.** Every new Go test function in U2 starts with one of these prefixes: `OAuth`, `Token`, `RateLimit`, `Queue`, `Projects`, `Recheck`, `Disconnect`, `Replace`, `SpaceChange`, `Restore`, `ConnectionChanged`, `ConnectionReader`, `Webhook`, `APIKeyHeader`, `ActionFailureLog`, `U2`. This lets the unit-scoped `-run` filter in `unit-test-instructions.md` select exactly the U2 tests. No existing U1 or U5 test name starts with these prefixes.

### Step 1 — Project structure and configuration skeleton

- [x] Confirm that U2 adds no Go package, no Go module and no npm dependency. All code goes into the existing `internal/backlog`, `internal/connection`, `internal/plugin` and `ui/src/settings`. Only `internal/plugin` imports `pluginsdk`.
- [x] Add fake-only JSON fixtures: `internal/backlog/testdata/token_ok.json` and `token_refreshed.json` (tokens built at test time where a test asserts on them); `token_invalid_grant.json`; `projects_ok.json` (keys `PROJ` and `DEMO`, fake names) and `projects_empty.json`; `error_500_bait.json`, a body with a bait placeholder that the test fills with a per-run key. Fixtures must pass `make check-secrets` (U5).
- [x] Extend `internal/testutil` with `Token()`. It returns a per-run random fake token built the same way as the U1 key helper (prefix plus 32 random hex characters), so the existing 8-character-window leak check works for tokens, client secrets and auth codes. Add a matching case to `testutil_test.go`.
- [x] Do not touch `Makefile` or `.github/workflows/*` (owned by U5).
- Stories: all U2 stories (fixtures). NFRs: NFR4.

### Step 2 — Test runner readiness (already in place from U1)

- [x] Confirm that these commands run before the first Red step: from the repository root, `go test -race ./internal/backlog/... ./internal/connection/... ./internal/plugin/... ./internal/testutil/... -run '^Test(OAuth|Token|RateLimit|Queue|Projects|Recheck|Disconnect|Replace|SpaceChange|Restore|ConnectionChanged|ConnectionReader|Webhook|APIKeyHeader|ActionFailureLog|U2)'` (reports `ok … [no tests to run]` per package); from `ui/`, `npx vitest run src/settings/state.test.ts src/settings/oauth.test.tsx src/settings/connected-panel.test.tsx src/settings/project-picker.test.tsx src/settings/confirm-dialog.test.tsx --passWithNoTests`.
- [x] Confirm that the U1 regression commands in U1 `unit-test-instructions.md` are still green.
- [x] Record the commands in `unit-test-instructions.md`. They are already written there.
- Stories: all.

### Step 3 — Data model, Red: tokens, projects, OAuth config, state token, change reasons

- [x] `internal/backlog/oauth_types_test.go`: `TokenParse…` decodes a token response; a missing `access_token`, a missing `refresh_token`, a non-positive `expires_in` or a `token_type` other than Bearer is rejected as `Unreachable`/`body`; `ExpiresAt` is the injected `Now` plus `expires_in`. `TokenSet` hides its tokens and `OAuthClient` hides its secret under `%v` and `%#v`, like `Credentials`.
- [x] `internal/backlog/projects_types_test.go` (`Projects…`): decoding the project list (id, `projectKey`, `name`, `archived`). A non-numeric id or an empty key is rejected. Archived projects are kept and flagged.
- [x] `internal/backlog/group_test.go` (`Queue…`, table-driven): GET on any path except the search paths → Read; `GET /api/v2/issues` and `GET /api/v2/issues/count` → Search; POST, PATCH and DELETE → Update; `POST /api/v2/oauth2/token` → Update.
- [x] `internal/connection/oauth_config_test.go` (`OAuth…`): `ParseOAuthConfig(map[string]any)` reads `oauth_client_id`, `oauth_client_secret` and `public_base_url`; any missing or blank gives `ErrOAuthNotConfigured`. `public_base_url` must be an absolute `https` URL with no query, fragment or userinfo; plain `http` only for `localhost` and `127.0.0.1` (assumption A2); one trailing slash is trimmed. `RedirectURI()` is exactly `<public_base_url>/api/plugins/nulab-backlog/webhooks/oauth-callback`. Error text never contains the client secret.
- [x] `internal/connection/state_token_test.go` (`OAuth…`): the token round-trips as base64url of (`workspaceId`, 32-byte nonce); malformed, truncated, empty, overlong and wrong-alphabet inputs are rejected; two tokens from the same workspace differ.
- [x] `internal/connection/project_key_test.go` (`Projects…`): `ValidateProjectKeys` trims each key, converts it to upper case, accepts `^[A-Z][A-Z0-9_]{0,24}$`, removes duplicates, rejects more than 100 keys, and rejects a non-string or a missing array (field `projectKeys`).
- [x] `internal/connection/change_test.go` (`ConnectionChanged…`, table-driven): the pure function `changeFor(prev record, next record)` returns `Reason` and `Restore`: first connect → `connected`, false; reconnect to the same host after a disconnect → `connected`, true; same host while connected → `credentials_replaced`, false; a different host → `space_changed`, true only when the new host equals the remembered previous host; disconnect → `disconnected`, false; projects added or removed → `projects_changed`, true when any project was added.
- [x] `internal/connection/record_compat_test.go` (`U2…`): a U1 `schemaVersion: 1` record and secret, which lack the new optional fields, still decode and still give `connected` with `authMethod: api_key`.
- [x] Run the tests and record the failing output.
- Stories: US1.3, US1.4, US1.7, US1.8, US1.9, US8.4. Contracts: C1 (TokenSet, Project), C3 (Reason, Restore), C6 (state). Findings: R-02, R-08. ACs: AC1.3.3, AC1.8.2, AC8.4.3.

### Step 4 — Data model, Green and Refactor

- [x] In `internal/backlog/oauth.go` and `projects.go`, implement the types only for now: `TokenSet`, `OAuthClient`, `parseTokenSet`, `Project`, `parseProjects`, and `group(method, path)`.
- [x] In `internal/connection/oauth.go`, `projects.go` and `change.go`, implement `OAuthConfig`, `ParseOAuthConfig`, `encodeState`/`decodeState`, `ValidateProjectKeys`, `Reason`, `changeFor`, and the new optional `record`/`secret` fields (record: `selectedProjects`, `previousSpaceHost`, `previousProjects`, `disconnected`, `signInAgain`; secret: `authMethod`, `accessToken`, `refreshToken`, `expiresAt`).
- [x] Refactor while green.
- Contracts: C1, C3, C6. Findings: R-02, R-08. NFRs: NFR3.

### Step 5 — Repository / data access, Red: OAuth secret, token update, disconnect record, pending state, projects

The new cases go into `internal/connection/store_test.go` and use the existing in-memory fakes, extended with `DeleteState`.

- [x] `OAuth…`: saving an OAuth connection writes the secret first, then the record; epoch = previous + 1; U1 rollback rules still apply. The view has `authMethod: oauth`, `hasOAuthToken: true`, `hasApiKey: false`.
- [x] `Token…`: `UpdateTokens` rewrites only the secret, keeps the epoch, never touches the record; if the epoch changed since the read, nothing is written and `ErrStale` is returned. `MarkSignInAgain` sets `signInAgain` on the record, and the view state becomes `sign_in_again`.
- [x] `OAuth…` pending state: `SavePending` writes workspace state key `oauth_pending` with `{schemaVersion, nonceHash, spaceHost, expiresAt}`; `TakePending` deletes the record before returning it (single use); a second `TakePending` gives `found == false`; the nonce comparison is constant-time over SHA-256 of the nonce; an expired record is deleted and refused.
- [x] `Disconnect…`: the secret is deleted first; the record is then rewritten as a disconnect record with epoch + 1 (`disconnected` true, `previousSpaceHost` and `previousProjects` kept, no user name); the view becomes `not_connected`; the next Connect gets epoch + 1 again, so the epoch never goes back; a failure to delete the secret writes no record and gives `ErrStore`.
- [x] `Projects…`: `SaveProjects` rewrites the secret (same values, epoch + 1), then the record with `selectedProjects`, in the BR2.8 order with the U1 rollback; a failed record write restores the previous secret.
- [x] `SpaceChange…`: saving a different host clears `selectedProjects` and remembers the old host and its projects in `previousSpaceHost` and `previousProjects`.
- [x] `Restore…`: saving the remembered previous host puts back `previousProjects` as `selectedProjects`.
- [x] `U2…`: every store call keeps the 1-second limit. Workspaces stay independent.
- [x] Run the tests and record the failing output.
- Stories: US1.3, US1.4, US1.5, US1.7, US1.8, US1.9. Rules: U1 BR2.8, BR2.11 (reused). Contracts: C3. ACs: AC1.5.4, AC1.8.2, AC1.4.2 (rotation stored).

### Step 6 — Repository / data access, Green and Refactor

- [x] In `internal/connection/store.go`, implement `SaveOAuth`, `UpdateTokens`, `MarkSignInAgain`, `SavePending`/`TakePending`, `Disconnect`, `SaveProjects`, and the space-change and restore handling in the shared write path.
- [x] Generalize the U1 `Save` into one `write(ctx, ws, rec, sec)` with the BR2.8 order and rollback, so every writer uses it.
- [x] Extend `viewOf`: connected when the secret holds an API key or an access token and its epoch and host match the record (BR2.11); `sign_in_again` when the record says so. Keep the U1 integration switch (`enabled`) behaviour unchanged.
- [x] Add `DeleteState` to `StateStore`.
- [x] Refactor while green.
- Rules: BR2.8, BR2.11. Contracts: C3. NFRs: NFR3, NFR5.

### Step 7 — Business logic, Red: BacklogGateway (headers, rate limits, queues, OAuth, projects, logs)

Tests use an `httptest` fake Backlog, an injected `Now` and `Wait`, run under `testing/synctest`, and never use a real sleep.

- [x] `APIKeyHeader…`: the API key is sent as `Backlog-API-Key` and never as a query parameter (replace U1's query-parameter test); an access token is sent as `Authorization: Bearer`; a request never carries both; URL redaction stays.
- [x] `RateLimit…` (table-driven, AC8.4.1): on a 429, the client waits until `X-RateLimit-Reset`; failing that, for `Retry-After`; failing that, 60 s; it then retries, and the fake sees the retry only after the virtual wait.
- [x] `RateLimit…` (AC8.4.2): after 3 retries (4 attempts in all) the client returns `RateLimited` with `RetryAfter`; a cancelled context during a wait returns `context.Canceled` with no further request; an `Interactive` call whose needed wait is over 3 s returns `RateLimited` at once; an `Interactive` call whose wait is 3 s or less waits and retries; a `Background` call waits the full time.
- [x] `Queue…` (AC8.4.3, `-race`): 3 goroutines send 10 Search-group and 10 Update-group requests to one host; the fake sees at most 1 concurrent request per group; each request in a group starts at least 1 s (virtual) after the previous one; Read-group calls are not queued; queues for two hosts do not block each other; for `Interactive` calls, queue time counts toward the 3-second limit.
- [x] `Queue…` logs: each wait writes one `backlog_wait` line with `group`, `reason` (`rate_limited` or `spacing`), `waitMs` and `attempt`, and no query string or secret (NFR11.3).
- [x] `OAuth…` in `oauth_test.go`: `ExchangeOAuthCode(ctx, spaceHost, OAuthClient, code, redirectURI)` POSTs the exact form fields to `/api/v2/oauth2/token` over https; 400 → `Invalid`, 401 → `Unauthorized`; the U1 redirect, 1 MiB and 10-second limits apply; error text and logs contain neither the client secret, the code nor any token.
- [x] `Token…` in `oauth_test.go`: `RefreshToken(ctx, spaceHost, OAuthClient, refreshToken)` sends the refresh grant and returns the rotated refresh token; 400 `invalid_grant` → `Invalid`; 401 → `Unauthorized`.
- [x] `Projects…` in `projects_test.go`: `Projects(ctx, Credentials)` calls `GET /api/v2/projects` with each credential kind; list and empty 200, plus 401, 403 and 500 with a bait body; the bait never appears in the error or the logs.
- [x] Run the tests and record the failing output.
- Stories: US8.4, US8.3, US1.3, US1.4, US1.7. Contracts: C1, C7. NFRs: NFR2, NFR3, NFR5, NFR11. ACs: AC8.4.1–AC8.4.3, AC8.3.1, AC7.3.2 (gateway part), AC1.1.7.

### Step 8 — Business logic, Green and Refactor: BacklogGateway

- [x] Implement: the auth header choice; one `send(ctx, req, group, class)` path with the retry loop (at most 3 retries) using the injected `Wait func(ctx, time.Duration) error` (default honours `ctx`); a per-(host, group) FIFO slot (`sync.Mutex` plus last-start time) for Search and Update with 1-second spacing; the `Interactive` 3-second budget covering queue and 429 waits; `backlog_wait` logging; `ExchangeOAuthCode`, `RefreshToken` and `Projects`.
- [x] Every secret (API key, access token, refresh token, client secret, code) is added to the context's redaction set before the request is built.
- [x] Refactor while green. The `Myself` contract stays unchanged.
- Contracts: C1, C7. NFRs: NFR2, NFR3, NFR5, NFR11.3.

### Step 9 — Business logic, Red: Connection service (reader, OAuth, re-check, disconnect, replace, projects, events)

Tests use the gateway fake, the store fakes, an injected clock and `synctest`.

- [x] `ConnectionReader…` (C3): `Current(ctx, ws)` returns `Snapshot{SpaceHost, AuthMethod, ConnectionEpoch, SelectedProjects}` or `ErrNotConnected`; `Credentials(ctx, ws)` applies BR2.11 before handing out any secret; a mismatched record and secret gives `ErrNotConnected`.
- [x] `Token…` (AC1.4.1–AC1.4.3): a token with less than 5 minutes left is refreshed before it is returned; five goroutines (`-race`) holding an expired token cause exactly 1 refresh request, all 5 get the new token, the rotated refresh token is stored with the epoch unchanged; a refused refresh (`Invalid`/`Unauthorized`) makes exactly one attempt, sets `MarkSignInAgain`, returns `ErrReconnectRequired`, and later `Credentials` calls make 0 Backlog requests until a new sign-in; an `Unreachable` refresh does not set `signInAgain`; a refresh finishing after the epoch changed writes nothing and returns `ErrReconnectRequired`.
- [x] `OAuth…` StartOAuth (AC1.3.1): the space address is validated as in U1; missing config gives `ErrOAuthNotConfigured` (field `oauth`); the pending state expires after 10 minutes; `authorizeUrl` is exactly `https://<host>/OAuth2AccessRequest.action` with `response_type=code`, `client_id`, `redirect_uri` and `state`; while Backlog is off for the workspace, the result is `integration_disabled`.
- [x] `OAuth…` CompleteOAuth (AC1.3.2, AC1.3.3), outcomes `connected`, `cancelled`, `failed`: `error=access_denied` deletes the pending state, gives `cancelled`, stores nothing; a missing, malformed, unknown, used or expired `state`, or a missing `code`, gives `failed` and 0 token requests; a failed exchange gives `failed` and stores nothing; on success: exchange, `Myself` with the Bearer token, a second switch check, `SaveOAuth`, then a `ConnectionChanged` event; Backlog turned off during the exchange gives `failed` and stores nothing.
- [x] `Recheck…` (AC1.5.1, AC1.5.2): `Test(ctx, ws)` calls `Myself` with the current credentials and returns the view with the fresh user name; a 401 gives `ErrReconnectRequired` and changes no stored state; `Unreachable` and `RateLimited` pass through.
- [x] `Disconnect…` (AC1.5.4): secrets deleted, then the disconnect record, then exactly one `disconnected` event with the new epoch; the pending OAuth state is deleted; afterwards `Credentials` gives `ErrNotConnected` and the gateway fake sees 0 calls; disconnecting when not connected is a successful no-op with no event.
- [x] `Replace…` (AC1.6.1, AC1.6.2): a verified new key on the same host gives `credentials_replaced`, epoch + 1, selected projects kept; a 401 keeps the old record and secret, no event.
- [x] `SpaceChange…` (AC1.8.2, AC1.8.3): a new host gives `space_changed`, the old secret is replaced, selected projects are cleared; a `Credentials` result taken before the change carries the old epoch and a later `Current` reports the higher epoch, so receivers can drop the late result.
- [x] `Restore…` (M12): reconnecting to the remembered previous host after a disconnect or a space change sends `Restore: true` with the previous projects reselected, and the view carries `restored: true`.
- [x] `Projects…` (AC1.7.1, AC1.7.3, AC1.9.2): `ListProjects` returns `[{projectKey, projectId, projectName, selected}]` (empty list valid); `SetProjects` calls `Projects` once, rejects keys not on the space (`validation`, field `projectKeys`), stores the selection, sends `projects_changed` with the new key list and `Restore` when a key was added; an unchanged selection sends no event.
- [x] `ConnectionChanged…`: subscribers get events in epoch order; a slow subscriber does not block the action (per-subscriber goroutine queue); a panicking subscriber is recovered and logged; events go out only after the state write succeeded.
- [x] `U2…` logs: exactly one line each for `oauth_started`, `oauth_completed`, `oauth_failed` (with `reason`), `token_refreshed`, `token_refresh_failed`, `connection_tested`, `connection_changed` (`reason`, `connectionEpoch`, `restore`, `projectCount`).
- [x] `U2…` leak test: logs, errors, views and events contain no window of the API key, access token, refresh token, client secret, code or state nonce, and no display name in the logs.
- [x] Run the tests and record the failing output.
- Stories: US1.3, US1.4, US1.5, US1.6, US1.7, US1.8, US1.9, US8.3. Contracts: C3, C6. Findings: R-08. ACs: AC1.3.1–3, AC1.4.1–3, AC1.5.1–2, AC1.5.4 (non-Git part), AC1.6.1–2, AC1.7.1, AC1.7.3, AC1.8.2–3 (non-Git part), AC1.9.2 (non-Git part), AC7.3.2 (token part).

### Step 10 — Business logic, Green and Refactor: Connection service

- [x] Implement `Service.Current` and `Service.Credentials` with a per-workspace refresh mutex that re-reads the secret after taking the lock (standard library only, no `singleflight`); `StartOAuth` and `CompleteOAuth`; `Test`, `Disconnect`, `ListProjects` and `SetProjects`; Reason and Restore handling on `Connect`; `Subscribe(func(ConnectionChanged)) (unsubscribe func())` with per-subscriber buffered delivery.
- [x] Every write (Connect, CompleteOAuth, Disconnect, SetProjects, UpdateTokens) runs under one per-workspace write lock. U1's try-lock for Connect stays.
- [x] Extend `Classify`: `ErrReconnectRequired` and `backlog` `Unauthorized`/`Forbidden` from non-Connect calls → `reconnect_required`; `ErrNotConnected` → `reconnect_required`; `ErrOAuthNotConfigured` → `validation` field `oauth`; `ErrInvalidProjects` → `validation` field `projectKeys`.
- [x] Refactor while green.
- Contracts: C3, C5 (codes), C6. Findings: R-06, R-08. NFRs: NFR3, NFR5, NFR11.

### Step 11 — API / endpoint, Red: actions, webhook, config, manifest, action logs

- [x] `internal/plugin/actions_test.go`, through `HandleAction` with the fake Host: `Recheck…` `connection.test` success and 401 → 401 `reconnect_required`; `Disconnect…` `connection.disconnect` returns the view and the Host secrets and pending state are gone; `Projects…` `connection.list_projects` and `connection.set_projects`, a bad body gives 400 `validation` on `projectKeys`, a 429 gives `Retry-After` and `retryAfterSeconds`; `OAuth…` `connection.start_oauth` returns `{authorizeUrl}` and reads config through `Host.GetConfig` on every call, missing config → 400 `validation` field `oauth`, a `GetConfig` failure → 500 `internal`; every new action except `connection.get` and `connection.set_enabled` is refused with 409 `integration_disabled` while Backlog is off (U1 guard, unchanged); the workspace always comes from the verified context.
- [x] `ActionFailureLog…` (AC8.3.1): a fake 500 during `connection.test` writes exactly one WARN `action_failed` line with `action`, `errorCode`, `durationMs`, `workspaceId` and `requestId`, and no secret.
- [x] `internal/plugin/webhook_test.go`, `Webhook…`: `oauth-callback` with each outcome answers `302`; `Location` is exactly `/settings/workspaces/<ws>/integrations/nulab-backlog?oauth=connected|cancelled|failed`, plus `&restored=1` on a restore, and never contains the code, a token, the state or Backlog error text; an undecodable state gives `302` to `/settings/integrations?oauth=failed`; an unknown webhook key gives 404; a non-GET method gives 405; a panic is recovered as `302 …oauth=failed`; `HandleWebhook` never returns a Go error.
- [x] `internal/plugin/config_test.go`, `OAuth…`: the `hostConfig` adapter passes `GetConfig` through; a missing Host gives `internal`.
- [x] `internal/plugin/manifest_test.go`, `U2Manifest…`: new actions with `scope`/`access` — `connection.start_oauth` workspace/admin, `connection.test` workspace/authenticated, `connection.disconnect` workspace/admin, `connection.list_projects` workspace/authenticated, `connection.set_projects` workspace/admin — each with `max_body_bytes: 8192` and a key matching the Kandev pattern; `webhooks: [{key: oauth-callback, method: GET, access: public}]`; `config_schema` has `oauth_client_id` (string), `oauth_client_secret` (string, `secret: true`) and `public_base_url` (string), none `required`.
- [x] Run the tests and record the failing output.
- Stories: US1.3, US1.5, US1.7, US1.9, US8.3, US8.4. Contracts: C5, C6, C8. Findings: R-01, R-02, R-06. ACs: AC1.3.1–3, AC1.5.1–2, AC1.5.4, AC1.7.1, AC8.3.1, AC8.4.2.

### Step 12 — API / endpoint, Green and Refactor

- [x] `internal/plugin/runtime.go`: add the 5 handlers and the new codes to `statusFor` (`reconnect_required` → 401); log `action_failed` (WARN) for every non-validation failure and keep ERROR for `internal`; wire `connection.Service` with `backlog.NewClient()` as the gateway.
- [x] `internal/plugin/webhook.go`: `HandleWebhook` routes `oauth-callback`, parses `req.Query` with `url.ParseQuery`, calls `CompleteOAuth` under the 12-second budget, builds the relative `Location`, and logs `oauth_callback` with the outcome only.
- [x] `internal/plugin/config.go`: the `hostConfig` adapter, `GetConfig` → `map[string]any`.
- [x] `manifest.yaml`: add the actions, the webhook and the `config_schema`.
- [x] Refactor while green.
- Contracts: C5, C6, C8. Findings: R-01, R-02. NFRs: NFR3, NFR11.

### Step 13 — Frontend behaviour, Red: M1 additions and the M12 notice

All tests render with the existing fake `host` (`ui/src/testing/harness.ts`), run `axe-core` with no violations allowed, check `data-testid` on every interactive element, and check that no literal text appears outside `messages/en.ts`.

- [x] `ui/src/settings/state.test.ts`, pure reducer: `reconnect_required` → "key is invalid or has been revoked" notice; `sign_in_again` view → Sign-in-again state; `rate_limited` → "Backlog is limiting requests. Retrying in N seconds", announced once; the `oauth` query outcomes `connected`, `cancelled`, `failed` → matching notices; `restored: true` → M12 notice; `validation`/`oauth` → "OAuth is not set up on this Kandev server"; `validation`/`projectKeys` → field error.
- [x] `ui/src/settings/oauth.test.tsx`: sign-in method `radiogroup` (API key or OAuth), API key default; "Sign in with Nulab" calls `connection.start_oauth` and navigates with `window.location.assign` only when `authorizeUrl` starts with `https://<the entered host>/OAuth2AccessRequest.action`, otherwise `connectFailed`; the button shows "Connecting…" and is locked while waiting; on mount `?oauth=cancelled` and `?oauth=failed` show and announce their notices exactly once; the Sign-in-again state shows a "Sign in again" button that starts OAuth with the stored host (AC1.4.3).
- [x] `ui/src/settings/connected-panel.test.tsx`: Test connection shows "Testing…", is locked, and on success shows and announces the user name (AC1.5.1); a 401 shows the revoked-key message (AC1.5.2); Disconnect opens the confirm dialog (`role="alertdialog"`, text saying credentials are deleted for everyone in the workspace, default focus on Cancel, Esc closes, focus returns to Disconnect) (AC1.5.3); Confirm calls `connection.disconnect` and shows the not-connected form; "Replace credentials" with the same host opens the replace confirmation and with a different host the change-space confirmation (which says selected projects are cleared); Cancel calls nothing (AC1.6.3); the M12 notice "Reconnected to <host>. Items from your earlier connection to this space were restored." is shown and announced.
- [x] `ui/src/settings/project-picker.test.tsx`: projects load with `connection.list_projects` into checkboxes inside a `fieldset` with a legend, and a search field filters them (AC1.7.1); an empty list shows "No projects available" (AC1.7.3); Save calls `connection.set_projects` with the checked keys; unchecking a selected project and saving opens a confirmation first, and Cancel saves nothing (AC1.9.1); a `rate_limited` reply shows the countdown and retries the load once at 0 with `vi.useFakeTimers` (AC8.4.4).
- [x] `ui/src/settings/confirm-dialog.test.tsx`: the shared dialog traps focus, closes on Esc, returns focus to its opener, shows "Working…" with a disabled confirm button while working, and keeps the dialog open with an inline error on failure.
- [x] Update `ui/src/settings/settings.test.tsx` (U1) only where the connected layout changed.
- [x] Run the tests and record the failing output.
- Stories: US1.3, US1.4, US1.5, US1.6, US1.7, US1.8, US1.9, US8.4. ACs: AC1.3.1–2, AC1.4.3 (settings part), AC1.5.1–3, AC1.6.3, AC1.7.1, AC1.7.3, AC1.8.1 (without counts), AC1.9.1 (without counts), AC8.4.4. NFRs: NFR9.

### Step 14 — Frontend behaviour, Green and Refactor

- [x] Implement kebab-case files `ui/src/settings/confirm-dialog.tsx`, `ui/src/settings/connected-panel.tsx` (Test connection, Disconnect, Replace credentials and the M12 notice), `ui/src/settings/project-picker.tsx`, `ui/src/settings/oauth.ts` (start, URL guard, reading the query result).
- [x] Extend `state.ts` and `SettingsScreen.tsx` without breaking the U1 switch-follow behaviour (`enabled-events.ts`). Add the new message keys to `messages/en.ts`.
- [x] Layout uses only host UI kit components and the U1 spacing classes (BR6.5).
- [x] Refactor while green.
- Rules: BR6.1–BR6.5 (reused). NFRs: NFR9, NFR10 (catalogue only).

### Step 15 — Environment and build configuration

- [x] No `Makefile` or CI change. The existing targets already cover the new files: `GO_PKGS` is `./internal/... ./server/...` and Vitest runs every file.
- [x] Run `make check-format vet lint test coverage check-secrets build package verify-package` from `make clean`. Confirm that Go coverage is at least 80% with only `server/main.go` excluded, and that `go mod tidy` leaves no diff.
- [x] Confirm that the packaged `manifest.yaml` carries the webhook and `config_schema`, and that `verifypkg` reports OK. Run `make contract-test KANDEV_MIN_DIR=../kandev-min` and confirm it still passes with the new manifest.
- Stories: US7.3 (gate applies). NFRs: NFR8.

### Step 16 — Documentation and traceability

- [x] `README.md`: append a section "Sign in with OAuth": registering a Nulab OAuth app (X5), the exact redirect URI `<public_base_url>/api/plugins/nulab-backlog/webhooks/oauth-callback`, and the three config fields in Settings > Plugins > Nulab Backlog. Only append.
- [x] `docs/manual-checks/TEMPLATE.md`: add steps for OAuth sign-in, the past-expiry refresh, disconnect and reconnect with the restore notice, and a fake-429 wait log line, matching the B3 demo.
- [x] Write `construction/connection/code-generation/code-summary.md`, `source-manifest.json` and `traceability.json`, including the upstream amendments listed below.
- Stories: all U2.

## Story-to-Step Map

| Story | Steps |
|-------|-------|
| US8.4 Respect Backlog's API rate limits | 3, 4, 7, 8, 11, 12, 13, 14 |
| US8.3 Logs for diagnosing errors | 7, 8, 9, 10, 11, 12 |
| US1.5 Re-check and disconnect | 5, 6, 9–14 |
| US1.6 Replace credentials in the same space | 5, 6, 9, 10, 13, 14 |
| US1.7 Choose projects | 3–6, 7–14 |
| US1.3 Connect with OAuth | 3–14, 16 |
| US1.4 Refresh the OAuth sign-in automatically | 3–10, 13, 14 |
| US1.8 Switch to another space | 3–6, 9, 10, 13, 14 |
| US1.9 Unselect a project in use | 3–6, 9, 10, 13, 14 |

**Shared files touched**: `manifest.yaml` (actions, webhook, `config_schema`), `README.md` (one appended section), `docs/manual-checks/TEMPLATE.md`, `internal/plugin/runtime.go`, `internal/connection/{store.go,service.go}`, `internal/backlog/client.go`, `internal/testutil/testutil.go`, `ui/src/messages/en.ts`, `ui/src/settings/{SettingsScreen.tsx,state.ts,settings.test.tsx}`. Not touched: `Makefile`, `.github/workflows/*`, `.golangci.yml`, `go.mod`, `go.sum`, `ui/package.json`, `internal/ci`, `cmd/`.

## Sources

- [desc] Initial description: Kandev plugin for Nulab Backlog.
- [scope] Workflow-selected scope: `feature`.
- `unit-of-work.md`, `unit-of-work-story-map.md`, `unit-of-work-dependency.md`; `stories.md`; `requirements.md`; `components.md`, `decisions.md`; `contract-summary.md` and `contract-design/reviews/review-01.md` (R-01, R-02, R-06, R-07, R-08); `mockups.md` (M1, M12), `interaction-spec.md`, `accessibility-checklist.md`; `bolt-plan.md`, `external-dependency-map.md`.
- U1: `functional-spec.md`, `rules.md`, `entities.md`, NFR design, `code-summary.md`, and the code at `99760f8`.
- `team.md`, `project.md`, `phases/construction.md`.
- Kandev v0.96.0: `pkg/pluginsdk/{plugin.go,host.go,types.go}`; `internal/plugins/handlers.go`; `internal/plugins/manifest/manifest.go`; `docs/public/plugins-manifest.md`; `docs/public/plugins-authoring.md`; `apps/packages/plugin-sdk/src/index.ts`; `apps/web/src/settings-routes.plugin.test.ts`.
- Backlog: <https://developer.nulab.com/docs/backlog/auth> and <https://developer.nulab.com/docs/backlog/rate-limit> (read 2026-10-06).

## Assumptions & Open Questions

- [assumption] There is no U2 functional, NFR or infrastructure design (user choice). The bolt plan asked for R-01, R-02, R-06 and R-08 to be settled in functional design before code. They are settled by the assumptions below and need the user's review at plan approval.
- [assumption] (R-01) `connection.start_oauth`, `connection.disconnect` and `connection.set_projects` are `access: admin`, because each changes the connection shared by the whole workspace. `connection.test` and `connection.list_projects` are `authenticated`.
- [assumption] (R-02) The OAuth client id, client secret and Kandev's public base URL come from the plugin `config_schema` (Settings > Plugins), with the client secret marked `secret: true`. The plugin reads them with `Host.GetConfig` on every `start_oauth` and callback. One OAuth app serves the whole Kandev instance. The redirect URI is `<public_base_url>/api/plugins/nulab-backlog/webhooks/oauth-callback`.
- [assumption] A Nulab OAuth app works for any Backlog space the user signs in to. Not verified; X5 confirms it.
- [assumption] (A2) `public_base_url` may be `http` only for `localhost` and `127.0.0.1`, for trial runs. Whether Nulab accepts such a callback is unverified.
- [assumption] PKCE is not used, because Backlog does not document it. The `state` is a 256-bit random nonce with the `workspaceId`, base64url-encoded, checked in constant time against a single-use server-side pending record (workspace state key `oauth_pending`) that expires after 10 minutes. This replaces the contract's "signed state" with the same guarantee: no forging, no replay. One pending sign-in per workspace; a new `start_oauth` replaces the old one.
- [assumption] (R-08) The epoch increases on every write of the connection: connect, credential replacement, space change, disconnect and project change. A token refresh does not change the epoch. Project changes rewrite the secret with the new epoch, so U1's BR2.11 pair rule stays the only consistency rule.
- [assumption] Disconnect keeps a disconnect record (epoch + 1, previous host and projects, no secret) instead of deleting the record, so the epoch never decreases and a later restore can be detected. New optional fields keep `schemaVersion: 1`, because U1 readers ignore unknown fields.
- [assumption] Restore: `Restore` is true when the new host equals the remembered previous host, or when a project selection adds a key. On a restore, the previous project selection is put back.
- [assumption] The M12 notice in U2 has no link or watch counts and no "Review watches" button; those belong to U3/U4 data (ADR-004). The disconnect, change-space and deselect dialogs show no counts (AC1.8.1/AC1.9.1 counts deferred). U3/U4 add an impact-count hook.
- [assumption] The confirmation before replacing credentials or changing space happens in the UI only. The backend replace path stays U1's verified-replace (BR2.5).
- [assumption] The token refresh threshold is 5 minutes before `expiresAt`. Backlog tokens last 3600 s.
- [assumption] A refresh refused with 400 (`invalid_grant`) or 401 means "sign in again". `Unreachable` and `RateLimited` refreshes are temporary and leave the state unchanged.
- [assumption] (R-06) For U2 calls (`users/myself`, `projects`), both 401 and 403 map to `reconnect_required`. During `connect_api_key` they still map to `validation` on the key field (U1). A per-project 403 is a U3 question.
- [assumption] (R-07) Calls without a `CallClass` (`Myself`, `Projects`, `ExchangeOAuthCode`, `RefreshToken`) are `Interactive`.
- [assumption] Rate-limit queues are per group (Search and Update separately, ADR-003) and per space host. The token endpoint counts as Update. Read calls are not queued but still retry on 429. No early slow-down from `X-RateLimit-Remaining`.
- [assumption] Retries: at most 3 retries after the first 429, then `rate_limited`. `Interactive` calls never wait more than 3 s in total.
- [assumption] The UI retries automatically after a `rate_limited` reply only for reads (`list_projects`), never for writes.
- [assumption] The API key moves from the `apiKey` query parameter (U1 BR3.4) to the `Backlog-API-Key` header, following NFR3, AC1.1.7, ADR-002 and Backlog's documentation. URL redaction stays. If the user prefers U1's behaviour, drop the `APIKeyHeader…` items.
- [assumption] C1 signatures change: `ExchangeOAuthCode(ctx, spaceHost, OAuthClient, code, redirectURI)` and `RefreshToken(ctx, spaceHost, OAuthClient, refreshToken)` (no `codeVerifier`, client credentials passed in). C3 methods take a `workspaceID`. Contracts C1, C3 and C5 (snake_case keys, `set_enabled`, `integration_disabled`) need upstream amendments.
- [assumption] The update interval (`setPollInterval`, `Snapshot.PollMinutes`) is left to U3, which owns US4.2.
- [assumption] `start_oauth` without config returns `validation` on field `oauth`, and the UI explains that an admin must set it up. `connection.get` does not call `GetConfig`.
- [assumption] The OAuth callback redirects to a relative `Location` on the Kandev origin; when the state is undecodable it goes to `/settings/integrations?oauth=failed`. The callback uses the 12-second action budget.
- [assumption] U1's integration-switch guard still blocks every new action while Backlog is off, `disconnect` included. A callback that arrives while Backlog is off gives `oauth=failed`.
- [assumption] New Go test secrets use U1's `test-api-key-` style run-time helper, which U5's secret scan accepts because the values are never committed.
- [assumption] `ConnectionChanged` has no subscriber yet in U2. Delivery is tested with test subscribers; U3 and U4 subscribe.
- Open: the Git-password parts of AC1.5.4, AC1.8.2 and AC1.9.2 are tested in U4 through `ConnectionChanged`.
- Open: the manual OAuth demo (B3) needs X5, a registered Nulab OAuth app. The automated tests do not.
