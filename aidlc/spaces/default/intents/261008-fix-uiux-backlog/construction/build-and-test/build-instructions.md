# Build Instructions - Fix UIUX (Backlog task badge, top bar, settings)

## Prerequisites

- Go 1.26.x on `PATH` (this machine: `~/.local/go/bin`, go1.26.8), with CGO for `-race`.
- Node.js and npm (this machine: Node v25.2.1); run `npm ci` in `ui/`.
- `../kandev` is a checkout of Kandev at `.kandev-sdk-ref` (f099a46, tag `v0.96.0`); here it is a symlink to `~/repo/kandev`. `make check-sdk` verifies it.
- gcc (only for `make contract-test`).

## Environment

No environment variables or secrets are needed to build or test. Tests use an `httptest` fake Backlog and the fake Kandev host in `ui/src/testing/harness.ts`.

## Build Commands

```bash
export PATH=$HOME/.local/go/bin:$PATH
make check-sdk                 # ../kandev at the pinned SDK commit
make check-format vet lint     # gofmt, go vet, golangci-lint+gosec, prettier, tsc, eslint, actionlint
make test                      # go test -race ./... and npx vitest run
make coverage                  # 80% floor over ./internal/... and ./server/... (profile under build/)
make check-secrets             # no credentials in the repo
make build package verify-package   # UI bundle + 4 platform binaries, dist/nulab-backlog-<version>.tar.gz, checksum + manifest check
make contract-test KANDEV_MIN_DIR=../kandev   # install and run the package on Kandev v0.96.0
```

## Build Verification

- `verifypkg: OK dist/nulab-backlog-0.4.2.tar.gz (nulab-backlog@0.4.2)` (the version bump belongs to Deployment Pipeline).
- `ci contract: OK nulab-backlog on Kandev v0.96.0`.
- `git status` shows no `coverage.out` at the repo root.

## Troubleshooting

- `check-sdk: ../kandev is missing`: link or check out Kandev v0.96.0 next to the repo.
- `go: command not found`: add `~/.local/go/bin` (or your Go 1.26 install) to `PATH`.
- `-race` needs CGO (`CGO_ENABLED=1`, gcc present).
- Vitest cannot resolve `@kandev/plugin-sdk`: the `ui/tsconfig.json` path alias points into `../kandev`; fix the checkout first.
