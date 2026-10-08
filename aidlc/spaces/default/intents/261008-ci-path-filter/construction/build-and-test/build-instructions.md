# Build Instructions — CI path filter

## Prerequisites

- Go 1.26.x on `PATH` (locally: `export PATH=$HOME/.local/go/bin:$PATH`).
- Node.js from `.nvmrc`, with `npm ci` run in `ui/`.
- `../kandev` checked out at the commit in `.kandev-sdk-ref` (`make check-sdk`). For the contract test, a Kandev checkout at `v<min_kandev_version>` (locally `../kandev` is at `v0.96.0`, which is both).
- gcc (the contract test builds Kandev with CGO).

## Build and Check Commands

The same standard targets CI runs in the `checks` job:

```bash
make check-format vet lint test coverage check-secrets build package verify-package
```

Packaged-host contract test (CI job `packaged-host-contract`):

```bash
make contract-test KANDEV_MIN_DIR=../kandev
```

New CI entry points added by this change (no Kandev checkout needed; `internal/ci` does not import the SDK):

```bash
go run ./cmd/ci changes -base <40-hex base sha> -head <40-hex head sha>   # prints app=true|false
make -o check-sdk check-secrets                                           # what the secret-scan job runs
```

## Verification

- `make` exits 0; `ci workflows: OK`; `ci secrets: OK`; `verifypkg: OK`; coverage at or above 80%.
- `make lint` includes actionlint and the repo workflow policy over `.github/workflows/ci.yml` and the new `.github/workflows/secrets.yml`.

## Troubleshooting

- `check-sdk: ../kandev is missing` — check out kdlbs/kandev at `.kandev-sdk-ref` next to the repo.
- `go: command not found` — add the Go toolchain to `PATH`.
- A stray `coverage.out` at the repository root — delete it; `make coverage` writes under `build/`.
