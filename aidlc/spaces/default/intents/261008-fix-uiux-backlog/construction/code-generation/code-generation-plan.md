# Code Generation Plan - 261008-fix-uiux-backlog

Zero-Unit express run (no Units Generation). Scoped from `inception/requirements-analysis/requirements.md` (FR1-FR5, NFR1-NFR5) and the code knowledge base `aidlc/spaces/default/codekb/kandev-plugin-nulab-backlog/`. Brownfield: every change is in place; no new package.

## Testing Contract

```json
{
  "version": 1,
  "methodology": "tdd",
  "source": "team",
  "ordering": "For each testable layer, we write a failing test first, then write the minimum code to make it pass, then refactor while the test stays green, before moving to the next behaviour or layer.",
  "scope": "express",
  "test_strategy": "minimal",
  "project_type": "brownfield",
  "applicable_notes": [
    {
      "layer": "org",
      "text": "We treat tests as a first-class deliverable in every Bolt. The specific\nmethodology (TDD, BDD, ATDD, or classic test-after) is affirmed at\npractices-discovery and recorded in `team.md` under this heading with explicit\n`Methodology` and `Ordering` fields; Code Generation resolves those fields\nindependently from coverage, tooling, and scope notes.\n\nWhen no posture has been affirmed, our default per scope is:\n- **Methodology**: test-after\n- **Ordering**: implement each applicable testable layer, then write and run\n  that layer's tests.\n- `mvp`, `enterprise`, `feature`, `infra`, `classic` add an 80% line-coverage\n  floor and CI execution before merge.\n- `bugfix`, `security-patch` add a targeted regression for the specific\n  bug/vulnerability and require the existing suite to remain green.\n- `express` uses the Minimal strategy: requirement-driven unit tests (one per\n  requirement, with a happy-path floor per component); existing tests remain\n  green.\n- `poc`, `refactor`, `workshop` add no extra new-test floor and require the\n  existing suite to remain green.\n\nThe active `Test Strategy` still applies in every scope and determines test\nvolume/types. Scope floors are additive; they never reduce or replace the\nselected strategy.\n\nBuild and Test verifies defined coverage floors and affirmed quality targets;\nthey may not be weakened to make a step pass.\n\nAffirm a stricter posture in `team.md` if the team commits to one."
    },
    {
      "layer": "team",
      "text": "- We treat tests as a deliverable of every Bolt, not extra work.\n- **Methodology**: tdd\n- **Ordering**: For each testable layer, we write a failing test first, then write the minimum code to make it pass, then refactor while the test stays green, before moving to the next behaviour or layer.\n- A floor of **80% line coverage for the Go code**, measured with `go test -coverprofile` and enforced by a `make coverage` target; CI blocks the merge when it is lower. The measured scope is `./internal/...` and `./server/...`; only the minimal wiring in `main` is excluded, and every exclusion must be listed explicitly in the `Makefile`. We do not lower the floor or add exclusions to make it pass.\n- Go tests always run with **`go test -race`** (needs CGO; works on the `ubuntu-latest` runner). This is an addition to the Bitbucket template, because token refresh and API rate limiting are two places prone to data races.\n- **Automated contract test with the real package on the minimum Kandev version**, like the template's `packaged-host-contract` job: CI checks out Kandev at exactly the `min_kandev_version` in `manifest.yaml`, builds a throwaway server, installs the packaged plugin, and runs it.\n- **Manual end-to-end check against a real Backlog space exactly twice**: when the walking skeleton is done, and before the first release. Each result is recorded. After the first release, we rely only on automated tests.\n- Unit and integration tests do not call real Backlog: they use a **fake Backlog server built with `net/http/httptest`**, with JSON sample data in `internal/<package>/testdata/`, including the 401, 429 (with `Retry-After`), and expired-token error cases.\n- Tools: `testing` + `github.com/stretchr/testify/require`, table-driven tests with `t.Run`, test names that describe behaviour. The clock and wait functions are injected into the code so rate limits and token expiry can be tested without a real `time.Sleep`. Tests are independent of each other and use `t.TempDir()` and `t.Setenv()`. The TypeScript UI (if any) is tested with Vitest.\n- There is a test asserting that API keys and tokens do not appear in logs, error messages, or responses sent to the UI.\n- We **do not add a dependency vulnerability scan** (`govulncheck`, `npm audit`, Dependabot) — this is your decision in Q4."
    },
    {
      "layer": "project",
      "text": "- Keep the coverage profile out of the repository root during AI-DLC stages (delete coverage.out after local make coverage runs, or write it under build/); reviewers never use -coverprofile, because an unclaimed coverage.out blocks the Code Generation gate (learned 2026-10-07) \n\n- Install the Go toolchain (Go 1.26.x) and link ../kandev to the pinned v0.96.0 checkout before Reverse Engineering, so the scan records a Go test baseline instead of skipping it (learned 2026-10-07) \n\n- Run the packaged-host contract test locally with make contract-test KANDEV_MIN_DIR=../kandev while the SDK checkout is at the minimum version tag, and run it 10 times to catch host startup races (learned 2026-10-07)"
    }
  ],
  "obligations": {
    "strategy": "minimal",
    "strategy_volume": [
      "One verifiable test per requirement at the narrowest effective level.",
      "At least one happy-path unit test per component.",
      "Unit tests are the default; a bugfix/security scope floor may require an integration or E2E regression when that is the narrowest level that reproduces the defect."
    ],
    "scope_floor": [
      "Keep the existing test suite green.",
      "This scope adds no extra new-test floor beyond the selected test strategy."
    ],
    "combination_rule": "Apply every selected-strategy obligation and every scope-floor obligation; neither replaces the other, and a targeted scope regression may add the narrowest necessary test type beyond the strategy default."
  },
  "plan_profile": {
    "methodology": "tdd",
    "runner_step": "Verify the existing test runner/configuration and record the exact unit-scoped command.",
    "runner_ready_before_first_test": true,
    "testable_layers": [
      "Data model / database behavior",
      "Repository / data access",
      "Business logic",
      "API / endpoint",
      "Frontend behavior"
    ],
    "steps": [
      "Project structure and production configuration skeleton.",
      "Verify the existing test runner/configuration and record the exact unit-scoped command.",
      "Data model / database behavior - Red: write the failing tests and record the failing command output.",
      "Data model / database behavior - Green: implement only enough behavior to pass.",
      "Data model / database behavior - Refactor: improve the implementation while tests stay green.",
      "Repository / data access - Red: write the failing tests and record the failing command output.",
      "Repository / data access - Green: implement only enough behavior to pass.",
      "Repository / data access - Refactor: improve the implementation while tests stay green.",
      "Business logic - Red: write the failing tests and record the failing command output.",
      "Business logic - Green: implement only enough behavior to pass.",
      "Business logic - Refactor: improve the implementation while tests stay green.",
      "API / endpoint - Red: write the failing tests and record the failing command output.",
      "API / endpoint - Green: implement only enough behavior to pass.",
      "API / endpoint - Refactor: improve the implementation while tests stay green.",
      "Frontend behavior - Red: write the failing tests and record the failing command output.",
      "Frontend behavior - Green: implement only enough behavior to pass.",
      "Frontend behavior - Refactor: improve the implementation while tests stay green.",
      "Environment/build configuration.",
      "Documentation and traceability."
    ]
  },
  "input_sha256": "sha256:4442c5fad82de74b31fc5e573b07e7444027e7437773b5002fb422bdfb2749f7",
  "contract_sha256": "sha256:6db87aafc6be218a4f8f762cfd33d3fec2e691661f16cdc25e6dcc4236038b2d"
}
```

## Approach

- Reuse, do not rebuild: the existing `IssueBadge` (`ui/src/issues/issue-badge.tsx`) already reads `{ taskId, workspaceId }` from `slotProps`, and both `task-row-metadata` and `chat-top-bar` pass those fields, so one component is registered for three slots (`task-card-tags`, `task-row-metadata`, `chat-top-bar`) and keeps sharing one `LinksStore` (NFR2).
- Hover card: wrap the badge in the host UI kit `Tooltip` / `TooltipTrigger` / `TooltipContent` (exposed to plugins in Kandev 0.96.0 `lib/plugins/host-api.ts`), opening on hover and keyboard focus (NFR4). Content: key, summary (when known), status.
- Summary: new optional `Summary` on the stored `Link` and on `LinkView` (`omitempty`, NFR5); set when a link is created and refreshed by the status sync loop (FR2).
- Home > Integrations: `initialize` becomes async-safe for the nav item only: before `registerNavItem`/`registerRoute`, await `connection.get` for every workspace with a bounded timeout; register when any workspace is ON, on any error or timeout (fail open, FR3.4, NFR3). Settings card always registered (FR3.3). If Kandev's initialize does not await a promise, the check is done synchronously-bounded inside the staged generation per the host contract; Step 1 verifies which.
- PRs in the top bar (FR5.4): no new UI; verify with a test that every linked PR summary from `git.prs.status` carries `taskStatus` (already set in `internal/git/service.go`).

## Steps

Each step follows the Testing Contract (TDD: Red, Green, Refactor). Test commands are in `unit-test-instructions.md`.

### Step 0: Environment and test baseline
- [x] Make the toolchain runnable: Go 1.26.x on PATH, `../kandev` linked to `~/repo/kandev` (v0.96.0, f099a46 = `.kandev-sdk-ref`), `npm ci` in `ui/`.
- [x] Record the baseline: `go test -race ./internal/issues/... ./internal/git/... ./internal/plugin/...` and the scoped Vitest files below; record counts in `code-summary.md`.
- [x] Verify in Kandev 0.96.0 source whether `KandevPlugin.initialize` may return a Promise that the host awaits inside the staged registration generation (`lib/plugins/host.ts`); pick the FR3 implementation accordingly.

### Step 1: FR2 - issue summary on links (Go, internal/issues)
- [x] Red: `service_test.go` - a new link to an issue with summary "Fix login" returns `summary` in `Links`; a stored link without summary still decodes and returns no summary.
- [x] Red: `sync_test.go` - the status refresh writes the current summary (old link gains it; renamed issue updates it).
- [x] Red: `leak_test.go` - a summary never appears in error messages or error responses (NFR1).
- [x] Green: add `Summary string \`json:"summary,omitempty"\`` to `Link` (types.go) and `LinkView` (service.go); set it in `newLink`; update it next to `LastKnownStatus` in sync.go; copy it in `Links`.
- [x] Refactor while green.

### Step 2: FR1.4 / FR5.2 - hover card content (UI, issues-state)
- [x] Red: `issues-state.test.ts` - a new `badgeHover(link)` returns key, summary and status; without summary returns key and status only, no empty line.
- [x] Green: add `summary?: string` to TS `LinkView`; add `badgeHover`; add catalogue keys in `messages/en.ts` if a label is needed.

### Step 3: FR1.4, FR1.5, NFR4 - badge with hover card (UI, issue-badge)
- [x] Red: `issue-badge.test.tsx` - the badge renders a Tooltip whose content has key, summary, status; the link still opens `https://<space>/view/<KEY>` in a new tab and stops propagation; accessible label includes key and summary; focus opens the card.
- [x] Green: wrap the existing badge in host `Tooltip`; keep the not-openable branch behaviour.

### Step 4: FR1.1-FR1.3, FR5.1, FR5.3, FR5.5, FR5.6 - register the badge on task rows and the task top bar (UI, index)
- [x] Red: `index.test.ts` - `initialize` registers the badge component for `task-card-tags`, `task-row-metadata` and `chat-top-bar`; all three share one links store (one `issues.links.list` per workspace).
- [x] Red: `issue-badge.test.tsx` - with `task-row-metadata` props (`surface: "task-list"` and `"sidebar"`) and `chat-top-bar` props (`presentation: "desktop"` and `"mobile"`) the badge renders for a linked task and renders nothing for an unlinked task; mobile presentation uses the 44px/16px sizing.
- [x] Green: register the slots in `ui/src/index.ts`; add the mobile sizing branch to the badge.

### Step 5: FR5.4 - Backlog PRs in the top-bar status area (Go, internal/git; check only)
- [x] Red-check: `service_test.go` - for a task with one and with two linked PRs, every summary from the PR status call carries `taskStatus` (number, state). If the test already passes, keep it as the regression guard; no production change.

### Step 6: FR3 - Home > Integrations entry only when ON (UI, index)
- [x] Red: `index.test.ts` - replace "registers the entry and the route even when Backlog is off everywhere" with: OFF in every workspace -> no nav item, no `/backlog` route, settings card still registered; ON in one workspace -> nav item and route; `connection.get` fails or times out -> nav item and route (fail open).
- [x] Green: implement the load-time check in `ui/src/index.ts` with a bounded timeout constant; update the comment that cites BR5.4/BR7.6/BR7.8 to say FR3 supersedes them.

### Step 7: FR4 - settings page fixes (UI, settings)
- [x] Red: `sections.test.tsx` - section order starts Connection, Projects, then the rest unchanged; empty Issue watches shows `issueWatchesEmpty` ("No issue watches yet"); empty Issue watches and empty PR watches each have exactly one Add watch button (the header one, no `-empty-add`).
- [x] Green: move the Projects block right after Connection in `SettingsScreen.tsx`; add `issueWatchesEmpty` to `messages/en.ts` and use it in `issue-watches-section.tsx`; drop the empty-state `add(...)` child in `issue-watches-section.tsx` and `pr-watches-section.tsx`.

### Step 8: Quality gates and docs
- [x] `make check-format vet lint` (gofmt, go vet, golangci-lint+gosec, tsc, eslint, prettier) clean.
- [x] `make coverage` keeps the 80% Go floor (profile under build/, no coverage.out at repo root).
- [x] README: note the Home > Integrations entry rule (shown when Backlog is ON in a workspace at load; reload after toggling) and the badge on task rows/top bar.
- [x] Write `code-summary.md`, `source-manifest.json`, `traceability.json`.

## Traceability

| Step | Requirements |
|------|--------------|
| 1 | FR2.1, FR2.2, FR2.3, NFR1, NFR5 |
| 2 | FR1.4, FR5.2 |
| 3 | FR1.4, FR1.5, FR1.6, FR5.2, NFR4 |
| 4 | FR1.1, FR1.2, FR1.3, FR5.1, FR5.3, FR5.5, FR5.6, NFR2 |
| 5 | FR5.4 |
| 6 | FR3.1, FR3.2, FR3.3, FR3.4, FR3.5, NFR3 |
| 7 | FR4.1, FR4.2, FR4.3 |

## Files expected to change

- internal/issues/types.go, service.go, sync.go and their tests (service_test.go, sync_test.go, leak_test.go)
- internal/git/service_test.go (test only)
- ui/src/index.ts, index.test.ts
- ui/src/issues/issue-badge.tsx, issue-badge.test.tsx, issues-state.ts, issues-state.test.ts
- ui/src/settings/SettingsScreen.tsx, issue-watches-section.tsx, pr-watches-section.tsx, sections.test.tsx
- ui/src/messages/en.ts
- README.md
