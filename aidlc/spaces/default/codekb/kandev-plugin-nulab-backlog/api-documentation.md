# API Documentation — kandev-plugin-nulab-backlog

## Kandev Host Contract (`manifest.yaml`)

- `api_version: 2`, `runtime.type: binary`, 4 executables since v0.4.2 (5 up to v0.4.1), `min_kandev_version: "0.96.0"`.
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

The multipart body is read fully before install, bounded by the server `ReadTimeout` (default 30 s, `server.readTimeout` / `KANDEV_SERVER_READTIMEOUT`, Kandev `common/config/catalog.go` line 61). An upload slower than the timeout is cut: backend returns 400 `missing multipart field "package"`, the `tailscale serve` proxy returns 502. Details: [code-quality-assessment.md](code-quality-assessment.md#known-issue-plugin-install-502).

## GitHub Actions Triggers and Required Checks

| Workflow | Trigger (file lines) | Path filter | Concurrency | Jobs |
|---|---|---|---|---|
| `ci.yml` | `pull_request: branches: [main]`, `push: branches: [main]` (3-7) | none | none | `checks`; `packaged-host-contract` (`needs: checks`, 30 min) |
| `release.yml` | `push: tags: ['v*']` (6-8) | none | `group: release`, `cancel-in-progress: false` | `verify`; `contract` (`needs: verify`); `publish` (`needs: [verify, contract]`) |

- Both workflows: top-level `permissions: contents: read`. Only `release` / `publish` has `contents: write`, `id-token: write`, `attestations: write`.
- `checks`: checkout plugin + Kandev at SDK ref, setup-go, setup-node, `npm ci`, `go mod tidy` diff check, `make check-format vet lint test coverage check-secrets build package verify-package`, upload `plugin-package`.
- `packaged-host-contract`: checks out Kandev at SDK ref and at `v<min_kandev_version>`, downloads the package, `sha256sum -c`, `make verify-package contract-test KANDEV_MIN_DIR=../kandev-min`.
- `verify`: as `checks` with `fetch-depth: 0`, plus `make release-preflight TAG=$TAG` (`GH_TOKEN`), uploads `release-package`. `contract`: as `packaged-host-contract` on `release-package`. `publish`: attests `dist/nulab-backlog-*.tar.gz`, `gh release create --verify-tag --generate-notes` (prerelease when the tag has `-`).
- GitHub applies `paths` / `paths-ignore` filters only to branch pushes and pull requests, not to tag pushes, so a path filter on `release.yml` would have no effect.

### Required status checks (`main` ruleset)

Repository ruleset `24580280` (read via `gh api repos/khuongdo/kandev-plugin-nulab-backlog/rulesets/24580280`, 2026-10-08):

- `required_status_checks`: contexts **`checks`** and **`packaged-host-contract`** (integration 15368 = GitHub Actions), `strict_required_status_checks_policy: false`.
- Other rules: `deletion`, `non_fast_forward`, `pull_request` with `allowed_merge_methods: [squash]` and 0 required approvals.
- Classic branch protection is not used (`branches/main/protection` returns 404).
- A required check matches by job name. A job skipped by `if:` reports as success; a workflow that never starts reports nothing, leaving the check "Expected - waiting".

## Outbound APIs

- Backlog REST v2 (`internal/backlog`): ~17 paths — `users/myself`, `projects`, statuses, project users, `issues`, `issues/count`, comments, attachments, git repositories, pull requests and count, `oauth2/token`.
- GitHub / GitLab / Bitbucket REST (read-only): repos, PRs/MRs, user.
