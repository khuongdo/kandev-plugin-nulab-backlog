# Technology Stack — kandev-plugin-nulab-backlog

## Languages and Runtimes

| Item | Version | Source |
|---|---|---|
| Go | `go 1.26.0` (baseline toolchain go1.26.8) | `go.mod` |
| Node.js | 22 (local machine had v25.2.1 during the scan) | `.nvmrc`, CI |
| TypeScript | ~6.0.3, `strict` | `ui/package.json`, `ui/tsconfig.json` |
| Kandev host | min `0.96.0`; SDK pin `f099a46dc7aab16f6ff5806cd29b2b480296303f` (tag `v0.96.0`) | `manifest.yaml`, `.kandev-sdk-ref` |

## Frameworks and Libraries

- Backend: Kandev `pkg/pluginsdk` (go-plugin over gRPC); stdlib `net/http` for Backlog, no third-party HTTP client or Backlog SDK. Versions: [dependencies.md](dependencies.md).
- UI runtime: React and `host.ui` (shadcn-based `@kandev/ui`) supplied by the host via `host.React` / `host.jsx`; types from `@kandev/plugin-sdk` (path-mapped). No React in the bundle.
- UI build: esbuild ^0.28.2 to `build/ui/bundle.js`.

## Build, Test and Quality Tooling

- Build: `Makefile` is the entry point and CI calls its targets (`check-sdk`, `check-format`, `vet`, `lint`, `test`, `coverage`, `check-secrets`, `build`, `ui-build`, `package`, `verify-package`, `clean`). `make package` = 5-platform `CGO_ENABLED=0` build + `ui-build` + Kandev `cmd/plugin-pack` → `dist/nulab-backlog-<ver>.tar.gz` + `checksums.txt`.
- Test: Go `testing` + testify, always `-race`, `httptest`; Vitest ^5.0.3 + jsdom ^30.1.2 + axe-core ^4.14.0.
- Lint and format: `gofmt`, `go vet`, golangci-lint v2.14.0 + `gosec`, actionlint v1.7.12 (both via `go run`); ESLint ^10.12.0 + typescript-eslint ^8.71.1, Prettier ^3.9.9 (`printWidth: 110`).
- CI/CD: GitHub Actions (`ci.yml`, `release.yml`), SHA-pinned actions, build provenance attestation.
