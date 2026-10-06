# Contract Design — Clarifying Questions

## Sources

- [desc] Initial description: "feature dựa trên plugin này https://github.com/kdlbs/kandev-plugin-bitbucket hãy tạo 1 plugin tương tự cho kandev (https://github.com/kdlbs/kandev) nhưng dùng cho nulab backlog dựa trên public api ở đây https://developer.nulab.com/docs/backlog/"
- [scope] Workflow-selected scope: `feature`.

This step settles the "contracts" at the boundaries. A contract is an agreement about which data crosses a boundary, in what shape, and what happens when there is an error.

There are three kinds of boundary:

- **Between units**: all 5 units run in the same Go process. So they connect through Go interfaces and in-process events, not over the network.
- **Between the UI and the backend**: the UI calls the plugin's "actions" through Kandev, the same way the Bitbucket plugin declares actions in its manifest.
- **With the outside**: the Backlog API (the plugin is the caller), the Kandev server, and the OAuth callback address that the browser returns to.

---

## Q1. Style of operations between the UI and the backend

How does the UI call the backend?

- A. Each operation is its own action with a clear name (for example `issues.list`, `issues.createTask`, `connection.connect`). Request and response are JSON with fixed types, like the Bitbucket plugin
- B. One single shared action that branches internally by command name
- C. Not yet defined
- X. Other (please specify)

[Answer]: A A

## Q2. Changing contract versions

The UI and the backend always sit in the same package, so their versions always match. How do we handle a contract change?

- A. No separate version mechanism between the UI and the backend. Just follow the package's semver: a breaking change bumps the major version. The plugin's stored data has a version number and is migrated automatically on startup
- B. Each action has its own version number (for example `issues.list.v1`), and the old version is kept when a new one is added
- C. Not yet defined
- X. Other (please specify)

[Answer]: A A

## Q3. Error shape returned to the UI

In what shape does the backend return errors to the UI?

- A. One fixed set of error codes shared by all actions:
  - `reconnect_required`
  - `rate_limited` (with the number of seconds to wait)
  - `unreachable`
  - `not_found`
  - `validation` (with the name of the invalid field)
  - `conflict`
  - `internal`

  The UI uses the code to pick the message to show in the current language. Errors never contain Backlog response content or secrets
- B. Each action defines its own error messages
- C. Not yet defined
- X. Other (please specify)

[Answer]: A A

## Q4. Time limit and retries when calling Backlog

The story step left two values open: the maximum wait time for each call (AC8.1.3) and the number of retries when rate limited (AC8.4.2). Which values do you choose?

- A. Each call takes at most 10 seconds; when rate limited, retry at most 3 times, then report a `rate_limited` error
- B. Each call takes at most 30 seconds; retry at most 5 times
- C. Not yet defined
- X. Other (please specify)

[Answer]: A A

## Q5. Operations that wait long because of rate limiting

A command the user clicks (for example, searching issues) may have to wait up to 60 seconds if Backlog is rate limiting, so it can exceed the 3-second target. What should the plugin do?

- A. A user-clicked command does not wait more than 3 seconds because of rate limiting. Beyond 3 seconds it returns `rate_limited` right away with the number of seconds to wait, and the UI retries automatically. Only background tasks wait the full time
- B. All commands wait the same way, and the UI shows a countdown
- C. Not yet defined
- X. Other (please specify)

[Answer]: A

## Q6. Detecting that a Kandev task was deleted

At the component design step, one assumption is still open: the plugin needs to know when a task is deleted. The reasons are to remove dangling issue links (US2.3), and so that PR watch does not re-create a task the user deleted. The Kandev source code shows the plugin can receive the `task.deleted` event, and can also ask whether a task still exists with `GetTask`. Which approach should the plugin use?

- A. Subscribe to the `task.deleted` event to remove links right away. In addition, every sync cycle re-checks the linked tasks with `GetTask`, to catch missed events (for example while the plugin was off)
- B. Only use the `task.deleted` event
- C. Only check with `GetTask` in every sync cycle
- D. Not yet defined
- X. Other (please specify)

[Answer]: A

## Consolidated Summary Confirmation

Summary of the answers:

- Q1: each operation is its own action with a clear name; request and response are JSON with fixed types (A).
- Q2: no separate version per action; follow the package's semver; stored data has a version number and is migrated automatically on startup (A).
- Q3: one shared error code set for all actions (`reconnect_required`, `rate_limited`, `unreachable`, `not_found`, `validation`, `conflict`, `internal`), with no Backlog response content or secrets (A).
- Q4: each Backlog call takes at most 10 seconds; when rate limited, retry at most 3 times, then report `rate_limited` (A).
- Q5: a user-clicked command does not wait more than 3 seconds because of rate limiting; beyond that it returns `rate_limited` with the number of seconds to wait; only background tasks wait the full time (A).
- Q6: receive the `task.deleted` event to remove links right away, and every sync cycle re-checks with `GetTask` (A).

Does this all look correct before I generate the artifact?

- Looks correct
- Request changes

[Answer]: Looks correct
