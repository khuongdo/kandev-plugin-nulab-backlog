# Smoke Test Results — release v0.5.1

## Executed (release artifact)

| Check | Result | Evidence |
|---|---|---|
| Package builds and verifies on `main` | pass | `verifypkg: OK dist/nulab-backlog-0.5.1.tar.gz (nulab-backlog@0.5.1)` |
| Packaged-host contract test on Kandev v0.96.0 (PR CI) | pass | job `packaged-host-contract` on #19 |
| Contract test in `release.yml` | pass | job `contract` of the `v0.5.1` release run |
| Release asset downloads and its provenance verifies | pass | `gh attestation verify` exit 0 |
| Release is not a pre-release and has both assets | pass | `nulab-backlog-0.5.1.tar.gz`, `checksums.txt` |

## Not executed (owner installs, scope Q1 = B)

Self-hosted Kandev checks after install:

- Settings > Integrations shows Nulab Backlog 0.5.1.
- Settings > Source control: GitHub / GitLab cards show "Use gh CLI login" / "Use glab CLI login"; existing token connections still "connected"; "Test" on an existing GitLab token connection succeeds (Bearer header change).
- With `gh auth login` done as the Kandev server user: "Use gh CLI login" → "Connected via gh CLI as <account>", "Test" succeeds.
