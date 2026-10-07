# Code Structure — kandev-plugin-nulab-backlog

## Layout

| Path | Classification | Notes |
|---|---|---|
| `server/main.go` | entrypoint | `pluginsdk.Serve(plugin.NewRuntime())` only |
| `internal/plugin/` | adapter | handlers map, webhook, events, host port, credentials; `Version`/`SDKRef` set by ldflags |
| `internal/connection/`, `internal/issues/`, `internal/git/`, `internal/scm/` | domain services | see [component-inventory.md](component-inventory.md) |
| `internal/backlog/`, `internal/github/`, `internal/gitlab/`, `internal/bitbucket/` | HTTP clients | stdlib `net/http`, `testdata/` JSON fixtures |
| `internal/redact/`, `internal/testutil/` | utility / test helper | |
| `internal/pkgverify/`, `cmd/verifypkg/` | build tooling | offline package verifier |
| `internal/ci/`, `cmd/ci/` | build tooling | `secrets`, `workflows`, `contract`, `preflight`, `marketplace` |
| `ui/src/` | UI bundle source | `index.ts` registration; areas `settings/`, `issues/`, `git/`, `page/`, `switch/`, `brand/`, `messages/`, `testing/` |
| `manifest.yaml` | plugin contract | see [api-documentation.md](api-documentation.md) |
| `.github/workflows/` | CI/CD | see [code-quality-assessment.md](code-quality-assessment.md) |
| `docs/` | docs | `brand/`, `manual-checks/` |

## Build and Packaging

`Makefile` targets (every Go target first runs `check-sdk`: `../kandev` HEAD must equal `.kandev-sdk-ref`):

1. `build` -> `build/server/plugin-{linux-amd64,linux-arm64,darwin-amd64,darwin-arm64,windows-amd64.exe}` (`CGO_ENABLED=0`, `-trimpath`, `-ldflags "-s -w -X ..."`).
2. `ui-build` -> `build/ui/bundle.js` (esbuild).
3. `package` -> stage `build/stage/{manifest.yaml,server/*,ui/bundle.js}` -> Kandev `cmd/plugin-pack` -> `dist/nulab-backlog-<version>.tar.gz` + `dist/checksums.txt`.
4. `verify-package` -> `cmd/verifypkg` (checksums, required files, manifest id/version, exactly 5 executables, no Nulab asset URL in the bundle).
5. `contract-test` -> builds Kandev at `min_kandev_version`, installs the package over loopback, calls two actions.

The platform set is declared three times: `manifest.yaml` `runtime.executables`, `Makefile` `PLATFORMS`, `internal/pkgverify` `executables` map.

## Code Patterns

- Ports-and-adapters: only `internal/plugin` maps domain errors to `pluginsdk` codes.
- Errors wrapped with `%w`; `context.Context` first on I/O; responses bounded by `io.LimitReader`.
- Tests co-located (`_test.go`, `*.test.ts[x]`), fake servers via `httptest`.
