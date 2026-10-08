**Collaborator:** aidlc-design-agent

## Contribution

Reviewed against the current GitHub card in `ui/src/settings/source-control-section.tsx` (one-click "Use gh CLI login" button, account line "Connected via {cli} CLI as {name}", results in one `role="status"` line). Proposed additions, integrable as acceptance-criteria edits:

1. **Picker placement and flow (US1.1, US1.2).** Keep the picker inline on the GitHub card, no dialog. "Use gh CLI login" first loads the accounts (button disabled, status "Loading gh accounts…"). If there is exactly one account, connect at once (AC1.1.4, so P2 sees no extra step). If there are two or more, show a labelled select ("GitHub account (gh)") with **Connect** and **Cancel** buttons under the token row. Use the host `Select` that `scm-watch-form.tsx` and `pr-list.tsx` already use. Cancel closes the picker and changes nothing.
2. **The active account is shown in text (AC1.1.1).** The option label is `alice (active in gh)`. Do not show "active" with colour or an icon alone (WCAG 1.4.1).
3. **Show the login, not only the display name (AC1.1.2, AC3.2.2).** The card stores the login as the identity, but today it shows `view.account`, which may be the GitHub `name`. In gh CLI mode the account line should read `Connected via gh CLI as @bob` (with the display name after it if it differs). Without this, P1 cannot tell which account a workspace uses.
4. **Change account (US1.2).** On a connected card in gh CLI mode, add a **Change account** button next to Test/Remove. It opens the same inline picker with the current login preselected. Proposed AC1.2.3: *Given W1 is connected as `bob`, When I open Change account and press Cancel, Then nothing is saved and the mappings are unchanged.*
5. **Make "pick another account" doable (US3.1).** The FR5 message tells the user to pick another account, so the card must offer that. Proposed AC3.1.4: *Given W1 shows the "bob is not logged in to gh" error, When I look at the card, Then Change account is available and its picker lists only the accounts gh still has, with nothing preselected that is not logged in.* The message should also say where gh runs, matching `scmCliUnavailable`: "bob is not logged in to gh on the Kandev server — log in again or pick another account."
6. **Messages are announced (accessibility).** Success messages stay in `role="status"`. Failures (cli unavailable, chosen login gone) use `role="alert"` so screen readers announce them; today both use `status`. When the picker opens, keyboard focus moves to the select. The select keeps arrow-key and Escape behaviour. Every new control has a visible label.
7. **Worktree note (US4.1).** Show it only in gh CLI mode, as muted text under the account line, never as a warning, so it does not look like an error. Proposed copy: "Agents working in task worktrees get their GitHub login from Kandev's own GitHub integration (or the executor profile), not from this plugin. Set it to the same account (@bob) for this workspace." Fill in the chosen login so the user knows which account to match. Read-only cards show the note and the login, but no picker.

## Positions

- AGREE: Breakdown A (one story per workflow step) — each story maps to one visible card state, which is easy to test in UI tests.
- AGREE: Worktree note as a Should story (Q2 = A) — it is the only visible place where P1 learns about the agent-shell limit.
- AGREE: Preselect the active account and connect directly when there is only one account — this keeps P2's upgrade and single-account use to zero extra clicks.
- OBJECT: AC1.1.2 / AC3.2.2 "shows connected as `bob`" — The card shows `view.account`, which may be the display name. The AC must require the login (`@bob`) to be visible, or the story's goal (knowing which account a workspace uses) cannot be checked.
- OBJECT: US3.1 has no AC for recovering on the card — The error says "pick another account", but no story makes the picker reachable while the card is in the error state. Add AC3.1.4 (item 5) or the message points to an action the UI does not offer.
