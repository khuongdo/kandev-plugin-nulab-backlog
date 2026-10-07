# Code Quality Assessment — kandev-plugin-nulab-backlog

## Test Coverage and Baselines

| Area | Suite | Baseline |
|---|---|---|
| Go | co-located `*_test.go` in every `internal/*` package (95 files; `internal/git` 17 incl. `harness_test.go`, `leak_test.go`; `internal/connection` 20 incl. `git_credential_test.go`); fixtures in `internal/backlog/testdata/`; `make coverage` floor 80% over `./internal/... ./server/...` | **Not recorded at `5bf88b9`**: Go toolchain not installed, `../kandev` link absent. Last recorded (commit `2b4325f`, intent 261007-github-parity-actions): all 9 packages pass with `-race`, total 92.7% |
| UI | 31 `*.test.{ts,tsx}` files (`ui/src/git` 7, `ui/src/settings` 7), axe-core checks | **Not recorded**: `ui/node_modules` absent. Last recorded at `2b4325f`: 286/286, `tsc`/ESLint/Prettier clean |

Per the project Testing Posture correction, install Go 1.26.x and link `../kandev` to v0.96.0 before Construction so a baseline exists before code changes.

## Linting, CI/CD, Documentation

- Lint: `gofmt`, `go vet`, golangci-lint + `gosec`, `go mod tidy` check; UI ESLint, Prettier, `tsc --strict`; actionlint.
- CI: `ci.yml` runs Makefile targets plus the packaged-host contract job; `release.yml` with provenance.
- Docs: `doc.go` per Go package; doc comments cite requirement IDs; secret-bearing types hide secrets in `String`/`GoString`/`Format`.

## Technical Debt

- **Backlog-only source control** (main debt for this intent): single provider id, Backlog-keyed Git identity, Backlog-typed gateway and errors, Git scope and credential taken from the Backlog connection, single Git settings section. Details: [architecture.md](architecture.md#source-control-coupling-intent-261007-source-control-agnostic).
- **Backlog heuristics in Git**: `DefaultBranch` guesses from recent PR bases with fallback `"master"`; `Branches` lists only branches seen in the newest 100 PRs (`internal/git/service.go:222-290`).
- **Large files**: `internal/git/service.go`, `internal/issues/service.go`, `ui/src/settings/SettingsScreen.tsx`.
- **Known `ponytail:` ceilings** (from the previous store, not re-verified): process-wide rate-limit queue `internal/backlog/client.go`; store helpers copied between `internal/git/store.go` and `internal/issues/store.go`; unpruned PR watch ledger; first 50 repositories only in `ui/src/git/git-state.ts`.

## Intent 261007-source-control-agnostic Risks

1. **Host-reserved provider ids**: Kandev v0.96.0 reserves `github`, `gitlab`, `azure_devops` for its native integrations, and its GitHub credential resolver runs first. The plugin cannot register `github`; it must use plugin-scoped ids (e.g. `nulab-backlog-github`) or only reference Kandev's native GitHub repositories as PR links without owning clone/credentials. Requirements decision.
2. **Exclusive provider ownership**: a bare `bitbucket` id would block activation next to a separate Bitbucket plugin; plugin-prefixed ids avoid it.
3. **Mandated host allowlist vs. new providers**: project.md "ALWAYS accept only `https` space addresses under `backlog.com`, `backlog.jp`, or `backlogtool.com`" and the gateway's host pinning forbid calling GitHub/Bitbucket hosts as written. Needs an explicit rule decision (e.g. the mandate covers only the Backlog space address, with its own fixed allowlist per new provider) and a new stdlib-only client package, not `internal/backlog`.
4. **Git coupled to the Backlog connection**: every Git action and both credential RPCs read the Backlog connection, its selected projects and the Backlog on/off switch, and `ConnectionChanged` disables Git items. Requirements must decide whether a GitHub/Bitbucket link survives Backlog disconnect, space change, project deselect, or the switch being off.
5. **Backward compatibility of v0.3.0 data**: `git.links/watches/queries` (`schemaVersion: 1`), the `backlog.git.<ws>` secret, link keys `spaceHost|repositoryId|number`, the `nulab_backlog_pr` task metadata and `reviewKey` format exist on live installs with no provider field. Any new schema must read them as the Backlog Git provider (field defaulting when absent, or a migrating schema bump), with a regression test.
6. **Issue-to-PR traceability differs by provider**: `RelatedIssueKey` and the `IssueID` attachment on PR create are Backlog-native; for external providers the Backlog issue key can only go in PR text.
7. **Error mapping**: `git.errorCode`/`isKind` and `connection.Classify` read `*backlog.Error`; each new client needs its own error type mapped to the same `ActionError` codes, with redaction of tokens in errors and logs.
8. **Missing local test baseline**: see [Test Coverage and Baselines](#test-coverage-and-baselines); brownfield regressions cannot be detected until it is recorded.
9. **Locked behaviour to keep**: existing action keys unchanged; settings card/switch/nav independent of the switch; secrets never sent to the UI or logs.
