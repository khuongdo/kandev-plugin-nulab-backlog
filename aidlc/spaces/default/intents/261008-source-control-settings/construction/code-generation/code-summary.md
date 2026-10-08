# Code Summary: Source Control Settings (one active service)

## Files Modified

No files were created or deleted; no new dependencies. Full list in `source-manifest.json`.

- Backend SCM: `internal/scm/store.go`, `types.go`, `errors.go`, `service.go`, `queries.go`, `links.go`, `watcher.go`
- Backlog Git: `internal/git/service.go`, `watcher.go`
- Plugin actions: `internal/plugin/scm_actions.go`, `runtime.go`, `git_actions.go`, `credential.go`, `manifest.yaml` (new `scm.active.set`; `testdata/v030/manifest.yaml` unchanged)
- UI: `ui/src/settings/source-control-section.tsx`, `state.ts`, `SettingsScreen.tsx`, `ui/src/git/git-state.ts`, `pr-list.tsx`, `watch-form.tsx`, `ui/src/messages/en.ts`, `ui/src/layout.ts`, `ui/src/icons.tsx`
- Docs: `README.md` upgrade note under "Unreleased"
- Tests: `internal/scm/*_test.go` (store, service, service_cli_account, harness, links, watcher), `internal/git/watcher_test.go`, `internal/plugin/actions_scm_test.go`, `ui/src/settings/source-control-section.test.tsx`, `sections.test.tsx`, `ui/src/git/pr-list.test.tsx`, `watch-form.test.tsx`

## Key Implementation Decisions

- Settings document gains `active` (`json:"active,omitempty"`), schema version stays 1. Effective service: stored value, else derived (0 connected external -> `backlog_git`, 1 -> that provider, 2-3 -> pending `""` with everything allowed).
- Guard at the `settings()` chokepoint plus `SetToken`, `UseCLI`, `SetMapping`; lists hide other providers' items; item operations on them return `InactiveError`. `RemoveToken` stays allowed and stores a derived active value first so the workspace does not switch silently.
- `scm.active.set` (admin) takes `{"service": ...}` and returns `{providers, active}`, the same shape `scm.providers.list` now returns.
- Backlog Git off while an external service is active: read-only git actions return empty results, other git actions return 409 `service_inactive` with `activeService`; `ResolveGitCredential` is refused and the credential binding is revoked. `internal/git` gets an injected `Active` func, so it does not import `internal/scm`.
- UI: Service selector (admin) / read-only line (member), confirm dialog on switch, pick notice while pending, one framed card with h3 heading, icon and status badge; repository labels name the service and mapped lines read `[GitHub] owner/name`. A reply without `active` is treated as pending, so older responses render the previous layout.

## Test Coverage Summary

- TDD Red excerpts recorded for each layer (build failures on the new store/service/action symbols; 11 failing new Vitest tests before Green).
- `make check-format vet lint test coverage`: pass; Go coverage 92.9% (floor 80%, only the existing `server/main.go` exclusion).
- UI `typecheck`, `lint`, `format:check`: pass; full Vitest run 34 files / 446 tests pass.
- Not run here: `make package verify-package` and the packaged-host contract test (10x) - left for Build and Test.

## Deviations from the Plan

- `scmStateNotConfigured` reworded to "Not connected" instead of adding `scmStateNotConnected`; extra keys `scmSwitchConfirm`, `scmPickBody`, `scmPickPlaceholder`.
- Git access heading in `SettingsScreen.tsx` changed h5 -> h4 to keep heading order under the new h3 card headings.
- Card "logo" is one generic Git-branch outline icon next to the service name, not brand marks, following the plugin's existing no-brand-marks policy (`docs/brand/backlog-logo.md`).
- `settings.test.tsx` needed no change (a missing `active` is treated as pending).
- Step 14 (helper deduplication) needed no change.

## Known Consequences

- While GitHub/GitLab/Bitbucket is active, tasks cannot fetch or push Backlog Git repositories (credential refused), as chosen in requirements Q8.
- In a pending workspace, removing tokens until one external provider remains makes it active automatically (FR3.2 rule).
