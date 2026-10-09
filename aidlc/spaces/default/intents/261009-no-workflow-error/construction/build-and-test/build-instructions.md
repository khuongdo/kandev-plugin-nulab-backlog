# Build Instructions — 261009-no-workflow-error

## Prerequisites

- Go 1.26.x with CGO (for `-race`). If missing, install to `~/.local/go` (checksum-verified from go.dev) and `export PATH=$HOME/.local/go/bin:$PATH`.
- Node.js + npm; run `npm ci` in `ui/` once.
- Kandev SDK checkout at `../kandev`, pinned to tag `v0.96.0` (this worktree uses a symlink `../kandev -> ~/repo/kandev`).

## Environment Setup

No environment variables or local services are needed for build and unit tests. The packaged-host contract test builds a throwaway Kandev server from `../kandev` itself.

## Build Commands

```bash
make check-format vet lint     # gofmt, go vet, golangci-lint (+gosec), tsc/ESLint/Prettier, actionlint, CI workflow check
make build                      # Go binaries + UI bundle
make package verify-package     # dist/nulab-backlog-<version>.tar.gz + checksums.txt, then package verification
```

## Build Verification

- `make verify-package` exits 0 and the package contains the four platform executables, `manifest.yaml` with `api_read: ["tasks", "repositories", "workflows"]` and the `workflows.status` action, and the UI bundle.
- `make contract-test KANDEV_MIN_DIR=../kandev` installs the package into a Kandev v0.96.0 server and exits 0.

## Troubleshooting

- `go: command not found` → put `~/.local/go/bin` on `PATH`.
- `replace ../kandev` errors → create the `../kandev` symlink/checkout at `v0.96.0`.
- `coverage.out` must not remain in the repo root; `make coverage` writes under `build/`.
