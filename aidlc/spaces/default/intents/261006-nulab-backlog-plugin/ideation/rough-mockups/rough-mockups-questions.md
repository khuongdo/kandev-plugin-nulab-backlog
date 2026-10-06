# Rough Mockups — Clarifying Questions

## Sources

- [desc] Initial description: "feature dựa trên plugin này https://github.com/kdlbs/kandev-plugin-bitbucket hãy tạo 1 plugin tương tự cho kandev (https://github.com/kdlbs/kandev) nhưng dùng cho nulab backlog dựa trên public api ở đây https://developer.nulab.com/docs/backlog/"
- [scope] Workflow-selected scope: `feature`.

Context: the plugin is shown inside the Kandev UI. The Bitbucket plugin uses Kandev's built-in extension points: the integration settings page, a sidebar entry, the task action menu, indicators on the task list, and `#` references. The in-scope capabilities are settled in `scope-document`. The questions below settle the screens and interaction flows at sketch level.

---

## Q1. User entry points

Where in Kandev will users reach the Backlog features?

- A. Like the Bitbucket plugin: an integration settings page to connect, plus actions in the task menu
- B. Add a separate "Backlog" page on the sidebar (browse issues, watch PRs, dashboard), plus the settings page to connect
- C. Both A and B
- D. Not yet defined
- X. Other (please specify)

[Answer]: A

## Q2. How to find and pick issues

How do users find a Backlog issue to create or link a task?

- A. An issue list page with filters (project, status, assignee, keyword), each row with "Create task" and "Link" buttons
- B. An issue picker dialog opened from a task (search by keyword or issue key such as `PROJ-123`)
- C. Both A and B
- D. Not yet defined
- X. Other (please specify)

[Answer]: A

## Q3. Space connection screen

What should connecting a Backlog space look like?

- A. A single form: enter the space address (for example `myteam.backlog.com`), choose the sign-in method (API key or OAuth), then click connect
- B. A multi-step wizard (space address → sign-in method → connection check → pick default project)
- C. Not yet defined
- X. Other (please specify)

[Answer]: A

## Q4. Look and style

Which standard should the plugin UI follow?

- A. Use Kandev's existing UI components and style exactly, like the Bitbucket plugin
- B. Have its own Backlog identity (colours, icons) inside the Kandev UI frame
- C. Not yet defined
- X. Other (please specify)

[Answer]: A

## Q5. Devices

The Bitbucket plugin shows pull request indicators on both desktop and phone. The Backlog plugin needs to support:

- A. Desktop and phone, like Kandev
- B. Desktop only
- C. Not yet defined
- X. Other (please specify)

[Answer]: A

## Q6. Display language

Which language does the plugin UI text use?

- A. English only, like Kandev
- B. English and Japanese (many Backlog users are in Japan)
- C. Follow the language Kandev is currently displaying (if Kandev supports several languages)
- D. Not yet defined
- X. Other (please specify)

[Answer]: C

## Q7. Accessibility

What level of accessibility (keyboard users, screen readers) is needed?

- A. Basic WCAG 2.1 AA: fully operable by keyboard, every input has a label, information is never conveyed by colour alone
- B. Whatever level Kandev already has, no separate target
- C. Not yet defined
- X. Other (please specify)

[Answer]: A

## Q8. Clarification: where the issue list page lives

In Q1 you chose entry points like the Bitbucket plugin (settings page and task menu, no separate page). In Q2 you chose an issue list page with filters. In addition, PR watch and saved dashboards (in scope per `scope-document`) also need a place to be shown. Kandev lets plugins add entries under "Integrations" on the sidebar. Where should the issue list page go?

- A. A Backlog entry under "Integrations" on the sidebar, containing the issue list, PR watch and dashboard
- B. No list page; switch to an issue picker dialog opened from a task, and put PR watch and dashboard under "Integrations"
- X. Other (please specify)

[Answer]: A

## Consolidated Summary Confirmation

Summary of the answers (with the Q8 clarification merged in):

- Q1: A. Entry points like the Bitbucket plugin: an integration settings page to connect, plus actions in the task menu
- Q2: A. Issue list page with filters; each row has "Create task" and "Link" buttons
- Q3: A. A single connection form: space address, sign-in method (API key or OAuth), connect button
- Q4: A. Use Kandev's UI components and style
- Q5: A. Support desktop and phone
- Q6: C. Display text follows the language Kandev is using
- Q7: A. Basic WCAG 2.1 AA accessibility
- Q8: A. A Backlog entry under "Integrations" on the sidebar, containing the issue list, PR watch and dashboard

Does this all look correct before I generate the artifact?

- Looks correct
- Request changes

[Answer]: Looks correct
