# Smoke Test Results — 261007-backlog-panel-retouch (release v0.4.1)

## Automated (release candidate)

| Check | Result | Evidence |
|---|---|---|
| CI `checks` on PR #11 (format, vet, lint, test, coverage, secrets, build, package, verify-package) | Pass | Actions run 37694739794 |
| CI `packaged-host-contract` on Kandev 0.96.0 | Pass | Actions run 37694739794 |
| Local contract test ×10 on Kandev 0.96.0 | 10/10 | `construction/build-and-test/test-results.md` |

## Manual Real-Host UI Check (pre-tag) — Passed (reported by the human, 2026-10-08)

Checklist: `construction/build-and-test/integration-test-instructions.md` § Manual Real-Host UI Check.

| # | Item | Result |
|---|---|---|
| 1 | `/backlog` issue list: one-row toolbar, typing does not reload, Enter reloads, searchable dropdowns, Save query | Pass |
| 2 | Phone width: filters stacked full width, no horizontal scroll, no "Filters (n)" button | Pass |
| 3 | Linked tasks: title shown, "Tasks (n)" menu, click opens task | Pass |
| 4 | Issue title opens the Backlog issue in a new tab | Pass |
| 5 | PR lists (Backlog Git and each connected provider): toolbar without search, provider selector first, "Status (n)", task indicator; review R-06/R-07 | Pass |
| 6 | Kanban badge: key · status, opens the issue in a new tab without opening the card, drag not started from the badge, unavailable badge shows its detail | Pass |

## Post-install Smoke Check (after v0.4.1 release) — Pending

Steps in `operation/deployment-pipeline/deployment-strategy.md` § Post-install Smoke Check; the human installs the released `v0.4.1` (GitHub Release asset `nulab-backlog-0.4.1.tar.gz`) and reports.
