# Reverse Engineering Timestamp — kandev-plugin-nulab-backlog

## Run Information

- Date: 2026-10-09
- Commit: `3803248f6aff958d9f4f837a7c8e7262208db862` (branch `feature/fix-kandev-has-no-wo-ik5`, on top of the v0.6.0 records on `main`)
- Intent: `261009-no-workflow-error` (scope bugfix, depth Minimal)
- Type: **FOCUSED scan, merged into a STALE store**. Sections covering the start-task flow, the "no workflow" notice, the issue row layout and the error-display styling were added or updated; prior prose outside that area was kept. The scope block records only this run's deep coverage; the previous run's deep paths (intent `261008-source-control-settings`, commit `3d6a080`) are demoted to `shallow` because v0.6.0 (PR #26) changed them and they were not re-read.
- Pre-scan snapshot: paths `ui/src/page/`, `ui/src/issues/`, `ui/src/messages/`, `internal/issues/`, `internal/plugin/`; store_generation `sha256:1a23213dcd0219b947c4b89e4f468a57a5f0932e90e18ba74630661df7a19008`; source_fingerprint `git:fd9d68ccaffb60b343acd0d871df9b3c358e4cbb`.
- Shallow reads outside the snapshot (evidence, not deep coverage): `ui/src/settings/issue-watch-dialog.tsx`, `ui/src/git/` watch forms and PR lists, `ui/src/layout.ts`, `ui/src/host-ui.ts`, `ui/src/testing/harness.ts`.
- Read-only host reference (outside the repo): Kandev checkout `~/repo/kandev` at v0.96.0 (`f099a46`, equal to `.kandev-sdk-ref`): `plugin-sdk/src/index.ts`, `web/lib/plugins/plugin-context-api.ts`, `web/src/spa-routes.tsx`, `web/hooks/use-workflows.ts`, `web/components/task-create-dialog-{types,computed,effects}.ts`, `web/components/integrations/change-request-list.tsx`.
- Baseline: **recorded** — Go `-race` for `internal/issues` and `internal/plugin` pass; Vitest `src/page src/issues` 14 files, 166 tests pass (see [code-quality-assessment.md](code-quality-assessment.md#test-coverage-and-baselines)).

## Scope of Analysis

```yaml
scope_version: 1
kind: partial
intent: 261009-no-workflow-error
fingerprint: e87a2413adc3bfaee913d59e3f873a379187afbf
analyzed:
  paths:
    - ui/src/page/start-task.tsx
    - ui/src/page/start-task.test.tsx
    - ui/src/page/BacklogPage.tsx
    - ui/src/issues/issues-page.tsx
    - ui/src/issues/link-task-dialog.tsx
    - ui/src/issues/task-menu.ts
    - ui/src/messages/en.ts
    - internal/issues/service.go
    - internal/issues/watch.go
    - internal/issues/watcher.go
    - internal/plugin/host_port.go
    - internal/plugin/issue_actions.go
  components:
    - Issues
    - KandevAdapter
    - UI Bundle
shallow:
  paths:
    - ui/src/settings/source-control-section.tsx
    - ui/src/settings/source-control-section.test.tsx
    - ui/src/settings/SettingsScreen.tsx
    - ui/src/settings/section-parts.tsx
    - ui/src/settings/state.ts
    - ui/src/settings/issue-watch-dialog.tsx
    - internal/scm/service.go
    - internal/scm/store.go
    - internal/scm/types.go
    - internal/scm/client.go
    - internal/scm/prs.go
    - internal/scm/queries.go
    - internal/scm/watcher.go
    - internal/scm/links.go
    - internal/scm/cli_token.go
    - internal/plugin/scm_actions.go
    - internal/plugin/runtime.go
    - internal/plugin/credential.go
    - internal/issues/types.go
    - internal/github/client.go
    - ui/src/git/watch-form.tsx
    - ui/src/git/scm-watch-form.tsx
    - ui/src/git/pr-list.tsx
    - ui/src/git/scm-pr-list.tsx
    - ui/src/layout.ts
    - ui/src/host-ui.ts
    - ui/src/testing/harness.ts
    - manifest.yaml
    - go.mod
    - .kandev-sdk-ref
    - internal/backlog/
    - internal/bitbucket/
    - internal/gitlab/
    - internal/github/
    - internal/ci/
    - cmd/ci/
    - cmd/verifypkg/
    - internal/pkgverify/
    - internal/connection/
    - internal/git/
    - internal/issues/
    - internal/scm/
    - internal/plugin/
    - internal/redact/
    - internal/testutil/
    - server/
    - ui/src/
    - ui/src/git/
    - ui/src/settings/
    - ui/package.json
    - .github/workflows/
    - Makefile
    - .golangci.yml
```
