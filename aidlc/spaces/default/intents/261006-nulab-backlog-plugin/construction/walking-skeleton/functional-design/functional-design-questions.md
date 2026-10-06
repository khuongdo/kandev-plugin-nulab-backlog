# Functional Design — walking-skeleton (U1) — Clarifying Questions

## Sources

- [desc] Initial description: "feature dựa trên plugin này https://github.com/kdlbs/kandev-plugin-bitbucket hãy tạo 1 plugin tương tự cho kandev (https://github.com/kdlbs/kandev) nhưng dùng cho nulab backlog dựa trên public api ở đây https://developer.nulab.com/docs/backlog/"
- [scope] Workflow-selected scope: `feature`.

U1 covers US7.1, US1.2, US1.1 and US7.2 (`unit-of-work-story-map`). Most behaviour is already fixed by the acceptance criteria, `contract-summary` (C1, C4, C5, C8) and `team-practices`. The questions below cover only the gaps left for this unit.

---

## Q1. How the API key reaches Backlog

The Backlog API documentation authenticates API keys with the `apiKey` URL query parameter. AC1.1.7 and NFR3 say the key goes in a header, not the URL. This is the open contract question that blocks U1 (`contract-summary`, C7).

- A. Use the documented `apiKey` query parameter. Redact the full URL in every log line, error and test output (`internal/redact`), and never put the URL in an error returned to the UI. Amend AC1.1.7 to "the key never appears in logs, errors or UI responses"
- B. Try a header first. Confirm during the first manual check against a real space; if Backlog rejects it, fall back to option A
- C. Not yet defined
- X. Other (please specify)

[Answer]: A

## Q2. Connecting when a connection already exists

In U1 there is no Disconnect or Replace yet; those arrive in U2 (US1.5, US1.6). What should `connection.connectApiKey` do if the workspace is already connected?

- A. Replace the existing connection, but only after the new key is verified. Keep the old one if verification fails. U2 adds the confirmation dialog
- B. Refuse with `conflict` until U2 adds Disconnect
- C. Not yet defined
- X. Other (please specify)

[Answer]: A

## Q3. Two Connect requests at the same time

AC1.1.5 says that clicking Connect many times must not create two connections. The button lock in the UI covers one browser, but not two admins clicking at the same moment.

- A. The backend also serializes connect per workspace. A second request that arrives while one is running gets `conflict`
- B. Rely on the UI lock only
- C. Not yet defined
- X. Other (please specify)

[Answer]: A

## Q4. Where the manual check record lives

AC7.2.1 requires a record of the first manual check: the date, Kandev version, plugin commit, space domain (no key), the steps and the result.

- A. A Markdown file in the repo at `docs/manual-checks/<YYYY-MM-DD>-walking-skeleton.md`, created from a template the unit adds
- B. Attached to the pull request description only
- C. Not yet defined
- X. Other (please specify)

[Answer]: A

## Q5. UI languages in U1

NFR10 says the UI follows Kandev's language, but the requirements leave open which languages ship translations. U1 adds only the reduced settings screen (M1).

- A. English only in U1, with every string behind a message key so U3 can add translations without code changes
- B. English and Vietnamese from U1
- C. English and Japanese from U1 (Backlog is a Japanese product)
- X. Other (please specify)

[Answer]: A

## Requested Changes Feedback

Verbatim feedback from the user after the first manual check (2026-10-06):

> - thêm option Backlog trong menu INtegrations
> - thêm toggle button trong card trong setting integrations
> - thay icon hiện tại bằng icon của backlog (nếu có)
> - retouch UI trong setting một chút, nó hơi sát nhau

Context: an earlier decision placed the per-workspace switch in U1, default on; when off, the Backlog entry stays so it can be turned back on, but Connect and every Backlog feature for that workspace are blocked. Kandev 0.96.0 offers `registerNavItem({ section: "integrations" })` for the menu entry, an `icon` field (curated name or plugin-owned component) on both the nav item and the integration settings card, an `action` component slot on the settings card, and `host.ui.IntegrationEnabledControl` plus `host.setIntegrationEnabled` for the switch.

## Q6. Where the Backlog entry in the Integrations menu leads

A. To the Backlog settings page for the current workspace (the same page as the settings card)
B. To a new plugin page showing the connection status and a link to settings
X. Other (please specify)

Clarification from the user: "menu ở đây là menu ở trang home, chứ không phải menu trong setting" (the menu is the Integrations menu on the Kandev home page, not the one in Settings). See Q6a.

[Answer]: X (clarified, see Q6a)

## Q7. Where the on/off value lives and who enforces it

A. The backend stores the value per workspace in plugin state and refuses Connect and every Backlog action while it is off; only admins can change it, like Connect
B. The UI stores the value with `host.storage`; only the UI hides or blocks features
X. Other (please specify)

[Answer]: A

## Q8. Icon

A. The official Backlog logo as a plugin-owned SVG component, taken from Nulab's official brand assets
B. A simple plugin-drawn glyph that does not copy the Backlog logo
C. A host built-in icon (for example `ticket`)
X. Other (please specify)

[Answer]: A

## Q9. Settings layout retouch

A. Spacing only: consistent gaps between fields, labels, buttons and messages, using the host UI kit
B. Spacing plus clear sections (connection status, connect form) with headings
X. Other (please specify)

[Answer]: A

## Q6a. What the home Integrations entry opens

On the Kandev home page, first-party entries in the Integrations menu (Jira, GitHub, Linear) open their own page (for example `/jira`). The Backlog entry is registered with `registerNavItem({ section: "integrations" })` and needs a target page.

A. A Backlog plugin page in the same style: in U1 it shows the connection status for the current workspace, whether Backlog is on or off, and a link to the Backlog settings; U3 later adds the issue list to this page
B. Go straight to the Backlog settings page of the current workspace
X. Other (please specify)

[Answer]: A

## Consolidated Summary Confirmation

Summary of the answers:

- Q1: send the API key with Backlog's documented `apiKey` query parameter; redact the full URL in every log, error and test output; AC1.1.7 is read as "the key never appears in logs, errors or UI responses" (A).
- Q2: connecting while already connected replaces the connection only after the new key is verified; a failed verification keeps the old one (A).
- Q3: the backend allows one Connect per workspace at a time; a concurrent second request gets `conflict` (A).
- Q4: the first manual check is recorded in `docs/manual-checks/<YYYY-MM-DD>-walking-skeleton.md`, created from a template added by this unit (A).
- Q5: English only in U1, with every UI string behind a message key (A).
- Q6/Q6a: a Backlog entry appears in the Integrations menu on the Kandev home page (`registerNavItem`, section `integrations`). It opens a new plugin page at `/backlog` that, in U1, shows the current workspace's connection status, whether Backlog is on or off, and a link to the Backlog settings; U3 later adds the issue list to this page (A).
- Q7: a per-workspace on/off switch on the Backlog card in Settings > Integrations (the card's action slot, using the host's integration switch). The backend stores the value per workspace and refuses Connect and every Backlog action while it is off; it defaults to on; only admins can change it, like Connect; when off, the Backlog menu entry, page and settings card stay visible so it can be turned back on, and the page and settings show that Backlog is off for this workspace (A).
- Q8: the official Backlog logo, as a plugin-owned SVG component taken from Nulab's official brand assets, replaces the current icon on the menu entry, the page title and the settings card; the user checks that Nulab's brand guidelines allow this use (A).
- Q9: settings layout retouch is spacing only: consistent gaps between fields, labels, buttons and messages, using the host UI kit (A).

Does this all look correct before I generate the artifact?

- Looks correct
- Request changes

[Answer]: Looks correct
