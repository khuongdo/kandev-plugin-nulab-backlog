# Build Instructions — CLI login for GitHub and GitLab

## Prerequisites

- Go 1.26.x on PATH: `export PATH="$HOME/.local/go/bin:$PATH"`.
- `../kandev` (next to the repo) is the Kandev checkout at tag `v0.96.0` (`min_kandev_version`, SDK commit `f099a46` in `.kandev-sdk-ref`). In this worktree it is a symlink to `~/repo/kandev`.
- Node per `.nvmrc`; `npm ci --prefix ui` once if `ui/node_modules` is missing.
- No new dependency was added by this change (Go or npm). No new environment variable or config file.
- Runtime prerequisite of the new feature (not of the build): `gh` ≥ 2.17 and/or `glab` installed and logged in for the user that runs the Kandev server.

## Build and package

All commands from the repo root; they are the same Makefile targets CI runs.

```bash
make check-format   # gofmt + Prettier
make vet            # go vet
make lint           # golangci-lint (+gosec), tsc, ESLint, actionlint, workflow checks
make build          # plugin binaries for the four platforms
make package        # build + UI bundle + dist/nulab-backlog-<version>.tar.gz + checksums.txt
make verify-package # checks archive, checksums, id and version
```

## Build verification

- `make verify-package` prints `verifypkg: OK dist/nulab-backlog-<version>.tar.gz`.
- `git status` shows no `coverage.out` at the repo root (the coverage profile is written under `build/`).

## Troubleshooting

- `check-sdk` fails: `../kandev` is missing or not at `v0.96.0`; create the link / check out the tag.
- Vitest fails with `document is not defined`: it was started from the repo root; run it from `ui/` (`cd ui && npx vitest run ...`) so `ui/vitest.config.ts` is loaded.
- npm warnings about unknown user config (`manage-package-manager-versions`, `lockfile-include-tarball-url`) come from the local npm config and are harmless.
