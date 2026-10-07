# Technology Stack — kandev-plugin-nulab-backlog

| Layer | Technology | Version |
|---|---|---|
| Backend language | Go | `go 1.26.0` (go.mod) |
| Plugin SDK | `github.com/kandev/kandev/pkg/pluginsdk` (local `replace` to `../kandev/apps/backend`) | Kandev `v0.96.0`, commit `f099a46dc7aab16f6ff5806cd29b2b480296303f` (`.kandev-sdk-ref`) |
| RPC (indirect) | `hashicorp/go-plugin` / gRPC | v1.8.0 / v1.83.1 |
| HTTP | Go stdlib `net/http` | — |
| UI language | TypeScript (strict), TSX with host-provided React | ~6.0.3 |
| UI bundler | esbuild | ^0.28.2 |
| Node | `.nvmrc` | 22 |
| Tests | Go `testing` + testify; Vitest + jsdom | testify v1.12.1; vitest ^5.0.3 |
| Lint/format | gofmt, go vet, golangci-lint (+gosec), ESLint, Prettier, actionlint | golangci-lint v2.14.0, actionlint v1.7.12 |
| CI/CD | GitHub Actions | `.github/workflows/ci.yml`, `release.yml` |

## Runtime Environment (observed, self-hosted)

- Kandev v0.97.0, systemd user service `kandev --headless` on `:38429`.
- Fronted by `tailscale serve` `https://webfrontier.tail152aaa.ts.net` -> `http://localhost:38429`.
- Drift: runtime 0.97.0 vs SDK pin and `min_kandev_version` 0.96.0.

Library inventory: [dependencies.md](dependencies.md).
