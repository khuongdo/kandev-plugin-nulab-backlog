# Code Generation Plan: Source Control Settings (one active service)

## Scope

Zero-Unit express work, brownfield. Source: `inception/requirements-analysis/requirements.md` (FR1-FR5, NFR1-NFR5) and the code knowledge base `aidlc/spaces/default/codekb/kandev-plugin-nulab-backlog/`. No new source files and no new dependencies; files are modified in place.

Design in one paragraph: the SCM settings document in `internal/scm/store.go` gains an additive `active` field (schema stays version 1). `scm.Service` derives the effective active service (stored value wins; otherwise 0 connected external providers -> `backlog_git`, 1 -> that provider, 2-3 -> "" pending, where everything works as today) and enforces it at its `settings()` chokepoint plus explicit guards. `internal/git` stays decoupled from `internal/scm`: `internal/plugin` guards Backlog Git actions and the Git credential, and injects an `Active` func into `git.Service` for its watcher. A new admin action `scm.active.set` switches the service; `scm.providers.list` also returns `active`. The UI shows a Service selector, a confirm dialog on switch, a pick notice while pending, one framed card for the active service, and service-named repository labels.

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
      "text": "- Keep the coverage profile out of the repository root during AI-DLC stages (delete coverage.out after local make coverage runs, or write it under build/); reviewers never use -coverprofile, because an unclaimed coverage.out blocks the Code Generation gate (learned 2026-10-07) \n\n- Install the Go toolchain (Go 1.26.x) and link ../kandev to the pinned v0.96.0 checkout before Reverse Engineering, so the scan records a Go test baseline instead of skipping it (learned 2026-10-07) \n\n- Run the packaged-host contract test locally with make contract-test KANDEV_MIN_DIR=../kandev while the SDK checkout is at the minimum version tag, and run it 10 times to catch host startup races (learned 2026-10-07) \n\n- When the worktree has no Go, install Go 1.26.x to ~/.local/go (checksum-verified from go.dev) and run make coverage, make package/verify-package and 10x make contract-test KANDEV_MIN_DIR=../kandev locally, so no target stays Unverified (learned 2026-10-08)"
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
  "input_sha256": "sha256:ca371ed31439c4d37a250d5e63d50b77aeaddf0755a2b7cfb8af1ee9221f1c2e",
  "contract_sha256": "sha256:dc2c3668efe5d53a084b8cac79c7c6d886c9148655498e84845031714e7646b7"
}
```

## Steps

Requirement IDs in brackets trace each step to `requirements.md`.

### Step 1: Project structure and configuration
- [x] No new packages, files or dependencies. Confirm `../kandev` resolves to the commit in `.kandev-sdk-ref` (v0.96.0) and `ui/node_modules` is installed. [NFR2]

### Step 2: Verify the test runners
- [x] Run the baseline commands from `unit-test-instructions.md` (Go `-race` for `internal/scm`, `internal/git`, `internal/plugin`; Vitest for the listed UI files) and record that they pass before any change.

### Step 3: Data model (settings document) - Red
- [x] `internal/scm/store_test.go`: a stored `active` value survives later `UpdateSettings` writes (SetToken/mapping) [FR1.3]; an existing document without `active` still loads with schema version 1 and no data loss [NFR2]. Record the failing output.

### Step 4: Data model - Green
- [x] `internal/scm/store.go`: add `Active string \`json:"active,omitempty"\`` to `doc[T]`; `update`/`put` keep the field on every write; add `Active(ctx, ws)` and `SetActive(ctx, ws, a)` under the workspace lock. `internal/scm/types.go`: add `BacklogGit Provider = "backlog_git"` and `Services = []Provider{BacklogGit, GitHub, GitLab, Bitbucket}`; `Providers`/`ParseProvider` unchanged.

### Step 5: Data model - Refactor
- [x] Remove duplication between `load`/`loadDoc`; tests stay green.

### Step 6: Business logic (active service rules) - Red
- [x] `internal/scm/service_test.go`, table-driven:
  - `SetActive` rejects `""` and unknown values; accepts the four services [FR1.1].
  - Derivation: 0 external connected -> `backlog_git` [FR1.2, FR3.1]; exactly 1 -> that provider [FR3.2]; 2-3 -> pending `""`, and both providers' `ListPRs` still work [FR3.3].
  - With GitHub active, every GitLab entry point (SetToken, UseCLI, SetMapping, Test, SearchRepos, ListPRs, SaveQuery, SaveWatch, RunQuery, RunWatch, Unlink) returns `InactiveError{Active: github}` whose message names GitHub; `RemoveToken` on GitLab still succeeds [FR1.4].
  - Switch GitHub -> GitLab -> GitHub: GitHub mappings, queries, watches and links are hidden while GitLab is active and come back unchanged [FR2.4].
  - `RemoveToken` on a provider that is active only by derivation stores the derived value first, so the workspace does not silently switch [FR2.4].
- [x] `internal/scm/watcher_test.go`: `runDue` and `RefreshLinks` only touch the active provider [FR1.5].
- [x] `internal/git/watcher_test.go`: when the injected `Active` func returns false, a watch cycle creates no task [FR1.6].
- [x] Update test harnesses once (`harness.connect` in `internal/scm/harness_test.go`) so existing tests keep their multi-provider setup: set active, connect, then clear the stored value so derivation reproduces today's behaviour. Record the failing output.

### Step 7: Business logic - Green
- [x] `internal/scm/errors.go`: `InactiveError{Active Provider}`.
- [x] `internal/scm/service.go`: `activeOf`, `Active`, `allow`, `SetActive`; guard in the `settings()` chokepoint and at the top of `SetToken`, `UseCLI`, `SetMapping`; `RemoveToken` stores the derived value before deleting.
- [x] `internal/scm/queries.go`, `links.go`: list functions hide items of a non-active provider; delete/default/state/unlink check the item's provider.
- [x] `internal/scm/watcher.go`: `RefreshLinks` loops only over allowed providers.
- [x] `internal/git/service.go`: exported `Active func(ctx context.Context, ws string) bool` (nil = always on); `internal/git/watcher.go` `cycleWatch` returns early when it is false.

### Step 8: Business logic - Refactor
- [x] Keep guards in as few places as possible; tests stay green.

### Step 9: API / plugin actions - Red
- [x] `internal/plugin/actions_scm_test.go`:
  - `scm.active.set` is admin-only, switches the service and returns `{providers, active}`; `scm.providers.list` returns `active` [FR1.3, FR2.2].
  - A non-active provider action returns status 409, code `service_inactive`, and `activeService` [FR1.4].
  - With GitHub active: `git.prs.list` and `connection.set_git_credential` are refused, `git.watches.list` / `git.queries.list` / `git.links.list` / `git.prs.status` return empty lists, and `ResolveGitCredential` is refused [FR1.6].
  - With `backlog_git` active, task PR summaries only include Backlog Git [FR1.5].
  - The new action passes the existing secret-leak check (no token in any response) [NFR1].
  - Add the key to `scmActions` and the admin list; update the action count. Update `scmRig.setToken` like the scm harness. Record the failing output.

### Step 10: API / plugin actions - Green
- [x] `internal/plugin/scm_actions.go`: `scm.active.set` handler; `scm.providers.list` adds `active`; `taskPRs` filters to the active service.
- [x] `internal/plugin/runtime.go`: `requireBacklogGit` guard in `HandleAction` for git handlers with the empty-list read table; map `InactiveError` to `service_inactive` / 409 with `activeService` in the error body; wire `git.Service.Active` after `wireSCM`.
- [x] `internal/plugin/credential.go`: `ResolveGitCredential` / `GetGitCredentialBinding` honour the Backlog Git guard.
- [x] `manifest.yaml`: add `scm.active.set` (workspace scope, admin, `max_body_bytes: 16384`). `internal/plugin/testdata/v030/manifest.yaml` stays unchanged.

### Step 11: API - Refactor
- [x] Tidy handler/table code; tests stay green.

### Step 12: Frontend behavior - Red
- [x] `ui/src/settings/source-control-section.test.tsx`:
  - Admin sees the Service selector (`backlog-scm-service`) with Backlog Git, GitHub, GitLab, Bitbucket and only the active service's card [FR2.1, FR1.1].
  - Member sees the active service read-only, no selector [FR2.2].
  - Changing the selector opens `backlog-scm-switch-dialog` naming the old and new service and that old data is kept but disabled; Confirm calls `scm.active.set`; Cancel sends no call [FR2.3, FR2.4].
  - Pending state (`active: ""`) shows `backlog-scm-pick-notice` and today's full layout [FR3.3].
  - The card is framed, has the service logo, an `h3` heading larger than the inner repository heading, and a Connected / Not connected badge [FR4.1, FR4.2].
  - Repository mapping heading, per-project label, search label and empty text name the service [FR5.1]; mapped lines render `[GitHub] acme/web` [FR5.2]; none of the new copy contains "scope" [FR5.3].
  - axe passes on the section, the selector is labelled [NFR3]; all text comes from `en.ts` [NFR4].
- [x] `ui/src/git/pr-list.test.tsx` / `watch-form.test.tsx`: with `backlog_git` active no provider selector is shown; with GitHub active only GitHub is offered; a `service_inactive` error shows a message naming the active service [FR1.5, FR1.4].
- [x] Update `scm.providers.list` mocks in `sections.test.tsx` and `settings.test.tsx` to include `active`. Keep all existing test ids. Record the failing output.

### Step 13: Frontend behavior - Green
- [x] `ui/src/settings/source-control-section.tsx`: load `active`; selector (reuse existing `Select`/`Label`), member read-only line, `createConfirmDialog` on switch, pick notice, framed card for the active service only, service-named labels and `[Service] owner/name` lines.
- [x] `ui/src/git/git-state.ts`: providers loader returns `active`; `prChoices` helper; `service_inactive` notice. `ui/src/settings/state.ts`: parse `activeService` in failures.
- [x] `ui/src/git/pr-list.tsx`, `watch-form.tsx`: use the active-service choices.
- [x] `ui/src/messages/en.ts`: reword `scmMappingsHeading`, `scmSearchLabel`, `scmManualLabel`, `scmNoRepos`, `scmRepoName`/`scmMappingLine` with `{provider}`; add `scmServiceLabel`, `scmServiceReadOnly`, `scmSwitchTitle`, `scmSwitchBody`, `scmPickNotice`, `scmServiceInactive`, `scmStateNotConnected`.

### Step 14: Frontend - Refactor
- [x] Remove the duplicated local field/button helpers in `source-control-section.tsx` only if touched code needs it; tests stay green.

### Step 15: Environment/build configuration
- [x] `make check-format vet lint test coverage` and `cd ui && npm run typecheck && npm run lint && npm run format:check` pass; coverage stays at or above 80% (delete `coverage.out` afterwards). No threshold is changed.

### Step 16: Documentation and traceability
- [x] README: short upgrade note (one active service; workspaces with several external services are asked to pick one; data of other services is kept but disabled).
- [x] Write `code-summary.md`, `source-manifest.json`, `traceability.json`.

## Risks

- Existing tests call `SetToken` on several providers; fixed centrally in the two harnesses, plus one-line changes where tests call the service directly.
- FR1.6 turns off the Backlog Git credential while an external service is active, so existing tasks cannot clone Backlog Git repositories then (as the user chose in Q8).
- Logo icon names in the host UI kit are confirmed during generation (open question in requirements).
