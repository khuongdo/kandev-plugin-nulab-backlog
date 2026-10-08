# Code Quality Assessment — kandev-plugin-nulab-backlog

## Test Coverage and Baselines

- Go: ~108 test files, about 704 `func Test`, co-located in every `internal/*` package (`httptest` fakes, `testdata/` JSON), run with `-race`.
- UI: 33 Vitest files under `ui/src/` (about 283 `it` cases), jsdom, shared fake host `ui/src/testing/harness.ts`.
- `make coverage`: 80% line floor over `./internal/... ./server/...`, sole exclusion `server/main.go`, profile at `build/coverage.out`.
- Packaging: `pkgverify_test.go`, `contract_test.go` (driver units); the real install path runs only in `make contract-test` (CI `packaged-host-contract`, release `contract`).
- Workflow policy is tested in `internal/ci/workflows_test.go`.
- Baseline: **not recorded** in the 261007 and 261008 scans. `261008-ci-path-filter` changes only CI configuration, so the Go suite is verified by CI itself; the `261007` and `261008-fix-uiux-backlog` scans had no Go on `PATH`, no `../kandev` in the worktree and no `ui/node_modules`. Link `../kandev` to `v0.96.0` (e.g. `~/repo/kandev`, already at `f099a46`) and record it before Construction.

## Linting

gofmt, go vet, golangci-lint (`.golangci.yml`, +gosec), `tsc --noEmit` strict, ESLint, Prettier, actionlint, `cmd/ci workflows` policy (`internal/ci/workflows.go` lines 36-40: SHA-pinned actions, top-level `permissions` exactly `contents: read`, no `pull_request_target`, write permission only in `release.yml` job `publish`). The policy does not inspect `paths` / `paths-ignore`.

## CI/CD

- `ci.yml`: `checks` (format, vet, lint, test, coverage, check-secrets, build, package, verify-package) and `packaged-host-contract` (installs the package on Kandev built at `v0.96.0`). Runs on every PR to `main` and every push to `main`; no path filter, no concurrency group.
- `release.yml`: `verify` -> `contract` -> `publish` (build provenance attestation, `gh release create`) on `v*` tag push.
- Required checks `checks` and `packaged-host-contract` on `main` via ruleset `24580280`; squash-only. Details: [api-documentation.md](api-documentation.md#github-actions-triggers-and-required-checks).
- Records-only PRs run the full CI today: PR #12 (`records/backlog-panel-retouch-v0.4.1`) shows `checks` 2m17s and `packaged-host-contract` 53s. PRs #3, #5, #7, #10, #12 are records-only.

## CI Path Filter Constraints

Facts any design for intent `261008-ci-path-filter` must respect:

1. **Required checks must still report.** A plain `paths` / `paths-ignore` on `ci.yml` stops the workflow from starting on a records-only PR, so `checks` and `packaged-host-contract` stay "Expected - waiting" and the PR cannot merge. A job skipped by `if:` does report success. Candidate patterns: an always-triggered workflow with a cheap change-detection job gating the two required jobs by `if:`, or a same-named no-op workflow on the inverse paths (fragile: names must match exactly).
2. **Release filtering has no trigger-level meaning.** `paths` filters do not apply to tag pushes, and a release is a deliberate manual tag. "Skip release for non-app changes" needs a different definition (for example `release-preflight` refusing a tag whose diff from the previous release touches no app path). Needs a user decision.
3. **`docs/manual-checks/` is a release input** (`release-preflight`, Makefile line 145, `internal/ci/release.go`).
4. **Secret scan coverage.** `check-secrets` scans the whole repo, including `aidlc/` and `docs/`. Skipping CI on records-only PRs drops that scan for those files (project Forbidden: no real credentials in the repo). Decide whether a light secret scan stays on skipped PRs.
5. **App set includes CI and build files.** `.github/workflows/**` and `Makefile` must count as app paths, because `make lint` checks the workflows and every job runs through the Makefile. Full classification: [code-structure.md](code-structure.md#top-level-path-classification).
6. **Workflow policy.** Any new action must be SHA-pinned and top-level `permissions: contents: read` kept, or `make lint` fails.
7. **Both events.** The filter must apply consistently to `pull_request` and `push` to `main`; the main-push run builds the package that a later release rebuilds.

## Documentation

README covers build, install (upload via Settings > Plugins), connection, CI, release, marketplace. Go packages have `doc.go`; every exported Go symbol and UI factory has a doc comment, often citing BR/FR/AC ids. No TODO/FIXME in the scanned paths. `docs/manual-checks/` holds the first-release record. Makefile targets carry traceability IDs.

## Intent Findings: 261008-fix-uiux-backlog

| # | Request | Evidence | Change shape |
|---|---|---|---|
| 1 | Backlog issue on Home > Tasks rows, hover summary, click opens issue | Badge registered only for `task-card-tags` (`ui/src/index.ts`); Home > Tasks mounts only `task-row-metadata`. Hover today is `title` = updated time. `Link` (`internal/issues/types.go`) and `LinkView` (`internal/issues/service.go`) have no `Summary`; only `Detail` reads it live. | Register `IssueBadge` for `task-row-metadata` too; add `Summary` to `Link` (set in `newLink`, refreshed in `sync.go`) and `LinkView`/TS `LinkView`; tooltip via host `Tooltip*`; stop pointer propagation; fall back to the key for old links |
| 2 | Hide Home > Integrations entry while OFF | `registerNavItem` unconditional in `initialize`; Kandev v0.96.0 has no `requires` on plugin nav items, no unregister, no late registration (see [architecture.md](architecture.md#ui-surfaces-host-slots)) | Needs a decision (register-only-if-ON at load, drop entry, or upstream Kandev change + `min_kandev_version` bump). Supersedes BR5.4/BR7.6/BR7.8 and the `index.test.ts` case "registers the entry and the route even when Backlog is off everywhere" |
| 3 | Projects right below the sign-in method | `SettingsScreen.tsx` order ends with `projects`; method dropdown is inside the Connection form | JSX reorder to after `connection`; update `SECTIONS` order in `sections.test.tsx` |
| 4 | Issue-watch empty message | `issue-watches-section.tsx` uses `messages.watchesEmpty` = "No PR watches yet" (`messages/en.ts`), shared with PR watches | New key (e.g. `issueWatchesEmpty`); update the `sections.test.tsx` issue-list assertion |
| 5 | Remove per-list Add watch button | Empty state renders `add("backlog-issue-watches-empty-add")` besides the header action; PR watches has the same duplicate (`backlog-pr-watches-empty-add`) | Drop the empty-state child (`ListEmpty` works without children); no test references `-empty-add` |

Risks: `task-row-metadata` also renders in the sidebar task list (`surface: "sidebar"`) — decide whether the badge shows there; a stored `Summary` is Backlog content, so `internal/issues/leak_test.go` expectations (no Backlog content in errors) must keep holding.

## Known Issue: Plugin install 502

Status: addressed in v0.4.2 (commit `f5a7529`, "Smaller package without Windows", squash-merged to `main` as `3a983ab`) by dropping `windows-amd64` from the package; not re-verified by the 2026-10-08 scans. Original analysis (intent `261007-plugin-install-502`), kept for history:

- Runtime: Kandev v0.97.0 (`kandev --headless`, `:38429`) behind `tailscale serve` (`https://webfrontier.tail152aaa.ts.net`).
- Kandev `server.readTimeout` default 30 s (`KANDEV_SERVER_READTIMEOUT`, `catalog.go` line 61). Multipart upload install parses the whole body before `Install`; a body slower than 30 s is cut.
- Backend logs: `POST /api/plugins/install` 400, `duration_ms` ~30000, 47-byte body `{"error":"missing multipart field \"package\""}`. `tailscale serve` turns the dropped upstream into 502.
- Reproduced: 29.5 MB v0.4.1 upload throttled to 800 KB/s -> 502 after 33.5 s; unthrottled uploads succeed.
- Classification: host timeout plus plugin package size. Remaining option: install by URL (see [architecture.md](architecture.md#improvement-opportunities)).

## Technical Debt

- `ci.yml` and `release.yml` duplicate setup and contract steps (lines 13-58 vs 145-191, 68-124 vs 200-256).
- `ci.yml` has no `concurrency` group; superseded PR pushes keep running.
- `release-preflight` uses `gh release list --limit 1000` (marked `ponytail:` in the Makefile).
- Shared `watchesEmpty` message and duplicate empty-state Add buttons (finding 4-5 above).
- Plugin-owned switch bus (`ui/src/switch/enabled-events.ts`) duplicates host state because Kandev v0.96.0 cannot expose the integration switch to plugins.
- Platform list declared in four places (`manifest.yaml`, `Makefile` `PLATFORMS`, `pkgverify`, `internal/plugin/manifest_test.go`; see [code-structure.md](code-structure.md#build-and-packaging)).
- `pkgverify` duplicates Kandev `pkgtar` rules; no size limit check.
- Contract test installs over loopback (30 s client timeout), so it never sees proxy/slow-upload failures.
- Runtime drift: Kandev 0.97.0 running vs 0.96.0 pin.
