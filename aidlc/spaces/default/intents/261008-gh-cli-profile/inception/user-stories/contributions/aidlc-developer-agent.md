**Collaborator:** aidlc-developer-agent

## Contribution

Focus: whether each AC can be built in this codebase, and the technical edge cases the stories miss. Checked against `internal/scm/{cli_token.go,service.go,store.go,prs.go,watcher.go,links.go}`, `internal/github/client.go`, `internal/plugin/scm_actions.go`, `manifest.yaml`, and the Kandev reference `~/repo/kandev/apps/backend/internal/github/gh_accounts.go` (v0.96.0).

### Facts from the code that shape the stories

1. **`AccountID` already is the GitHub login.** `internal/github/client.go` `CurrentUser` returns `User{ID: u.Login, Name: u.Name-or-login}`, and `Settings.AccountID` is the key for the "mine" filter (`prs.go` `byAuthor`, `watcher.go`). For GitHub in CLI mode, the chosen login can simply be `AccountID`; no new stored field is needed. With that choice, US3.2 / FR4.1 needs no migration code, only a regression test, and AC2.1.3 ("mine" filters by `bob`) works with no change.
2. **Every GitHub call goes through one function.** `Service.credential(ctx, ws, p)` is the only token source for Test, repo search, mappings, PR list, links and watches (6 call sites). Passing the workspace's login into `cliToken` there covers all of FR3.1 / US2.1 in one place. US2.1 is therefore small. Most of its cost is tests, not code.
3. **The cache is per provider only** (`map[Provider]cachedToken`). FR3.2 needs a `{provider, login}` key. `forgetCLI(p)` should then drop every login of `p`; this is simple and safe, because Test and UseCLI already force a new CLI read.
4. **`runCLI` does NOT remove `GH_TOKEN` / `GITHUB_TOKEN` from the child process environment.** `cmd.Env` is never set, so gh inherits the Kandev server's environment. NFR1 says this removal already exists; it does not. If the Kandev server process has `GH_TOKEN` set, gh reports that token's account and ignores the stored logins, which breaks US1.1 and US2.1. This is new work and needs a test (see Positions).
5. **`runCLI` discards stderr and caps stdout at 4 KiB.** Two consequences:
   - The plugin cannot tell "gh missing" from "this login is not in gh" from the error text alone.
   - `gh auth status --json hosts` prints about 200–300 bytes per account (it includes `scopes`, `tokenSource` and `gitProtocol`), so roughly 12 or more accounts would be cut off and the JSON would no longer parse. A cut-off result must give `cli_unavailable` and must never return a partial list. Optionally, allow a larger cap (for example 32 KiB) for the status call only.
6. **The Settings view has no login.** `ProviderView.Account` is the display name (`u.Name`). AC1.1.2 / AC3.2.2 ("connected as `bob`") and the FR5 message ("`bob` is not logged in…") need the login. Add a `login` field to `ProviderView` (taken from `AccountID`), or change the AC wording to "display name (login)".
7. **Test overwrites the account today** (`service.go`: `st.Account, st.AccountID = user.Name, user.ID`). FR5.2 / AC3.1.2 need a guard: in CLI mode, if `user.ID` is not the chosen login (`strings.EqualFold`, because GitHub logins ignore case), record the account-missing error and keep `AccountID`. With `--user`, the IDs always match, so the guard protects the old-gh path and the case where the Kandev server has `GH_TOKEN` set.
8. **The `use_cli` input** (`providerBody`) has only `provider` and `query`. It needs a `login` field. For GitLab, `login` must be empty: reject a non-empty value, or ignore it, so GitLab behaviour stays unchanged (AC3.2.3). For GitHub, an empty `login` should mean "use the active account". This keeps the old request shape working, and it also covers the single-account case (AC1.1.4).

### Edge cases the ACs should cover

- **Login validation (NFR1):** Go's `regexp` has no lookahead. Use `^[A-Za-z0-9](?:-?[A-Za-z0-9])*$` with `len <= 39`. Apply it in two places:
  - to the `login` received from the browser, before it is passed as an argument to gh. A login starting with `-` is rejected, which blocks injecting a gh option;
  - to every login read from gh output, before it is returned to the browser.
- **Unknown login on connect:** with UseCLI(`bob`) when gh has no `bob`, `gh auth token --user bob` fails. The result should be the FR5 account-missing error, not a generic `cli_unavailable`, and nothing should be saved (this extends AC1.1.3).
- **Telling "gh missing" from "login missing":** stderr is discarded (fact 5). When the token call fails, list the accounts. If the list call also fails, return `cli_unavailable`. If the list works but `bob` is not in it, return the new code `cli_account_missing`. This is a new error code: map it in `errorCode`, in `scm_actions.go`, and in the UI. The UI builds the message from this code plus the stored login, so the backend never returns free text.
- **Accounts in an error state:** each JSON entry has a `state` (`success`/`error`, for example an expired keyring token). The picker should list only `success` entries for host `github.com`, or show the others as disabled. An empty list means "not logged in" (AC1.1.3).
- **Exit code of `gh auth status`:** it exits non-zero when any account has an error. Decode stdout even when the exit code is non-zero; treat it as a failure only when stdout does not decode.
- **Older gh (FR3.4, AC2.1.4), simplest approach that works:**
  1. Always run `gh auth token --hostname github.com --user <login>` first.
  2. If it fails and the account list does not have the login, return account-missing.
  3. Otherwise run the plain `gh auth token --hostname github.com` and accept the token only after `CurrentUser(tok).ID` matches `<login>`, using `EqualFold`.
  4. Cache the token under that login.

  This needs no `--help` probe and no parsing of the old text output, and it can never serve another account's token. The extra `/user` call happens at most once per 5-minute cache period.
- **Reading the account list on older gh:** `gh auth status --json` is much newer than `gh auth token --user`, which arrived in 2.40 with multi-account support. The `--json` flag was added in a 2.7x release; Functional Design must confirm the exact version. On gh 2.40 up to that release, the JSON call fails. Fallback for the picker: show only the active account, found with the plain token plus `/user`. Do not parse the human-readable text output: it may go to stderr, and stderr is discarded. AC1.1.1 then needs a variant: "Given gh without `auth status --json`, the picker shows only the active account".
- **Never switch the account (FR3.3):** you cannot see gh's active account from inside a unit test. Rephrase AC1.1.2's second clause so it can be tested: "the fake CLI runner never receives `auth switch`, and no call is made without `--hostname github.com`".
- **Lock scope:** listing accounts and reading tokens both run under `cli.mu`. Acceptable: at most 2–3 CLI runs per cache miss, each with a 10 s limit. Listing accounts must not fill the token cache.

### Story sizing

| Story | Size | Notes |
|---|---|---|
| US1.1 | M (largest) | New `scm.providers.cli_accounts` action and manifest entry, an account-list parser, a `login` field in `use_cli`, login validation, `login` in the view, and a UI picker. Could be split into backend and UI, but one story is fine with a single picker. |
| US1.2 | S | Calls `use_cli` again with the new login. `updateProvider` only changes the account fields, so mappings and watches stay. AC1.2.2 holds automatically once the cache is keyed by login. |
| US2.1 | S | One change in `credential()` plus tests that use a fake runner and an httptest server per login. |
| US3.1 | S | New error code, the Test guard (fact 7), and the list-based check. |
| US3.2 | XS | Only a regression test if `AccountID` is reused as the chosen login. |
| US4.1 | XS | Card text and the README. |

Every story is achievable with SDK v0.96.0; none needs a host API change.

## Positions

- AGREE: Breakdown into six stories (Q1 = A) and US4.1 as a Should story (Q2 = A). — Each story maps to a separate code path and test, and none is larger than M.
- AGREE: US2.1 as a single story for all GitHub calls. — `Service.credential()` is the only token source, so one change covers Test, repos, PR list, the "mine" filter, links and watches.
- OBJECT: NFR1 says the `GH_TOKEN`/`GITHUB_TOKEN` removal in `runCLI` already exists. — `runCLI` never sets `cmd.Env`, so gh inherits the server's environment. A `GH_TOKEN` there overrides every stored login and breaks US1.1 and US2.1. Make the removal an explicit AC (for example on US2.1: "Given the Kandev server has `GH_TOKEN` set, When W1 calls GitHub, Then it still uses `bob`'s stored gh login"), with a test for the environment passed to the child process.
- OBJECT: The requirements assume gh 2.40+ supports `gh auth status --json hosts`. — Only `gh auth token --user` dates from 2.40; `--json` on `auth status` is much later. Add an AC for the picker fallback (show only the active account, found through token plus `/user`). Functional Design must confirm the exact gh version.
- OBJECT: AC1.1.2 and AC3.2.2 say "connected as `bob`", but the view only has the display name. — `ProviderView.Account` is `u.Name`. Either add a `login` field to the view, or change the wording to "shows `bob`'s display name and login `bob`". Otherwise the AC cannot be checked against what the card actually shows.
- OBJECT: AC1.1.2's clause "`gh auth status` still shows `alice` as active" cannot be checked in a unit test. — Restate it as "the CLI runner never receives `auth switch`", and keep the real-gh check for the manual end-to-end run.
- AGREE: Keep the chosen login in the existing `AccountID` instead of adding a new field (a suggestion for Functional Design, not a change to the stories). — `AccountID` is already the GitHub login and the "mine" filter key. FR4.1 then holds with no migration, and NFR3 holds because nothing in the stored format changes.
- AGREE: Add to US1.1 / US3.1 that an unknown or invalid login on connect saves nothing and returns the account-missing error. — `gh auth token --user` fails for an unknown login, and AC1.1.3 only covers "gh missing / no login at all".
