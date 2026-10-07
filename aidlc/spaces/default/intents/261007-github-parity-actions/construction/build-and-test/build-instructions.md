# Build Instructions — github-parity-actions

## Prerequisites

- Go 1.26.x on `PATH` (used here: go1.26.8, `export PATH=$HOME/.local/go/bin:$PATH`), with CGO available for `-race`.
- `../kandev` must be a checkout of Kandev at tag `v0.96.0` (commit `f099a46`). It is needed by the `replace` in `go.mod` and by the contract test. In this worktree it is a symlink to `/home/k_do_webfrontier/repo/kandev`.
- Node.js 22 (the CI version; Node 25 also worked locally) and `npm ci` in `ui/`.
- gcc, which the contract test needs to build Kandev.

## Build and Package

```bash
make check-format vet lint      # gofmt, go vet, golangci-lint (+gosec), CI workflow check
make build                      # Go plugin binary
make package                    # Go binary + UI bundle -> dist/nulab-backlog-0.2.0.tar.gz + checksums.txt
make verify-package             # verifypkg: id nulab-backlog, version 0.2.0
(cd ui && npm run typecheck && npm run lint && npm run format:check)
```

## Verification

- `make verify-package` prints `verifypkg: OK dist/nulab-backlog-0.2.0.tar.gz (nulab-backlog@0.2.0)`.
- `git diff go.mod go.sum` is empty. No dependency was added.

## Troubleshooting

- `check-sdk` fails: `../kandev` is missing or is not at `v0.96.0`. Fix it with `git -C ../kandev checkout v0.96.0`.
- The coverage profile goes to `build/coverage.out`, which is gitignored. Never leave a `coverage.out` in the repo root (project rule).
