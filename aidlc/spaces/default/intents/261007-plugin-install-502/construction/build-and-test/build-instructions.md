# Build Instructions — Plugin install failed: 502

## Prerequisites

- Go 1.26.x on `PATH` (this machine: `~/.local/go/bin`, go1.26.8). CGO and a C compiler for `-race`.
- `../kandev` (sibling of the repository root) at the commit in `.kandev-sdk-ref` (Kandev `v0.96.0`, `f099a46dc7aab16f6ff5806cd29b2b480296303f`). In this worktree it is a detached `git worktree` of `~/repo/kandev`.
- Node.js per `.nvmrc` and `npm ci` in `ui/` (needed by `check-format`, `lint`, `test` and `ui-build`).

## Build Commands

Run from the repository root, in CI order:

```bash
export PATH=$HOME/.local/go/bin:$PATH
make check-format vet lint test coverage check-secrets build package verify-package
```

- `build` now empties `build/server/` first and builds 4 executables: `linux-amd64`, `linux-arm64`, `darwin-amd64`, `darwin-arm64` (no Windows).
- `package` writes `dist/nulab-backlog-<version>.tar.gz` and `dist/checksums.txt`; `verify-package` runs `cmd/verifypkg`.
- Packaged-host contract test on the minimum Kandev version (team Testing Posture):

```bash
make contract-test KANDEV_MIN_DIR=../kandev
```

## Build Verification

- `verifypkg: OK dist/nulab-backlog-<version>.tar.gz` is printed.
- `tar tzf dist/nulab-backlog-<version>.tar.gz` lists `manifest.yaml`, `ui/bundle.js`, `checksums.txt` and exactly the 4 `server/plugin-*` executables.
- `ls -l dist/*.tar.gz` shows at most 25 MB (NFR1).

## Troubleshooting

- `check-sdk: ../kandev is missing` / wrong commit: recreate the worktree at `.kandev-sdk-ref`.
- `verifypkg: unexpected executable: server/...`: a stray file reached the staging dir; run `make package` again (it rebuilds `build/server/` from empty).
- `coverage.out` must stay under `build/` (project rule); never pass `-coverprofile` at the repository root.
