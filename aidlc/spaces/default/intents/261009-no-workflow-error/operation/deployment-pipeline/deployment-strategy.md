# Deployment Strategy: v0.6.1

## Strategy

Tag-driven release with manual install, per the team Deployment practice: no blue/green or canary, because the plugin runs inside each self-hosted Kandev and is installed by an admin.

1. Open a PR from `feature/fix-kandev-has-no-wo-ik5` with the code change, the version bump to 0.6.1 and the README heading rename; CI must be green; self-merge with squash.
2. Re-check `gh release list` and `origin/main`; tag `v0.6.1` on the merge commit of `main`.
3. `release.yml` publishes the GitHub Release; put the README 0.6.1 section at the top of the generated notes (`gh release edit`).
4. Install the released package on the self-hosted Kandev; approve the new `workflows` read permission.
5. Send the marketplace registry PR.

## Approvals

- Creating the tag is the production approval (team rule).
- Package verification must pass before tagging (project Mandated rule) — enforced again by `release.yml`.

## Upgrade Impact

- Admins may be asked to re-approve the plugin's permissions because of `api_read: workflows`.
- No data migration; no state or secret format change.

## Success Criteria (smoke check after install)

- Open `/backlog` directly (reload) → "+ Task" opens Kandev's task dialog with a selectable workflow; the created task links to the issue.
- An error (e.g. connection off) shows as a toast or a full-width wrapping alert at phone and desktop widths.
