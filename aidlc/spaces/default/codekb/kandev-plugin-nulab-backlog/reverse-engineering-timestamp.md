# Reverse Engineering Timestamp — kandev-plugin-nulab-backlog

## Run Information

- Date: 2026-10-08
- Commit: `3d6a080eafc466772c6640799ac42167963ea517` (branch `feature/refactor-source-cont-c9o`, after the v0.5.3 records on `main`)
- Intent: `261008-source-control-settings` (scope express, depth Minimal)
- Type: **FOCUSED scan, merged into a STALE store**. Sections covering the Source Control settings page and the SCM provider model were updated; prior prose outside that area was kept (and the shipped v0.5.3 gh-account facts were refreshed). The scope block records only this run's deep coverage; the previous run's deep paths (intent `261008-gh-cli-profile`, commit `ca8146c`) that were not re-read are demoted to `shallow`.
- Pre-scan snapshot: paths `ui/src/settings/`, `internal/scm/`, `internal/github/`, `internal/gitlab/`, `internal/bitbucket/`, `internal/plugin/`; store_generation `sha256:519bb55ef508a1dfff86b9d862f36c10aee826c3708fb11f25bb362050c44fc5`; source_fingerprint `git:3d6a080eafc466772c6640799ac42167963ea517`.
- Shallow reads outside the snapshot (evidence, not deep coverage): `ui/src/git/` provider-list consumers, `ui/src/issues/issue-prs.tsx`, `ui/src/page/start-task.tsx`, `ui/src/messages/en.ts` (`scm*` keys), `manifest.yaml`, `Makefile`, `ui/package.json`, `.github/workflows/`.
- Baseline: **recorded** — targeted Go packages `ok` with `-race`, Vitest `src/settings` 121 passing (see [code-quality-assessment.md](code-quality-assessment.md#test-coverage-and-baselines)).

## Scope of Analysis

```yaml
scope_version: 1
kind: partial
intent: 261008-source-control-settings
fingerprint: ba46bb96ee325ea30a4c75cc66cebbe2b3b4a9cc
analyzed:
  paths:
    - ui/src/settings/source-control-section.tsx
    - ui/src/settings/source-control-section.test.tsx
    - ui/src/settings/SettingsScreen.tsx
    - ui/src/settings/section-parts.tsx
    - ui/src/settings/state.ts
    - internal/scm/service.go
    - internal/scm/store.go
    - internal/scm/types.go
    - internal/scm/client.go
    - internal/scm/prs.go
    - internal/scm/queries.go
    - internal/scm/watcher.go
    - internal/scm/links.go
    - internal/plugin/scm_actions.go
  components:
    - SCM
    - KandevAdapter
    - UI Bundle
shallow:
  paths:
    - internal/scm/cli_token.go
    - internal/github/client.go
    - internal/plugin/runtime.go
    - internal/plugin/host_port.go
    - internal/plugin/credential.go
    - internal/issues/service.go
    - internal/issues/types.go
    - ui/src/page/start-task.tsx
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
    - ui/src/messages/en.ts
    - ui/package.json
    - .github/workflows/
    - Makefile
    - .golangci.yml
```
