# Unit Test Instructions: Source Control Settings

## Framework and Setup

- Go: `testing` + `github.com/stretchr/testify/require`, table-driven with `t.Run`, always `-race`. Requires `../kandev` checked out at the commit in `.kandev-sdk-ref` (v0.96.0); `go.mod` uses a `replace` to it.
- UI: Vitest with jsdom and axe-core (existing `ui/vitest` setup); run from `ui/` after `npm ci`.
- No new test configuration is needed.

## Commands (scoped to this change)

Run before the first Red step to confirm the runners work, then after each Red/Green step:

```bash
go test -race -count=1 ./internal/scm/ ./internal/git/ ./internal/plugin/
```

```bash
cd ui && npx vitest run src/settings/source-control-section.test.tsx src/settings/sections.test.tsx src/settings/settings.test.tsx src/git/pr-list.test.tsx src/git/watch-form.test.tsx
```

Gates before handing over (team Code Style and Testing Posture):

```bash
make check-format vet lint test coverage
cd ui && npm run typecheck && npm run lint && npm run format:check
```

Delete `coverage.out` (or keep it under `build/`) after `make coverage`.

## Tests to Write (Minimal strategy: one per requirement)

| Requirement | Test location |
|---|---|
| FR1.1 four services, no "none" | `internal/scm/service_test.go` (SetActive validation); `source-control-section.test.tsx` (selector options) |
| FR1.2 / FR3.1 default Backlog Git | `internal/scm/service_test.go` derivation table |
| FR1.3 stored by backend | `internal/scm/store_test.go` (active survives settings writes); `actions_scm_test.go` (`scm.providers.list` returns active) |
| FR1.4 refuse non-active service | `internal/scm/service_test.go` (InactiveError names active; RemoveToken allowed); `actions_scm_test.go` (409 `service_inactive`) |
| FR1.5 only active service in lists/watchers | `internal/scm/watcher_test.go`; `pr-list.test.tsx` / `watch-form.test.tsx` |
| FR1.6 Backlog Git off when external active | `actions_scm_test.go` (git actions, Git credential); `internal/git/watcher_test.go` |
| FR2.1 selector + single card | `source-control-section.test.tsx` |
| FR2.2 admin-only switch | `actions_scm_test.go` (admin list); `source-control-section.test.tsx` (member read-only) |
| FR2.3 confirm on switch | `source-control-section.test.tsx` (dialog text, Cancel sends nothing) |
| FR2.4 keep old data disabled | `internal/scm/service_test.go` (switch away and back) |
| FR3.2 one external auto-picked | `internal/scm/service_test.go` derivation table |
| FR3.3 pick notice while pending | `internal/scm/service_test.go` (both work); `source-control-section.test.tsx` (notice) |
| FR4.1 / FR4.2 framed card | `source-control-section.test.tsx` |
| FR5.1-FR5.3 service-named labels | `source-control-section.test.tsx` |
| NFR1 no secrets | existing leak test extended to `scm.active.set` in `actions_scm_test.go` |
| NFR2 old document loads | `internal/scm/store_test.go` |
| NFR3 accessibility | axe check in `source-control-section.test.tsx` |

## Coverage Target

Go line coverage at or above 80% over `./internal/...` and `./server/...` via `make coverage`; no new exclusions, no lowered floor.

## Mocking and Test Data

- Go: use the existing fake host / httptest provider servers in `internal/scm/harness_test.go`, `internal/scm/fakes_test.go` and the `scmRig` in `internal/plugin/actions_scm_test.go`. Change the harness `connect`/`setToken` helpers once so multi-provider setups keep working (set active, connect, clear stored value).
- UI: mock `scm.providers.list` responses with an `active` field; reuse existing fixtures and test ids.
- Never use real tokens; fixtures use obvious fake values.
