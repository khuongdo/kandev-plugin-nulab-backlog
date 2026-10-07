# Reverse Engineering Timestamp — kandev-plugin-nulab-backlog

## Run Information

- Date: 2026-10-07
- Commit: `2b4325fcbd0c14ae0beb23603c6249849fc0ddda` (branch `feature/add-default-queries-87j`)
- Intent: `261007-github-parity-actions` (scope express, depth Minimal)
- Type: FOCUSED SCAN merged into a STALE store (prior store: intent `261007-uiux-github-style`, full rescan at `86ae473`). Snapshot paths: `ui/src/`, `internal/plugin/`, `internal/git/`, `internal/issues/`, `manifest.yaml`. Prose outside the focus area (gateway, connection, tooling, dependencies) is preserved from the prior store; the prior deep coverage is demoted to shallow because it could not be re-verified.
- External reference (not in scope): `/home/k_do_webfrontier/repo/kandev`, tag `v0.96.0`, commit `f099a46dc7aab16f6ff5806cd29b2b480296303f`
- Baselines: Go `go test -race` all 9 test packages pass, total 92.7%; UI Vitest 286/286, `tsc`, ESLint, Prettier clean. Details: [code-quality-assessment.md](code-quality-assessment.md#test-coverage-and-baselines).

## Scope of Analysis

```yaml
scope_version: 1
kind: partial
intent: 261007-github-parity-actions
fingerprint: 45039258deef5df0bf71376f6f420a5208512146
analyzed:
  paths:
    - manifest.yaml
    - ui/src/index.ts
    - ui/src/layout.ts
    - ui/src/host-ui.ts
    - ui/src/icons.tsx
    - ui/src/jsx.d.ts
    - ui/src/page/BacklogPage.tsx
    - ui/src/issues/issues-page.tsx
    - ui/src/issues/issues-state.ts
    - ui/src/issues/task-menu.ts
    - ui/src/issues/i18n.ts
    - ui/src/git/pr-list.tsx
    - ui/src/git/pr-toolbar.tsx
    - ui/src/git/save-query-dialog.tsx
    - ui/src/git/git-state.ts
    - ui/src/settings/saved-queries-section.tsx
    - ui/src/settings/section-parts.tsx
    - ui/src/settings/use-list.ts
    - ui/src/settings/SettingsScreen.tsx
    - internal/plugin/runtime.go
    - internal/plugin/issue_actions.go
    - internal/plugin/git_actions.go
    - internal/plugin/host_port.go
    - internal/issues/types.go
    - internal/issues/service.go
    - internal/issues/store.go
    - internal/issues/watch.go
    - internal/git/types.go
    - internal/git/service.go
    - internal/git/store.go
    - internal/git/watcher.go
  components:
    - KandevAdapter
    - Issues
    - Git
    - UI Registration
    - UI Shared Kit
    - UI Page
    - UI Settings
    - UI Issues
    - UI Git
shallow:
  paths:
    - Makefile
    - go.mod
    - .kandev-sdk-ref
    - .nvmrc
    - .gitignore
    - server/main.go
    - internal/plugin/
    - internal/git/
    - internal/issues/
    - ui/package.json
    - ui/tsconfig.json
    - ui/vitest.config.ts
    - ui/eslint.config.js
    - ui/.prettierrc
    - ui/src/index.test.ts
    - ui/src/brand/
    - ui/src/page/
    - ui/src/settings/
    - ui/src/git/
    - ui/src/issues/
    - ui/src/switch/
    - ui/src/testing/harness.ts
    - ui/src/messages/en.ts
    - docs/brand/backlog-logo.md
    - internal/backlog/
    - internal/connection/
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
