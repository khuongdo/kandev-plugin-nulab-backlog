# Manual check: <check name>

> **Never paste an API key, a token, or any URL with a query string (`?...`) into this file.**
> Record the space domain only, for example `myteam.backlog.com`.

Copy this file to `docs/manual-checks/<YYYY-MM-DD>-<check name>.md`
(`walking-skeleton` or `first-release`) and fill in every field.

| Field | Value |
|-------|-------|
| Date | YYYY-MM-DD |
| Check name | walking-skeleton / first-release |
| Kandev version | |
| Plugin commit | |
| Space domain | |
| Result | pass / fail |
| Connect duration (seconds) | |
| Logs checked for the key | yes / no |

## Steps

1. Install `dist/nulab-backlog-<version>.tar.gz` through **Settings > Plugins**.
2. Open **Settings > Integrations > Backlog**.
3. Enter the space domain and an API key, then click **Connect**.
4. Confirm the page shows `Connected as <name> @ <space domain>`.
5. Reload the page and confirm the connection is still shown and the key field is empty.
6. On the Kandev home page, open **Integrations > Backlog** and confirm the `/backlog` page
   shows the Backlog logo and `Connected as <name> @ <space domain>`, and that its settings
   link opens the Backlog settings of the same workspace.
7. On the Backlog card, turn the switch off and click **Save**. Confirm the settings screen
   says Backlog is turned off and shows no Connect form, and the `/backlog` page says Backlog
   is turned off.
8. Turn the switch back on and click **Save**. Confirm `Connected as <name> @ <space domain>`
   is shown again without connecting again.
9. Search the plugin logs for the key and for `apiKey=`; neither may appear. Confirm the log
   has one `integration_switch_changed` event per save.

### Connection (U2) steps

These steps need a registered Nulab OAuth application (see "Sign in with OAuth" in the README).

10. In **Settings > Plugins > Nulab Backlog**, fill in `oauth_client_id`, `oauth_client_secret` and
    `public_base_url`. In the Backlog settings, choose **OAuth**, enter the space domain and click
    **Sign in with Nulab**. Grant access on the Backlog page and confirm the page returns with
    `Connected as <name> @ <space domain>`. Repeat and click cancel on the Backlog page; confirm
    `Sign-in was cancelled` and that the earlier connection is unchanged.
11. Token refresh past expiry: leave the OAuth connection for more than one hour (or until the token
    has expired), then click **Test connection**. Confirm it succeeds and the log has one
    `token_refreshed` event and no `token_refresh_failed`.
12. Select one project, click **Disconnect** and confirm the dialog says everyone in the workspace
    loses the connection, with the focus on **Cancel**. Confirm, then connect to the same space
    again and confirm the notice `Reconnected to <space domain>. Items from your earlier connection
    to this space were restored.` and that the project is selected again.
13. Rate-limit wait: point the plugin at a fake Backlog that answers `429` with `Retry-After: 2`
    (or run `go test -race ./internal/backlog/... -run '^TestQueueWaitsAreLogged' -v`) and confirm
    one `backlog_wait` log line with `group`, `reason`, `waitMs` and `attempt`, and no query string.
14. Search the plugin logs for the client secret, the access token and `code=`; none may appear.

| U2 field | Value |
|----------|-------|
| OAuth sign-in | pass / fail |
| Refresh after expiry | pass / fail |
| Disconnect and restore | pass / fail |
| Rate-limit wait log line | pass / fail |

### Git and pull requests (U4, B5 demo) steps

These steps need pull requests and Git hosting on the space's plan (A1) and a test repository
in a selected project. Never write the Git password into this file.

15. In the Backlog settings, open **Git access (optional)**, enter the Git user name and password
    and click **Save Git access**. Confirm **Saved** and **Stored**, and that the password field is
    empty. Click **Test connection** and confirm no Git error is shown.
16. Create a Kandev task on the Backlog repository (repository source **Backlog**). Let the agent
    change a file and push. Confirm the clone and the push worked without asking for a password
    (AC5.6.2), and that the plugin log has no `git_credential_refused` line.
17. In the task, click **Create PR**. Confirm a pull request appears in Backlog for the task's
    branch, that the task card shows its badge (for example `Open – <assignee>`), and that a
    second **Create PR** asks whether to link the open pull request instead.
18. In **Integrations > PR watches**, create a watch for the repository with status Open. Open a
    new pull request in Backlog, click **Run** (or wait 5 minutes) and confirm exactly one task
    `Review PR #<n>: <title>` is created and linked; run again and confirm no second task.
19. Click **Disconnect** and confirm the dialog counts the PR links and watches that turn off, and
    that the watch then shows **Not connected**. Connect to the same space again and confirm the
    notice says `PR watches are Paused.`, the PR badge is back, and the watch shows **Paused**.
20. Search the plugin logs for the Git password and the API key; neither may appear.

| U4 field | Value |
|----------|-------|
| Git access saved and tested | pass / fail |
| Clone and push with Git access | pass / fail |
| Create PR and badge | pass / fail |
| Watch creates one task | pass / fail |
| Disconnect and restore (watches Paused) | pass / fail |

### Backlog issues (U3, B4 demo) steps

These steps need at least one selected project with issues. Never write the API key into this file.

21. Open **Integrations > Backlog**, search `PROJ-12` (a key of your project), and click **Create task**
    in the issue's **…** menu. Confirm `Created <task key>` and that the row lists the task.
22. Open the Kanban board and confirm the task card shows the badge `<KEY> · <status>`.
23. Change the issue's status in Backlog. Wait one sync cycle (5 minutes, or click **Refresh** on
    the Issues page) and confirm the badge shows the new status.
24. Delete the task in Kandev and confirm the link disappears from the issue's row.
25. Load the Issues page with 20 rows 20 times and record the slowest time (AC8.1.2: at most 3 s).
26. Use the Issues page, the Link to task dialog and the Backlog issue panel with the keyboard only
    and with a screen reader, then at 320 px width (AC8.2.1, AC8.2.4). Every control must be
    reachable, named, and usable without horizontal page scroll.

| U3 field | Value |
|----------|-------|
| Create task from an issue | pass / fail |
| Badge on the card | pass / fail |
| Status change within one cycle | pass / fail |
| Deleted task removes the link | pass / fail |
| Slowest 20-row list (seconds, of 20) | |
| Keyboard, screen reader and 320 px | pass / fail |

## Notes

<!-- Anything unexpected. No keys, no URLs with a query string. -->
