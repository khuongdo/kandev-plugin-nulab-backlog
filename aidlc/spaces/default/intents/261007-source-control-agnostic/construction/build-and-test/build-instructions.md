# Build Instructions — Multi-provider source control

## Prerequisites

- Go 1.26.x with CGO (for `-race`). In this environment it is at `~/.local/go/bin`; add it to `PATH`.
- Node.js + npm (Node 25 was used); run `npm ci` in `ui/` once.
- `../kandev` (next to the repository) must be the Kandev checkout at tag `v0.96.0`, matching `.kandev-sdk-ref`. `go.mod` replaces `github.com/kandev/kandev` with `../kandev/apps/backend`, and the UI `tsconfig` aliases `@kandev/plugin-sdk` into it. In this worktree it is a symlink to `/home/k_do_webfrontier/repo/kandev`.
- No environment variables, local services or new dependencies are needed. Provider tokens are only entered at runtime in Settings.

## Build Commands

From the repository root:

```bash
export PATH=$HOME/.local/go/bin:$PATH
make check-sdk          # ../kandev matches .kandev-sdk-ref
make build              # UI bundle + server binaries for all release platforms
make package            # dist/nulab-backlog-<version>.tar.gz + checksums.txt
make verify-package     # archive layout, manifest id/version, checksums
```

## Build Verification

- `make check-format vet lint check-secrets` must exit 0 (gofmt, prettier, go vet, golangci-lint + gosec, tsc strict, eslint, actionlint, CI workflow and secrets checks).
- `go mod tidy` must leave `go.mod` / `go.sum` unchanged (no new dependency).
- `make verify-package` prints `verifypkg: OK dist/nulab-backlog-<version>.tar.gz (nulab-backlog@<version>)`.
- `make contract-test KANDEV_MIN_DIR=../kandev` installs the packaged plugin into a throwaway Kandev v0.96.0 server and must pass; run it 10 times to catch host startup races (project rule).

## Troubleshooting

- `go: command not found`: put `~/.local/go/bin` on `PATH`.
- `check-sdk` fails or imports of `github.com/kandev/kandev` do not resolve: `../kandev` is missing or not at `v0.96.0`; re-create the symlink and run `git -C ../kandev describe --tags`.
- UI type errors about `@kandev/plugin-sdk`: same cause as above, or `ui/node_modules` missing (`npm ci`).
- `build/` and `dist/` are git-ignored; `make clean` removes them. Keep `coverage.out` under `build/` only.
