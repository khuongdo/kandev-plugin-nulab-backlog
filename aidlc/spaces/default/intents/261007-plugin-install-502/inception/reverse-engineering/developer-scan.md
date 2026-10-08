# Developer Code Scan - kandev-plugin-nulab-backlog

Intent: 261007-plugin-install-502 (bugfix, Minimal depth, full rescan of `./`).
Scanned commit: `1819cc3` (branch `feature/plugin-install-faile-9qm`), manifest version `0.4.1`.

## Developer Code Scan Results

### Scan Coverage
- **Analyzed deeply**:
  - `manifest.yaml`
  - `Makefile`
  - `go.mod`, `.kandev-sdk-ref`, `.nvmrc`, `.gitignore`
  - `server/main.go`
  - `cmd/verifypkg/main.go`
  - `internal/pkgverify/pkgverify.go`
  - `internal/ci/contract.go`, `internal/ci/marketplace.go`
  - `.github/workflows/ci.yml`, `.github/workflows/release.yml`
  - `README.md` (Build, Install on Kandev, CI checks, Releasing, Marketplace sections)
  - `ui/package.json`, `ui/vitest.config.ts`, `.golangci.yml`
- **Skimmed only**:
  - `internal/plugin/` (runtime entry `NewRuntime`, `Version`/`SDKRef` ldflags vars; action handlers not read)
  - `internal/backlog/`, `internal/connection/`, `internal/issues/`, `internal/git/`, `internal/scm/`, `internal/github/`, `internal/gitlab/`, `internal/bitbucket/`, `internal/redact/`, `internal/testutil/` (package doc comments, file/test counts, Backlog endpoint strings)
  - `internal/ci/` other files (`run.go`, `manifest.go`, `release.go`, `secrets.go`, `workflows.go`), `cmd/ci/main.go`
  - `ui/src/` (directory layout only), `ui/tsconfig.json`, `ui/eslint.config.js`
  - `docs/` (`brand/`, `manual-checks/`), `LICENSE`
- **External evidence consulted (read-only, outside the snapshot, not part of coverage)**: the local Kandev checkout `~/repo/kandev` at tag `v0.96.0` (`apps/backend/internal/plugins/handlers.go`, `service_install.go`, `pkgtar/pkgtar.go`, `backendapp/httpserver.go`, `common/config/catalog.go`, `apps/web/lib/api/domains/plugins-api.ts`), the running Kandev server's logs `~/.kandev/logs/backend-logs.log`, the installed runtime version (npm `kandev` 0.97.0), and `gh release view v0.4.1`.

### Packages Found
- `server` (cmd `main`) — binary entry — Go — calls `pluginsdk.Serve(plugin.NewRuntime())` only.
- `cmd/verifypkg` — CLI — Go — thin wrapper over `internal/pkgverify`.
- `cmd/ci` — CLI — Go — CI/release subcommands (`secrets`, `workflows`, `contract`, `preflight`, `marketplace`) over `internal/ci`.
- `internal/plugin` — adapter — Go — KandevAdapter; the only package (with `server`) importing `pluginsdk`; action/webhook/event handlers, host port, references, credentials.
- `internal/backlog` — HTTP client — Go — BacklogGateway: Backlog REST v2 (`/api/v2/users/myself`, `projects`, `issues`, `issues/count`, comments, attachments, statuses, project users, git repositories, pull requests, `oauth2/token`).
- `internal/connection` — domain service — Go — address validation, API key and OAuth connect, lifecycle, project selection, Git credential, persisted state.
- `internal/issues` — domain service — Go — issue list/filters, tasks from issues, links, sync, watches, quick actions, saved queries.
- `internal/git` — domain service — Go — Backlog Git repository provider, PR links/create/status, watches, queries.
- `internal/scm` — domain service — Go — provider-neutral source-control context (GitHub/GitLab/Bitbucket links, queries, watches).
- `internal/github`, `internal/gitlab`, `internal/bitbucket` — HTTP clients — Go — read-only SCM REST clients.
- `internal/redact` — utility — Go — masks secrets and Backlog URL query strings.
- `internal/pkgverify` — build tooling — Go — offline package verifier (replaces Kandev's missing verify CLI at v0.96.0).
- `internal/ci` — build tooling — Go — contract test driver, manifest/marketplace/release/secret/workflow checks.
- `internal/testutil` — test helper — Go.
- `ui` — browser bundle — TypeScript/TSX — one ESM bundle (`ui/bundle.js`), React provided by the host (`h`/`Fragment` JSX factory); areas `settings/`, `issues/`, `git/`, `page/`, `switch/`, `brand/`, `messages/`.

### Build System
- **Type**: Go modules (`go 1.26.0`) + npm (`ui/package-lock.json`, Node from `.nvmrc` = 22) driven by `make`.
- **Config Files**: `Makefile`, `go.mod`, `go.sum`, `.kandev-sdk-ref` (`f099a46dc7aab16f6ff5806cd29b2b480296303f` = Kandev `v0.96.0`), `ui/package.json`, `ui/tsconfig.json`, `ui/vitest.config.ts`, `ui/eslint.config.js`, `ui/.prettierrc`, `.golangci.yml`.
- **Build Dependencies**:
  - `go.mod` `replace github.com/kandev/kandev => ../kandev/apps/backend`; every Go target depends on `check-sdk` (fails unless `../kandev` HEAD equals `.kandev-sdk-ref`).
  - `build` -> 5 cross-compiled binaries `build/server/plugin-{linux-amd64,linux-arm64,darwin-amd64,darwin-arm64,windows-amd64.exe}` (`CGO_ENABLED=0`, `-trimpath`, `-ldflags "-s -w -X ...Version -X ...SDKRef"`).
  - `ui-build` -> `build/ui/bundle.js` (esbuild, fails if React was bundled).
  - `package` (build + ui-build) -> stage `build/stage/{manifest.yaml,server/*,ui/bundle.js}` -> Kandev's own `go -C ../kandev/apps/backend run ./cmd/plugin-pack` -> `dist/nulab-backlog-<version>.tar.gz` + `dist/checksums.txt`.
  - `verify-package` -> `go run ./cmd/verifypkg` (dist checksum, in-archive `checksums.txt`, required files, no Nulab asset URL in the bundle, manifest id/version, exactly the 5 executables).
  - `contract-test` -> builds Kandev at `v<min_kandev_version>` from `KANDEV_MIN_DIR`, starts it on port 38529 (agentctl 39529), installs the package via multipart `POST /api/plugins/install`, waits for `active`, calls `connection.get` and `connection.connect_api_key`.
  - `release-preflight`, `marketplace-entry` -> `go run ./cmd/ci ...`.

### APIs Discovered
- Kandev plugin actions — `manifest.yaml` `actions` — 77 actions (`connection.*` 9, `repositories.*` 2, `git.*` 19, `issues.*` 23, `scm.*` 20 approx.), `scope` workspace/task, `access` authenticated/admin, `max_body_bytes` 8 KiB-256 KiB.
- Kandev plugin webhook — `manifest.yaml` `webhooks` — 1 (`oauth-callback`, GET, public, 1 KiB).
- Kandev host contract — `manifest.yaml` — `api_version: 2`, `runtime.type: binary`, 5 executables, `min_kandev_version: "0.96.0"`, capabilities (`state`, `secrets`, `api_read: [tasks, repositories]`, `api_write: [tasks]`, `events: [task.deleted]`), `repository_providers: [nulab-backlog]`, 1 `reference_sources` entry, `config_schema` (OAuth client id/secret, public base URL), `ui.bundle: /ui/bundle.js`.
- Kandev install API used by the contract test — `internal/ci/contract.go` — `GET /ready`, `POST /api/plugins/install` (multipart field `package`, expects 201 and no `warning`), `GET /api/plugins/<id>`, workspace list and plugin action calls.
- Backlog REST API v2 client — `internal/backlog/` — about 17 distinct endpoint paths (users, projects, statuses, users per project, issues, issue count, comments, attachments, git repositories, pull requests and count, OAuth token).
- GitHub / GitLab / Bitbucket REST clients — `internal/github`, `internal/gitlab`, `internal/bitbucket` — read-only repos, PRs/MRs, user.

### Frameworks & Libraries
- Kandev plugin SDK (`github.com/kandev/kandev/pkg/pluginsdk`) — local replace at v0.96.0 — host RPC, built on `hashicorp/go-plugin` v1.8.0 / gRPC v1.83.1 (indirect).
- `github.com/stretchr/testify` — v1.12.1 — Go test assertions.
- `gopkg.in/yaml.v3` — v3.0.1 — manifest parsing in `pkgverify` and `ci`.
- Go standard library `net/http` for all external HTTP (no third-party HTTP client).
- UI dev dependencies: `esbuild` ^0.28.2, `typescript` ~6.0.3, `vitest` ^5.0.3, `jsdom` ^30.1.2, `eslint` ^10.12.0 + `typescript-eslint` ^8.71.1, `prettier` ^3.9.9, `react`/`react-dom` ^19.3.0 (types and tests only; never bundled), `axe-core` ^4.14.0.
- Tooling pinned in `Makefile`: `golangci-lint` v2.14.0, `actionlint` v1.7.12.

### Test Coverage
- **Test Directories**: co-located `_test.go` in every `internal/*` package (about 108 Go test files; `testdata/` JSON fixtures in `backlog`, `github`, `gitlab`, `bitbucket`); `ui/src/**/*.test.{ts,tsx}` (about 40 files).
- **Test Frameworks**: Go `testing` + testify `require`, `net/http/httptest` fake servers, `go test -race`; Vitest with jsdom.
- **Coverage Config**: present — `make coverage` enforces an 80% line floor over `./internal/... ./server/...`, sole exclusion `server/main.go`, profile written to `build/coverage.out`.
- **Packaging tests**: `internal/pkgverify/pkgverify_test.go`, `internal/ci/contract_test.go` (driver unit tests); the real install path is exercised only by `make contract-test` in CI (`packaged-host-contract`) and release (`contract`).
- **Baseline**: not run in this scan — no Go toolchain on `PATH` and no `../kandev` sibling checkout in this worktree, so `check-sdk` would fail.

### Code Quality Indicators
- **Linting**: `.golangci.yml` (v2, standard set + `gosec`); `gofmt` via `check-format`; `go vet`; UI `tsc --noEmit` (strict), ESLint (`ui/eslint.config.js`), Prettier (`ui/.prettierrc`); `actionlint` plus `cmd/ci workflows` policy (SHA-pinned actions, `contents: read` default, no `pull_request_target`).
- **CI/CD**: `.github/workflows/ci.yml` (`checks`: format/vet/lint/test/coverage/check-secrets/build/package/verify-package; `packaged-host-contract`: installs the uploaded package on Kandev built at `v0.96.0`); `.github/workflows/release.yml` (`verify` -> `contract` -> `publish` with build provenance attestation and `gh release create`).
- **Documentation**: `README.md` covers build, install (upload via Settings > Plugins), API-key and OAuth connection, CI, releasing, marketplace; Go packages carry `doc.go` package comments; `docs/manual-checks/` holds the first-release record.

### Technical Debt Signals
- Package size is about 29.5 MB (`nulab-backlog-0.4.1.tar.gz` = 29,498,475 bytes per `gh release view`); 5 stripped Go binaries of 14.4-15.9 MB each make up almost all of it (`ui/bundle.js` is about 128 KB). Every user downloads/uploads all five platforms to install one.
- `internal/pkgverify` hard-codes exactly 5 executables (`executables` map) and the Makefile `PLATFORMS` list duplicates it; changing the platform set needs edits in `manifest.yaml`, `Makefile`, and `pkgverify` together.
- `pkgverify` duplicates Kandev's internal `pkgtar` checksum rules because v0.96.0 ships no verify CLI; it does not check size limits or the host's upload path.
- The contract test installs over loopback with a 30 s client timeout (`internal/ci/contract.go`), so it never reproduces a slow upload through the Kandev web/proxy layer.
- Release builds go to `dist/` in the sibling repo checkout `~/repo/kandev-plugin-nulab-backlog/dist`, which still holds stale `0.0.1`/`0.1.0` tarballs (about 28.8 MB each); easy to upload the wrong file manually.
- Local runtime drift: the running Kandev is npm `kandev` 0.97.0, while the SDK pin and `min_kandev_version` are 0.96.0 and the reference checkout `~/repo/kandev` is at `v0.96.0`.

## Handoff Summary
- **Intent-relevant finding**: The host, not the plugin manifest, is cutting the upload. Evidence:
  1. Kandev logs `~/.kandev/logs/backend-logs.log` lines 26817 and 26823 (2026-10-08 07:32:04 and 07:32:45): `POST /api/plugins/install` -> status 400, `duration_ms` 30003 / 30000, `bytes` 47. 47 bytes is exactly `{"error":"missing multipart field \"package\""}`, which `installFromRequest` (`apps/backend/internal/plugins/handlers.go:155-160`, v0.96.0) returns when `ctx.FormFile("package")` fails.
  2. The backend HTTP server sets `ReadTimeout: cfg.Server.ReadTimeoutDuration()` (`apps/backend/internal/backendapp/httpserver.go:117`) with default `server.readTimeout` = 30 s (`apps/backend/internal/common/config/catalog.go:61`, env `KANDEV_SERVER_READTIMEOUT`). The 29.5 MB multipart body did not finish arriving within 30 s, the read was cut, and the multipart parse failed.
  3. The Kandev install handler never answers 502 (it maps errors to 409/400/500, `handlers.go:187-204`), so the `Plugin install failed: 502 ...` text (built in `apps/web/lib/api/domains/plugins-api.ts:79` from the status line when the body has no `error`) comes from a proxy layer between the browser and the backend that saw the upstream connection closed mid-upload. Earlier installs failed differently (07 Oct 10:57: 500 `pkgtar: invalid gzip stream`, i.e. a non-gzip file uploaded).
  4. Plugin-side contributor: the package carries all 5 platform binaries (about 29.5 MB). The manifest, `min_kandev_version`, checksums and layout pass `verify-package` and the CI contract test, so nothing indicates a malformed package.
- **Risks / follow-up**:
  - Fix options for the architect to weigh: (a) install by URL (`POST /api/plugins/install` JSON `{"url": ...}` pointing at the GitHub Release asset; the backend downloads it itself with a 100 MiB cap, `service_install.go:22-24,293-327`), avoiding the browser upload and the 30 s read timeout; (b) document raising `KANDEV_SERVER_READTIMEOUT`; (c) shrink the package (fewer platforms or per-platform packages), which requires changing `manifest.yaml` `runtime.executables`, `Makefile` `PLATFORMS`, and `internal/pkgverify` `executables` together and keeping `make verify-package` and the contract test green; Kandev rejects a package whose platform is missing (`pkgtar.ErrPlatformNotSupported`).
  - Unconfirmed: which proxy produced the 502 (the `kandev --headless` launcher or another reverse proxy) and why a loopback upload took more than 30 s; the 0.97.0 runtime source was not inspected (only v0.96.0). Confirm before changing code.
  - No Go test baseline was recorded (no Go on `PATH`, no `../kandev` in this worktree); link `../kandev` to the `v0.96.0` checkout before Construction, per `project.md` Testing Posture.
  - Keep `pkgverify`'s 5-executable contract, the 80% coverage floor and the contract test unchanged unless the chosen fix deliberately changes the platform set.
