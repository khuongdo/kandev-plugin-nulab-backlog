# Scalability Requirements — walking-skeleton (U1)

Inputs: `requirements` (FR1.1), `functional-spec` and `rules` for U1, `contract-summary` (C8), and the answers in `nfr-requirements-questions.md`.

U1 has a small load profile. Each workspace has at most one Connect running at a time. Settings reads happen when a person opens the screen.

| ID | Dimension | Requirement | Pass/fail criterion | Source |
|----|-----------|-------------|---------------------|--------|
| NFR5.5 | Workspaces per Kandev server | The plugin keeps state per workspace and has no shared global connection, so any number of workspaces can be connected independently | A test with two workspaces connects both to different spaces, and each `connection.get` returns its own space | FR1.1, BR2.7 |
| NFR5.6 | Concurrent Connects | At most one Connect runs per workspace (BR2.6). Connects in different workspaces run in parallel and do not block each other | A `-race` test runs Connect in two workspaces at once (both succeed) and twice in one workspace (one gets `conflict`) | BR2.6, NFR8 |
| NFR5.10 | Switch per workspace | The switch is stored per workspace; changing it in one workspace never affects another | A test turns Backlog off in one of two workspaces; Connect still works in the other | BR7.1, BR7.3 |
| NFR5.7 | Process state | Restarting the plugin loses nothing, because all connection data lives in Kandev's state and secret stores | A test creates a new plugin instance over the same fake stores and reads the same connection | FR1.1 |

Scaling beyond one plugin process per Kandev server is out of scope. Kandev runs one process per installed plugin (C8).

## Sources

- [desc] Initial description: Kandev plugin for Nulab Backlog.
- [scope] Workflow-selected scope: `feature`.
- `functional-spec.md`, `rules.md` (U1); `requirements.md`; `contract-summary.md`.

## Assumptions & Open Questions

None.
