# Health Check Report — 261007-opt-in-default (v0.3.0)

## Pre-merge Health (rebased branch, commit `9d6d312` on top of v0.2.0)

| Check | Command | Result |
|---|---|---|
| Format, vet, lint, SAST, workflows | `make check-format vet lint` | `0 issues.`, `ci workflows: OK` |
| Secret scan | `make check-secrets` | `ci secrets: OK` |
| Go tests + coverage | `make coverage` (`go test -race`) | all packages ok; 92.8% (floor 80%) |
| UI tests | `cd ui && npx vitest run` | 31 files, 322/322 passed |
| Package | `make package verify-package` | `verifypkg: OK dist/nulab-backlog-0.3.0.tar.gz (nulab-backlog@0.3.0)` |
| Packaged-host contract on Kandev v0.96.0 | `make contract-test KANDEV_MIN_DIR=../kandev` ×3 | `ci contract: OK` 3/3 |
| Pull request CI | `ci.yml` on PR #8 | Started; result to be checked by the maintainer before merge |

Environment: Go 1.26.x (`~/.local/go/bin`), `../kandev` at v0.96.0 (`f099a46dc`).

## Runtime Health (self-hosted Kandev)

Not applicable yet: the release is not published or installed (Q2 = C). When `v0.3.0` is installed, the health signals are: the plugin shows version `0.3.0` and starts without errors in the Kandev log; Settings > Integrations > Backlog loads (`connection.get` answers with `enabled`).
