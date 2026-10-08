# Technology Stack — kandev-plugin-nulab-backlog

| Layer | Technology | Version |
|---|---|---|
| Backend language | Go | `go 1.26.0` (go.mod); Go 1.26.8 used for the run 3 baseline |
| Plugin SDK | `github.com/kandev/kandev/pkg/pluginsdk` (local `replace` to `../kandev/apps/backend`) | Kandev `v0.96.0`, commit `f099a46dc7aab16f6ff5806cd29b2b480296303f` (`.kandev-sdk-ref`) |
| UI plugin SDK | `@kandev/plugin-sdk` (tsconfig path alias to `../../kandev/apps/packages/plugin-sdk/src/index.ts`), types only; actions via `host.api.invokeAction` | same Kandev pin |
| RPC (indirect) | `hashicorp/go-plugin` / gRPC | v1.8.0 / v1.83.1 |
| HTTP | Go stdlib `net/http` (team rule: no third-party HTTP client or provider SDK) | — |
| Process execution | Go stdlib `os/exec` (used today only in `internal/ci/changes.go`) | — |
| UI language | TypeScript (strict), TSX via `h` factory with host-provided React and host UI kit | ~6.0.3 |
| UI bundler | esbuild | ^0.28.2 |
| Node | `.nvmrc` | 22 |
| Tests | Go `testing` + testify (`-race`); Vitest + jsdom + axe | testify v1.12.1; vitest ^5.0.3, jsdom ^30.1.2 |
| Lint/format | gofmt, go vet, golangci-lint (+gosec), ESLint, Prettier, actionlint | golangci-lint v2.14.0, actionlint v1.7.12 (both via `go run`, pinned in `Makefile`); eslint ^10, prettier ^3.9 |
| Build | GNU Make (bash, `-eu -o pipefail`) | — |
| CI/CD | GitHub Actions | `.github/workflows/ci.yml`, `release.yml`, `secrets.yml` |
| Merge gating | GitHub repository ruleset on `main` | ruleset `24580280` |

## GitHub Actions (pinned by full SHA, recorded by run 1)

| Action | Version | SHA |
|---|---|---|
| actions/checkout | v7.0.1 | `3d3c42e5aac5ba805825da76410c181273ba90b1` |
| actions/setup-go | v7.0.0 | `b7ad1dad31e06c5925ef5d2fc7ad053ef454303e` |
| actions/setup-node | v7.0.0 | `820762786026740c76f36085b0efc47a31fe5020` |
| actions/upload-artifact | v7.0.1 | `043fb46d1a93c77aae656e7c1c64a875d1fc6a0a` |
| actions/download-artifact | v8.0.1 | `3e5f45b2cfb9172054b4087a40e8e0b5a5461e7c` |
| actions/attest-build-provenance (release only) | v4.2.2 | `4d101475d8b20a2381f78447822ac1eab6504dd8` |

## Runtime Environment

- Observed self-hosted install (intent 261007): Kandev v0.97.0, systemd user service `kandev --headless`, fronted by `tailscale serve`. Drift: runtime 0.97.0 vs SDK pin and `min_kandev_version` 0.96.0.
- The plugin binary inherits the Kandev process environment (`PATH`, `HOME`, `GH_TOKEN`, `GH_CONFIG_DIR`). The GitHub CLI (`gh`) is **not** a dependency today; whether it is installed and logged in on the Kandev host is unknown to the plugin. Kandev itself uses `gh auth token` for its own GitHub integration (external reference).

Library inventory: [dependencies.md](dependencies.md).
