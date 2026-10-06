# Monitoring Design — walking-skeleton (U1)

Inputs: `observability-design` and `reliability-design` (U1 NFR design), `logical-components`, `functional-spec` (U1), `contract-summary`, and the answers in `infrastructure-design-questions.md`.

U1 monitoring is logs only, collected by Kandev from the plugin's standard error. Metrics, alerts and dashboards belong to the Operation phase. The rows below say what U1 provides for them.

## Metrics & KPIs

| Metric | Source | Threshold | Why it matters |
|--------|--------|-----------|----------------|
| Connect duration | `durationMs` on `connect_succeeded` and `connect_failed` | p95 ≤ 3 s (NFR1.2) | Checked in tests and during the manual check; no live collection in U1 |
| Connect outcomes by `errorCode` | `connect_failed` events | — | Shows whether failures come from keys, Backlog or storage |
| Inconsistent storage | `connection_inconsistent` events | Any occurrence | The only state that needs a person to reconnect |
| Switch changes | `integration_switched` events | — | Shows when and in which workspace Backlog was turned on or off |
| Actions refused while off | `action_refused_disabled` events | — | Explains "Backlog is off" reports from users |

## Alerts

| Alert | Condition | Severity | Routes to |
|-------|-----------|----------|-----------|
| None in U1 | — | — | Self-hosted Kandev has no alert routing for plugins. Operators read the Kandev logs |

## SLIs / SLOs

| SLI | SLO target | Measurement window |
|-----|------------|--------------------|
| None in U1 | — | The plugin runs on your server, and its availability is the server's. Performance targets (NFR1.1, NFR1.2) are checked by tests, not measured live |

## Logs & Tracing

- **Collection**: JSON lines on standard error, collected by Kandev into its own logs (`observability-design`). Level INFO by default; DEBUG via `KANDEV_PLUGIN_LOG_LEVEL=debug`.
- **Correlation**: every line carries `requestId`, which is also returned in `ActionError.requestId`, so an error the user sees can be found in the logs.
- **Tracing**: none. There is one process and one outbound call per action.
- **Dashboards**: none.
- **Manual check record**: `docs/manual-checks/<YYYY-MM-DD>-walking-skeleton.md`. It records the Kandev version (0.96.0), the plugin commit and the result. This is the evidence for the skeleton checkpoint. The second manual check, before the first release, also covers the switch, the home Integrations entry, the `/backlog` page, keyboard focus and contrast (NFR9.1), and the logo brand confirmation (Q6 of the NFR requirements).

## Sources

- [desc] Initial description: Kandev plugin for Nulab Backlog.
- [scope] Workflow-selected scope: `feature`.
- `observability-design.md`, `reliability-design.md`, `logical-components.md` (U1 NFR design); `functional-spec.md` (U1); `contract-summary.md`.

## Assumptions & Open Questions

None.
