# Build Instructions — 261007-backlog-panel-retouch

## Prerequisites

- Node.js with npm; UI dependencies installed with `npm --prefix ui ci`.
- Go 1.26.x **with the `covdata` tool** (needed by `make coverage`). The toolchain that `~/go-sdk/bin/go` auto-switches to (`golang.org/toolchain@v0.0.1-go1.26.0`, module cache) ships without `covdata`, so `make coverage` fails with `go: no such tool "covdata"`. Use a complete Go 1.26 toolchain on `PATH` with `GOTOOLCHAIN=local` (this run used a writable copy at `<scratchpad>/go126` that contains `pkg/tool/linux_amd64/covdata`).
- `../kandev` → Kandev checkout at tag `v0.96.0` (commit `f099a46`, equal to `.kandev-sdk-ref` and `min_kandev_version`). In this worktree: `../kandev -> /home/k_do_webfrontier/repo/kandev`.

## Environment

No environment variables or config files are required for build or test. No secrets are used (tests use fakes only).

## Build Commands

```bash
export PATH=<go1.26-with-covdata>/bin:$PATH GOTOOLCHAIN=local
make check-format vet lint      # gofmt, go vet, golangci-lint (+gosec), actionlint, ci workflows
npm --prefix ui run typecheck   # tsc --noEmit (strict)
npm --prefix ui run lint        # ESLint
npm --prefix ui run format:check
make package                    # Go server binaries + UI bundle → dist/nulab-backlog-0.3.0.tar.gz
make verify-package             # package verification (id, version, checksums)
```

## Build Verification

- `make verify-package` prints `verifypkg: OK dist/nulab-backlog-0.3.0.tar.gz (nulab-backlog@0.3.0)`.
- `make contract-test KANDEV_MIN_DIR=../kandev` builds Kandev v0.96.0, installs the package and prints `ci contract: OK nulab-backlog on Kandev v0.96.0`.

## Troubleshooting

| Symptom | Cause | Fix |
|---|---|---|
| `go: no such tool "covdata"` on `make coverage` | Module-cache toolchain without `covdata` | Use a complete Go 1.26 toolchain with `GOTOOLCHAIN=local` |
| `contract-test: ../kandev is missing` / wrong HEAD | Sibling checkout missing or not at the tag | Link `../kandev` to a checkout at `v0.96.0` |
| 54 Vitest failures when running a scoped command from the repo root | Vitest started without the UI config | Add `--root ui` (see unit-test-instructions.md) |
| `coverage.out` at the repo root | Ad-hoc `-coverprofile` | Use `make coverage` (writes under `build/`) and delete stray profiles |
