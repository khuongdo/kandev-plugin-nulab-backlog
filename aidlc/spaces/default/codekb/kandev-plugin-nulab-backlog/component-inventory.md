# Component Inventory — kandev-plugin-nulab-backlog

Health ratings: healthy / at-risk / degraded. "At-risk" here means at risk for intent 261007-source-control-agnostic, not broken. No test baseline was recorded this run ([code-quality-assessment.md](code-quality-assessment.md#test-coverage-and-baselines)).

## Backend Components (Go)

### Server Entrypoint
- `server/main.go`. Calls `pluginsdk.Serve(plugin.NewRuntime())` only. Health: healthy. Not re-scanned.

### KandevAdapter
- `internal/plugin/`. Routes 55 actions through one `handlers` map (`guarded` refuses all but two while Backlog is off), OAuth webhook, `task.deleted`, Git credential RPCs (`credential.go`), host port incl. Repository. Builds one `backlog.Client` that satisfies all three domain gateway ports. Depends on: Connection, Issues, Git, BacklogGateway, Redact, `pluginsdk`. Health: at-risk (single combined gateway; credential RPCs gated by the Backlog switch).

### BacklogGateway
- `internal/backlog/`. Only outbound client (REST v2 + Git smart-HTTP probe), pinned to the space host; per-group rate limiter; OAuth. Depends on: Redact. Health: healthy for Backlog; not reusable for other hosts by design. Skimmed this run.

### Connection
- `internal/connection/`. One Backlog connection per workspace (space host, API key or OAuth, selected projects, epoch), the integration switch (opt-in), and the single Git credential secret `backlog.git.<ws>` valid only for the connected host; publishes `ConnectionChanged`. Depends on: BacklogGateway, Redact. Health: at-risk (Git credential lives inside the issue-tracker connection and is deleted with it).

### Issues
- `internal/issues/`. Issue list, tasks, links, sync, watches, quick actions, saved issue queries. Depends on: BacklogGateway, Connection, Redact; does not import Git. Health: healthy (not affected by this intent beyond PR-to-issue references). Skimmed this run.

### Git
- `internal/git/`. Repository provider (inspect, branches via PR-base heuristics), PR link/unlink/create/status/list, PR watches and watcher, saved PR queries, Git credential lease. Identity, validation, state map and gateway are Backlog-specific; single `ProviderID`. Depends on: BacklogGateway, Connection, Redact. Health: at-risk (primary change surface for this intent; large `service.go`).

### Redact
- `internal/redact/`. Masks secrets and Backlog URL query strings. Health: healthy. Not re-scanned.

### CI Tooling
- `internal/ci/`, `cmd/ci/`. Secret scan, release preflight, marketplace entry, workflow policy, packaged-host contract driver. Health: healthy. Not re-scanned.

### PackageVerify
- `internal/pkgverify/`, `cmd/verifypkg/`. Package verification. Health: healthy. Not re-scanned.

### TestUtil
- `internal/testutil/`. Fake keys and tokens. Health: healthy. Not re-scanned.

## UI Components (TypeScript)

### UI Registration
- `ui/src/index.ts`. All extension points ([api-documentation.md](api-documentation.md#kandev-ui-extension-points)); registers exactly one repository and one review provider. Health: healthy. Skimmed.

### UI Shared Kit
- `ui/src/layout.ts`, `ui/src/host-ui.ts`, `ui/src/icons.tsx`. Class constants, `hostUi(host)`, inline icons. Health: healthy. Not re-scanned.

### UI Brand
- `ui/src/brand/backlog-logo.tsx`. `PLUGIN_ICON`. Health: healthy. Not re-scanned.

### UI Page
- `ui/src/page/BacklogPage.tsx`. `/backlog` Issues / Pull requests tabs. Health: healthy. Not re-scanned.

### UI Settings
- `ui/src/settings/`. `SettingsScreen` stacks sections (connection, PR watches, issue watches, saved queries, quick actions, issue sync, Git access, projects). One Backlog-only, admin-only Git access section shown only while connected; no provider, repository or owner/workspace settings. Health: at-risk (per-provider settings land here).

### UI Issues
- `ui/src/issues/`. Issue list, quick actions, panel, badge, linking. Health: healthy. Not re-scanned.

### UI Git
- `ui/src/git/`. Repository provider (`BACKLOG_GIT_URL`), review provider, PR link task action, create-PR flow, PR list, watch form, Git access form; repository options from Backlog selected projects. Health: at-risk (provider id, URL matcher and repo picker are Backlog-only).

### UI Switch
- `ui/src/switch/`. Enable switch; `PLUGIN_ID` also used as provider registration id. Health: healthy. Not re-scanned.

### UI Messages
- `ui/src/messages/en.ts`. English catalogue. Health: healthy. Not re-scanned.

### UI Test Harness
- `ui/src/testing/harness.ts`. Fake host. Health: healthy. Not re-scanned.
