# Build Instructions — Opt-in by default

## Prerequisites

- Go 1.26.x. On this machine it lives at `~/.local/go/bin` (not on the default `PATH`): `export PATH=$HOME/.local/go/bin:$HOME/go/bin:$PATH`.
- `../kandev` next to the repo, checked out at the pinned SDK commit (`.kandev-sdk-ref`, Kandev v0.96.0). `make check-sdk` verifies it.
- Node per `.nvmrc`; run `cd ui && npm ci` once.
- gcc (CGO) for `go test -race` and the contract test.

## Environment

No environment variables or config files are needed for build or tests. Tests use fakes and an `httptest` fake Backlog server; no real credentials.

## Build Commands

```bash
make check-format vet lint check-secrets   # gofmt, go vet, golangci-lint+gosec, tsc, ESLint, Prettier, actionlint, secret scan
make build                                  # cross-platform server binaries under build/server/
make package                                # build + ui-build + plugin-pack -> dist/nulab-backlog-<version>.tar.gz + checksums.txt
make verify-package                         # checksum, contents and manifest check
```

## Build Verification

- `make package` ends with `plugin-pack: wrote dist/nulab-backlog-<version>.tar.gz`.
- `make verify-package` prints `verifypkg: OK`.
- No `coverage.out` at the repo root (`make coverage` writes under `build/`).

## Troubleshooting

- `go: command not found`: add `~/.local/go/bin` to `PATH`.
- `check-sdk` fails: re-point `../kandev` at the v0.96.0 checkout.
- Contract test: the Makefile default is `KANDEV_MIN_DIR=../kandev-min`; pass `KANDEV_MIN_DIR=../kandev` locally.
