# CI Pipeline — Questions

Context: the CI pipeline was already built by the ci-release Unit (U5): `.github/workflows/ci.yml` (jobs `checks` and `packaged-host-contract`) and `.github/workflows/release.yml` (jobs `verify`, `contract`, `publish`). The last two `ci` runs on `main` passed. A read of the GitHub ruleset `main` (id 24580280) on 2026-10-07 shows only one rule, `deletion`: pull requests, required status checks and the force-push block are not enforced, which conflicts with the team rule that `main` is protected.

## Q1. CI tool

Which CI tool should the pipeline use?

A. GitHub Actions, keeping the existing `ci.yml` and `release.yml` (team Way of Working and Deployment)
B. Another tool
X. Other (please specify)

[Answer]: A

## Q2. Branch strategy and protection of `main`

The ruleset now only blocks branch deletion. What should be done?

A. Restore the full team protection on the `main` ruleset: require a pull request, require the status checks `checks` and `packaged-host-contract`, block force-push, keep the deletion block
B. Leave the ruleset as it is for now and record the gap as an open item
X. Other (please specify)

[Answer]: A

## Q3. Quality gates before merge

Which gates must pass before merge?

A. The existing gates: `go mod tidy` diff check, then `make check-format vet lint test coverage check-secrets build package verify-package` (job `checks`), then `make verify-package contract-test` on Kandev at `min_kandev_version` against the exact uploaded package (job `packaged-host-contract`)
B. Add more gates
X. Other (please specify)

[Answer]: A

## Q4. Artifact repositories

Where are build outputs stored?

A. GitHub Actions artifacts for CI (`plugin-package`, `release-package`) and GitHub Releases for published packages with `checksums.txt` and a build provenance attestation; no container or package registry
B. Add a registry
X. Other (please specify)

[Answer]: A

## Consolidated Summary Confirmation

Does this all look correct before I generate the artifact?

- Looks correct
- Request changes

[Answer]: Looks correct
