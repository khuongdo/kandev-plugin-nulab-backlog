# Code Quality Assessment — kandev-plugin-nulab-backlog

## Test Coverage and Baselines

- Go: ~108 test files, about 704 `func Test` (count from the original full scan), co-located in every `internal/*` package (`httptest` fakes, `testdata/` JSON), run with `-race`.
- UI: 33 Vitest files under `ui/src/` (about 283 `it` cases, original full scan), jsdom, shared fake host `ui/src/testing/harness.ts`.
- `make coverage`: 80% line floor over `./internal/... ./server/...`, sole exclusion `server/main.go`, profile at `build/coverage.out`.
- SCM tests: fake `Client`, `fakeSecrets`, `fakeState` and `newHarness` (`internal/scm/harness_test.go:232-250`); GitHub client tests use an `httptest` fake GitHub with user/repo/pull fixtures and 401, 403 rate limit, 404, 429 cases; plugin-level SCM tests in `internal/plugin/actions_scm_test.go`; secret-leak tests in `internal/scm/httpx_test.go` and `internal/plugin/actions_test.go`; UI tests in `ui/src/settings/source-control-section.test.tsx`.
- **Baseline (run 3, 2026-10-08, commit `d3d17e5`)**: `go test -race ./internal/scm/... ./internal/github/... ./internal/connection/...` -> **407 passed in 3 packages**; `go test -race ./internal/plugin/... ./internal/git/...` -> **302 passed in 2 packages** (709 total, 0 failures). No `-coverprofile`; UI (Vitest) baseline not recorded. Environment: Go 1.26.8 at `~/.local/go/bin` (not on `PATH`); `../kandev` is missing in this worktree, so the run used a temporary `go.work` in the session scratchpad that points the SDK at `~/repo/kandev/apps/backend` (v0.96.0, matches `.kandev-sdk-ref`).
- Earlier runs (261007, 261008-ci-path-filter, 261008-fix-uiux-backlog) recorded no baseline.

## Linting

gofmt, go vet, golangci-lint (`.golangci.yml`, +gosec), `tsc --noEmit` strict, ESLint, Prettier, actionlint, `cmd/ci workflows` policy (SHA-pinned actions, top-level `permissions` exactly `contents: read`, no `pull_request_target`, write permission only in `release.yml` job `publish`). Existing gosec conventions: `//nolint:gosec // G204` on `exec.CommandContext` with fixed arguments (`internal/ci/changes.go:41`); `//nolint:gosec // G101` on the `scm.providers.set_token` action key; `//nolint:gosec // G117` on `Credential.Token`.

## CI/CD

- `ci.yml`: `checks` (format, vet, lint, test, coverage, check-secrets, build, package, verify-package) and `packaged-host-contract` (installs the package on Kandev built at `v0.96.0`). A CI path filter for app-only changes was added after run 1 (PR #14) and is not re-verified here.
- `secrets.yml`: separate credential scan.
- `release.yml`: `verify` -> `contract` -> `publish` (build provenance attestation, `gh release create`) on `v*` tag push.
- Required checks `checks` and `packaged-host-contract` on `main` via ruleset `24580280`; squash-only. Details: [api-documentation.md](api-documentation.md#github-actions-triggers-and-required-checks).

## Intent Findings: 261008-gh-cli-auth

Request: add GitHub CLI login as a second way to connect the GitHub source-control provider, next to the access token. Current flow: [architecture.md](architecture.md#interaction-diagrams).

| # | Area | Evidence | Change shape |
|---|---|---|---|
| 1 | Credential source | One private read path `(*scm.Service).credential` (`internal/scm/service.go:221-237`) with 7 callers, including the 1-minute watcher | Choose the source inside `credential()`: for GitHub with the CLI method, run `gh auth token --hostname github.com` instead of reading the secret; callers stay unchanged |
| 2 | Stored method | `scm.Settings` (`internal/scm/store.go:41-48`) has no notion of credential source; schema version 1, `load` checks only `schemaVersion` | Add an `omitempty` field (e.g. auth method `token` / `gh_cli`); old documents stay readable |
| 3 | Connect action | Only `scm.providers.set_token` exists; manifest/handler parity test (`internal/plugin/manifest_test.go:136-142`) | New admin action (e.g. `scm.providers.use_gh_cli`) mirroring `SetToken`: resolve the token, validate with `CurrentUser`, record account and method; add the manifest entry (key must match `^[a-z0-9][a-z0-9._-]*$`) |
| 4 | View and state | `view` derives `state` only from `HasToken` + `LastError`; UI `hasToken = state !== "not_configured"`; `usableProviders` keeps only `connected` | A CLI-backed provider must still report connected (set `HasToken`-equivalent truth) and expose the method as a non-secret `ProviderView` field so the card can show it |
| 5 | Remove | `RemoveToken` deletes the secret and clears account fields | Also clear the method; deleting a missing secret must stay harmless |
| 6 | Errors | `ErrNoToken` text says "add one under Source control"; `classifySCM` maps it to `validation`/`token` | New sentinel for an unusable CLI (e.g. not installed, not logged in, timeout) with its own mapping; keep `gh` stderr out of messages |
| 7 | UI | `ProviderCard` in `ui/src/settings/source-control-section.tsx` offers only the token form | GitHub-only control to use the CLI login, plus showing which method is active; new `scm*` strings in `ui/src/messages/en.ts`; TS `ProviderView` in `ui/src/git/git-state.ts` |

Constraints and risks:

1. **Where `gh` runs.** The token is read on the Kandev server host, from the environment of the OS account that runs Kandev (see [architecture.md](architecture.md#system-overview)). In Docker or when `gh` is absent or logged out, the action must fail with a clear typed error. The host offers no API to borrow Kandev's own GitHub credential, so the plugin must exec `gh` itself (stdlib `os/exec`).
2. **Resolve per call vs import once.** Per call keeps the token current (survives `gh auth refresh`) but runs a subprocess on every `credential()` call, including the watcher each minute — needs a short timeout and possibly a small in-memory TTL cache. Import once into the secret store means no runtime exec but the copy goes stale when `gh` rotates the token. Decision belongs to Requirements/Design.
3. **Secrets.** Trim `gh` stdout, wrap it with `redact.WithSecrets`, never log or return `gh` stderr, never persist the CLI token in plugin state (a cache, if any, stays in memory). Project rules: redact tokens in logs, errors and test output; no real credentials in tests. Add a leak test like the existing ones.
4. **Testability.** Inject the command runner as a `Service` field (like `Now`) so tests never exec a real `gh`; wire the fake in `newHarness`.
5. **Lint.** `exec.CommandContext("gh", ...)` with fixed arguments needs `//nolint:gosec // G204: fixed arguments`.
6. **Scope.** Only GitHub has this path; GitLab (`glab`) and Bitbucket stay token-only unless requirements say otherwise. `scm.task_prs.list` uses Kandev's own PR data and is unaffected.
7. **Reference shape.** Kandev itself runs `gh auth token --hostname <host> [--user <login>]` and handles older `gh` without `--user` and an empty token (Kandev `internal/github/gh_accounts.go:201-215`, external).
8. **Construction environment.** Link `../kandev` to `~/repo/kandev` (v0.96.0) or use an out-of-tree `go.work`; put Go 1.26.x on `PATH`.

## CI Path Filter Constraints

Recorded by run 1 for intent `261008-ci-path-filter` (since implemented in PR #14; kept for history):

1. Required checks must still report; a plain `paths` filter on `ci.yml` leaves them "Expected - waiting". A job skipped by `if:` does report success.
2. `paths` filters do not apply to tag pushes, so release filtering has no trigger-level meaning.
3. `docs/manual-checks/` is a release input (`release-preflight`).
4. `check-secrets` scans the whole repo; skipping CI on records-only PRs drops that scan unless a light scan stays (now `secrets.yml`).
5. `.github/workflows/**` and `Makefile` count as app paths. Classification: [code-structure.md](code-structure.md#top-level-path-classification).
6. Any new action must be SHA-pinned with top-level `permissions: contents: read`.
7. The filter must apply consistently to `pull_request` and `push` to `main`.

## Documentation

README covers build, install (upload via Settings > Plugins), connection, CI, release, marketplace. Go packages have `doc.go`; exported Go symbols and UI factories have doc comments citing FR/NFR/AC ids. `docs/manual-checks/` holds the first-release record. Makefile targets carry traceability IDs. The README does not describe SCM provider connection methods beyond tokens (not re-read in run 3).

## Intent Findings: 261008-fix-uiux-backlog

Recorded by run 2; released in v0.5.0 (PR #17) and not re-verified here.

| # | Request | Evidence at run 2 | Change shape proposed |
|---|---|---|---|
| 1 | Backlog issue on Home > Tasks rows, hover summary, click opens issue | Badge registered only for `task-card-tags`; `Link`/`LinkView` had no `Summary` | Register `IssueBadge` for `task-row-metadata`; add `Summary` to `Link` and `LinkView` |
| 2 | Hide Home > Integrations entry while OFF | `registerNavItem` unconditional; Kandev v0.96.0 has no nav `requires`, no unregister, no late registration | Needs a decision |
| 3 | Projects right below the sign-in method | Section order ended with `projects` | JSX reorder |
| 4 | Issue-watch empty message | Shared `watchesEmpty` = "No PR watches yet" | New message key |
| 5 | Remove per-list Add watch button | Empty state duplicated the header action | Drop the empty-state child |

## Known Issue: Plugin install 502

Status: addressed in v0.4.2 by dropping `windows-amd64` from the package. History (intent `261007-plugin-install-502`): Kandev `server.readTimeout` default 30 s (`KANDEV_SERVER_READTIMEOUT`); a multipart upload slower than 30 s is cut (400 `missing multipart field "package"`), which `tailscale serve` shows as 502. Remaining option: install by URL (see [architecture.md](architecture.md#improvement-opportunities)).

## Technical Debt

- SCM: only one credential method (`set_token`); `Settings` has no credential-source field; `ErrNoToken` wording assumes a pasted token (intent findings above).
- SCM: `ponytail:` markers in `internal/scm/store.go` (unbounded dismissed list / ledger) and `internal/scm/watcher.go` (single watcher worker) — unrelated to the current intent.
- `ci.yml` and `release.yml` duplicate setup and contract steps (run 1).
- `release-preflight` uses `gh release list --limit 1000` (marked `ponytail:` in the Makefile).
- Plugin-owned switch bus (`ui/src/switch/enabled-events.ts`) duplicates host state because Kandev v0.96.0 cannot expose the integration switch to plugins.
- Platform list declared in four places (see [code-structure.md](code-structure.md#build-and-packaging)).
- `pkgverify` duplicates Kandev `pkgtar` rules; no size limit check.
- Contract test installs over loopback (30 s client timeout), so it never sees proxy/slow-upload failures.
- Runtime drift: Kandev 0.97.0 running vs 0.96.0 pin.
- Dev environment: `../kandev` missing in fresh worktrees and Go not on `PATH` (run 3).
