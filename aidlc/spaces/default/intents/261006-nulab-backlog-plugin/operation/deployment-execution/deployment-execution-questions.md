# Deployment Execution — Questions

Context: there is no automatic deployment. "Deployment" for this plugin is: get the code onto `main` through a pull request, run the manual check against a real Backlog space, tag `vX.Y.Z` on `main` (the tag triggers `release.yml`, which publishes a GitHub Release with the package, `checksums.txt` and an attestation), install the package by hand on the self-hosted Kandev, and send the marketplace registry PR (team Deployment). `main` is at `4b0a090`; commit `74edd48` (Build and Test loop-backs) and the CI Pipeline records are local only. `manifest.yaml` has `version: "0.0.1"`. `docs/manual-checks/` holds only `TEMPLATE.md`, so `release-preflight` refuses a first release until a passing record exists. No database migrations and no dependent services other than Backlog and Kandev.

## Q1. Getting the local work onto `main`

How should `74edd48` and the CI Pipeline records reach `main`? (`main` now requires a pull request and the checks `checks` and `packaged-host-contract`.)

A. Create a short-lived branch, push it, open a pull request, and squash-merge it after both checks are green
B. Wait; do not push yet
X. Other (please specify)

[Answer]: A

## Q2. Version of the first release

Which version should the first release use?

A. `v0.1.0`: bump `manifest.yaml` to `0.1.0` in the same pull request
B. `v0.0.1`: keep the current manifest version
X. Other (please specify)

[Answer]: A

## Q3. Manual check against a real Backlog space

Who runs the manual check (walking skeleton and first release, steps in `docs/manual-checks/TEMPLATE.md`), and when?

A. The user runs it now on the self-hosted Kandev with a real Backlog space and reports the results; the record is written to `docs/manual-checks/<date>-first-release.md` without any key or token
B. Postpone: park the workflow here until the check is done
X. Other (please specify)

[Answer]: A

## Q4. Tag, release, installation and marketplace

What should happen after the check passes?

A. Tag `vX.Y.Z` on `main` only after the user explicitly confirms the tag; then watch `release.yml`, verify the attestation with `gh attestation verify`, the user installs the package on the self-hosted Kandev, and the marketplace registry PR is drafted for the user to send
B. Stop after the pull request is merged; release later
X. Other (please specify)

[Answer]: A

## Consolidated Summary Confirmation

Does this all look correct before I generate the artifact?

- Looks correct
- Request changes

[Answer]: Looks correct
