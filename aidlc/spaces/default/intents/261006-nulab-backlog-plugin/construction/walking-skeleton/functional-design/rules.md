# Business Rules — walking-skeleton (U1)

Inputs: `unit-of-work`, `unit-of-work-story-map`, `components`, `requirements`, `contract-summary`, `team-practices`, and answers Q1–Q9 in `functional-design-questions.md`.

```yaml
rules:
  # BR1 — space address
  - id: BR1.1
    statement: Only Backlog space hosts are accepted.
    category: validation
    applies_to: SpaceAddress
    trigger: Connect request
    logic: IF the parsed host is exactly one ASCII label of 1-63 characters from [a-z0-9-], not starting or ending with "-", followed by ".backlog.com", ".backlog.jp" or ".backlogtool.com" THEN accept, ELSE reject. The check runs on the parsed host, never on the raw string.
    violation: validation error on field spaceUrl, with the expected format
    source: FR1.2, project.md (Mandated)
  - id: BR1.2
    statement: The address is normalised in fixed steps, and anything other than a bare https host is rejected.
    category: validation
    applies_to: SpaceAddress
    trigger: Connect request
    logic: >
      1) Trim surrounding whitespace.
      2) IF the input contains "://" THEN the scheme is the text before it, compared case-insensitively; it must be "https".
         ELSE prepend "https://".
      3) Parse as a URL.
      4) Lower-case the host.
      5) Reject IF parsing fails, OR any of these is present: a port, userinfo, a query, a fragment, a path other than "" or "/", a host ending with ".", a host that is an IP address, or any non-ASCII character.
    violation: validation error on field spaceUrl
    source: FR1.2
  - id: BR1.3
    statement: A rejected address causes no network request.
    category: policy
    applies_to: Connect workflow
    trigger: BR1.1 or BR1.2 rejects
    logic: IF the address is rejected THEN return before any call to Backlog.
    violation: none (guarantee)
    source: FR1.2
  - id: BR1.4
    statement: Redirects are never followed.
    category: policy
    applies_to: every Backlog call
    trigger: Backlog answers with a 3xx status
    logic: IF the response is a redirect THEN do not follow it, send nothing further, and return Kind Unreachable.
    violation: error code unreachable
    source: FR1.2, NFR3

  # BR2 — connect
  - id: BR2.1
    statement: Only a Kandev admin can connect.
    category: authorization
    applies_to: connection.connect_api_key
    trigger: Connect request
    logic: The action is declared with admin access, so Kandev refuses non-admins before the plugin runs.
    violation: HTTP 403 from Kandev
    source: contract C5
  - id: BR2.2
    statement: A key is saved only after Backlog confirms it.
    category: policy
    applies_to: SpaceConnection, ApiKeySecret
    trigger: Connect request with a valid address and key
    logic: Call Myself with the key. IF it returns 200 with a body that maps to User (numeric id, non-empty name) THEN store per BR2.8 and return the connection view with the user name and space host.
    violation: see BR2.3, BR2.4, BR2.9, BR2.10
    source: FR1.3
  - id: BR2.3
    statement: A rejected key stores nothing and leaves any existing connection untouched.
    category: validation
    applies_to: Connect workflow
    trigger: Backlog returns 401 or 403
    logic: IF Myself returns Kind Unauthorized or Forbidden THEN return a validation error on field apiKey ("invalid key", with how to create one) and write nothing.
    violation: validation error on field apiKey
    source: FR1.3
  - id: BR2.4
    statement: An unreachable or misbehaving Backlog is never reported as a wrong key.
    category: policy
    applies_to: Connect workflow
    trigger: Myself returns Kind Unreachable or Invalid or Conflict
    logic: >
      IF the call fails with a network error, a timeout, a redirect, a 5xx, an oversized body,
      a 200 body that does not map to User, or any 4xx not covered by BR2.3, BR2.9 or BR2.10
      THEN return unreachable (with Retry in the UI) and write nothing.
    violation: error code unreachable
    source: NFR5
  - id: BR2.5
    statement: An existing connection is replaced only by a verified new one.
    category: policy
    applies_to: SpaceConnection, ApiKeySecret
    trigger: Connect while already connected
    logic: Verify the new key first (BR2.2). IF it succeeds THEN store per BR2.8 with connectionEpoch = previous + 1. IF it fails THEN keep the old secret and record unchanged. U1 does not ask for confirmation; U2 adds the confirmation that FR1.1 requires.
    violation: none (old connection kept)
    source: Q2, FR1.1
  - id: BR2.6
    statement: One Connect per workspace at a time.
    category: constraint
    applies_to: ConnectAttempt
    trigger: Connect request
    logic: IF a ConnectAttempt exists for the workspace THEN reject the new request; ELSE create one, and remove it when the Connect returns on every path.
    violation: error code conflict
    source: Q3, AC1.1.5
  - id: BR2.7
    statement: The connection is shared by the whole workspace.
    category: policy
    applies_to: SpaceConnection
    trigger: any member reads the connection
    logic: The connection is keyed by workspace only. Any authenticated member may read the connection view. There is no per-user connection.
    violation: none
    source: FR1.1
  - id: BR2.8
    statement: The secret and the record are linked by the epoch, written in a fixed order, and rolled back on failure.
    category: constraint
    applies_to: SpaceConnection, ApiKeySecret
    trigger: storing after a successful verification
    logic: >
      1) GetSecret the previous ApiKeySecret (it may be absent).
      2) SetSecret the new ApiKeySecret {apiKey, spaceHost, connectionEpoch = previous + 1, or 1}.
      3) Write the SpaceConnection record with the same spaceHost and connectionEpoch.
      4) IF step 2 fails in a way that may have written the secret, OR step 3 fails, THEN roll back on a fresh context with its own 2-second limit (NFR1.4), independent of the action's
         context: SetSecret the previous ApiKeySecret, or DeleteSecret when there was none. Then return internal.
      5) IF the rollback also fails THEN log a redacted "connection_inconsistent" event and return internal.
         BR2.11 makes the leftover state visible and harmless.
    violation: error code internal; the previous state is kept or flagged by BR2.11
    source: NFR5
  - id: BR2.9
    statement: A space that does not exist is reported on the address field.
    category: validation
    applies_to: Connect workflow
    trigger: Myself returns 404
    logic: IF Myself returns Kind NotFound THEN return a validation error on field spaceUrl ("space not found") and write nothing.
    violation: validation error on field spaceUrl
    source: FR1.2
  - id: BR2.10
    statement: A rate-limited Connect returns at once with a wait time.
    category: policy
    applies_to: Connect workflow
    trigger: Myself returns 429
    logic: >
      IF Myself returns Kind RateLimited THEN return rate_limited with retryAfterSeconds taken from
      X-RateLimit-Reset, else Retry-After, else 60. Write nothing and do not retry.
      Connect is a user action with a 3-second wait cap, and U1 has no rate-limit queue (contract-summary).
    violation: error code rate_limited
    source: NFR2, contract-summary (timeouts and retries)
  - id: BR2.11
    statement: A connection is valid only when the record and the secret agree.
    category: constraint
    applies_to: connection.get
    trigger: reading the connection
    logic: >
      IF there is no record THEN not_connected (a leftover secret is ignored and overwritten by the next Connect).
      IF there is a record AND a secret with the same connectionEpoch and spaceHost THEN connected.
      IF there is a record but the secret is missing or its epoch or host differs THEN state error ("connect again");
      such a connection is never used for a Backlog call.
    violation: state error in the connection view
    source: NFR5

  # BR3 — secrets
  - id: BR3.1
    statement: The key is stored only in Kandev's encrypted secret store.
    category: policy
    applies_to: ApiKeySecret
    trigger: storing a key
    logic: The key is written only through SetSecret; never to plugin state, settings or files.
    violation: none (design guarantee, checked by test)
    source: NFR3
  - id: BR3.2
    statement: The key is never returned to the UI.
    category: policy
    applies_to: connection.get, connection.connect_api_key
    trigger: any response
    logic: Responses carry the derived hasApiKey flag only. No part of the key value appears.
    violation: none (design guarantee, checked by test)
    source: NFR3
  - id: BR3.3
    statement: The key and any URL that carries it are redacted everywhere.
    category: policy
    applies_to: logs, errors, test output
    trigger: any log line or error
    logic: Before any text leaves the gateway, replace the key value and the whole query string of any Backlog URL with a redaction marker. Errors returned to the UI carry only an error code and an optional field name.
    violation: none (design guarantee, checked by test)
    source: NFR3, project.md (Mandated), Q1
  - id: BR3.4
    statement: The key is sent as Backlog's documented apiKey query parameter, over https only.
    category: policy
    applies_to: every Backlog call made with a key
    trigger: Backlog call
    logic: Add apiKey, query-encoded, to an https URL built from SpaceAddress. Never send it to any other host.
    violation: none
    source: Q1, contract C7
  - id: BR3.5
    statement: The key input is checked before any call.
    category: validation
    applies_to: connection.connect_api_key
    trigger: Connect request
    logic: Trim surrounding whitespace. IF the result is empty, longer than 256 characters, or contains anything other than printable ASCII without spaces THEN return a validation error on field apiKey and make no call.
    violation: validation error on field apiKey
    source: NFR3, NFR5

  # BR4 — calls
  - id: BR4.1
    statement: Every Backlog call is bounded in time.
    category: constraint
    applies_to: every Backlog call
    trigger: Backlog call
    logic: Each call ends after at most 10 seconds, or earlier when the caller cancels.
    violation: Kind Unreachable (timeout), or the caller's cancellation is returned unchanged
    source: NFR5, contract-summary (timeouts), AC8.1.3
  - id: BR4.2
    statement: Response bodies are size-limited.
    category: constraint
    applies_to: every Backlog call
    trigger: reading a response
    logic: Read at most 1 MiB. A larger body is treated as Kind Unreachable.
    violation: Kind Unreachable
    source: team-practices (Code Style)

  # BR5 — package, build and checks
  - id: BR5.1
    statement: The package contains the executables for exactly the five supported platforms.
    category: constraint
    applies_to: PluginPackage
    trigger: packaging
    logic: The package holds server/plugin-linux-amd64, server/plugin-linux-arm64, server/plugin-darwin-amd64, server/plugin-darwin-arm64 and server/plugin-windows-amd64.exe, all listed under the manifest runtime.executables, plus manifest.yaml and the UI bundle.
    violation: packaging fails
    source: NFR7, FR7.1
  - id: BR5.2
    statement: Package verification checks contents as well as checksums.
    category: validation
    applies_to: PluginPackage
    trigger: make verify-package
    logic: >
      Fail IF the package's SHA-256 differs from its line in dist/checksums.txt,
      OR any file inside differs from the per-file checksums stored in the package,
      OR any of the five executables, manifest.yaml or the UI bundle is missing,
      OR the manifest is not valid or does not list exactly the five executables.
    violation: verification fails
    source: FR7.1, AC7.1.1, AC7.1.3
  - id: BR5.3
    statement: A manual check is recorded in a fixed shape, without secrets.
    category: policy
    applies_to: ManualCheckRecord
    trigger: completing a manual check
    logic: Copy docs/manual-checks/TEMPLATE.md to docs/manual-checks/<YYYY-MM-DD>-<checkName>.md and fill in date, Kandev version, plugin commit, space domain, steps and result.
    violation: the skeleton checkpoint is not ready
    source: Q4, team-practices (Testing Posture)
  - id: BR5.4
    statement: The installed plugin shows its settings screen under Settings > Integrations, with the Backlog logo.
    category: policy
    applies_to: KandevAdapter
    trigger: plugin start
    logic: Register the Backlog integration settings card (label, description, BacklogLogo icon, the settings screen as its page, and the BR7.4 switch as its action) when the plugin starts.
    violation: none
    source: FR7.1, Q7, Q8
  - id: BR5.5
    statement: The build uses the pinned Kandev SDK.
    category: constraint
    applies_to: build
    trigger: make build and every target that compiles Go
    logic: go.mod replaces the Kandev SDK module with the ../kandev checkout. IF ../kandev is missing, OR its HEAD commit differs from the commit in .kandev-sdk-ref THEN the build stops with a message naming the expected commit.
    violation: build fails
    source: team-practices (Walking Skeleton), unit-of-work (U1 boundary)
  - id: BR5.6
    statement: U1 delivers every standard Makefile target, and coverage enforces the team floor.
    category: constraint
    applies_to: Makefile
    trigger: make coverage
    logic: Provide check-format, vet, lint, test, coverage, build, package and verify-package. coverage runs go test -race with a coverage profile over ./internal/... and ./server/..., excludes only the wiring in server/main.go (listed in the Makefile), and fails below 80% line coverage.
    violation: target fails
    source: team-practices (Testing Posture, Code Style), contract C4

  # BR6 — settings screen
  - id: BR6.1
    statement: The Connect button is locked while a Connect is running.
    category: policy
    applies_to: M1 settings screen
    trigger: Connect clicked
    logic: The button shows "Connecting..." and ignores clicks until the request returns.
    violation: none
    source: AC1.1.5
  - id: BR6.2
    statement: Every UI string comes from a message key; U1 ships English only.
    category: policy
    applies_to: M1 settings screen
    trigger: rendering
    logic: No literal user-facing text in components; English messages only in U1.
    violation: none
    source: Q5, NFR10
  - id: BR6.3
    statement: The API key field is empty after a Connect.
    category: policy
    applies_to: M1 settings screen
    trigger: Connect returns
    logic: Clear the key input on success and on failure. The page never shows a stored key.
    violation: none
    source: NFR3
  - id: BR6.4
    statement: Field errors and results are accessible.
    category: policy
    applies_to: M1 settings screen
    trigger: rendering errors and results
    logic: Each input has a visible label. A field error is linked to its input and marks it invalid. The Connect result is announced through a polite live region.
    violation: none
    source: NFR9
  - id: BR6.5
    statement: The settings screen uses consistent spacing.
    category: policy
    applies_to: M1 settings screen
    trigger: rendering
    logic: >
      Sections, fields, buttons and messages are laid out as one vertical stack with one consistent gap between items,
      and a smaller consistent gap between a label and its input, using the host UI kit and host styles only.
      The plugin ships no CSS framework of its own.
    violation: none
    source: Q9

  # BR7 — integration switch, home entry and page
  - id: BR7.1
    statement: Backlog can be turned on or off per workspace, and it is on by default.
    category: policy
    applies_to: IntegrationSwitch
    trigger: reading or changing the switch
    logic: IF no IntegrationSwitch record exists for the workspace THEN Backlog is on. Every ConnectionView carries the current value as enabled.
    violation: none
    source: Q7
  - id: BR7.2
    statement: Only a Kandev admin can turn Backlog on or off.
    category: authorization
    applies_to: connection.set_enabled
    trigger: switch changed
    logic: >
      connection.set_enabled is declared with admin access, so Kandev refuses non-admins before the plugin runs.
      It takes {enabled}; the workspace comes from the verified action context, never from the body.
      It writes the IntegrationSwitch record and returns the ConnectionView.
      A non-boolean enabled value is a validation error.
    violation: HTTP 403 from Kandev; validation error on field enabled
    source: Q7
  - id: BR7.3
    statement: While Backlog is off, the backend refuses Connect and every Backlog action for that workspace.
    category: policy
    applies_to: every plugin action except connection.get and connection.set_enabled
    trigger: an action request for a workspace whose switch is off
    logic: >
      Read the IntegrationSwitch before anything else. IF enabled is false THEN return integration_disabled at once,
      with no Backlog call, no secret read and no state write. connection.get and connection.set_enabled always work,
      so the UI can show the state and turn Backlog back on. Connect checks the switch again just before it stores
      anything, so a Connect that was already running when Backlog was turned off writes nothing.
      Later units add their actions under this same rule.
    violation: error code integration_disabled
    source: Q7
  - id: BR7.4
    statement: Turning Backlog off keeps the connection, and turning it back on restores it without a reconnect.
    category: policy
    applies_to: IntegrationSwitch, SpaceConnection, ApiKeySecret
    trigger: switch changed
    logic: Changing the switch writes only the IntegrationSwitch record. It never deletes or changes the SpaceConnection or the ApiKeySecret.
    violation: none
    source: Q7
  - id: BR7.5
    statement: The switch on the settings card reflects the stored value in every Kandev surface.
    category: policy
    applies_to: IntegrationSettingsCard action, PluginUI
    trigger: plugin start, workspace list change, switch change
    logic: >
      The card's action renders the host's drafted integration switch (id nulab-backlog) for the routed workspace.
      A flip marks the page as changed; the host calls persist when the user clicks Save. Persist calls
      connection.set_enabled and rejects with a message from the error code if it fails, so the host keeps the
      change unsaved and reports the failed save. Only after a successful load or save does the UI publish the value
      with setIntegrationEnabled("nulab-backlog", workspaceId, enabled). At start and whenever the workspace list
      changes, it loads and publishes the value for each workspace. A non-admin is refused by Kandev with 403.
    violation: the host reports the failed save; the stored and published values stay unchanged
    source: Q7
  - id: BR7.6
    statement: A Backlog entry in the home Integrations menu opens the Backlog page.
    category: policy
    applies_to: HomeIntegrationsEntry, BacklogPage
    trigger: plugin start
    logic: >
      Register the nav item {id backlog, path /backlog, section integrations, BacklogLogo icon} and the /backlog route.
      Both stay registered whether Backlog is on or off, so it can always be turned back on.
    violation: none
    source: Q6, Q6a
  - id: BR7.7
    statement: In U1 the Backlog page shows the workspace's connection status, whether Backlog is on, and a link to its settings.
    category: policy
    applies_to: BacklogPage
    trigger: opening /backlog
    logic: >
      Load connection.get for the active workspace. Show connected (space host and user name), not connected, or
      "connect again" (state error); show whether Backlog is on or off for this workspace; and link to the Backlog
      settings page of that workspace. IF no workspace is active THEN say so. IF the load fails THEN show an error with Retry.
    violation: none
    source: Q6a
  - id: BR7.8
    statement: The official Backlog logo is the plugin's icon.
    category: policy
    applies_to: BacklogLogo
    trigger: rendering the home entry, the page title and the settings card
    logic: Use the BacklogLogo SVG component from Nulab's official brand assets. It is aria-hidden next to a text label. The asset's source is recorded in the repository.
    violation: none
    source: Q8
```

## Rules Summary

| ID | Rule | Category |
|----|------|----------|
| BR1.1 | One ASCII label plus an allowed suffix, checked on the parsed host | validation |
| BR1.2 | Fixed normalisation steps; only a bare `https` host passes | validation |
| BR1.3 | A rejected address makes no network call | policy |
| BR1.4 | Never follow redirects | policy |
| BR2.1 | Only admins connect | authorization |
| BR2.2 | Save only after Backlog confirms the key | policy |
| BR2.3 | 401 or 403: the key is rejected, nothing stored | validation |
| BR2.4 | Unreachable or unexpected: never reported as a wrong key | policy |
| BR2.5 | Replace only with a verified new connection; confirmation comes in U2 | policy |
| BR2.6 | One Connect per workspace at a time | constraint |
| BR2.7 | One shared connection per workspace | policy |
| BR2.8 | Epoch-linked writes in a fixed order, with rollback | constraint |
| BR2.9 | 404: space not found, on the address field | validation |
| BR2.10 | 429: `rate_limited` at once, no retry | policy |
| BR2.11 | A connection is valid only when record and secret agree | constraint |
| BR3.1 | Key only in the encrypted secret store | policy |
| BR3.2 | Key never returned to the UI | policy |
| BR3.3 | Key and key-bearing URLs redacted everywhere | policy |
| BR3.4 | Key sent as query-encoded `apiKey` over `https` only | policy |
| BR3.5 | Key input trimmed and checked (1–256 printable ASCII) | validation |
| BR4.1 | 10-second limit per Backlog call | constraint |
| BR4.2 | 1 MiB response size limit | constraint |
| BR5.1 | Package holds exactly the 5 named platform executables | constraint |
| BR5.2 | Verification checks contents and checksums | validation |
| BR5.3 | Manual checks recorded from the template | policy |
| BR5.4 | Settings card under Settings > Integrations, with the logo and the switch | policy |
| BR5.5 | Build uses the pinned Kandev SDK | constraint |
| BR5.6 | All standard Makefile targets; coverage ≥ 80% with `-race` | constraint |
| BR6.1 | Connect button locked while running | policy |
| BR6.2 | Message keys; English only in U1 | policy |
| BR6.3 | Key field cleared after Connect | policy |
| BR6.4 | Accessible labels, field errors and result announcement | policy |
| BR6.5 | Consistent spacing using the host UI kit | policy |
| BR7.1 | Per-workspace on/off switch, on by default | policy |
| BR7.2 | Only admins change the switch (`connection.set_enabled`) | authorization |
| BR7.3 | While off, every action except get and set_enabled returns `integration_disabled` | policy |
| BR7.4 | Switching never touches the connection or the key | policy |
| BR7.5 | The card switch shows the stored value and publishes it to Kandev | policy |
| BR7.6 | Home Integrations entry opens `/backlog`, always registered | policy |
| BR7.7 | `/backlog` shows status, on/off and a settings link in U1 | policy |
| BR7.8 | Official Backlog logo as the plugin icon | policy |

## Sources

- [desc] Initial description: Kandev plugin for Nulab Backlog.
- [scope] Workflow-selected scope: `feature`.
- [Q1]–[Q9]: answers in `functional-design-questions.md`.
- `unit-of-work.md`, `unit-of-work-story-map.md`, `components.md`, `requirements.md`, `contract-summary.md`, `stories.md`, `team-practices.md`, `project.md`.

## Assumptions & Open Questions

- [assumption] Backlog returns 401 for an invalid API key. BR2.3 also treats 403 as a rejected key, because during Connect no project-level permission is involved yet.
- [assumption] A request to a space that does not exist returns 404. If DNS fails instead, BR2.4 reports it as unreachable.
- [assumption] `integration_disabled` is a new error code next to the C5 list (`reconnect_required`, `rate_limited`, `unreachable`, `not_found`, `validation`, `conflict`, `internal`). Contract C5 is amended for it, together with the snake_case action keys (`connection.connect_api_key`, `connection.set_enabled`) that Kandev requires.
