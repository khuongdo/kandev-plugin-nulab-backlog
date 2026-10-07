# Code Generation Plan — Multi-provider source control (stage-level, express)

Target: zero-Unit, stage-level iteration (`<record>/construction/code-generation/`).
Inputs: `inception/requirements-analysis/requirements.md` (FR1-FR7, NFR1-NFR8, C1-C4, A1-A4), codekb `aidlc/spaces/default/codekb/kandev-plugin-nulab-backlog/`, developer scan `inception/reverse-engineering/developer-scan.md`. Units Generation and Functional Design are skipped by the express scope; the design below is derived from requirements + codekb.

## Design Summary

The smallest design that meets the requirements leaves Backlog Git (`internal/git`, `git.*` actions, v0.3.0 data) untouched and adds a separate, provider-neutral bounded context for the three external services.

- **`internal/scm`** (new): provider-neutral types (`Provider` id `github|gitlab|bitbucket`, `Repo`, `PullRequest`, `PRState`), the consumer-side `Client` port (4 real implementations exist: 3 here plus a test fake), a shared `HTTPError` type, a small stdlib HTTP helper (fixed host allow-list, `io.LimitReader`, rate-limit parsing with injected clock), the store (settings, mappings, links, dismissed auto-links, queries, watches, ledger) and the service (settings, mapping, PR list, link/unlink/auto-link, queries, watches, status refresh, watcher loop).
- **`internal/github`, `internal/gitlab`, `internal/bitbucket`** (new): stdlib clients implementing `scm.Client`: `CurrentUser`, `SearchRepos`, `GetRepo`, `ListPRs`, `GetPR`. Fixed API hosts `api.github.com`, `gitlab.com` (`/api/v4`), `api.bitbucket.org` (`/2.0`).
- **`internal/plugin`**: wires the clients, registers the new `scm.*` actions, adds them to `guarded` (FR6.1), maps `scm.HTTPError` to `ActionError` codes (NFR5).
- **UI**: a "Source control" Settings section (token, test, remove, project → repository mapping), a provider selector on the `/backlog` PR list, provider-aware saved queries and PR watches, and a PR section on the issue panel.
- **Tokens**: Kandev secret store, key `backlog.scm.<provider>.<workspaceId>`; never returned to the UI (NFR1).
- **No ConnectionChanged subscription** for `internal/scm`: items survive space change, disconnect and project deselect (FR6.2); views filter by the currently selected Backlog projects. Backlog Git keeps its coupling (FR6.3).

### Deviations from requirements wording (for approval)

- **NFR5**: one shared `scm.HTTPError{Provider, Status, RetryAfter}` type in `internal/scm` instead of three identical per-client error types. It is mapped once in `internal/plugin`; the requirement's intent (own type, mapped only in `internal/plugin`, no response body or secret in messages) is kept.
- **FR7.2**: new behaviour uses new `scm.*` action keys instead of adding an optional `provider` input to `git.*` actions. Existing `git.*` keys and inputs are therefore unchanged by construction, and Backlog Git stays the implicit provider of `git.*`.
- **OQ1** (Mandated host rule wording in `project.md`) is left to the human; this plan enforces NFR2 in code and tests and does not edit memory files.
- **OQ2** token scopes shown in Settings: GitHub fine-grained "Metadata: read" + "Pull requests: read" (classic: `repo`, or `public_repo` for public repositories); GitLab `read_api`; Bitbucket API token scopes `read:repository:bitbucket` + `read:pullrequest:bitbucket` (app password: Repositories Read, Pull requests Read).

## Blast Radius

| Package / area | Files | Change | Impact |
|---|---|---|---|
| `internal/scm` (new) | `doc.go`, `types.go`, `httpx.go`, `errors.go`, `store.go`, `service.go`, `links.go`, `queries.go`, `watcher.go` + tests, `testdata/` | create | low (new) |
| `internal/github`, `internal/gitlab`, `internal/bitbucket` (new) | `doc.go`, `client.go` + `client_test.go`, `testdata/*.json` each | create | low (new) |
| `internal/plugin` | `runtime.go` (wiring, `guarded`), new `scm_actions.go`, `errors` mapping, `host_port.go` (task PRs read for FR5.4) + tests, `manifest_test.go` expectations | modify | medium (action router shared by all features) |
| `manifest.yaml` | new `scm.*` action declarations only | modify | low |
| `ui/src/settings` | new `source-control-section.tsx`, `SettingsScreen.tsx` (mount section; move "Git access" into it) + tests | create/modify | medium |
| `ui/src/git` | `pr-list.tsx`, `pr-toolbar.tsx`, `save-query-dialog.tsx`, `git-state.ts`, `watch-form.tsx` (provider selector / provider field) + tests | modify | medium |
| `ui/src/issues` | `issue-panel.tsx` (PR section) + test | modify | low |
| `ui/src/settings/pr-watches-section.tsx`, `saved-queries-section.tsx` | provider column / filter | modify | low |
| docs | `README.md` (Source control section, token scopes), `CHANGELOG` if present | modify | low |
| Untouched | `internal/git`, `internal/connection`, `internal/backlog`, `internal/issues` (except none) | — | — |

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

### Step 1 — Project structure skeleton
- [x] Create package directories with `doc.go` (package doc comments citing FR IDs): `internal/scm`, `internal/github`, `internal/gitlab`, `internal/bitbucket`. No behaviour yet. (FR1.1, C2)

### Step 2 — Test runner readiness and brownfield baseline
- [x] Install Go 1.26.x if `go version` fails (user-local toolchain under `$HOME`, added to `PATH` for this session only).
- [x] Link the SDK: `ln -s /home/k_do_webfrontier/repo/kandev ../kandev` (worktree parent) and confirm `git -C ../kandev describe --tags` prints `v0.96.0` and matches `.kandev-sdk-ref`.
- [x] `npm ci` in `ui/`.
- [x] Record the baseline before any change: `make test`, `make coverage` (profile written under `build/`, not the repo root), `cd ui && npx vitest run`, `make lint`. Write counts (pass/fail/skip, coverage %) into `code-summary.md` § Baseline. Any pre-existing failure is reported, not fixed silently.
- [x] Confirm the exact scoped commands in `unit-test-instructions.md` run (they report "no test files" for new packages at this point).

### Step 3 — Data model (Red → Green → Refactor) — `internal/scm/types.go`, `store.go`
- [x] Red: `internal/scm/types_test.go`, `store_test.go`: provider ids are exactly `github|gitlab|bitbucket` (FR1.1); repository full-name validation per provider (`owner/name`, `group[/subgroup]/project`, `workspace/repo`) (FR3.2); PR web-URL parsing accepts only `https://github.com/<o>/<r>/pull/<n>`, `https://gitlab.com/<path>/-/merge_requests/<n>`, `https://bitbucket.org/<w>/<r>/pull-requests/<n>` and rejects other hosts, userinfo, query, fragment (FR1.2, FR5.1, NFR2); link key `provider|repo|number`; documents `scm.settings`, `scm.links`, `scm.dismissed`, `scm.queries`, `scm.watches`, `scm.ledger` round-trip with `schemaVersion: 1` and caps (links 500, queries 50, watches 50) (FR4.2); saved query validation (name 1-100 runes, ≥1 state, `anyone|me`) (FR4.2); state mapping incl. draft/declined labels (A3). Record failing output.
- [x] Green: implement types, validators and the store on the existing host state port pattern (same `StateStore` shape as `internal/git`).
- [x] Refactor while green.

### Step 4 — Repository / data access: HTTP helper and the three clients (Red → Green → Refactor)
- [x] Red: `internal/scm/httpx_test.go`: requests only to the configured fixed host over `https`; any other host is refused before dialing (NFR2); body read through `io.LimitReader` (NFR4); 401/403/404/429 → `HTTPError` with `Status`; `Retry-After` and GitHub `X-RateLimit-Remaining: 0` + `X-RateLimit-Reset` produce `RetryAfter` via injected clock (NFR3); error strings contain neither the token nor the response body (NFR1, NFR5); context cancellation returned unchanged.
- [x] Red: `internal/github/client_test.go`, `internal/gitlab/client_test.go`, `internal/bitbucket/client_test.go` against `httptest` fakes with `testdata/*.json` fixtures (fake tokens only): `CurrentUser` (FR2.4), `SearchRepos` with paging ≤100 per page (FR3.2, NFR4), `GetRepo`, `ListPRs` filtered by state with paging (FR4.1), `GetPR` (FR4.4, FR5.1); auth header shape (GitHub `Bearer`, GitLab `PRIVATE-TOKEN`, Bitbucket Basic `username:token`) (FR2.2); 401/403/404/429 fixtures (NFR3, NFR5).
- [x] Green: implement `httpx.go`, `errors.go` and the three clients (stdlib only, NFR6), each with a `BaseURL` override used only by tests.
- [x] Refactor while green.

### Step 5 — Business logic: `internal/scm` service (Red → Green → Refactor)
- [x] Red: `service_test.go` with a fake `Client`:
  - settings: set token stores it under `backlog.scm.<provider>.<ws>` and saves account name from `CurrentUser`; test returns account or classified error; remove deletes secret and keeps mappings/links/queries (disabled until a token exists) (FR2.2, FR2.4); state listing never contains the token (NFR1).
  - mapping: set per Backlog project → repositories, repositories validated through `GetRepo` (FR3.1, FR3.2); removing a mapping marks dependent queries/watches `unmapped` without deleting them (FR3.4).
  - PR list: only mapped repositories, grouped by Backlog project, page of ≤20 (FR3.3, FR4.1).
  - links: manual link by URL or picked PR to a task/issue, refused for unmapped repositories (FR5.1); auto-link when an issue key of a currently selected Backlog project appears in branch or title (FR5.2, A4); unlinking an auto-link records it in `scm.dismissed` and it is never re-created (FR5.3).
  - queries: save/list/delete/set default per provider (FR4.2).
  - watches: run creates at most one Kandev task per watch per run, respects the per-watch interval (default 5 min) via injected clock, ledger prevents duplicates (FR4.3).
  - status refresh: linked PR states refreshed on the watcher cycle; a 429 for one provider does not block other providers (FR4.4, NFR3).
  - coupling: no reaction to Backlog `ConnectionChanged`; items stay after space change/disconnect; views filter by selected projects (FR6.2).
  - redaction: run the service with a capturing `slog` handler and assert no token appears in logs or errors (NFR1).
- [x] Green: implement `service.go`, `links.go`, `queries.go`, `watcher.go` (one goroutine started/stopped by the runtime like `git.Watcher`).
- [x] Refactor while green (split files above ~400 lines).

### Step 6 — API / endpoint: `internal/plugin` + `manifest.yaml` (Red → Green → Refactor)
- [x] Red: `internal/plugin/actions_scm_test.go`, `manifest_test.go`:
  - every new action is declared in `manifest.yaml` with key pattern `^[a-z0-9][a-z0-9._-]*$`: `scm.providers.list`, `scm.providers.set_token` (admin), `scm.providers.test` (admin), `scm.providers.remove` (admin), `scm.repos.search` (admin), `scm.mappings.set` (admin), `scm.prs.list`, `scm.prs.link`, `scm.prs.unlink`, `scm.links.list`, `scm.task_prs.list`, `scm.queries.{list,save,delete,run,set_default}`, `scm.watches.{list,save,delete,run,pause,resume}` (FR2.6, FR4, FR5);
  - all `scm.*` actions are refused while Backlog is off except `scm.providers.list` (FR6.1);
  - existing action keys and their inputs are unchanged (count and names compared with the v0.3.0 list) (FR7.2);
  - `scm.HTTPError` mapping: 401/403 → reconnect required, 429 → unavailable, 404 → not found; public message has no body or token (NFR5, NFR1);
  - `scm.task_prs.list` returns Kandev's `TaskPullRequest` data (provider `github|gitlab`) for tasks linked to an issue without any plugin token (FR5.4);
  - v0.3.0 regression: load fixtures of `git.links`, `git.watches`, `git.queries`, the `backlog.git.<ws>` secret and `nulab_backlog_pr` metadata into the fake host, run `git.links.list`, `git.queries.list`, `git.watches.list`, `git.prs.status` and assert unchanged results (FR1.3, FR7.1).
- [x] Green: `scm_actions.go`, runtime wiring (clients, service, watcher start/stop), `guarded` update, error mapping, manifest entries.
- [x] Refactor while green.

### Step 7 — Frontend behaviour (Red → Green → Refactor)
- [x] Red (Vitest + jsdom + axe):
  - `ui/src/settings/source-control-section.test.tsx`: lists Backlog Git + 3 providers with state (FR2.1, FR2.5); admin can enter, test (shows account name or plain error), replace and remove a token; the token input is never re-filled from state (FR2.2-FR2.4, NFR1); required scopes text per provider (FR2.3); non-admin sees read-only view (FR2.6); mapping editor per selected Backlog project with repository search and manual `owner/name` entry (FR3.1, FR3.2); `data-testid` on every interactive control; axe passes.
  - `ui/src/git/pr-list.test.tsx` (new) + `git-state.test.ts`: provider selector; non-Backlog providers call `scm.prs.list` and show number, title, author, state incl. draft, branches, updated time, external link (FR4.1); Backlog Git path unchanged (FR1.3).
  - `save-query-dialog` / `watch-form.test.tsx`: provider field, default query per provider (FR4.2, FR4.3).
  - `ui/src/issues/issue-panel.test.tsx`: PR section shows scm links (manual and "auto-linked" badge with remove) and Kandev task PRs (FR5.1-FR5.4).
- [x] Green: implement the components; `SettingsScreen.tsx` mounts the new section and keeps the existing "Git access" form inside it for Backlog Git.
- [x] Refactor while green.

### Step 8 — Environment / build configuration
- [x] `go mod tidy` leaves `go.mod`/`go.sum` unchanged (no new dependency) (NFR6).
- [x] Run `make check-format vet lint test coverage` (coverage profile under `build/`), `cd ui && npx tsc --noEmit && npx eslint . && npx prettier --check . && npx vitest run`, `make package verify-package`, and `make contract-test KANDEV_MIN_DIR=../kandev` 10 times. The 80% floor is not lowered and no exclusion is added (NFR7).

### Step 9 — Documentation and traceability
- [x] `README.md`: "Source control" section (supported services, cloud only, token scopes from OQ2, project → repository mapping, auto-link rule, behaviour when Backlog is off/disconnected).
- [x] Write `code-summary.md` (files, decisions, baseline vs. after test counts, deviations), `source-manifest.json` (every created/modified path), `traceability.json` (FR1-FR7 sub-IDs and NFR1-NFR8 → implementation/test files).
