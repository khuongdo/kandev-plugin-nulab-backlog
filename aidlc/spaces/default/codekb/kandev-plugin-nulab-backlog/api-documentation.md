# API Documentation — kandev-plugin-nulab-backlog

## Kandev Host Contract (`manifest.yaml`)

- `api_version: 2`, `runtime.type: binary`, 5 executables, `min_kandev_version: "0.96.0"`.
- Capabilities: `state`, `secrets`, `api_read: [tasks, repositories]`, `api_write: [tasks]`, `events: [task.deleted]`.
- `repository_providers: [nulab-backlog]`, 1 `reference_sources` entry, `config_schema` (OAuth client id/secret, public base URL), `ui.bundle: /ui/bundle.js`.

## Plugin Actions (77)

| Prefix | Count | Notes |
|---|---|---|
| `connection.*` | 9 | connect API key / OAuth, get, disconnect, projects, switch, Git credential; admin-only where they change state |
| `repositories.*` | 2 | repository provider |
| `git.*` | 19 | Backlog PRs, watches, queries |
| `issues.*` | 25 | list, tasks, links, sync, watches, quick actions, queries |
| `scm.*` | 22 | GitHub/GitLab/Bitbucket context |

Each action declares `scope` (workspace/task), `access` (authenticated/admin), `max_body_bytes` (8-256 KiB).

## Webhooks

- `oauth-callback`: GET, public, 1 KiB body limit.

## Kandev Install API (consumed by operators and by the contract test)

- `POST /api/plugins/install`, either:
  - multipart form with field `package` (the `.tar.gz`) — used by the web UI upload and by `internal/ci/contract.go` (expects 201, no `warning`); or
  - JSON `{"url": "<package url>"}` — the backend downloads the package itself (100 MiB cap, Kandev `service_install.go`).
- `GET /ready`, `GET /api/plugins/<id>`.
- Error mapping in the install handler: 409 / 400 / 500 only; it never returns 502.

### Known issue: upload install over a slow link

The multipart body is read fully before install, bounded by the server `ReadTimeout` (default 30 s, `server.readTimeout` / `KANDEV_SERVER_READTIMEOUT`, Kandev `common/config/catalog.go` line 61). A 29.5 MB upload slower than ~1 MB/s is cut: backend returns 400 `missing multipart field "package"`, the `tailscale serve` proxy returns 502. Details: [code-quality-assessment.md](code-quality-assessment.md#known-issue-plugin-install-502).

## Outbound APIs

- Backlog REST v2 (`internal/backlog`): ~17 paths — `users/myself`, `projects`, statuses, project users, `issues`, `issues/count`, comments, attachments, git repositories, pull requests and count, `oauth2/token`.
- GitHub / GitLab / Bitbucket REST (read-only): repos, PRs/MRs, user.
