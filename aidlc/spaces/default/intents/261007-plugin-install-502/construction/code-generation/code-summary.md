# Code Summary — Plugin install failed: 502

Zero-Unit bugfix. The package now carries 4 server executables (Linux and macOS, amd64 and arm64) instead of 5, the verifier enforces that set, and the README documents install by URL, the 502 troubleshooting note and the supported platforms.

## Files Changed

| File | Change | Requirements |
|---|---|---|
| `internal/pkgverify/pkgverify.go` | `executables` map without `windows-amd64`; new check that every archive file under `server/` is one of the listed executables (`unexpected executable: <path>`); manifest error now `manifest must list exactly the four executables of BR5.1`; doc comments updated | FR1.1, FR1.2, NFR3 |
| `internal/pkgverify/pkgverify_test.go` | good fixture with the 4 executables; new regression cases (manifest still lists `windows-amd64`; package contains `server/plugin-windows-amd64.exe`; manifest missing an executable); removed the obsolete "four executables" / "sixth executable" cases | FR1.1, FR1.2, NFR3 |
| `internal/plugin/manifest_test.go` | `TestManifestListsExactlyTheFiveExecutables` renamed to `TestManifestListsExactlyTheFourExecutables`, expected map without Windows (see Deviations) | FR1.2 |
| `manifest.yaml` | `runtime.executables` without `windows-amd64` | FR1.1, FR1.2 |
| `Makefile` | `PLATFORMS := linux-amd64 linux-arm64 darwin-amd64 darwin-arm64`; `.exe` branch removed; `build` runs `rm -rf $(BUILD)/server` first so a stale executable is never packaged | FR1.1, FR1.2 |
| `README.md` | "Install on Kandev": supported platforms, **From URL** (recommended, GitHub Release asset URL) and **Upload file** (alternative); new "Troubleshooting: `Plugin install failed: 502`" subsection | FR2.1, FR2.2, FR2.3 |

`version` in `manifest.yaml` is unchanged (set at release, FR1.3).

## Key Decisions

- The unexpected-executable check is a separate loop in `Verify` after the required-file check, so a missing executable is still reported by name first, and any extra file under `server/` (the Windows `.exe` or anything else) is rejected even when it is listed in the in-archive `checksums.txt`.
- The README wording was checked against Kandev v0.96.0 (the manifest's `min_kandev_version`, checkout at `.kandev-sdk-ref`): its **Install plugin** dialog already has the **From URL** tab (default) and the **Upload file** tab (`apps/web/components/settings/plugins/install-plugin-dialog.tsx`, labels from `apps/web/src/locales/en/plugins.json`); the backend downloads with a 60 s timeout and a 100 MiB cap and follows redirects (default `http.Client`), so the GitHub Release asset URL works; `server.readTimeout` defaults to `30` (`KANDEV_SERVER_READTIMEOUT`, `apps/backend/internal/common/config/catalog.go`).
- `internal/plugin/testdata/v030/manifest.yaml` still lists `windows-amd64`. It is a frozen fixture of the released 0.3.0 manifest and is left unchanged.

## Test Runner and Baseline (Step 1)

- Go `go1.26.8` at `~/.local/go`; `../kandev` created as a detached `git worktree` of `~/repo/kandev` at `f099a46dc7aab16f6ff5806cd29b2b480296303f` (tag `v0.96.0`).
- Scoped command: `go test -race ./internal/pkgverify/...` -> `ok`, 28 passing tests and subtests (9 top-level tests).
- Whole suite baseline: `go test -race ./...` -> 1324 passing tests and subtests, 0 failing.

## Red Output (Step 2)

Command: `go test -race ./internal/pkgverify/...` after changing only the test file. Result: `FAIL`, 16 failing and 13 passing tests and subtests. Every failure comes from the old verifier still requiring the Windows executable or allowing it:

```
--- FAIL: TestVerifyAcceptsAGoodPackage
--- FAIL: TestVerifyRejects
    --- FAIL: TestVerifyRejects/a_wrong_plugin_id
    --- FAIL: TestVerifyRejects/a_wrong_version
    --- FAIL: TestVerifyRejects/a_manifest_missing_an_executable
    --- FAIL: TestVerifyRejects/a_manifest_that_still_lists_windows-amd64
    --- FAIL: TestVerifyRejects/a_package_containing_the_Windows_executable
    --- FAIL: TestVerifyRejects/an_invalid_manifest
--- FAIL: TestRunExitCodes
--- FAIL: TestVerifyRejectsABundleLoadingANulabOrBacklogAsset (5 subtests)
--- FAIL: TestVerifyAcceptsABundleNamingBacklogHostsWithoutURLs
Error: "missing required file: server/plugin-windows-amd64.exe" does not contain "manifest must list exactly the four executables"
Error: "manifest must list exactly the 5 executables of BR5.1" does not contain "unexpected executable: server/plugin-windows-amd64.exe"
FAIL	github.com/khuongdo/kandev-plugin-nulab-backlog/internal/pkgverify
```

## Green and Refactor (Steps 3-4)

- After the verifier change: `go test -race ./internal/pkgverify/...` -> `ok`, 29 passing tests and subtests, 0 failing.
- After the doc-comment refactor: still `ok`; `gofmt -l server internal` prints nothing; `go vet ./...` clean.

## Package Check and Full Suite (Step 6)

- `make package verify-package` -> `verifypkg: OK dist/nulab-backlog-0.4.1.tar.gz (nulab-backlog@0.4.1)`.
- Package size: `23393498` bytes (23.4 MB, 22.3 MiB), within the 25 MB cap of NFR1. The previous 5-executable package was 29.5 MB.
- Archive contents: `manifest.yaml`, `server/plugin-darwin-amd64`, `server/plugin-darwin-arm64`, `server/plugin-linux-amd64`, `server/plugin-linux-arm64`, `ui/bundle.js`, `checksums.txt`; no `.exe`.
- Uncompressed executables: 14.7 to 16.3 MB each, below the verifier's 128 MiB entry bound.
- `go test -race ./...` -> 1325 passing tests and subtests, 0 failing (baseline 1324; net +1 subtest in `pkgverify`).
- `make coverage` -> `coverage: 92.8% (floor 80%, excluded: server/main.go)`; profile under `build/`, no `coverage.out` at the repository root.
- Not run in this stage: `make lint` (golangci-lint, tsc, eslint, actionlint), `make check-secrets` and the packaged-host contract test; these belong to Build and Test.

## Deviations

- **Fourth platform list**: the plan (and constraint C4) names three places that declare the platform list. A fourth exists: `internal/plugin/manifest_test.go` asserts the exact `runtime.executables` map of `manifest.yaml`. It failed after Step 5, so it was updated (renamed to `TestManifestListsExactlyTheFourExecutables`, Windows entry removed). No other behaviour changed.
- **Local setup**: `ui/node_modules` was installed with `npm ci` (gitignored) to run `make package`; `build/` and `dist/` are build outputs and are not part of the change.
