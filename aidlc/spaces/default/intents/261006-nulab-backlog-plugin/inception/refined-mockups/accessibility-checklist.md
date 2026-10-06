# Accessibility Checklist — Kandev Plugin for Nulab Backlog

Inputs:

- Baseline WCAG 2.1 AA target, settled at the rough-mockups step (Q7). Source documents: `wireframes`, `user-flow`.
- NFR9 in `requirements`.
- US8.2 and US5.4 in `stories`.
- Screens M1–M12 in `mockups.md`.

Classification per `team-practices`:

- **[auto]**: checked by an automated scanning tool in the UI tests.
- **[manual]**: checked by hand, with the result recorded.

## By Criterion

| # | WCAG criterion | Specific requirement | Screens | How checked |
|---|---------------|----------------|----------|-----------|
| 1 | 1.1.1 Non-text Content | Icons have text labels; decorative icons have `aria-hidden` | All | [auto] |
| 2 | 1.3.1 Info and Relationships | Headings at the right level (h1 for the page, h2 for dialogs and blocks); tables and lists have real structure; checkbox groups sit inside a `fieldset` | M1–M8 | [auto] |
| 3 | 1.4.1 Use of Color | Every status has accompanying text (Open, Paused, not connected, may be out of date) | M2, M4, M6, M8, M12 | [manual] |
| 4 | 1.4.3 Contrast | Text contrast at least 4.5:1; labels and borders at least 3:1 | All | [auto] |
| 5 | 1.4.10 Reflow | At 320px width no horizontal scrolling is needed; lists switch to cards | M2m, M4, M5, M6 | [manual] |
| 6 | 2.1.1 Keyboard | Every action can be done with Tab, Enter, Space, Esc and the arrow keys | All | [manual] |
| 7 | 2.1.2 No Keyboard Trap | Esc can always close a dialog; focus never gets stuck | M3, M4, M5, M7, ConfirmDialog | [manual] |
| 8 | 2.4.3 Focus Order | Tab order follows reading order; closing a dialog returns focus to the button that opened it | All | [manual] |
| 9 | 2.4.7 Focus Visible | Clear focus ring, contrast at least 3:1 | All | [manual] |
| 10 | 2.5.5 (recommended) Target Size | On phones, buttons have a touch target of at least 44x44px | M2m, M6 | [manual] |
| 11 | 3.3.1 Error Identification | Errors are written in text, placed right under the field and linked with `aria-describedby` | M1, M4, M5, M7, M11 | [auto] |
| 12 | 3.3.2 Labels or Instructions | Every input has a visible label; required fields are marked | M1, M4, M5, M7, M11 | [auto] |
| 13 | 4.1.2 Name, Role, Value | Dialogs use `dialog`/`alertdialog`; expand buttons have `aria-expanded`; disabled menu items use `aria-disabled` so they still receive focus | All | [auto] |
| 14 | 4.1.3 Status Messages | Connection results, load complete, errors, rate limiting and restore are all announced via `aria-live="polite"`, each event announced only once | M1, M2, M4, M7, M12 | [manual] |

## By Screen

| Screen | Heading | Region | Keyboard entry point |
|----------|---------|------|-------------------|
| M1 | h1 "Backlog" | main | Space address field |
| M2, M2m | h1 "Backlog issues" | main | Project filter, or the Filters button on phones |
| M3 | h2 inside dialog | dialog | Task search field |
| M4 | h1 "PR watches" | main | New watch button |
| M5 | h1 "Backlog dashboard" | main | Query select field |
| M6 | (owned by Kandev) | — | Task menu |
| M7 | h2 inside dialog | dialog | PR number or link field |
| M8 | h2 "Backlog" | complementary | Collapse/expand block button |
| M9–M11 | (owned by Kandev) | — | Follows Kandev's flow |
| M12 | (inside M1 and the task card) | status | Review watches button |

## Tools

- **Automated**: axe-core runs in the UI tests (Vitest). No level A or AA violations are accepted (US8.2, AC8.2.2).
- **Manual**: keyboard and screen reader checks (NVDA or VoiceOver), plus the 320px width, during the manual check when the thin slice is done and before the first release (`team-practices`).

## Sources

- [desc] Initial description: a Kandev plugin for Nulab Backlog.
- [scope] Workflow-selected scope: `feature`.
- [Q3]: answer in `refined-mockups-questions.md`.
- `requirements` NFR9; `stories` US5.4, US8.2; `mockups.md`.

## Assumptions & Open Questions

- [assumption] The parts Kandev renders (task menu, `#` suggestions, PR creation flow) meet Kandev's own accessibility level. The plugin is only responsible for the data and labels it provides.
