# Entities — walking-skeleton (U1)

Inputs:

- `unit-of-work`: the U1 boundary.
- `unit-of-work-story-map`: US7.1, US1.2, US1.1, US7.2.
- `components`: Connection, BacklogGateway, KandevAdapter, PluginUI.
- `requirements`: FR1.1–FR1.3, FR7.1, NFR3, NFR5, NFR7.
- `contract-summary`: C1, C2, C4, C5, C8.
- Answers Q1–Q9 in `functional-design-questions.md` (Q6–Q9 came from the change request after the first manual check).

U1 uses only the parts of the domain model that the API-key connection needs. OAuth, project selection and the poll interval come in U2.

## Domain Entities

```yaml
entities:
  - name: SpaceAddress
    kind: value-object
    description: A validated, normalised Backlog space host. It is built only from input that passes BR1.1 and BR1.2.
    attributes:
      - { name: host, type: string, required: true, constraints: "lower-case ASCII; exactly one label of 1-63 characters [a-z0-9-], not starting or ending with '-', followed by an allowed suffix" }
      - { name: suffix, type: enum, required: true, allowed: [backlog.com, backlog.jp, backlogtool.com] }
      - { name: baseUrl, type: string, required: true, constraints: "always https://<host>" }

  - name: SpaceConnection
    kind: entity
    description: The public record of the single Backlog connection shared by every member of one Kandev workspace. It holds no secret.
    identifier: workspaceId
    storage: "Kandev plugin state, scope workspace, key 'connection'"
    attributes:
      - { name: workspaceId, type: string, required: true, unique: true }
      - { name: spaceHost, type: string, required: true, constraints: "SpaceAddress.host" }
      - { name: authMethod, type: enum, required: true, allowed: [api_key], default: api_key, constraints: "oauth is added in U2" }
      - { name: connectedUserName, type: string, required: true, constraints: "copied from BacklogUser.name" }
      - { name: connectedUserId, type: integer, required: true, constraints: "copied from BacklogUser.id" }
      - { name: connectionEpoch, type: integer, required: true, min: 1, constraints: "increases by 1 on every successful connect or replace" }
      - { name: connectedAt, type: timestamp, required: true }
      - { name: schemaVersion, type: integer, required: true, default: 1 }
    constraints:
      - "At most one SpaceConnection per workspace (FR1.1)."
      - "Valid only when an ApiKeySecret exists with the same connectionEpoch and spaceHost (BR2.8)."

  - name: ApiKeySecret
    kind: entity
    description: The Backlog API key plus the connection identity it belongs to, held only in Kandev's encrypted secret store. One write to this entity is atomic.
    identifier: workspaceId
    storage: "Kandev secret store, plugin-owned key 'backlog.connection.<workspaceId>'"
    attributes:
      - { name: workspaceId, type: string, required: true, unique: true }
      - { name: apiKey, type: secret-string, required: true, constraints: "trimmed; 1-256 printable ASCII characters, no spaces (BR3.5); never returned to the UI, logs or errors" }
      - { name: spaceHost, type: string, required: true }
      - { name: connectionEpoch, type: integer, required: true, min: 1 }
    operations: "GetSecret, SetSecret and DeleteSecret, all available with capabilities.secrets (C8)"

  - name: IntegrationSwitch
    kind: entity
    description: Whether Backlog is turned on for one Kandev workspace. It is independent of the connection, so turning Backlog off keeps the connection and its secret, and turning it back on needs no reconnect.
    identifier: workspaceId
    storage: "Kandev plugin state, scope workspace, key 'integration'"
    attributes:
      - { name: workspaceId, type: string, required: true, unique: true }
      - { name: enabled, type: boolean, required: true, default: true, constraints: "an absent record means enabled (BR7.1)" }
      - { name: changedAt, type: timestamp, required: false }
      - { name: schemaVersion, type: integer, required: true, default: 1 }
    constraints:
      - "At most one IntegrationSwitch per workspace."
      - "Changed only through connection.set_enabled, which is admin-only (BR7.2)."

  - name: ConnectionView
    kind: value-object
    description: What connection.get, connection.connect_api_key and connection.set_enabled return to the UI (the C5 ConnectionView with the U1 fields; functional-spec WF2 step 6 lists the same fields). It is derived on every call and never stored.
    attributes:
      - { name: connected, type: boolean, required: true, constraints: "true only when state is connected" }
      - { name: state, type: enum, required: true, allowed: [not_connected, connected, error], constraints: "derived by BR2.11" }
      - { name: enabled, type: boolean, required: true, constraints: "IntegrationSwitch.enabled, default true; added to C5" }
      - { name: spaceHost, type: string, required: false, constraints: "present when state is connected or error" }
      - { name: authMethod, type: enum, required: false, allowed: [api_key], constraints: "present when a record exists" }
      - { name: connectedUserName, type: string, required: false, constraints: "present when state is connected" }
      - { name: hasApiKey, type: boolean, required: true, constraints: "derived; the key itself is never included (BR3.2)" }
      - { name: hasOAuthToken, type: boolean, required: true, default: false, constraints: "always false in U1" }
      - { name: hasGitCredential, type: boolean, required: true, default: false, constraints: "always false in U1" }
      - { name: connectionEpoch, type: integer, required: false, constraints: "present when a record exists" }

  - name: ConnectAttempt
    kind: entity
    description: An in-flight Connect for a workspace. It exists only while verification and storage run, and enforces one Connect at a time.
    identifier: workspaceId
    attributes:
      - { name: workspaceId, type: string, required: true, unique: true }
      - { name: startedAt, type: timestamp, required: true }
    constraints:
      - "In memory only. It ends when the Connect returns, on every path."

  - name: ManualCheckRecord
    kind: entity
    description: The record of a manual end-to-end check against a real Backlog space (AC7.2.1). It is a Markdown file in the repository.
    identifier: date + checkName
    storage: "docs/manual-checks/<YYYY-MM-DD>-<checkName>.md, copied from docs/manual-checks/TEMPLATE.md"
    attributes:
      - { name: date, type: date, required: true }
      - { name: checkName, type: enum, required: true, allowed: [walking-skeleton, first-release] }
      - { name: kandevVersion, type: string, required: true }
      - { name: pluginCommit, type: string, required: true }
      - { name: spaceDomain, type: string, required: true, constraints: "host only; never a key or a URL with a key" }
      - { name: steps, type: list<string>, required: true }
      - { name: result, type: enum, required: true, allowed: [pass, fail] }
      - { name: notes, type: string, required: false }

  - name: PluginPackage
    kind: entity
    description: The distributable plugin package produced by the build.
    identifier: version
    storage: "dist/<pluginId>-<version>.tar.gz plus dist/checksums.txt"
    attributes:
      - { name: pluginId, type: string, required: true, constraints: "manifest id" }
      - { name: version, type: string, required: true, constraints: "semver, read from the manifest version field" }
      - { name: executables, type: list<PlatformExecutable>, required: true, constraints: "exactly the 5 targets in BR5.1" }
      - { name: manifest, type: file, required: true, constraints: "manifest.yaml at the package root" }
      - { name: uiBundle, type: file, required: true, constraints: "the built PluginUI bundle" }
      - { name: minKandevVersion, type: string, required: true }

  - name: PlatformExecutable
    kind: value-object
    description: One server executable inside the package.
    attributes:
      - { name: platformKey, type: enum, required: true, allowed: [linux-amd64, linux-arm64, darwin-amd64, darwin-arm64, windows-amd64] }
      - { name: path, type: string, required: true, constraints: "server/plugin-<platformKey>, with .exe for windows-amd64; listed under manifest runtime.executables" }
```

## UI Registrations

These are not stored data, but they are part of U1's shape in Kandev, so they are listed here and specified in `functional-spec.md`.

```yaml
ui_registrations:
  - name: BacklogLogo
    kind: value-object
    description: The official Backlog logo as a plugin-owned SVG icon component, taken from Nulab's official brand assets [Q8]. It is decorative (aria-hidden) wherever a text label sits next to it.
    used_by: [HomeIntegrationsEntry, BacklogPage topbar, IntegrationSettingsCard]
  - name: HomeIntegrationsEntry
    kind: value-object
    description: The Backlog item in the Integrations menu on the Kandev home page [Q6].
    attributes:
      - { name: id, type: string, value: "backlog" }
      - { name: label, type: message-key, value: "integrationLabel" }
      - { name: path, type: string, value: "/backlog" }
      - { name: section, type: enum, value: integrations }
      - { name: icon, type: BacklogLogo }
  - name: BacklogPage
    kind: value-object
    description: The plugin page at /backlog [Q6a]. In U1 it shows the connection status, whether Backlog is on for the current workspace, and a link to the Backlog settings. U3 adds the issue list.
  - name: IntegrationSettingsCard
    kind: value-object
    description: The Backlog card under Settings > Integrations, with the BacklogLogo icon, the settings screen as its page, and the on/off switch as its action [Q7].
```

## Contract Types Owned by U1 (C1)

U1 owns contract C1 (`backlog.Client`) in `contract-summary`. These are the shapes U1 fixes. U2, U3 and U4 add methods but keep these types unchanged.

```yaml
contract_types:
  - name: Credentials
    owner: walking-skeleton
    fields:
      - { name: SpaceHost, type: string, constraints: "a SpaceAddress host only" }
      - { name: APIKey, type: secret-string, constraints: "empty when OAuth is used (U2)" }
      - { name: AccessToken, type: secret-string, constraints: "empty in U1; used by U2" }

  - name: User
    owner: walking-skeleton
    description: The result of Myself, mapped from GET /api/v2/users/myself.
    fields:
      - { name: ID, type: integer, source: "id", required: true }
      - { name: UserID, type: string, source: "userId", required: false }
      - { name: Name, type: string, source: "name", required: true, constraints: "non-empty" }

  - name: Error
    owner: walking-skeleton
    description: The only error type the client returns. Its text is already redacted (BR3.3) and never contains a response body or a URL with a query string.
    fields:
      - { name: Kind, type: enum, allowed: [Unauthorized, Forbidden, RateLimited, Unreachable, NotFound, Invalid, Conflict] }
      - { name: Status, type: integer, constraints: "HTTP status from Backlog; 0 for network errors and timeouts" }
      - { name: RetryAfter, type: duration, constraints: "set only for RateLimited" }
    status_to_kind:
      "401": Unauthorized
      "403": Forbidden
      "404": NotFound
      "409": Conflict
      "400, 422": Invalid
      "429": RateLimited
      "other 4xx, 3xx, 5xx, timeout, network error, oversized or unparsable body": Unreachable

  - name: CallClass
    owner: walking-skeleton
    allowed: [Interactive, Background]
    rule: "Every Client method takes a CallClass. U1's only caller (Connect) uses Interactive."
```

`Forbidden` is a separate kind from `Unauthorized`. This is an addition to the C1 kind list, and it lets later units tell "no permission on this project" apart from "bad credentials" (contract finding R-06).

C2 (`HostPort`) is not built in U1. No U1 workflow creates tasks, sets labels or resolves Git credentials, and `unit-of-work` does not list `HostPort` in the U1 boundary. It is deferred to U3, the first unit that creates tasks.

## Summary

- **SpaceAddress** is the only way a host reaches the gateway, so an invalid host can never be called.
- **SpaceConnection** is the public record. **ApiKeySecret** holds the key together with the host and epoch it belongs to, written in one atomic secret write. The epoch links the two (BR2.8).
- `hasApiKey` is not stored. It is derived when the connection view is built (BR3.2).
- **IntegrationSwitch** turns Backlog on or off per workspace, on by default. It is kept apart from the connection, so turning it off loses nothing [Q7].
- **ConnectionView** is the one response shape the UI reads; it now carries `enabled`.
- The **UI registrations** add the home Integrations entry, the `/backlog` page, the Backlog logo and the switch on the settings card [Q6]–[Q8].
- **ConnectAttempt** is a short-lived lock: one Connect per workspace at a time [Q3].
- **ManualCheckRecord** gives the manual check a fixed shape, stored under `docs/manual-checks/` [Q4].
- **PluginPackage** and **PlatformExecutable** define what CI builds and verifies.
- The C1 types are fixed here for the later units.

## Sources

- [desc] Initial description: Kandev plugin for Nulab Backlog.
- [scope] Workflow-selected scope: `feature`.
- [Q1]–[Q9]: answers in `functional-design-questions.md`.
- Kandev v0.96.0 plugin SDK: `registerNavItem` (section `integrations`), `registerRoute`, `registerIntegrationSettings` (`icon`, `action`), `host.ui.IntegrationEnabledControl`, `host.setIntegrationEnabled`.
- `unit-of-work.md`, `unit-of-work-story-map.md`, `components.md`, `requirements.md`, `contract-summary.md`, `stories.md`, `team-practices.md`.
- Kandev `docs/public/plugins-manifest.md` (`runtime.executables` keys `<goos>-<goarch>`) and `docs/public/plugins-authoring.md` (GetSecret, SetSecret, DeleteSecret; `.tar.gz` packages).

## Assumptions & Open Questions

- [assumption] The 5 platform targets are linux-amd64, linux-arm64, darwin-amd64, darwin-arm64 and windows-amd64, the same as the Bitbucket template; windows-arm64 is not built. NFR7 names the operating systems and architectures but not the exact five.
- [assumption] Nulab's brand guidelines allow the Backlog logo to identify the integration in a third-party plugin; the user confirms this before release [Q8]. If they do not, the icon falls back to a host built-in icon without any other design change.
- [assumption] The Backlog "current user" response carries a numeric `id` and a non-empty `name`. The first manual check against a real space confirms this.
