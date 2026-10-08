# Reverse Engineering Timestamp — kandev-plugin-nulab-backlog

## Run Information

- Date: 2026-10-08
- Commit: `ca8146ca529f0bf6caa4604cf44239c3f555eda5` (v0.5.2 records on `main`; branch `feature/gh-cli-profile-scope-q1o`)
- Intent: `261008-gh-cli-profile` (scope express, depth Minimal)
- Type: **FULL RESCAN** of `./` at Minimal depth. All 9 artifacts were replaced; the scope block below records only this run. Prior focused-run history lives in the earlier intent records.
- Pre-scan snapshot: paths `./`, store_generation `sha256:36ae6e8147b7ca6e24623d66b33daf14aeb8910cc6c15d9f4c94478df2376e4a`, source_fingerprint `git:f9ee5e321fa96ae433d8f38544b5034c3b896c33`.
- Depth note: the whole repo was in the snapshot, but only the files listed under `analyzed.paths` were read deeply (intent area: gh CLI credential, per-workspace SCM settings, task creation from Backlog issues). Everything else was skimmed, so the block is `kind: partial`.
- External evidence (not coverage): Kandev checkout `~/repo/kandev` at `v0.96.0` (`f099a46dc`): `internal/github/gh_accounts.go`, `internal/github/auth_resolver.go`, `internal/orchestrator/executor/executor_credentials.go`, `pkg/pluginsdk/{host.go,data_types.go,plugin.go,types.go}`.
- Baseline: **recorded** — 1383 Go test results passing, 0 failures (see [code-quality-assessment.md](code-quality-assessment.md#test-coverage-and-baselines)).

## Scope of Analysis

```yaml
scope_version: 1
kind: partial
intent: 261008-gh-cli-profile
fingerprint: 0579d12edc9ff08dd841a3b17d4aac2f68a48f5e
analyzed:
  paths:
    - internal/scm/cli_token.go
    - internal/scm/service.go
    - internal/scm/types.go
    - internal/scm/store.go
    - internal/scm/client.go
    - internal/github/client.go
    - internal/plugin/scm_actions.go
    - internal/plugin/runtime.go
    - internal/plugin/host_port.go
    - internal/plugin/credential.go
    - internal/issues/service.go
    - internal/issues/types.go
    - ui/src/settings/source-control-section.tsx
    - ui/src/page/start-task.tsx
    - manifest.yaml
    - go.mod
    - .kandev-sdk-ref
  components:
    - SCM
    - GitHub Client
    - KandevAdapter
    - Issues
    - UI Bundle
shallow:
  paths:
    - internal/backlog/
    - internal/bitbucket/
    - internal/gitlab/
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
    - .github/workflows/
    - Makefile
    - .golangci.yml
```
