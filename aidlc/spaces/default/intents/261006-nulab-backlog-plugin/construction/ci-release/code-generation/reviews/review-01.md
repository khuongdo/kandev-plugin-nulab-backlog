## Review

**Verdict:** READY
**Reviewer:** aidlc-architecture-reviewer-agent
**Date:** 2026-10-06T09:53:32Z
**Iteration:** 1

### Findings

| ID | Severity | Location | Finding | Required action | Status |
|---|---|---|---|---|---|
| R-01 | Major | internal/ci/release.go > hasEarlierRelease and CheckRelease; Makefile > release-preflight (`-tags "$(git tag -l 'v*')"`) | The "first release" test that arms the AC7.5.3 manual-check gate counts any other stable `v*` git tag, not an existing GitHub Release. Failure path: tag v0.1.0 is pushed, the preflight refuses (no passing first-release record), and the tag stays in the repo with no Release. A later tag such as v0.1.1 then sees an "earlier release", returns nil, and skips the gate. The first real Release ships without the recorded manual check against a real Backlog space (team.md Testing Posture, AC7.5.3). | Derive "is first release" from published Releases, for example `gh release list` has no non-prerelease entry, and pass that as a Fact. Add a test where an unreleased stable tag exists and the gate still fires. | New |
| R-02 | Minor | .github/workflows/release.yml > job contract (step "Package and run the packaged-host contract test"); .github/workflows/ci.yml > packaged-host-contract | The contract job runs `make package` itself. The package it installs is a rebuild, not the `release-package` artifact that `verify` uploads and `publish` attests and ships. The tested bytes and the released bytes are only the same if the build is reproducible. I did not verify reproducibility. | In release.yml, make `contract` depend on `verify`, download `release-package` into `plugin/dist/`, and run `contract-test` without rebuilding. Or record that the build is deterministic. | New |
| R-03 | Minor | Makefile > release-preflight (`released=false; gh release view ... && released=true`) | Any `gh` failure, such as an auth or network error, is treated as "no Release exists", so the check fails open. `publish` still fails at `gh release create`, so a released tag is not overwritten, and this is not blocking. The refusal for an existing Release (AC7.5.2) is therefore not reliable on its own. | Distinguish "release not found" from other `gh` errors, and fail the preflight on the latter. | New |
| R-04 | Minor | internal/ci/secrets.go > `identifier` regex and lineRule (developer-reported deviation) | The relaxation is acceptable. It skips only runs shaped `^[A-Za-z]+[0-9]*$`. A random 32+ character mixed key almost never has all its digits at the end. The bait test and the planted-key evidence in code-summary.md back this. The residual gap is a real secret of letters followed by trailing digits, and also tokens that lack upper case, lower case or a digit, such as lowercase-hex keys. | Accepted as the user-approved heuristic. Optionally narrow the exemption to Go test-function names (`func Test...`) so it cannot hide a token elsewhere. | Accepted risk |
| R-05 | Minor | internal/ci/release.go > CheckRelease (`tag.Prerelease` returns nil before the manual-check gate) | A `-suffix` tag skips the first-release manual check but still creates a public GitHub Release, marked `--prerelease`. AC7.5.3 says "before the first release". This follows the plan's assumption, but it is a way around the gate. | Confirm with the user that pre-release tags may skip the manual check, and record the decision in the plan. | New |

### Validation Tool Results

| Tool | Result | Interpretation |
|---|---|---|
| gofmt -l . | no output | Clean. |
| go vet ./... | no output | Clean. |
| go test -race -cover ./internal/... ./server/... | all packages ok; internal/ci 90.7% | Matches code-summary.md. `server` is 0.0% and is the one declared coverage exclusion. |
| go run ./cmd/ci secrets -root . | `ci secrets: OK` | Passes. The scan wrote nothing; git status was identical before and after. |
| go run ./cmd/ci workflows -dir .github/workflows | `ci workflows: OK` | Passes. |
| Action SHAs (gh api) | attest-build-provenance 4d101475... = v4.2.2; download-artifact 3e5f45b2... = v8.0.1; upload-artifact 043fb46d... = v7.0.1 | The pins resolve to the claimed tags. |
| Marketplace entry vs kandev plugin-registry/schema.json | id and repo patterns match; `categories` is an array of strings; `featured` is never set | The entry shape is valid. |
| Not run (read-only constraint) | contract test, make targets, actionlint | I relied on code-summary.md. Its results are plausible against the code. |

### Summary

U5 is implementable and consistent with the contracts and team.md: pinned actions, a single write-permission job, a fail-closed on-main check, a secrets scan that never prints matches, and a contract driver with injected waits. The one real defect is R-01: a failed first tag attempt disarms the AC7.5.3 manual-check gate for the next tag. The other findings are minor. The developer deviations (secret-scan relaxation, stricter workflow policy, `build-agentctl` with its own port, the `../kandev-min` worktree) are sound. The attestation and `gh attestation verify` (AC7.5.1) and branch-protection settings are still unproven until the first tag, as code-summary.md already says.
