# Build Instructions — Link Task modal, GitHub-style

## Prerequisites

- Go 1.26.x on `PATH`. Here Go 1.26.8 was installed to `~/.local/go` from go.dev with its SHA-256 checksum verified. CGO and `gcc` are needed for `-race`.
- Node (the CI version; v25.2.1 works locally with an engine warning) and `npm`.
- `../kandev` must point to the Kandev checkout at `v0.96.0`. Here it is a symlink to `~/repo/kandev`. Both the `make` targets (`check-sdk`) and the UI SDK types need it.

## Environment Setup

```bash
export PATH=$HOME/.local/go/bin:$PATH
ln -s ~/repo/kandev ../kandev        # only if absent
cd ui && npm ci && cd ..
```

No environment variables or local services are needed. Tests use fakes; no real Backlog space or credentials.

## Build Commands

```bash
cd ui && npm run build      # bundles ui/src/index.ts to build/ui/bundle.js (build/ is git-ignored)
make vet                    # go vet ./...
```

The full package and verification (`make package`, `make verify-package`) are unchanged by this UI-only bugfix and run in CI.

## Build Verification

- `npm run build` prints `../build/ui/bundle.js 226.9kb` and `Done`.
- `make vet` exits 0 with no findings.

## Troubleshooting

- `check-sdk` fails: `../kandev` is missing or at the wrong tag. Fix the symlink.
- `-race` fails with a cgo error: install `gcc` and keep `CGO_ENABLED=1`.
- `npm warn Unknown user config ...`: harmless local npm config warnings.
