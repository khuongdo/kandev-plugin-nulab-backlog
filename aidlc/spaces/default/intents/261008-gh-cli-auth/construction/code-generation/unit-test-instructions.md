# Unit Test Instructions — CLI login for GitHub and GitLab

## Framework and setup

- Go: `testing` + `github.com/stretchr/testify/require`, table-driven tests with `t.Run`, always `-race`. Already configured; no new dependency.
- UI: Vitest (`ui/vitest.config.ts`), already configured.
- Prerequisites (one time, nothing in the repo changes):
  - Go on PATH: `export PATH="$HOME/.local/go/bin:$PATH"`.
  - `../kandev` (next to the repo) is a link to the pinned v0.96.0 checkout: `ln -s ~/repo/kandev ../kandev` if it is missing.
  - `ui/node_modules` installed: `npm ci --prefix ui` if missing.

## Commands for this change only (runnable before the first Red step)

```bash
# Go: the packages this change touches
go test -race ./internal/scm/... ./internal/gitlab/... ./internal/plugin/...

# Go: one test file's tests while iterating, e.g. the CLI token source
go test -race ./internal/scm/ -run 'CLI'

# UI: the source-control settings card only (run from ui/ so ui/vitest.config.ts is loaded)
cd ui && npx vitest run src/settings/source-control-section.test.tsx src/git/git-state.test.ts
```

Do not pass `-coverprofile` here. Coverage is checked once in plan Step 21 with `make coverage`, and the profile is deleted afterwards (project rule).

## Tests to write (Minimal strategy: one per requirement, happy path per component)

About 15 tests in total:
- `internal/scm/cli_token_test.go` (new): command and arguments per provider, trimming, every failure becomes `ErrCLIUnavailable` without stderr or token text, Bitbucket refused, 5-minute cache on an injected clock, forget.
- `internal/scm/service_test.go`: `UseCLI` happy path (GitHub, GitLab), Bitbucket refused, CLI failure and 401 change nothing, `credential()` uses the CLI not the secret store, `SetToken` and `RemoveToken` switch back and clear the cache, `Test` records and clears `cli_unavailable`, new token after cache expiry, watcher forgets on 401, v0.5.0 settings read as `method: token`, leak test.
- `internal/gitlab/client_test.go`: `Authorization: Bearer` header.
- `internal/plugin` tests: `scm.providers.use_cli` handler, `classifySCM` → `cli_unavailable`, manifest/runtime parity (existing test).
- `ui/src/settings/source-control-section.test.tsx`: button on GitHub/GitLab only and only for admins, action call, "Connected via gh CLI as …" line, CLI error text.

## Mocking and test data

- Never run a real `gh` or `glab`: the service takes a `CLIRunner` function; tests pass a fake that records name and arguments and returns canned stdout or an error (`exec.ErrNotFound`, an exit error, `context.DeadlineExceeded`).
- Providers are faked with the existing `httptest` servers / fake clients in `internal/scm` and `internal/gitlab`; the clock is `Service.Now`.
- Tokens in tests are obvious fakes (e.g. `gho_fake_cli_token`); never real credentials.

## Coverage target

80% line coverage for `./internal/...` and `./server/...` stays the floor (`make coverage`); it is not lowered.
