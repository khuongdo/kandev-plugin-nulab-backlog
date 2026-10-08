# Reverse Engineering Timestamp — kandev-plugin-nulab-backlog

## Run Information

- Date: 2026-10-08
- Commit: `f5a7529baaa685e2d02e30cf879a247ef0dd106b` (v0.4.2; branch `feature/plugin-install-faile-9qm`)
- Intent: `261008-ci-path-filter` (depth Minimal)
- Type: FOCUSED SCAN merged into the existing store (prior verdict STALE; prior store built by `261007-plugin-install-502` as `kind: full`). Prior prose is preserved; sections on CI/release workflows, the Makefile, required checks and the app vs non-app path classification were updated. Per the STALE rule, this block records only this run; the prior `./` deep coverage is demoted to `shallow.paths`.
- Pre-scan snapshot: paths `.github/,Makefile`, store_generation `sha256:e0c900e8059aba039141269d61ddd08641452607633ea5d17ad9323c99832e87`, source_fingerprint `git:0a67930243050497755920e35d45db5302fb220f`.
- External evidence (outside the snapshot, not coverage): GitHub repository ruleset `24580280` and PR check history, read with `gh api` / `gh pr` on 2026-10-08.
- Baselines: not recorded (CI-configuration-only intent). See [code-quality-assessment.md](code-quality-assessment.md#test-coverage-and-baselines).

## Scope of Analysis

```yaml
scope_version: 1
kind: partial
intent: 261008-ci-path-filter
fingerprint: 0a67930243050497755920e35d45db5302fb220f
analyzed:
  paths:
    - .github/
    - Makefile
  components:
    - CI Workflows
    - Build Makefile
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
    - ui/tsconfig.json
    - ui/eslint.config.js
    - build/
    - dist/
    - manifest.yaml
    - go.mod
    - go.sum
    - README.md
    - LICENSE
    - .golangci.yml
    - .kandev-sdk-ref
    - .nvmrc
    - .gitignore
```
