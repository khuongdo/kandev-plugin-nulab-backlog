## Review

**Verdict:** READY
**Reviewer:** aidlc-architecture-reviewer-agent
**Date:** 2026-10-07T23:16:44Z
**Iteration:** 1

### Findings

| ID | Severity | Location | Finding | Required action | Status |
|---|---|---|---|---|---|
| R-01 | Minor | aidlc/spaces/default/intents/261007-plugin-install-502/inception/requirements-analysis/requirements.md > NFR1; construction/code-generation/code-summary.md > Package Check | The measured package is 23,393,498 bytes (about 22.3 MiB). At the Kandev 30 s read limit this still fails on any upload link slower than about 0.78 MB/s. The failing upload in the root cause (29.5 MB in more than 30 s) implies a link below about 0.98 MB/s, so a link between 0.78 and 0.98 MB/s now succeeds and a slower one still fails. The size reduction only reduces the failure window, and FR2.1/FR2.2 (From URL, KANDEV_SERVER_READTIMEOUT) are the real fix. The requirements record this limit, and the README troubleshooting note covers it. | When the human approves, accept that the user's own upload may still fail. Consider a one-line note in the release notes (Deployment Pipeline) that From URL is the dependable path. No code change needed. | New |
| R-02 | Minor | manifest.yaml > runtime.executables; README.md > Install on Kandev | Dropping `windows-amd64` is a breaking change for any Windows-hosted Kandev already running 0.4.1: an upgrade to the next patch is refused with `ErrPlatformNotSupported`. The README states the platform list, but a patch version number hides that this is a platform drop. Requirements C1 acknowledges it. | Put the Windows drop explicitly in the release notes and upgrade notes in the Deployment Pipeline stage (FR1.3). | New |
| R-03 | Minor | construction/code-generation/code-generation-plan.md > Testing Contract (plan_profile.steps and testable_layers) | The embedded contract lists the data model, repository, API and frontend layers as testable, but the plan applies only business logic. The prose after the contract justifies the omission, but the contract and the Steps list disagree. No reader is misled in practice. | None needed. Optionally state the omission inside the contract as "not applicable" so the plan reads as consistent. | New |
| R-04 | Minor | internal/pkgverify/pkgverify.go > Verify (unexpected-executable loop) | The check calls `slices.Contains(values(executables), name)` once per archive file, so the executables map is rebuilt into a slice for every file. The cost is negligible, since an archive has a handful of entries. The behaviour is correct: any `server/` file outside the four is rejected, even one listed in the in-archive checksums. | None required. Optionally compute the allowed set once before the loop. | New |

### Validation Tool Results

| Tool | Result | Interpretation |
|---|---|---|
| `go test -race ./internal/pkgverify/... ./internal/plugin/...` | PASS (ok, ok) | The regression cases (manifest still lists Windows, package contains the `.exe`, manifest missing an executable) and the updated `TestManifestListsExactlyTheFourExecutables` pass. |
| Differential check across the three platform lists (`manifest.yaml`, `Makefile` `PLATFORMS`, `pkgverify.executables`) | Consistent | All three list the same four entries (C4, FR1.2). A grep for `windows` or `.exe` outside `aidlc/` finds only the README prose, comments, the new regression tests and the frozen fixture `internal/plugin/testdata/v030/manifest.yaml`. That fixture is a released 0.3.0 manifest used by `v030_test.go`, so leaving it unchanged is correct. |
| Makefile `build` clears `build/server` | Verified in diff | The stale-binary risk is closed: a leftover `plugin-windows-amd64.exe` can no longer reach the package. The new verifier check is a second guard. |
| `dist/` package size | 23,393,498 bytes | Within the NFR1 limit of 25 MB. |

### Summary

The change is small and coherent. The four-executable set is consistent across the manifest, Makefile and verifier. The verifier rejects both a listed Windows entry and a stray Windows file under `server/`, with targeted regression tests, and the suite is green. Only minor residual risks remain: the package is still slower than the read limit on links below about 0.78 MB/s (already covered by the URL-install documentation), and the breaking effect on Windows servers should be called out in the release notes.
