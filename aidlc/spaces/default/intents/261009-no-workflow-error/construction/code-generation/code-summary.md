# Code Summary — 261009-no-workflow-error

## Files Created / Modified

Created:
- `internal/plugin/workflow_actions.go` — `workflows.status` action → `{"hasWorkflow": bool}`.
- `internal/plugin/workflow_actions_test.go` — `HasWorkflow` and endpoint tests.
- `ui/src/error-alert.tsx`, `ui/src/error-alert.test.tsx` — shared `ErrorAlert` (host `Alert` `variant="destructive"`, `role="alert"`, `w-full min-w-0 whitespace-normal break-words`, optional `title`, action slot, `messageTestId`).
- `ui/src/git/save-query-dialog.test.tsx`.

Modified:
- Backend/manifest: `manifest.yaml` (`api_read: ["tasks", "repositories", "workflows"]`, `workflows.status` action), `internal/plugin/host_port.go` (`HasWorkflow`, `Workflows().List` with `Limit: 1`), `internal/plugin/actions_u4_test.go` (fake `Workflows()`), `internal/plugin/manifest_test.go`, `README.md` (Next release upgrade note).
- UI: `ui/src/page/start-task.tsx`, `ui/src/page/BacklogPage.tsx`, `ui/src/issues/issues-page.tsx`, `ui/src/issues/issue-panel.tsx`, `ui/src/issues/issue-prs.tsx`, `ui/src/issues/link-task-dialog.tsx`, `ui/src/git/pr-list.tsx`, `ui/src/git/scm-pr-list.tsx`, `ui/src/git/save-query-dialog.tsx`.
- UI tests: `ui/src/testing/harness.ts` (`data-steps` on the dialog double, `expectErrorAlert` helper), `start-task.test.tsx`, `backlog-page.test.tsx`, `backlog-lists.test.tsx`, `issues-page.test.tsx`, `issue-panel.test.tsx`, `link-task-dialog.test.tsx`, `pr-list.test.tsx`.

Generated (gitignored, not source): `build/coverage.out`, `build/coverage.filtered.out`, `build/ui/bundle.js`, `dist/nulab-backlog-0.6.0.tar.gz`, `dist/checksums.txt`.

## Key Implementation Decisions

- With no Kandev context, `TaskCreateDialog` gets `workflowId={null}`, `defaultStepId={null}`, `steps={[]}` (Kandev v0.96.0 requires these props; its sidebar passes the same when no steps are loaded) instead of omitting them as the plan said.
- Only an explicit `hasWorkflow === false` blocks the dialog with the `errorWorkflow` toast; any error or empty reply opens the dialog (FR1.4). Repeat clicks while the check runs are ignored (`useRef`).
- `workflows.status` stays behind the plugin's enabled/disabled guard like every other action; when Backlog is off it returns `integration_disabled` and the UI opens the dialog (fail-open).
- Host errors map to `internal` (500) like other host errors; the body carries only the code, never host text (NFR4).
- `HasWorkflow` and endpoint tests share `workflow_actions_test.go` instead of `host_port_test.go`; `ErrorAlert` lives at `ui/src/error-alert.tsx` (next to `host-ui.ts`) instead of `ui/src/ui/`.
- The page "not connected / off" alert now uses the destructive style, because FR2.2 classes it as a page-state error.

## Error-to-Style Mapping

| Surface | File | `en.ts` keys | Style |
|---|---|---|---|
| "+ Task" no workflow | `page/start-task.tsx` | `errorWorkflow` | toast |
| "+ Task" link failed | `page/start-task.tsx` | `taskNotLinked` | toast |
| Page load failed (+Retry) | `page/BacklogPage.tsx` | `pageLoadFailed` | page alert `backlog-page-error` |
| Page unavailable (+Open settings) | `page/BacklogPage.tsx` | `pageAlertTitle` + `pageNotConnected` / `pageOff` / incomplete | page alert `backlog-page-alert` |
| Saved query delete / set default failed | `page/BacklogPage.tsx` | gitNotice keys | toast |
| Issues load failed (+Retry) | `issues/issues-page.tsx` | `issuesLoadFailed` + issueNotice | page alert `backlog-issues-error` |
| Issues not connected / no project / sign in again | `issues/issues-page.tsx` | `pageNotConnected`, `issuesNoProject`, `issuesSignInAgain` | page alert `backlog-issues-state` |
| Issues refresh failed | `issues/issues-page.tsx` | issueNotice keys | toast |
| Task menu action failed (unchanged) | `issues/task-menu.ts` | issueNotice keys | toast |
| Link-task dialog | `issues/link-task-dialog.tsx` | issueNotice keys | dialog alert above footer |
| Issue panel load failed (+Retry) | `issues/issue-panel.tsx` | `issueLoadFailed` + issueNotice | page alert `backlog-issue-error` |
| Comments failed (+Retry) | `issues/issue-panel.tsx` | `commentsFailed` + issueNotice | section alert |
| Attachments failed | `issues/issue-panel.tsx` | issueNotice | section alert |
| Issue PRs load failed | `issues/issue-prs.tsx` | `prsLoadFailed` + scmNotice | section alert |
| Issue PR link / unlink failed | `issues/issue-prs.tsx` | `errorPrUrl`, scmNotice keys | toast |
| Backlog Git PR list failed (+Retry) | `git/pr-list.tsx` | `prsLoadFailed` + gitNotice | page alert |
| SCM PR list failed (+Retry) | `git/scm-pr-list.tsx` | `prsLoadFailed` + scmNotice | page alert |
| Save query dialog | `git/save-query-dialog.tsx` | gitNotice keys (`errorName` stays by its field) | dialog alert above actions |

Left unchanged: the rate-limit countdown (`rateLimitedRetrying`) and `issueUnavailableLong` / "Reconnect host" (states, not errors); review-provider and integration-switch (Kandev renders their errors); watch-form, scm-watch-form, git-access, poll-interval (rendered only on the Settings page, out of scope per FR2.5).

## Test Coverage Summary

- Red evidence: Step 3 `HasWorkflow undefined` (compile); Step 6 `undefined: actionWorkflowsStatus`; Step 9 Vitest 7 failed / 6 passed (regression, no-workflow toast, fail-open, link, link-fail, PR row); Step 12 unresolved `./error-alert`, then 1 failed / 3 passed for `title`; Step 14 16 failed across 7 files.
- Final: Go `-race` 13/13 packages ok, coverage 92.9% (floor 80%), new Go code 100%; Vitest 36 files, 462/462 passed; `make check-format vet lint` clean (golangci-lint 0 issues, tsc, ESLint, Prettier, actionlint); `make package verify-package` ok; `make contract-test KANDEV_MIN_DIR=../kandev` passed once on v0.96.0.

## Deviations from the Plan

- Dialog props passed as `null` / `[]` instead of omitted (Kandev requires them).
- Test and component file locations as noted above.
- Watch forms, git-access and poll-interval were not retouched: they render only on the Settings page, which FR2.5 excludes (the requirements listed them as Backlog-page surfaces in error).
- `save-query-dialog` is shared with Settings, so its new dialog alert also appears there.
- Step 21 (contract test x10, manual real-Kandev check at 320/360/768/1280 px) is left to Build and Test.
