# Unit Test Instructions — Plugin install failed: 502

## Framework and Setup

- Go `testing` + `github.com/stretchr/testify/require`, table-driven with `t.Run` (team Testing Posture). No new test dependency.
- Prerequisites: Go 1.26.x on `PATH`; `../kandev` (sibling of the repository root) checked out at the commit in `.kandev-sdk-ref` (Kandev v0.96.0), because `go.mod` replaces the SDK with that path.
- `-race` needs CGO (a C compiler); it is available on the local machine and the `ubuntu-latest` runner.

## Run This Fix's Tests

Exact, scoped command (runnable before the first Red step):

```bash
go test -race ./internal/pkgverify/...
```

Package-level check after the build change (Step 6 of the plan):

```bash
make package verify-package && ls -l dist/
```

## Tests in Scope

| Test | Requirement |
|---|---|
| `TestVerifyAcceptsAGoodPackage` (4 Linux/macOS executables) | FR1.1 happy path |
| `TestVerifyRejects` / "a manifest that still lists windows-amd64" | FR1.2, NFR3 (regression) |
| `TestVerifyRejects` / "a package containing the Windows executable" -> `unexpected executable: server/plugin-windows-amd64.exe` | FR1.1, NFR3 (regression) |
| `TestVerifyRejects` / "a manifest missing an executable" | FR1.2 |
| `TestVerifyRejects` / "a missing executable" (existing) | FR1.2 |

All other existing `pkgverify` tests stay green unchanged except the fixture and the expected manifest error text (`manifest must list exactly the four executables`).

## Coverage

- Team floor: 80% line coverage over `./internal/...` and `./server/...`, measured by `make coverage` in Build and Test (profile under `build/`, never the repository root). This fix must not lower it.

## Mocking and Test Data

- No mocks. Tests build in-memory `.tar.gz` packages with the existing `pkg.write` helper into `t.TempDir()`. Executables are short placeholder strings; no real binaries, no network, no credentials.
