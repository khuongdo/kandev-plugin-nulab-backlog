# Unit Test Instructions — Multi-provider source control

## Framework Setup

- Go 1.26.x with CGO enabled (needed by `-race`); `../kandev` must point to the Kandev v0.96.0 checkout (`.kandev-sdk-ref`), because `go.mod` replaces `github.com/kandev/kandev` with `../kandev/apps/backend` and the UI `tsconfig` aliases `@kandev/plugin-sdk` into it.
- Go tests: `testing` + `github.com/stretchr/testify/require` (already in `go.mod`), table-driven with `t.Run`, behaviour-describing names.
- UI tests: Vitest + jsdom + axe-core from `ui/package.json`; run `npm ci` in `ui/` once.
- No new test configuration files are needed; the existing `Makefile`, `ui/vitest` configuration and `.golangci.yml` are reused.

## How to Run This Work's Tests

Go (from the repository root):

```bash
go test -race -count=1 ./internal/scm/... ./internal/github/... ./internal/gitlab/... ./internal/bitbucket/...
go test -race -count=1 -run 'SCM|Scm|Manifest|V030' ./internal/plugin/...
```

UI (from `ui/`):

```bash
npx vitest run src/settings/source-control-section.test.tsx src/git/pr-list.test.tsx src/git/git-state.test.ts src/git/watch-form.test.tsx src/issues/issue-panel.test.tsx
```

The first Go command is runnable before the first Red step (it reports `no test files` until Step 3 adds them). Every Red step records the failing output of the matching command above.

## Coverage Targets

- Go: the team floor of 80% line coverage over `./internal/...` and `./server/...` (`make coverage`); new packages are expected at or above 85% on their own. Write the profile under `build/` (never the repository root) and delete it after the run.
- One verifiable test per requirement (FR1.1-FR7.2, NFR1-NFR8), plus one happy-path test per new component (`scm` store, service, watcher, HTTP helper, each client, the plugin action group, each new/changed UI component).

## Mocking and Stubbing

- External services: `net/http/httptest` fake servers per client, serving JSON fixtures from `internal/<pkg>/testdata/`; clients get the fake's URL through a test-only `BaseURL` field. Include 401, 403, 404, 429 with `Retry-After`, and GitHub 403 with `X-RateLimit-Remaining: 0`.
- Service tests use a fake `scm.Client` and the existing in-memory host state/secret fakes pattern from `internal/git` (`fakes_test.go`).
- Time: inject the clock and wait functions; no real `time.Sleep`.
- Logs: a capturing `slog.Handler` asserts no token appears in log records or error strings (NFR1).
- UI: stub `host.actions.call` / `host.ui` the same way the existing `ui/src/git` and `ui/src/settings` tests do.

## Test Data Management

- Only fake tokens such as `ghp_FAKE_TOKEN_FOR_TESTS`, `glpat-FAKE`, `ATBB-FAKE`; never real credentials (project Forbidden rule).
- v0.3.0 regression fixtures (`git.links`, `git.watches`, `git.queries`, `backlog.git.<ws>`, `nulab_backlog_pr`) live in `internal/plugin/testdata/v030/` and are copied from the shapes used by existing `internal/git` tests.
- Tests use `t.TempDir()` / `t.Setenv()` and share no state.
