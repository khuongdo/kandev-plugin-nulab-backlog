# Dependencies — kandev-plugin-nulab-backlog

## External (Go, `go.mod`)

- `github.com/kandev/kandev` — `replace` to `../kandev/apps/backend`; build fails unless `../kandev` HEAD equals `.kandev-sdk-ref`.
- `github.com/stretchr/testify` v1.12.1 — tests.
- `gopkg.in/yaml.v3` v3.0.1 — manifest parsing in `pkgverify` and `ci`.
- Indirect via SDK: `hashicorp/go-plugin` v1.8.0, `google.golang.org/grpc` v1.83.1.

## External (UI dev, `ui/package.json`)

esbuild, typescript, vitest, jsdom, eslint + typescript-eslint, prettier, react/react-dom (types and tests only, never bundled), axe-core. Versions: [technology-stack.md](technology-stack.md).

## External Services

- Kandev host (install API, plugin RPC, state, secrets): [api-documentation.md](api-documentation.md).
- Backlog REST v2, GitHub, GitLab, Bitbucket REST.
- Packaging tool: Kandev `cmd/plugin-pack` from the sibling checkout.

## Internal (cross-package)

```
server -> plugin
plugin -> connection, issues, git, scm, backlog, redact, pluginsdk
connection -> backlog, redact
issues -> backlog, connection, redact
git -> backlog, connection, redact
scm -> github, gitlab, bitbucket
backlog -> redact
cmd/verifypkg -> pkgverify
cmd/ci -> ci
```

The graph is acyclic; only `plugin` and `server` import `pluginsdk`.
