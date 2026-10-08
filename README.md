# Kandev plugin for Nulab Backlog

A [Kandev](https://github.com/kdlbs/kandev) plugin that connects a Kandev workspace to a
[Nulab Backlog](https://backlog.com) space.

## What it does

- Adds a **Backlog** card under **Settings > Integrations**, with the plugin's outline icon and
  an on/off switch. The card holds every Backlog setting as framed sections: Connection,
  Projects, PR watches, Issue watches, Saved queries, Quick actions, Issue sync and Source
  control (with Git access).
- Adds one **Backlog** entry to the **Integrations** menu on the Kandev home page when Backlog is
  on in at least one workspace when Kandev loads (see [Turn Backlog on or off](#turn-backlog-on-or-off)). It opens the
  `/backlog` page with an **Issues** and a **Pull requests** list, like the GitHub integration:
  a scope bar with the two kinds, a built-in preset and a **Saved** menu, a toolbar with the
  filters and refresh, and the results below.
  When the workspace is not connected or Backlog is off, the page shows an alert with a link to
  the Backlog settings instead.
- **Quick actions**: every issue and pull request row has a **+ Task** menu (Implement,
  Investigate, Reproduce for issues; Review, Address feedback, Fix CI for pull requests). A pick
  opens Kandev's own create-task dialog with a ready prompt (`{{url}}` and `{{title}}` filled in);
  the created task is linked to the issue or pull request. Edit, add, delete or reset the
  actions per workspace under **Settings > Integrations > Backlog > Quick actions**.
- **Default queries**: the Issues list opens on **Assigned to me, open** and the Pull requests
  list on **Open, assigned to me** (first repository), unless a saved query of that kind is
  starred as the default in the **Saved** menu or in Settings.
- A workspace admin connects the workspace to one Backlog space with an **API key**.
  The plugin checks the key with Backlog (`GET /api/v2/users/myself`) before saving it,
  then shows `Connected as <name> @ <space>`.
- Only `https` space addresses on `backlog.com`, `backlog.jp` or `backlogtool.com` are accepted.
- The key is stored only in Kandev's encrypted secret store. It is never returned to the
  browser and never written to logs.

## Upgrade notes

### 0.5.2: link a Backlog issue from the task, GitHub-style

- The task's Link menu (Kanban card menu, task switcher) has a new "Link Backlog issue" item, next
  to "GitHub Issue" and "Link Backlog pull request". Type an issue key such as `PROJ-123` or paste
  the issue link (`https://<space>.backlog.com/view/PROJ-123`, also `.backlog.jp` and
  `.backlogtool.com`) and press Save or Enter. The item is hidden while the task already has a
  Backlog issue; use "Unlink Backlog issue" first to change it.
- The "Link to task" dialog on the Backlog issues page now looks like Kandev's GitHub link dialog:
  a short description, a Save button, Enter saves the chosen task, errors show in red under the
  list, and a "linked" message appears.
- After linking from either place, the issue badge on the task updates at once instead of at the
  next refresh.
- Nothing to do after upgrading: no setting or permission changes.

### 0.5.1: connect GitHub or GitLab with the gh / glab CLI login

- Settings > Source control: the GitHub card has a "Use gh CLI login" button and the GitLab card
  a "Use glab CLI login" button, next to the token field. The plugin then reads the token from
  the CLI that is logged in on the Kandev server (`gh auth token`, `glab config get token`),
  checks it, and never stores it. It follows `gh auth login` / `refresh` / `switch` within
  5 minutes.
- The CLI runs on the machine that runs Kandev, as the Kandev server user. It needs `gh` 2.17 or
  later / `glab`, logged in for that user; it does not work when Kandev runs where the CLI is not
  installed (for example a Docker image without it). If the CLI stops working, the card shows
  "The gh CLI is not available or not logged in on the Kandev server."; there is no fallback to a
  typed token.
- Saving a typed token switches back to the token method; Remove clears either method.
- GitLab requests now send the token as `Authorization: Bearer`. Existing GitLab personal access
  tokens keep working; nothing to do.
- Nothing to do after upgrading: existing connections stay on the token method.

### 0.5.0: Backlog issue on task rows and the task top bar

- The Backlog issue badge now shows on Home > Tasks rows, in the sidebar task list and in the
  task top bar (right of the workflow steps), not only on Kanban cards. Hover or focus it to see
  the issue key, summary and status; click it to open the issue.
- The issue summary is filled in for existing links at the next issue sync; until then the badge
  shows the key and status.
- Home > Integrations shows Backlog only when Backlog is on in at least one workspace. After
  turning Backlog on or off, reload the page. The Settings > Integrations card is always there.
- Backlog settings: Projects now comes right after the connection; the empty issue watch list
  says "No issue watches yet"; the extra "Add watch" button inside empty watch lists is gone.
- Nothing to do after upgrading: no setting or permission changes.

### 0.4.2: smaller package, install From URL

- The package is smaller (about 23 MB instead of 29.5 MB): it now carries server
  executables for Linux and macOS on amd64 and arm64 only.
- **Windows servers are no longer supported.** Kandev running on Windows cannot install
  or upgrade to 0.4.2; stay on 0.4.1 there.
- Install **From URL** (Settings > Plugins > Install plugin) with the GitHub Release
  package URL. Uploading the file still works, but Kandev stops reading an upload after
  30 seconds, which shows as `Plugin install failed: 502` on slow connections (see
  "Troubleshooting" below).
- Nothing else changes: no data, setting or permission changes.

### 0.4.1: GitHub-style lists

- Issue search runs when you press **Enter** (or leave the search box after editing), not while
  you type.
- The issue list and every pull request list (Backlog Git, GitHub, GitLab and Bitbucket) use a
  GitHub-style toolbar: title and count, searchable dropdown filters without field labels, a
  **Status (n)** menu for pull request statuses, then last updated and refresh. On phones the
  filters stack at full width.
- Linked tasks show the task title; a row with several tasks shows a **Tasks (n)** menu.
- The Backlog badge on a Kanban card opens the Backlog issue in a new tab.
- Nothing to do after upgrading: no data, setting or permission changes.

### 0.4.0: GitHub, GitLab and Bitbucket pull requests

- New: pull requests from GitHub, GitLab and Bitbucket (cloud only) next to Backlog Git: PR
  list, saved queries, PR watches, PR status, and links to tasks and Backlog issues,
  including automatic links when a Backlog issue key is in the branch name or PR title.
- Setup: an admin adds one read-only token per service and maps Backlog projects to
  repositories in **Source control** on the Backlog settings card. The required token
  scopes are shown there. See [GitHub, GitLab and Bitbucket](#github-gitlab-and-bitbucket).
- Moved: the Backlog Git **Git access** form is now inside the **Source control** section.
- Nothing to do after upgrading: existing Backlog Git links, watches, saved queries and
  credentials keep working unchanged.

### 0.3.0: Backlog is now off by default (upgrading from v0.2.0 or earlier)

Backlog is now opt-in: after installation it is off in every workspace until an admin turns
it on. A workspace that never touched the switch on v0.1.0 to v0.2.0 turns **off** after the
upgrade. Its saved connection is kept, but issue sync, issue and PR watches and Git
credentials pause. To resume, a Kandev admin turns on the switch on the Backlog card and
clicks **Save**; there is no need to connect again. Workspaces where the switch was saved
(on or off) keep their setting.

### 0.2.0

- Rows on `/backlog` get a **+ Task** quick action menu. The issue row's old **Create task** item
  is gone; **Link to task** stays in the row menu.
- The Issues / Pull requests tabs became a scope bar. Saved queries are picked from its
  **Saved** menu (the saved-query dropdown in the Pull requests toolbar was removed), and
  issue filters can now be saved too.
- The lists open on a default query instead of an empty filter. Star a saved query to make it
  the default.
- Existing saved PR queries, links and watches keep working; no data migration is needed.

### 0.1.1

- The **Integrations** menu now has a single **Backlog** entry. The `/backlog` page has an
  **Issues** tab and a **Pull requests** tab (`/backlog?scope=prs`).
- The `/backlog/watches` and `/backlog/dashboard` pages were removed. Bookmarks to them no
  longer work.
- PR watches, the new **issue watches** and saved PR queries now live in
  **Settings > Integrations > Backlog**. New saved queries are created from the
  **Pull requests** tab with **Save query**; edit and delete them in Settings.
- The plugin icon is a new outline icon; the plugin no longer ships the Nulab logo.
- Existing PR watches, saved queries and issue links are kept. No data migration is needed.

## Turn Backlog on or off

Backlog is off by default after installation, in every workspace. A Kandev admin turns it on
for one workspace with the switch on the Backlog card, then **Save**, and then connects. The
same switch turns it off again. While it is off, the plugin refuses to connect or call Backlog
for that workspace, but it keeps the existing connection: turning it back on restores it
without connecting again. The **Backlog** card under **Settings > Integrations** always stays,
so Backlog can always be turned back on.

The **Backlog** entry in the home **Integrations** menu (and the `/backlog` page) is added only
when Backlog is on in at least one workspace when Kandev loads. Kandev cannot hide a menu entry
after it has loaded, so after turning Backlog on or off, reload the page to show or hide the
entry. If the plugin cannot read the on/off state at load (an error, or no answer within
3 seconds), it shows the entry.

The plugin icon is an original outline drawing, not the Nulab logo; see
[docs/brand/backlog-logo.md](docs/brand/backlog-logo.md).

## Build

Requirements: Go 1.26, Node.js 22, `make`, and a checkout of Kandev next to this repository
at the commit in `.kandev-sdk-ref`:

```
git clone https://github.com/kdlbs/kandev ../kandev
git -C ../kandev checkout "$(cat .kandev-sdk-ref)"
(cd ui && npm ci)
make package verify-package
```

This writes `dist/nulab-backlog-<version>.tar.gz` and `dist/checksums.txt`.

Other targets: `make check-format vet lint test coverage build`. Go tests run with `-race`,
and `make coverage` fails below 80% line coverage.

## Install on Kandev

Supported Kandev server platforms: Linux and macOS, on amd64 and arm64. Windows servers
are not supported: the package carries no Windows executable, so Kandev on Windows
refuses to install it.

1. Sign in to your Kandev server (0.96.0 or later) as an admin.
2. Open **Settings > Plugins**, click **Install plugin** and install the package in one of
   two ways:
   - **From URL** (recommended): paste the package URL of the GitHub Release, for example
     `https://github.com/khuongdo/kandev-plugin-nulab-backlog/releases/download/v<version>/nulab-backlog-<version>.tar.gz`.
     The Kandev server downloads the package itself, so the speed of your own connection
     does not matter.
   - **Upload file**: upload `nulab-backlog-<version>.tar.gz` (from the GitHub Release or
     your own `dist/`). The whole upload must finish within the server's read limit
     (30 seconds by default).
3. Open **Settings > Integrations > Backlog**, turn the switch on and click **Save**
   (Backlog is off until an admin turns it on).

### Troubleshooting: `Plugin install failed: 502`

If an upload fails with `Plugin install failed: 502` after about 30 seconds, and the
Kandev server log shows a 400 response with `missing multipart field "package"`, the
upload was slower than the server's read limit: Kandev stopped reading the request after
30 seconds, and a proxy in front of it (for example `tailscale serve`) reported 502.
Either:

- install **From URL** instead (see above), or
- raise the limit on the Kandev server with `KANDEV_SERVER_READTIMEOUT` (in seconds, for
  example `KANDEV_SERVER_READTIMEOUT=120`) and restart Kandev.

## Connect with an API key

1. In Backlog, open **Personal Settings > API** and create an API key.
2. In Kandev, enter your space address (for example `myteam.backlog.com`) and the key.
3. Click **Connect**. The page shows `Connected as <name> @ <space>` when Backlog accepts the key.

Connecting again with another key replaces the connection only after Backlog accepts the new key.

## CI checks

Every pull request and every push to `main` runs `.github/workflows/ci.yml`. Its first job,
`changes`, lists the changed files (`go run ./cmd/ci changes -base <sha> -head <sha>`). When every
changed file is under `aidlc/`, `.claude/` or `docs/`, or is `README.md`, `LICENSE` or `.gitignore`,
the two app jobs below are skipped, and GitHub reports a skipped job as passing. Any other file, or
a change set that cannot be listed, runs them in full. The list lives only in `internal/ci/changes.go`.

- `checks` runs `make check-format vet lint test coverage check-secrets build package verify-package`.
  `lint` also runs `actionlint` and the workflow policy (every action pinned to a full commit SHA,
  default `permissions: contents: read`, no `pull_request_target`, write permissions only in the
  release `publish` job). `check-secrets` fails when test data or test artifacts hold a
  credential-shaped string; fake secrets in tests must start with `TESTSECRET-`.
- `packaged-host-contract` builds Kandev at exactly the `min_kandev_version` in `manifest.yaml`,
  starts a throwaway server, installs the packaged plugin and runs two of its actions
  (`make contract-test`). Locally: `git -C ../kandev worktree add ../kandev-min v<min_kandev_version>`,
  then `make package contract-test` (needs gcc; uses ports 38529 and 39529).

`.github/workflows/secrets.yml` runs the job `secret-scan` (`make check-secrets`) on every pull
request and every push to `main`, with no path filter, so records and docs are scanned too.

In the repository settings, protect `main`: require a pull request, block force-pushes and
deletion, and require `checks`, `packaged-host-contract` and `secret-scan` to pass before merging.
Also add a tag rule so `v*` tags cannot be moved or deleted.

## Releasing

1. In a pull request, set `version` in `manifest.yaml` to the new version `X.Y.Z`.
2. For the first release only, commit the second manual-check record
   `docs/manual-checks/<YYYY-MM-DD>-first-release.md` (from `TEMPLATE.md`) with `| Result | pass |`.
   While no stable GitHub Release is published yet, the release preflight refuses a `vX.Y.Z` tag without it.
3. After the pull request is merged, run `make package verify-package` on `main`, then tag and push:
   `git tag vX.Y.Z && git push origin vX.Y.Z`. Creating the tag is the approval step.
   The `release` workflow re-runs every check, package verification and the contract test, refuses a
   tag that is not on `main`, differs from the manifest version or already has a Release, and then
   publishes a GitHub Release with `nulab-backlog-X.Y.Z.tar.gz`, `checksums.txt` and a build
   provenance attestation. A tag with a `-suffix` (for example `v0.0.1-rc.1`) is published as a
   pre-release.
4. Check the download:
   `gh attestation verify nulab-backlog-X.Y.Z.tar.gz --repo khuongdo/kandev-plugin-nulab-backlog`.
5. If a release is broken, mark it as broken in its notes, reinstall the previous version and fix
   forward with a new patch release. A released tag is never deleted or overwritten.

Installing a release on a Kandev server is a manual step (see "Install on Kandev").

## Marketplace

After a release, run `make marketplace-entry` to print the entry for the Kandev marketplace
registry. Add `REGISTRY=<path to plugin-registry/plugins.yaml>` to check it against the current
catalogue: the run fails when the catalogue lists this repository with another id, or uses the id
for another repository. Then fork `kdlbs/kandev`, add the entry to `plugin-registry/plugins.yaml`
and open a pull request; a Kandev maintainer reviews it. Kandev shows the installed package as
unsigned: it checks the package's checksums but not signatures. The GitHub attestation above is how
users check where the package came from.

## License

MIT. See [LICENSE](LICENSE).

## Sign in with OAuth

An API key always works. To also offer **Sign in with Nulab** (OAuth), a Kandev admin sets up
one Nulab OAuth application for the whole Kandev server:

1. In the Nulab developer site (<https://developer.nulab.com/>), register an application of the
   Backlog type. Use this redirect URI, exactly:
   `<public_base_url>/api/plugins/nulab-backlog/webhooks/oauth-callback`,
   where `<public_base_url>` is the address users open Kandev with, for example
   `https://kandev.example.com`. It must be `https`; plain `http` is accepted only for
   `localhost` and `127.0.0.1`, for a local trial.
2. In Kandev, open **Settings > Plugins > Nulab Backlog** and fill in the three fields:
   - `oauth_client_id`: the application's client ID.
   - `oauth_client_secret`: the client secret. Kandev keeps it in its encrypted vault and shows
     it masked.
   - `public_base_url`: the same address as in step 1, without a trailing path such as
     `/settings`.
3. In **Settings > Integrations > Backlog**, a workspace admin chooses **OAuth** as the sign-in
   method, enters the space address and clicks **Sign in with Nulab**.

The plugin refreshes the OAuth sign-in by itself before it expires. If Backlog refuses the
refresh, the settings page shows **Sign in again**. Without the three fields, choosing OAuth
says that OAuth is not set up on this Kandev server.

**Unsupported: Kandev under a path prefix.** OAuth sign-in works only when Kandev is served at
the root of its address (for example `https://kandev.example.com`). If `public_base_url` has a
path (for example `https://example.com/kandev`), the browser does not send the sign-in cookie to
the callback, because the cookie is scoped to `/api/plugins/nulab-backlog/webhooks/oauth-callback`
without the prefix. Sign-in then fails safely with "failed" and nothing is stored. Use an API key
in that setup.

## Backlog Git and pull requests

These features need pull requests and Git hosting on the Backlog plan of the space (assumption
A1). They only use repositories of the projects selected in the Backlog settings.

- **Git access.** In **Settings > Integrations > Backlog**, the **Git access (optional)** section takes
  your Git user name and password: your Backlog password, or the Git password Backlog gives you
  when two-step verification is on. Kandev stores them in its encrypted vault and never shows them
  again; the page only says **Stored**. **Test connection** also checks them against the first
  repository of the first selected project. Disconnecting, or connecting to another space, deletes
  them.
- **Repository source.** When you create a Kandev task, **Backlog** is a repository source. Kandev
  fetches and pushes with the stored Git access, one 15-minute lease at a time. Backlog has no
  branch list, so the branch choices are the branch names of the repository's newest pull
  requests, and the default branch is the most used base branch (or `master`).
- **Link a pull request.** In a task's **Link** menu, choose **Link Backlog pull request** and paste
  a pull request link, or enter `<repository>#<number>`, or just a number when the task has one
  Backlog repository. The task card then shows the pull request's state (for example
  `Open – Lan`); Kandev refreshes it. Unlinking never changes Backlog.
- **Create a pull request.** Kandev's **Create PR** dialog pushes the task's branch and then creates
  the pull request in Backlog, linked to the task. If the title or body names an issue of a selected
  project, the pull request is related to that issue with `Related: <KEY>`; it never closes it. If
  a pull request is already open for the branch, you are asked whether to link it instead.
- **Pull requests list.** On `/backlog`, **Pull requests** lists the pull requests of one
  repository of a selected project, 20 per page, newest first, with status, assignee and creator
  filters (Backlog has no pull request search). Each row shows the number, title, status, author,
  assignee, last update and the linked Kandev task. `/backlog?scope=prs` opens this list directly.
- **PR watches.** In **Settings > Integrations > Backlog**, the **PR watches** section creates
  Kandev tasks for new pull requests that match a filter (repository, status, assignee, creator,
  linked issue). Add and edit watches in a dialog; each row's menu has Edit, Run now,
  Pause/Resume and Delete. Every 5 minutes each watch creates at most 10 tasks, oldest pull
  request first, and never creates a task twice for one pull request, even after a deleted task or
  a restart. Disconnecting turns watches off; reconnecting to the same space brings them back
  **Paused**. Any signed-in member can manage watches.
- **Saved queries.** **Save query** on the Pull requests list saves the current repository and
  filters; saved queries then appear as presets there. Rename or delete them in the **Saved PR
  queries** section of the settings.

## GitHub, GitLab and Bitbucket

Next to Backlog Git, the plugin reads pull requests from **GitHub**, **GitLab** and **Bitbucket
Cloud** and links them to Kandev tasks and Backlog issues. Only the cloud services are supported:
`github.com`, `gitlab.com` and `bitbucket.org` (GitHub Enterprise Server, GitLab self-managed and
Bitbucket Data Center are not). The plugin calls only `https://api.github.com`,
`https://gitlab.com/api/v4` and `https://api.bitbucket.org/2.0`, reads only, and never creates,
merges or comments on pull requests or hands out clone or push credentials for these services.

- **Tokens.** In **Settings > Integrations > Backlog > Source control**, a workspace admin enters
  one access token per service for the workspace, then can **Test**, replace or **Remove** it. The
  token is checked with the service's current-user call before it is saved, kept only in Kandev's
  encrypted secret store, and never shown again or written to logs. Every member sees the same pull
  requests. Members see the section read-only. Read scopes needed:
  - GitHub: a fine-grained token with **Metadata: read** and **Pull requests: read**, or a classic
    token with `repo` (`public_repo` for public repositories only).
  - GitLab: a personal access token with `read_api`.
  - Bitbucket: an API token with `read:repository:bitbucket` and `read:pullrequest:bitbucket`, or an
    app password with **Repositories: Read** and **Pull requests: Read**, together with your
    Bitbucket user name or email.
- **CLI login instead of a token (GitHub and GitLab).** See
  [Connect GitHub or GitLab with the CLI login](#connect-github-or-gitlab-with-the-cli-login).
- **Repositories per Backlog project.** For each selected Backlog project, an admin maps the
  repositories of each service: search the repositories the token can read, or type
  `owner/name` (GitHub), `group/project` (GitLab) or `workspace/repo` (Bitbucket). Each repository
  is checked with the service before it is saved. Lists, saved queries, watches and automatic
  linking only cover mapped repositories. Removing a mapping disables, but never deletes, the
  saved queries and watches that use it; they show **Repository no longer mapped**.
- **Pull requests list.** On `/backlog`, **Pull requests** has a **Provider** choice. For a service
  it lists one mapped repository of a selected project, 20 per page, with status (open, closed,
  merged; drafts and Bitbucket's declined are shown as such) and author filters, the branches,
  the last update and a link to the pull request. **Save query** saves the filters with the
  service; one saved query per service can be the default.
- **Linking.** Create a task from a row with **+ Task**, or open a task's **Backlog issue** panel and
  paste a pull request URL with **Link a pull request**; the URL must belong to a mapped
  repository. When a list refresh or a watch first sees a pull request whose source branch or title
  contains an issue key of a selected project mapped to that repository (for example
  `PROJ-123`), the pull request is **auto-linked** to that issue. Removing an auto-link is final:
  it is not created again. The panel also shows the GitHub and GitLab pull requests Kandev itself
  attached to the tasks of that issue, without any plugin token.
- **Watches.** PR watches work for every service: each run creates at most one new task, every
  N minutes (5 by default) as set in the watch dialog, and never twice for one pull request.
  Linked pull request states are refreshed every 5 minutes, like Backlog Git watches. A rate limit
  from one service never holds up another; the plugin does not retry.
- **Backlog switch and connection.** When Backlog is off for the workspace, every source control
  action is refused except reading the settings, and watches do not run. Unlike Backlog Git, the
  tokens, mappings, links, queries and watches of these services stay after a Backlog disconnect,
  a space change or a project deselection; items of a project that is no longer selected are
  only hidden from project-based views.

### Connect GitHub or GitLab with the CLI login

Instead of pasting a token, a workspace admin can press **Use gh CLI login** (GitHub) or **Use
glab CLI login** (GitLab) on the service's card. Bitbucket has no CLI option.

- The plugin runs the CLI **on the machine that runs the Kandev server, as the server's user and
  with its environment** — not on the admin's own computer. Log in there first: `gh auth login`
  (gh 2.17 or later, which has `gh auth token`) or `glab auth login`.
- The plugin reads the token with `gh auth token --hostname github.com` or
  `glab config get token --host gitlab.com`, checks it with the service's current-user call, and
  stores only the method and the account name. The token is never saved, logged or shown; any
  typed token of that service is deleted. The card then shows **Connected via gh CLI as …**.
- The token is read again at most every **5 minutes**, so `gh auth refresh`, `gh auth switch` or
  a new `glab auth login` on the server takes effect within 5 minutes without any action in
  Kandev. **Test** always reads it again.
- If the CLI is missing, not logged in, slow (10 s limit) or its token is refused, actions fail
  with "The gh CLI is not available or not logged in on the Kandev server" (glab respectively).
  There is no fallback to a typed token. This is the case when Kandev runs in Docker or on a
  machine without the CLI: use a token there.
- glab: only the login glab stores is read; a token given to glab only through the
  `GITLAB_TOKEN` environment variable is not seen.
- Saving a token switches the service back to the token method; **Remove token** disconnects
  either method and keeps the mappings. GitLab requests now send the token as
  `Authorization: Bearer`, which works for personal access tokens and glab's login.

## Backlog issues

These features use only the projects selected in the Backlog settings. The plugin only reads
from Backlog: it never creates, changes or comments on a Backlog issue.

- **Issues list.** **Integrations > Backlog** opens on **Issues**: the issues of the selected projects, newest
  updated first, 20 per page. Filter by project, status and assignee, or search; typing a full
  issue key such as `PROJ-123` puts that issue on top. Each row shows the Kandev tasks linked to
  it. On a phone the rows become cards and the filters open from **Filters (n)**.
- **Create task.** In a row's **…** menu, **Create task** creates a Kandev task in the workspace's
  default workflow: the issue title, its description plus a link back to Backlog, and its priority
  (High, Normal or Low). The task is linked to the issue. If the issue already has a task, you are
  asked whether to open it or create another one.
- **Link to task.** **Link to task** in the same menu links an existing task. A task links one issue
  only; several tasks may link the same issue. **Unlink Backlog issue** in the task's menu removes
  the link and changes nothing in Backlog.
- **Link an issue from the task.** In a task's **Link** menu, next to **GitHub Issue**, choose
  **Link Backlog issue** and enter an issue key (`PROJ-123`) or paste the issue's
  `https://<space>/view/PROJ-123` link. The entry is hidden while the task is linked; unlink first
  to link another issue.
- **`#` references.** In the message composer, type `#` and choose **Backlog issues** to reference
  an issue by key or title. Kandev checks the issue again when the message is sent.
- **Backlog panel in the task.** Open the **Backlog issue** panel from the task's **+** panel menu to
  see the issue's status, assignee, priority, due date, comments and attachments, read live from
  Backlog. Attachments open in Backlog; files over 10 MB are not previewed.
- **Issue watches.** In **Settings > Integrations > Backlog**, the **Issue watches** section creates
  a Kandev task for each Backlog issue that matches a filter: a selected project, one or more
  statuses, assignee and creator (anyone or me), and the interval in minutes (default 5). Each run
  creates at most one task, oldest issue first, including issues that existed before the watch was
  saved, and links it like **Create task**. An issue that already has a task, or that the watch
  handled before, never gets another one. A watch stops with a visible error when Backlog refuses
  the sign-in, the workflow was removed, or its ledger of 5000 handled issues is full. Rows have
  Edit, Run now, Pause/Resume and Delete; deleting a watch keeps its tasks and links.
- **Badge and status sync.** A linked task shows a badge such as `PROJ-120 · Resolved` on its
  Kanban card, on its row in **Home > Tasks** and the sidebar task list, and in the task top bar
  (right of the workflow steps). Hovering or focusing the badge shows the issue key, summary and
  status; clicking it opens the issue in a new tab. The summary is stored with the link and
  refreshed with the status, so older links show it after the next check. The
  plugin checks the status of linked issues every 5 minutes by default; an admin can set the
  interval (at least 1 minute) in the **Issue sync** section of the Backlog settings, and
  **Refresh** on the Issues list checks at once. A badge says **may be out of date** after three failed checks,
  and **Issue unavailable** when the issue was deleted or hidden. Deleting a task removes its link.
