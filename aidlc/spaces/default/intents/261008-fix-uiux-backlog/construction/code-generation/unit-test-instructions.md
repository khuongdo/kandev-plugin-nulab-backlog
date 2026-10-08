# Unit Test Instructions - 261008-fix-uiux-backlog

Test strategy: Minimal (one verifiable test per requirement, happy-path floor per component). Methodology: TDD (Testing Contract in `code-generation-plan.md`). Scope floor: keep the existing suite green.

## Framework Setup

- Go: stdlib `testing` + `github.com/stretchr/testify/require`, table-driven with `t.Run`, fake Backlog via `net/http/httptest`; always `-race` (CGO on).
- UI: Vitest + jsdom with the shared fake host `ui/src/testing/harness.ts` (`ui/vitest.config.ts`).
- Prerequisites: Go 1.26.x on PATH; `../kandev` is a checkout of Kandev v0.96.0 (`.kandev-sdk-ref` f099a46), for example a symlink to `~/repo/kandev`; `npm ci` run in `ui/`.

## How to Run This Change's Tests

Run from the repository root. These commands are scoped to the packages and files this change touches.

Go (FR2, FR5.4, NFR1, NFR5):

```bash
go test -race ./internal/issues/ ./internal/git/
```

UI (FR1, FR3, FR4, FR5.1-FR5.3, FR5.5, FR5.6, NFR2-NFR4):

```bash
(cd ui && npx vitest run src/index.test.ts src/issues/issue-badge.test.tsx src/issues/issues-state.test.ts src/settings/sections.test.tsx)
```

A TDD Red step runs the single affected test first, for example `go test -race ./internal/issues/ -run TestLinks` or `(cd ui && npx vitest run src/issues/issue-badge.test.tsx -t "hover")`, and records the failing output before Green.

## Tests to Add or Change (one per requirement)

| Requirement | Test file | Test |
|-------------|-----------|------|
| FR2.1, FR2.3, NFR5 | internal/issues/service_test.go | new link carries the summary in `Links`; an old stored link without summary still decodes |
| FR2.2 | internal/issues/sync_test.go | status refresh sets/updates the stored summary |
| NFR1 | internal/issues/leak_test.go | the summary never appears in error messages or error responses |
| FR5.4 | internal/git/service_test.go | one and two linked PRs: every summary carries `taskStatus` |
| FR1.4, FR5.2 | ui/src/issues/issues-state.test.ts | `badgeHover` returns key, summary, status; no summary -> key and status only |
| FR1.4, FR1.5, FR1.6, NFR4 | ui/src/issues/issue-badge.test.tsx | Tooltip content; new-tab link; stops propagation; focus opens; accessible label |
| FR1.1, FR1.2, FR5.1, FR5.3, FR5.5 | ui/src/issues/issue-badge.test.tsx | renders for `task-row-metadata` (task-list, sidebar) and `chat-top-bar` (desktop, mobile) props; nothing when unlinked |
| FR1.3, FR5.6, NFR2 | ui/src/index.test.ts | badge registered for `task-card-tags`, `task-row-metadata`, `chat-top-bar`; one `issues.links.list` per workspace |
| FR3.1-FR3.5, NFR3 | ui/src/index.test.ts | OFF everywhere -> no nav item/route, settings card kept; ON in one -> entry; error/timeout -> entry |
| FR4.1 | ui/src/settings/sections.test.tsx | Connection, Projects, then the remaining sections |
| FR4.2 | ui/src/settings/sections.test.tsx | empty Issue watches shows "No issue watches yet" |
| FR4.3 | ui/src/settings/sections.test.tsx | empty Issue and PR watches each show exactly one Add watch button |

Expected volume: about 15 new or changed tests.

## Coverage Targets

- Go: the team floor of 80% line coverage over `./internal/...` and `./server/...` stays met (`make coverage`, profile under `build/`; delete any `coverage.out` at the repo root).
- UI: no numeric floor; every requirement above has a test.

## Mocking and Stubbing

- Backlog: `httptest` fake server with JSON in `internal/issues/testdata/`; never real Backlog, never real credentials.
- Kandev host (UI): `ui/src/testing/harness.ts` fake host; stub `connection.get` per workspace (ON, OFF, rejecting, never resolving with fake timers) for FR3.
- Time: inject the clock / use Vitest fake timers for the FR3 timeout; no real sleeps.

## Test Data

- Issue `PROJ-12`, summary "Fix login", status "In Progress", space `https://example.backlog.com`.
- A stored link JSON without `summary` for the backward-compatibility case.
