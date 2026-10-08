**Collaborator:** aidlc-quality-agent

## Contribution

Focus: whether each acceptance criterion has a clear pass/fail, covers happy and error paths, and has an obvious test home. Test homes used below: **Go-unit** = Go unit test with a fake CLI runner that records argv and env and returns scripted stdout/stderr/exit/timeout; **Go-http** = Go test with an `httptest` GitHub server that records the `Authorization` header per request; **Vitest** = UI test of the Source control settings card. All Go tests run with `-race`; the clock is injected for TTL checks (team Testing Posture).

### 1. Test home per existing AC

| AC | Pass/fail is clear? | Test home | Note |
|----|--------------------|-----------|------|
| AC1.1.1 | Yes | Go-unit (action returns `alice` active + `bob`) + Vitest (picker lists both, `alice` preselected) | Two layers; both needed. |
| AC1.1.2 | Partly | Go-unit + Vitest | "`gh auth status` still shows `alice`" cannot be asserted in a test. Reword to: "and the plugin ran no command that changes gh's active account". Assert this with the fake runner's argv log. |
| AC1.1.3 | Partly | Go-unit table + Vitest | Combines two cases and leaves out FR1.2's "fails" (non-zero exit, timeout). Make it a table: not installed, no github.com login, non-zero exit, timeout. Each row returns `cli_unavailable` and saves nothing. |
| AC1.1.4 | Yes | Vitest | |
| AC1.2.1 | Yes | Go-unit (service: login changed, mappings and watches unchanged) + Vitest | |
| AC1.2.2 | Yes | Go-unit + Go-http | Seed the cache with `alice`, switch to `bob`, then assert the runner was called with `--user bob` and the server saw `bob`'s token. |
| AC2.1.1 | Yes | Go-http | Run both workspaces concurrently under `-race`. A sequential test does not prove FR3.2. |
| AC2.1.2 | Yes | Go-http | |
| AC2.1.3 | Yes | Go-http | Assert the query or filter value is `bob`. |
| AC2.1.4 | Partly | Go-unit | "Gh does not support `--user`" needs a defined detection signal (for example the stderr pattern `unknown flag: --user`). Functional Design must define it so the fake runner can reproduce it. Positive case missing, see §2. |
| AC3.1.1 | Yes | Go-unit | Exact string compare, including the em dash. |
| AC3.1.2 | Yes | Go-unit + Go-http | |
| AC3.1.3 | No | Go-unit table | "Any gh failure" has no bound. Use the same rows as AC1.1.3, plus stderr that contains a sentinel token-like string and output over the 4 KiB cap. Assert that the error, the logs and the action response contain neither the sentinel nor the raw stderr (reuse the existing redaction-test pattern). |
| AC3.2.1 | Yes | Go-unit | Use a frozen v0.5.2 settings JSON fixture in `testdata/` and do not regenerate it. |
| AC3.2.2 | Yes | Vitest | |
| AC3.2.3 | No | Existing suites | "Behaviour is unchanged" has no pass/fail. Reword to: "the existing token-mode, GitLab and Bitbucket tests pass unchanged, and settings stored without the new field decode to the same values". |
| AC4.1.1 | Yes | Vitest | Also assert the note is **absent** in token mode. |
| AC4.1.2 | Manual | Doc review at code review | Acceptable for a Should doc item. |

### 2. Missing ACs (proposed, BDD form)

- **US1.1 / NFR1, login validation (trust boundary)**: AC1.1.5: Given a connect or change request with login `-x`, `a;b`, an empty string, or 40 characters, When it reaches the plugin, Then it is rejected as invalid input, no gh command runs, and nothing is saved. (Go-unit; the fake runner asserts zero calls.)
- **US1.1 / NFR1, accounts call hardening**: AC1.1.6: Given `GH_TOKEN` and `GITHUB_TOKEN` are set in the plugin's environment, When the accounts action runs, Then the gh child process runs without them and the response carries only `login` and `active` per account, with no other fields from `gh auth status` output. (Go-unit: the fake runner records the env; the test allowlists the response fields.)
- **US1.1 / FR5.1, connect with a login gh does not have**: AC1.1.7: Given gh has only `alice`, When a connect request names `bob`, Then it fails with the FR5 error and nothing is saved. Without this AC, a stale picker or a crafted call saves an unusable connection.
- **US1.2, error path**: AC1.2.3: Given W1 is connected as `alice`, When I pick `bob` and `bob` is not logged in to gh, Then the change is refused with the FR5 error and W1 stays connected as `alice`. US1.2 currently has only happy paths.
- **US2.1 / FR3.1, every call site**: AC2.1.5: Given W1 chose `bob` and `alice` is active, When each of Test, repo list, PR list, PR link, the "mine" filter, and a PR-watch poll runs for W1, Then every GitHub request is authenticated as `bob`. (Go-http table, one row per call site.) Repo list and PR watches are not covered today. The PR-watch poll runs in the background with no browser request, so it resolves the workspace in a different way, which makes it the call site most likely to use the wrong account.
- **US2.1 / FR3.4, old gh positive case**: AC2.1.6: Given gh does not support `--user` and the chosen login is the active account, When W1 calls GitHub, Then the call succeeds with that token.
- **US2.1 / FR3.3, global guard**: The fake runner should allow only `auth status --json hosts` and `auth token --hostname github.com --user <login>` (plus the old-gh fallback form) and fail the test on any other argv. One check then covers "never `gh auth switch`" for every test.
- **US2.1 / NFR2, cache TTL per login**: AC2.1.7: Given W1's `bob` token was fetched, When W1 calls again within 5 min, Then gh is not run again, and after 5 min it is run again. (Go-unit with an injected clock, no real sleep.)
- **US3.2, legacy record with no AccountID**: The story does not say what happens when a v0.5.2 gh CLI record has neither a chosen login nor an `AccountID`. Propose AC3.2.4: Then the card asks the user to pick an account, and GitHub calls fail with the FR5-style error and do not use the active account. If this case cannot occur, record that as an assumption instead.

### 3. Traceability

- Each traceability.json row targets a story ID, never an AC ID. Story IDs are valid under the schema. If the coverage targets name the AC that proves each item, Build and Test can map tests to requirements directly. Examples: FR1.2 → AC1.1.3, FR3.3 → AC1.1.2 plus the runner allowlist, NFR2 → AC2.1.7.
- **NFR1** is mapped to US1.1, but no current AC in US1.1 tests it. Login validation and env stripping have no AC (§2), so coverage is `OK` on paper only.
- **NFR2** is mapped to US1.1, but no AC tests the timeout or the TTL. Add AC2.1.7, or mark the timeout part `Deferred` to code-generation the way NFR4 is.
- **FR2.4** is mapped to US1.1, but only US2.1 (AC2.1.1) tests it. Drop US1.1 from that target.
- **FR3.1** is `OK`, but only part of it is covered: repo list and PR watches have no AC (§2).
- **FR5.2**: "may refresh the display name of the same login" has no AC. That is acceptable because it is optional, but the overwrite guard (AC3.1.2) should assert that the login is unchanged, not just that Test succeeds.
- **NFR4** `Deferred` to code-generation is correct.
- Independence (Inception rule): US1.2, US2.1 and US3.1 depend on US1.1 only for setup. Each one can be tested on its own by seeding stored workspace settings. State this in the stories so the tests do not chain.

## Positions

- AGREE: Breakdown by workflow (Q1 = A), about 6 stories — each story maps to one test cluster, so tests stay independent.
- AGREE: US4.1 as a Should story (Q2 = A) — the card note can be checked with Vitest, and the README is checked at code review.
- AGREE: NFR4 Deferred to code-generation — the TDD, `-race` and coverage floor are pipeline gates, not story behaviour.
- OBJECT: No AC covers login validation or hardening of the accounts call (NFR1) — this is a trust boundary where user input becomes a process argument; add AC1.1.5 and AC1.1.6 before approval.
- OBJECT: FR3.1 is marked OK, but repo list and PR-watch polling have no AC — the background poll is the call site most likely to use the wrong account; add the per-call-site table AC2.1.5.
- OBJECT: US1.2 has no error path, and connecting with a login gh does not have is unspecified — add AC1.1.7 and AC1.2.3 so the FR5 behaviour applies when the account is saved, not only when it is used.
- OBJECT: AC3.1.3 ("any gh failure") and AC3.2.3 ("behaviour is unchanged") have no pass/fail — replace them with the bounded table and the reworded criterion in §1.
