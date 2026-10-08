# Code Summary — CLI login for GitHub (gh) and GitLab (glab)

Zero-Unit express run. TDD per the Testing Contract (`sha256:6db87aaf…`). Brownfield: every file modified in place.

## Files

Created:
- `internal/scm/cli_token.go` — `CLIRunner`, `cliCommand` (fixed argument lists), `runCLI` (no shell, 10 s timeout, 4 KiB stdout cap, stderr discarded), the 5-minute per-provider cache (`cliToken`, `forgetCLI`).
- `internal/scm/cli_token_test.go` — fake runner; commands, trimming, every failure becomes `ErrCLIUnavailable`, caller cancellation, Bitbucket refused, cache and expiry on the injected clock, forget, failures not cached, default runner output cap and missing binary (helper-process test, no real `gh`/`glab`).

Modified:
- `internal/scm/store.go` — `Settings.Source` (`json:"source,omitempty"`).
- `internal/scm/errors.go` — `ErrCLIUnavailable`.
- `internal/scm/service.go` — `MethodToken`/`MethodCLI`, `ProviderView.Method`, `ErrorCLIUnavailable`, `Service.CLI` + cache, `UseCLI`, `credential()` branch on `Source`, `SetToken`/`RemoveToken` reset the source and clear the cache, `Test` forgets the cache and records `cli_unavailable` as a result, `errorCode` maps `ErrCLIUnavailable`, `validToken` helper shared by typed and CLI tokens.
- `internal/scm/watcher.go` — `refreshProvider` forgets the cached CLI token on 401.
- `internal/scm/service_test.go`, `internal/scm/store_test.go` — service and data-model tests (and the existing exact-view assertion now includes `Method: token`).
- `internal/gitlab/client.go`, `internal/gitlab/client_test.go` — `Authorization: Bearer` instead of `PRIVATE-TOKEN`.
- `internal/plugin/scm_actions.go` — `scm.providers.use_cli` handler, `codeCLIUnavailable`, `classifySCM` mapping.
- `internal/plugin/runtime.go` — `statusFor`: `cli_unavailable` → 503.
- `internal/plugin/actions_scm_test.go` — action, classify and manifest tests (23 scm actions).
- `manifest.yaml` — `scm.providers.use_cli` (`workspace`, `admin`, 16384).
- `ui/src/git/git-state.ts` (+ test) — `ProviderView.method`, `scmNotice` handles `cli_unavailable`.
- `ui/src/messages/en.ts` — `scmAccountCli`, `scmUseCli`, `scmCliConnected`, `scmCliUnavailable` (the only message catalogue).
- `ui/src/settings/source-control-section.tsx` (+ test) — "Use gh/glab CLI login" button (`backlog-scm-<provider>-use-cli`), "Connected via … CLI as …" line, CLI error text from the action and from Test.
- `README.md` — "Connect GitHub or GitLab with the CLI login" section.

## Key Decisions

- The CLI cache is per provider, not per workspace: the CLI login belongs to the Kandev server. One mutex guards the cache and the CLI run, so at most one CLI process runs at a time (`ponytail:` comment names the ceiling).
- `UseCLI` forgets the cache first, so pressing the button always reads the current login; a refused CLI token (e.g. 401) is forgotten and nothing is stored.
- `SetToken` and `RemoveToken` write `Source = ""` (not `"token"`), so stored settings keep the v0.5.0 shape; the view reports `method: token` for any connected non-CLI provider (FR6.1).
- `credential()` reads the settings first; a CLI provider never touches the secret store (FR3.1). Token providers pay one extra state read per call.
- `Test` treats `ErrCLIUnavailable` as a result (`lastError: cli_unavailable`, state `error`), like a refused token, and clears it on the next good run (FR4.2, FR4.3).
- `cli_unavailable` is a new action error code with HTTP 503; the UI shows "The gh CLI is not available or not logged in on the Kandev server." (glab on the GitLab card; "gh / glab" outside the card).
- glab command: `glab config get token --host gitlab.com` (works on every glab version that stores a login; `GITLAB_TOKEN`-only logins are not seen — documented). gh needs ≥ 2.17 for `gh auth token`.

## Test Results

| Suite | Baseline (before changes) | After |
|---|---|---|
| `go test -race ./internal/scm/... ./internal/gitlab/... ./internal/plugin/...` (top-level tests) | 194 passed | 214 passed |
| `go test -race ./...` (via `make test`) | all packages ok | all packages ok |
| Vitest, whole UI (`cd ui && npx vitest run`) | 33 files, 387 passed | 33 files, 391 passed |
| `src/settings/source-control-section.test.tsx` | 12 passed | 16 passed |

`make check-format vet lint test` is clean (gofmt, go vet, golangci-lint with gosec, Prettier, ESLint, `tsc --noEmit`, actionlint). `go mod tidy` changes nothing.

## Red Evidence

Step 3 — `go test -race ./internal/scm/ -run 'TestSettings_'`
```
store_test.go:146:34: v.Method undefined (type ProviderView has no field or method Method)
store_test.go:153:39: unknown field Source in struct literal of type Settings
store_test.go:153:47: undefined: MethodCLI
FAIL internal/scm [build failed]
```

Step 6 — `go test -race ./internal/scm/ -run 'CLI'`
```
cli_token_test.go:84:22: h.svc.cliToken undefined (type *Service has no field or method cliToken)
cli_token_test.go:85:28: undefined: ErrCLIUnavailable
cli_token_test.go:124:20: undefined: cliTTL
FAIL internal/scm [build failed]
```

Step 9 — `go test -race ./internal/scm/`
```
service_test.go:382:18: h.svc.UseCLI undefined (type *Service has no field or method UseCLI)
service_test.go:450:19: undefined: ErrorCLIUnavailable
FAIL internal/scm [build failed]
```

Step 12 — `go test -race ./internal/gitlab/`
```
--- FAIL: TestCurrentUser  client_test.go:61: Not equal: expected: "Bearer test-token-<generated>"  actual: ""
FAIL internal/gitlab
```

Step 15 — `go test -race ./internal/plugin/`
```
actions_scm_test.go:347:26: undefined: actionSCMUseCLI
actions_scm_test.go:422:19: undefined: codeCLIUnavailable
FAIL internal/plugin [build failed]
```

Step 18 — `cd ui && npx vitest run src/settings/source-control-section.test.tsx`
```
× offers the CLI login on GitHub and GitLab only, to admins only (FR1.1, FR1.4, FR2.1, FR2.3)
× connects with the CLI login and shows the method (FR1.2, FR5.2)
× explains an unusable CLI, from the action and from Test (FR4.2)
AssertionError: expected 'Connected as Lan' to be 'Connected via gh CLI as Lan'
Tests  3 failed | 13 passed (16)
```

## Coverage

`make coverage`: **92.9 %** total (floor 80 %, exclusion `server/main.go` only, unchanged). `internal/scm` 93.7 %, `internal/gitlab` 88.7 %, `internal/plugin` 93.8 %. The profile was written under `build/` and deleted afterwards; no `coverage.out` in the repository root.

## Deviations From the Plan

- The approved UI command `npm --prefix ui exec -- vitest run src/settings/source-control-section.test.tsx` run from the repository root does not load `ui/vitest.config.ts` (jsdom), so every test fails with `document is not defined` even at baseline. The same test file was run from `ui/` (`npx vitest run …`), which is what `make test` does; the instructions file was not edited.
- `internal/plugin/runtime.go` (not listed in the plan) gained one `statusFor` case so `cli_unavailable` answers 503 instead of the 500 default.
- `ui/src/git/git-state.test.ts` (not listed) gained one assertion for `scmNotice` with `cli_unavailable`.
- Plan Step 1: `../kandev` did not exist in this worktree; created the symlink `../kandev -> ~/repo/kandev` (v0.96.0), outside the repository.
- The packaged-host contract test (NFR5, second half) was not run in this stage; it belongs to Build and Test.
