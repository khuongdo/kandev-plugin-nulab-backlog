# Cross-Unit Traceability — Final Coverage Gate

## Verdict: FAIL

- 176 IDs enumerated: 36 FR (7 are the FR1–FR7 section headings), 11 NFR, 129 AC (every bold `ACn.m.k` in `stories.md`).
- **153 covered with status `OK`** by at least one unit's `construction/*/code-generation/traceability.json`, and every OK target path exists in the repo.
- **23 have no `OK` entry:**
  - 7 FR section headings (FR1–FR7). They are covered through their children.
  - 7 FR/NFR IDs that no unit lists directly (FR1.2, FR1.3, FR7.1, FR7.2, FR7.3, NFR6, NFR7). Each is covered indirectly through child IDs or ACs, but the walking-skeleton and ci-release traceability files never list the parent IDs (ci-release has empty `upstream_ids`).
  - 9 ACs marked `Deferred` (manual checks or GitHub-only steps).
- No `GAP`, `ORPHAN` or `N/A` statuses.

The gate fails because 9 ACs are only Deferred and 7 FR/NFR IDs lack a direct `OK` entry.

## Uncovered elements

### Not listed by any unit (indirect coverage)

| ID | Indirect evidence |
|----|-------------------|
| FR1–FR7 (headings) | all children of FR1–FR6 covered; FR7 children partly deferred |
| FR1.2 | US1.2 → AC1.2.1–AC1.2.3 OK; BR1.1, NFR3.4/NFR3.7 OK (walking-skeleton) |
| FR1.3 | US1.1 → AC1.1.1–AC1.1.7 OK (walking-skeleton, connection) |
| FR7.1 | AC7.1.1, AC7.1.3 OK (walking-skeleton); AC7.1.2 Deferred |
| FR7.2 | AC7.5.2, AC7.5.3 OK (ci-release); AC7.5.1 Deferred |
| FR7.3 | AC7.6.2 OK (ci-release); AC7.6.1 Deferred |
| NFR6 | NFR6.1 OK (walking-skeleton, `internal/plugin/manifest_test.go`); AC7.4.1, AC7.4.2 OK |
| NFR7 | NFR7.1 OK (walking-skeleton, `Makefile`); AC7.1.1 OK |

### Deferred

| ID | Unit | Justification |
|----|------|---------------|
| AC5.4.2 | git-pr | [manual] 320 px layout and screen-reader wording rendered by Kandev; B5 demo (TEMPLATE.md step 17) |
| AC5.6.2 | git-pr | [manual] real clone and push need real Kandev and Backlog; B5 demo (step 16), dependency X6 |
| AC7.1.2 | walking-skeleton | manual install through Kandev Settings > Plugins at the skeleton checkpoint |
| AC7.2.1 | walking-skeleton | manual check against a real Backlog space at the skeleton checkpoint (no record yet) |
| AC7.5.1 | ci-release | Release, attestation and `gh attestation verify` run only on GitHub after a tag push |
| AC7.6.1 | ci-release | catalogue PR sent after v0.1.0 |
| AC8.1.2 | issues | [manual] real 20-row list measured 20 times; B4 demo (step 25) |
| AC8.2.1 | issues | [manual] keyboard-only and screen-reader pass; B4 demo (step 26) |
| AC8.2.4 | issues | [manual] 320 px and contrast; B4 demo (step 26) |

## Per-ID coverage

"none" = no unit lists the ID; Target gives indirect evidence.

| ID | Status | Owning unit(s) | Target |
|----|--------|----------------|--------|
| FR1 | none (heading) | - | children FR1.1–FR1.7 |
| FR1.1 | OK | connection | internal/connection/lifecycle_test.go |
| FR1.2 | none | - | indirect: AC1.2.1–AC1.2.3, BR1.1 (walking-skeleton) |
| FR1.3 | none | - | indirect: AC1.1.1–AC1.1.7 (walking-skeleton) |
| FR1.4 | OK | connection | internal/connection/oauth_flow_test.go |
| FR1.5 | OK | connection | internal/connection/credentials_test.go |
| FR1.6 | OK | connection, git-pr | internal/connection/lifecycle_test.go; internal/connection/store.go |
| FR1.7 | OK | connection, git-pr, issues | internal/connection/projects_service_test.go; internal/git/events.go; internal/issues/types.go |
| FR2 | none (heading) | - | children FR2.1–FR2.3 |
| FR2.1 | OK | issues | internal/issues/service.go |
| FR2.2 | OK | issues | internal/issues/types.go |
| FR2.3 | OK | issues | internal/issues/service.go |
| FR3 | none (heading) | - | children FR3.1–FR3.5 |
| FR3.1 | OK | issues | internal/issues/service.go |
| FR3.2 | OK | issues | internal/issues/service.go |
| FR3.3 | OK | issues | internal/issues/service.go |
| FR3.4 | OK | issues | internal/issues/service.go |
| FR3.5 | OK | issues | internal/plugin/references.go |
| FR4 | none (heading) | - | children FR4.1–FR4.4 |
| FR4.1 | OK | issues | internal/issues/sync.go |
| FR4.2 | OK | issues | internal/issues/store.go |
| FR4.3 | OK | issues | internal/issues/sync.go |
| FR4.4 | OK | issues | internal/issues/sync.go |
| FR5 | none (heading) | - | children FR5.1–FR5.4 |
| FR5.1 | OK | git-pr | internal/git/service.go |
| FR5.2 | OK | git-pr | internal/git/service.go |
| FR5.3 | OK | git-pr | internal/git/service.go |
| FR5.4 | OK | git-pr | internal/git/service.go |
| FR6 | none (heading) | - | children FR6.1–FR6.3 |
| FR6.1 | OK | git-pr | internal/git/service.go |
| FR6.2 | OK | git-pr | internal/git/watcher.go |
| FR6.3 | OK | git-pr | internal/git/service.go |
| FR7 | none (heading) | - | children FR7.1–FR7.3 |
| FR7.1 | none | - | indirect: AC7.1.1, AC7.1.3 OK; AC7.1.2 Deferred |
| FR7.2 | none | - | indirect: AC7.5.2, AC7.5.3 OK; AC7.5.1 Deferred |
| FR7.3 | none | - | indirect: AC7.6.2 OK; AC7.6.1 Deferred |
| NFR1 | OK | issues | internal/issues/list_test.go |
| NFR2 | OK | connection, git-pr, issues | internal/backlog/client.go; internal/backlog/issues.go; internal/backlog/limiter_test.go |
| NFR3 | OK | connection, git-pr, issues | internal/connection/events_test.go; internal/git/leak_test.go; internal/issues/leak_test.go |
| NFR4 | OK | connection, git-pr | internal/backlog/testdata/pullrequests_ok.json; internal/testutil/testutil.go |
| NFR5 | OK | connection, git-pr, issues | internal/backlog/client.go; internal/git/store.go; internal/issues/store.go |
| NFR6 | none | - | indirect: NFR6.1 OK (internal/plugin/manifest_test.go); AC7.4.1, AC7.4.2 OK |
| NFR7 | none | - | indirect: NFR7.1 OK (Makefile); AC7.1.1 OK |
| NFR8 | OK | connection, git-pr, issues | Makefile; internal/git/watcher_test.go; internal/issues/service.go |
| NFR9 | OK | connection, git-pr, issues | ui/src/git/watch-form.tsx; ui/src/issues/issues-page.tsx; ui/src/settings/confirm-dialog.test.tsx |
| NFR10 | OK | git-pr, issues | ui/src/issues/i18n.ts; ui/src/messages/en.ts |
| NFR11 | OK | connection, git-pr, issues | internal/backlog/limiter_test.go; internal/git/watcher.go; internal/issues/sync.go |
| AC1.1.1 | OK | walking-skeleton | internal/plugin/actions_test.go |
| AC1.1.2 | OK | walking-skeleton | internal/connection/service_test.go |
| AC1.1.3 | OK | walking-skeleton | internal/connection/service_test.go |
| AC1.1.4 | OK | walking-skeleton | internal/plugin/actions_test.go |
| AC1.1.5 | OK | walking-skeleton | ui/src/settings/settings.test.tsx |
| AC1.1.6 | OK | walking-skeleton | internal/plugin/actions_test.go |
| AC1.1.7 | OK | connection, walking-skeleton | internal/backlog/client_test.go; internal/connection/store_test.go |
| AC1.2.1 | OK | walking-skeleton | internal/connection/address_test.go |
| AC1.2.2 | OK | walking-skeleton | internal/connection/address_test.go |
| AC1.2.3 | OK | walking-skeleton | internal/backlog/client_test.go |
| AC1.3.1 | OK | connection | internal/connection/oauth_flow_test.go |
| AC1.3.2 | OK | connection | internal/connection/oauth_flow_test.go |
| AC1.3.3 | OK | connection | internal/connection/oauth_flow_test.go |
| AC1.4.1 | OK | connection | internal/connection/credentials_test.go |
| AC1.4.2 | OK | connection | internal/connection/credentials_test.go |
| AC1.4.3 | OK | connection, issues | internal/connection/credentials_test.go; ui/src/issues/issues-page.test.tsx |
| AC1.5.1 | OK | connection | ui/src/settings/connected-panel.test.tsx |
| AC1.5.2 | OK | connection | ui/src/settings/connected-panel.test.tsx |
| AC1.5.3 | OK | connection | ui/src/settings/connected-panel.test.tsx |
| AC1.5.4 | OK | connection, git-pr, issues | internal/connection/lifecycle_test.go; internal/connection/store_test.go; internal/issues/sync_test.go |
| AC1.6.1 | OK | connection | internal/connection/lifecycle_test.go |
| AC1.6.2 | OK | connection | internal/connection/lifecycle_test.go |
| AC1.6.3 | OK | connection | ui/src/settings/connected-panel.test.tsx |
| AC1.7.1 | OK | connection, git-pr, issues | internal/connection/projects_service_test.go; internal/git/repos_test.go; internal/issues/list_test.go |
| AC1.7.2 | OK | issues | internal/issues/list_test.go |
| AC1.7.3 | OK | connection | ui/src/settings/project-picker.test.tsx |
| AC1.8.1 | OK | git-pr, issues | ui/src/settings/connected-panel.test.tsx; ui/src/settings/connected-panel.tsx |
| AC1.8.2 | OK | connection, git-pr, issues | internal/connection/lifecycle_test.go; internal/connection/store_test.go; internal/issues/events_test.go |
| AC1.8.3 | OK | connection, git-pr, issues | internal/connection/lifecycle_test.go; internal/git/watcher_test.go; internal/issues/sync_test.go |
| AC1.9.1 | OK | git-pr, issues | ui/src/settings/project-picker.test.tsx; ui/src/settings/project-picker.tsx |
| AC1.9.2 | OK | connection, git-pr, issues | internal/connection/projects_service_test.go; internal/git/events_test.go; internal/issues/events_test.go |
| AC2.1.1 | OK | issues | internal/issues/list_test.go |
| AC2.1.2 | OK | issues | ui/src/issues/issues-page.test.tsx |
| AC2.1.3 | OK | issues | ui/src/issues/issues-page.test.tsx |
| AC2.1.4 | OK | issues | internal/issues/list_test.go |
| AC2.1.5 | OK | issues | ui/src/issues/issues-page.test.tsx |
| AC2.1.6 | OK | issues | ui/src/issues/issues-page.test.tsx |
| AC2.2.1 | OK | issues | internal/issues/list_test.go |
| AC2.2.2 | OK | issues | internal/issues/list_test.go |
| AC2.2.3 | OK | issues | internal/issues/list_test.go |
| AC2.2.4 | OK | issues | internal/backlog/issues_types_test.go |
| AC2.3.1 | OK | issues | internal/issues/list_test.go |
| AC2.3.2 | OK | issues | internal/issues/sync_test.go |
| AC3.1.1 | OK | issues | internal/issues/create_test.go |
| AC3.1.2 | OK | issues | internal/issues/create_test.go |
| AC3.1.3 | OK | issues | internal/issues/create_test.go |
| AC3.1.4 | OK | issues | internal/issues/create_test.go |
| AC3.2.1 | OK | issues | internal/issues/detail_test.go |
| AC3.2.2 | OK | issues | internal/issues/detail_test.go |
| AC3.2.3 | OK | issues | internal/backlog/issues_client_test.go |
| AC3.3.1 | OK | issues | ui/src/issues/link-task-dialog.test.tsx |
| AC3.3.2 | OK | issues | internal/issues/links_test.go |
| AC3.3.3 | OK | issues | internal/issues/links_test.go |
| AC3.3.4 | OK | issues | ui/src/issues/link-task-dialog.test.tsx |
| AC3.3.5 | OK | issues | internal/issues/links_test.go |
| AC3.4.1 | OK | issues | internal/issues/suggest_test.go |
| AC3.4.2 | OK | issues | internal/issues/suggest_test.go |
| AC3.4.3 | OK | issues | internal/issues/suggest_test.go |
| AC3.5.1 | OK | issues | internal/issues/detail_test.go |
| AC3.5.2 | OK | issues | internal/issues/detail_test.go |
| AC3.5.3 | OK | issues | internal/issues/detail_test.go |
| AC4.1.1 | OK | issues | internal/issues/sync_test.go |
| AC4.1.2 | OK | issues | internal/issues/sync_test.go |
| AC4.1.3 | OK | issues | internal/issues/sync_test.go |
| AC4.1.4 | OK | issues | internal/issues/sync_test.go |
| AC4.1.5 | OK | issues | internal/issues/sync_test.go |
| AC4.2.1 | OK | issues | internal/issues/sync_test.go |
| AC4.2.2 | OK | issues | internal/issues/types_test.go |
| AC4.2.3 | OK | issues | internal/issues/sync_test.go |
| AC4.2.4 | OK | issues | internal/issues/sync_test.go |
| AC5.1.1 | OK | git-pr | internal/git/repos_test.go |
| AC5.1.2 | OK | git-pr | ui/src/git/repository-provider.test.ts |
| AC5.2.1 | OK | git-pr | internal/git/links_test.go |
| AC5.2.2 | OK | git-pr | ui/src/git/pr-link.test.ts |
| AC5.2.3 | OK | git-pr | internal/git/links_test.go |
| AC5.3.1 | OK | git-pr | internal/git/types_test.go |
| AC5.3.2 | OK | git-pr | internal/git/create_test.go |
| AC5.3.3 | OK | git-pr | internal/git/create_test.go |
| AC5.3.4 | OK | git-pr | ui/src/git/create-pr.test.ts |
| AC5.4.1 | OK | git-pr | internal/git/status_test.go |
| AC5.4.2 | Deferred | git-pr | [manual] B5 demo, TEMPLATE.md step 17 |
| AC5.4.3 | OK | git-pr | internal/git/status_test.go |
| AC5.5.1 | OK | git-pr | internal/connection/git_credential_test.go |
| AC5.5.2 | OK | git-pr | internal/connection/git_credential_test.go |
| AC5.6.1 | OK | git-pr | internal/git/resolver_test.go |
| AC5.6.2 | Deferred | git-pr | [manual] B5 demo, TEMPLATE.md step 16 (dependency X6) |
| AC5.6.3 | OK | git-pr | internal/git/resolver_test.go |
| AC6.1.1 | OK | git-pr | internal/git/watches_test.go |
| AC6.1.2 | OK | git-pr | ui/src/git/watch-form.test.tsx |
| AC6.1.3 | OK | git-pr | internal/git/watches_test.go |
| AC6.1.4 | OK | git-pr | internal/git/watches_test.go |
| AC6.1.5 | OK | git-pr | ui/src/git/watches-page.test.tsx |
| AC6.2.1 | OK | git-pr | internal/git/watcher_test.go |
| AC6.2.2 | OK | git-pr | internal/git/watcher_test.go |
| AC6.2.3 | OK | git-pr | internal/git/watcher_test.go |
| AC6.2.4 | OK | git-pr | internal/git/watcher_test.go |
| AC6.2.5 | OK | git-pr | internal/git/watcher_test.go |
| AC6.3.1 | OK | git-pr | ui/src/git/dashboard-page.test.tsx |
| AC6.3.2 | OK | git-pr | internal/git/queries_test.go |
| AC7.1.1 | OK | walking-skeleton | Makefile |
| AC7.1.2 | Deferred | walking-skeleton | manual install at skeleton checkpoint (`make package`, README) |
| AC7.1.3 | OK | walking-skeleton | internal/pkgverify/pkgverify_test.go |
| AC7.2.1 | Deferred | walking-skeleton | manual real-space check (docs/manual-checks/TEMPLATE.md) |
| AC7.3.1 | OK | ci-release | .github/workflows/ci.yml |
| AC7.3.2 | OK | ci-release, connection, git-pr, issues | internal/plugin/actions_test.go; internal/backlog/projects_test.go; internal/backlog/git_client_test.go; internal/backlog/issues_client_test.go |
| AC7.3.3 | OK | ci-release | internal/backlog/client_test.go |
| AC7.3.4 | OK | ci-release | internal/ci/secrets_test.go |
| AC7.4.1 | OK | ci-release | internal/ci/contract_test.go |
| AC7.4.2 | OK | ci-release | internal/ci/contract_test.go |
| AC7.5.1 | Deferred | ci-release | .github/workflows/release.yml (GitHub after tag push) |
| AC7.5.2 | OK | ci-release | internal/ci/release_test.go |
| AC7.5.3 | OK | ci-release | internal/ci/release_test.go |
| AC7.6.1 | Deferred | ci-release | internal/ci/marketplace_test.go (catalogue PR after v0.1.0) |
| AC7.6.2 | OK | ci-release | internal/ci/marketplace_test.go |
| AC8.1.1 | OK | issues | internal/issues/list_test.go |
| AC8.1.2 | Deferred | issues | [manual] B4 demo, TEMPLATE.md step 25 |
| AC8.1.3 | OK | issues | internal/backlog/issues_client_test.go |
| AC8.2.1 | Deferred | issues | [manual] B4 demo, TEMPLATE.md step 26 |
| AC8.2.2 | OK | issues | ui/src/issues/issues-page.test.tsx |
| AC8.2.3 | OK | issues | ui/src/issues/link-task-dialog.test.tsx |
| AC8.2.4 | Deferred | issues | [manual] B4 demo, TEMPLATE.md step 26 |
| AC8.3.1 | OK | connection, git-pr, issues | internal/plugin/actions_u2_test.go; internal/plugin/actions_u3_test.go; internal/plugin/actions_u4_test.go |
| AC8.3.2 | OK | issues | internal/issues/sync_test.go |
| AC8.4.1 | OK | connection | internal/backlog/limiter_test.go |
| AC8.4.2 | OK | connection | internal/backlog/limiter_test.go |
| AC8.4.3 | OK | connection | internal/backlog/limiter_test.go |
| AC8.4.4 | OK | connection, issues | ui/src/issues/issues-state.test.ts; ui/src/settings/project-picker.test.tsx |
| AC8.5.1 | OK | issues | ui/src/issues/i18n.test.ts |
| AC8.5.2 | OK | issues | ui/src/issues/i18n.test.ts |
