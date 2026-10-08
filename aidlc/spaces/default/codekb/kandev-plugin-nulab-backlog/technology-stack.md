# Technology Stack — kandev-plugin-nulab-backlog

| Layer | Technology | Version |
|---|---|---|
| Backend language | Go | `go 1.26` (go.mod); Go 1.26.x at `~/.local/go` for this run's baseline |
| Plugin SDK | `github.com/kandev/kandev/pkg/pluginsdk` (`replace` to `../kandev/apps/backend`) | Kandev `v0.96.0`, commit `f099a46d` (`.kandev-sdk-ref`) |
| RPC (indirect) | hashicorp go-plugin / gRPC | via SDK |
| HTTP | Go stdlib `net/http` (no third-party HTTP client or provider SDK) | — |
| Process execution | Go stdlib `os/exec` (`internal/scm/cli_token.go` for `gh` / `glab`; `internal/ci/` for CI) | — |
| UI language | TypeScript strict, TSX with host-provided React (`host.jsx`) and UI kit (`hostUi(host)`) | TypeScript ~6.0.3, `@types/react` 19.3 |
| UI plugin SDK | `@kandev/plugin-sdk` (types) | same Kandev pin |
| UI build | esbuild; Node from `.nvmrc` | esbuild 0.28 |
| Tests | Go `testing` + testify `require` (`-race`), `httptest`; Vitest + jsdom + axe-core | testify v1.12.1; Vitest 5.0; jsdom 30; axe-core 4.14 |
| Lint / format | gofmt, go vet, golangci-lint + gosec, `tsc --noEmit`, ESLint 10 + typescript-eslint, Prettier 3.9, actionlint | pinned in `Makefile` / `ui/package.json` |
| Build | GNU Make | — |
| CI/CD | GitHub Actions (`ci.yml`, `release.yml`, `secrets.yml`), ruleset on `main` | — |

## Runtime Environment

- The plugin binary runs on the Kandev server as a child process and inherits its environment.
- Optional host tools for the SCM CLI method: `gh` (GitHub; `--user` needs >= 2.40, `auth status --json` needs >= 2.81.0) and `glab` (GitLab), logged in for the OS account that runs Kandev. Token-only use needs neither.
- Task worktrees and agent environments are prepared by Kandev's executor, not the plugin.

Library inventory: [dependencies.md](dependencies.md).
