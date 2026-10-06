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

## Notes

<!-- Anything unexpected. No keys, no URLs with a query string. -->
