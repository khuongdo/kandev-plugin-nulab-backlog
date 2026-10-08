# Reverse Engineering Timestamp — kandev-plugin-nulab-backlog

## Run Information

- Latest run date: 2026-10-08
- The store has received three FOCUSED SCANS, each merged into a STALE store (the original store was built by `261007-plugin-install-502` as `kind: full`). Each run demotes the prior deep coverage to shallow, so the scope block below records **only Run 3**; prose from earlier runs is preserved.
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
  - Depth note: the developer also read `ui/package.json`, `ui/vitest.config.ts` and `ui/tsconfig.json`; they were outside the snapshot, so they were recorded as shallow.
  - External evidence (outside the snapshot, not coverage): Kandev `v0.96.0` checkout `~/repo/kandev` (HEAD `f099a46`): task-row slots, navigation destinations, plugin registry and host API.
  - Baselines: not recorded (no Go toolchain, no `../kandev` link, no `ui/node_modules`).
- Run 3 (261008-gh-cli-auth):
  - Date: 2026-10-08
  - Commit: `d3d17e508534650a495f566c7382cca867121fa6` (v0.5.0; branch `feature/th-m-auth-method-cho-0pe`)
  - Intent: `261008-gh-cli-auth` (scope express, depth Minimal): add GitHub CLI login (`gh auth token`) as a second way to connect the GitHub source-control provider, next to the personal access token.
  - Type: FOCUSED SCAN merged into a STALE store. Prose for SCM provider connection, the GitHub client, credential handling, the SCM plugin actions, the Source control settings card and the manifest `scm.*` entries was updated or added; other prose is preserved. Run 2's deep coverage is demoted to `shallow.paths`.
  - Pre-scan snapshot: paths `internal/scm/,internal/github/,internal/connection/,internal/git/,internal/plugin/,ui/src/settings/,manifest.yaml`, store_generation `sha256:584cda89d96e10dc00f0e7bdbb349875393270f93c9e4bb40cd7ab8c690b9575`, source_fingerprint `git:186e7a716ded2d01e4c46414618ebea1aeced744`.
  - Depth note: `internal/connection/` and `internal/git/` were inside the snapshot but only skimmed, so they are shallow. `ui/src/git/git-state.ts`, `ui/src/messages/en.ts`, `internal/redact/`, `internal/ci/changes.go` and `Makefile` were read outside the snapshot and are shallow.
  - External evidence (outside the repo, not coverage): Kandev `v0.96.0` checkout `~/repo/kandev` (HEAD `f099a46`): `pkg/pluginsdk/host.go` method list, plugin process spawn in `internal/plugins/runtime/manager.go`, `gh auth token` use in `internal/github/gh_accounts.go`.
  - Baselines: **recorded** — 709 Go tests passing (see [code-quality-assessment.md](code-quality-assessment.md#test-coverage-and-baselines)).

## Scope of Analysis

```yaml
scope_version: 1
kind: partial
intent: 261008-gh-cli-auth
fingerprint: 05ba6f4e57844516580ce283f766d26e0eb62638
analyzed:
  paths:
    - internal/scm/
    - internal/github/
    - internal/plugin/scm_actions.go
    - internal/plugin/runtime.go
    - internal/plugin/credential.go
    - internal/plugin/manifest_test.go
    - ui/src/settings/source-control-section.tsx
    - manifest.yaml
  components:
    - SCM
    - GitHub Client
shallow:
  paths:
    - ./
    - .github/
    - Makefile
    - aidlc/
    - docs/
    - .claude/
    - internal/
    - internal/plugin/
    - internal/plugin/issue_actions.go
    - internal/backlog/
    - internal/connection/
    - internal/connection/service.go
    - internal/issues/
    - internal/issues/service.go
    - internal/issues/types.go
    - internal/issues/sync.go
    - internal/git/
    - internal/gitlab/
    - internal/bitbucket/
    - internal/redact/
    - internal/testutil/
    - internal/ci/changes.go
    - server/
    - cmd/
    - cmd/ci/
    - ui/
    - ui/src/
    - ui/src/index.ts
    - ui/src/index.test.ts
    - ui/src/issues/issue-badge.tsx
    - ui/src/issues/issue-badge.test.tsx
    - ui/src/issues/issues-state.ts
    - ui/src/issues/links-store.ts
    - ui/src/issues/i18n.ts
    - ui/src/switch/enabled-events.ts
    - ui/src/switch/integration-switch.tsx
    - ui/src/settings/SettingsScreen.tsx
    - ui/src/settings/issue-watches-section.tsx
    - ui/src/settings/pr-watches-section.tsx
    - ui/src/settings/section-parts.tsx
    - ui/src/settings/sections.test.tsx
    - ui/src/settings/source-control-section.test.tsx
    - ui/src/git/git-state.ts
    - ui/src/messages/en.ts
    - ui/src/host-ui.ts
    - ui/src/testing/harness.ts
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
