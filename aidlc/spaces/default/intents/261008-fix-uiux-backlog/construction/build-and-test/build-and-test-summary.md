# Build and Test Summary - Fix UIUX

## Overall Status

Build, package verification, every test suite and the contract test on Kandev v0.96.0 pass. Prerequisites: Go 1.26.x, `../kandev` at v0.96.0, `npm ci` in `ui/` (see [build-instructions.md](build-instructions.md)).

## Test Type Inventory

| Type | Instruction file | Notes |
|---|---|---|
| Unit (one per requirement) | `../code-generation/unit-test-instructions.md` | Go `internal/issues`, `internal/git`; Vitest index, badge, issues-state, settings sections |
| Full existing suite | [build-instructions.md](build-instructions.md) | `make test` (Go `-race` + Vitest) |
| Packaged-host contract (integration) | [integration-test-instructions.md](integration-test-instructions.md) | Kandev v0.96.0, 10 runs |
| Bounded behaviour | [performance-test-instructions.md](performance-test-instructions.md) | NFR2 shared store, NFR3 3 s timeout |
| Security checks | [security-test-instructions.md](security-test-instructions.md) | NFR1 leak test, lint/gosec, secret scan |

Minimal strategy: no new integration, performance or security suites; the files above record which checks cover each concern.

## Coverage Expectations

Zero-Unit express run: one stage-level target, 80% Go line coverage over `./internal/...` and `./server/...`. Measured 92.8%. UI: one test per requirement (no numeric floor).

## Target Verification Matrix

| Target ID | Source | Expected | Actual | Evidence | Owning Stage | Verdict |
|---|---|---|---|---|---|---|
| T-COV | `code-generation-plan.md` > Testing Contract (team) | >= 80% Go line coverage | 92.8% | [test-results.md](test-results.md) | build-and-test | Met |
| T-RACE | Testing Contract (team) | Go tests green with `-race` | 709/709 pass | [test-results.md](test-results.md) | build-and-test | Met |
| T-CONTRACT | Testing Contract (team) + project note (10 runs) | contract test passes on `min_kandev_version` | 10/10 pass on v0.96.0 | [test-results.md](test-results.md) | build-and-test | Met |
| T-REQ-TESTS | Testing Contract obligations (Minimal) | one test per requirement | 28/28 covered | [cross-unit-traceability.md](cross-unit-traceability.md) | build-and-test | Met |
| T-SUITE | Testing Contract scope floor (express) | existing suite green | 709 Go + 387 UI pass | [test-results.md](test-results.md) | build-and-test | Met |
| T-LINT | Team Code Style | format, vet, lint clean | 0 issues | [test-results.md](test-results.md) | build-and-test | Met |
| T-PKG | Project Mandated rule | package verification passes | verifypkg OK | [test-results.md](test-results.md) | build-and-test | Met |

No `nfr-requirements/` or `nfr-design/` artifacts exist (express scope skips them).

## Readiness Assessment

- **Build-ready**: yes.
- **Test-ready**: yes; every target Met.
- **Deployment-ready**: yes for a minor release once Deployment Pipeline sets the version; the release notes should say the Home > Integrations entry now follows the Backlog switch (reload after a toggle) and the badge shows on task rows, the sidebar and the task top bar.

## Known Limitations

- The UI changes were verified against the fake Kandev host and the contract test (which does not render the browser UI). A look in a real Kandev 0.96.0 browser session (badge and hover card on Home > Tasks, sidebar and task top bar on desktop and phone; Integrations entry hidden when OFF) is still a manual check.
- If Kandev loads plugins before its workspace list is ready, the plugin sees no workspaces and shows the Integrations entry (fail open), so it may appear while Backlog is OFF everywhere; it disappears after a reload once workspaces are known.
- Turning Backlog on or off takes effect in the Integrations menu only after a page reload (Kandev v0.96.0 has no way to hide a plugin menu entry later).
