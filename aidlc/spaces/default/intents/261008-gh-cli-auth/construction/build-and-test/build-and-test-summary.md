# Build and Test Summary — CLI login for GitHub and GitLab

## Build Status

Build, package and package verification succeed (`dist/nulab-backlog-0.5.0.tar.gz`, `verifypkg: OK`). Prerequisites: Go 1.26.x, `../kandev` at `v0.96.0`, `ui/node_modules`. See [build-instructions.md](build-instructions.md).

## Test Type Inventory

- Unit tests (Go + Vitest) written in Code Generation — see `../code-generation/unit-test-instructions.md`.
- Integration: existing packaged-host contract test (10 runs) and manifest parity test — [integration-test-instructions.md](integration-test-instructions.md).
- Security: leak/redaction tests, gosec, secret scan — [security-test-instructions.md](security-test-instructions.md).
- Performance: no load test (Minimal strategy); the cache bound is a unit test — [performance-test-instructions.md](performance-test-instructions.md).

## Coverage Expectations

80% line coverage floor for `./internal/...` and `./server/...` (only `server/main.go` excluded). Actual 92.9%.

## Target Verification Matrix

| Target ID | Source | Expected | Actual | Evidence | Owning Stage | Verdict |
|---|---|---|---|---|---|---|
| TC-COV | Testing Contract (team) | ≥ 80% | 92.9% | `make coverage` | build-and-test | Met |
| TC-RACE | Testing Contract (team) | `-race` on all Go tests | yes | Makefile `test`, unit commands | build-and-test | Met |
| TC-CONTRACT | Testing Contract (team) + project rule | 10/10 on Kandev v0.96.0 | 10/10 | `make contract-test` × 10 | build-and-test | Met |
| TC-LEAK | Testing Contract (team) + NFR1 | leak test passes | passes | security command | build-and-test | Met |
| TC-SUITE | Testing Contract (scope floor) | existing suite green | green | `make test` | build-and-test | Met |
| TC-MINIMAL | Testing Contract (strategy) | ≥ 1 test per requirement | all 29 IDs traced | cross-unit-traceability.md | build-and-test | Met |
| TC-STYLE | team Code Style | format/vet/lint pass | pass | `make check-format vet lint` | build-and-test | Met |
| TC-PKG | project Mandated | package verification passes | `verifypkg: OK` | `make verify-package` | build-and-test | Met |
| NFR2 | requirements.md | fixed args, ≤ 10 s, capped output | met | CLI token tests | build-and-test | Met |
| NFR4 | requirements.md | ≤ 1 CLI run / provider / 5 min | met | cache test | build-and-test | Met |

Details and evidence: [test-results.md](test-results.md).

## Readiness Assessment

- Build-ready: yes.
- Test-ready: yes — all commands pass, every target Met, traceability gate PASS.
- Deployment-ready: yes for a release PR; the version in `manifest.yaml` is still `0.5.0` and is set in the Deployment Pipeline stage.

## Known Limitations

- A cached CLI token is forgotten on 401 only in Test and the PR watcher; other actions recover on the next Test or after the 5-minute expiry (documented `ponytail:` limit).
- `glab config get token --host gitlab.com` does not see a token given to glab only through `GITLAB_TOKEN`.
- The CLI runs as the Kandev server user; it does not work when the server has no `gh`/`glab` (e.g. Docker image without them).
- No automated test runs a real `gh`/`glab`; an optional manual check on the self-hosted Kandev is recommended before release.
