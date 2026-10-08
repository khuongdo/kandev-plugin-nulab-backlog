# Smoke Test Results - v0.5.2

## Automated Checks (run here)

| Check | Environment | Result |
|-------|-------------|--------|
| Packaged-host contract: install the package and start it | Local throwaway Kandev v0.96.0, 10 runs on the rebased branch | 10/10 `ci contract: OK nulab-backlog on Kandev v0.96.0` |
| Packaged-host contract | PR #21 CI (`packaged-host-contract`) | pass |
| Packaged-host contract | `release.yml` job `contract` on tag `v0.5.2` | success |
| Package verification | `release.yml` job `verify` + local `make verify-package` | success, `verifypkg: OK ... (nulab-backlog@0.5.2)` |
| Release assets | GitHub Release v0.5.2 | `nulab-backlog-0.5.2.tar.gz` and `checksums.txt` present |

## Self-Hosted Smoke Checks (yours, scope B)

The six checks in `operation/deployment-pipeline/deployment-strategy.md` need the self-hosted Kandev and a real Backlog space. They were **not run here**, because the release scope stops at the GitHub Release.

| # | Check | Result |
|---|-------|--------|
| 1 | Plugin 0.5.2 running | Pending (yours) |
| 2 | "Link Backlog issue" in the task Link menu | Pending (yours) |
| 3 | Link by key and by link; badge updates at once | Pending (yours) |
| 4 | Inline errors for unknown key / non-Backlog link | Pending (yours) |
| 5 | Hidden while linked; back after Unlink | Pending (yours) |
| 6 | Restyled "Link to task" dialog; keyboard Enter saves (R-04) | Pending (yours) |
