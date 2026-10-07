# Quality Gates

Every gate is a `Makefile` target or a workflow step, and a failure stops the job. Gates in the `ci` workflow block merges into `main` through the required status checks `checks` and `packaged-host-contract` ([Q2] A, [Q3] A).

## Merge gates (`ci` workflow)

| # | Gate | Command | Fails when | Rule |
|---|------|---------|------------|------|
| G1 | Module tidy | `go mod tidy` + `git diff --exit-code -- go.mod go.sum` | `go.mod`/`go.sum` would change | team Code Style |
| G2 | Format | `make check-format` | `gofmt -l server internal` lists a file, or `prettier --check` fails in `ui/` | team Code Style |
| G3 | Vet | `make vet` | `go vet ./...` reports a problem, or `../kandev` is not at `.kandev-sdk-ref` | team Code Style |
| G4 | Lint | `make lint` | `golangci-lint` (default rules + `gosec`) finds an issue; `tsc --noEmit` (strict) or ESLint fails; `actionlint` fails; an action is not pinned to a full SHA | team Code Style, Deployment |
| G5 | Tests | `make test` | any `go test -race ./internal/... ./server/...` or Vitest test fails | team Testing Posture |
| G6 | Coverage | `make coverage` | Go line coverage is below 80% after excluding only `server/main.go` | team Testing Posture |
| G7 | Secrets | `make check-secrets` | a credential-shaped string is found in test data or test artifacts (including `coverage.out`) | project Forbidden / Mandated |
| G8 | Build | `make build` | a cross-compile for `linux-amd64`, `linux-arm64`, `darwin-amd64`, `darwin-arm64` or `windows-amd64` fails | US7.1 |
| G9 | Package | `make package` | the UI bundle contains React, or `plugin-pack` fails | US7.1 |
| G10 | Package verification | `make verify-package` | the archive checksum, per-file checksums, contents, plugin id or version do not match | project Mandated, BR5.2 |
| G11 | Packaged-host contract | `sha256sum -c checksums.txt` + `make verify-package contract-test` | the uploaded package changed, or it cannot be installed and run on Kandev `v<min_kandev_version>` | team Testing Posture, US7.4 |

G1–G10 run in job `checks`; G11 runs in job `packaged-host-contract`.

## Release gates (`release` workflow)

| # | Gate | Fails when |
|---|------|------------|
| R1 | G1–G10 again on the tagged commit | as above |
| R2 | `make release-preflight TAG=<tag>` | the tag is malformed, differs from `manifest.yaml`, is not on `origin/main`, already has a Release, or is the first release without a passing record in `docs/manual-checks` |
| R3 | G11 on the `release-package` artifact | as above |
| R4 | Checksum before publishing | `sha256sum -c checksums.txt` fails in `publish` |

Order: package verification (G10) runs before the preflight (R2) and before any tag is published, as the project Mandated rule requires.

## Enforcement on GitHub

| Control | Setting |
|---------|---------|
| Required status checks | `checks`, `packaged-host-contract` |
| Pull request required | yes, 0 approvals, squash only |
| Force-push | blocked |
| Branch deletion | blocked |
| Bypass | none |

## Mapping to the Build and Test commands

These are the commands recorded in `construction/build-and-test/build-instructions.md` and run in `test-results.md`:

| Build and Test command | Gate |
|------------------------|------|
| `make check-format vet lint` | G2, G3, G4 |
| `make test` | G5 |
| `make coverage` | G6 |
| `make check-secrets` | G7 |
| `make build package verify-package` | G8, G9, G10 |
| `make contract-test` (repeated 10 times after Loop-back 2) | G11, once per CI run |

CI runs the contract test once per run. The 10-run repetition is a local check for startup races (project learning), not a CI gate.

## Not gated in CI

These targets from `build-and-test-summary.md` stay Unverified and are carried to Operation stages, as accepted at Build and Test: T-PERF-02, T-SEC-06, T-SEC-11, T-OBS-03, T-ACC-02, T-REL-08, T-MKT-01, T-MAN-01. The manual real-space checks (walking skeleton and pre-release) are enforced only for the first release, through R2.
