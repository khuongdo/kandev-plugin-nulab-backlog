# Kandev plugin for Nulab Backlog

A [Kandev](https://github.com/kdlbs/kandev) plugin that connects a Kandev workspace to a
[Nulab Backlog](https://backlog.com) space.

## What it does

- Adds a **Backlog** card under **Settings > Integrations**, with the plugin's outline icon and
  an on/off switch. The card holds every Backlog setting as framed sections: Connection,
  PR watches, Issue watches, Saved queries, Quick actions, Issue sync, Git access and Projects.
- Adds one **Backlog** entry to the **Integrations** menu on the Kandev home page. It opens the
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
without connecting again. The home entry and the `/backlog` page stay
available, so Backlog can always be turned back on.

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

1. Sign in to your Kandev server (0.96.0 or later) as an admin.
2. Open **Settings > Plugins** and upload `dist/nulab-backlog-<version>.tar.gz`.
3. Open **Settings > Integrations > Backlog**, turn the switch on and click **Save**
   (Backlog is off until an admin turns it on).

## Connect with an API key

1. In Backlog, open **Personal Settings > API** and create an API key.
2. In Kandev, enter your space address (for example `myteam.backlog.com`) and the key.
3. Click **Connect**. The page shows `Connected as <name> @ <space>` when Backlog accepts the key.

Connecting again with another key replaces the connection only after Backlog accepts the new key.

## CI checks

Every pull request and every push to `main` runs two jobs in `.github/workflows/ci.yml`:

- `checks` runs `make check-format vet lint test coverage check-secrets build package verify-package`.
  `lint` also runs `actionlint` and the workflow policy (every action pinned to a full commit SHA,
  default `permissions: contents: read`, no `pull_request_target`, write permissions only in the
  release `publish` job). `check-secrets` fails when test data or test artifacts hold a
  credential-shaped string; fake secrets in tests must start with `TESTSECRET-`.
- `packaged-host-contract` builds Kandev at exactly the `min_kandev_version` in `manifest.yaml`,
  starts a throwaway server, installs the packaged plugin and runs two of its actions
  (`make contract-test`). Locally: `git -C ../kandev worktree add ../kandev-min v<min_kandev_version>`,
  then `make package contract-test` (needs gcc; uses ports 38529 and 39529).

In the repository settings, protect `main`: require a pull request, block force-pushes and
deletion, and require **both** `checks` and `packaged-host-contract` to pass before merging.
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
- **Badge and status sync.** A linked task's card shows a badge such as `PROJ-120 · Resolved`. The
  plugin checks the status of linked issues every 5 minutes by default; an admin can set the
  interval (at least 1 minute) in the **Issue sync** section of the Backlog settings, and
  **Refresh** on the Issues list checks at once. A badge says **may be out of date** after three failed checks,
  and **Issue unavailable** when the issue was deleted or hidden. Deleting a task removes its link.
