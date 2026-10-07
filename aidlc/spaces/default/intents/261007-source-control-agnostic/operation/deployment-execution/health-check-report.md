# Health Check Report — 261007-source-control-agnostic (v0.4.0)

## Release Pipeline Health

| Check | Result |
|---|---|
| Pull request CI (`checks`) | Pass, 2m17s |
| Pull request CI (`packaged-host-contract`, Kandev 0.96.0) | Pass, 1m1s |
| `release.yml` `verify` (full make chain + `release-preflight TAG=v0.4.0`) | Success |
| `release.yml` `contract` (release package on Kandev 0.96.0) | Success |
| `release.yml` `publish` (provenance attestation + GitHub Release) | Success |
| GitHub Release assets | `nulab-backlog-0.4.0.tar.gz`, `checksums.txt` |

## Local Re-verification Before Commit (version 0.4.0)

`make check-format test coverage package verify-package`: all OK; Go coverage 92.8% (floor 80%); `verifypkg: OK dist/nulab-backlog-0.4.0.tar.gz (nulab-backlog@0.4.0)`.

## Runtime Health (self-hosted Kandev)

Pending install by the maintainer (Q2 = A). Health signals to check after install: plugin process running with version `0.4.0`, no errors in the Kandev log at startup, and no `rate_limited` / `reconnect required` errors for providers that have no token (none are called until a token is added).

## Migrations

None. No existing data changes; new `scm.*` documents and secrets are created only when an admin configures a provider.
