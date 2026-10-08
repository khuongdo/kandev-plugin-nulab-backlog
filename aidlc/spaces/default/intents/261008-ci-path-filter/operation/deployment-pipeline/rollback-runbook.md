# Rollback Runbook — CI path filter

## Triggers

- A records-only pull request still cannot merge (a required check stays "Expected — waiting").
- An app change is wrongly classified as non-app (`changes` prints `app=false` while app files changed), so `checks` is skipped.
- `secret-scan` fails for reasons unrelated to credentials and blocks every pull request.

## Rollback Steps

1. **If a required check blocks merging** (`secret-scan` broken): remove `secret-scan` from ruleset 24580280, keeping `checks` and `packaged-host-contract`:

   ```bash
   gh api repos/khuongdo/kandev-plugin-nulab-backlog/rulesets/24580280 \
     --jq '{name, target, enforcement, conditions, bypass_actors, rules: [.rules[] | if .type == "required_status_checks" then .parameters.required_status_checks |= map(select(.context != "secret-scan")) else . end]} | with_entries(select(.value != null))' \
     > ruleset-24580280.json
   gh api -X PUT repos/khuongdo/kandev-plugin-nulab-backlog/rulesets/24580280 --input ruleset-24580280.json
   rm ruleset-24580280.json
   ```

2. **Revert the change**: open a pull request with `git revert <squash commit>` on `main`. The revert touches `.github/workflows/` and `internal/ci`, so it is classified as app and runs full CI; after merge, `ci.yml` runs every job on every change again.
3. Re-read the ruleset (`gh api …/rulesets/24580280 --jq '[.rules[] | select(.type == "required_status_checks") | .parameters.required_status_checks[].context]'`) and confirm it matches the workflows that exist on `main`.

## Notes

- No plugin release is involved, so nothing needs to be reinstalled on the self-hosted Kandev, and no tag is touched (released tags are never deleted or overwritten).
- A wrong classification is fixed forward by editing the single non-app list in `internal/ci/changes.go` and its table test.
