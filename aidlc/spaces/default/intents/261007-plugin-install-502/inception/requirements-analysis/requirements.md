# Requirements — Plugin install failed: 502

## Intent Analysis

- **Request** [desc]: "Investigate and fix Plugin install failed: 502 when installing plugin. Read local Kandev server logs if needed".
- **Type**: bug fix, single component (packaging and install documentation), Minimal depth.
- **Root cause (confirmed)**: the Kandev server v0.97.0 stops reading a request after 30 s (`server.readTimeout`, env `KANDEV_SERVER_READTIMEOUT`). A browser upload of the 29.5 MB package `nulab-backlog-0.4.1.tar.gz` through `tailscale serve` took longer, the body read was cut (backend: 400, 47 bytes, `missing multipart field "package"`), and the proxy answered 502, which the UI shows as `Plugin install failed: 502`. Reproduced with a throttled 29.5 MB upload (502 after 33.5 s). Source: [codekb code-quality-assessment.md](../../../../codekb/kandev-plugin-nulab-backlog/code-quality-assessment.md), [architecture.md](../../../../codekb/kandev-plugin-nulab-backlog/architecture.md).
- **Goal**: make installing the plugin on a self-hosted Kandev server work reliably: a smaller package for browser upload, and a documented install path that does not depend on upload speed [Q1].

## Functional Requirements

### FR1 — Smaller package without Windows [Q1, Q2, Q4]

- **FR1.1**: The package shall contain exactly 4 server executables: `server/plugin-linux-amd64`, `server/plugin-linux-arm64`, `server/plugin-darwin-amd64`, `server/plugin-darwin-arm64`. `server/plugin-windows-amd64.exe` is no longer built or packaged.
  - Given the release build, When `make package` runs, Then the archive lists those 4 executables and no `.exe` file.
- **FR1.2**: `manifest.yaml` `runtime.executables` shall list exactly those 4 platforms, and the `Makefile` `PLATFORMS` and the package verifier (`internal/pkgverify`) shall use the same 4-entry list.
  - Given a package that still contains or lists `windows-amd64`, When `make verify-package` runs, Then verification fails.
  - Given a package missing one of the 4 executables, When `make verify-package` runs, Then verification fails naming the missing file.
- **FR1.3**: The fix shall be released as a patch version (next after the latest GitHub release, checked with `gh release list` and `origin/main` before choosing the number, per project Deployment rules).

### FR2 — Install documentation [Q1, Q3, Q4]

- **FR2.1**: The README "Install on Kandev" section shall describe installing by release URL (Kandev downloads the package from the GitHub Release asset itself, so the 30 s upload limit does not apply) as the recommended path, and browser upload of the `.tar.gz` as the alternative.
- **FR2.2**: The README shall contain a troubleshooting note: `Plugin install failed: 502` (or a 400 `missing multipart field "package"` in the server log) after about 30 s means the upload was slower than the server's 30 s read limit; fix by installing by URL, or by raising `KANDEV_SERVER_READTIMEOUT` (seconds) on the Kandev server.
- **FR2.3**: The README shall state the supported server platforms: Linux and macOS on amd64 and arm64; Windows servers are not supported from this version on.

## Non-Functional Requirements

- **NFR1 (size)**: The release package shall be at most 25 MB (expected about 24 MB) [Q4]. Browser upload on links slower than about 0.8 MB/s may still exceed the server limit; that case is covered by FR2.1 and FR2.2, not by package size [Q4].
- **NFR2 (no regression)**: The existing CI checks (`make check-format vet lint test coverage check-secrets build package verify-package`) and the packaged-host contract test stay green; coverage stays at or above 80% [team Testing Posture].
- **NFR3 (regression test)**: A targeted test in `internal/pkgverify` shall fail if a Windows executable is present or listed, and if one of the 4 required executables is missing [org Testing Posture, bugfix scope].

## Constraints

- **C1**: Kandev rejects a package that has no executable for the server's own platform (`pkgtar.ErrPlatformNotSupported`), so Windows-hosted Kandev servers can no longer install the plugin from this version on [Q2]. This supersedes NFR7 of intent `261006-nulab-backlog-plugin` (all 5 platforms).
- **C2**: The 30 s read limit is a Kandev server setting; this repository cannot change it [codekb].
- **C3**: No change to the local Kandev server or its systemd unit [Q3].
- **C4**: The platform list is declared in three places (`manifest.yaml`, `Makefile`, `internal/pkgverify`) and must change together [codekb code-structure.md].

## Assumptions

- [assumption] No current user runs Kandev on Windows; owner: user, validated at release (user chose to drop Windows in Q2).
- [assumption] Each executable compresses to about 6 MB, so 4 executables give about 24 MB; validated by measuring the built package in Build and Test.

## Out of Scope

- Changing Kandev's `server.readTimeout` default or its install route (upstream Kandev).
- Changing the local server configuration [Q3].
- Further compression or per-platform packages [Q4].
- Linux-only packaging [Q4].

## Open Questions

- None.
