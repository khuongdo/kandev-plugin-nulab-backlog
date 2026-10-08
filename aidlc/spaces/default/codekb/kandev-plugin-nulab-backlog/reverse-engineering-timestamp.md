# Reverse Engineering Timestamp — kandev-plugin-nulab-backlog

## Run Information

- Date: 2026-10-08
- Two FOCUSED SCANS were merged into the same STALE store (prior store built by `261007-plugin-install-502` as `kind: full`). The scope block below is their union; the fingerprint is not re-minted, so the next intent sees the store as STALE and rescans.
- Run 1 (261008-ci-path-filter):
  - Commit: `f5a7529baaa685e2d02e30cf879a247ef0dd106b` (v0.4.2; branch `feature/plugin-install-faile-9qm`)
  - Intent: `261008-ci-path-filter` (depth Minimal)
  - Type: FOCUSED SCAN merged into the existing store (prior verdict STALE). Prior prose is preserved; sections on CI/release workflows, the Makefile, required checks and the app vs non-app path classification were updated. Per the STALE rule, the prior `./` deep coverage is demoted to `shallow.paths`.
  - Pre-scan snapshot: paths `.github/,Makefile`, store_generation `sha256:e0c900e8059aba039141269d61ddd08641452607633ea5d17ad9323c99832e87`, source_fingerprint `git:0a67930243050497755920e35d45db5302fb220f`.
  - External evidence (outside the snapshot, not coverage): GitHub repository ruleset `24580280` and PR check history, read with `gh api` / `gh pr` on 2026-10-08.
  - Baselines: not recorded (CI-configuration-only intent). See [code-quality-assessment.md](code-quality-assessment.md#test-coverage-and-baselines).
- Run 2 (261008-fix-uiux-backlog):
  - Commit: `3a983aba8ef37d8ac216ccca5d6ce707bce68fe2` (v0.4.2; branch `feature/fix-uiux-i41`)
  - Intent: `261008-fix-uiux-backlog` (scope express, depth Minimal)
  - Type: FOCUSED SCAN merged into a STALE store. Prior prose preserved outside the newly analyzed area; the prior `./` deep coverage is demoted to `shallow.paths` because it could not be re-verified.
  - Pre-scan snapshot: paths `ui/src/,internal/plugin/,internal/issues/,manifest.yaml`, store_generation `sha256:e0c900e8059aba039141269d61ddd08641452607633ea5d17ad9323c99832e87`, source_fingerprint `git:a17ba3672510156b6adc01a57b753172142b9076`.
  - Depth note: `analyzed.paths` lists only the files read deeply inside the snapshot paths. The developer also read `ui/package.json`, `ui/vitest.config.ts` and `ui/tsconfig.json`; they are outside the snapshot, so they are recorded as shallow.
  - External evidence (outside the snapshot, not coverage): Kandev `v0.96.0` checkout `~/repo/kandev` (HEAD `f099a46`): task-row slots, navigation destinations, plugin registry and host API.
  - Baselines: not recorded (no Go toolchain, no `../kandev` link, no `ui/node_modules`). See [code-quality-assessment.md](code-quality-assessment.md#test-coverage-and-baselines).

## Scope of Analysis

```yaml
scope_version: 1
kind: partial
intent: 261008-fix-uiux-backlog
fingerprint: bf70f7c7b814e501f1c28abc8f2bd79d91bd1f37
analyzed:
  paths:
    - .github/
    - Makefile
    - ui/src/index.ts
    - ui/src/index.test.ts
    - ui/src/issues/issue-badge.tsx
    - ui/src/issues/issues-state.ts
    - ui/src/issues/links-store.ts
    - ui/src/issues/i18n.ts
    - ui/src/switch/enabled-events.ts
    - ui/src/switch/integration-switch.tsx
    - ui/src/settings/SettingsScreen.tsx
    - ui/src/settings/issue-watches-section.tsx
    - ui/src/settings/section-parts.tsx
    - ui/src/settings/sections.test.tsx
    - ui/src/messages/en.ts
    - ui/src/host-ui.ts
    - internal/issues/service.go
    - internal/issues/types.go
    - internal/issues/sync.go
    - internal/plugin/issue_actions.go
    - manifest.yaml
  components:
    - CI Workflows
    - Build Makefile
    - UI Bundle
    - Issues
shallow:
  paths:
    - ./
    - aidlc/
    - docs/
    - .claude/
    - internal/
    - internal/plugin/
    - internal/backlog/
    - internal/connection/
    - internal/issues/
    - internal/git/
    - internal/scm/
    - internal/github/
    - internal/gitlab/
    - internal/bitbucket/
    - internal/redact/
    - internal/testutil/
    - server/
    - cmd/
    - cmd/ci/
    - ui/
    - ui/src/
    - ui/src/settings/pr-watches-section.tsx
    - ui/src/settings/source-control-section.tsx
    - ui/src/issues/issue-badge.test.tsx
    - ui/src/testing/harness.ts
    - internal/plugin/runtime.go
    - internal/connection/service.go
    - ui/package.json
    - ui/vitest.config.ts
    - ui/tsconfig.json
    - ui/eslint.config.js
    - build/
    - dist/
    - go.mod
    - go.sum
    - README.md
    - LICENSE
    - .golangci.yml
    - .kandev-sdk-ref
    - .nvmrc
    - .gitignore
```
