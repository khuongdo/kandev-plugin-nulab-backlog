# Smoke Test Results — 261007-source-control-agnostic (v0.4.0)

## Status

**Pending (Q2 = A).** The maintainer installs `v0.4.0` on the self-hosted Kandev and runs the smoke check later; results are recorded here when reported.

## Automated Smoke Evidence Already Available

| Check | Result | Evidence |
|---|---|---|
| Packaged plugin installs and starts on Kandev 0.96.0 | Pass | `release.yml` `contract` job on the exact release package; local `make contract-test` 10/10 |
| Package integrity | Pass | `verify-package` in `release.yml` `verify` job; `checksums.txt` on the Release |

## Manual Smoke Check (pending)

From `operation/deployment-pipeline/deployment-strategy.md`:

| # | Check | Result |
|---|---|---|
| 1 | Plugin shows `0.4.0`, no startup errors in the Kandev log | Pending |
| 2 | Existing Backlog features unchanged (issue list, linked Backlog Git PR status, existing saved PR query runs) | Pending |
| 3 | "Source control" section lists Backlog Git, GitHub, GitLab, Bitbucket; "Git access" inside it; read-only for non-admins | Pending |
| 4 | Optional real-account check: add a read-only token, Test shows the account, map one project → repository, PR list loads | Pending |
