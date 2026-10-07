# Requirements — github-parity-actions

## Intent Analysis

Initial description [desc]: "Thêm chức năng tương tự như Github integration trong kandev: (1) Quick actions (default Implement, Investigate, ...) cho phép người dùng tùy biến prompt, thêm quick actions tương tự github; (2) Thêm default queries cho Issue và PR -> cài sẵn 1 query mặc định; (3) Sửa UI, margin, vị trí buttons tương tự như giao diện github integration"

Goal: make the `/backlog` page of the Nulab Backlog plugin work like Kandev's first-party GitHub page. Users start agent tasks from an issue or pull request with one menu pick and a ready-made prompt. The page opens on a useful list right away instead of an empty one. Layout and button positions match the GitHub page users already know.

Type: enhancement of an existing brownfield plugin (v0.1.1). Multi-component: Go backend (`internal/issues`, `internal/git`, `internal/plugin`, manifest) and TypeScript UI (`ui/src/page`, `ui/src/issues`, `ui/src/git`, `ui/src/settings`). Depth: Minimal.

Sources: current behaviour comes from [business-overview.md](../../../../codekb/kandev-plugin-nulab-backlog/business-overview.md), [architecture.md](../../../../codekb/kandev-plugin-nulab-backlog/architecture.md) (section "External Reference: Kandev GitHub Integration (v0.96.0)"), and [code-structure.md](../../../../codekb/kandev-plugin-nulab-backlog/code-structure.md). Decisions come from [requirements-analysis-questions.md](requirements-analysis-questions.md) (Q1-Q5, all answered A). Team practices come from `aidlc/spaces/default/memory/team.md`.

## Functional Requirements

### FR1 Quick actions: launch a task from a row [Q1, Q2]
- **FR1.1** Every issue row and every pull request row on `/backlog` shows a "+ Task" menu at the end of the row. The menu uses the host `IntegrationStartTaskMenu` in the `ChangeRequestRow` `action` slot. It lists the workspace's quick actions for that kind (issue or PR), each with its icon, label and hint.
- **FR1.2** Picking a quick action opens Kandev's native `TaskCreateDialog`, prefilled with:
  - title `"<action label>: <issue summary | PR title>"`, truncated the same way as Kandev;
  - description = the action's prompt template with `{{url}}` (the Backlog web URL of the issue/PR) and `{{title}}` replaced. Unknown placeholders stay as written.

  The user can edit both fields before creating the task. Cancelling creates nothing.
- **FR1.3** When the dialog's create succeeds, the plugin links the new task to the issue (`issues.link`) or PR (`git.prs.link`) automatically. If the link call fails, the task is kept, and the user sees an error that says the task was created but not linked.
- **FR1.4** The issue row's current "Create task" item, which created a task with no prompt, is replaced by the "+ Task" menu. "Link to task" stays available on the row [assumption].
  - Given an issue row, when the user opens "+ Task" and picks "Investigate", then the create dialog opens with title "Investigate: <summary>" and the interpolated Investigate prompt.
  - Given the dialog was confirmed, when Kandev returns the created task, then the issue shows as linked to that task.

### FR2 Quick actions: defaults and customization [Q2]
- **FR2.1** Default quick actions match Kandev's GitHub defaults, with the wording changed from GitHub to Backlog:
  - Issues: Implement (`code`), Investigate (`search`), Reproduce (`bug`).
  - Pull requests: Review (`eye`), Address feedback (`message`), Fix CI (`tool`).

  The default texts are not translated (persisted seed, as in Kandev).
- **FR2.2** Quick actions are stored per workspace in plugin state. When the stored list for a kind is empty or missing, the defaults for that kind are used.
- **FR2.3** The plugin settings have a "Quick actions" section with Issues and Pull requests tabs. In it the user can edit each action's icon, label, hint and prompt template, add an action (new: label "New action", icon `sparkle`, empty prompt), delete an action, and reset a kind to its defaults.
- **FR2.4** Validation: label is required and at most 100 characters; prompt template is at most 4000 characters [assumption]; at most 20 actions per kind [assumption]; icon is one of `eye message tool code search bug sparkle check`. Invalid input is refused with a clear message and nothing is saved.
  - Given a workspace with no stored actions, when the issue "+ Task" menu opens, then Implement, Investigate and Reproduce are listed.
  - Given the user deleted every PR action and saved, when the PR "+ Task" menu opens, then the PR defaults are listed (fallback).
  - Given the user edited and saved the Implement prompt, when Implement is picked, then the dialog uses the edited prompt.

### FR3 Default query for the pull request list [Q3]
- **FR3.1** The PR list has a built-in preset "Open, assigned to me". It is not stored. It means status open, assignee = the connected user, and it runs on the first repository of the selected projects (ordered as `git.repos`/repository picker order).
- **FR3.2** When the page opens on the Pull requests view, the starred default saved PR query is applied if one exists. Otherwise the built-in preset is applied, so the list is never empty only because nothing was chosen.
- **FR3.3** A saved PR query can be starred as the default. At most one saved PR query per workspace is the default. Starring another moves the star, and un-starring leaves none. Deleting the default query removes the default.
- **FR3.4** If the selected projects have no repository, the PR list shows the existing empty/connect guidance instead of an error.
  - Given no starred PR query, when the user opens Pull requests, then open PRs assigned to the connected user in the first repository are listed.
  - Given saved query "Q" is starred, when the user opens Pull requests, then "Q" is applied.

### FR4 Default query and saved queries for the issue list [Q4]
- **FR4.1** The issue list has a built-in preset "Assigned to me, open". It is not stored. It means assignee = the connected user (`Connection.ConnectedUserID`), and statuses are every status except the project's closed/"Closed" status, across the selected projects.
- **FR4.2** `issues.list` accepts the assignee value `me`. The server resolves it to the connected user's id. The browser never sends or needs the numeric id for this.
- **FR4.3** Users can save the current issue filters (keyword, project, status, assignee incl. `me`) as a named saved issue query. Saved issue queries are stored per workspace: name required and at most 100 runes, at most 50 per workspace, the same rules as saved PR queries. They can be renamed and deleted in Settings, next to the saved PR queries.
- **FR4.4** A saved issue query can be starred as the default, at most one per workspace, with the same rules as FR3.3.
- **FR4.5** When the page opens on the Issues view, the starred default saved issue query is applied if one exists. Otherwise the built-in preset is applied.
  - Given no starred issue query, when the user opens `/backlog`, then only open issues assigned to the connected user are listed.
  - Given the user saves filters as "My bugs" and stars it, when they reopen the page, then "My bugs" is applied.

### FR5 GitHub-aligned page layout [Q5, desc]
- **FR5.1** The Issues/Pull requests tabs are replaced by a scope bar (host `IntegrationScopeBar`). The bar holds the kind switch (Issues / Pull requests), the built-in preset pill(s) of that kind, and a "Saved" menu on the right that lists the saved queries of that kind with a star toggle for the default. Bar padding: `px-4 py-2 sm:px-6`.
- **FR5.2** Below the scope bar is a toolbar laid out like the host `IntegrationListToolbar`: `border-b px-4 py-2.5 sm:px-6`. It has the title and result count on the left, the filters in the middle, and last-updated text plus a ghost refresh icon button on the right. The "Save query" action stays reachable from the toolbar.
- **FR5.3** Result lists use `px-3 py-4 md:px-6` padding and `ChangeRequestRow` rows for both issues and PRs, with the "+ Task" menu at the row end (FR1.1) and pagination at the bottom.
- **FR5.4** Settings card, enable switch and nav entry keep their current behaviour and stay independent of the enabled state (BR5.4/BR7.6/BR7.8). `settingsHref()` keeps its format.
  - Given the `/backlog` page at 1280px width, when it renders, then the scope bar and toolbar have 24px side padding and refresh sits at the right end of the toolbar.

## Non-Functional Requirements

- **NFR1 Security**: prompt templates and saved queries are plain user data. They never contain or log API keys, tokens or Git passwords, and the existing redaction test covers the new actions. New actions use `authenticated` access, except where they change workspace-wide settings: editing quick actions and setting a default query are workspace-wide but keep `authenticated`, the same as the existing `git.queries.*` [assumption].
- **NFR2 Compatibility**: works on the minimum Kandev version `0.96.0` and uses only host components exposed there. Existing action keys stay unchanged. New keys match `^[a-z0-9][a-z0-9._-]*$`. Existing stored data (PR saved queries, links, watches) keeps working without migration. A missing `isDefault` field means not default.
- **NFR3 Quality**: Go coverage floor 80% for `./internal/...` and `./server/...` with `go test -race`. TDD ordering per team practice. Minimal test strategy: at least one test per FR sub-group, plus the existing suite (Go 9 packages, Vitest 286 tests) stays green. `tsc`, ESLint, Prettier, golangci-lint stay clean.
- **NFR4 Performance**: opening `/backlog` makes no more Backlog API calls than today plus the single list call for the default query. Quick action menus render from already-loaded data, with no extra request when the menu opens.
- **NFR5 Accessibility**: the "+ Task" trigger and star toggles have accessible labels. Everything works by keyboard, as the host components do.

## Constraints

- Plugin SDK and host UI at Kandev `v0.96.0` (`.kandev-sdk-ref`). Only `internal/plugin` and `server/` import `pluginsdk`.
- Backlog's PR API is per repository: one `git.prs.list` call covers one repository.
- Team code style, Makefile targets and CI gates in `team.md` apply. No new third-party dependencies.

## Assumptions

- [assumption] "Link to task" stays as a row action, and the old prompt-less "Create task" item is removed in favour of the "+ Task" menu (FR1.4).
- [assumption] Prompt template max 4000 characters and at most 20 quick actions per kind (FR2.4).
- [assumption] Prompt wording follows Kandev's GitHub defaults with "GitHub" replaced by "Backlog", e.g. Implement: `Implement the changes described in the Backlog issue at {{url}} (title: "{{title}}"). Open a pull request when complete.`
- [assumption] Quick actions and default stars are workspace-wide (shared by every member), with `authenticated` access like `git.queries.*` (NFR1).
- [assumption] The Quick actions section saves through its own Save button in the section (the plugin's per-section save model), not through a page-level Save/Discard bar.
- [assumption] "Open" for issues means every status except the project's closed status (Backlog status id 4 "Closed" or the project's last status), resolved on the server.

## Out of Scope

- Customizing the built-in preset list (Kandev's "customizable defaults"). Only one built-in preset per kind ships.
- Extra GitHub presets (Mentions, Drafts, Recently merged, Recently closed).
- Starting the agent automatically (`StartAgent`/`Launch.Prompt`). The native dialog decides.
- Mobile views picker (`MobileViewsPicker`).
- Per-user (rather than per-workspace) quick actions or defaults.

## Open Questions

- Which exact Backlog status counts as "closed" for projects with custom statuses. Resolve in Code Generation from `internal/backlog` status data. Default: status id 4.
