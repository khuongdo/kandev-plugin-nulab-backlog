# Observability Design — walking-skeleton (U1)

Inputs: `observability-requirements` (NFR11.1–NFR11.6), `security-requirements` (NFR3.3), `tech-stack-decisions` (log/slog), `functional-spec` (U1), `contract-summary`, and `nfr-design-questions.md`.

## Log Pipeline

1. `log/slog` with a JSON handler writes to standard error, at level INFO by default. The `KANDEV_PLUGIN_LOG_LEVEL=debug` environment variable switches to DEBUG.
2. A wrapping handler runs every attribute through `internal/redact` before writing (`security-design.md`, NFR3.3).
3. Each action gets a `requestId`: a random 16-hex string created when the handler starts. It is attached to every log line written during the action, and returned in `ActionError.requestId` so a user report can be matched to the logs.

## Events

| ID | Event | Level | Fields |
|----|-------|-------|--------|
| NFR11.1 | (all) | — | `time`, `level`, `event`, `workspaceId`, `requestId` |
| NFR11.2 | `connect_succeeded` | INFO | `spaceHost`, `backlogUserId`, `connectionEpoch`, `durationMs` |
| NFR11.2 | `connect_failed` | WARN | `spaceHost` (only if valid), `errorCode`, `backlogStatus` (if any), `durationMs` |
| NFR11.3 | `backlog_call` | DEBUG | `method`, `path` (no query), `status`, `durationMs`, `errorKind`, `errorClass` |
| NFR11.4 | `connection_inconsistent` | ERROR | `previousEpoch`, `newEpoch` |
| NFR11.6 | `integration_switched` | INFO | `previous`, `enabled`, `durationMs` |
| NFR11.6 | `action_refused_disabled` | INFO | `action` (the action key). A Connect refused by the guard logs only this event. A Connect refused at its second switch read logs `connect_failed` with `errorCode` `integration_disabled`, so every Connect that passed the guard still has exactly one NFR11.2 outcome event |
| — | `plugin_started` | INFO | `version`, `platform` (`<goos>-<goarch>`), `sdkRef` |

NFR11.5 is enforced by construction: no event has a field for a key, a URL, a body or a display name. The leak tests (NFR3.3) also search for the fake user's display name.

Metrics, dashboards and alerts are out of scope for U1. They come with the Operation phase.

## Sources

- [desc] Initial description: Kandev plugin for Nulab Backlog.
- [scope] Workflow-selected scope: `feature`.
- `observability-requirements.md`, `security-requirements.md`, `tech-stack-decisions.md` (U1 NFR requirements); `functional-spec.md` (U1); `contract-summary.md`.

## Assumptions & Open Questions

- [assumption] Kandev passes environment variables such as `KANDEV_PLUGIN_LOG_LEVEL` to the plugin process. If it does not, the level stays INFO.
