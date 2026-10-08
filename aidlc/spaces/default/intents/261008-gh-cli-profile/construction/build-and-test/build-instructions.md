# Build Instructions — 261008-gh-cli-profile

## Dependencies

- Go 1.26.x with CGO (for `-race`); local install used here: `~/.local/go` (go1.26.8).
- `../kandev` → Kandev checkout at tag `v0.96.0` (commit `f099a46dc…`, matches `.kandev-sdk-ref`). In this worktree it is a symlink to `~/repo/kandev`. `go.mod` keeps `replace github.com/kandev/kandev => ../kandev/apps/backend`; never edit it.
- Node per `.nvmrc`; `npm ci` in `ui/`.
- golangci-lint at the version pinned in CI (v2.14.0).
- gh CLI is NOT needed to build or test (tests use a fake runner); gh ≥ 2.40 is needed at runtime for `--user`, gh ≥ 2.81.0 for the multi-account picker (`gh auth status --json`).

## Environment

No env vars or local services. Tests never call real GitHub, Backlog or gh.

## Build and Verify (Makefile targets, same as CI)

```bash
export PATH=$HOME/.local/go/bin:$PATH
make check-sdk check-format vet lint check-secrets
make coverage                      # go test -race, 80% floor, profile under build/
(cd ui && npx vitest run && npx tsc --noEmit && npx eslint . && npx prettier --check .)
make package verify-package        # dist/nulab-backlog-0.5.3.tar.gz + checksums
make contract-test KANDEV_MIN_DIR=../kandev   # run 10 times
```

## Troubleshooting

- `check-sdk` fails: `../kandev` missing or not at the `.kandev-sdk-ref` commit — recreate the symlink / checkout `v0.96.0`.
- `-race` errors about CGO: install a C compiler (`gcc`).
- A stray `coverage.out` in the repo root blocks the Code Generation gate; `make coverage` writes under `build/`.
