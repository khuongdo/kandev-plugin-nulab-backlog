# Build Instructions — 261007-uiux-github-style

## Dependencies

- Go 1.26.x with CGO (gcc) for `-race`: `export PATH="$HOME/.local/go/bin:$HOME/go/bin:$PATH"`.
- Kandev SDK checkout next to the repo: `../kandev` at the commit in `.kandev-sdk-ref` (`f099a46…`, tag `v0.96.0`). `make check-sdk` verifies it.
- Node 22 (`.nvmrc`) and `npm ci` in `ui/`.
- `golangci-lint` (with `gosec`) and `actionlint` at the versions pinned in CI, for `make lint`.

## Environment Setup

No environment variables or local services are needed for build and unit tests. The contract test starts a throwaway Kandev on ports 38529/39529 with a temporary home directory.

## Build Commands

Run from the repository root:

```bash
make check-format vet lint test coverage check-secrets build package verify-package
```

- `build` compiles the plugin binary and the UI bundle (`build/ui/bundle.js`, React never bundled).
- `package` writes the plugin package under `build/`; `verify-package` checks its manifest, files and forbidden content.
- `coverage` writes its profile under `build/` (never the repository root) and fails below 80%.

Contract test on the minimum Kandev version (0.96.0). The SDK checkout is already at `v0.96.0`, so it doubles as the minimum-version checkout:

```bash
make contract-test KANDEV_MIN_DIR=../kandev
```

## Build Verification

- Every `make` target exits 0.
- `git status` shows no `coverage.out` at the repository root and no change to `go.mod`/`go.sum` from `go mod tidy`.
- `verify-package` prints its pass line.

## Troubleshooting

- `check-sdk: ../kandev is missing` — link or check out Kandev at `.kandev-sdk-ref`.
- `-race requires cgo` — install gcc and keep `CGO_ENABLED=1`.
- UI build fails on fragments — `make ui-build` and `npm run build` both pass `--jsx-fragment=Fragment`.
- Contract test cannot bind a port — another process uses 38529/39529; override `CONTRACT_PORT` / `CONTRACT_AGENTCTL_PORT`.
