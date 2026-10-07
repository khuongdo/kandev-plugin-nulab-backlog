# Technology Stack — kandev-plugin-nulab-backlog

## Languages and Runtimes

| Item | Version | Source |
|---|---|---|
| Go | `go 1.26.0` | `go.mod` |
| Node.js | 22 | `.nvmrc`, CI |
| TypeScript | ~6.0.3, `strict` | `ui/package.json`, `ui/tsconfig.json` |
| Kandev host | min `0.96.0`; SDK pin `f099a46dc7aab16f6ff5806cd29b2b480296303f` (tag `v0.96.0`) | `manifest.yaml`, `.kandev-sdk-ref` |

## Frameworks and Libraries

- Backend: Kandev `pkg/pluginsdk` (go-plugin over gRPC, incl. `GitCredentialHandler`); stdlib `net/http` for Backlog. No third-party HTTP client and no vendor SDK (team rule) — a GitHub/Bitbucket client would also be stdlib-only. Versions: [dependencies.md](dependencies.md).
- UI runtime: React 19 and `host.ui` supplied by the host; types from `@kandev/plugin-sdk` (path-mapped). No React and no CSS in the bundle.
- UI build: esbuild ^0.28.2 to `build/ui/bundle.js`.

## Build, Test and Quality Tooling

- Build: `Makefile` targets (`check-sdk`, `check-format`, `vet`, `lint`, `test`, `coverage`, `check-secrets`, `ui-build`, `build`, `package`, `verify-package`, `contract-test`, `release-preflight`, `marketplace-entry`); CI calls them.
- Test: Go `testing` + testify, always `-race`, `httptest`; Vitest ^5.0.3 + jsdom + axe-core.
- Lint and format: `gofmt`, `go vet`, golangci-lint + `gosec`; ESLint ^10.12.0, Prettier ^3.9.9.
- CI/CD: GitHub Actions (`ci.yml`, `release.yml`), SHA-pinned actions, build provenance attestation.
