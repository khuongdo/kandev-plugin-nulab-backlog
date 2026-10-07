# Code Quality Assessment — kandev-plugin-nulab-backlog

## Test Coverage and Baselines

- ~108 Go test files co-located in every `internal/*` package (`httptest` fakes, `testdata/` JSON), run with `-race`; ~40 Vitest files under `ui/src/`.
- `make coverage`: 80% line floor over `./internal/... ./server/...`, sole exclusion `server/main.go`, profile at `build/coverage.out`.
- Packaging: `pkgverify_test.go`, `contract_test.go` (driver units); the real install path runs only in `make contract-test` (CI `packaged-host-contract`, release `contract`).
- Baseline: **not recorded** this run (no Go on `PATH`, no `../kandev` in this worktree). Link `../kandev` to `v0.96.0` before Construction.

## Linting

gofmt, go vet, golangci-lint (`.golangci.yml`, +gosec), `tsc --noEmit` strict, ESLint, Prettier, actionlint, `cmd/ci workflows` policy (SHA-pinned actions, `contents: read`, no `pull_request_target`).

## CI/CD

- `ci.yml`: `checks` (format, vet, lint, test, coverage, check-secrets, build, package, verify-package) and `packaged-host-contract` (installs the package on Kandev built at `v0.96.0`).
- `release.yml`: `verify` -> `contract` -> `publish` (build provenance attestation, `gh release create`).

## Documentation

README covers build, install (upload via Settings > Plugins), connection, CI, release, marketplace. Go packages have `doc.go`. `docs/manual-checks/` holds the first-release record.

## Known Issue: Plugin install 502

Symptom: Kandev web UI shows `Plugin install failed: 502` when uploading `nulab-backlog-0.4.1.tar.gz`.

Verified evidence:
- Runtime: Kandev v0.97.0 (`kandev --headless`, `:38429`) behind `tailscale serve` (`https://webfrontier.tail152aaa.ts.net`).
- Kandev v0.97.0 `server.readTimeout` default 30 s (`KANDEV_SERVER_READTIMEOUT`, `catalog.go` line 61). Multipart upload install parses the whole body before `Install`; a body slower than 30 s is cut.
- Backend logs: `POST /api/plugins/install` 400, `duration_ms` ~30000, 47-byte body `{"error":"missing multipart field \"package\""}`. The install handler never returns 502; `tailscale serve` turns the dropped upstream into 502, and the web UI prints it from the status line.
- Reproduced: 29.5 MB upload throttled to 800 KB/s via the tailscale URL -> 502 after 33.5 s; unthrottled uploads succeed.
- Package: 29,498,475 bytes; 5 server binaries of 14.7-16.4 MB each uncompressed, already `-trimpath -ldflags "-s -w"`; `ui/bundle.js` 226 KB. The package itself is valid (passes `verify-package` and the contract test).

Classification: host timeout plus plugin package size. Not a manifest or checksum defect. Fix options: [architecture.md](architecture.md#improvement-opportunities).

## Technical Debt

- Package ~29.5 MB: every install uploads all 5 platforms to use one.
- Platform set declared three times (`manifest.yaml`, `Makefile` `PLATFORMS`, `pkgverify` `executables`).
- `pkgverify` duplicates Kandev `pkgtar` rules; no size limit check.
- Contract test installs over loopback (30 s client timeout), so it never sees proxy/slow-upload failures.
- Sibling checkout `~/repo/kandev-plugin-nulab-backlog/dist` holds stale `0.0.1`/`0.1.0` tarballs; easy to upload the wrong file.
- Runtime drift: Kandev 0.97.0 running vs 0.96.0 pin.
