# Cross-Unit Traceability — 261007-uiux-github-style

**Verdict: PASS** — every FR and NFR in `requirements.md` is covered with status `OK` in `construction/code-generation/traceability.json`, and every target file exists. User Stories did not run (refactor scope), so there are no AC IDs to check. This intent has no Units; the single owner is the stage-level Code Generation record.

| ID | Status | Target file | Exists |
|---|---|---|---|
| FR1 | OK | ui/src/settings/SettingsScreen.tsx | yes |
| FR1.1 | OK | ui/src/settings/SettingsScreen.tsx | yes |
| FR1.2 | OK | ui/src/settings/pr-watches-section.tsx | yes |
| FR1.3 | OK | ui/src/settings/issue-watches-section.tsx | yes |
| FR1.4 | OK | manifest.yaml | yes |
| FR1.5 | OK | ui/src/settings/sections.test.tsx | yes |
| FR1.6 | OK | ui/src/settings/SettingsScreen.tsx | yes |
| FR2 | OK | ui/src/page/BacklogPage.tsx | yes |
| FR2.1 | OK | ui/src/index.ts | yes |
| FR2.2 | OK | ui/src/index.test.ts | yes |
| FR2.3 | OK | ui/src/page/BacklogPage.tsx | yes |
| FR2.4 | OK | ui/src/issues/issues-page.tsx | yes |
| FR2.5 | OK | ui/src/git/pr-list.tsx | yes |
| FR2.6 | OK | ui/src/git/save-query-dialog.tsx | yes |
| FR2.7 | OK | ui/src/page/BacklogPage.tsx | yes |
| FR3 | OK | internal/issues/watcher.go | yes |
| FR3.1 | OK | internal/issues/watch.go | yes |
| FR3.2 | OK | internal/issues/watcher.go | yes |
| FR3.3 | OK | internal/issues/watcher_test.go | yes |
| FR3.4 | OK | internal/issues/watch_store.go | yes |
| FR3.5 | OK | internal/plugin/issue_actions.go | yes |
| FR3.6 | OK | internal/plugin/host_port.go | yes |
| FR4 | OK | internal/git/prs.go | yes |
| FR4.1 | OK | internal/git/prs.go | yes |
| FR4.2 | OK | internal/git/prs_test.go | yes |
| FR4.3 | OK | internal/plugin/git_actions.go | yes |
| FR5 | OK | ui/src/controls.test.ts | yes |
| FR5.1 | OK | ui/src/controls.test.ts | yes |
| FR5.2 | OK | ui/src/settings/section-parts.tsx | yes |
| FR5.3 | OK | ui/src/settings/issue-watch-dialog.tsx | yes |
| FR6 | OK | ui/src/brand/backlog-logo.tsx | yes |
| FR6.1 | OK | ui/src/brand/backlog-logo.tsx | yes |
| FR6.2 | OK | ui/src/index.ts | yes |
| FR6.3 | OK | ui/src/brand/backlog-logo.test.tsx | yes |
| NFR1 | OK | Makefile | yes |
| NFR2 | OK | internal/plugin/actions_watch_test.go | yes |
| NFR3 | OK | internal/issues/watcher.go | yes |
| NFR4 | OK | ui/src/page/backlog-lists.test.tsx | yes |
| NFR5 | OK | ui/src/host-ui.ts | yes |
| NFR6 | OK | internal/plugin/actions_watch_test.go | yes |

## Uncovered Elements

None at FR/NFR level. Known partial delivery below requirement level, accepted at the Code Generation gate: business rule BR1.4 (hide "Sign in with Nulab" when OAuth is not configured) is `Deferred` because `connection.get` does not report OAuth configuration.
