# Code Generation Plan — CLI login for GitHub (gh) and GitLab (glab)

## Scope

Zero-Unit express work. Source: `inception/requirements-analysis/requirements.md` (FR1–FR6, NFR1–NFR5) and the CodeKB `kandev-plugin-nulab-backlog` (scm credential flow). Brownfield: modify files in place.

Design in one paragraph: `scm.Settings` gains a `Source` field (`""`/`"token"` = typed token, `"cli"` = CLI login). `scm.Service.credential` — the single place every provider call gets its token — reads the token from the CLI when `Source == "cli"`, through an injectable runner with a 5-minute in-memory cache (clock = `Service.Now`). A new admin action `scm.providers.use_cli` runs the CLI once, checks the token with `CurrentUser`, deletes any typed-token secret and stores `Source = "cli"` + account. `SetToken` resets the source to token; `RemoveToken` clears it and the cache. CLI failures are one sentinel `scm.ErrCLIUnavailable` (no stderr, no token text). `ProviderView` gains `method` (`token`/`cli`). GitLab switches to `Authorization: Bearer` for all tokens (works for personal access tokens and glab OAuth tokens). The UI adds a "Use gh CLI login" / "Use glab CLI login" button on the GitHub and GitLab cards.

Files expected to change:
- `internal/scm/cli_token.go` (new), `internal/scm/cli_token_test.go` (new)
- `internal/scm/store.go`, `internal/scm/service.go`, `internal/scm/errors.go`, `internal/scm/watcher.go`, `internal/scm/service_test.go` (+ fakes/harness as needed)
- `internal/gitlab/client.go`, `internal/gitlab/client_test.go`
- `internal/plugin/scm_actions.go`, `internal/plugin/*_test.go` (classify + handler), `manifest.yaml`
- `ui/src/git/git-state.ts`, `ui/src/settings/source-control-section.tsx`, `ui/src/settings/source-control-section.test.tsx`, `ui/src/messages/*.ts`
- `README.md` (short section on the CLI login option)

## Testing Contract

```json
{
  "version": 1,
  "methodology": "tdd",
  "source": "team",
  "ordering": "For each testable layer, we write a failing test first, then write the minimum code to make it pass, then refactor while the test stays green, before moving to the next behaviour or layer.",
  "scope": "express",
  "test_strategy": "minimal",
  "project_type": "brownfield",
  "applicable_notes": [
    {
      "layer": "org",
      "text": "We treat tests as a first-class deliverable in every Bolt. The specific\nmethodology (TDD, BDD, ATDD, or classic test-after) is affirmed at\npractices-discovery and recorded in `team.md` under this heading with explicit\n`Methodology` and `Ordering` fields; Code Generation resolves those fields\nindependently from coverage, tooling, and scope notes.\n\nWhen no posture has been affirmed, our default per scope is:\n- **Methodology**: test-after\n- **Ordering**: implement each applicable testable layer, then write and run\n  that layer's tests.\n- `mvp`, `enterprise`, `feature`, `infra`, `classic` add an 80% line-coverage\n  floor and CI execution before merge.\n- `bugfix`, `security-patch` add a targeted regression for the specific\n  bug/vulnerability and require the existing suite to remain green.\n- `express` uses the Minimal strategy: requirement-driven unit tests (one per\n  requirement, with a happy-path floor per component); existing tests remain\n  green.\n- `poc`, `refactor`, `workshop` add no extra new-test floor and require the\n  existing suite to remain green.\n\nThe active `Test Strategy` still applies in every scope and determines test\nvolume/types. Scope floors are additive; they never reduce or replace the\nselected strategy.\n\nBuild and Test verifies defined coverage floors and affirmed quality targets;\nthey may not be weakened to make a step pass.\n\nAffirm a stricter posture in `team.md` if the team commits to one."
    },
    {
      "layer": "team",
      "text": "- We treat tests as a deliverable of every Bolt, not extra work.\n- **Methodology**: tdd\n- **Ordering**: For each testable layer, we write a failing test first, then write the minimum code to make it pass, then refactor while the test stays green, before moving to the next behaviour or layer.\n- A floor of **80% line coverage for the Go code**, measured with `go test -coverprofile` and enforced by a `make coverage` target; CI blocks the merge when it is lower. The measured scope is `./internal/...` and `./server/...`; only the minimal wiring in `main` is excluded, and every exclusion must be listed explicitly in the `Makefile`. We do not lower the floor or add exclusions to make it pass.\n- Go tests always run with **`go test -race`** (needs CGO; works on the `ubuntu-latest` runner). This is an addition to the Bitbucket template, because token refresh and API rate limiting are two places prone to data races.\n- **Automated contract test with the real package on the minimum Kandev version**, like the template's `packaged-host-contract` job: CI checks out Kandev at exactly the `min_kandev_version` in `manifest.yaml`, builds a throwaway server, installs the packaged plugin, and runs it.\n- **Manual end-to-end check against a real Backlog space exactly twice**: when the walking skeleton is done, and before the first release. Each result is recorded. After the first release, we rely only on automated tests.\n- Unit and integration tests do not call real Backlog: they use a **fake Backlog server built with `net/http/httptest`**, with JSON sample data in `internal/<package>/testdata/`, including the 401, 429 (with `Retry-After`), and expired-token error cases.\n- Tools: `testing` + `github.com/stretchr/testify/require`, table-driven tests with `t.Run`, test names that describe behaviour. The clock and wait functions are injected into the code so rate limits and token expiry can be tested without a real `time.Sleep`. Tests are independent of each other and use `t.TempDir()` and `t.Setenv()`. The TypeScript UI (if any) is tested with Vitest.\n- There is a test asserting that API keys and tokens do not appear in logs, error messages, or responses sent to the UI.\n- We **do not add a dependency vulnerability scan** (`govulncheck`, `npm audit`, Dependabot) — this is your decision in Q4."
    },
    {
      "layer": "project",
      "text": "- Keep the coverage profile out of the repository root during AI-DLC stages (delete coverage.out after local make coverage runs, or write it under build/); reviewers never use -coverprofile, because an unclaimed coverage.out blocks the Code Generation gate (learned 2026-10-07) \n\n- Install the Go toolchain (Go 1.26.x) and link ../kandev to the pinned v0.96.0 checkout before Reverse Engineering, so the scan records a Go test baseline instead of skipping it (learned 2026-10-07) \n\n- Run the packaged-host contract test locally with make contract-test KANDEV_MIN_DIR=../kandev while the SDK checkout is at the minimum version tag, and run it 10 times to catch host startup races (learned 2026-10-07)"
    }
  ],
  "obligations": {
    "strategy": "minimal",
    "strategy_volume": [
      "One verifiable test per requirement at the narrowest effective level.",
      "At least one happy-path unit test per component.",
      "Unit tests are the default; a bugfix/security scope floor may require an integration or E2E regression when that is the narrowest level that reproduces the defect."
    ],
    "scope_floor": [
      "Keep the existing test suite green.",
      "This scope adds no extra new-test floor beyond the selected test strategy."
    ],
    "combination_rule": "Apply every selected-strategy obligation and every scope-floor obligation; neither replaces the other, and a targeted scope regression may add the narrowest necessary test type beyond the strategy default."
  },
  "plan_profile": {
    "methodology": "tdd",
    "runner_step": "Verify the existing test runner/configuration and record the exact unit-scoped command.",
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
      "Verify the existing test runner/configuration and record the exact unit-scoped command.",
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
  "input_sha256": "sha256:4442c5fad82de74b31fc5e573b07e7444027e7437773b5002fb422bdfb2749f7",
  "contract_sha256": "sha256:6db87aafc6be218a4f8f762cfd33d3fec2e691661f16cdc25e6dcc4236038b2d"
}
```

## Steps

Layers that do not apply: no database or migration (provider settings stay in the existing plugin-state JSON; the new `Source` field is additive, FR6.1).

- [x] **Step 1 — Environment.** Make sure Go is on PATH (`~/.local/go/bin`) and `../kandev` exists as a link to the pinned v0.96.0 checkout `~/repo/kandev` (outside the repo; nothing in the repo changes). Run `npm ci` in `ui/` if `ui/node_modules` is missing. (Contract: project structure / environment)
- [x] **Step 2 — Runner check + baseline.** Run the exact commands in `unit-test-instructions.md` before any change and record the pass counts (brownfield test baseline).
- [x] **Step 3 — Data model, Red.** In `internal/scm`: tests that a `Settings` without `source` (v0.5.0 JSON) reads as the token method and its `ProviderView.method` is `token` (FR6.1, FR5.1); a `Settings` with `source: "cli"` shows `method: "cli"`, state `connected`, and no token field in the view's JSON (NFR1). Record the failing output.
- [x] **Step 4 — Data model, Green.** Add `Source string \`json:"source,omitempty"\`` to `Settings`, constants `MethodToken = "token"`, `MethodCLI = "cli"`, and `Method string \`json:"method,omitempty"\`` to `ProviderView` set by `view()` (empty when not configured).
- [x] **Step 5 — Data model, Refactor** while green.
- [x] **Step 6 — CLI token source, Red.** New `internal/scm/cli_token_test.go` with a fake runner (no real `gh`/`glab`): (a) GitHub runs `gh auth token --hostname github.com`, GitLab runs `glab config get token --host gitlab.com` — fixed argument lists; (b) stdout is trimmed; (c) empty output, output with spaces/control characters, a non-zero exit, a missing binary (`exec.ErrNotFound`) and a timeout all return `ErrCLIUnavailable` whose message contains neither the stderr text nor the token (NFR1); (d) Bitbucket is refused; (e) the cache returns the same token for 5 minutes on an injected clock and asks the runner again after expiry (FR3.2, NFR4); (f) `forget` clears one provider's cache. Record the failing output.
- [x] **Step 7 — CLI token source, Green.** New `internal/scm/cli_token.go`: `ErrCLIUnavailable` (in `errors.go`), a `CLIRunner func(ctx context.Context, name string, args ...string) ([]byte, error)` default using `exec.CommandContext` with a 10-second timeout, stdout limited to 4 KiB, stderr discarded (`//nolint:gosec // G204: fixed arguments`) (NFR2); `cliCommand(p)`; a small mutex-guarded cache `map[Provider]cachedToken{token, at}` on `Service`, fields `CLI CLIRunner` and `cliTTL = 5 * time.Minute`, using `Service.Now`.
- [x] **Step 8 — CLI token source, Refactor** while green; run with `-race`.
- [x] **Step 9 — Service, Red.** In `internal/scm/service_test.go` (harness with the fake runner): (a) `UseCLI(github)` runs the CLI, calls `CurrentUser` with that token, deletes the typed-token secret, stores `Source=cli`, account and `HasToken=true`, returns `method: cli` (FR1.2, FR1.3); (b) `UseCLI(gitlab)` works the same (FR2.1); (c) `UseCLI(bitbucket)` is a field error on `provider` (FR2.3); (d) `UseCLI` with a failing CLI returns `ErrCLIUnavailable` and changes nothing (FR4.1); (e) `UseCLI` with a token the provider rejects (401) returns that error and changes nothing; (f) `credential()` for a CLI provider returns the CLI token and never reads the secret store (FR3.1); (g) `SetToken` after `UseCLI` resets the method to token and clears the cache (FR1.3); (h) `RemoveToken` on a CLI provider clears source, account and cache, keeps mappings (FR5.3); (i) `Test` on a CLI provider forgets the cache first, and a failing CLI records `LastError = "cli_unavailable"` with state `error` (FR4.2, FR4.3), and a later working CLI clears it; (j) after the CLI's token changes, calls use the new token once the cache expires (FR3.3); (k) the PR watcher's refresh for a CLI provider that gets a 401 forgets the cached token (FR3.2); (l) leak test: no CLI token in logs, error strings or the provider-list JSON (NFR1). Record the failing output.
- [x] **Step 10 — Service, Green.** `Service.UseCLI(ctx, ws, p)`, `credential()` branch on `Source`, `SetToken`/`RemoveToken` updates, `Test` forget + `cli_unavailable` (`ErrorCLIUnavailable` constant) recorded as a result, `refreshProvider` forget on 401, `errorCode` maps `ErrCLIUnavailable`. CLI tokens registered with `redact.WithSecrets`.
- [x] **Step 11 — Service, Refactor** while green.
- [x] **Step 12 — GitLab client, Red.** Update `internal/gitlab/client_test.go` to expect `Authorization: Bearer <token>` instead of `PRIVATE-TOKEN` (FR2.2). Record the failing output.
- [x] **Step 13 — GitLab client, Green.** Change the header in `gitlab.New()`.
- [x] **Step 14 — GitLab client, Refactor** while green.
- [x] **Step 15 — Plugin action, Red.** In `internal/plugin` tests: (a) handler `scm.providers.use_cli` decodes `{provider}` and calls `UseCLI`; (b) `classifySCM(ErrCLIUnavailable)` returns code `cli_unavailable`; (c) the manifest/runtime parity test fails until `scm.providers.use_cli` is declared (FR6.2). Record the failing output.
- [x] **Step 16 — Plugin action, Green.** Add `actionSCMUseCLI = "scm.providers.use_cli"` and its handler; map `ErrCLIUnavailable` in `classifySCM`; declare the action in `manifest.yaml` with `scope: workspace`, `access: admin`, `max_body_bytes: 16384` (FR1.4).
- [x] **Step 17 — Plugin action, Refactor** while green.
- [x] **Step 18 — UI, Red.** In `ui/src/settings/source-control-section.test.tsx`: (a) the GitHub card shows a `backlog-scm-github-use-cli` button labelled "Use gh CLI login" and the GitLab card `backlog-scm-gitlab-use-cli` "Use glab CLI login"; Bitbucket has none; members (read-only) see none (FR1.1, FR2.1, FR2.3, FR1.4); (b) clicking it invokes `scm.providers.use_cli` with `{provider}` and shows the returned state; (c) a view with `method: "cli"` shows "Connected via gh CLI as <account>" (FR5.2); (d) a `cli_unavailable` failure from the action and a `lastError: "cli_unavailable"` after Test both show the CLI message (FR4.2). Record the failing output.
- [x] **Step 19 — UI, Green.** `ProviderView.method?: "token" | "cli"` in `ui/src/git/git-state.ts` (and `scmNotice` handles code `cli_unavailable`); button + account line in `ProviderCard`; `TEST_ERRORS.cli_unavailable`; new message keys in every `ui/src/messages/*.ts` file.
- [x] **Step 20 — UI, Refactor** while green; `tsc --noEmit`, ESLint and Prettier clean.
- [x] **Step 21 — Full check.** Run `make check-format vet lint test` (and `make coverage` with the profile kept out of the repo root, then delete it) so the existing suite stays green and the 80% floor holds.
- [x] **Step 22 — Documentation.** README: a short "Connect GitHub or GitLab with the CLI login" section — the CLI runs on the Kandev server under the server's user, needs `gh` ≥ 2.17 / `glab` logged in, does not work when the server has no CLI (e.g. Docker); token cache 5 minutes.
- [x] **Step 23 — Records.** `code-summary.md`, `source-manifest.json`, `traceability.json` (FR/NFR → files).

## Requirement-to-step traceability

| Requirement | Steps |
|---|---|
| FR1.1–FR1.4 | 9, 10, 15, 16, 18, 19 |
| FR2.1–FR2.3 | 6, 7, 9, 10, 12, 13, 18, 19 |
| FR3.1–FR3.3 | 6, 7, 9, 10 |
| FR4.1–FR4.3 | 6, 7, 9, 10, 15, 16, 18, 19 |
| FR5.1–FR5.3 | 3, 4, 9, 10, 18, 19 |
| FR6.1–FR6.2 | 3, 4, 15, 16 |
| NFR1 | 3, 6, 9 (leak test) |
| NFR2 | 6, 7 |
| NFR3 | 6–20 (fake runner, injected clock, `-race`), 21 |
| NFR4 | 6, 7 |
| NFR5 | 2, 21 |

## Known limits

- `ponytail:` a cached CLI token is forgotten on 401 only in `Test` and the PR watcher; an action that gets a 401 reports "reconnect required" and the next Test or the 5-minute expiry picks up the new token.
- The glab command `glab config get token --host gitlab.com` reads glab's stored login; a token given to glab only through the `GITLAB_TOKEN` environment variable is not seen. Documented in the README.
