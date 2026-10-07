# Build Instructions — Kandev plugin for Nulab Backlog

## Prerequisites

| Tool | Version | Notes |
|------|---------|-------|
| Go | 1.26.x | `GOTOOLCHAIN=local`. On this machine the toolchain is `~/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.26.8.linux-amd64/bin`; put it first on `PATH`. CGO is needed for `-race` (gcc). |
| Node.js / npm | Node 20+ (v25 used locally) | UI lives in `ui/`; install with `npm ci`. |
| Kandev SDK checkout | `../kandev` at the commit in `.kandev-sdk-ref` (v0.96.0) | `go.mod` has `replace` to `../kandev`; `make check-sdk` verifies the pin. |
| Kandev minimum-version checkout | `../kandev-min` at tag `v0.96.0` | Only for `make contract-test`. |
| golangci-lint | v2.14.0 (pinned in the Makefile, run with `go run`) | Includes `gosec`. |
| actionlint | pinned in the Makefile | Workflow lint. |

## Environment setup

```
export PATH=$HOME/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.26.8.linux-amd64/bin:$PATH
export GOTOOLCHAIN=local
(cd ui && npm ci)
```

No environment variables, secrets or local services are needed to build or test. Tests never call real Backlog or real Kandev (httptest fakes and an in-memory host). Never put a real API key, token or password in the repo or in test data.

## Build commands

All commands run from the repository root; CI calls exactly these targets.

```
make clean
make check-format vet lint          # gofmt, go vet, golangci-lint+gosec, tsc --noEmit, ESLint, Prettier, actionlint
make test                           # go test -race ./... and npx vitest run
make coverage                       # go test -race -coverprofile; floor 80%, only server/main.go excluded
make check-secrets                  # repository secret scan (go run ./cmd/ci secrets -root .)
make build                          # 5 executables: linux/darwin amd64+arm64, windows amd64; UI bundle via esbuild
make package                        # dist/nulab-backlog-<version>.tar.gz + dist/checksums.txt (Kandev plugin-pack)
make verify-package                 # go run ./cmd/verifypkg against the archive and checksum
make contract-test KANDEV_MIN_DIR=../kandev-min   # builds a throwaway Kandev v0.96.0, installs and runs the package
```

The single gate used in this stage:

```
make clean check-format vet lint test coverage check-secrets build package verify-package
make contract-test KANDEV_MIN_DIR=../kandev-min
```

## Build verification

- `make coverage` prints `coverage: NN.N% (floor 80%, excluded: server/main.go)` and fails below 80%.
- `make verify-package` prints `verifypkg: OK dist/nulab-backlog-<version>.tar.gz (nulab-backlog@<version>)`.
- `make contract-test` prints `ci contract: OK nulab-backlog on Kandev v0.96.0`.
- `go mod tidy` must leave `go.mod`/`go.sum` unchanged (CI checks this).
- `make coverage` writes `coverage.out` at the repo root (git-ignored). Delete it after local runs inside an AI-DLC Code Generation attempt; the workflow treats it as an unclaimed source change.

## Troubleshooting

| Symptom | Cause | Fix |
|---------|-------|-----|
| `gofmt: command not found` / wrong Go version | Go 1.26 toolchain not on `PATH` | Export the `PATH` and `GOTOOLCHAIN=local` above. |
| `check-sdk` fails | `../kandev` not at `.kandev-sdk-ref` | `git -C ../kandev checkout v0.96.0`. |
| `-race` build error about cgo | No C compiler | Install gcc. |
| `contract-test` port in use | Ports 38529/39529 busy | Stop the process using them. Never touch the user's real Kandev on 38429/39429. |
| npm warnings about unknown config keys | User npm config | Harmless. |
