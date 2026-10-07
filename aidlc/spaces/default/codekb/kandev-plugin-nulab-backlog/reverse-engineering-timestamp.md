# Reverse Engineering Timestamp — kandev-plugin-nulab-backlog

## Run Information

- Date: 2026-10-07
- Commit: `5bf88b95f6317231d83dd14c811a7057c1d7a5fd` (v0.3.0; branch `feature/source-control-agnos-2jr`)
- Intent: `261007-source-control-agnostic` (scope express, depth Minimal)
- Type: FOCUSED SCAN merged into a STALE store (prior store: intent `261007-github-parity-actions`, focused scan at `2b4325f`). Snapshot paths: `internal/git`, `internal/connection`, `internal/plugin`, `internal/backlog`, `ui/src/git`, `ui/src/settings`, `manifest.yaml`. `internal/backlog/` was only skimmed. Analyzed paths are written exactly as the snapshot paths (no trailing slash) so the publish coverage check matches them literally; each is a directory except `manifest.yaml`. Prose outside the focus area is preserved from the prior store; the prior deep coverage is demoted to shallow because it could not be re-verified.
- External reference (not in scope, read only for extension points): `/home/k_do_webfrontier/repo/kandev`, tag `v0.96.0`, commit `f099a46dc7aab16f6ff5806cd29b2b480296303f`.
- Baselines: not recorded this run (no Go toolchain, no `ui/node_modules`, no `../kandev` link). Details: [code-quality-assessment.md](code-quality-assessment.md#test-coverage-and-baselines).

## Scope of Analysis

```yaml
scope_version: 1
kind: partial
intent: 261007-source-control-agnostic
fingerprint: 7b88df716ef7d54ee9aa69f6fbde15a27bd68711
analyzed:
  paths:
    - internal/git
    - internal/connection
    - internal/plugin
    - ui/src/git
    - ui/src/settings
    - manifest.yaml
  components:
    - KandevAdapter
    - Connection
    - Git
    - UI Git
    - UI Settings
shallow:
  paths:
    - internal/backlog/
    - internal/issues/
    - ui/src/index.ts
    - ui/src/layout.ts
    - ui/src/host-ui.ts
    - ui/src/icons.tsx
    - ui/src/jsx.d.ts
    - ui/src/page/
    - ui/src/issues/
    - ui/src/switch/
    - ui/src/brand/
    - ui/src/index.test.ts
    - ui/src/testing/harness.ts
    - ui/src/messages/en.ts
    - ui/package.json
    - ui/tsconfig.json
    - ui/vitest.config.ts
    - ui/eslint.config.js
    - ui/.prettierrc
    - Makefile
    - go.mod
    - .kandev-sdk-ref
    - .nvmrc
    - .gitignore
    - server/main.go
    - internal/redact/
    - internal/testutil/
    - internal/ci/
    - cmd/ci/
    - internal/pkgverify/
    - cmd/verifypkg/
    - .github/workflows/
    - .golangci.yml
    - README.md
    - docs/brand/backlog-logo.md
    - docs/manual-checks/
```
