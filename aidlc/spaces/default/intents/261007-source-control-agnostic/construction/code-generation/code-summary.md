# Code Summary — Multi-provider source control

Stage-level, zero-Unit iteration (express scope). The plan is `code-generation-plan.md` (Steps 1-9, all ticked). The method was TDD per the embedded Testing Contract: each layer went Red, then Green, then Refactor.

## Baseline (Step 2, before any source change)

- Environment: Go 1.26.8 was already installed at `~/.local/go`, so nothing was installed. `../kandev` was linked to `/home/k_do_webfrontier/repo/kandev` at tag `v0.96.0`, which matches `.kandev-sdk-ref`. `npm ci` was run in `ui/`.
- `go test -race ./internal/... ./server/...`: 1201 passed, 0 failed, 0 skipped.
- `make coverage`: 92.8% (floor 80%). The profile was written under `build/`.
- `npx vitest run`: 31 files, 322 tests, all passed.
- `make lint`: clean.
- No test was failing before the change.

## After

| Check | Result |
|---|---|
| `go test -race ./internal/... ./server/...` | 1324 passed, 0 failed, 0 skipped (+123) |
| `make coverage` | 92.8% (floor 80%, no exclusion added). New packages: scm 93.1%, github 91.7%, gitlab 88.7%, bitbucket 90.2%. plugin 93.8% |
| Vitest | 33 files, 364 tests, all passed (+2 files, +42 tests) |
| `make check-format vet lint` | clean (gofmt, go vet, golangci-lint 0 issues incl. gosec, tsc strict, eslint, prettier, actionlint, ci workflows) |
| `make check-secrets` | OK |
| `go mod tidy` | `go.mod`/`go.sum` unchanged (NFR6: no new dependency) |
| `make package verify-package` | OK (`nulab-backlog@0.3.0`) |
| `make contract-test KANDEV_MIN_DIR=../kandev` | passed 10/10 on Kandev v0.96.0 |

## Files

92 application paths are listed in `source-manifest.json`.

- **`internal/scm`** (new, 11 source and 9 test files). It holds the provider-neutral types and validation (`types.go`), the `Client` port and the `Credential` type that hides its token (`client.go`), `HTTPError` and the sentinel errors (`errors.go`), and the stdlib HTTP helper (`httpx.go`). The helper sends HTTPS only to a fixed host, follows no redirects, reads bodies through `io.LimitReader`, parses rate limits with an injected clock and never retries. The rest is the store (`store.go`), the service (`service.go`, `prs.go`, `links.go`, `queries.go`) and the watcher (`watcher.go`).
- **`internal/github`, `internal/gitlab`, `internal/bitbucket`** (new). Each has `doc.go`, `client.go`, `client_test.go` and `testdata/` fixtures, which include the 401, 403, 404 and 429 cases.
- **`internal/plugin`**:
  - New: `scm_actions.go` (22 `scm.*` actions, `classifySCM`, `taskPRs`, the `scmHost` adapter), `actions_scm_test.go`, `v030_test.go` and the fixtures in `testdata/v030/`.
  - Modified: `runtime.go` (wiring, starting and stopping the watcher, `guarded` exempts `scm.providers.list`, `classify` delegates to `classifySCM`), `host_port.go` (`taskPullRequests`) and `manifest_test.go` (it now skips the `scm.*` keys, which `TestSCM_Manifest_Actions` covers).
- **`manifest.yaml`**: the 22 `scm.*` actions were added. No existing entry changed.
- **UI**:
  - New: `settings/source-control-section.tsx` (+test), `git/scm-pr-list.tsx`, `git/scm-watch-form.tsx`, `issues/issue-prs.tsx` and `git/pr-list.test.tsx`.
  - Modified: `git/git-state.ts`, `git/pr-list.tsx`, `git/save-query-dialog.tsx`, `git/watch-form.tsx`, `issues/issue-panel.tsx`, `messages/en.ts`, `page/BacklogPage.tsx`, `page/start-task.tsx`, `settings/SettingsScreen.tsx`, `settings/pr-watches-section.tsx` and `settings/saved-queries-section.tsx`.
  - Tests modified: `git-state.test.ts`, `watch-form.test.tsx`, `issue-panel.test.tsx` and `sections.test.tsx`.
- **Docs**: `README.md` gained a "GitHub, GitLab and Bitbucket" section with the OQ2 token scopes, and the settings list in "What it does" was updated.
- **Untouched**: `internal/git`, `internal/connection`, `internal/backlog` and `internal/issues`.

## Key decisions

- A separate bounded context, `internal/scm`, holds the new providers. Backlog Git keeps its own code, data and ConnectionChanged coupling (FR1.3, FR6.3). `internal/scm` never subscribes to ConnectionChanged (FR6.2). The Backlog switch is enforced by the action guard and by the watcher's `RequireEnabled` check (FR6.1).
- Tokens are stored as the secret `backlog.scm.<provider>.<workspaceId>`, a JSON `{token, username}`. `scm.settings` keeps only `hasToken`, the account and the last test error. No reply carries the token, and tests check the replies, logs, state and errors for it (NFR1).
- PR links in replies are built by `scm.PRURL`, never taken from the provider's `html_url`.
- Unmapped queries and watches are computed on read, not stored, so mapping a repository again enables them again (FR3.4).
- Auto-link covers the issue keys of projects that are both selected and mapped to the PR's repository (FR5.2, A4). It is best effort: a failure is logged and never fails the list or the watch run.
- `scm.task_prs.list` reads Kandev's `TaskPullRequest` data for every task linked to the same issue as the current task, capped at 20 tasks. A deleted task is skipped. Only `github` and `gitlab` PRs are kept, and no plugin token is used (FR5.4).

## Deviations and interpretations

1. **NFR5** (approved in the plan): there is one shared `scm.HTTPError` rather than one error type per client.
2. **FR7.2** (approved in the plan): the new behaviour uses new `scm.*` keys. `TestV030_ActionKeysAreUnchanged` compares all 55 v0.3.0 actions (key, scope, access and body limit) with a snapshot of the v0.3.0 manifest.
3. **429 mapping (NFR3, team rule "429 → Unavailable")**: provider 429s, and GitHub's 403 with `X-RateLimit-Remaining: 0`, map to the existing `rate_limited` code (HTTP 429 with `Retry-After`). That is the code Backlog 429s already use; the codebase has no separate "unavailable" code. 401 and 403 map to `reconnect_required`, 404 to `not_found`, and anything else to `unreachable`.
4. **No retries** for the providers. This meets "never more retries than the Backlog client".
5. **Token test (FR2.4)**: `scm.providers.test` answers 200 with `state: error` and a plain `lastError` (`invalid_token`, `missing_scope`, `rate_limited` or `unreachable`). `set_token` checks the token first and stores nothing if the provider refuses it.
6. **Watch run**: `scm.watches.run` runs synchronously and returns `{created}`. Backlog Git's run is queued, but a provider run creates at most one task, so it fits the action budget (FR4.3). The watcher ticks every minute, runs watches whose interval has passed, and refreshes link states every 5 minutes (FR4.4).
7. **Crash window**: the ledger reserves a PR before its task is created. A failed creation releases the reservation. A crash between the two loses at most one task rather than creating a duplicate. This is marked with a `ponytail:` comment; Backlog Git recovers through task metadata.
8. **Issue panel**: the existing test requires the panel to open with no inputs (read-only). The PR URL field therefore opens from a "Link a pull request" button rather than being shown at once.
9. **Settings layout**: the former "Git access" section is now inside the new "Source control" section, as the plan says. Its id changes from `backlog-section-git-access` to `backlog-section-source-control`. Members now see this section read-only (FR2.6). `sections.test.tsx` was updated to match.
10. **Provider choice in lists and forms**: only providers in the `connected` state are offered. A provider in the `error` state comes back after a successful Test.
11. **Files beyond the plan's blast-radius table**: `scm-pr-list.tsx`, `scm-watch-form.tsx` and `issue-prs.tsx` were split out to keep the existing components small. `pr-toolbar.tsx` needed no change, because it is reused through `idPrefix`.
12. **Step order**: the baseline (Step 2) was recorded before Step 1's `doc.go` files were written. This follows the dispatch instruction to record the baseline before any source change; the `doc.go` files hold no behaviour.
13. **Version**: `manifest.yaml` stays at `0.3.0`. The release version is set at Deployment Pipeline.

## Red/Green log

- **Step 3 (data model)**
  - Red: `go test ./internal/scm/...` failed to build (undefined `NewStore`, `Settings`, `GitHub` and others).
  - Green: `ok`, 98.8% coverage.
  - Refactor: a local that shadowed a builtin was renamed. Still green.
- **Step 4 (HTTP helper and clients)**
  - Red: all four packages failed to build (undefined `Credential`, `API`, `NewAPI`, `scm.Client`, `New` and others).
  - Green: scm 94.4%, github 91.7%, gitlab 88.7%, bitbucket 90.2%.
  - Refactor: PR URLs are now built by `scm.PRURL`, and `gofmt` and `go vet` are clean. Still green.
- **Step 5 (service)**
  - Red: the build failed (undefined `NewService`, `TokenInput`, `MappingInput`, `PRListInput` and others).
  - Green: `ok`, 61 tests, 93.5%.
  - Refactor: the PR list moved to `prs.go`, so `service.go` is 389 lines. gosec `//nolint` was added on fixture reads, following the repo's convention. Lint shows 0 issues. Still green.
- **Step 6 (plugin and manifest)**
  - Red: the build failed (undefined `actionSCMProviders` and others).
  - Green: on the first run, `TestSCM_Actions_EndToEnd` exposed a real bug. Unlinking an auto-link also removed the task's manual link to the same PR. It was fixed in `scm.Unlink`, and `TestUnlink_TargetsOneKindOfLink` was added. After that: plugin 93.8%, the full Go suite green, lint 0 issues, total coverage 92.8%.
  - Refactor: the scm error mapping stays in `scm_actions.go`. Still green.
- **Step 7 (UI)**
  - Red: 6 test files failed, with 24 tests failing and the source-control test failing to import.
  - Green: the existing read-only check on the issue panel failed once. The URL input was moved behind a button. After that, all 364 tests passed, and `tsc`, `eslint` and `prettier` were clean.
  - Refactor: `key?:` was added to the props types, following the repo's convention. Still green.

## Open issues

- **OQ1**: the Mandated host rule in `project.md` was not edited. Only the human can do that. NFR2 is enforced in code and covered by `TestAPI_RefusesAnyOtherHostOrSchemeBeforeDialing`, `TestAPI_NeverFollowsARedirect` and the `ParsePRURL` tests.
- **OQ2**: the token scopes in Settings and the README come from the providers' documentation as known at authoring time. They were not re-checked online during this stage.
- **Real providers**: there was no manual end-to-end check against real GitHub, GitLab or Bitbucket accounts. All tests use `httptest` fakes with fake tokens.
- `../kandev` is a symlink outside the repository, created for this stage.
