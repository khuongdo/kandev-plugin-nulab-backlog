# Code Summary — 261008-gh-cli-profile (zero-Unit, express)

Per-workspace gh CLI account for the GitHub source control connection. Methodology: TDD (team Testing Contract), Red → Green → Refactor per layer.

## Files

Modified:
- `internal/scm/cli_token.go` — login rule (`validLogin`), `CLIAccount`, `ListCLIAccounts`, login-aware `cliToken(ctx, p, login)` with a `{provider, login}` cache key, `ghToken` (`--user` read, then account-list classification and old-gh fallback), `ghAccounts` (`gh auth status --json hosts --hostname github.com`), `ghActive` (plain token + GitHub `/user`), `cliEnv` + `runCLI` env stripping, per-call output cap with `errCLIOutputCut`, `forgetCLI` dropping every login of a provider.
- `internal/scm/errors.go` — `ErrCLIAccountMissing`.
- `internal/scm/service.go` — `UseCLI(ctx, ws, p, login)`, `activeLogin`, `credential()` passes `AccountID`, Test keeps the chosen login (FR5.2), `ErrorCLIAccountMissing` code, `ProviderView.Login` (GitHub CLI only).
- `internal/scm/types.go` — `FieldLogin`.
- `internal/plugin/scm_actions.go` — `scm.providers.cli_accounts` action, `login` in the provider body, `cli_account_missing` classification.
- `internal/plugin/runtime.go` — `cli_account_missing` → HTTP 409.
- `manifest.yaml` — `scm.providers.cli_accounts` (workspace/admin, like `use_cli`); version 0.5.2 → 0.5.3.
- `ui/src/settings/source-control-section.tsx` — inline gh account picker (host `Select`), Change account, `@login` account line, account-missing alert, failures as `role="alert"`, worktree note.
- `ui/src/git/git-state.ts` — `ProviderView.login`, `CliAccount`, `cli_account_missing` notice for other screens.
- `ui/src/messages/en.ts` — new strings.
- `README.md` — 0.5.3 upgrade note; CLI section: per-workspace account, gh versions, env stripping, worktree note (AC4.1.2).
- Tests: `internal/scm/cli_token_test.go`, `internal/scm/harness_test.go`, `internal/scm/service_test.go`, `internal/plugin/actions_scm_test.go`, `ui/src/settings/source-control-section.test.tsx`, `ui/src/git/git-state.test.ts`.

Created:
- `internal/scm/service_cli_account_test.go`
- `internal/scm/testdata/gh-auth-status-two.json`, `internal/scm/testdata/gh-auth-status-error-state.json`
- `internal/scm/testdata/v052-settings-cli.json` (frozen v0.5.2 gh CLI settings record; never regenerate)

## Key Decisions

- The chosen login is the existing `Settings.AccountID`; no new stored field, no migration. `UseCLI` stores GitHub's canonical `user.ID` after checking it equals the requested login (case-insensitive).
- Confirmed gh versions: `gh auth status --json` was added in **gh 2.81.0** (cli/cli PR #11544, released 2025-10-01); `gh auth token --user` exists since gh 2.40 (multi-account). Local check with gh 2.97.0 confirmed the JSON shape (`hosts["github.com"][]` with `login`, `active`, `state`).
- Token read for login L: `gh auth token --hostname github.com --user L`; on failure the account list decides: list fails → `cli_unavailable`; L unlisted → `cli_account_missing`; L listed → plain `gh auth token --hostname github.com`, used only when GitHub `/user` says L.
- Account list decodes stdout even on a non-zero exit, keeps `state == "success"` entries with a valid login on `github.com` only, and returns only `login` + `active`. A cut-off output (32 KiB cap, this call only; token reads keep 4 KiB) is `cli_unavailable`. Undecodable output (gh < 2.81.0 or no gh) falls back to the active account via plain token + `/user`; if that fails too, `cli_unavailable`.
- `runCLI` removes `GH_TOKEN`, `GITHUB_TOKEN`, `GH_ENTERPRISE_TOKEN`, `GITHUB_ENTERPRISE_TOKEN` (name match without case, like Kandev's reference) from the child environment.
- `CLIRunner` gained a `limit int` parameter so the output cap is per call.
- `use_cli` for GitHub with no login resolves the active account through the list; any login for GitLab, or an invalid GitHub login, is a `login` field error before any CLI run.
- Test in gh CLI mode never rewrites `AccountID`; it refreshes `Account` only for the same login, else records `cli_account_missing`.
- `cli_account_missing` maps to HTTP 409; the UI builds "<login> is not logged in to gh on the Kandev server — log in again or pick another account" from the code plus the login (picked or stored). A gh record with no login shows "Pick the gh account this workspace uses."
- UI: the display name is shown after `@login` only when it differs from the login exactly.

## Red → Green Log

| Layer | Red command | Red result | Green |
|---|---|---|---|
| CLI (`internal/scm`) | `go test -race -count=1 ./internal/scm/` | build failed: `CLIRunner` signature, undefined `validLogin`, `errCLIOutputCut`, `ListCLIAccounts`, `cliToken` arity | all CLI tests pass |
| Service (`internal/scm`) | `go test -race -count=1 ./internal/scm/` | build failed: `UseCLI` arity, undefined `FieldLogin`, `ProviderView.Login`, `ErrorCLIAccountMissing` | `ok internal/scm` |
| Action (`internal/plugin`) | `go test -race -count=1 ./internal/plugin/` | build failed: undefined `actionSCMCLIAccounts`, `codeCLIAccountMissing`; `UseCLI` arity | `ok internal/plugin` |
| UI | `npx vitest run src/settings/source-control-section.test.tsx src/git/git-state.test.ts` | 11 failed, 24 passed | 35 passed |

Refactor steps: no structural change needed beyond the Green code; the `ponytail:` lock note in `cli_token.go` was updated (the lock now also covers the old-gh `/user` check).

## Test Results

- `make coverage` (profile under `build/`): all packages `ok` with `-race`; total **92.9%** (floor 80%, only `server/main.go` excluded). `internal/scm` 93.7%, `internal/plugin` 93.8%.
- Go test results: **1433 passed, 0 failed** (baseline 1383).
- Concurrency tests re-run 5 times with `-race`: green.
- `golangci-lint v2.14.0` (pinned): 0 issues. `gofmt -l`: clean. `go vet`: clean. `go run ./cmd/ci secrets -root .`: OK.
- UI: `tsc --noEmit` clean, `eslint src` clean, `prettier --check .` clean, Vitest **435 passed** (34 files).

## Deviations

- "Go-http" checks: `internal/scm` cannot import `internal/github` (cycle), so the per-workspace account proof uses the in-package fake `Client`, which maps each token to its GitHub login and records the workspace (context tag) and login of every call. The `Authorization: Bearer <token>` mapping of the real GitHub client stays covered by the existing `internal/github` tests.
- Env stripping (AC1.1.6, AC2.1.8) is tested on the real `runCLI` with a helper subprocess that prints its environment, instead of a fake runner recording env (the env is built inside `runCLI`).
- Action tests live in the existing `internal/plugin/actions_scm_test.go` (the plan named `scm_actions_test.go` / `manifest_test.go`); the manifest declaration test there (`TestSCM_Manifest_Actions`) now expects 24 actions.
- Version bumped to 0.5.3 (patch, per the project's convention for small additions); Deployment Pipeline still confirms the release version.
- Workspace setup created a `../kandev` symlink to `~/repo/kandev` (v0.96.0, commit matches `.kandev-sdk-ref`), outside the repo.
