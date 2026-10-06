# Code Generation Plan — walking-skeleton (U1)

Inputs:

- `functional-spec`, `rules` and `entities` for U1.
- U1 NFR requirements and NFR design.
- `infrastructure-specification` and `cicd-pipeline` for U1.
- `contract-summary`: C1, C4, C5, C8.
- `unit-of-work`.
- `requirements`.
- `team-practices`.

Stories in U1: US7.1, US1.2, US1.1, US7.2.

## Kandev SDK Facts Used by This Plan

These are taken from the `../kandev` checkout:

- **Go module**:
  - Module `github.com/kandev/kandev`, rooted at `apps/backend`, Go 1.26.
  - Import the SDK as `github.com/kandev/kandev/pkg/pluginsdk`.
  - In `go.mod`, add `replace github.com/kandev/kandev => ../kandev/apps/backend`.
- **Plugin interface**:
  - Embed `pluginsdk.UnimplementedPlugin`, which provides `SetHost` and `Host()`.
  - Implement `ActionHandler.HandleAction(ctx, *PluginActionRequest) (*PluginActionResponse, error)`.
  - `PluginActionRequest.Context.WorkspaceID` is verified by Kandev.
- **Host stores**:
  - State: `GetState(ctx, scope, scopeID, key) (map[string]any, found, error)` and `SetState`.
  - Secrets: `GetSecret(ctx, key) (string, found, error)`, `SetSecret` and `DeleteSecret`.
  - A missing key returns `found == false` with no error.
- **Packaging**:
  - `go run github.com/kandev/kandev/cmd/plugin-pack -dir <stage> -out <id>-<version>.tar.gz`. This generates `checksums.txt` inside the archive.
  - `go run github.com/kandev/kandev/cmd/plugin-package-verify -archive <file> -expected-id <id> -expected-version <ver>`.
- **UI**:
  - One self-contained ES module, `ui/bundle.js`, that calls `window.registerKandevPlugin("<id>", { initialize(registry, host) {}, destroy() {} })`.
  - Rendering uses `host.jsx` and the components in `host.ui`.
  - The settings screen is registered with `registerIntegrationSettings`, which also takes an `icon` and an `action` component (the card's switch slot).
  - The home Integrations menu entry is `registerNavItem({ section: "integrations" })`, and its page is `registerRoute`.
  - The per-workspace switch is `host.ui.IntegrationEnabledControl`; the value is published with `host.setIntegrationEnabled`.
  - Backend calls use `host.api.invokeAction`.
- **Action keys**: Kandev only accepts keys matching `^[a-z0-9][a-z0-9._-]*$`, so the code uses `connection.connect_api_key` and `connection.set_enabled`.
- **Package verification**: `plugin-package-verify` does not exist at the v0.96.0 tag, so `verify-package` runs the in-repo verifier `cmd/verifypkg` (`internal/pkgverify`).
- **Admin status**: there is no admin flag for the viewer. The M1 screen uses the fallback from `functional-spec`: it shows the form, and a 403 from `connectApiKey` switches to the member message.

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

Each TDD layer follows Red → Green → Refactor. Each Red step records the failing command output in `code-summary.md`.

### Step 1 — Project structure and production configuration skeleton

- [x] Create `go.mod`:
  - module `github.com/khuongdo/kandev-plugin-nulab-backlog`;
  - `go 1.26`;
  - `require github.com/kandev/kandev` with `replace github.com/kandev/kandev => ../kandev/apps/backend`;
  - `require github.com/stretchr/testify`.
- [x] Create `.kandev-sdk-ref`, holding the full commit SHA of the current `../kandev` HEAD (the v0.96.0 tag).
- [x] Create `manifest.yaml`:
  - `id: nulab-backlog`, `api_version: 2`, `version: 0.0.1`, `min_kandev_version: "0.96.0"`;
  - `runtime.type: binary` with the 5 executables of BR5.1;
  - `capabilities: {state: true, secrets: true}`;
  - actions `connection.get` (workspace, authenticated) and `connection.connectApiKey` (workspace, admin), each with `max_body_bytes: 8192`;
  - `ui.bundle: /ui/bundle.js`.
- [x] Create `server/main.go`. It contains only `pluginsdk.Serve(plugin.NewRuntime())`.
- [x] Create empty packages `internal/plugin`, `internal/connection`, `internal/backlog` and `internal/redact`, each with a doc comment.
- [x] Create `ui/package.json` (private; devDependencies `typescript`, `vitest`, `jsdom`, `esbuild`, `eslint`, `prettier`, `axe-core`), `ui/tsconfig.json` (`strict: true`), and `.nvmrc` (Node 22).
- [x] Add `LICENSE` (MIT). Add `dist/`, `node_modules/` and `coverage.out` to `.gitignore`.
- Stories: US7.1. Rules: BR5.1, BR5.5.

### Step 2 — Bootstrap the test runners and record the unit-scoped commands

- [x] Add a Go smoke test in `internal/redact`, so that `go test -race ./internal/redact/...` runs.
- [x] Add `ui/vitest.config.ts` (environment `jsdom`) and a smoke test, so that `npx vitest run src/settings` runs from `ui/`.
- [x] Record both exact commands in `unit-test-instructions.md`. They are already written there.
- Stories: US7.1.

### Step 3 — Data model, Red: value objects and input validation

- [x] `internal/connection/address_test.go`: a table-driven test with every row of the functional-spec "Address examples" table (BR1.1, BR1.2), including trimming, `://` detection, an upper-case scheme, a non-ASCII label and a 64-character label.
- [x] `internal/connection/apikey_test.go`: trimming, empty, 257 characters, an inner space, a control character, and a valid key (BR3.5).
- [x] `internal/backlog/types_test.go`: `User` decoding (missing `id`, empty `name`, non-numeric `id`). Status-to-Kind mapping for 200, 401, 403, 404, 409, 400, 422, 429, 418, 302 and 503 (entities.md, C1 `Error`). `Error.Error()` shows only Kind and Status.
- [x] Run the tests and record the failing output.

### Step 4 — Data model, Green and Refactor

- [x] Implement `connection.ParseSpaceAddress`, `connection.ValidateAPIKey`, `backlog.User`, `backlog.Error`, `backlog.Kind`, `backlog.CallClass` and `backlog.Credentials` until the tests pass.
- [x] Refactor while green.
- Rules: BR1.1, BR1.2, BR3.5. NFR3.7.

### Step 5 — Repository / data access, Red: epoch-linked storage

- [x] `internal/connection/store_test.go` uses in-memory fakes of two small interfaces declared in `internal/connection`: `SecretStore` (Get, Set, Delete) and `StateStore` (Get, Set). Cases:
  - first save writes the secret, then the record, with epoch 1;
  - replace writes epoch previous + 1;
  - a record-write failure restores the previous secret;
  - a record-write failure on a first connect deletes the new secret;
  - a rollback failure leaves an `error` view;
  - reading with no record gives `not_connected`;
  - a record with a matching secret gives `connected`;
  - a record with a mismatched or missing secret gives `error`;
  - the rollback still runs when the action context is cancelled (fresh context);
  - the next epoch never decreases when the secret is missing but the record exists (epoch = max(record, secret) + 1).
- [x] Run the tests and record the failing output.

### Step 6 — Repository / data access, Green and Refactor

- [x] Implement `connection.Store` with BR2.8 ordering and BR2.11 reconciliation.
  - Secret key: `backlog.connection.<workspaceId>`.
  - State: scope `workspace`, key `connection`, JSON with `schemaVersion: 1`.
  - Each store call has a 1-second limit. The rollback uses a fresh 2-second context.
- [x] Refactor while green.
- Rules: BR2.8, BR2.11, BR3.1. NFR5.4, NFR5.5, NFR5.7.

### Step 7 — Business logic, Red: redaction, gateway and Connect service

- [x] `internal/redact/redact_test.go`:
  - registered secrets are masked;
  - Backlog URL queries are replaced with `?REDACTED`;
  - the `slog` handler wrapper masks attributes;
  - masking works for a value nested in an error chain.
- [x] `internal/backlog/client_test.go` uses an `httptest` fake Backlog with JSON in `internal/backlog/testdata/`. Cases:
  - `Myself` 200;
  - the `apiKey` query parameter is sent query-encoded over https only;
  - a redirect is not followed and the other host gets 0 requests;
  - a body over 1 MiB gives `Unreachable`;
  - a timeout gives `Unreachable`, using an injected short deadline;
  - a cancelled context returns `context.Canceled` unchanged;
  - a 429 gives `RateLimited` with RetryAfter from `X-RateLimit-Reset`, then `Retry-After`, then 60, clamped to at least 1 s;
  - transport errors never return `*url.Error` text, and the error string holds no key or query.
- [x] `internal/connection/service_test.go` covers `Connect` with fakes. Cases:
  - every row of the functional-spec WF3 outcome table;
  - nothing is written on any failure;
  - a second Connect in the same workspace gets `conflict` (`-race`, two goroutines);
  - Connects in different workspaces run in parallel;
  - the 13-second deadline with an 8-second Backlog sub-deadline, using an injected clock and a fake that blocks;
  - log events `connect_succeeded` and `connect_failed` with the NFR11.2 fields;
  - a leak test: logs, errors and views contain neither the key nor any 8-character window of its random part, nor the display name in logs.
- [x] Run the tests and record the failing output.

### Step 8 — Business logic, Green and Refactor

- [x] Implement:
  - `redact` (mask set carried on the context, URL query masking, `slog.Handler` wrapper);
  - `backlog.Client.Myself` (shared `http.Client`; `CheckRedirect` returns an error; https-only transport; 1 MiB limited reader; Kind mapping);
  - `connection.Service.Connect` (validation → try-lock → Myself → Store.Save → view, under the 13-second deadline).
- [x] Refactor while green.
- Rules: BR1.3, BR1.4, BR2.2–BR2.7, BR2.9, BR2.10, BR3.3, BR3.4, BR4.1, BR4.2. NFR1.2, NFR2.1, NFR3.3, NFR3.4, NFR3.6, NFR5.1–NFR5.3, NFR5.6, NFR5.8, NFR11.2–NFR11.5.

### Step 9 — API / endpoint, Red: Kandev adapter

- [x] `internal/plugin/actions_test.go` drives `HandleAction` with a fake `pluginsdk.Host`. Cases:
  - `connection.get` in each state;
  - `connection.connectApiKey` success;
  - every error code maps to its HTTP status and `ActionError` JSON, with `Retry-After` on 429;
  - `requestId` is in both the error body and the logs;
  - an unknown action key gives 404 JSON;
  - a malformed body gives 400 `validation`;
  - a panic in a handler is recovered as 500 `internal`;
  - the handler never returns a Go error;
  - responses never contain the key.
- [x] `internal/plugin/manifest_test.go` parses `manifest.yaml` and asserts:
  - the action keys and `access` values (NFR3.5);
  - `min_kandev_version: "0.96.0"` (NFR6.1);
  - the 5 executables (BR5.1);
  - `capabilities.state` and `capabilities.secrets`.
- [x] Run the tests and record the failing output.

### Step 10 — API / endpoint, Green and Refactor

- [x] Implement `plugin.NewRuntime()`:
  - it embeds `pluginsdk.UnimplementedPlugin`;
  - a `HandleAction` router;
  - one error-code mapping table;
  - store adapters from `pluginsdk.Host` to `connection.SecretStore` and `connection.StateStore`;
  - `requestId` creation;
  - a `plugin_started` log line.
- [x] Refactor while green.
- Rules: BR2.1, BR5.4. NFR3.2, NFR3.5, NFR5.3, NFR6.1, NFR11.1.

### Step 11 — Frontend behaviour, Red: M1 settings screen

- [x] `ui/src/settings/settings.test.tsx` renders the screen with a fake `host`. It covers every M1 state in `functional-spec`: loading, load failed, not connected, connected, error, connecting (button locked), field errors on `spaceUrl` and `apiKey`, unreachable, rate limited, busy, and the member message after a 403.
- [x] It also checks that the key input is cleared after Connect (BR6.3), that each field error is linked to its input with `aria-describedby` and `aria-invalid`, and that the result is announced in a polite live region (BR6.4).
- [x] It checks that `data-testid` attributes are present on interactive elements, that `axe-core` finds no violations, and that there is no literal text outside the message catalogue (BR6.2).
- [x] Run the tests and record the failing output.

### Step 12 — Frontend behaviour, Green and Refactor

- [x] Implement:
  - `ui/src/index.ts` (`window.registerKandevPlugin("nulab-backlog", …)`, which calls `registerIntegrationSettings`);
  - `ui/src/settings/SettingsScreen.tsx` (rendered with `host.jsx` and `host.ui`);
  - `ui/src/settings/state.ts` (a pure state machine that maps action results to screen states);
  - `ui/src/messages/en.ts`.
- [x] Refactor while green.
- Rules: BR6.1–BR6.4. NFR1.1, NFR9.1, NFR10.1.

### Step 13 — Environment and build configuration

- [x] Write the `Makefile` targets:
  - `check-sdk` fails if `../kandev` is missing or its HEAD differs from `.kandev-sdk-ref` (BR5.5); all Go targets depend on it;
  - `check-format`: `gofmt -l`, plus Prettier check;
  - `vet`;
  - `lint`: `golangci-lint` at a pinned version installed with `go run`; `tsc --noEmit`; ESLint;
  - `test`: `go test -race` over `./internal/... ./server/...`, plus `vitest run`;
  - `coverage`: `-race -coverprofile`, excluding only `server/main.go`, with the exclusion listed in a variable; fails below 80%;
  - `ui-build`: esbuild into `build/ui/bundle.js` as one ES module with no React;
  - `build`: `CGO_ENABLED=0` cross-compile of the 5 targets into the staging `build/server/`;
  - `package`: stage `manifest.yaml`, `server/` and `ui/`, then `plugin-pack`, writing `dist/nulab-backlog-<version>.tar.gz` and `dist/checksums.txt`;
  - `verify-package`: `plugin-package-verify` with the expected id and version, plus a contents check (the 5 executables, `manifest.yaml`, `ui/bundle.js`), plus a comparison of the package SHA-256 with `dist/checksums.txt` (BR5.2).
- [x] Add `.golangci.yml` (default linters plus `gosec`), `ui/eslint.config.js` and `ui/.prettierrc`.
- [x] Add `.github/workflows/ci.yml` as in `cicd-pipeline.md`: actions pinned by SHA, `permissions: contents: read`, Kandev checked out at `.kandev-sdk-ref` into `../kandev`, a `go mod tidy` diff check, and the `make` targets.
- [x] Run every `make` target locally and confirm that coverage is at least 80%.
- Rules: BR5.1, BR5.2, BR5.5, BR5.6. NFR4.1, NFR7.1, NFR8.1.

### Step 14 — Documentation and traceability

- [x] Add `README.md`: what the plugin does today, how to build (`make package`), how to install on Kandev, and how to connect with an API key.
- [x] Add `docs/manual-checks/TEMPLATE.md`. It starts with a reminder never to paste keys or URLs with a query, and has fields for date, Kandev version, plugin commit, space domain, steps, result, Connect duration, and "logs checked for key: yes/no" (BR5.3, NFR4.2).
- [x] Write `code-summary.md`, `source-manifest.json` and `traceability.json` for this unit.
- Stories: US7.2 (the manual check itself happens at the skeleton checkpoint).

## Revision 2 — Changes After the First Manual Check

Steps 1–14 are already built and stay as they are. The U1 design was revised after the first manual check (functional-design Q6–Q9, nfr-requirements Q4–Q6). Steps 15–22 modify the existing code in place; they follow the same Red → Green → Refactor order per layer.

### Step 15 — Data model and storage, Red: integration switch

- [x] `internal/connection/store_test.go` gains cases for the IntegrationSwitch record (state scope `workspace`, key `integration`, `schemaVersion: 1`):
  - no record means enabled (BR7.1);
  - saving `false` then `true` round-trips, with `changedAt` from the injected clock;
  - saving the switch never touches the `connection` record or the secret (BR7.4);
  - an undecodable stored value is an error, not "enabled" (fail closed, NFR3.9);
  - the `ConnectionView` carries `enabled` in every state.
- [x] Run the tests and record the failing output.

### Step 16 — Data model and storage, Green and Refactor

- [x] Implement `Store.LoadSwitch` / `Store.SaveSwitch` and add `Enabled` to the view. Each call keeps the 1-second store limit.
- [x] Refactor while green.
- Rules: BR7.1, BR7.4. NFR1.3, NFR3.9.

### Step 17 — Business logic, Red: switch service, disabled guard and the 14-second budget

- [x] `internal/connection/service_test.go` gains cases:
  - `SetEnabled` writes only the switch and returns the view; it logs `integration_switch_changed` with the workspace id, previous value, new value and duration, and nothing else (nfr-requirements Q5);
  - `Connect` while off returns `integration_disabled` with no Backlog call, no secret read and no write;
  - `Connect` reads the switch again just before storing; if it turned off during the Backlog call, nothing is written (BR7.3);
  - the new budget with an injected clock: pre-call steps 1 s, the `Myself` call 10 s and never past the 12-second deadline minus 2 s, store steps 2 s together, rollback 2 s on a fresh context; the total never exceeds 14 s (performance-design, replaces the 13 s / 8 s budget);
  - a fake Backlog that delays 2 s still answers well under 3 s; one that never answers returns `unreachable` at about 10 s.
- [x] Run the tests and record the failing output.

### Step 18 — Business logic, Green and Refactor

- [x] Implement `connection.Service.SetEnabled`, the switch check at the start of `Connect` and again in its store step, and the new deadlines (`Deadline` 12 s, Backlog call 10 s capped at deadline − 2 s, store 2 s, rollback 2 s).
- [x] Refactor while green.
- Rules: BR7.2–BR7.4. NFR1.2, NFR1.4, NFR3.9, NFR5.1.

### Step 19 — API / endpoint, Red and Green: `connection.set_enabled` and the guard

- [x] Red in `internal/plugin/actions_test.go`:
  - `connection.set_enabled` with `{enabled: true|false}` returns the view; a non-boolean or missing `enabled` gives 400 `validation` on field `enabled`;
  - the workspace comes only from the verified action context, never from the body;
  - while off, every action except `connection.get` and `connection.set_enabled` returns 409 `integration_disabled`; a failing or undecodable switch read gives 500 `internal` and the action does not run;
  - `connection.get` and `connection.set_enabled` always work while off.
- [x] Red in `internal/plugin/manifest_test.go`: `connection.set_enabled` is declared with `access: admin`, `scope: workspace` and `max_body_bytes: 8192`.
- [x] Record the failing output, then Green: add the action to `manifest.yaml`, the router, the one guard wrapper, and `integration_disabled` → 409 in the error mapping table. Refactor while green.
- Rules: BR7.2, BR7.3. NFR3.5, NFR3.9.

### Step 20 — Frontend behaviour, Red: logo, switch, home entry, `/backlog` page, spacing

- [x] `ui/src/brand/backlog-logo.test.tsx`: the logo renders only `svg` and `path` elements with fixed attributes and `aria-hidden`, and has no `script`, `foreignObject`, `href` or `on*` attribute (NFR3.10).
- [x] `ui/src/index.test.ts` gains cases: registration of the settings card with the logo `icon` and the switch `action`; `registerNavItem({ id: "backlog", path: "/backlog", section: "integrations", icon })` and `registerRoute("/backlog", …)`, both registered whether Backlog is on or off (BR7.6); at start and when the workspace list changes, the UI loads `connection.get` per workspace and calls `setIntegrationEnabled("nulab-backlog", workspaceId, enabled)` only after a successful load (BR7.5).
- [x] `ui/src/switch/switch.test.tsx`: the card action renders `IntegrationEnabledControl` for the routed workspace; persist calls `connection.set_enabled`, publishes with `setIntegrationEnabled` only on success, and rejects with a catalogue message on failure (403, `validation`, `internal`), so the host keeps the change unsaved.
- [x] `ui/src/page/backlog-page.test.tsx`: every BacklogPage state from `functional-spec` WF7 — loading, load failed with Retry, no active workspace, Off, Not connected, Connected ("Connected as <name> @ <host>"), Incomplete — each with a link to the Backlog settings of that workspace; `data-testid` on interactive elements; `axe-core` finds no violations; no literal text outside the catalogue.
- [x] `ui/src/settings/settings.test.tsx` gains cases: the screen shows "Backlog is turned off for this workspace" and blocks Connect when `enabled` is false; an `integration_disabled` reply shows the same message; the layout uses one vertical stack with one consistent gap and a smaller label-to-input gap, built from the host UI kit only (BR6.5).
- [x] Run the tests and record the failing output.

### Step 21 — Frontend behaviour, Green and Refactor

- [x] Implement:
  - `ui/src/brand/backlog-logo.tsx` (inline SVG from Nulab's official brand assets, exported through one `PLUGIN_ICON` constant so a fallback to a host built-in icon is a one-line change);
  - `ui/src/switch/integration-switch.tsx` (the card action);
  - `ui/src/page/BacklogPage.tsx` and its pure state function;
  - the new registrations and workspace-list sync in `ui/src/index.ts`;
  - the spacing retouch and the Off state in `ui/src/settings/SettingsScreen.tsx` and `state.ts`;
  - the new message keys in `ui/src/messages/en.ts`.
- [x] Refactor while green.
- Rules: BR6.5, BR7.5–BR7.8. NFR3.10, NFR9.1, NFR10.1.

### Step 22 — Build checks, documentation and traceability

- [x] Extend the `verify-package` bundle check (`internal/pkgverify`) to fail on any asset URL from a Nulab or Backlog domain inside `ui/bundle.js` (NFR3.10), with a test in `pkgverify_test.go`.
- [x] Add `docs/brand/backlog-logo.md` with the logo's source URL and terms, and the reminder that the user confirms Nulab's brand guidelines before release.
- [x] Update `README.md` (the switch, the home entry and the `/backlog` page) and `docs/manual-checks/TEMPLATE.md` (steps for the switch, the home entry and the page).
- [x] Run every `make` target locally and confirm coverage is still at least 80%.
- [x] Update `code-summary.md`, `source-manifest.json` and `traceability.json` for the new rules and NFRs.
- Stories: US1.1, US7.1, US7.2 (scope added to U1 by the change request, functional-design Q6–Q9).

## Story-to-Step Map

| Story | Steps |
|-------|-------|
| US7.1 Plugin skeleton installs into Kandev | 1, 2, 9, 10, 13, 19–22 |
| US1.2 Validate the space address | 3, 4, 7, 8 |
| US1.1 Connect with an API key | 3–12, 15–18 |
| US7.2 First Backlog call from a real server | 7, 8, 14 (template), 22, then the manual check |

## Sources

- [desc] Initial description: Kandev plugin for Nulab Backlog.
- [scope] Workflow-selected scope: `feature`.
- U1 `functional-spec.md`, `rules.md`, `entities.md`; U1 NFR requirements and design; U1 `infrastructure-specification.md`, `cicd-pipeline.md`; `contract-summary.md`; `unit-of-work.md`; `requirements.md`; `team-practices.md`.

## Assumptions & Open Questions

- [assumption] The GitHub repository will be `github.com/khuongdo/kandev-plugin-nulab-backlog`, the GitHub owner the user named.
- [assumption] The plugin id is `nulab-backlog`.
