# Deployment Strategy — release 0.5.1

## Strategy

Tagged release + manual install (team Deployment practice). A Kandev plugin is one package installed per server; there is no traffic to shift, so blue/green, canary and rolling do not apply.

- Version: **0.5.1** (patch, user's choice at Q1). The change adds an opt-in connection method and is backward compatible: existing token connections are read as method `token` with no migration.
- Approval: creating the tag `v0.5.1` on `main` is the production approval.
- Install: manual, by the self-hosted Kandev admin (From URL or upload of the Release package).

## Feature flags

None. The new method is opt-in per provider: nothing changes until an admin presses "Use gh CLI login" / "Use glab CLI login".

## Compatibility and risk

| Change | Risk | Mitigation |
|---|---|---|
| New admin action `scm.providers.use_cli` in `manifest.yaml` | Kandev older than 0.96.0 | `min_kandev_version` stays `0.96.0`; contract test on 0.96.0 passed 10/10 |
| GitLab token header `PRIVATE-TOKEN` → `Authorization: Bearer` | an existing GitLab token stops working | GitLab accepts personal access tokens as Bearer; covered by `internal/gitlab` tests; smoke check "Test" on an existing GitLab connection after install |
| `Settings.Source` field in plugin state | old state unreadable | field is optional (`omitempty`); v0.5.0 state reads as token (test `FR6.1`) |
| CLI runs on the server as the Kandev user | admins of any workspace on the server can use the server's CLI login | action is admin-only; documented in README |

## Security implications (operation phase rule)

No infrastructure, IAM, network or encryption change. The only new runtime behaviour is running `gh` / `glab` with fixed arguments, no shell, a 10-second timeout and no stored token (see `construction/build-and-test/security-test-instructions.md`).
