# User Stories — 261008-gh-cli-profile

Breakdown: by workflow. Personas: P1 multi-account developer, P2 upgrading gh CLI user (see `personas.md`). Requirements: `../requirements-analysis/requirements.md`. Mob contributions: `contributions/`.

Test homes: **Go-unit** = Go test with a fake CLI runner that records argv and env and returns scripted stdout/exit/timeout; **Go-http** = Go test with an `httptest` GitHub server recording the `Authorization` header; **Vitest** = UI test of the Source control card. Go tests run with `-race`; the clock is injected for TTL checks. Every story can be tested on its own by seeding stored workspace settings; dependencies below are build order, not test chaining.

Global test guard (FR3.3): the fake runner allows only `gh auth status --json hosts --hostname github.com`, `gh auth token --hostname github.com --user <login>` and the old-gh fallback `gh auth token --hostname github.com`; any other argv (including `auth switch`) fails the test.

## US1 — Choose the gh account for a workspace

### US1.1 — Pick an account when connecting with gh CLI
As P1, I want to pick which logged-in gh account a workspace uses when I choose "Use gh CLI login" for GitHub, so that this workspace works as that account no matter which account gh has active.

- Priority: Must
- Size: M
- Requirements: FR1.1, FR1.2, FR2.1, FR2.2, FR3.3, NFR1, NFR2
- UX: the picker is inline on the GitHub card (no dialog), using the host `Select`. "Use gh CLI login" loads accounts first (button disabled, status "Loading gh accounts…"). With two or more accounts a labelled select "GitHub account (gh)" appears with **Connect** and **Cancel**; focus moves to the select. The active account's option reads `alice (active in gh)` (text, not colour only). Failures use `role="alert"`, successes `role="status"`.
- Acceptance criteria:
  - **AC1.1.1** Given gh has `alice` (active) and `bob` logged in for github.com, When I choose "Use gh CLI login" on the GitHub card, Then I see a picker listing `alice (active in gh)` and `bob`, with `alice` preselected. (Go-unit + Vitest)
  - **AC1.1.2** Given I pick `bob` and press Connect, When the connection is saved, Then the card reads "Connected via gh CLI as @bob" (display name after it if different), and the plugin ran no command that changes gh's active account (runner argv log). (Go-unit + Vitest; real gh checked in the manual end-to-end run)
  - **AC1.1.3** Given each of: gh not installed, no github.com login, non-zero exit with undecodable output, timeout, When I choose "Use gh CLI login", Then I see the existing "gh CLI not available on the Kandev server" message as an alert and nothing is saved. (Go-unit table + Vitest)
  - **AC1.1.4** Given gh has exactly one account, When I choose "Use gh CLI login", Then it connects at once with that account, with no picker step. (Vitest)
  - **AC1.1.5** Given a connect or change request whose login is `-x`, `a;b`, empty-but-not-omitted with whitespace, or 40 characters long, When it reaches the plugin, Then it is rejected as invalid input, no gh command runs, and nothing is saved. Logins read from gh output are validated the same way before being returned. Rule: `^[A-Za-z0-9](?:-?[A-Za-z0-9])*$`, length ≤ 39. (Go-unit, runner asserts zero calls)
  - **AC1.1.6** Given `GH_TOKEN` and `GITHUB_TOKEN` are set in the plugin process environment, When the accounts action or a token read runs gh, Then the child process environment contains neither, and the accounts response carries only `login` and `active` per account. (Go-unit: runner records env; response fields allowlisted)
  - **AC1.1.7** Given gh has only `alice`, When a connect request names `bob`, Then it fails with the account-missing error (US3.1) and nothing is saved. (Go-unit)
  - **AC1.1.8** Given gh output lists an account in an error state (e.g. expired keyring token), When the accounts are listed, Then that account is not offered (or shown disabled); an empty usable list is "not logged in" (AC1.1.3). Given `gh auth status` exits non-zero but its stdout decodes, Then the decoded accounts are used. Given the output exceeds the size cap and is cut off, Then the result is `cli_unavailable`, never a partial list. (Go-unit)
  - **AC1.1.9** Given a gh version without `auth status --json`, When I choose "Use gh CLI login", Then the picker offers only the active account (found with the plain token read plus GitHub `/user`). (Go-unit + Vitest)
- Dependencies: none.
- INVEST: independent; largest story (new action, parser, validation, view login, picker); testable at two layers.

### US1.2 — Change the account later
As P1, I want to change a connected workspace to another gh account from the same card, so that I can fix a wrong choice without removing the connection and its repo mappings.

- Priority: Must
- Size: S
- Requirements: FR2.3, FR3.2
- UX: a connected gh CLI card has a **Change account** button next to Test/Remove; it opens the same inline picker with the current login preselected.
- Acceptance criteria:
  - **AC1.2.1** Given workspace W1 is connected as `alice`, When I pick `bob` under Change account and press Connect, Then W1 is connected as `@bob` and its repo mappings and watches are kept. (Go-unit + Vitest)
  - **AC1.2.2** Given W1's cache holds an `alice` token and W1 changes to `bob`, When W1 next calls GitHub, Then the runner is called with `--user bob` and the server sees `bob`'s token. (Go-unit + Go-http)
  - **AC1.2.3** Given W1 is connected as `alice`, When I pick `bob` and `bob` is not logged in to gh, Then the change is refused with the account-missing error and W1 stays connected as `@alice`. (Go-unit)
  - **AC1.2.4** Given W1 is connected as `bob`, When I open Change account and press Cancel, Then nothing is saved and nothing else changes. (Vitest)
- Dependencies: US1.1 (build order).

## US2 — Plugin GitHub work runs as the chosen account

### US2.1 — Every plugin GitHub call for the workspace uses the chosen account
As P1, I want every GitHub call the plugin makes for a workspace — Test, repo list, PR list and "mine" filter, PR links and watches, and work on tasks created from Backlog issues — to use that workspace's chosen account, so that two workspaces can use two accounts at the same time.

- Priority: Must
- Size: S
- Requirements: FR2.4, FR3.1, FR3.2, FR3.3, FR3.4, NFR1, NFR2
- Acceptance criteria:
  - **AC2.1.1** Given W1 chose `alice` and W2 chose `bob`, When both list their PRs concurrently, Then W1's requests are authenticated as `alice` and W2's as `bob`. (Go-http, concurrent under `-race`)
  - **AC2.1.2** Given W1 chose `bob` and gh's active account is `alice`, When a task created from a Backlog issue in W1 links its PR, Then the GitHub request is authenticated as `bob`. (Go-http)
  - **AC2.1.3** Given W1 chose `bob`, When the PR "mine" filter runs, Then it filters by `bob`. (Go-http)
  - **AC2.1.4** Given gh does not support `--user` (the `--user` token read fails while `bob` is in the account list) and `bob` is not the active account, When W1 calls GitHub, Then the plain token read's `/user` login does not match `bob` and the call fails with the account-missing error instead of using `alice`. (Go-unit + Go-http)
  - **AC2.1.5** Given W1 chose `bob` and `alice` is active, When each of Test, repo list, PR list, PR link, the "mine" filter and a background PR-watch poll runs for W1, Then every GitHub request is authenticated as `bob`. (Go-http table, one row per call site)
  - **AC2.1.6** Given gh does not support `--user` and the chosen login is the active account, When W1 calls GitHub, Then the call succeeds with that token. (Go-unit + Go-http)
  - **AC2.1.7** Given W1's `bob` token was fetched, When W1 calls again within 5 minutes, Then gh is not run again; after 5 minutes it is run again; an `alice` cache entry is never served to `bob`. (Go-unit, injected clock)
  - **AC2.1.8** Given the Kandev server process has `GH_TOKEN` set to `alice`'s token, When W1 (chose `bob`) calls GitHub, Then it still uses `bob`'s stored gh login. (Go-unit env assertion + Go-http)
- Dependencies: US1.1 (build order).

## US3 — Keep the account stable

### US3.1 — Clear error when the chosen account is gone
As P1, I want a clear error naming the account when gh no longer has my chosen login, and a way to pick another one right there, so that I re-login or re-pick instead of silently working as someone else.

- Priority: Must
- Size: S
- Requirements: FR5.1, FR5.2, NFR1
- Acceptance criteria:
  - **AC3.1.1** Given W1 chose `bob` and `bob` was logged out of gh, When W1 runs Test, Then the card shows the alert "bob is not logged in to gh on the Kandev server — log in again or pick another account" (built in the UI from the new error code `cli_account_missing` plus the stored login) and the saved login stays `bob`. (Go-unit + Vitest, exact string)
  - **AC3.1.2** Given W1 chose `bob` and the user ran `gh auth switch` to `alice`, When W1 runs Test, Then Test succeeds as `bob` and the stored login is asserted unchanged (`bob`). (Go-unit + Go-http)
  - **AC3.1.3** Given each gh failure row of AC1.1.3, plus a token read failure, plus gh stderr containing a token-like sentinel string, plus output over the size cap, When the error reaches the action response and the logs, Then neither the sentinel nor any raw gh output appears. (Go-unit table, existing redaction-test pattern)
  - **AC3.1.4** Given W1 shows the account-missing error, When I look at the card, Then Change account is available and its picker lists only the accounts gh still has, with nothing preselected that is not logged in. (Vitest)
  - **AC3.1.5** Given the token read fails, When the plugin classifies it, Then: account list also fails → `cli_unavailable`; list works but lacks the login → `cli_account_missing`. (Go-unit)
- Dependencies: US1.1 (build order).

### US3.2 — Upgrade keeps each workspace's account
As P2, I want my existing gh CLI workspaces to keep the account they were connected with after upgrading, so that nothing changes account without me choosing it.

- Priority: Must
- Size: XS
- Requirements: FR4.1, NFR3
- Acceptance criteria:
  - **AC3.2.1** Given a frozen v0.5.2 settings fixture in gh CLI mode with GitHub login `alice` (`AccountID`), When it is loaded by the new version and gh's active account is `bob`, Then the workspace works as `alice`. (Go-unit, fixture in `testdata/`, never regenerated)
  - **AC3.2.2** Given that upgraded workspace, When I open its card, Then it reads "Connected via gh CLI as @alice" and Change account is available. (Vitest)
  - **AC3.2.3** Given the existing token-mode, GitLab and Bitbucket tests, When the change is applied, Then they pass unchanged, and settings stored before the change decode to the same values. A GitLab `use_cli` request with a non-empty login is rejected; without one it behaves as before. (existing suites + Go-unit)
  - **AC3.2.4** Given a gh CLI record with no `AccountID` (not expected from v0.5.2, guarded anyway), When W1 calls GitHub, Then it fails with the account-missing style error and the card asks to pick an account; it never uses the active account silently. (Go-unit + Vitest)
- Dependencies: none.

## US4 — Agent shell in task worktrees

### US4.1 — Know where the agent's gh credential comes from
As P1, I want the settings card and the README to tell me that the agent's shell in a task worktree takes its GitHub credential from Kandev's own GitHub integration (or the executor profile), so that I set it to the same account as the plugin for that workspace.

- Priority: Should
- Size: XS
- Requirements: FR6.1
- Acceptance criteria:
  - **AC4.1.1** Given the GitHub card is in gh CLI mode, When I look at the card, Then I see muted text: "Agents working in task worktrees get their GitHub login from Kandev's own GitHub integration (or the executor profile), not from this plugin. Set it to the same account (@bob) for this workspace." Given token mode, Then the note is absent. Read-only cards show the note and the login but no picker. (Vitest)
  - **AC4.1.2** Given the README's source control section, When I read it, Then it explains the per-workspace gh account and the worktree note. (doc review at code review)
- Dependencies: US1.1.

## Story Dependencies

- US1.1 → US1.2, US2.1, US3.1, US4.1 (build order)
- US3.2 is independent.

## Notes for Code Generation (from the mob)

- `AccountID` already holds the GitHub login and is the "mine" filter key; reuse it as the chosen login (no new stored field, no migration). Add a `login` field to the provider view for the card.
- `Service.credential()` is the single token source; pass the workspace login into the CLI token read there.
- `runCLI` currently does not set `cmd.Env`; stripping `GH_TOKEN`/`GITHUB_TOKEN` is new work (AC1.1.6, AC2.1.8), not existing behaviour as requirement NFR1 implies.
- Test currently overwrites `Account`/`AccountID`; guard with case-insensitive login compare (AC3.1.2).
- Exact gh version that added `auth status --json` must be confirmed during Code Generation (AC1.1.9 fallback covers older gh).
