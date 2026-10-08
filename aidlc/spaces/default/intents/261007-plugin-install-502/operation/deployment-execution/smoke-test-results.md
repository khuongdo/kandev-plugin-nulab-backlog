# Smoke Test Results — release v0.4.2

## Release Artifact Checks (run)

| Check | Command | Result |
|---|---|---|
| Release workflow | `release.yml` run 37704397347 | verify, contract (Kandev 0.96.0 on the exact package bytes), publish: all success |
| Checksum | `sha256sum -c checksums.txt` on the downloaded assets | `nulab-backlog-0.4.2.tar.gz: OK` |
| Provenance | `gh attestation verify nulab-backlog-0.4.2.tar.gz -R khuongdo/kandev-plugin-nulab-backlog` | exit 0 (verified) |
| Package contents | `tar tzf nulab-backlog-0.4.2.tar.gz` | `manifest.yaml`, 4 `server/plugin-{linux,darwin}-{amd64,arm64}`, `ui/bundle.js`, `checksums.txt`; no Windows executable |
| Package size | release asset size | 23,337,444 bytes (NFR1 ≤ 25 MB: met) |
| From URL source | `curl -sL` the release asset URL | HTTP 200 after redirect, 23,337,444 bytes in 2.9 s |

## Self-Hosted Install Smoke Check (not run)

Not run in this stage: the human chose to install 0.4.2 on the self-hosted Kandev themselves (Q1 = B). The checklist to run after installing From URL is in `operation/deployment-pipeline/deployment-strategy.md` ("Smoke Check").
