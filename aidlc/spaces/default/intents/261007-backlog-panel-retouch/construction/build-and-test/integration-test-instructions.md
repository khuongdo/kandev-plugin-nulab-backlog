# Integration Test Instructions — 261007-backlog-panel-retouch

## Applicability

Test strategy is Minimal, so no separate integration suite is generated. The one integration-level check this refactor needs is the existing **packaged-host contract test**, which is a Testing Contract obligation (team Testing Posture) and is run here.

## Packaged-Host Contract Test

- Setup: `../kandev` at tag `v0.96.0`; Go 1.26 with `GOTOOLCHAIN=local` (see build-instructions.md).
- Command (run 10 times to catch host start-up races, project rule):

```bash
for i in $(seq 1 10); do make contract-test KANDEV_MIN_DIR=../kandev || break; done
```

- Expected: every run prints `ci contract: OK nulab-backlog on Kandev v0.96.0`.
- Scope limit: the contract test builds Kandev, installs the package and exercises the backend; it does **not** render the plugin UI in a browser. Host components (`IntegrationListToolbar`, `IntegrationRepositoryFilter`, `TaskRowIndicator`, `Popover`) inside the plugin route are verified only by the manual check below.

## Manual Real-Host UI Check (owned by Deployment Execution)

On the self-hosted Kandev (≥ v0.96.0) with the new package installed and a real Backlog space connected:

1. `/backlog` issue list: toolbar is one row on desktop; typing does not reload; Enter reloads; Project/Status/Assignee dropdowns are searchable; Save query works.
2. Phone width (≤ 390 px): filters stack at full width; no horizontal scroll; no "Filters (n)" button.
3. Issue with 1 and with ≥2 linked tasks: task title shows; "Tasks (n)" menu lists them; clicking opens the task.
4. Issue title opens the Backlog issue in a new tab.
5. PR list (Backlog Git and each connected provider — GitHub, GitLab, Bitbucket): same toolbar layout without a search box; provider selector first; "Status (n)" popover; task indicator. Also check review findings R-06 (provider list "All repositories") and R-07.
6. Kanban card badge: shows key · status; click opens the Backlog issue in a new tab without opening the card; dragging the card by the badge does not start from the badge; unavailable badge shows its detail on focus/tap.

Record the result (date, Kandev version, pass/fail per item) in the Deployment Execution record.
