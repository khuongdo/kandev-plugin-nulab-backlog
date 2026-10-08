# Requirements — CLI login as a second auth method for GitHub and GitLab

## Intent Analysis

Initial description: "Hiện tại link với github source control chỉ có qua apikey. Thêm 1 phương thức nữa là bằng gh cli auth" [desc]

Today an admin connects GitHub (and GitLab, Bitbucket) in Settings > Source Control only by pasting a personal access token. On a self-hosted Kandev server where the admin has already run `gh auth login` (or `glab auth login`), pasting a token is extra work and the token goes stale. The goal is to let the plugin reuse the CLI login that already exists on the Kandev server, so the connection follows the CLI's own login, refresh and account switches. [desc] [Q1] [Q2]

- Type: enhancement of an existing feature (source-control connection).
- Scope: multi-component inside one repo — `internal/scm` (credential resolution, settings, errors), `internal/github`, `internal/gitlab`, `internal/plugin` (actions, error mapping, manifest), `ui/src/settings/source-control-section.tsx`.
- Complexity: standard; depth Minimal (express). [scope]

Code facts this builds on (CodeKB `kandev-plugin-nulab-backlog`):
- Every provider call gets its credential from one function, `scm.Service.credential` (7 callers, including the background PR watcher). Clients only add the token to a header.
- The plugin is a process started by the Kandev server and inherits its environment (PATH, HOME, `GH_TOKEN`, `GH_CONFIG_DIR`), so the plugin can run `gh` / `glab` itself. The Kandev host API offers no way to borrow Kandev's own GitHub credential.
- Provider hosts are fixed: `api.github.com` and `gitlab.com/api/v4`.

## Functional Requirements

### FR1 — Connect GitHub with the gh CLI login [Q4] [Q2]
- FR1.1 The GitHub card in Settings > Source Control shall keep the existing token field and add a "Use gh CLI login" button next to it.
- FR1.2 When an admin presses the button, the plugin shall run `gh auth token --hostname github.com` on the Kandev server, verify the token by reading the current GitHub user, and on success save the connection with source `cli` and the GitHub account name. The token itself is not saved.
- FR1.3 Connecting with the gh CLI shall replace a saved typed token (the saved token secret is deleted). Saving a typed token later shall replace the CLI source. Only one method is active at a time.
- FR1.4 Only admins can connect, switch or disconnect (same guard as the existing token actions).

### FR2 — Connect GitLab with the glab CLI login [Q2]
- FR2.1 The GitLab card shall get a "Use glab CLI login" button next to the token field, with the same behaviour as FR1.2–FR1.4, using `glab` for host `gitlab.com`.
- FR2.2 Requests made with a glab-sourced token shall authenticate in a way that works for both personal access tokens and OAuth tokens from `glab auth login` (e.g. `Authorization: Bearer`).
- FR2.3 Bitbucket stays token-only; its card does not show a CLI button.

### FR3 — Live token lookup [Q1]
- FR3.1 While a provider's source is `cli`, every credential lookup shall get the token from the CLI instead of the secret store.
- FR3.2 The plugin shall cache the CLI token in memory for at most 5 minutes per provider, so the 1-minute PR watcher does not start a CLI process on every run. The cache is cleared when the provider is disconnected or switched to a token, and when a request with the cached token returns 401.
- FR3.3 After `gh auth login`, `gh auth refresh` or `gh auth switch` on the server, the plugin shall use the new token no later than the cache expiry, without any action in Kandev.

### FR4 — Clear failure when the CLI is not usable [Q3]
- FR4.1 If the CLI is not installed, not logged in, returns an empty token, times out, or the token is rejected (401), the action or watcher run shall fail with a dedicated error, e.g. "gh CLI is not available or not logged in on the Kandev server" (glab respectively). There is no fallback to any other credential.
- FR4.2 The provider card shall show the error state and that message, as it does today for a failed token test.
- FR4.3 "Test connection" shall work for the CLI source and clear the error once the CLI works again.

### FR5 — Show the active method [Q4]
- FR5.1 The provider list shown to the UI shall include the active method (`token` or `cli`) and the account name. It shall never include a token.
- FR5.2 The card shall show which method is active (e.g. "Connected via gh CLI as <account>").
- FR5.3 Disconnect ("Remove") shall clear the CLI source the same way it clears a typed token, and keep repository mappings as today.

### FR6 — Compatibility
- FR6.1 Existing connections saved by v0.5.0 (typed token, no method field) shall keep working and be shown as method `token` without any migration step.
- FR6.2 The new admin actions shall be declared in `manifest.yaml` (the manifest/runtime parity test must stay green).

## Non-Functional Requirements

- NFR1 Secrets: the CLI token shall never be written to plugin state, the secret store, logs, error messages or responses to the UI. CLI stdout is trimmed and registered for redaction; CLI stderr is never surfaced. A test asserts no leak (team rule: redact API keys and tokens).
- NFR2 Bounded execution: each CLI call shall run with a fixed argument list (no shell, no user input in arguments) and a timeout of at most 10 seconds; output read is size-limited.
- NFR3 Testability: the command runner and clock shall be injectable so tests never run a real `gh`/`glab` and never sleep. TDD, `go test -race`, 80% coverage floor on `./internal/...` and `./server/...` stay in force.
- NFR4 Load: with the cache, a provider using the CLI source starts at most one CLI process per 5 minutes in steady state (plus one per 401).
- NFR5 Existing tests stay green; the packaged-host contract test on the minimum Kandev version still passes.

## Constraints

- The CLI runs on the machine running the Kandev server, under the server's user and environment. It does not use the browser user's CLI login. If Kandev runs in Docker or the CLI is missing there, only the token method works.
- No new Go or npm dependency; use `os/exec` from the standard library (team rule: prefer the standard library).
- Fixed hosts only (`github.com`, `gitlab.com`); no GitHub Enterprise or self-managed GitLab.
- Only `internal/plugin` imports `pluginsdk`.

## Assumptions

- [assumption] `gh auth token --hostname github.com` prints the token on stdout (gh ≥ 2.17). For `glab`, the exact token command depends on the installed version (`glab auth token` where available, otherwise `glab config get token --host gitlab.com`); Code Generation picks the command and documents the minimum glab version.
- [assumption] Q1, Q3 and Q4 were asked about GitHub; they apply to GitLab in the same way (see the interpretation notes in the questions file).
- [assumption] The 5-minute cache value comes from the Q1 option text and is acceptable.

## Out of Scope

- CLI auth for Bitbucket.
- Fallback from the CLI to a typed token.
- GitHub Enterprise / self-managed GitLab hosts.
- Using Kandev's own GitHub integration credential.
- Backlog connection auth (API key / OAuth) — unchanged.

## Open Questions

- Which `glab` command and minimum version to support (resolved in Code Generation, see Assumptions).
