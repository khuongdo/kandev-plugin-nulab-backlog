# Code Quality Assessment — kandev-plugin-nulab-backlog

## Test Coverage and Baselines

- Go tests co-located in every `internal/*` package; `httptest` fakes with `testdata/` fixtures; SCM tests use `newHarness` with fake `Client`, secrets, state, clock and `CLIRunner` (no real `gh` runs).
- UI tests: Vitest + jsdom + axe-core under `ui/src/**/*.test.ts(x)` with the shared harness `ui/src/testing/harness` (`fakeHost`, `mount`, `byTestId`, `axeViolations`, `expectOnlyCatalogueText`, `expectTestIds`, `pseudoCatalogue`, `rawControls`).
- `make coverage`: 80% line floor over `./internal/... ./server/...`.
- **Baseline (this run, 2026-10-09, commit `3803248`)**: `go test -race ./internal/issues/... ./internal/plugin/...` (Go 1.26.8) -> both `ok`, no coverage profile written; `npx vitest run src/page src/issues` -> 14 files, 166 tests passed. Full `make coverage` and the contract test were not run.
- `ui/src/page/start-task.test.tsx:132-147` ("explains a missing workflow instead of opening the dialog") pins today's null-context behaviour and will change with the fix. The harness stub (`ui/src/testing/harness.ts:872`) always returns a context and fakes `TaskCreateDialog`, so "workflow exists but steps not hydrated" and the real host dialog are not exercised.
- Previous baseline (2026-10-08, commit `3d6a080`): `go test -race -count=1 ./internal/scm/... ./internal/github/... ./internal/gitlab/... ./internal/bitbucket/... ./internal/plugin/...` -> all 5 packages `ok`; `npx vitest run src/settings` -> 8 files, 121 tests passed. Full `make coverage` and the contract test were not run.
- Previous full baseline (commit `ca8146c`): `go test ./...` 13 packages, 1383 results passing, 0 failures.

## Linting

gofmt, go vet, golangci-lint with gosec (`.golangci.yml`), `tsc --noEmit` strict, ESLint (`ui/eslint.config.js`), Prettier (`ui/.prettierrc`), actionlint plus `cmd/ci workflows` policy. Conventions: `//nolint:gosec // G204` on `exec.CommandContext` with fixed args; `G101` on credential-like action keys.

## CI/CD

`ci.yml` (`checks`, `packaged-host-contract`), `secrets.yml`, `release.yml` (`verify` -> `contract` -> `publish` with attestation). Required checks on `main` via ruleset; squash only.

## Documentation

README present; `doc.go` per package; Go and TS doc comments cite FR/NFR/AC ids; `ui/src/messages/en.ts` is the only source of UI text (tests enforce it).

## Intent Findings: 261009-no-workflow-error

Request: "+ Task" on `/backlog` shows "Kandev has no workflow for this workspace yet." although a workflow exists; retouch the error display for mobile and desktop. Flow and host mechanics: [architecture.md](architecture.md#task-creation-context-host-derived-kandev-v0960).

| # | Area | Evidence | Change shape |
|---|---|---|---|
| 1 | Root cause | `start-task.tsx:54-56` treats a `null` `getTaskCreationContext` as "no workflow"; the host returns `null` also when steps are not in the web store, and the plugin route never loads them (`plugin-context-api.ts:14-25`, `spa-routes.tsx:369-380`) | Open `TaskCreateDialog` with `workflowId: null`, `defaultStepId: null`, `steps: []` and let the host resolve and fetch (`task-create-dialog-computed.ts:62-93`, `-effects.ts:32-57`); verify on the real host |
| 2 | Hidden-workflow fallback | Host picks the first workspace workflow when `activeId` is elsewhere; list includes hidden workflows (`plugin-context-api.ts:7-11`, `use-workflows.ts`) | Covered by finding 1 (the dialog resolves its own workflow) |
| 3 | Sibling callers | Same read and `errorWorkflow` notice in `settings/issue-watch-dialog.tsx:118-121`, `git/watch-form.tsx:168-171`, `git/scm-watch-form.tsx:119-122`; they must persist a real `workflowId` + `workflowStepId` | Decide in Requirements: in scope or not; `subscribeTaskCreationContext` and/or a notice with a next step |
| 4 | Real "zero workflows" case | After finding 1 the plugin no longer detects it; the host dialog must show it clearly | Verify host behaviour; keep `errorWorkflow` (`en.ts:157`) only where still reachable |
| 5 | Notice placement | Inline `<span role="alert" className="text-xs text-destructive">` beside the trigger (`start-task.tsx:79-92`) inside the row action slot (`issues-page.tsx:393-418`, `ROW` = `flex flex-wrap items-center gap-2`), which the host wraps in `div.shrink-0` (`change-request-list.tsx:84`): a sentence squeezes the `min-w-0 flex-1` title on mobile and shifts trigger and menu per row on desktop. Same in `git/pr-list.tsx:334`, `git/scm-pr-list.tsx:283` | Use `host.toast.error` (already used in `issues/task-menu.ts:34`) or a wrapping notice outside the action slot |
| 6 | Notice lifecycle | Cleared only by the next click; no dismiss; no next step (e.g. link to the board or workflow settings) | Toast auto-dismisses; add a next-step hint if a notice remains |
| 7 | Inconsistent notice styles | Unstyled `<p role="alert">` in `issues-page.tsx:582-586` and `BacklogPage.tsx:357`; `text-xs text-destructive` in `link-task-dialog.tsx:152-156` | One notice style across the plugin |

Not the cause: the workspace id (`BacklogPage.tsx:109,130` uses `getActiveWorkspaceId` / `subscribeActiveWorkspace`) and the Go backend (the notice is raised client-side; `issues.create_task` requires `workflowId`, `service.go:369-429`).

Constraints to preserve: catalogue-only text and axe checks, `backlog-` test ids (`${testId}-start`, `-start-item`, `-start-notice`), `min_kandev_version` 0.96.0, the `issues.link` / `git.prs.link` / `scm.prs.link` follow-up after task creation.

## Intent Findings: 261008-source-control-settings

Pre-0.6.0 analysis, kept for history; v0.6.0 implemented the one-service rule.

Request: clear separation between source control services on the settings page, repo-scope labels that name the service, and only one source control service usable at a time.

| # | Area | Evidence | Change shape |
|---|---|---|---|
| 1 | Visual separation | Backlog Git + 3 provider blocks are bare `<section className={STACK}>` siblings in one `flex flex-col gap-6` inside one `SettingsSection` (`source-control-section.tsx:466-482`); no frame, border or divider; provider name is `h4 text-sm font-medium` (`:367`), same weight as the repo `h5` (`:424`) | Frame each service (host UI kit card or bordered block), stronger heading with status; keep test ids |
| 2 | Repo-scope labels | `scmMappingsHeading` "Repositories of each selected Backlog project", `scmSearchLabel` "Search repositories for {project}", `scmManualLabel` "Repository name for {project}", `scmNoRepos` / `scmMappingLine` "{project}: ...", one shared `scmManualPlaceholder` "owner/name, group/project or workspace/repo" (`ui/src/messages/en.ts:390-398`); the service is implied only by DOM nesting | Add a `{provider}` param to these keys or new per-provider keys; per-provider placeholder format |
| 3 | Overloaded "scope" | `scmScopes*` texts (`en.ts:365-369`) are token permission scopes, unrelated to repo mapping | Avoid "scope" for the mapping area in new copy |
| 4 | One service at a time | No active-provider concept: per-provider `Settings` (`store.go:38-50`), `Providers` returns all three (`service.go:165-175`), connect/remove only touch their own entry (`service.go:227-406`), PR list and watch selectors offer every connected provider (`pr-list.tsx:127-132,217`; `watch-form.tsx:91-115`), link refresh loops over all (`watcher.go:137`) | Needs product decisions (below) before design |
| 5 | CLI notice | `scmNotice` maps `cli_unavailable` to literal `{cli: "gh / glab"}` (`ui/src/git/git-state.ts:194`) | Name the card's own CLI |

Open decisions for "one service at a time" (Requirements Analysis):

- **(a) Backlog Git**: does it count as a service? It is always on through the Backlog connection and its Git access form sits in this section (`SettingsScreen.tsx:424-438`).
- **(b) Enforcement**: UI only (hide/disable other cards) or a backend invariant (e.g. additive `activeProvider` in `scm.settings`, or "connecting one disconnects the others")? UI-only leaves the actions able to connect several.
- **(c) Inactive providers' data**: mappings, queries, watches and links of non-active providers — keep disabled (the `RemoveToken` / FR2.2 precedent, `service.go:392-406`) or delete?
- **(d) Upgrade path**: existing workspaces may have two or three providers connected; pick one automatically or ask the admin. `scm.settings` is schema version 1 and `load` refuses other versions (`store.go:229-233`): prefer an additive field over a version bump.
- **(e) Manifest / action keys**: a new key (e.g. `scm.providers.set_active`) changes `manifest.yaml`, `scmHandlers`, `TestSCM_Manifest_Actions` and probably the `guarded()` allow-list; `internal/plugin/testdata/v030/manifest.yaml` stays frozen.

Constraints to preserve:

- Test ids `backlog-section-source-control`, `backlog-scm-<provider>-*`, `backlog-scm-<provider>-map-<project>`, `backlog-scm-backlog`; catalogue-only text and axe checks (`source-control-section.test.tsx`, 25 tests; `sections.test.tsx` asserts no `backlog-scm-github-save` for members).
- `TestProviders_ListsAllThreeNotConfigured` assumes three entries; changing the list shape changes this test.
- NFR1: never show or refill tokens (`setToken("")` after every action, `source-control-section.tsx:252`); `ProviderView` stays non-secret.
- Everything must work on Kandev 0.96.0 (`min_kandev_version`).

## Technical Debt

- `ui/src/settings/source-control-section.tsx` (488 lines) holds three components with duplicated local `field`/`input` and `button` helpers (`:138-162` vs `:302-362`).
- `internal/scm/store.go:105` — `ponytail:` one settings document per list; `:165`, `:181` dismissed list / ledger never pruned.
- `internal/scm/cli_token.go:109` — `ponytail:` one global CLI lock; `internal/scm/watcher.go:200` — one watcher worker for all workspaces.
- `internal/plugin/host_port.go:139` — `workflowRefused` matches gRPC error text (`ponytail:`, previous run).
- `internal/github/client.go` — `SearchRepos` reads only the 100 most recent repos (`ponytail:`, previous run).
- Platform list declared in four places (see [code-structure.md](code-structure.md#build-and-packaging)).
- Dev environment: `../kandev` missing in fresh worktrees (required by `go.mod` and the `ui/tsconfig.json` alias `@kandev/plugin-sdk`).
- `subscribeTaskCreationContext` is available but unused; every caller reads the context once at click/open time.
