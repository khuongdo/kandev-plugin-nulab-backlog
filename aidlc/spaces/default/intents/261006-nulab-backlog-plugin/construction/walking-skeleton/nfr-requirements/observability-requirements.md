# Observability Requirements — walking-skeleton (U1)

Inputs: `requirements` (NFR11, NFR3), `functional-spec` and `rules` for U1, `contract-summary`, `team-practices`, and answers Q3 and Q5 in `nfr-requirements-questions.md`.

U1 uses structured logs only. Metrics, dashboards and alerts belong to the Operation phase. Kandev captures what the plugin writes to its log.

| ID | Requirement | Pass/fail criterion | Source |
|----|-------------|---------------------|--------|
| NFR11.1 | Logs are structured JSON lines with at least `time`, `level`, `event`, `workspaceId` and `requestId`, written to the plugin's standard error | A test parses every captured line as JSON with these fields | NFR11 |
| NFR11.2 | Each Connect logs exactly one outcome event. `connect_succeeded` (INFO) carries `spaceHost`, `backlogUserId`, `connectionEpoch` and `durationMs`. `connect_failed` (WARN) carries `spaceHost` (when valid), `errorCode`, `backlogStatus` (when any) and `durationMs` | A test checks one event per Connect and its fields | NFR11, Q3 |
| NFR11.3 | Each Backlog call logs `backlog_call` (DEBUG) with method, path without query, status, `durationMs` and error kind. Each wait, which starts in U2, will log `backlog_wait` | A test checks the event and that no query string appears | NFR11, BR3.3 |
| NFR11.4 | A failed rollback logs `connection_inconsistent` (ERROR) with `workspaceId` and both epochs, and nothing else | A test with an injected double failure checks the event | BR2.8 |
| NFR11.6 | Each switch change logs one `integration_switched` (INFO) event with `previous`, `enabled` and `durationMs`; a refused action logs `action_refused_disabled` (INFO) with the action key; both also carry the NFR11.1 base fields | A test checks one event per change and per refused action, with the NFR11.1 base fields plus these fields and nothing else | Q5, BR7.3 |
| NFR11.5 | Logs never contain the API key, a URL query string, a Backlog response body, or the Backlog display name | The leak test of NFR3.3 also searches for the display name of the fake user | NFR3, NFR11, Q3 |

## Log Levels

| Level | Used for |
|-------|----------|
| ERROR | Inconsistent storage; unexpected internal errors |
| WARN | Failed Connect (any code) |
| INFO | Successful Connect; switch change; action refused because Backlog is off; plugin start with version and platform |
| DEBUG | Each Backlog call. Off by default |

## Sources

- [desc] Initial description: Kandev plugin for Nulab Backlog.
- [scope] Workflow-selected scope: `feature`.
- [Q3], [Q5]: answers in `nfr-requirements-questions.md`.
- `functional-spec.md`, `rules.md` (U1); `requirements.md`; `contract-summary.md`; `team-practices.md`.

## Assumptions & Open Questions

- [assumption] Kandev collects plugin standard error into its own logs, as the Bitbucket plugin relies on. The first manual check confirms this.
