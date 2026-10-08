# Unit Test Instructions — 261008-gh-cli-profile

## Framework Setup

- Go 1.26.x with CGO (for `-race`), `testing` + `github.com/stretchr/testify/require`, table-driven `t.Run` tests, `httptest` fake GitHub server.
- `go.mod` replaces `github.com/kandev/kandev => ../kandev/apps/backend`. Before running, `../kandev` must point at the pinned v0.96.0 checkout (`~/repo/kandev`, commit in `.kandev-sdk-ref`), e.g. a symlink next to the repo. Never edit `go.mod` for this.
- UI: Node from `.nvmrc`, `npm ci` in `ui/`, Vitest (`ui/vitest.config.ts`), existing harness `ui/src/testing/harness.ts`.

## Run This Change's Tests

Scoped to the packages and files this change touches (run from the repo root):

```bash
go test -race -count=1 ./internal/scm/ ./internal/plugin/
```

```bash
cd ui && npx vitest run src/settings/source-control-section.test.tsx
```

Static checks for the touched code:

```bash
gofmt -l internal/scm internal/plugin
go vet ./internal/scm/ ./internal/plugin/
cd ui && npx tsc --noEmit && npx eslint src/settings
```

Both test commands are runnable before the first Red step (existing tests in those packages/files are green at baseline).

## Expected Tests (Minimal strategy, requirement-driven)

About 12-15 test functions (table rows count as cases, not separate tests), at least one per requirement and one happy path per component:

- `internal/scm` CLI layer: login validation table; env stripping; account list parse (active, error-state, non-zero exit, cut-off, field allowlist); `--user` token + per-login cache with injected clock; old-gh fallback accept/refuse; unavailable vs account-missing; no-`--json` fallback.
- `internal/scm` service layer: UseCLI with/without login, unknown login, GitLab login rejected; call-site table (Test, repos, PR list, link, mine filter, watch poll) with concurrent W1/W2; Test guard; v0.5.2 fixture; missing `AccountID`; redaction table.
- `internal/plugin`: `scm.providers.cli_accounts` action; `use_cli` login field; error code passthrough; manifest declares the action.
- UI: picker flow, single-account connect, login display, Change account + Cancel, account-missing alert, `role="alert"` on failures, no-`--json` single option, worktree note on/off.

## Coverage Target

- Go: 80% line coverage floor over `./internal/...` and `./server/...` (`make coverage`, run in Build and Test; write the profile under `build/` or delete `coverage.out` afterwards — never leave it in the repo root).
- Existing suite stays green (baseline 1383 Go test results).

## Mocking / Stubbing

- `scm.CLIRunner` fake: records argv and env per call, returns scripted stdout / error / timeout; fails the test on any argv outside the allowlist (`auth status --json hosts --hostname github.com`, `auth token --hostname github.com --user <login>`, `auth token --hostname github.com`; glab's existing command for GitLab rows).
- Injected clock (`Service.Now`) for the 5-minute cache TTL; no `time.Sleep`.
- `httptest` GitHub server keyed by `Authorization` token → login, to prove which account each request used.
- UI: the existing host/action harness; stub `scm.providers.cli_accounts` and `scm.providers.use_cli` responses.

## Test Data

- `internal/scm/testdata/`: sample `gh auth status --json hosts` outputs (two accounts, error-state account, oversized), and a frozen v0.5.2 settings JSON fixture in gh CLI mode with `accountId: "alice"` — never regenerate it.
- Token values in tests are sentinels (e.g. `gho_SENTINEL_alice`) so redaction tests can assert absence.
