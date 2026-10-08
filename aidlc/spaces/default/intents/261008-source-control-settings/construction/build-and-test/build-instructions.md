# Build Instructions

## Prerequisites

- Go 1.26.x on `PATH` (in this worktree: `export PATH=$HOME/.local/go/bin:$HOME/go/bin:$PATH`; `gofmt` must resolve).
- `../kandev` checked out at the commit in `.kandev-sdk-ref` (v0.96.0); `go.mod` uses a `replace` to it. In a fresh worktree, link it to the local Kandev checkout (`ln -s ~/repo/kandev ../kandev`).
- Node.js with `ui/node_modules` installed (`cd ui && npm ci`).
- `golangci-lint` (version pinned in CI) on `PATH` for `make lint`.

## Environment

No environment variables or local services are needed for build and unit tests. Tests use fake hosts and `httptest` servers; never real tokens.

## Build Commands

```bash
make check-format vet lint      # gofmt, go vet, golangci-lint (+gosec), tsc, eslint, actionlint
make test                       # Vitest + go test -race
make coverage                   # go test -race -coverprofile=build/coverage.out, floor 80%
cd ui && npm run typecheck && npm run lint && npm run format:check
make package verify-package     # builds dist/nulab-backlog-<version>.tar.gz + checksums, verifies
```

## Build Verification

- `make coverage` prints `coverage: N% (floor 80%, excluded: server/main.go)` and exits 0.
- `verifypkg: OK dist/nulab-backlog-<version>.tar.gz` is printed by `make verify-package`.
- Delete `build/coverage.out` after a local run so it does not land in the AI-DLC record.

## Troubleshooting

- `gofmt: command not found` (exit 127 from `check-format`): Go is not on `PATH`; export the path above.
- `replace` errors from `go build`: `../kandev` is missing or not at the pinned commit.
- Contract test start-up failures: the Kandev host is injected asynchronously; rerun to see whether it is intermittent and keep the 10-run check.
