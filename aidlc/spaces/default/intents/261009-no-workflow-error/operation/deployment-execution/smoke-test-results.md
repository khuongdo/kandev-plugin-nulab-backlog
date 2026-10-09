# Smoke Test Results - v0.6.1

## Automated (release pipeline)

| Check | Result | Evidence |
|---|---|---|
| `release.yml` verify (format, vet, lint, test, package, package verification) | Pass | run 37923098232, job `verify` |
| Packaged-host contract test on Kandev 0.96.0 | Pass | run 37923098232, job `contract` |
| Released asset checksum | Pass | `sha256sum -c checksums.txt` |
| Released asset package verification | Pass | `verifypkg: OK ... (nulab-backlog@0.6.1)` |
| Build provenance attestation | Pass | `gh attestation verify` exit 0 |

## Functional (real Kandev v0.96.0)

| Check | Result | Evidence |
|---|---|---|
| Direct `/backlog` load → "+ Task" opens Kandev's dialog with a selectable workflow | Pass | user's manual check during Build and Test (same code, pre-version-bump build) |
| Created task links to the issue | Pass | same |
| Errors display as toast / full-width wrapping alert on mobile and desktop | Pass | same |

## Pending (user)

- Repeat the functional smoke check after installing the released 0.6.1 package on the self-hosted Kandev (see `operation/deployment-pipeline/deployment-strategy.md` § Success Criteria).
