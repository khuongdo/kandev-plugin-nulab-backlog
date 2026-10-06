# Kandev plugin for Nulab Backlog

A [Kandev](https://github.com/kdlbs/kandev) plugin that connects a Kandev workspace to a
[Nulab Backlog](https://backlog.com) space.

## What it does today (walking skeleton)

- Adds a **Backlog** card under **Settings > Integrations**, with the Backlog logo and an
  on/off switch.
- Adds a **Backlog** entry to the **Integrations** menu on the Kandev home page. It opens the
  `/backlog` page, which shows whether Backlog is on, the connection status of the current
  workspace, and a link to its Backlog settings.
- A workspace admin connects the workspace to one Backlog space with an **API key**.
  The plugin checks the key with Backlog (`GET /api/v2/users/myself`) before saving it,
  then shows `Connected as <name> @ <space>`.
- Only `https` space addresses on `backlog.com`, `backlog.jp` or `backlogtool.com` are accepted.
- The key is stored only in Kandev's encrypted secret store. It is never returned to the
  browser and never written to logs.

OAuth, project selection, issues and pull requests come in later releases.

## Turn Backlog on or off

Backlog is on by default in every workspace. A Kandev admin can turn it off for one workspace
with the switch on the Backlog card, then **Save**. While it is off, the plugin refuses to
connect or call Backlog for that workspace, but it keeps the existing connection: turning it
back on restores it without connecting again. The home entry and the `/backlog` page stay
available, so Backlog can always be turned back on.

The Backlog logo is used under Nulab's brand guidelines; see
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
3. Open **Settings > Integrations > Backlog**.

## Connect with an API key

1. In Backlog, open **Personal Settings > API** and create an API key.
2. In Kandev, enter your space address (for example `myteam.backlog.com`) and the key.
3. Click **Connect**. The page shows `Connected as <name> @ <space>` when Backlog accepts the key.

Connecting again with another key replaces the connection only after Backlog accepts the new key.

## License

MIT. See [LICENSE](LICENSE).
