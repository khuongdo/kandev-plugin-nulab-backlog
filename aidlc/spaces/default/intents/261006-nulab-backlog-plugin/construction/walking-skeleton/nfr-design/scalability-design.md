# Scalability Design — walking-skeleton (U1)

Inputs: `scalability-requirements` (NFR5.5–NFR5.7, NFR5.10), `functional-spec` (U1), `contract-summary` (C8), and `nfr-design-questions.md`.

| ID | Design |
|----|--------|
| NFR5.5 | Every record and secret key includes the workspace id: state scope `workspace` + key `connection`, and secret `backlog.connection.<workspaceId>`. No package-level variable holds connection data |
| NFR5.6 | The Connect lock is a map from workspace id to an in-flight marker, guarded by a mutex. Acquiring the lock is a non-blocking try: if a marker exists, return `conflict` at once. Markers are removed in a `defer`. Different workspaces never share a marker, so they run in parallel |
| NFR5.10 | The switch is keyed by workspace (state scope `workspace`, key `integration`); the guard reads only the context workspace's value, so one workspace's switch never affects another. The UI loads and publishes each workspace's value separately |
| NFR5.7 | The process is stateless apart from the lock map. A restart drops in-flight markers, which is safe because the in-flight Connect ends with the process |

Kandev runs one process per installed plugin, so U1 needs no cross-process locking.

## Sources

- [desc] Initial description: Kandev plugin for Nulab Backlog.
- [scope] Workflow-selected scope: `feature`.
- `scalability-requirements.md` (U1 NFR requirements); `functional-spec.md` (U1); `contract-summary.md`.

## Assumptions & Open Questions

None.
