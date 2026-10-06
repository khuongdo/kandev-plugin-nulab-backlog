# Code Summary — ci-release (U5)

Methodology: TDD (Testing Contract `sha256:376593af89ee0d9cddcd078826773d6d5ca3215ec737b9af952712bee2395bcc`).
Every plan step 1–11 was executed in order. Base commit `99760f8` on `main`; nothing was committed,
pushed or tagged.

## Files

Created:

- `internal/ci/doc.go`: package doc.
- `internal/ci/manifest.go`, `manifest_test.go`: `Manifest`, `ReadManifest`.
- `internal/ci/release.go`, `release_test.go`: `Tag`, `ParseTag`, `Facts`, `CheckRelease`, `FirstReleaseRecorded`.
- `internal/ci/secrets.go`, `secrets_test.go`: `ScanSecrets`, `Finding` and the three rules.
- `internal/ci/marketplace.go`, `marketplace_test.go`: `MarketplaceEntry`, `CheckRegistry`.
- `internal/ci/workflows.go`, `workflows_test.go`: `CheckWorkflows`, plus the real-repository policy test.
- `internal/ci/contract.go`, `contract_test.go`: `Config`, `RunContract` and a fake Kandev built with `httptest`.
- `internal/ci/run.go`, `run_test.go`: `Run` with the subcommands `secrets`, `preflight`, `marketplace`, `workflows` and `contract`.
- `cmd/ci/main.go`: one-line wrapper, like `cmd/verifypkg`.
- `.github/workflows/release.yml`.

Modified:

- `Makefile`: new variables `MIN_KANDEV_VERSION` (read from `manifest.yaml`), `ACTIONLINT_VERSION`,
  `KANDEV_MIN_DIR`, `CONTRACT_PORT`, `CONTRACT_AGENTCTL_PORT`, `MARKETPLACE_REPO`, `REGISTRY`.
  `lint` now also runs actionlint and the workflow policy. New targets: `check-secrets`,
  `contract-test`, `release-preflight`, `marketplace-entry`. `help` and `.PHONY` are updated. No
  existing C4 target was renamed or removed, and `COVERAGE_MIN` and `COVERAGE_EXCLUDE` are unchanged.
- `.github/workflows/ci.yml`: `checks` runs `check-secrets` after `coverage`. There is a new job,
  `packaged-host-contract`.
- `README.md`: three new sections, "CI checks", "Releasing" and "Marketplace".

The step 2 smoke test `internal/ci/smoke_test.go` was created, then deleted at the first Red step
(the same pattern as U1).

No application code was touched under `internal/connection`, `internal/backlog`, `internal/plugin`,
`internal/redact`, `internal/testutil` or `ui/src`. No module dependency was added: `go mod tidy`
leaves no diff.

## Key Decisions

- **Secret scan heuristic**: the plan's rule is "32 or more `[A-Za-z0-9]` characters with upper
  case, lower case and a digit". On the real tree it flagged the existing U1 test name
  `TestIntegrationDisabledMapsTo409` (`internal/plugin/actions_test.go:447`). The plan lists long
  CamelCase test names as non-findings, and `internal/plugin` must not be changed. The rule now
  also skips identifier-shaped runs: letters followed only by trailing digits
  (`^[A-Za-z]+[0-9]*$`). Generated keys mix digits into the letters, so they are still found. The
  bait test and a planted random key both prove this. A new row in the non-findings table covers
  the test-name case.
- **Workflow policy**: it flags **any** `write` permission (and `write-all`) outside `publish` in
  `release.yml`, not only `id-token` and `attestations`. This matches the brief ("only the release
  `publish` job has `contents: write`, `id-token: write`, `attestations: write`") and the plan's
  assumption that `publish` is the only job with any write permission.
- **Marketplace entry**: it is printed with the registry's two-space list indentation, so it can be
  pasted under `plugins:`.
- **Contract polling**: there is no clock. `poll` makes `timeout / PollInterval` attempts (at least
  one) and calls the injected `Wait` between attempts, so tests never sleep.
- **No echo of the bait key**: for the action steps, the driver reports only the HTTP status,
  `code` and `field`, and replaces the bait key in the message with `[redacted]`. Response bodies
  are never echoed. For install failures, Kandev's `error` text is reported (it carries no key).
  Every response read is capped at 1 MiB with `io.LimitReader`, and `ctx` errors are returned
  unchanged.
- **Release**: the `gh` CLI is used instead of `softprops/action-gh-release`. A tag with a
  `-suffix` gets `--prerelease`. All `${{ }}` values reach `run:` scripts through `env:`.

## Red Evidence (failing runs before Green)

Step 3, data model:

```
$ go test -race ./internal/ci/... -run 'TestReadManifest|TestParseTag'
internal/ci/manifest_test.go:20:12: undefined: ReadManifest
internal/ci/manifest_test.go:22:19: undefined: Manifest
internal/ci/release_test.go:28:16: undefined: ParseTag
internal/ci/release_test.go:34:21: undefined: Tag
FAIL	github.com/khuongdo/kandev-plugin-nulab-backlog/internal/ci [build failed]
```

Step 5, business logic:

```
$ go test -race ./internal/ci/... -run 'TestScanSecrets|TestCheckRelease|TestFirstRelease|TestMarketplace|TestCheckRegistry|TestCheckWorkflows'
internal/ci/release_test.go:47:31: undefined: Facts
internal/ci/release_test.go:52:21: undefined: CheckRelease
internal/ci/marketplace_test.go:14:14: undefined: MarketplaceEntry
internal/ci/marketplace_test.go:46:19: undefined: CheckRegistry
FAIL	github.com/khuongdo/kandev-plugin-nulab-backlog/internal/ci [build failed]
(after ScanSecrets, CheckRelease and the marketplace functions were green:)
internal/ci/workflows_test.go:50:21: undefined: CheckWorkflows
FAIL	github.com/khuongdo/kandev-plugin-nulab-backlog/internal/ci [build failed]
(first full run, the repository regression guard:)
--- FAIL: TestScanSecretsFindsNothingInTheRepository
    Error: Should be empty, but was [internal/plugin/actions_test.go:447: credential-shaped-token]
```

Step 7, API / endpoint:

```
$ go test -race ./internal/ci/... -run 'TestRunContract|TestRun'
internal/ci/contract_test.go:109:46: undefined: Config
internal/ci/contract_test.go:127:21: undefined: RunContract
FAIL	github.com/khuongdo/kandev-plugin-nulab-backlog/internal/ci [build failed]
```

Step 9, environment and build configuration:

```
$ go test -race ./internal/ci/... -run TestRepositoryWorkflowsFollowThePolicy
--- FAIL: TestRepositoryWorkflowsFollowThePolicy (0.00s)
    Error: Received unexpected error:
           open ../../.github/workflows/release.yml: no such file or directory
    Messages: release.yml must exist
FAIL
```

Refactor, while green: golangci-lint (gosec) reported three issues. Two G101 false positives on
the rule-name constants were annotated with `//nolint:gosec // G101: ...`. One G120 in the fake
Kandev was fixed for real: the fake now reads the upload as a stream, with `MultipartReader` and a
1 MiB `LimitReader`. The result is `0 issues.` No lint rule was disabled.

## Test and Coverage Results

`make clean && make check-format vet lint test coverage check-secrets build package verify-package`
passes (exit 0):

- golangci-lint: `0 issues.`
- actionlint v1.7.12: clean. It was also run with shellcheck v0.11.0 (as on `ubuntu-latest`), and is
  clean there too.
- Workflow policy: `ci workflows: OK`.
- Go tests with `-race`: every package passes. Vitest: 65 passed.
- `internal/ci`: 38 test functions (93 runs including subtests). Package coverage: 90.7%.
- `make coverage`: `coverage: 93.2% (floor 80%, excluded: server/main.go)`.
- `ci secrets: OK`. Note that `coverage.out` and `build/coverage.filtered.out` were scanned.
- `verifypkg: OK dist/nulab-backlog-0.0.1.tar.gz (nulab-backlog@0.0.1)`.
- `go mod tidy`: no diff in `go.mod` or `go.sum`.

## Contract Test Result (real Kandev v0.96.0)

To keep build outputs out of the SDK checkout `../kandev` (the plan's Open note), a separate
worktree was created: `git -C ../kandev worktree add ../kandev-min v0.96.0`. It is outside this
repository, at commit `f099a46dc7aab16f6ff5806cd29b2b480296303f`, the same commit as
`.kandev-sdk-ref`.

1. First run, `make contract-test KANDEV_MIN_DIR=../kandev-min`: **failed**, as the plan
   anticipated. The backend does not start without its agentctl sidecar
   (`Failed to start agentctl subprocess ... exec: "agentctl": executable file not found`). The
   driver then failed closed with `contract: host not ready within 3m0s`.
2. Fix, without weakening any assertion:
   - `contract-test` now builds `build-kandev build-agentctl`. Kandev finds `bin/agentctl` next to
     its own binary.
   - The sidecar gets its own port, `AGENTCTL_PORT=$(CONTRACT_AGENTCTL_PORT)` (39529). This
     machine's own Kandev uses 38429 and 39429, and Kandev tries to adopt an agentctl already on
     its port.
3. Second run: **pass**, in about 63 s including the build:
   `ci contract: OK nulab-backlog on Kandev v0.96.0`. The run covered `/ready` `ok` at exactly
   `v0.96.0`, a multipart install that returned 201 with no warning, status `active`, the first
   workspace, `connection.get` returning `not_connected` with `enabled: true`, and
   `connection.connect_api_key` returning 400 `validation` on `spaceUrl`. The temporary home was
   removed. The host's own Kandev (38429 and 39429) was not touched, and `../kandev-min` has no
   tracked changes (`bin/` is ignored).

## Negative Evidence (no committed threshold changed)

| Check | Command | Result |
|-------|---------|--------|
| Coverage gate fails closed (AC7.3.1) | `make coverage COVERAGE_MIN=100` | `coverage: below the 100% floor`, make Error 1 |
| Planted key found, not printed (AC7.3.4) | `go run ./cmd/ci secrets -root <scratchpad tree with a run-time random 33-char key>` | exit 1, `internal/x/testdata/a.json:1: credential-shaped-token`; `grep` of the output for the key finds 0 matches |
| Not on main (AC7.5.2) | `preflight ... -on-main=false` | exit 1, `tag v0.1.0 points at a commit that is not on main` |
| Release exists (AC7.5.2) | `preflight ... -released=true` (flag replaced by `-releases` in R-01; see Review fixes) | exit 1, `tag v0.1.0 already has a GitHub Release; a released tag is never overwritten` |
| Version mismatch | `preflight -tag v0.1.1 -version 0.1.0 ...` | exit 1, `tag v0.1.1 does not match the manifest version "0.1.0"` |
| First release without a record (AC7.5.3) | `preflight ... -manual-checks <empty dir>` | exit 1, `... is the first release and ... has no *-first-release.md record with Result pass` |
| First release with a `fail` record (AC7.5.3) | `preflight ... -manual-checks <dir with Result fail>` | exit 1, same refusal |
| Wrong catalogue id (AC7.6.2) | `marketplace -registry <copy of plugins.yaml plus id: backlog for this repo>` | exit 1, `registry id "backlog" for khuongdo/kandev-plugin-nulab-backlog differs from the manifest id "nulab-backlog"` |
| Current catalogue | `make marketplace-entry REGISTRY=../kandev/plugin-registry/plugins.yaml` | exit 0, `listed=false`, the entry is printed |

## Pinned Actions and Tools

Each SHA was resolved from the official repository's release tag with `gh api` (annotated tags
dereferenced to commits) on 2026-10-06:

| Action | SHA | Tag |
|--------|-----|-----|
| `actions/checkout` | `3d3c42e5aac5ba805825da76410c181273ba90b1` | v7.0.1 (unchanged from U1, re-verified) |
| `actions/setup-go` | `b7ad1dad31e06c5925ef5d2fc7ad053ef454303e` | v7.0.0 (re-verified) |
| `actions/setup-node` | `820762786026740c76f36085b0efc47a31fe5020` | v7.0.0 (re-verified) |
| `actions/upload-artifact` | `043fb46d1a93c77aae656e7c1c64a875d1fc6a0a` | v7.0.1 (re-verified) |
| `actions/download-artifact` | `3e5f45b2cfb9172054b4087a40e8e0b5a5461e7c` | v8.0.1 (latest; it enforces digest checks by default) |
| `actions/attest-build-provenance` | `4d101475d8b20a2381f78447822ac1eab6504dd8` | v4.2.2 (latest; a wrapper over `actions/attest`) |

Tools:

- actionlint: `github.com/rhysd/actionlint/cmd/actionlint@v1.7.12` (latest release), run through
  `go run`.
- golangci-lint: still v2.14.0.

## Deviations and Open Items

- **Contract-test build target**: `build-agentctl` and a separate agentctl port were added, as the
  plan's assumption allowed.
- **Step 10 used `KANDEV_MIN_DIR=../kandev-min`** instead of `../kandev`, following the plan's Open
  note. The worktree remains in place for later runs. Remove it with
  `git -C ../kandev worktree remove ../kandev-min`.
- **Secret-scan rule refinement**: identifier-shaped runs are skipped (see Key Decisions). It needs
  a reviewer's acknowledgement.
- **Dead host is not detected early**: if the throwaway backend exits early, the driver still waits
  for the full ready limit (3 min) before it fails. The result is correct; it is only slow.
- **Not proven before merge (Open)**: the release workflow, the provenance attestation and
  `gh attestation verify` (AC7.5.1) run only on GitHub after a tag is pushed. The user still has to
  decide whether to use a pre-release tag or wait for `v0.1.0`. Branch protection must require both
  `checks` and `packaged-host-contract` (X4); the user sets this in GitHub. The marketplace pull
  request is sent after `v0.1.0` (AC7.6.1).
- **Different Go cache keys in CI**: `packaged-host-contract` caches on two `go.sum` files and
  `checks` on one.

## Review fixes (R-01..R-03)

Fixes for the findings in `reviews/review-01.md`, done TDD (failing test first, minimal fix,
refactor while green). The plan and `unit-test-instructions.md` are unchanged; where they name
`ReleaseExists`/`ExistingTags`/`-released`/`-tags`, read them as replaced by `Releases`/`-releases`.

### R-01 (Major): first release comes from GitHub Releases, not git tags

- **Red** (behavioural, against the old API; temporary test removed after the run):
  `Facts{Tag: "v0.1.1", ManifestVersion: "0.1.1", OnMain: true, ExistingTags: ["v0.1.0","v0.1.1"], ManualChecksDir: <empty>}`
  ```
  --- FAIL: TestR01RedRefusedThenRetaggedStillRequiresRecord (0.00s)
      Error:      An error is expected but got nil.
  FAIL	.../internal/ci	0.010s
  ```
  Then the new tests against the old code (compile Red):
  ```
  internal/ci/release_test.go:48:70: unknown field Releases in struct literal of type Facts
  internal/ci/release_test.go:48:82: undefined: Release
  FAIL	.../internal/ci [build failed]
  ```
- **Change**: `internal/ci/release.go` drops `Facts.ReleaseExists` and `Facts.ExistingTags` for
  `Facts.Releases []Release` (`tagName`, `isDraft`, `isPrerelease`, as `gh release list --json`
  prints them). `CheckRelease` stays pure: "already released" is any Release with the tag (drafts
  included); "first release" is "no other Release that is neither a draft nor a pre-release". A
  stable tag the preflight refused has no Release, so it never removes the gate. `Makefile`
  `release-preflight` now passes `-releases "$(gh release list --limit 1000 --json tagName,isDraft,isPrerelease)"`
  and no longer reads `git tag -l`. README step 2 wording follows.
- **Tests** (`internal/ci/release_test.go`): `TestCheckReleaseStillRequiresTheRecordWithoutAPublishedStableRelease`
  (refused v0.1.0 then retagged v0.1.1; only drafts and pre-releases) still require the record;
  `TestCheckReleaseSkipsTheRecordForLaterAndPreReleases` passes with a published stable `v0.1.0`
  Release; refusal table adds "draft release already exists".

### R-02 (Minor): the contract job tests the shipped bytes

- No Go test applies (workflow wiring); the checks are the workflow policy check, actionlint and a
  local run of the new job command.
- **Change**: `release.yml` `contract` now `needs: verify`, downloads the `release-package` artifact
  into `plugin/dist` (`actions/download-artifact@3e5f45b2… # v8.0.1`, the SHA already used by
  `publish`), checks `sha256sum -c checksums.txt`, and runs `make verify-package contract-test`
  with no `make package`. The Node.js setup and `npm ci` steps were removed from the job (no UI
  build). `publish` still needs both jobs. No Makefile change was needed: `contract-test` already
  installs `$(PACKAGE)` without building it.
- `ci.yml` got the same change (`packaged-host-contract` `needs: checks` and downloads
  `plugin-package`): it was cheap, and the PR check still installs the full packaged plugin on the
  minimum Kandev, now the exact bytes `checks` built. The cost is that the two jobs run one after
  the other instead of in parallel. If `checks` fails, `packaged-host-contract` is skipped; the pull
  request stays blocked by `checks`.
- No new action and no permission change; top-level `permissions: contents: read` is unchanged.

### R-03 (Minor): preflight fails closed when the Release list is unknown

- **Red**: compile Red as above (`-releases` and `ParseReleases` did not exist); with the old
  Makefile any `gh release view` error became `-released=false`.
- **Change**: `ParseReleases` accepts only a JSON array; empty output, `null`, an object or an
  error text returns `could not list GitHub Releases; refusing to release without knowing which exist`.
  `CheckRelease` also refuses `Releases == nil`. The Makefile turns a failed `gh` call into an empty
  `-releases` (`|| releases=""`, gh's own error stays visible in the log), so the CLI refuses with
  exit 1 and one line.
- **Tests**: `TestParseReleases` (valid list, `[]` is known-empty, four unknown inputs);
  `TestRunPreflightRefusesWhenTheReleaseListIsUnknown` (no flag, empty, error text: exit 1, one
  line, no stdout); `TestRunPreflight` (`-releases '[]'` OK, existing Release refused); refusal
  table row "release list unknown".

### Results

| Check | Result |
|-------|--------|
| `make clean` then `make check-format vet lint test coverage check-secrets build package verify-package` | exit 0; actionlint and `ci workflows: OK`; `internal/ci` 91.0%; total `coverage: 93.3% (floor 80%, excluded: server/main.go)`; `ci secrets: OK`; `verifypkg: OK dist/nulab-backlog-0.0.1.tar.gz` |
| New contract job command on the built `dist/` (no rebuild): `sha256sum -c checksums.txt` then `make verify-package contract-test KANDEV_MIN_DIR=../kandev-min` | `nulab-backlog-0.0.1.tar.gz: OK`; `ci contract: OK nulab-backlog on Kandev v0.96.0`, exit 0 |
| `gh release list --limit 1000 --json tagName,isDraft,isPrerelease` on this repo | `[]`, exit 0 (no Release yet, so the first-release gate applies) |
| R-03 end to end: the Makefile's command against a repository that does not exist | gh: `Could not resolve to a Repository ...`; `ci preflight: could not list GitHub Releases; refusing to release without knowing which exist`, exit 1 |

- `gh release list` sees drafts only with push access; the release job's read-only token sees
  published Releases. A draft for the same tag is still refused by `gh release create` in
  `publish`. `--limit 1000` is a known ceiling (marked `ponytail:` in the Makefile).

