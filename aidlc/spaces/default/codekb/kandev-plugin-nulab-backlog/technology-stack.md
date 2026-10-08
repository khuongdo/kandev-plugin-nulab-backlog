# Technology Stack — kandev-plugin-nulab-backlog

| Layer | Technology | Version |
|---|---|---|
| Backend language | Go | `go 1.26` (go.mod); Go 1.26.8 used for this run's baseline |
| Plugin SDK | `github.com/kandev/kandev/pkg/pluginsdk` (`replace` to `../kandev/apps/backend`) | Kandev `v0.96.0`, commit `f099a46dc` (`.kandev-sdk-ref`) |
| RPC (indirect) | hashicorp go-plugin / gRPC | via SDK |
| HTTP | Go stdlib `net/http` (no third-party HTTP client or provider SDK) | — |
| Process execution | Go stdlib `os/exec` (`internal/scm/cli_token.go` for `gh` / `glab`; `internal/ci/changes.go` for CI) | — |
| UI language | TypeScript strict, TSX with host-provided React and UI kit | — |
| UI plugin SDK | `@kandev/plugin-sdk` (types; `host.api.invokeAction`, `TaskCreateDialog`, `IntegrationStartTaskMenu`) | same Kandev pin |
| UI build | esbuild; Node from `.nvmrc` | — |
| Tests | Go `testing` + testify (`-race` in CI), `httptest`; Vitest | testify v1.12.1; vitest ^5.0.3 |
| Lint / format | gofmt, go vet, golangci-lint + gosec, `tsc --noEmit`, ESLint, Prettier, actionlint | pinned in `Makefile` / `ui/package.json` |
| Build | GNU Make | — |
| CI/CD | GitHub Actions (`ci.yml`, `release.yml`, `secrets.yml`), ruleset on `main` | — |

## Runtime Environment

- The plugin binary runs on the Kandev server as a child process and inherits its environment.
- Optional host tools for the SCM CLI method: `gh` (GitHub) and `glab` (GitLab), installed and logged in for the OS account that runs Kandev. `gh` 2.97.0 on the scan host (supports `gh auth token --user` and `gh auth status --json hosts`). Token-only use needs neither.
- Task worktrees and agent environments are prepared by Kandev's executor, not the plugin.

Library inventory: [dependencies.md](dependencies.md).
