# Reverse Engineering Timestamp — kandev-plugin-nulab-backlog

## Run Information

- Date: 2026-10-08
- Commit: `1819cc34dcd1fa89950d904d7f2c58c01b51cd95` (v0.4.1; branch `feature/plugin-install-faile-9qm`)
- Intent: `261007-plugin-install-502` (scope bugfix, depth Minimal)
- Type: FULL RESCAN by human decision (prior store verdict STALE). All 9 artifacts replaced wholesale; this block is built only from this run.
- Pre-scan snapshot: paths `./`, store_generation `sha256:3839bc6a51780effeb3a33171e5151e517523a62e7a0383a5ccd104319cad241`, source_fingerprint `git:db8a8ceecd7a0f03e0662036d298e3a1a25be9cf`.
- Depth note: build, packaging, manifest, verifier and CI paths were read deeply; domain packages and `ui/src/` were skimmed (listed under `shallow.paths`).
- External evidence (outside the snapshot, not coverage): Kandev `v0.96.0` checkout `~/repo/kandev`, running Kandev v0.97.0 logs and config, `gh release view v0.4.1`, throttled upload reproduction.
- Baselines: not recorded (no Go toolchain, no `../kandev` link). See [code-quality-assessment.md](code-quality-assessment.md#test-coverage-and-baselines).

## Scope of Analysis

```yaml
scope_version: 1
kind: full
intent: 261007-plugin-install-502
fingerprint: db8a8ceecd7a0f03e0662036d298e3a1a25be9cf
analyzed:
  paths:
    - ./
  components:
    - Server Entrypoint
    - KandevAdapter
    - BacklogGateway
    - Connection
    - Issues
    - Git
    - SCM
    - SCM Clients
    - Redact
    - PackageVerify
    - CI Tooling
    - TestUtil
    - UI Bundle
shallow:
  paths:
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
    - cmd/ci/
    - ui/src/
    - ui/tsconfig.json
    - ui/eslint.config.js
    - docs/
    - LICENSE
```
