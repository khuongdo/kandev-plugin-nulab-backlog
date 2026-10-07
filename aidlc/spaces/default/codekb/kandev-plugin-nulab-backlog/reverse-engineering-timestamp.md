# Reverse Engineering Timestamp — kandev-plugin-nulab-backlog

## Run Information

- Date: 2026-10-07
- Commit: `86ae4737ab535ed40462bf58ea880f8a7fe41af3` (branch `feature/refactor-uiux-2bi`)
- Intent: `261007-uiux-github-style` (scope refactor, depth Minimal)
- Type: FULL RESCAN (replaces the previous store from the same intent's earlier attempt)
- External reference (not in scope): `/home/k_do_webfrontier/repo/kandev`, tag `v0.96.0`, commit `f099a46dc7aab16f6ff5806cd29b2b480296303f`
- Baselines: Go `go test -race` all 9 test packages pass, total 92.9% (go1.26.8, `../kandev` at `v0.96.0`); UI Vitest 229/229, `tsc`, ESLint, Prettier clean. Details: [code-quality-assessment.md](code-quality-assessment.md#test-coverage-and-baselines).

## Scope of Analysis

```yaml
scope_version: 1
kind: full
intent: 261007-uiux-github-style
fingerprint: b08ebb70423240a1b303c00058a8fce7c052a466
analyzed:
  paths:
    - ./
    - manifest.yaml
    - Makefile
    - go.mod
    - .kandev-sdk-ref
    - .nvmrc
    - .gitignore
    - server/main.go
    - internal/plugin/
    - internal/git/types.go
    - internal/git/service.go
    - internal/git/watcher.go
    - ui/package.json
    - ui/tsconfig.json
    - ui/vitest.config.ts
    - ui/eslint.config.js
    - ui/.prettierrc
    - ui/src/index.ts
    - ui/src/index.test.ts
    - ui/src/jsx.d.ts
    - ui/src/brand/
    - ui/src/page/
    - ui/src/settings/
    - ui/src/git/
    - ui/src/issues/
    - ui/src/switch/
    - ui/src/testing/harness.ts
    - ui/src/messages/en.ts
    - docs/brand/backlog-logo.md
  components:
    - Server Entrypoint
    - KandevAdapter
    - BacklogGateway
    - Connection
    - Issues
    - Git
    - Redact
    - CI Tooling
    - PackageVerify
    - TestUtil
    - UI Registration
    - UI Brand
    - UI Page
    - UI Settings
    - UI Issues
    - UI Git
    - UI Switch
    - UI Messages
    - UI Test Harness
shallow:
  paths:
    - internal/backlog/
    - internal/connection/
    - internal/issues/
    - internal/git/
    - internal/redact/
    - internal/testutil/
    - internal/ci/
    - cmd/ci/
    - internal/pkgverify/
    - cmd/verifypkg/
    - .github/workflows/
    - .golangci.yml
    - README.md
    - docs/manual-checks/
```
