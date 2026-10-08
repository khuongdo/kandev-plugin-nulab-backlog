## Developer Code Scan Results

Intent: add a second way to link GitHub as a source-control provider: use the GitHub CLI's existing login on the host (`gh auth token`) instead of pasting a personal access token. Focused scan, depth Minimal. Commit at scan time: `d3d17e5` (branch `feature/th-m-auth-method-cho-0pe`).

### Scan Coverage
- **Analyzed deeply**:
  - internal/scm/ (client.go, types.go, errors.go, service.go, store.go, httpx.go, watcher.go credential call sites; harness_test.go and fakes_test.go test patterns)
  - internal/github/ (client.go, doc.go, client_test.go, testdata/)
  - internal/plugin/scm_actions.go
  - internal/plugin/runtime.go
  - internal/plugin/credential.go
  - internal/plugin/manifest_test.go (action/handler parity tests only)
  - ui/src/settings/source-control-section.tsx
  - manifest.yaml (runtime, capabilities, scm.* action entries)
- **Skimmed only**:
  - internal/connection/ (doc.go, git_credential.go header, outcome codes in service.go): Backlog connection (API key / OAuth) and Backlog Git credential; supplies `FieldError`, `ErrStore`, `Code*` outcome codes reused by scm. Not the GitHub link.
  - internal/git/ (doc.go): Backlog Git integration (U4). Not GitHub; unaffected by this intent.
  - ui/src/settings/source-control-section.test.tsx (test names only)
  - ui/src/git/git-state.ts (`ProviderView`, `scmNotice`, `usableProviders`) - outside snapshot, shallow
  - ui/src/messages/en.ts (scm* strings) - outside snapshot, shallow
  - internal/redact/ (exported functions only) - outside snapshot, shallow
  - internal/ci/changes.go (only existing `os/exec` use) - outside snapshot, shallow
  - Makefile coverage target - outside snapshot, shallow
  - ~/repo/kandev (v0.96.0, commit f099a46, = `.kandev-sdk-ref`): apps/backend/pkg/pluginsdk/host.go method list, apps/backend/internal/plugins/runtime/manager.go plugin spawn, apps/backend/internal/github/gh_accounts.go `gh auth token` usage - outside the repo, shallow reference only

### Packages Found
- internal/scm - Go library - Go - provider-neutral source control service for GitHub/GitLab/Bitbucket: settings, token storage, mappings, PR lists, links, queries, watches, background watcher. Owns the single credential read path.
- internal/github - Go library - Go - read-only GitHub REST client (`scm.Client` implementation) for `https://api.github.com`, Bearer token auth.
- internal/gitlab, internal/bitbucket - Go library - Go - sibling `scm.Client` implementations (not read; same pattern per scm/doc.go).
- internal/plugin - Go adapter - Go - the only package importing `pluginsdk`; action routing (`handlers` map), error-to-code mapping (`classify`, `classifySCM`), Host adapters (`hostStores`, `hostPort`), wiring (`wireSCM`, `scmClients`).
- internal/connection - Go library - Go - Backlog space connection; shared error/outcome types.
- internal/git - Go library - Go - Backlog Git (repository provider, PRs, credential leases).
- internal/redact - Go library - Go - context-scoped secret masking for logs/errors.
- ui/src/settings - TypeScript UI - TSX (React via host) - settings screen; `createSourceControlSection` renders the provider cards.
- server - Go binary entry (`pluginsdk.Serve`).

### Build System
- **Type**: Go modules (go 1.26.0) + npm/Vite UI; orchestrated by `Makefile`
- **Config Files**: go.mod (`replace github.com/kandev/kandev => ../kandev/apps/backend`), .kandev-sdk-ref (f099a46 = Kandev v0.96.0), Makefile (`check-format vet lint test coverage build package verify-package`, `COVERAGE_MIN := 80`), .golangci.yml, manifest.yaml, .github/workflows/{ci,release,secrets}.yml
- **Build Dependencies**: plugin -> scm -> connection, redact; github -> scm; scm never imports pluginsdk (scm/doc.go); github imports only scm + stdlib.

### APIs Discovered
- Plugin actions (browser -> plugin, manifest.yaml lines 270-360) - 22 `scm.*` keys. Auth-relevant: `scm.providers.list` (authenticated, unguarded), `scm.providers.set_token` / `scm.providers.test` / `scm.providers.remove` / `scm.repos.search` (admin). Body cap 16384.
- `scm.Client` interface (internal/scm/client.go:83) - 5 methods, each takes `cred scm.Credential` per call: `CurrentUser`, `SearchRepos`, `GetRepo`, `ListPRs`, `GetPR`.
- GitHub REST (internal/github/client.go) - GET `/user`, `/user/repos`, `/repos/{o}/{r}`, `/repos/{o}/{r}/pulls`, `/repos/{o}/{r}/pulls/{n}`.
- Kandev Host (pluginsdk v0.96.0, shallow) - state, config, secrets (`GetSecret/SetSecret/DeleteSecret/RevealSecret`), events, tasks, sessions, workspaces, workflows, repositories, messages, utility agent. **No host method exposes Kandev's own GitHub credential or a process-exec API.**

### Frameworks & Libraries
- Go stdlib `net/http` (no third-party HTTP client, team rule) - provider calls via `scm.API`
- github.com/kandev/kandev/pkg/pluginsdk - pinned v0.96.0 via local replace - plugin protocol (hashicorp go-plugin gRPC)
- github.com/stretchr/testify/require - tests
- `@kandev/plugin-sdk` (host React/jsx, `host.api.invokeAction`) - UI; Vitest + axe for UI tests

### Test Coverage
- **Test Directories**: internal/scm/*_test.go (fake `Client`, `fakeSecrets`, `fakeState`, `newHarness`), internal/github/client_test.go + testdata/ (httptest fake GitHub with user/repo/pull JSON and 401/403-rate-limit/404/429 fixtures), internal/plugin/actions_scm_test.go, manifest_test.go, ui/src/settings/source-control-section.test.tsx
- **Test Frameworks**: Go `testing` + testify (`-race`), Vitest
- **Coverage Config**: present (Makefile `coverage`, 80% floor over `./internal/...` and `./server/...`, only `server/main.go` excluded)
- **Baseline (recorded 2026-10-08)**: `go test -race ./internal/scm/... ./internal/github/... ./internal/connection/...` -> **407 passed in 3 packages**; additionally `./internal/plugin/... ./internal/git/...` -> **302 passed in 2 packages**. Go 1.26.8 from `~/.local/go/bin` (not on PATH). `../kandev` does not exist in this worktree; tests ran through a temporary `go.work` in the session scratchpad replacing the SDK with `~/repo/kandev/apps/backend` (v0.96.0, matches `.kandev-sdk-ref`). No `-coverprofile`; working tree unchanged.

### Code Quality Indicators
- **Linting**: golangci-lint (default + gosec) via `.golangci.yml`; `//nolint:gosec // G204` convention already used for `exec.CommandContext` in internal/ci/changes.go:41; tsc strict + ESLint + Prettier for UI
- **CI/CD**: .github/workflows/ci.yml, release.yml, secrets.yml (credential scan)
- **Documentation**: package doc comments present and reference FR/NFR ids; exported names documented

### Technical Debt Signals
- `scm.TokenInput` / `set_token` is the only way to configure a provider; `Settings` has no notion of credential source (internal/scm/store.go:41-48) - the new method needs a field here (schemaVersion 1 document; adding an `omitempty` field keeps old documents readable since `load` only checks `schemaVersion`).
- `ProviderView.State` is derived purely from `HasToken` + `LastError` (service.go `view`, from line 110); the UI's `hasToken` is `state !== "not_configured"` (source-control-section.tsx). A gh-backed provider must still report `HasToken`-equivalent truthiness or the cards/PR lists (`usableProviders` filters `state === "connected"`) hide it.
- Error text `ErrNoToken` (scm/errors.go) says "add one under Source control" - wording assumes PAT only.
- `ponytail:` markers in scm/store.go and watcher.go (unbounded dismissed list/ledger, single watcher worker) - unrelated to this intent.

## Handoff Summary
- **Intent-relevant finding**: Every GitHub API call obtains its token through ONE private function, `(*scm.Service).credential(ctx, ws, p)` (internal/scm/service.go:221-237), which reads secret `backlog.scm.<provider>.<workspace>` as JSON `scm.Credential{Token, Username}` and wraps ctx with `redact.WithSecrets`. Its 7 callers: `Test` (service.go:242), `SearchRepos` (:294), `checkRepos`/`SetMapping` (:350), `ListPRs` (prs.go:53), `Link` (links.go:31), `runWatch` (watcher.go:50) and the background `refreshProvider` (watcher.go:170). The GitHub client is credential-agnostic: `github.New()` sets `Authorization: Bearer <cred.Token>` (internal/github/client.go:20-23), so a token from `gh auth token` works unchanged. Adding the gh CLI method therefore reduces to (a) a new credential source chosen inside `credential()` (e.g. a `Settings.AuthMethod`/`Source` field = `token` | `gh_cli`; for `gh_cli` run `gh auth token --hostname github.com` instead of reading the secret), (b) a new admin action (e.g. `scm.providers.use_gh_cli`) mirroring `SetToken` - validate with `CurrentUser`, record account, mark the source - plus a manifest entry (the `TestU2_ManifestActionKeysMatchTheRuntime` parity test, internal/plugin/manifest_test.go:136-142, requires both), (c) `RemoveToken` clearing the source, and (d) a UI control in `ProviderCard` (source-control-section.tsx) shown for `p === "github"` only, with `ProviderView` gaining the source so the card can show it.
- **Process execution is possible and unrestricted**: the plugin is a native binary spawned by Kandev with `cmd.Env = append(os.Environ(), KANDEV_PLUGIN_DATA_DIR=...)` and `cmd.Dir = installPath` (kandev manager.go:392-394, v0.96.0), so it inherits the Kandev server user's PATH/HOME/GH_TOKEN/GH_CONFIG_DIR. manifest.yaml has no exec/permission capability to declare (capabilities: state, secrets, api_read, api_write, events). pluginsdk Host offers no GitHub-credential API, so the plugin must exec `gh` itself (`os/exec`, stdlib). Kandev itself already does the same (`gh auth token --hostname <host> [--user <login>]`, kandev internal/github/gh_accounts.go:201-215; executor_host_gh_bridge.go:47) - a reference for argument shape, older-gh `--user` support detection and empty-token handling.
- **Risks / follow-up**:
  - The gh token is resolved on the Kandev server host, not the browser user's machine: it works only when `gh` is installed and logged in for the account that runs Kandev (self-hosted). In Docker or when `gh` is absent, the action must fail with a clear, typed error (new sentinel, e.g. `ErrGHCLIUnavailable`, mapped in `classifySCM`, internal/plugin/scm_actions.go:178-200; currently `ErrNoToken` maps to `validation`/field `token`).
  - Decide: resolve the token on every `credential()` call (always current, survives `gh auth refresh`; costs one subprocess per call including the 1-minute watcher, so use a short timeout and possibly a small TTL cache) versus importing it once into the secret store (zero exec at runtime, but goes stale when gh rotates the token). Never persist the gh token in plugin state; if cached in memory it must be redacted the same way.
  - Redaction: the token from `gh` stdout must be `strings.TrimSpace`d, wrapped with `redact.WithSecrets`, and gh's stderr must never be surfaced in action errors or logs (team rule: no secrets / provider response text in errors). `scm.Credential` already hides `Token` in all fmt verbs (client.go:15-22). Add a leak test like the existing ones (internal/scm/httpx_test.go, internal/plugin/actions_test.go).
  - Testability: inject the command runner (e.g. `func(ctx) (string, error)` field on `scm.Service`, like the existing `Now` clock) so tests never exec a real `gh`; `newHarness` (internal/scm/harness_test.go:232-250) builds the service and is the place to wire a fake.
  - Lint: `exec.CommandContext("gh", ...)` with fixed args needs `//nolint:gosec // G204: fixed arguments` per existing convention (internal/ci/changes.go:41).
  - Scope: only GitHub has `gh`; GitLab (`glab auth token` exists, Kandev uses `glab auth status -t`) and Bitbucket stay token-only unless the requirements say otherwise. `taskPRs` (scm_actions.go:218-252) already uses Kandev's own PR data, not a plugin token - unaffected.
  - `scm.providers.list` is unguarded and member-readable; any new source field in `ProviderView` must stay non-secret (NFR1).
  - Environment: `go` is not on PATH (use `~/.local/go/bin`) and `../kandev` is missing in this worktree; Construction must link `../kandev` to `~/repo/kandev` (v0.96.0) or use an out-of-tree `go.work` as this scan did.
