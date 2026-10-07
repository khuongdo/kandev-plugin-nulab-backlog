# Code Summary — walking-skeleton (U1)

All 14 plan steps were carried out in order, under the TDD Testing Contract (`contract_sha256 sha256:376593af89ee0d9cddcd078826773d6d5ca3215ec737b9af952712bee2395bcc`). Every `make` target passes from a clean tree: `check-format vet lint test coverage build package verify-package`. Go line coverage is **94.6%**; the floor is 80%. These results were re-run after the follow-up below, at the pinned Kandev v0.96.0.

**Follow-up (user decision).** `../kandev` is now checked out at the v0.96.0 tag, and `.kandev-sdk-ref` is `f099a46dc7aab16f6ff5806cd29b2b480296303f`; this run changed neither. Kandev's `cmd/plugin-package-verify` does not exist at that commit. `make verify-package` now uses a verifier in this repo (`internal/pkgverify` plus `cmd/verifypkg`), built test-first, and keeps every BR5.2 check.

The coordinator corrected the GitHub owner during the run. The module path is `github.com/khuongdo/kandev-plugin-nulab-backlog`, and `manifest.yaml` (`repo_url`, `author`) and `LICENSE` use `khuongdo`. No file contains `khuongdo-nicosys`.

## Files

Application code is at the repository root. `source-manifest.json` lists every path.

| Area | Files |
|------|-------|
| Build and SDK pin | `go.mod`, `go.sum`, `.kandev-sdk-ref`, `Makefile`, `.golangci.yml`, `.nvmrc`, `.gitignore` (adds `dist/`, `build/`, `node_modules/`, `coverage.out`) |
| Manifest | `manifest.yaml` (id `nulab-backlog`, api_version 2, version 0.0.1, `min_kandev_version: "0.96.0"`, 5 executables, `state` + `secrets`, 2 actions with `max_body_bytes: 8192`, `ui.bundle: /ui/bundle.js`) |
| Entry point | `server/main.go`, which only calls `pluginsdk.Serve(plugin.NewRuntime())` |
| KandevAdapter | `internal/plugin/runtime.go` (router, one code→status table, `ActionError` JSON, `requestId`, panic recovery, `hostStores` adapter, `plugin_started`), plus `actions_test.go` and `manifest_test.go` |
| Connection | `internal/connection/address.go`, `apikey.go`, `store.go` (BR2.8 writes, BR2.11 view), `service.go` (Connect, lock, `Classify`, outcome logs), plus tests and `fakes_test.go` |
| BacklogGateway | `internal/backlog/types.go` (C1 types), `client.go` (`Myself`, https only, no redirects, 1 MiB limit, Kind mapping, 429 wait, `backlog_call` log), tests, and `testdata/*.json` |
| Redaction | `internal/redact/redact.go` (secret set on the context, URL query masking, `slog.Handler` wrapper, logger on the context), `redact_test.go` |
| Package verifier | `internal/pkgverify/pkgverify.go` (`Verify`, `Run`), `internal/pkgverify/pkgverify_test.go`, `cmd/verifypkg/main.go` (only calls `pkgverify.Run`) |
| Test helper | `internal/testutil/testutil.go` (per-run `test-api-key-` + 32 hex keys, 8-character-window leak assertion), `testutil_test.go` |
| UI | `ui/src/index.ts`, `ui/src/settings/SettingsScreen.tsx`, `ui/src/settings/state.ts`, `ui/src/messages/en.ts`, `ui/src/jsx.d.ts`, tests `ui/src/settings/settings.test.tsx` and `ui/src/index.test.ts`, config `package.json`, `package-lock.json`, `tsconfig.json`, `vitest.config.ts`, `eslint.config.js`, `.prettierrc`, `.prettierignore` |
| CI | `.github/workflows/ci.yml`: job `checks`; actions pinned by full SHA (checkout v7.0.1, setup-go v7.0.0, setup-node v7.0.0, upload-artifact v7.0.1); `permissions: contents: read`; Kandev checked out at `.kandev-sdk-ref` into `kandev/`; `go mod tidy` diff check; the standard `make` targets; `dist/` uploaded |
| Docs | `README.md`, `docs/manual-checks/TEMPLATE.md`, `LICENSE` (MIT) |

## Key Decisions

- **The store interfaces use the same method names as `pluginsdk.Host`.** `connection.SecretStore` (`GetSecret`, `SetSecret`, `DeleteSecret`) and `connection.StateStore` (`GetState`, `SetState`) match the Host signatures. The adapter in `internal/plugin` (`hostStores`) is therefore a small pass-through. It reads `Runtime.Host()` on every call, because Kandev injects the Host from a background goroutine after `NewRuntime` returns. A missing Host gives `internal`.
- **The logger travels on the context.** `redact.WithLogger` and `redact.Logger` carry it. `HandleAction` builds a logger with `workspaceId` and `requestId` (NFR11.1), so the gateway (`backlog_call`), the store (`connection_inconsistent`) and the service (`connect_*`) all write lines that can be matched to a request. Every line goes through `redact.NewHandler`.
- **Domain error codes are mapped once, in `connection.Classify`.** The C5 code is also needed for the `connect_failed` log field. `internal/plugin` owns the only code→HTTP-status table (`statusFor`) and the `ActionError` JSON. No `pluginsdk` error type is used: the C5 codes differ from the SDK's `ActionErrorCode` set.
- **A timeout is `Unreachable` but still matches `errors.Is(err, context.DeadlineExceeded)`.** `backlog.Error.Is` makes this possible. The caller's `context.Canceled` is returned unchanged. A `*url.Error` never escapes: transport errors become `Error{Kind: Unreachable, Class: timeout|dns|tls|connection}`.
- **Redirects return the 3xx response itself.** `CheckRedirect` returns `http.ErrUseLastResponse`, and the 3xx is mapped to `Unreachable` with class `redirect`. The test proves the other host receives 0 requests.
- **A record with an unknown `schemaVersion` gives `internal`.** It is never overwritten (contract versioning rule).
- **Requests rejected before the service run are logged.** A malformed body or an unknown action logs `action_rejected` (WARN), so its `requestId` can be found.
- **Retry on the "Could not reach Backlog" screen resends the last submitted input.** The input is held only in memory, in a ref. The key field is still cleared after every Connect (BR6.3).
- **The UI bundle contains no React.** The JSX factory `h` is `host.jsx`, and host components come from `host.ui`. `@kandev/plugin-sdk` types are resolved from `../kandev/apps/packages/plugin-sdk` through a tsconfig path, using type-only imports. `make ui-build` fails if React is found in the bundle. The bundle is 9 kB.
- **Packaging runs from the pinned checkout; verification runs in this repo.** The Makefile uses `go -C ../kandev/apps/backend run ./cmd/plugin-pack`, which keeps the packer's dependencies out of this module's `go.sum`, so `go mod tidy` leaves no diff. `make verify-package` runs `go run ./cmd/verifypkg`.
- **The package verifier is reimplemented, not imported.** At v0.96.0, Kandev's checksum logic lives in `apps/backend/internal/plugins/pkgtar`. Go's `internal` rule stops this module from importing it, and there is no public equivalent under `apps/backend/pkg`. `internal/pkgverify` mirrors pkgtar's rules: the `<sha256>  <path>` format, no unlisted files, no listed file missing, and no path escaping the package. It adds the BR5.2 checks: the package SHA-256 against its line in `dist/checksums.txt`; the 7 required files; the manifest `id` and `version`; and `runtime.executables` equal to exactly the 5 entries. The checks live under `./internal/...`, so they count toward coverage. The `cmd` is a one-line wrapper.

## Red evidence

Each Red step ran before any production code for that layer existed. The key lines are trimmed below.

**Step 2, runner bootstrap.** Smoke tests ran green: `go test -race ./internal/redact/...` → `ok`, and `npx vitest run src/settings` → `1 passed`. Both smoke tests were deleted at the first Red step of their layer.

**Step 3, data model.** `go test -race ./internal/connection/... ./internal/backlog/...`

```
internal/backlog/types_test.go:24:11: undefined: User
internal/backlog/types_test.go:40:16: undefined: parseUser
internal/connection/address_test.go:50:17: undefined: ParseSpaceAddress
internal/connection/apikey_test.go:32:16: undefined: ValidateAPIKey
FAIL	.../internal/connection [build failed]
```

**Step 5, repository / data access.** `go test -race ./internal/connection/...`

```
internal/connection/fakes_test.go:154:60: undefined: Store
internal/connection/fakes_test.go:155:7: undefined: NewStore
internal/connection/store_test.go:61:19: undefined: View
internal/connection/store_test.go:94:26: undefined: ErrStore
```

**Step 7, business logic.** `go test -race ./internal/redact/... ./internal/backlog/... ./internal/connection/...`

```
internal/redact/redact_test.go:19:9: undefined: WithSecrets
internal/redact/redact_test.go:45:18: undefined: NewHandler
FAIL	.../internal/redact [build failed]
internal/connection/service_test.go:62:11: undefined: Service
internal/connection/service_test.go:123:180: undefined: CodeUnreachable
internal/backlog/client_test.go:29:54: undefined: Client
```

**Step 9, API / endpoint.** `go test -race ./internal/plugin/...`

```
internal/plugin/actions_test.go:112:8: undefined: newRuntime
internal/plugin/actions_test.go:125:20: r.rt.HandleAction undefined (type *Runtime has no field or method HandleAction)
internal/plugin/actions_test.go:229:24: undefined: statusFor
```

**Step 11, frontend.** `npx vitest run` (in `ui/`)

```
FAIL  src/index.test.ts
Error: Failed to resolve import "./index" from "src/index.test.ts". Does the file exist?
FAIL  src/settings/settings.test.tsx
Error: Failed to resolve import "./SettingsScreen" from "src/settings/settings.test.tsx". Does the file exist?
```

**Step 13, defect found by `make verify-package`.** The fix was test-first. A key-pattern assertion was added to `TestManifestActionsAndAccess` and failed first:

```
--- FAIL: TestManifestActionsAndAccess
    Error: Expect "connection.connectApiKey" to match "^[a-z0-9][a-z0-9._-]*$"
```

The key was then renamed (see Deviations).

**Follow-up, package verifier.** `go test -race ./internal/pkgverify/...` (Go 1.26.8, Kandev at v0.96.0):

```
internal/pkgverify/pkgverify_test.go:31:12: undefined: Expect
internal/pkgverify/pkgverify_test.go:116:21: undefined: Verify
internal/pkgverify/pkgverify_test.go:158:11: undefined: Verify
FAIL	.../internal/pkgverify [build failed]
```

Green came at the first implementation: 93.0% package coverage. The refactor step renamed a helper (`sortedValues` became `values`).

**Green-phase failures (fixed while green was being reached):**

- `TestConnectCancelledWritesNothingAndReturnsCancellation`. The fake gateway reported cancellation as a timeout; it now mirrors the real client.
- `TestErrorCodesMapToStatusAndActionError/malformed_body`. No log line carried the `requestId`; `action_rejected` was added.

## Test and Coverage Results

**Attempt 2 (2026-10-06).** The stage was restarted only because the previous attempt's review could not be recorded: build outputs were regenerated during that review. No code was regenerated or rewritten. The module path is `github.com/khuongdo/kandev-plugin-nulab-backlog`. The Kandev SDK is pinned at v0.96.0 (`f099a46dc7aab16f6ff5806cd29b2b480296303f`, in `.kandev-sdk-ref` and matching `../kandev` HEAD). A fresh run of `make check-format vet lint test coverage build package verify-package` (Go 1.26.8, `GOTOOLCHAIN=local`) passed every target. Go tests ran with `-race`, and all 22 Vitest tests passed. Total Go coverage is 94.6% against the 80% floor, with only `server/main.go` excluded. `make package` wrote `dist/nulab-backlog-0.0.1.tar.gz` and `checksums.txt`, and `verifypkg` reported `OK (nulab-backlog@0.0.1)`. The Red evidence above and the table below are from attempt 1 and still apply.

Re-run at Kandev v0.96.0 (`f099a46…`) with Go 1.26.8 and `GOTOOLCHAIN=local`, from `make clean`:

| Target | Result |
|--------|--------|
| `make check-sdk` | pass. A missing `../kandev` or a different HEAD stops the build and names the expected commit (both cases checked by hand) |
| `make check-format` | pass (`gofmt -l` empty; Prettier clean) |
| `make vet` | pass |
| `make lint` | pass. golangci-lint v2.14.0 (standard linters plus gosec): `0 issues`. `tsc --noEmit` (strict) and ESLint are clean |
| `make test` | pass. Go with `-race`: 82 top-level tests and 120 subtests, 0 failures. Vitest: 22 tests passed |
| `make coverage` | pass, **94.6%** with only `server/main.go` excluded. Per package: backlog 96.8%, connection 97.7%, pkgverify 93.0%, plugin 89.6%, redact 97.4%, testutil 78.9% |
| `make build` | pass. 5 executables built with `CGO_ENABLED=0` |
| `make package` | pass. `plugin-pack` from v0.96.0 writes `dist/nulab-backlog-0.0.1.tar.gz` (`manifest.yaml`, 5 executables, `ui/bundle.js`, in-archive `checksums.txt`) and `dist/checksums.txt` |
| `make verify-package` | pass: `verifypkg: OK dist/nulab-backlog-0.0.1.tar.gz (nulab-backlog@0.0.1)` |
| `go mod tidy` | no diff in `go.mod` or `go.sum` |
| AC7.1.3 by hand, file changed and the archive rebuilt with a fresh dist line | `verifypkg: checksum mismatch: ui/bundle.js` (exit 1) |
| AC7.1.3 by hand, archive bytes changed | `verifypkg: package checksum mismatch: nulab-backlog-0.0.1.tar.gz` (exit 1) |

The `//nolint` exceptions:

- `gosec` G304 in `types_test.go` (fixture names are test constants).
- `gosec` G117 in `store.go` (the marshalled `apiKey` goes only to the secret store).
- `gosec` G101 in `testutil.go` (the fake key prefix).
- `gosec` G304, twice in `pkgverify.go` (the paths are the operator's own build output).
- `errcheck` was fixed in code, not suppressed.

## Deviations from the Plan

1. **The action key `connection.connectApiKey` was renamed to `connection.connect_api_key`.** Kandev rejects action keys that do not match `^[a-z0-9][a-z0-9._-]*$`. Packaging accepted the key, and Kandev's `plugin-package-verify` refused it; that run used the earlier pin `abc7a85…`. The same rule exists at v0.96.0 (`apps/backend/internal/plugins/manifest/validate.go`, `actionKeyPattern`), so Kandev would also refuse the old key at install. The new key is used in the manifest, `internal/plugin`, the UI and the tests. **Upstream C5 in `contract-summary.md` needs amending.** Every later camelCase action name has the same problem: `connection.startOAuth`, `setProjects`, `listProjects`, `setPollInterval`, `setGitCredential`, `issues.createTask`, and others.
2. **The SDK pin was first set to the wrong commit, then resolved by the user.** Code generation first pinned `../kandev` HEAD `abc7a85…` (`v0.96.0-86`). The user then moved `../kandev` to the v0.96.0 tag, and `.kandev-sdk-ref` is now `f099a46dc7aab16f6ff5806cd29b2b480296303f`, which matches `min_kandev_version: "0.96.0"`.
10. **`make verify-package` no longer uses Kandev's `plugin-package-verify`.** That command is absent at the pinned v0.96.0. The in-repo `internal/pkgverify` (via `cmd/verifypkg`) performs every BR5.2 check, as the Key Decisions section describes. The earlier shell checks (`sha256sum -c`, the `tar -t` contents list, the executables count) are folded into the Go verifier, so `verify-package` no longer needs `sha256sum` or `tar`.
3. **The rollback limit is 2 s.** This follows the plan and `performance-design.md`. `functional-spec.md` WF3 step 7.4, `rules.md` BR2.8 and NFR5.4 still say 5 s, and the upstream documents disagree with each other.
4. **The leak scans use 8-character windows of the random key part.** This follows `security-design.md` NFR4.1 and the brief. NFR3.2's verification text says 4-character substrings, but 4-hex windows collide by chance with the random 16-hex `requestId`, which would make the tests flaky.
5. **Performance checks use injected short deadlines instead of real delays.** The team practice forbids real sleeps.
   - NFR1.2 ("fake delays 2 s, p95 < 3 s") is covered by the sub-deadline tests.
   - NFR1.1 ("timing over 100 calls") has no automated timing test. By construction, `connection.get` makes two store reads and no network call.
6. **There is an added package, `internal/testutil`.** It holds the shared test-key helper that the plan's mocking rules ask for. It is inside the coverage scope and has its own tests.
7. **The UI has extra devDependencies.** `react`, `react-dom`, `@types/react` and `@types/react-dom` build the fake `host` in tests only. The bundle does not include them. The tool versions are TypeScript ~6.0.3 (typescript-eslint supports TypeScript below 6.1), ESLint 10, Vitest 5, jsdom 30, esbuild 0.28 and Prettier 3.9.
8. **The local toolchain differs from CI.** No Go was installed, so Go 1.26.8 was downloaded into the session scratchpad and its SHA-256 was verified against go.dev. Nothing was installed system-wide. Local Node is v25.2.1, which triggers an engines warning from Vitest. `.nvmrc` and CI use Node 22.
9. **Smoke tests were removed.** The Step 2 smoke tests (`internal/redact/smoke_test.go`, `ui/src/settings/smoke.test.ts`) were created and then deleted when the real tests replaced them. They are not in the final tree.

## Revision 2 — Steps 15–22 (after the first manual check)

Steps 15–22 changed the existing code in place, under the same TDD Testing Contract (`sha256:376593af89ee0d9cddcd078826773d6d5ca3215ec737b9af952712bee2395bcc`). Each layer went Red → Green → Refactor. A fresh run of `make check-format vet lint test coverage build package verify-package` passed every target from `make clean` (Go 1.26.8, `GOTOOLCHAIN=local`, Kandev v0.96.0 `f099a46…`). Go line coverage is **94.9%** against the 80% floor, with only `server/main.go` excluded. `go mod tidy` leaves no diff, and no dependency was added.

### Files

| Change | Files |
|--------|-------|
| Created | `ui/src/brand/backlog-logo.tsx`, `ui/src/brand/backlog-logo.test.tsx`, `ui/src/switch/integration-switch.tsx`, `ui/src/switch/switch.test.tsx`, `ui/src/page/BacklogPage.tsx`, `ui/src/page/backlog-page.test.tsx`, `ui/src/testing/harness.ts` (shared test helpers, not bundled), `docs/brand/backlog-logo.md` |
| Modified | `internal/connection/store.go`, `store_test.go`, `service.go`, `service_test.go`, `fakes_test.go`; `internal/plugin/runtime.go`, `actions_test.go`, `manifest_test.go`; `internal/pkgverify/pkgverify.go`, `pkgverify_test.go`; `manifest.yaml`; `ui/src/index.ts`, `index.test.ts`, `messages/en.ts`, `settings/SettingsScreen.tsx`, `settings/state.ts`, `settings/settings.test.tsx`; `README.md`; `docs/manual-checks/TEMPLATE.md` |
| Deleted | none |

### Key decisions

- **IntegrationSwitch storage.** `Store.LoadSwitch` and `Store.SaveSwitch` use state scope `workspace`, key `integration`, `{schemaVersion: 1, enabled, changedAt}`. `changedAt` comes from the store's injected clock. Each call keeps the 1-second store limit. `enabled` is decoded through a `*bool`, so a missing, non-boolean or unknown-schema value is `ErrStore` and never "on" (NFR3.9). `SaveSwitch` writes only that record (BR7.4). `Store.Load` reads the switch first and puts `enabled` on every view.
- **`Service.SetEnabled`** reads the previous value, writes the switch, logs exactly one `integration_switch_changed` (INFO; `previousEnabled`, `enabled`, `durationMs`; `workspaceId` and `requestId` come from the request logger), then returns `Store.Load`. If the previous value cannot be read, the write still happens and `previousEnabled` is logged as `"unknown"`. This lets an admin repair a broken switch that the fail-closed guard would otherwise lock forever.
- **Connect budget (NFR1.4).** `Deadline` is 12 s, `PreCallTimeout` 1 s (switch read, validation, lock), and `BacklogTimeout` 10 s, capped at the time left minus `StoreTimeout`. `StoreTimeout` is 2 s (second switch read, then `Save`), and `RollbackTimeout` stays 2 s on a fresh context. A store step that runs out of time is now `internal`, not `unreachable`: `storeErr` and `rollback` return only `context.Canceled` unchanged, and wrap a deadline with `ErrStore`. `Classify` checks `ErrStore` before the deadline case. The wrapped error still matches `errors.Is(err, context.DeadlineExceeded)`.
- **The budget tests use `testing/synctest`** (Go standard library). Its bubble gives a virtual clock for `time.Now`, timers and `context.WithTimeout`, so the real 1/10/2/2-second limits are asserted exactly without any real waiting.
- **One guard in `internal/plugin`.** A `handlers` map routes actions. `guarded(key)` is true for every action except `connection.get` and `connection.set_enabled`, so later units' actions are guarded by default. The guard calls `connection.Service.RequireEnabled`, which fails closed. An unknown action is still 404 and is checked before the guard. The guard runs inside the action's 12-second deadline, so the 14-second total also covers its read; a test proves this.
- **`connection.set_enabled`** decodes `{enabled *bool}`. A missing, `null`, non-boolean or malformed body is `validation` on field `enabled`. The body has no workspace field. `integration_disabled` maps to 409 in `statusFor`. The manifest declares it `scope: workspace`, `access: admin`, `max_body_bytes: 8192`.
- **Logo.** `ui/src/brand/backlog-logo.tsx` renders the official Backlog product icon from Nulab's media kit (`Nulab-all-logos.zip`, `Icons/Backlog/Backlog-Product-Icon-Color.svg`). The two paths and their colours are copied unchanged. The SVG is `aria-hidden`, not focusable, and passes through only the host's `className`. `PLUGIN_ICON` is the one place the icon is chosen. Source, hash and terms are in `docs/brand/backlog-logo.md`.
- **Card switch.** `createIntegrationSwitch` loads `connection.get` for the routed workspace. After a successful load it publishes the value, then renders `host.ui.IntegrationEnabledControl` with `{id: "nulab-backlog", enabled, persist, name}`. Kandev v0.96.0's drafted control takes exactly these props and prefixes the id with the plugin id. `persist` calls `connection.set_enabled` and publishes with `setIntegrationEnabled` only after success. On failure it throws a catalogue message (`switchForbidden` for 403, `errorInput` for `validation`, `switchFailed` otherwise), so the host keeps the change unsaved. The host control takes no `data-testid`, so a wrapping `span` carries `backlog-integration-switch`.
- **Registrations and workspace sync (`ui/src/index.ts`).** The code registers the card (`icon`, `action`), `registerNavItem({id: "backlog", label, path: "/backlog", section: "integrations", icon})` and `registerRoute("/backlog", …, {topbar: {title, icon}})`, none of them conditional. It then calls `connection.get` for each id from `host.context.getWorkspaceIds()`, and again on every `subscribeWorkspaces` change, and publishes each value only after a successful load. A failed load is logged with `console.warn` (workspace id only) and left unpublished. `destroy()` unsubscribes.
- **Backlog page.** `pageState` is a pure function over `{workspaceId, load, view}`. Its order is no workspace → loading/failed → off → connected/incomplete/not connected, so Off wins over the connection state. The settings link is `/settings/workspaces/<id>/integrations/nulab-backlog` (Kandev's `workspaceSettingsHref` + plugin id), opened with `host.navigate`. The page follows the active workspace via `subscribeActiveWorkspace`.
- **Settings screen.** When `enabled` is false, the screen shows `integrationOff` and no form, and keeps the status line. An `integration_disabled` reply sets `view.enabled = false` and announces the same message. Layout (BR6.5): the screen, form and notice use `flex flex-col gap-4`, and each field uses `flex flex-col gap-2`. These are host utility classes; there is no inline style and no stylesheet.
- **Bundle check.** `pkgverify.Verify` fails when `ui/bundle.js` matches a URL (`http(s):` or protocol-relative `//`) on `nulab`, `nulab-inc`, `backlog` or `backlogtool` `.com`/`.jp`. Bare host names, such as the input placeholder `myteam.backlog.com`, are allowed. The bundle is 17.5 kB and has no React.

### Red evidence

**Step 15, data model and storage.** `go test -race ./internal/connection/...`

```
internal/connection/store_test.go:297:65: newTestStore(newFakeSecrets(), newFakeState()).LoadSwitch undefined (type *Store has no field or method LoadSwitch)
internal/connection/store_test.go:307:29: store.SaveSwitch undefined (type *Store has no field or method SaveSwitch)
internal/connection/store_test.go:393:36: view.Enabled undefined (type View has no field or method Enabled)
FAIL	github.com/khuongdo/kandev-plugin-nulab-backlog/internal/connection [build failed]
```

**Step 17, business logic.** `go test -race ./internal/connection/...`

```
internal/connection/service_test.go:253:36: svc.PreCallTimeout undefined (type *Service has no field or method PreCallTimeout)
internal/connection/service_test.go:255:38: svc.StoreTimeout undefined (type *Service has no field or method StoreTimeout)
internal/connection/service_test.go:403:21: h.svc.SetEnabled undefined (type *Service has no field or method SetEnabled)
internal/connection/service_test.go:467:26: undefined: ErrIntegrationDisabled
FAIL	github.com/khuongdo/kandev-plugin-nulab-backlog/internal/connection [build failed]
```

The first Green run then failed one assertion (`Expected nil, but got: "<nil>"`): the redaction handler turns an `any` attribute into a string. The test and the code were changed to log `"unknown"`.

**Step 19, API / endpoint (actions and manifest; same package).** `go test -race ./internal/plugin/...`

```
internal/plugin/actions_test.go:325:26: undefined: actionSetEnabled
internal/plugin/actions_test.go:404:19: undefined: guarded
FAIL	github.com/khuongdo/kandev-plugin-nulab-backlog/internal/plugin [build failed]
```

**Step 19 follow-up, guard inside the budget.** `go test -race ./internal/plugin/... -run TestGuardRunsInsideTheConnectDeadline`

```
Error: Not equal: expected: 10s  actual: 10.9s
FAIL	github.com/khuongdo/kandev-plugin-nulab-backlog/internal/plugin
```

**Step 20, frontend behaviour.** `cd ui && npx vitest run`

```
FAIL src/brand/backlog-logo.test.tsx  Error: Failed to resolve import "./backlog-logo"
FAIL src/page/backlog-page.test.tsx   Error: Failed to resolve import "./BacklogPage"
FAIL src/switch/switch.test.tsx       Error: Failed to resolve import "./integration-switch"
FAIL src/index.test.ts > registers the settings card with the logo icon and the switch action
     AssertionError: expected "vi.fn()" to be called with arguments: [ ObjectContaining{…} ]
FAIL src/index.test.ts > registers the entry and the route even when Backlog is off everywhere
     AssertionError: expected "vi.fn()" to be called 1 times, but got 0 times
FAIL src/settings/settings.test.tsx > shows the Off message and no Connect form while Backlog is off ...
     TypeError: Cannot read properties of null (reading 'textContent')
FAIL src/settings/settings.test.tsx > lays out one vertical stack with one gap ... (BR6.5)
     AssertionError: expected '' to be 'flex flex-col gap-4'
Test Files  5 failed (5)   Tests  10 failed | 21 passed (31)
```

The first Green run failed two tests because of a bug in the test helpers: a default parameter replaced an explicit `undefined` workspace. The helpers now take `null` for "no workspace".

**Step 22, bundle check.** `go test -race ./internal/pkgverify/...`

```
--- FAIL: TestVerifyRejectsABundleLoadingANulabOrBacklogAsset
    Error: An error is expected but got nil.   (x5, one per URL form)
FAIL	github.com/khuongdo/kandev-plugin-nulab-backlog/internal/pkgverify
```

### Test and coverage results (Revision 2)

| Check | Result |
|-------|--------|
| `make check-format` | gofmt clean; Prettier: all files formatted |
| `make vet` | clean |
| `make lint` | golangci-lint v2.14.0 (with gosec): 0 issues; `tsc --noEmit` strict: clean; ESLint: clean |
| `make test` | Go `-race`: all packages `ok` (105 top-level tests, 149 subtests); Vitest: 5 files, 60 tests passed (was 22) |
| `make coverage` | **94.9%** (floor 80%, excluded `server/main.go`). `internal/connection` 97.5%, `internal/plugin` 91.4%, `internal/pkgverify` 93.2%, `internal/backlog` 96.8%, `internal/redact` 97.4% |
| `make build` / `package` | 5 executables; `dist/nulab-backlog-0.0.1.tar.gz` and `checksums.txt` |
| `make verify-package` | `verifypkg: OK dist/nulab-backlog-0.0.1.tar.gz (nulab-backlog@0.0.1)` |

New test cases, by plan step:

- Step 15: 7 store tests, among them a table of 3 undecodable values and a 3 × 2 table of enabled-per-state.
- Step 17: 5 budget subtests, run under synctest:
  - a stalled pre-call is internal at 1 s;
  - a Backlog that never answers is unreachable at 10 s;
  - a Backlog that answers in 2 s connects in under 3 s;
  - the call is capped at 9.1 s after 0.9 s of pre-call;
  - the worst case is 13.9 s, which is at most 14 s.

  Step 17 also added 7 tests for `SetEnabled` and the guard.
- Step 19: 8 action tests (one of them a table of 6 invalid bodies), plus the manifest access table.
- Step 20: 5 logo tests, 6 registration tests, 7 switch tests, 7 `pageState` rows, 11 page tests and 4 settings tests.
- Step 22: 2 pkgverify tests, one of them a table of 5 URL forms.

### Deviations (Revision 2)

1. **NFR1.3 and the returned view.** `performance-design.md` describes `connection.set_enabled` as "one state read and one state write, with no network call and no secret access". WF6 step 4 and BR7.2 require it to return the full ConnectionView, which needs the record and the secret to work out `state` and `hasApiKey`. `SetEnabled` writes only the switch, as the plan says ("writes only the switch and returns the view"), and the requirement NFR1.3 ("no Backlog call; one state write") holds. It then reads the view through `Store.Load`, which reads the secret only to compare epoch and host and never uses the key. `performance-design.md` should be amended to say "one state write; the view read of WF2".
2. **The guard and Connect's own first switch read.** The plan asks for a switch check at the start of `Connect` and the NFR3.9 design asks for a plugin-level guard, so a guarded Connect reads the switch twice before the Backlog call. Both are cheap. The plugin's action deadline starts before the guard, so the 14-second total holds; `TestGuardRunsInsideTheConnectDeadline` covers this.
3. **"Injected clock" is the `testing/synctest` virtual clock** for the deadline tests. The store's injected `Now` is still used for `changedAt`. The earlier tests that set short real deadlines (`TestConnectBacklogCallUsesTheSubDeadline`, `TestConnectOverallDeadlineBoundsTheStoreWrites`) were replaced by the synctest budget tests. This also partly closes Revision 1 Deviation 5: NFR1.2 now has a 2-second fake-delay test.
4. **Settings links appear in the four view states only.** The plan's test line says "each with a link". The `functional-spec` P1 table puts the link in Off, Not connected, Connected and Incomplete. Loading and No workspace have no workspace or view to link to, and Load failed shows Retry. The tests follow the table.
5. **The BacklogPage file name.** The plan names it `BacklogPage.tsx`, and it was kept as written. Other new files are kebab-case.
6. **Two message keys were reused.** `errorInput` is the switch's `validation` message, and `incomplete` is the page's Incomplete text.

## Review fixes (R-01..R-03)

The code-generation architecture review (`reviews/review-02.md`) found three defects; the user chose to fix exactly these. Each went Red → Green → Refactor under the same TDD Testing Contract. Existing code was changed in place; the plan and `unit-test-instructions.md` are unchanged.

### R-01 (Major): `connection.set_enabled` violated NFR1.3 and could report a saved switch as failed

**Red.** `go test -race ./internal/connection/ ./internal/plugin/`

```
--- FAIL: TestSetEnabledIsSuccessOnceTheSwitchIsWritten
    Error: Received unexpected error: read secret: connection store failure: injected store failure
--- FAIL: TestSetEnabledDoesOneStateReadOneStateWriteAndNoSecretAccess
    expected: []string{"state:integration"}
    actual  : []string{"state:integration", "state:integration", "state:connection", "secret:backlog.connection.ws-1"}
--- FAIL: TestSetEnabledSucceedsWhenTheWriteSucceeds
    expected: 200  actual: 500
```

**Change.** `Service.SetEnabled` no longer calls `Store.Load` after the write. It does one state read (the previous switch value, needed for the `integration_switch_changed` log), one state write (the switch) and no secret access, and returns `nil` error once the write succeeded. It returns a new `connection.SwitchView{enabled}` instead of a `View`. The action `handler` type in `internal/plugin` now returns `any`, so `set_enabled` replies `{"enabled": <bool>}`. The UI ignored the `set_enabled` reply already (the switch keeps its own value and the settings screen and page use `connection.get`), so only the TypeScript generic in `integration-switch.tsx` changed. This supersedes Revision 2 Deviation 1 and the "then returns `Store.Load`" part of the `Service.SetEnabled` key decision.

**Tests.** `internal/plugin/actions_test.go`: `TestSetEnabledDoesOneStateReadOneStateWriteAndNoSecretAccess` counts host store calls through the real action handler (reads exactly `["state:integration"]`, one write, secrets unchanged, reply exactly `{enabled}`); it replaces `TestSetEnabledReturnsTheView`. `TestSetEnabledSucceedsWhenTheWriteSucceeds` makes the secret store fail and still gets 200 (while `connection.get` reports `internal`). `internal/connection/service_test.go`: `TestSetEnabledIsSuccessOnceTheSwitchIsWritten` (no error, zero secret reads, switch stored) and `TestSetEnabledWritesOnlyTheSwitchAndReturnsItsValue` (renamed; it checks the kept connection through `Get`). The fake host gained `failSecretRead`.

**Spec ambiguity (not resolved in code).** `functional-spec.md` WF6 step 4 and `entities.md` (ConnectionView: "what … connection.set_enabled return[s]") say the reply is the full ConnectionView, but `state`, `connected` and `hasApiKey` need the record and the secret (BR2.11), which `performance-design.md` NFR1.3 forbids, and even the record alone would be a second state read. The reply therefore carries only the switch value, which is the one thing the write knows. WF6, `entities.md` and contract C5 should be amended to say that `set_enabled` returns `{enabled}`.

### R-02 (Major): the settings screen did not follow the card switch

**Red.** `cd ui && npx vitest run src/switch`

```
FAIL src/switch/switch.test.tsx > settings screen and switch together > restores the Connect form after a saved Off then On, without a reload
AssertionError: expected <p data-testid="backlog-off"></p> to be null
Tests  1 failed | 8 passed (9)
```

**Change.** Kandev v0.96.0 gives a plugin only `setIntegrationEnabled`; `PluginHostApi` (`apps/packages/plugin-sdk/src/index.ts`) has no getter or subscription for the published value (the registry's `getIntegrationEnabled` is host-internal). So the plugin uses a local channel: new `ui/src/switch/enabled-events.ts` with `publishEnabled(host, workspaceId, enabled)` (calls `host.setIntegrationEnabled` and then notifies local listeners) and `subscribeEnabled(listener)`. `PLUGIN_ID` moved there. The card switch (after a successful load and after a successful persist) and `index.ts` publish through it. `SettingsScreen` subscribes for its workspace and dispatches a new `enabledChanged` event; the reducer patches only `view.enabled` (the connection did not change, BR7.4), so no reload and no network call happen. A failed save publishes nothing, so the screen stays as it was.

**Tests** (`ui/src/switch/switch.test.tsx`, switch and screen mounted together): Off → On restores the Connect form and On → Off hides it, with exactly two `connection.get` calls (the first loads only); a value published for another workspace is ignored; a failed save leaves the screen in the Off state.

### R-03 (Minor): out-of-order replies on the Backlog page

**Red.** `cd ui && npx vitest run src/page`

```
× ignores a late reply for the previous workspace when loads overlap out of order
  Expected: "Connected as Test User @ example-space.backlog.com"  Received: "Backlog is turned off for this workspace."
× ignores a late failure for the previous workspace
  Expected: "Backlog is not connected."  Received: "Could not load Backlog."
Tests  2 failed | 17 passed (19)
```

**Change.** `BacklogPage.reload` takes a sequence number from a `useRef` counter and applies a reply or a failure only when it is still the latest load; the effect cleanup bumps the counter when the workspace changes or the page closes. This also covers a Retry overlapping a load.

**Tests** (`ui/src/page/backlog-page.test.tsx`): two overlapping loads (ws-1, then ws-2) resolving out of order, one with a late success and one with a late failure for ws-1; the page shows ws-2's state and link.

### Results

A fresh run of `make check-format vet lint test coverage build package verify-package` from `make clean` (Go 1.26.8, `GOTOOLCHAIN=local`, Kandev v0.96.0 `f099a46…`) passed every target. golangci-lint (with gosec): 0 issues; `tsc --noEmit`, ESLint and Prettier: clean. Go tests ran with `-race`, all packages `ok`. Vitest: 5 files, 65 tests passed (was 60). Go coverage **94.9%** (floor 80%, only `server/main.go` excluded); `internal/connection` 97.5%, `internal/plugin` 91.4%. `verifypkg: OK dist/nulab-backlog-0.0.1.tar.gz (nulab-backlog@0.0.1)`. `go mod tidy` leaves no diff; no dependency was added.

| Change | Files |
|--------|-------|
| Created | `ui/src/switch/enabled-events.ts` |
| Modified | `internal/connection/service.go`, `service_test.go`; `internal/plugin/runtime.go`, `actions_test.go`; `ui/src/index.ts`, `switch/integration-switch.tsx`, `switch/switch.test.tsx`, `settings/SettingsScreen.tsx`, `settings/state.ts`, `page/BacklogPage.tsx`, `page/backlog-page.test.tsx` |

## Loop-back 1 repairs (Steps 23–25)

Build and Test Loop-back 1 (`construction/build-and-test/test-results.md`) found T-RATE-01 and T-SEC-02 Not Met and T-PERF-03 Unverified. Steps 23–25 fix them in place with Red → Green → Refactor. Go 1.26.8, `GOTOOLCHAIN=local`, Kandev v0.96.0.

### Step 23 — Connect returns `rate_limited` at once (T-RATE-01, NFR2.1)

**Red.** `go test -race ./internal/backlog/ ./internal/connection/ -run 'TestNoRetry|TestRetryDisabled|TestConnectAsksForNoRetryOn429|TestOAuthSignInAsksForNoRetry|TestConnectionTestKeepsTheRetry'`. The first run failed to build (`undefined: NoRetry`, `undefined: RetryDisabled`). With no-op stubs:

```
--- FAIL: TestNoRetryReturns429AtOnce
    client_test.go:264: expected: 1 actual: 2   exactly one Backlog request   (short Retry-After)
    client_test.go:264: expected: 1 actual: 4   exactly one Backlog request   (reset 1 s ahead)
    client_test.go:264: expected: 1 actual: 4   exactly one Backlog request   (past reset is clamped)
--- FAIL: TestRetryDisabledReadsTheContext   Should be true
--- FAIL: TestConnectAsksForNoRetryOn429    Should be true  the Connect Myself call must not retry
--- FAIL: TestOAuthSignInAsksForNoRetry     Should be true  the sign-in Myself call must not retry
FAIL internal/backlog, FAIL internal/connection
```

**Green.** `backlog.NoRetry(ctx)` marks a context, and `backlog.RetryDisabled(ctx)` reads the mark. `Client.send` returns a 429 after the first attempt when the mark is set, so there is no wait and no retry. `RetryAfter` is still computed as before: `X-RateLimit-Reset`, then `Retry-After`, else 60 s, and never less than 1 s. `Service.verify` (API-key Connect) and `signIn` (the OAuth sign-in check) pass `backlog.NoRetry(...)` to `Myself`. Every other call keeps the shared policy: `TestMyselfRateLimitWait` still sees 4 hits on the clamp case (AC8.4.1–AC8.4.3), and `TestConnectionTestKeepsTheRetry` shows `connection.test` does not set the mark. The context option was chosen over a new Gateway method, so no Gateway interface or fake changed. Refactor: none needed.

### Step 24 — 4-character leak windows (T-SEC-02, NFR3.2)

**Red 1 (window).** `testutil.AssertNoLeak` lost its window argument and now always scans `testutil.LeakWindow` characters. All 55 call sites in the units' tests dropped their `, 8` argument (a mechanical edit that the compiler checks). With `LeakWindow = 8`, `go test -race ./internal/testutil/`:

```
--- FAIL: TestAssertNoLeak_Catches4CharWindow   Should be true  window at 0 must be caught   (x 58, every window of an APIKey and a Token)
--- FAIL: TestAssertNoLeak_PassesWithout4Char   Should be true  cdef is a 4-character window
FAIL internal/testutil
```

**Green 1.** `LeakWindow = 4`. Every package then passed once. A 40-times rerun (`go test -race -count=40` over backlog, connection, git, issues, plugin, redact, testutil) failed intermittently in 16 tests across backlog, connection, issues and plugin, each in at most 9 of 40 runs. Examples: `TestConnectNeverLeaksTheKeyOrDisplayName`, `TestU2_NoSecretInLogsErrorsViewsOrEvents`, `TestWebhookRedirectsWithTheOutcome`, `TestU3_IssueCalls_ErrorsNeverEchoTheBody`. A real leak of 4 or more characters would fail on every run. These failures were chance matches between the secrets' random hex and the digits, timestamps and hex IDs in log and fixture text (for example `2026` or `0710`). **No genuine leak was found.**

**Red 2 (generator).** As the plan directs, the generator was changed and the window was not widened. Tests first, `go test -race ./internal/testutil/`:

```
--- FAIL: TestAPIKeyHasFakePrefixAndRandomPart   Expect "test-api-key-<32 hex, redacted>" to match "^test-api-key-[GHJKLMNPQRSVWXYZ]{32}$"
--- FAIL: TestU2_TokenHasFakePrefixAndRandomPart Expect "test-token-<32 hex, redacted>" to match "^test-token-[GHJKLMNPQRSVWXYZ]{32}$"
--- FAIL: TestSecretRandomPartAvoidsHexAndDigits Expect "<32 hex, redacted>" to NOT match "[0-9a-z]"
FAIL internal/testutil
```

**Green 2.** `randomPart` draws 32 characters from `GHJKLMNPQRSVWXYZ`: 16 uppercase letters with no hex digit, no digit and no vowel. T is left out so that "HTTP" cannot be formed. That is 4 bits per character, 128 random bits in all. The `test-api-key-` and `test-token-` prefixes are unchanged, and `Windows` still skips them. After the change, `go test -race -count=40 ./internal/...` passed in every package with no failure. Refactor: `randomHex` was replaced, and the `encoding/hex` import was removed.

### Step 25 — 100-call timing of `connection.get` and `connection.set_enabled` (T-PERF-03, NFR1.1, NFR1.3)

`internal/plugin/actions_timing_test.go` runs inside `testing/synctest`. A `slowHost` wraps the rig's fake host and adds `storeLatency = 100 ms` of virtual time to each `GetState`, `SetState`, `GetSecret` and `SetSecret` call. That value is half the "well under 200 ms" store latency that `performance-requirements` assumes. After one Connect, each test calls the action 100 times through `HandleAction`, reads each duration from the synctest clock, and asserts that p95 ≤ 500 ms and that the fake gateway is never called. Both tests passed on the first run, so no code change was made. The measured p95 values were **400 ms** for `connection.get` (4 store reads) and **200 ms** for `connection.set_enabled` (1 read and 1 write). Raising the latency to 200 ms makes the `connection.get` test fail at 800 ms, so the test is not vacuous.

### Results

`make check-format vet lint test coverage check-secrets` passed. Results: gofmt clean; `go vet` clean; golangci-lint with gosec, 0 issues; `tsc`, ESLint and Prettier clean. Go tests ran with `-race` and every package was `ok`. Vitest: 28 files, 224 tests passed. Go coverage was **92.9%** against the 80% floor, with only `server/main.go` excluded. Per package: backlog 96.0%, ci 91.0%, connection 94.5%, git 90.8%, issues 91.3%, pkgverify 93.2%, plugin 93.6%, redact 97.4%, testutil 88.0%. `ci secrets: OK`. `go mod tidy` left no diff, and no dependency was added. The generated `coverage.out` was deleted.

| Change | Files |
|--------|-------|
| Created | `internal/plugin/actions_timing_test.go` |
| Modified (code) | `internal/backlog/client.go`, `internal/connection/service.go`, `internal/connection/oauth.go` (U2 file, one line), `internal/testutil/testutil.go` (shared by every unit) |
| Modified (tests: new tests) | `internal/backlog/client_test.go`, `internal/connection/service_test.go`, `internal/testutil/testutil_test.go` |
| Modified (tests: only the `, 8` argument removed) | `internal/backlog/{client,git_client,issues_client,limiter,oauth,oauth_types,projects,types}_test.go`, `internal/ci/contract_test.go`, `internal/connection/{apikey,events,git_credential,git_credential_types,oauth_config,service,store}_test.go`, `internal/git/{leak,resolver}_test.go`, `internal/issues/leak_test.go`, `internal/plugin/{actions,actions_u2,actions_u3,actions_u4,credential,webhook}_test.go`, `internal/redact/redact_test.go` |

### Findings and deviations (Loop-back 1)

- **Finding (NFR1.1 design drift, not fixed here):** `performance-design.md` says `connection.get` makes three store reads: record, secret and switch. It now makes **four**: the switch, the record, the connection secret and the Git credential secret `backlog.git.<ws>`. U4 added the Git credential read for `hasGitCredential`. The target is still met at 100 ms per read (400 ms), but the headroom is smaller. If a read takes more than 125 ms, the p95 goes over 500 ms. Either the design's read count is amended, or a later change drops the fourth read.
- **Deviation (test helper signature):** the plan said "change the window from 8 to 4". It was done by removing the window argument from `AssertNoLeak`, so no caller can choose a wider window again. This touched every unit's leak tests, but only to remove the argument.
- **Deviation (exported API):** `backlog.RetryDisabled` is exported so that the connection tests can assert the mark through their fake Gateway. Production uses it only inside `Client.send`.

## Loop-back 2 repair (Step 26)

Build and Test Run 2 found T-COMPAT-01 Not Met: the packaged-host contract test failed 2 of 7 runs with `connection.get answered 500`. Kandev v0.96.0's `pluginsdk/serve.go` dials the host broker from a background goroutine (up to 30 s) and only then calls `SetHost`, so an action could arrive first, and `hostStores.get` returned `errNoHost` at once.

### Step 26 — wait for the Host instead of failing (T-COMPAT-01, NFR6.1, R-04)

**Red.** `internal/plugin/runtime_host_test.go` builds the runtime without a Host (`hostlessRig`) and runs inside `testing/synctest`, so no test sleeps in real time:

- `TestHost_EarlyActionWaitsForHost` (a): `connection.get` sent before `SetHost` is still pending after 1 s of virtual time, then answers 200 at exactly 1 s, when `SetHost` is called.
- `TestHost_DeadlineFailsClosed` (b): a 1 s action deadline before `SetHost` answers 500 `internal` at exactly 1 s, logs `errNoHost`, and the Host set afterwards sees no read and no write.
- `TestHost_WaitIsCappedAt5s` (b): with no deadline, the wait stops at the 5 s cap.
- `TestHost_NilHostDoesNotOpen` (b): `SetHost(nil)` does not release the waiters.
- `TestHost_SetHostMeansNoWait` (c): after `SetHost`, three actions take zero virtual time.
- `TestHost_SetHostTwiceIsSafe` (c): eight concurrent `SetHost` calls racing eight actions under `-race`, then a second `SetHost` with another Host: no panic, and the latest Host serves the call.
- `TestHost_WorkersWaitWithoutStorm` (d): the PR watcher and the issue sync tick at 1 min with no Host, the Host arrives 500 ms later, and the logs have no ERROR line, at most two cycle lines per worker, and no cycle with `errors` > 0.

Failing output (after adding only the `hostWait` constant so the file compiled; before that it failed to build on `undefined: hostWait`):

```text
--- FAIL: TestHost_EarlyActionWaitsForHost (0.00s)   the action answered 500 before the Host was set
--- FAIL: TestHost_DeadlineFailsClosed (0.00s)       expected: 1s, actual: 0s
--- FAIL: TestHost_WaitIsCappedAt5s (0.00s)          expected: 5s, actual: 0s
--- FAIL: TestHost_NilHostDoesNotOpen (0.00s)        expected: 5s, actual: 0s
--- FAIL: TestHost_WorkersWaitWithoutStorm (0.00s)   watch_cycle and issue_sync_cycle at 09:01 with "errors":2
FAIL	github.com/khuongdo/kandev-plugin-nulab-backlog/internal/plugin
```

(c) passed already in Red: the old code never waited once a Host was set.

**Green.** `Runtime` overrides `SetHost`: it calls the embedded `UnimplementedPlugin.SetHost` (the latest Host wins) and, for a non-nil Host, closes `hostReady` through a `sync.Once`, so concurrent and repeated calls are safe. `waitHost(ctx, host, ready)` returns the Host at once when it is set; otherwise it waits on `ready`, `ctx.Done()` or the `hostWait` (5 s) timer, whichever is first, then fails closed with `errNoHost` (code `internal`). `hostStores.get` and `hostPort.get` now take the call's context and use `waitHost`; `HandleAction` calls it once before the switch guard, after the guarded action's 12 s deadline is set. A first Green that waited only inside the stores still failed (a) and the cap test: each store call runs under the store's 1 s `CallTimeout`, so an early `connection.get` failed after 1 s. Waiting once in `HandleAction`, inside the action's own budget, fixed it without changing any timeout. Refactor: none beyond the shared `waitHost`; tests stayed green.

`make package`, then `make contract-test KANDEV_MIN_DIR=../kandev-min` 10 times in a row: **10/10 passed** (`ci contract: OK nulab-backlog on Kandev v0.96.0` each time).

### Results (Loop-back 2)

`make check-format vet lint test coverage check-secrets build package verify-package` passed: gofmt clean, `go vet` clean, golangci-lint with gosec 0 issues, `ci workflows: OK`; Go tests with `-race` `ok` in every package; Vitest 28 files, 229 tests passed. Go coverage **92.9%** against the 80% floor, only `server/main.go` excluded; plugin 93.6%, connection 94.6%. `ci secrets: OK`; `verifypkg: OK dist/nulab-backlog-0.0.1.tar.gz`. No dependency added. `coverage.out` was deleted.

| Change | Files |
|--------|-------|
| Created | `internal/plugin/runtime_host_test.go` |
| Modified (code) | `internal/plugin/runtime.go`, `internal/plugin/host_port.go` (U4 file), `internal/plugin/config.go` (U2 file) |

### Findings and deviations (Loop-back 2)

- **Deviation (where the action waits):** the plan put the wait in `hostStores.get` and `hostPort`. It is there, but the store's 1 s per-call limit caps it at 1 s for actions, so `HandleAction` also waits once, up front, bounded by the action context and the 5 s cap. Budgets are unchanged: a guarded action's wait counts toward its 12 s deadline (14 s with rollback).
- **Note (workers):** background cycles still wait inside a 1 s store call. A Host that arrives later than that makes one cycle count errors in its single cycle line (no ERROR line, no storm), and the next tick runs normally.
- **Note (fail closed):** a wait that ends on the context returns `errNoHost`, not `ctx.Err()`, so the action reports `internal` as the plan requires rather than `unreachable`.

## Sources

- [desc] Initial description: Kandev plugin for Nulab Backlog.
- [scope] Workflow-selected scope: `feature`.
- `code-generation-plan.md` (Testing Contract and Steps 1–14) and `unit-test-instructions.md`.
- U1 `functional-spec.md`, `rules.md`, `entities.md`; `nfr-requirements/*`; `nfr-design/*`; `infrastructure-specification.md`; `cicd-pipeline.md`.
- `contract-summary.md` (C1, C4, C5, C8) and `stories.md` (US7.1, US1.2, US1.1, US7.2).
- `team.md` (Code Style, Testing Posture, Deployment) and `project.md` (Mandated, Forbidden).
- Kandev checkout `../kandev`:
  - `apps/backend/pkg/pluginsdk/{plugin.go,host.go,types.go}`;
  - `apps/backend/cmd/plugin-pack` and `apps/backend/internal/plugins/pkgtar/pkgtar.go` (the checksum rules mirrored by `internal/pkgverify`), at v0.96.0;
  - `apps/packages/plugin-sdk/src/index.ts`;
  - `apps/web/lib/plugins/host-api.ts` and `apps/web/lib/api/client.ts` (the `ApiError` shape the UI reads);
  - `docs/public/plugins-manifest.md`.
- Revision 2: plan Steps 15–22; functional-design Q6–Q9 (WF6, WF7, BR6.5, BR7.1–BR7.8); nfr-requirements Q4–Q6 (NFR1.3, NFR1.4, NFR3.9, NFR3.10); `performance-design.md` and `security-design.md`.
- Revision 2, Kandev v0.96.0:
  - `apps/packages/plugin-sdk/src/index.ts`: `registerNavItem`, `registerRoute` with `PluginRouteOptions.topbar`, `registerIntegrationSettings` with `icon` and `action`, `IntegrationSettingsActionProps`, `PluginIconProps`, `PluginContextApi.getWorkspaceIds` and `subscribeWorkspaces`, `setIntegrationEnabled`, `navigate`;
  - `apps/web/components/integrations/drafted-integration-enabled-control.tsx` and `lib/plugins/host-api.ts`: `IntegrationEnabledControl` props `{id, enabled, persist, name}`, with the id prefixed by the plugin id;
  - `apps/web/lib/settings/workspace-settings-tabs.ts` and `components/app-sidebar/sections/settings/settings-menu-branches.ts`: the `/settings/workspaces/<id>/integrations/<pluginId>` route.
- Loop-back 2: plan Step 26; `construction/build-and-test/test-results.md` (Run 2, Loop-back 2); Kandev v0.96.0 `apps/backend/pkg/pluginsdk/serve.go` (background host dial) and `plugin.go` (`HostSetter`, `UnimplementedPlugin`).
- Loop-back 1: plan Steps 23–25; `construction/build-and-test/test-results.md` (Loop-back 1) and `build-and-test-summary.md` (T-RATE-01, T-SEC-02, T-PERF-03); `performance-requirements.md` (NFR1.1, NFR1.3, NFR2.1 and the store-latency assumption).
- Nulab media assets <https://nulab.com/press/media-assets/> and logo guidelines <https://nulab.com/logo-guidelines/> (retrieved 2026-10-06).

## Assumptions & Open Questions

- [assumption] The settings screen cannot tell from Kandev whether the viewer is an admin, so it uses the functional-spec fallback: it shows the form, and a 403 from Connect switches to the member message.
- [assumption] Kandev passes `KANDEV_PLUGIN_LOG_LEVEL` to the plugin process (observability-design). If it does not, the level stays INFO.
- [assumption] Backlog returns the current user with a numeric `id` and a non-empty `name`. The first manual check against a real space confirms this.
- Resolved: the SDK pin is the v0.96.0 tag (user decision).
- Open: the in-repo verifier checks checksums, contents, identity and executables, but not Kandev's full manifest schema. For example, Kandev's action-key rule is enforced only by `TestManifestActionsAndAccess`. Kandev applies its full schema at install time, which the manual install at the skeleton checkpoint covers.
- Open: the action-key rename in Deviation 1 must be reflected upstream in `contract-summary.md` C5 before U2 adds more actions.
- Open (Revision 2): **the user must confirm Nulab's brand terms for the logo before the first release.** The logo guidelines allow logos "for development and marketing purposes ONLY" and forbid modifying them. The media assets page says its downloads are "for editorial purposes only" and that commercial use is prohibited. The two texts are not clearly consistent for a third-party plugin; see `docs/brand/backlog-logo.md`. The fallback is a one-line change to `PLUGIN_ICON`.
- Superseded (Review fix R-01): `set_enabled` now meets NFR1.3 and returns `{enabled}`. Open: amend WF6 step 4, `entities.md` ConnectionView and contract C5 to say so.
- [assumption] (Revision 2) Kandev's host CSS includes the `flex`, `flex-col`, `gap-4` and `gap-2` utility classes used for BR6.5. These are standard Tailwind utilities that Kandev's own settings pages use. The manual check confirms the spacing visually.
