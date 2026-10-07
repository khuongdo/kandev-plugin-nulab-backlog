# Deployment Log — v0.1.0

## Summary

First release `v0.1.0` of the Nulab Backlog plugin, published on 2026-10-07 and installed by hand on the self-hosted Kandev. There is no automatic deployment target; the deployment path is the one in team Deployment and `construction/ci-pipeline/ci-config.md` (the Deployment Pipeline and Environment Provisioning stages were skipped for that reason). Answers: `deployment-execution-questions.md` [Q1]–[Q4] all A.

## Pre-deployment checks

| Check | Result |
|-------|--------|
| Build and Test approved (`construction/build-and-test/test-results.md`, accepted failure for real-space items) | done |
| CI Pipeline approved; `main` ruleset requires PR + `checks` + `packaged-host-contract`, force-push blocked | done |
| Database migrations | none (the plugin keeps its state in Kandev plugin storage) |
| Dependent services | Backlog (`khuongdo.backlog.com`) and the self-hosted Kandev, both reachable during the manual check |
| Deployment window | none needed (single maintainer, manual install) |

## Timeline (2026-10-07)

| Step | Evidence | Result |
|------|----------|--------|
| PR #1: Build and Test loop-backs, CI Pipeline records, `manifest.yaml` 0.1.0 | run 37558151237: `checks` pass (2m2s), `packaged-host-contract` pass (51s); squash-merged as `0545212` | done |
| Local package built from an identical tree; `verifypkg` OK | `dist/nulab-backlog-0.1.0.tar.gz` | done |
| Manual check against a real Backlog space, Kandev v0.97, steps 1–9 | `docs/manual-checks/2026-10-07-first-release.md` (Result pass; `make check-secrets` OK) | pass |
| PR #2: manual-check record | checks pass on `4a7ada6`; squash-merged as `cdaa7c4` | done |
| Tag | annotated `v0.1.0` on `cdaa7c4`, created after the maintainer's explicit confirmation | done |
| `release.yml` | run 37559126917: `verify` (incl. `verify-package` then `release-preflight`) success, `contract` success, `publish` success | done |
| GitHub Release | https://github.com/khuongdo/kandev-plugin-nulab-backlog/releases/tag/v0.1.0 — `nulab-backlog-0.1.0.tar.gz`, `checksums.txt`; not a prerelease | done |
| Download check | `sha256sum -c checksums.txt` OK; `gh attestation verify` OK, signer `release.yml@refs/tags/v0.1.0` | pass |
| Install on self-hosted Kandev | maintainer installed the Release asset through Settings > Plugins | done |
| Marketplace registry PR | entry and text drafted in the appendix below; postponed by the maintainer | open |

## Notes

- The Release package (sha256 `6d3fbe7c…837e`) differs in bytes from the local pre-release build (`e998aa7f…02dd7`) because builds are not byte-reproducible; the code is the same tree. The Release asset is the one installed.
- The manual check ran on Kandev v0.97; the minimum version v0.96.0 is covered by the CI contract test.

## Rollback

No in-place rollback (team Deployment): if `v0.1.0` is broken, mark the Release, reinstall the previous version on the self-hosted Kandev (none exists for the first release: uninstall the plugin instead), and fix forward with `v0.1.1`. The tag `v0.1.0` is never deleted or overwritten.

## Appendix: marketplace registry PR draft

Target: `kdlbs/kandev`, file `plugin-registry/plugins.yaml`. Fork the repo, add the entry at the end of `plugins:`, open a pull request.

`make marketplace-entry REGISTRY=../kandev/plugin-registry/plugins.yaml` (2026-10-07, against `../kandev` at v0.96.0) reports `listed=false` and prints:

```yaml
  - id: nulab-backlog
    repo: khuongdo/kandev-plugin-nulab-backlog
    categories: [integrations]
```

### PR title

Add nulab-backlog (Nulab Backlog integration) to the plugin registry

### PR body

Adds `nulab-backlog`, a Kandev plugin that connects a workspace to a Nulab Backlog space (API key or OAuth), creates and links Kandev tasks from Backlog issues, and supports Backlog Git repositories and pull requests.

- Repo: https://github.com/khuongdo/kandev-plugin-nulab-backlog (MIT)
- Release: https://github.com/khuongdo/kandev-plugin-nulab-backlog/releases/tag/v0.1.0 with `nulab-backlog-0.1.0.tar.gz` and `checksums.txt`
- Manifest `id`: `nulab-backlog`; `min_kandev_version`: `0.96.0`
- Build provenance: `gh attestation verify nulab-backlog-0.1.0.tar.gz -R khuongdo/kandev-plugin-nulab-backlog`
- Tested: packaged-host contract test on Kandev v0.96.0 in CI; manual check against a real Backlog space on Kandev v0.97
