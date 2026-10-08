# Code Quality Assessment — kandev-plugin-nulab-backlog

## Test Coverage and Baselines

- Go tests co-located in every `internal/*` package; `httptest` fakes with `testdata/` fixtures; SCM tests use `newHarness` with fake `Client`, secrets, state, clock and `CLIRunner` (no real `gh` runs).
- UI tests: Vitest under `ui/src/**/*.test.ts(x)` with the shared fake host `ui/src/testing/harness.ts`.
- `make coverage`: 80% line floor over `./internal/... ./server/...`.
- **Baseline (this run, 2026-10-08, commit `ca8146c`)**: `go test -count=1 ./...` passed — 13 packages with tests, 1383 passing results (tests plus subtests), 0 failures, 0 skips. Go 1.26.8 with a scratch `-modfile` pointing the SDK at `~/repo/kandev/apps/backend` (v0.96.0) because `../kandev` is missing; no repo file changed. Not run with `-race`; Vitest not run.

## Linting

gofmt, go vet, golangci-lint with gosec (`.golangci.yml`), `tsc --noEmit` strict, ESLint, Prettier, actionlint plus `cmd/ci workflows` policy. Conventions: `//nolint:gosec // G204` on `exec.CommandContext` with fixed args; `G101` on credential-like action keys.

## CI/CD

`ci.yml` (`checks`, `packaged-host-contract`), `secrets.yml`, `release.yml` (`verify` -> `contract` -> `publish` with attestation). Required checks on `main` via ruleset; squash only.

## Documentation

README present; `doc.go` per package; doc comments cite FR/NFR/AC ids.

## Intent Findings: 261008-gh-cli-profile

Request: per-workspace choice of gh account for the GitHub connection, and `gh` in a task worktree from a Backlog task uses that account.

| # | Area | Evidence | Change shape |
|---|---|---|---|
| 1 | CLI command | `cliCommand` (`internal/scm/cli_token.go:28-36`) is fixed to `gh auth token --hostname github.com`, no `--user` | Take the login: `gh auth token --hostname github.com --user <login>`; keep fixed-arg, no-shell exec. Fallback for gh without `--user`: accept only the active login, else `cli_unavailable` (Kandev pattern) |
| 2 | Account list | Nothing lists gh accounts | Run `gh auth status --json hosts` (bounded output, never echoed) and return logins + active flag; new admin action (e.g. `scm.providers.cli_accounts`) in `manifest.yaml` and `scm_actions.go` (parity test) |
| 3 | Stored choice | `scm.Settings` (`store.go:42-50`) has `Source`, `Account`, `AccountID` but no chosen login | Add an `omitempty` login field (or reuse `AccountID` as the chosen login); old documents stay readable. Migration default: existing CLI records keep their current `AccountID` |
| 4 | Cache | `cliCache` keyed by provider only, "not per workspace" (`cli_token.go:59-63`) | Key by provider + login; `forgetCLI` likewise |
| 5 | Connect / test | `UseCLI` (`service.go:250-271`) and `credential` (`:276-298`) use whoever is active; `Test` (`:310-335`) rewrites `Account`/`AccountID` on `gh auth switch` | Pass the chosen login through `UseCLI` (body gains `login`), `credential` and `Test`; a test must not silently change the identity |
| 6 | UI | GitHub card shows one "Use gh CLI login" button (`ui/src/settings/source-control-section.tsx:233,302`) | Account picker next to it; `ProviderView` may expose the chosen login (non-secret) |
| 7 | Worktree gh | Tasks come from Kandev's `TaskCreateDialog` (`ui/src/page/start-task.tsx:100-115`) or `Tasks().Create` without `Repositories`/`Launch` (`host_port.go:120-130`, `scm_actions.go:268-275`); Kandev's executor sets worktree `GH_TOKEN` | Not reachable by plugin code with pluginsdk v0.96.0. Options for Requirements: (a) document that Kandev's own GitHub integration must use the same gh account; (b) set `Launch.ExecutorProfileID` to a profile whose env carries the right token — only for tasks the plugin creates itself, not the dialog path; (c) record as a limitation / Kandev feature request |

Constraints:

- **Secrets (NFR1, project rule)**: never log or return `gh` stdout/stderr (`gh auth status` can print token sources); keep redaction, 10 s timeout, 4 KiB cap; strip `GH_TOKEN` / `GITHUB_TOKEN` from the child env if the stored login must win over an env override (Kandev does).
- **Scope**: GitLab `glab` has the same single-account behaviour; the intent names only gh, so scope it explicitly.
- **Compatibility**: everything used must exist in Kandev 0.96.0 (`min_kandev_version`).
- **Test environment**: link `../kandev` to v0.96.0 or reuse the scratch `-modfile` approach; run `-race` in Code Generation.

## Technical Debt

- `internal/scm/cli_token.go:59-63` — `ponytail:` one mutex for every provider with the CLI run under it.
- `internal/plugin/host_port.go:139` — `workflowRefused` matches gRPC error text (`ponytail:`).
- `internal/scm/store.go`, `watcher.go` — `ponytail:` unbounded dismissed list / ledger, single watcher worker.
- `internal/github/client.go` — `SearchRepos` reads only the 100 most recent repos (`ponytail:`).
- Platform list declared in four places (see [code-structure.md](code-structure.md#build-and-packaging)).
- Dev environment: `../kandev` missing in fresh worktrees.
