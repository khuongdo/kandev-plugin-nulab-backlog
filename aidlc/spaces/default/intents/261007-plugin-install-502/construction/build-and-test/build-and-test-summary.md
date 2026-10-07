# Build and Test Summary — Plugin install failed: 502

## Overall Status

Build and all tests pass. Prerequisites: Go 1.26.x, `../kandev` at `v0.96.0`, `npm ci` in `ui/` (see [build-instructions.md](build-instructions.md)).

## Test Type Inventory

| Type | Instruction file | Notes |
|---|---|---|
| Unit (fix's regression) | `../code-generation/unit-test-instructions.md` | `go test -race ./internal/pkgverify/...` |
| Full existing suite | [build-instructions.md](build-instructions.md) | Go + Vitest via `make test` |
| Packaged-host contract (integration) | [integration-test-instructions.md](integration-test-instructions.md) | Kandev v0.96.0, 10 runs |
| Package size | [performance-test-instructions.md](performance-test-instructions.md) | NFR1 |
| Security checks | [security-test-instructions.md](security-test-instructions.md) | lint/gosec, secret scan, package verifier |

Minimal strategy: no new integration, performance or security suites were generated; the files above record which existing checks cover each concern.

## Coverage Expectations

Zero-Unit bugfix: one stage-level target, 80% line coverage over `./internal/...` and `./server/...`. Measured 92.8%.

## Target Verification Matrix

| Target ID | Source | Expected | Actual | Evidence | Owning Stage | Verdict |
|---|---|---|---|---|---|---|
| T-COV | `code-generation-plan.md` > Testing Contract (team floor) | ≥ 80% line coverage | 92.8% | [test-results.md](test-results.md) | build-and-test | Met |
| T-RACE | Testing Contract (team) | Go tests green with `-race` | 1325/1325 pass | [test-results.md](test-results.md) | build-and-test | Met |
| T-CONTRACT | Testing Contract (team, project note: 10 runs) | contract test passes on `min_kandev_version` | 10/10 pass on v0.96.0 | [test-results.md](test-results.md) | build-and-test | Met |
| T-REGRESSION | Testing Contract scope floor (bugfix) / requirements NFR3 | targeted regression present and green | present, green | [test-results.md](test-results.md) | build-and-test | Met |
| T-SUITE | Testing Contract scope floor / NFR2 | existing suite green | all pass | [test-results.md](test-results.md) | build-and-test | Met |
| T-SIZE | requirements NFR1 | package ≤ 25 MB | 23,393,498 bytes | [test-results.md](test-results.md) | build-and-test | Met |

No `nfr-requirements/` or `nfr-design/` artifacts exist (bugfix scope skips them).

## Readiness Assessment

- **Build-ready**: yes.
- **Test-ready**: yes; every target Met.
- **Deployment-ready**: yes for a patch release, once Deployment Pipeline sets the version (FR1.3) and the release notes state that Windows is no longer packaged.

## Known Limitations

- A 23.4 MB browser upload still exceeds Kandev's 30 s read limit on links slower than about 0.78 MB/s; the README documents install From URL and `KANDEV_SERVER_READTIMEOUT` for that case.
- Windows-hosted Kandev servers can no longer install or upgrade the plugin from the next version on (requirements C1).
- FR1.3 (patch release) is owned by Deployment Pipeline and is not covered here.
