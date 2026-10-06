# Security Requirements — walking-skeleton (U1)

Inputs:

- `requirements`: NFR3, NFR4, FR1.2.
- `functional-spec` and `rules` for U1: BR1.x, BR2.x, BR3.x, BR4.x.
- `contract-summary`: C1, C5, C6, C8.
- `project.md`: Mandated and Forbidden rules.
- Answers Q1–Q6 in `nfr-requirements-questions.md`.

## Data Classification

| Data | Class | Where it lives | Notes |
|------|-------|----------------|-------|
| Backlog API key | Restricted | Kandev secret store only (ApiKeySecret) | Grants the full rights of the Backlog user |
| Space host | Internal | Plugin state (SpaceConnection), secret, logs | Not secret, but identifies the customer |
| Backlog user id (numeric) | Internal | Plugin state, logs | Used to identify the account in logs [Q3] |
| Backlog display name | Confidential (personal data) | Plugin state, connection view | Never logged [Q3] |
| Manual check record | Internal | Repository | Must not contain a key (NFR4) |
| Integration switch value | Internal | Plugin state (IntegrationSwitch), logs | Not secret [Q5] |
| Backlog logo | Public | UI bundle | Third-party brand asset; source and terms recorded [Q6] |

## Requirements

| ID | Requirement | Pass/fail criterion | Source |
|----|-------------|---------------------|--------|
| NFR3.1 | The API key is stored only through Kandev's encrypted secret store (SetSecret), never in plugin state, settings, files or environment variables | A test after Connect finds the key in the fake secret store and nowhere in the fake state store | NFR3, BR3.1 |
| NFR3.2 | No response sent to the UI contains the key or any part of it; the view carries only the derived `hasApiKey` flag | A test scans the JSON of every `connection.get`, `connection.connect_api_key` and `connection.set_enabled` response for the key and for any 4-character substring of it; none is found | NFR3, BR3.2 |
| NFR3.3 | The key and the query string of every Backlog URL are redacted before any log line, error message or test output is produced | A test captures all logs and errors from a run that passes through every row of the WF3 outcome table; the key never appears, and no Backlog URL appears with its query | NFR3, BR3.3, project.md (Mandated) |
| NFR3.4 | The key is sent only over `https`, only to a host that passed BR1.1, and never after a redirect | A test with a fake server that redirects to another host shows 0 requests at the other host; the gateway refuses `http` URLs | FR1.2, BR1.4, BR3.4, project.md (Mandated) |
| NFR3.5 | Only a Kandev admin (Kandev's instance-wide admin role) can call `connection.connect_api_key` and `connection.set_enabled`; any authenticated member can call `connection.get` | The manifest declares `access: admin` for Connect and set_enabled and `authenticated` for get; a contract test (U5) confirms that a non-admin gets 403 | contract C5, BR2.1, BR7.2 |
| NFR3.8 | `connection.get`, `connection.connect_api_key` and `connection.set_enabled` are declared with `scope: workspace`, and every action takes the workspace from the verified action context Kandev passes, never from the request body | The manifest declares `scope: workspace` for all three; a test sends a body naming another workspace and the action reads and writes only the context workspace | BR7.2, contract C5 |
| NFR3.9 | While Backlog is off for a workspace, every action except `connection.get` and `connection.set_enabled` returns `integration_disabled` before reading the secret or calling Backlog; Connect checks the switch again before storing | Tests with the switch off show 0 secret reads, 0 requests to the fake server and 0 state writes; a test that turns the switch off during the Myself call shows nothing stored | BR7.3 |
| NFR3.10 | The Backlog logo is an inline SVG compiled into the UI bundle; the plugin makes no runtime request for it, and the SVG contains no script, external reference or event handler | A test renders the logo and finds no `script`, `href`/`xlink:href` to an external URL, or `on*` attribute; a bundle check finds no Nulab URL fetched at runtime | Q6 |
| NFR3.6 | The plugin never echoes a Backlog response body to the UI or to logs | Tests for 4xx and 5xx with distinctive body text show that text in no response and no log | NFR3, contract-summary (error codes) |
| NFR3.7 | Inputs are validated before any network call: the address is checked per BR1.1 and BR1.2, and the key per BR3.5 | Each AC1.2.2 case and each invalid key form results in 0 requests to the fake server | FR1.2, BR1.3, BR3.5 |
| NFR4.1 | No real credential is committed to the repository, test data or test artifacts. Fixtures use clearly fake keys | The CI secret check (U5) passes, and every fixture key matches a fake pattern such as `test-api-key-*` | NFR4, project.md (Forbidden) |
| NFR4.2 | The manual check record contains the space domain only, never a key or a URL with a query | The template has no key field. A review of each record confirms this | NFR4, BR5.3 |

## Threat Model (STRIDE, U1 scope)

U1 has two trust boundaries:

- browser → Kandev → plugin, through actions;
- plugin → Backlog, over HTTPS.

| Threat | Element | Risk | Mitigation |
|--------|---------|------|------------|
| Spoofing: a non-admin sets the workspace connection or the switch | `connection.connect_api_key`, `connection.set_enabled` | Medium | NFR3.5 (admin access enforced by Kandev before the plugin runs) |
| Tampering: a request body names another workspace | All actions | Medium | NFR3.8 (workspace from the verified action context only) |
| Tampering: a Connect stores a connection after Backlog was turned off | Connect, switch | Low | NFR3.9 (switch read again before storing) |
| Tampering: a crafted `spaceUrl` makes the plugin call an attacker host (SSRF) and leak the key | Connect workflow | High | NFR3.4, NFR3.7 (allowlisted suffixes, parsed-host checks, no IPs, no redirects) |
| Repudiation: who changed the connection or the switch is unknown | Connect, set_enabled | Low | NFR11.2 and NFR11.6 log every Connect and switch change. Kandev records which actor called the action [Q5] |
| Information disclosure: the key leaks through logs, errors or responses | Gateway, KandevAdapter | High | NFR3.1, NFR3.2, NFR3.3, NFR3.6 |
| Information disclosure: a malicious package swap | Package | Medium | BR5.2 checksums and contents. Release attestation comes in U5 |
| Denial of service: a slow or huge Backlog response blocks the plugin | Gateway | Medium | NFR5.1 (10 s), NFR5.2 (1 MiB), BR2.6 (one Connect per workspace) |
| Elevation of privilege: the stored key is used for something other than the connection | Plugin | Low | U1 uses the key only in WF3. From U2 on, BR2.11 is checked before every use |
| Tampering: a malicious or modified logo SVG runs script in Kandev's page | UI bundle | Low | NFR3.10 (inline, script-free SVG; no runtime fetch) |

## Compliance

- There is no regulated data (no payment, health or special-category data).
- The Backlog display name is personal data. It is shown only to workspace members and is never logged [Q3].
- Disconnecting deletes it (U2). U1 replaces it on reconnect.
- The Nulab Terms of Use allow API use with the user's own key (C-R2 in `requirements`).
- The Backlog logo is a Nulab trademark. Its source URL and terms of use are recorded in `docs/brand/backlog-logo.md` next to the asset, and the pre-release manual check record states that the user confirmed Nulab's brand guidelines allow this use; without that statement the release falls back to a host built-in icon [Q6].

## Sources

- [desc] Initial description: Kandev plugin for Nulab Backlog.
- [scope] Workflow-selected scope: `feature`.
- [Q1]–[Q6]: answers in `nfr-requirements-questions.md`.
- `functional-spec.md`, `rules.md` (U1); `requirements.md`; `contract-summary.md`; `team-practices.md`; `project.md`.

## Assumptions & Open Questions

- [assumption] Kandev records which actor invoked each plugin action, so the plugin does not need its own audit log for Connect.
