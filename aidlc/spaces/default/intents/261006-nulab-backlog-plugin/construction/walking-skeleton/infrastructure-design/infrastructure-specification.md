# Infrastructure Specification — walking-skeleton (U1)

Inputs:

- U1 NFR design: `performance-design`, `security-design`, `scalability-design`, `reliability-design`, `observability-design`, `logical-components`.
- `components`: external dependencies.
- `functional-spec` (U1).
- `contract-summary` (C4, C8).
- Answers Q1–Q3 in `infrastructure-design-questions.md`.

The plugin owns no cloud resources. It runs as a process started by a self-hosted Kandev server, and it keeps all data in Kandev's stores. Its only other infrastructure is the build and check pipeline (`cicd-pipeline.md`).

## Deployment

| Facet | Choice | Rationale |
|-------|--------|-----------|
| Compute model | One plugin process per Kandev server, started by Kandev from the platform executable in the package (`runtime.executables`) | Kandev's plugin model (C8); no separate hosting |
| Networking | Ingress: only from Kandev over the local go-plugin gRPC channel. Egress: HTTPS (443) to the connected `*.backlog.com`, `*.backlog.jp` or `*.backlogtool.com` host only | `security-design` (NFR3.4); no listening port of its own |
| Storage | Kandev state store, workspace scope, keys `connection` and `integration`. Kandev secret store, key `backlog.connection.<workspaceId>`. No files, no database | `security-design` (NFR3.1), `scalability-design` (NFR5.5) |
| Environments | Local development (fake Backlog via `httptest`, no Kandev); CI (GitHub Actions, no Kandev in U1); manual check on your existing self-hosted Kandev server at exactly 0.96.0 with a real Backlog space [Q1, Q3] | `team-practices` (Testing Posture); NFR6.1 |
| Installation | Manual. A Kandev admin uploads `dist/<pluginId>-<version>.tar.gz` through Settings > Plugins | `team-practices` (Deployment: manual install on self-hosted Kandev) |
| IaC approach | None. Nothing is provisioned. The only configuration as code is the GitHub Actions workflow and the repository settings listed in `cicd-pipeline.md` | No cloud resources |
| Resource sizing | No sizing. The process is idle except during actions. Memory is bounded by the 1 MiB response limit and the per-workspace lock map | `performance-design`, `reliability-design` |

## Infrastructure Services

| Service | Role | Configuration | Notes |
|---------|------|---------------|-------|
| Kandev state store | Database (key-value) | Scope `workspace`, keys `connection` (SpaceConnection) and `integration` (IntegrationSwitch), JSON values with `schemaVersion: 1`. Requires `capabilities.state: true` | Owned by Kandev; Connect's store calls share a 2 s budget (`performance-design`) |
| Kandev secret store | Secrets vault | Plugin-owned key `backlog.connection.<workspaceId>`, encrypted by Kandev. Requires `capabilities.secrets: true` | GetSecret, SetSecret, DeleteSecret |
| Kandev plugin host | Process supervisor and action router | `min_kandev_version: "0.96.0"`, `api_version: 2`. Actions `connection.get` (authenticated), `connection.connect_api_key` and `connection.set_enabled` (admin), all `scope: workspace` | Kandev enforces the 15 s action limit, admin access and the verified workspace context |
| Backlog API v2 | External API | HTTPS only, allowlisted host, no redirects; a call is limited to 10 s and, during Connect, never runs past the 12 s deadline minus 2 s (about 9 s when the pre-call steps take their full 1 s) | Owned by Nulab; contract C7 |
| GitHub | Source hosting and CI | Public repository, protected `main`, one required check (`cicd-pipeline.md`) | Owned by you (external dependency X4 in `external-dependency-map`) |
| Kandev web app (plugin UI host) | Renders the plugin bundle | The bundle registers the settings card (icon and switch action), the home Integrations nav item and the `/backlog` route; the Backlog logo is inline in the bundle, with its source recorded in `docs/brand/backlog-logo.md` | No runtime asset requests (NFR3.10) |


No queue, cache, CDN, DNS or load balancer is used.

Shared Infrastructure is not applicable to U1. U1 creates the state and secret keys. Later units read them through `ConnectionReader` (contract C3) rather than sharing them directly.

## Sources

- [desc] Initial description: Kandev plugin for Nulab Backlog.
- [scope] Workflow-selected scope: `feature`.
- [Q1]–[Q3]: answers in `infrastructure-design-questions.md`.
- U1 NFR design (all six files); `functional-spec.md` (U1); `components.md`; `contract-summary.md`; `team-practices.md`.

## Assumptions & Open Questions

None.
