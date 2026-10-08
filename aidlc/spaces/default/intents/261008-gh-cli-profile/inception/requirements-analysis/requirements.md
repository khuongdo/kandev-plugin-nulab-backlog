# Requirements — Per-workspace gh CLI account (261008-gh-cli-profile)

## Intent Analysis

Initial request (verbatim): "Hiện tại khi kết nối với github thông qua gh cli chỉ có thể dùng credential của active user. Cho phép chọn profile theo workspace giống có github integration làm. Khi tạo task worktree từ backlog task đảm bảo gh cli sẽ hoạt động bằng profile đã chọn."

- **Goal**: a person with several GitHub accounts logged in to `gh` can bind each Kandev workspace to one of them, so the plugin's GitHub work for that workspace always runs as that account, whatever account `gh` currently has active.
- **Type**: enhancement of the existing "Use gh CLI login" connection (v0.5.1), brownfield, narrow scope (`internal/scm`, `internal/plugin` SCM actions, `manifest.yaml`, Source control settings UI).
- **Today**: `internal/scm/cli_token.go` runs `gh auth token --hostname github.com` without `--user`, cached per provider only, so every workspace uses the active gh account; a Test after `gh auth switch` silently rewrites the workspace's stored account (codekb `code-quality-assessment.md#intent-findings-261008-gh-cli-profile`).
- **Reference behaviour**: Kandev v0.96.0 `internal/github/gh_accounts.go` lists accounts with `gh auth status --json hosts` and reads a token with `gh auth token --hostname <host> --user <login>`, never switching the active account.
- **Worktree meaning (Q1 = A)**: "gh in a task worktree uses the chosen profile" is met for everything the plugin does for a task created from a Backlog issue. The agent's own shell `GH_TOKEN` is set by Kandev, which the plugin SDK cannot influence; this is documented, not coded.

## Functional Requirements

### FR1 — List gh accounts
- **FR1.1** The system shall provide a browser action (`scm.providers.cli_accounts`, declared in `manifest.yaml`) that returns the GitHub logins logged in to `gh` for host `github.com`, with which one is currently active.
- **FR1.2** When `gh` is missing, not logged in, or fails, the action shall return the existing `cli_unavailable` style error with a plain message and no gh output.
- Acceptance:
  - Given gh has logins `alice` (active) and `bob`, When Settings asks for accounts, Then it receives `alice` (active) and `bob`.
  - Given gh is not installed, When Settings asks for accounts, Then it receives a "gh CLI not available" error.

### FR2 — Choose an account per workspace
- **FR2.1** "Use gh CLI login" for GitHub shall take a chosen login and store it in that workspace's source control settings.
- **FR2.2** The Source control settings card shall show a picker of the accounts from FR1 when the user chooses gh CLI for GitHub, defaulting to the active account, and shall show the chosen login on a connected card.
- **FR2.3** The user shall be able to change the chosen account later from the same card without removing the connection.
- **FR2.4** The choice shall be per workspace: two workspaces may use different gh accounts at the same time.
- Acceptance:
  - Given workspace W1 chose `alice` and W2 chose `bob`, When each lists its PRs, Then W1 calls GitHub with alice's token and W2 with bob's.
  - Given the user picks `bob` while `alice` is active in gh, Then the connection is saved as `bob` and gh's active account is still `alice`.

### FR3 — Use the chosen account for every plugin GitHub call
- **FR3.1** Every GitHub call the plugin makes for a workspace in CLI mode (Test, repo list, PR list/link, "mine" filter, PR watches, tasks created from a Backlog issue) shall use the token of that workspace's chosen login (`gh auth token --hostname github.com --user <login>`).
- **FR3.2** The CLI token cache shall be keyed by provider and login, so one login's token is never served to another.
- **FR3.3** The plugin shall never run `gh auth switch` or change gh's active account.
- **FR3.4** On a gh version without `--user`, the plugin shall use the token only when the chosen login is the active account; otherwise it shall return the FR5 error.
- Acceptance:
  - Given W1 chose `bob` and gh's active account is `alice`, When a task created from a Backlog issue links its PR, Then the GitHub request is authenticated as `bob`.

### FR4 — Existing CLI connections (Q3 = A)
- **FR4.1** A workspace connected with gh CLI before this change, with no chosen login stored, shall treat its saved GitHub login (`AccountID`) as the chosen login, with no user action.
- Acceptance:
  - Given a v0.5.2 workspace saved with `AccountID = alice`, When the plugin is upgraded and `bob` becomes active in gh, Then the workspace still uses `alice`.

### FR5 — Chosen account no longer available (Q4 = A)
- **FR5.1** When gh no longer has the chosen login, the plugin shall fail the call with a clear error naming the login ("<login> is not logged in to gh — log in again or pick another account") and shall not fall back to another account.
- **FR5.2** Test shall never overwrite the stored chosen login with a different gh account; it may refresh the display name of the same login.
- Acceptance:
  - Given W1 chose `bob` and the user runs `gh auth logout --user bob`, When W1 runs Test, Then it shows the "bob is not logged in to gh" error and the stored login stays `bob`.

### FR6 — Agent shell in the task worktree (Q1 = A)
- **FR6.1** The README and the Source control settings card shall state that the agent's shell in a task worktree gets its GitHub credential from Kandev's own GitHub integration (or executor profile), and that it should be set to the same gh account as the plugin's choice for that workspace.

## Non-Functional Requirements

- **NFR1 Security**: tokens and raw gh stdout/stderr are never logged, returned to the browser, or stored; existing redaction and `runCLI` limits (no shell, 10 s timeout, 4 KiB output cap, `GH_TOKEN`/`GITHUB_TOKEN` stripped from the child env) apply to the new `gh auth status` call. Logins are validated (GitHub login charset, length ≤ 39) before being passed as an argument.
- **NFR2 Performance**: listing accounts completes within the existing 10 s CLI timeout; tokens stay cached for the existing 5 min TTL per provider+login.
- **NFR3 Compatibility**: existing token-mode connections, GitLab and Bitbucket behaviour are unchanged; stored settings stay readable by the new version (additive field only).
- **NFR4 Quality**: team posture applies — TDD, Go tests pass with `-race`, 80% coverage floor over `./internal/...` and `./server/...`; UI changes covered by Vitest; existing 1383 Go test results stay green.

## Constraints

- Kandev plugin SDK v0.96.0 (`.kandev-sdk-ref`): no API to set env for a task's agent; `CreateTaskInput` has no env field.
- Only host `github.com` (current CLI support scope).
- Trunk-based, PR to `main`, squash merge (team practices).

## Assumptions

- gh on the user's machine is 2.40+ (supports `gh auth status --json hosts` and `gh auth token --user`); older gh follows FR3.4. [assumption]
- The display name shown on the card may stay the GitHub `name` from `/user`; the login is the identity. [assumption]

## Out of Scope

- GitLab `glab` account choice (Q2 = A).
- Setting the agent shell's `GH_TOKEN` in the worktree, executor-profile selection, and any change to Kandev itself (Q1 = A).
- GitHub Enterprise hosts.

## Open Questions

None.
