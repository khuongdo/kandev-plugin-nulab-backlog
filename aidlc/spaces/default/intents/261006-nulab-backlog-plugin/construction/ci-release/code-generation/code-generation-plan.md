# Code Generation Plan — ci-release (U5)

Inputs:

- `unit-of-work` (U5 boundary), `unit-of-work-story-map`, `unit-of-work-dependency`.
- `stories` (US7.3, US7.4, US7.5, US7.6).
- `requirements` (FR7.1, FR7.2, FR7.3, NFR3, NFR4, NFR6, NFR8).
- `contract-summary`: C4 (Makefile targets) and C8 (Kandev server). U5 has no runtime contract.
- `bolt-plan` (B2 Definition of Done, Release Milestone), `external-dependency-map` (X4, X8), `risk-and-sequencing-rationale` (R-E, B2 merges before B3).
- `team-practices` and memory `team.md` (Way of Working, Testing Posture, Deployment, Code Style) and `project.md` (Forbidden, Mandated).
- U1 `cicd-pipeline.md` and `infrastructure-specification.md`; U1 `code-summary.md`.
- The current repository: `.github/workflows/ci.yml`, `Makefile`, `manifest.yaml`, `cmd/verifypkg`, `internal/pkgverify`, `internal/testutil`.

U5 has no unit-level functional, NFR or infrastructure design (user decision). Behaviour comes from the inputs above. Every design decision taken here is listed as `[assumption]` at the end.

Stories in U5: US7.3 (Must), US7.4 (Should), US7.5 (Should), US7.6 (Should).

## Facts Used by This Plan

**Kandev v0.96.0** (`../kandev`, the same commit as `.kandev-sdk-ref`):

- **Server binary**:
  - `apps/backend/cmd/kandev/main.go`: `kandev __backend` runs the backend directly, without the launcher.
  - `make -C apps/backend build-kandev` builds `apps/backend/bin/kandev`. It uses CGO and `-tags fts5`.
  - The Makefile stamps `main.Version` from `VERSION ?= $(git describe --tags …)`.
- **Minimum-version check**: `internal/plugins/manifest/semver.go` `CheckMinimumKandevVersion` skips the check when the running version is `dev` or is not a release version. The throwaway server must therefore be built with `VERSION=v<min_kandev_version>`, so the host enforces the minimum for real.
- **Configuration** (`docs/public/configuration.md`): `KANDEV_HOME_DIR` is the data root; `KANDEV_SERVER_HOST` sets the listen host; `KANDEV_SERVER_PORT` sets the port (default `38429`).
- **Authentication**: disabled in every shipped profile (`docs/public/authentication.md`). The middleware then injects a synthetic single-user admin (`internal/auth/authn/identity.go`), so a throwaway server needs no credentials to install a plugin or call an admin action.
- **Install** (`internal/plugins/handlers.go`): `POST /api/plugins/install` takes a multipart upload in the field `package`. Success is `201 {"plugin": <record>, "warning"?: "…"}`. The same version again gives 409. A package error gives 400. `docs/public/plugins-authoring.md` ("Disposable-instance smoke test") documents `curl -F "package=@<file>" http://localhost:38429/api/plugins/install`.
- **Plugin status**: `GET /api/plugins/:id` returns the record with `status`. `active` is `store.StatusActive` (`internal/plugins/store/store.go`).
- **Action relay**: `POST /api/plugins/:id/actions/:key` takes the envelope `{workspaceId, taskId, sessionId, repositoryId, body}` (`internal/plugins/action_handlers.go`). Kandev checks access and the workspace scope, then passes the plugin's status and body through unchanged.
- **Default workspace and liveness**: a "Default Workspace" is created on first start (`internal/task/repository/sqlite/defaults.go`). `GET /api/v1/workspaces` returns `{"workspaces": [{"id": …}], "total": n}`. `/health` and `/ready` both carry the build `version`; `/ready` reports `status: "ok"` once the server is ready (`internal/backendapp/health_test.go`).
- **No upstream verifier**: `cmd/plugin-package-verify` does not exist at v0.96.0. U1's `cmd/verifypkg` and `internal/pkgverify` replace it.
- **Signing**: Kandev checks the package's internal `checksums.txt` but does not check signatures; installed packages show as unsigned (`docs/decisions/2026-08-01-bitbucket-initial-release-remains-unsigned.md`). The GitHub build provenance attestation is verified with `gh attestation verify`, not by Kandev.

**Kandev marketplace registry** (`../kandev/plugin-registry/` at v0.96.0):

- `plugins.yaml` is a pointer list of `{id, repo, categories?}` entries. `featured` is for maintainers only.
- `schema.json` sets the patterns: `id` `^[a-z0-9][a-z0-9-]*$`, `repo` `^[^/]+/[^/]+$`.
- The `id` MUST equal the manifest `id` of the repository's latest release.
- The index build needs a release asset named `<id>-<version>.tar.gz`. A release-level `checksums.txt` is optional.
- To submit, fork `kdlbs/kandev`, add one entry, and open a pull request. The index workflow validates the entry, and a maintainer merges it.
- The Bitbucket entry uses `categories: [integrations]`.

**Bitbucket plugin workflows** (`github.com/kdlbs/kandev-plugin-bitbucket`, branch `main`, fetched 2026-10-06):

- Three workflows: `build.yml`, `ci.yml` and `release.yml`. Every action is pinned by SHA, and the default is `permissions: contents: read`.
- `ci.yml` `packaged-host-contract` job: reads `min_kandev_version` from `manifest.yaml` with `sed` and checks its format; checks out `kdlbs/kandev` at `v<min>` into `kandev-min/`; runs `make build-backend build-web-e2e`, then a Playwright spec that lives in the Kandev repo (`tests/plugins/bitbucket-packaged-plugin.spec.ts`). That spec cannot be reused for this plugin.
- `release.yml`: triggered by a `v*` tag, or by a `workflow_dispatch` prepare job that creates the tag. The publish job has `contents: write` and uses `softprops/action-gh-release`. The release `checksums.txt` is the archive's internal checksums plus the tarball's SHA-256. There is no provenance attestation; this plugin adds one (team Deployment rule).

**This repository**:

- `Makefile`: `VERSION` is parsed from `manifest.yaml` with `sed`; `GO_PKGS := ./internal/... ./server/...`; only `server/main.go` is excluded from coverage; golangci-lint runs through `go run …@$(GOLANGCI_LINT_VERSION)`, so it adds no repository dependency.
- `ci.yml` has one job, `checks`. `gopkg.in/yaml.v3` is already a direct dependency. No `actionlint` binary is installed locally.
- `internal/testutil` builds test keys at run time (`test-api-key-` plus 32 hex characters), so they never appear in committed files.
- Existing literals that a secret scanner must not flag: `"k-123456"` in `internal/plugin/actions_test.go`; `"test-api-key-0000"` in `ui/src/settings/settings.test.tsx`; `apiKey=abc123` in `internal/redact/redact_test.go`; long CamelCase test names.

## Testing Contract

```json
{
  "version": 1,
  "methodology": "tdd",
  "source": "team",
  "ordering": "For each testable layer, we write a failing test first, then write the minimum code to make it pass, then refactor while the test stays green, before moving to the next behaviour or layer.",
  "scope": "feature",
  "test_strategy": "standard",
  "project_type": "greenfield",
  "applicable_notes": [
    {
      "layer": "org",
      "text": "We treat tests as a first-class deliverable in every Bolt. The specific\nmethodology (TDD, BDD, ATDD, or classic test-after) is affirmed at\npractices-discovery and recorded in `team.md` under this heading with explicit\n`Methodology` and `Ordering` fields; Code Generation resolves those fields\nindependently from coverage, tooling, and scope notes.\n\nWhen no posture has been affirmed, our default per scope is:\n- **Methodology**: test-after\n- **Ordering**: implement each applicable testable layer, then write and run\n  that layer's tests.\n- `mvp`, `enterprise`, `feature`, `infra`, `classic` add an 80% line-coverage\n  floor and CI execution before merge.\n- `bugfix`, `security-patch` add a targeted regression for the specific\n  bug/vulnerability and require the existing suite to remain green.\n- `express` uses the Minimal strategy: requirement-driven unit tests (one per\n  requirement, with a happy-path floor per component); existing tests remain\n  green.\n- `poc`, `refactor`, `workshop` add no extra new-test floor and require the\n  existing suite to remain green.\n\nThe active `Test Strategy` still applies in every scope and determines test\nvolume/types. Scope floors are additive; they never reduce or replace the\nselected strategy.\n\nBuild and Test verifies defined coverage floors and affirmed quality targets;\nthey may not be weakened to make a step pass.\n\nAffirm a stricter posture in `team.md` if the team commits to one."
    },
    {
      "layer": "team",
      "text": "- We treat tests as a deliverable of every Bolt, not extra work.\n- **Methodology**: tdd\n- **Ordering**: For each testable layer, we write a failing test first, then write the minimum code to make it pass, then refactor while the test stays green, before moving to the next behaviour or layer.\n- A floor of **80% line coverage for the Go code**, measured with `go test -coverprofile` and enforced by a `make coverage` target; CI blocks the merge when it is lower. The measured scope is `./internal/...` and `./server/...`; only the minimal wiring in `main` is excluded, and every exclusion must be listed explicitly in the `Makefile`. We do not lower the floor or add exclusions to make it pass.\n- Go tests always run with **`go test -race`** (needs CGO; works on the `ubuntu-latest` runner). This is an addition to the Bitbucket template, because token refresh and API rate limiting are two places prone to data races.\n- **Automated contract test with the real package on the minimum Kandev version**, like the template's `packaged-host-contract` job: CI checks out Kandev at exactly the `min_kandev_version` in `manifest.yaml`, builds a throwaway server, installs the packaged plugin, and runs it.\n- **Manual end-to-end check against a real Backlog space exactly twice**: when the walking skeleton is done, and before the first release. Each result is recorded. After the first release, we rely only on automated tests.\n- Unit and integration tests do not call real Backlog: they use a **fake Backlog server built with `net/http/httptest`**, with JSON sample data in `internal/<package>/testdata/`, including the 401, 429 (with `Retry-After`), and expired-token error cases.\n- Tools: `testing` + `github.com/stretchr/testify/require`, table-driven tests with `t.Run`, test names that describe behaviour. The clock and wait functions are injected into the code so rate limits and token expiry can be tested without a real `time.Sleep`. Tests are independent of each other and use `t.TempDir()` and `t.Setenv()`. The TypeScript UI (if any) is tested with Vitest.\n- There is a test asserting that API keys and tokens do not appear in logs, error messages, or responses sent to the UI.\n- We **do not add a dependency vulnerability scan** (`govulncheck`, `npm audit`, Dependabot) — this is your decision in Q4."
    }
  ],
  "obligations": {
    "strategy": "standard",
    "strategy_volume": [
      "Five to eight tests per component.",
      "Unit tests plus integration tests for key boundaries.",
      "Add E2E, performance, or security tests when requirements demand them."
    ],
    "scope_floor": [
      "Meet an 80% line-coverage floor.",
      "Run the selected tests in CI before merge."
    ],
    "combination_rule": "Apply every selected-strategy obligation and every scope-floor obligation; neither replaces the other, and a targeted scope regression may add the narrowest necessary test type beyond the strategy default."
  },
  "plan_profile": {
    "methodology": "tdd",
    "runner_step": "Bootstrap the minimal test runner/configuration and record the exact unit-scoped command.",
    "runner_ready_before_first_test": true,
    "testable_layers": [
      "Data model / database behavior",
      "Repository / data access",
      "Business logic",
      "API / endpoint",
      "Frontend behavior"
    ],
    "steps": [
      "Project structure and production configuration skeleton.",
      "Bootstrap the minimal test runner/configuration and record the exact unit-scoped command.",
      "Data model / database behavior - Red: write the failing tests and record the failing command output.",
      "Data model / database behavior - Green: implement only enough behavior to pass.",
      "Data model / database behavior - Refactor: improve the implementation while tests stay green.",
      "Repository / data access - Red: write the failing tests and record the failing command output.",
      "Repository / data access - Green: implement only enough behavior to pass.",
      "Repository / data access - Refactor: improve the implementation while tests stay green.",
      "Business logic - Red: write the failing tests and record the failing command output.",
      "Business logic - Green: implement only enough behavior to pass.",
      "Business logic - Refactor: improve the implementation while tests stay green.",
      "API / endpoint - Red: write the failing tests and record the failing command output.",
      "API / endpoint - Green: implement only enough behavior to pass.",
      "API / endpoint - Refactor: improve the implementation while tests stay green.",
      "Frontend behavior - Red: write the failing tests and record the failing command output.",
      "Frontend behavior - Green: implement only enough behavior to pass.",
      "Frontend behavior - Refactor: improve the implementation while tests stay green.",
      "Environment/build configuration.",
      "Documentation and traceability."
    ]
  },
  "input_sha256": "sha256:c08268ec87b4a805bbc352db69f8ea8730f7297e5e63201ea588835bf2850378",
  "contract_sha256": "sha256:376593af89ee0d9cddcd078826773d6d5ca3215ec737b9af952712bee2395bcc"
}
```

## Plan Steps

Each TDD layer follows Red, then Green and Refactor. Each Red step records its failing command output in `code-summary.md`. All new Go logic lives in one package, `internal/ci`, so it counts toward the 80% coverage floor. `cmd/ci/main.go` is a one-line wrapper, like `cmd/verifypkg`. No application code under `internal/connection`, `internal/backlog`, `internal/plugin`, `internal/redact`, `internal/testutil` or `ui/src` is touched.

Layers that do not apply: Repository / data access (no storage) and Frontend behaviour (U5 has no UI).

### Step 1 — Project structure and configuration skeleton

- [x] Create `internal/ci/doc.go` with a package doc comment: CI and release checks for the plugin (secret scan, release preflight, marketplace entry, workflow policy, packaged-host contract driver).
- [x] Confirm that no new module dependency is needed: `yaml.v3`, `testify`, `net/http`, `mime/multipart` and `regexp` are enough.
- Stories: US7.3. Rules: team Code Style (Layout, Dependencies).

### Step 2 — Test runner readiness

- [x] The Go runner already exists from U1. Add a smoke test `internal/ci/smoke_test.go` so that `go test -race ./internal/ci/...` runs. Delete it at the first Red step (Step 3), as U1 did.
- [x] Confirm that these commands run, and record them in `unit-test-instructions.md` (already written there): `go test -race ./internal/ci/...` and `go test -race ./internal/pkgverify/...`.
- Stories: US7.3.

### Step 3 — Data model, Red: manifest facts and release tags

- [x] `internal/ci/manifest_test.go`:
  - `ReadManifest` returns `ID`, `Version` and `MinKandevVersion` from a manifest written to `t.TempDir()`;
  - it fails on a missing `id`, `version` or `min_kandev_version`;
  - it fails on a `min_kandev_version` that is not a numeric `X.Y.Z` (Kandev's `NormalizeReleaseVersion` rule);
  - it reads the repository's own `manifest.yaml` and gets `nulab-backlog` and `0.96.0` (US7.4: the version comes from the manifest, not from code).
- [x] `internal/ci/release_test.go` (tag part), a table-driven test of `ParseTag`: accepts `v0.1.0`, `v1.20.3` and `v0.0.1-rc.1` (pre-release); rejects `0.1.0`, `v1.2`, `v01.2.3`, `v1.2.3+build`, `v1.2.3 ` and `refs/tags/v1.2.3`; returns `Version` without the `v`, and `Prerelease`.
- [x] Run the tests and record the failing output.
- Stories: US7.4 (AC7.4.1), US7.5 (AC7.5.2). Rules: team Deployment (`vX.Y.Z` semver tags).

### Step 4 — Data model, Green and Refactor

- [x] Implement `ci.Manifest`, `ci.ReadManifest`, `ci.Tag` and `ci.ParseTag`, using `yaml.v3` and one anchored regular expression.
- [x] Refactor while green.
- Stories: US7.4, US7.5.

### Step 5 — Business logic, Red: secret scan, release preflight, marketplace entry, workflow policy

- [x] `internal/ci/secrets_test.go` covers `ScanSecrets(root) ([]Finding, error)`. Every bait string is built at run time (for example by joining parts), so the test file itself never matches.
  - A credential-shaped token in a `testdata/` file is found, and the finding gives path, line and rule. The finding never contains the matched text (`project.md` Mandated: redact keys in test output).
  - A credential-shaped token is a run of at least 32 `[A-Za-z0-9]` characters with at least one upper-case letter, one lower-case letter and one digit.
  - The same token with the `TESTSECRET-` prefix is not a finding (AC7.3.4).
  - A quoted `password`/`passwd` literal that does not start with `TESTSECRET-` is a finding.
  - An `apiKey=` query value of 8 or more characters that does not start with `TESTSECRET-` is a finding.
  - These are not findings: lower-case hex digests (SHA-256, commit SHAs), long CamelCase test names, `"k-123456"`, `"test-api-key-0000"` and `apiKey=abc123`.
  - Only the scan scope is read: `**/testdata/**`, `*_test.go`, `ui/src/**/*.test.ts(x)`, `ui/src/testing/**`, `docs/manual-checks/**`, and the test artifacts `coverage.out` and `build/coverage.filtered.out`. `go.sum`, `ui/package-lock.json`, `node_modules/`, `.git/` and `dist/` are not read.
  - Scanning the repository's current tree (`../..`) finds nothing. This is the regression guard for U2 to U4.
- [x] `internal/ci/release_test.go` (preflight part) covers `CheckRelease(Facts) error`. `Facts` holds `Tag`, `ManifestVersion`, `OnMain`, `ReleaseExists`, `ExistingTags` and `ManualChecksDir`.
  - It passes for `v0.1.0`, manifest `0.1.0`, on `main`, with no release yet and a passing first-release record.
  - It refuses a malformed tag.
  - It refuses a tag that differs from the manifest version (`v0.1.1` against `0.1.0`).
  - It refuses a tag whose commit is not on `main` (AC7.5.2).
  - It refuses a tag that already has a Release (AC7.5.2; a released tag is never overwritten).
  - For the first non-pre-release tag (no earlier `vX.Y.Z` tag without a suffix), it refuses unless a `docs/manual-checks/*-first-release.md` record exists with `| Result | pass |`. A record with `fail` is refused (AC7.5.3).
  - A later release, or a pre-release tag, does not need the record.
  - Every refusal names its reason in one line.
- [x] `internal/ci/marketplace_test.go`:
  - `MarketplaceEntry(m Manifest, repo string, categories []string)` returns the YAML entry `- id: nulab-backlog` / `repo: khuongdo/kandev-plugin-nulab-backlog` / `categories: [integrations]`, with no `featured`;
  - it rejects an id that does not match `^[a-z0-9][a-z0-9-]*$`, or a repo that is not `owner/name`;
  - `CheckRegistry(plugins.yaml bytes, manifestID, repo)` reports an error when the entry for `repo` has a different `id` (AC7.6.2);
  - it reports an error when the manifest id is already used by another repo;
  - it passes when the entry matches;
  - it reports "not listed" without error when the repo has no entry yet.
- [x] `internal/ci/workflows_test.go` covers `CheckWorkflows(dir) error` on workflow files written to `t.TempDir()`. It flags: a `uses:` that is not `owner/repo[/path]@<40 hex>`; top-level `permissions` that are missing or are anything other than exactly `contents: read`; a `pull_request_target` trigger; `id-token: write` or `attestations: write` in any job except `publish` in `release.yml`. A compliant pair of files passes.
- [x] Run the tests and record the failing output.
- Stories: US7.3 (AC7.3.1, AC7.3.4), US7.5 (AC7.5.2, AC7.5.3), US7.6 (AC7.6.1, AC7.6.2). Rules: `project.md` NEVER real credentials and NEVER overwrite a released tag; ALWAYS redact; team Deployment (SHA pins, default permissions, no `pull_request_target`, only `publish` has `id-token`/`attestations`).

### Step 6 — Business logic, Green and Refactor

- [x] Implement `ScanSecrets` (scope rules, three regular expressions, line numbers, redacted findings), `CheckRelease` with `FirstReleaseRecorded(dir)`, `MarketplaceEntry` with `CheckRegistry`, and `CheckWorkflows`. Workflows are parsed with `yaml.v3` into a small struct (`on`, `permissions`, `jobs.*.permissions`, `jobs.*.steps[].uses`).
- [x] Refactor while green.
- Stories: US7.3, US7.5, US7.6.

### Step 7 — API / endpoint, Red: packaged-host contract driver and CLI

- [x] `internal/ci/contract_test.go` drives `RunContract(ctx, Config)` against an `httptest` fake Kandev. `Config` holds the base URL, package path, plugin id, expected host version, an injected `Wait` function and the time limits; tests use millisecond limits and never sleep for real.
  - Happy path, in order:
    1. `/ready` reaches `status: "ok"`, and its `version` equals `v0.96.0` (the host is exactly the minimum version).
    2. `POST /api/plugins/install` sends the multipart field `package` with the archive bytes and gets 201.
    3. `GET /api/plugins/nulab-backlog` reaches `status: "active"`.
    4. `GET /api/v1/workspaces` returns the first workspace id.
    5. `POST …/actions/connection.get` with `{workspaceId, body: {}}` returns 200 with `state: "not_connected"` and `enabled: true`.
    6. `POST …/actions/connection.connect_api_key` with an invalid `spaceUrl` and a `TESTSECRET-` key returns 400 with `code: "validation"` and `field: "spaceUrl"`, with no network call. This exercises an admin action and the error mapping.
  - Install answers 400 or 409, or 201 with a `warning`: the run fails, and the error includes Kandev's message.
  - The status reaches `error`: the run fails at once. The status never reaches `active`: the run fails at the time limit.
  - `/ready` never reaches `ok`, or reports another version: the run fails (AC7.4.2).
  - There is no workspace, or `connection.get` returns non-200 or the wrong state: the run fails.
  - No error text contains the `TESTSECRET-` bait key (`testutil.AssertNoLeak`).
- [x] `internal/ci/run_test.go` covers `Run(args, stdout, stderr) int`. It returns 0, 1 or 2, like `pkgverify.Run`. The subcommands are `secrets`, `preflight`, `marketplace`, `workflows` and `contract`. An unknown subcommand or a missing required flag returns 2. Findings and refusals return 1. `secrets` on a clean root prints `OK` and returns 0.
- [x] Run the tests and record the failing output.
- Stories: US7.4 (AC7.4.1, AC7.4.2), US7.3. Rules: team Testing Posture (contract test on `min_kandev_version`; injected wait functions; testify/require).

### Step 8 — API / endpoint, Green and Refactor

- [x] Implement `ci.RunContract`, using the standard `net/http` and `mime/multipart` packages, with a 1 MiB limit on response reads (`io.LimitReader`) and `context.Context` first. Implement `ci.Run`. Add `cmd/ci/main.go`, which only calls `os.Exit(ci.Run(os.Args[1:], os.Stdout, os.Stderr))`.
- [x] Refactor while green.
- Stories: US7.3, US7.4, US7.5, US7.6.

### Step 9 — Environment and build configuration (Makefile and workflows)

- [x] Red: add `TestRepositoryWorkflowsFollowThePolicy` to `internal/ci/workflows_test.go`. It runs `CheckWorkflows("../../.github/workflows")` and requires that `release.yml` exists with a `publish` job holding `id-token: write` and `attestations: write`. Run it and record the failing output (`release.yml` is missing).
- [x] `Makefile` (additive; C4's existing targets keep their names and behaviour):
  - `MIN_KANDEV_VERSION` is parsed from `manifest.yaml` with `sed`, like `VERSION` (AC7.4.1).
  - `ACTIONLINT_VERSION` is a pinned tag. `lint` additionally runs `go run github.com/rhysd/actionlint/cmd/actionlint@$(ACTIONLINT_VERSION)` and `go run ./cmd/ci workflows -dir .github/workflows`.
  - `check-secrets`: `go run ./cmd/ci secrets -root .`
  - `contract-test`: (1) `KANDEV_MIN_DIR ?= ../kandev-min`; fail unless its HEAD is the tag `v$(MIN_KANDEV_VERSION)`. (2) `make -C $(KANDEV_MIN_DIR)/apps/backend build-kandev VERSION=v$(MIN_KANDEV_VERSION)`. (3) Start `bin/kandev __backend` with `KANDEV_HOME_DIR=$$(mktemp -d)`, `KANDEV_SERVER_HOST=127.0.0.1` and `KANDEV_SERVER_PORT=$(CONTRACT_PORT)`; a `trap` stops it and removes the temp home. (4) Run `go run ./cmd/ci contract -base-url … -package $(PACKAGE) -plugin-id $(PLUGIN_ID) -host-version v$(MIN_KANDEV_VERSION)`.
  - `release-preflight`: require `TAG`; compute `on_main` with `git merge-base --is-ancestor "$(TAG)^{commit}" origin/main`; compute `released` with `gh release view "$(TAG)"`; run `go run ./cmd/ci preflight -tag … -version $(VERSION) -on-main … -released … -tags "$$(git tag -l 'v*')" -manual-checks docs/manual-checks`.
  - `marketplace-entry`: `go run ./cmd/ci marketplace -repo khuongdo/kandev-plugin-nulab-backlog -categories integrations [-registry <plugins.yaml>]`.
  - Update `help` and `.PHONY`.
- [x] `.github/workflows/ci.yml`:
  - Job `checks` runs `make check-format vet lint test coverage check-secrets build package verify-package`. `check-secrets` comes after `coverage` so `coverage.out` is scanned.
  - New job `packaged-host-contract` (`timeout-minutes: 30`, no extra permissions): check out the plugin into `plugin/`; read `.kandev-sdk-ref` and `min_kandev_version` with `sed`, and check both formats; check out `kdlbs/kandev` at `.kandev-sdk-ref` into `kandev/`, and at `v<min>` into `kandev-min/`; set up Go and Node; run `npm ci`; run `make package verify-package contract-test KANDEV_MIN_DIR=../kandev-min`.
  - Keep `pull_request` and `push` to `main`, `permissions: contents: read`, and every action SHA-pinned.
- [x] `.github/workflows/release.yml`:
  - Triggered on `push: tags: ['v*']`. Top-level `permissions: contents: read`; concurrency group `release`, no cancel.
  - Job `verify`: check out with `fetch-depth: 0`, plus the SDK checkout as in `ci.yml`; set up Go and Node; run `npm ci`; run the `go mod tidy` diff check; run `make check-format vet lint test coverage check-secrets build package verify-package release-preflight TAG=${{ github.ref_name }}`, with `GH_TOKEN: ${{ github.token }}` (package verification runs before preflight passes, `project.md` Mandated); upload `dist/`.
  - Job `contract`: the same steps as `packaged-host-contract` (US7.4: tested on the minimum version before release).
  - Job `publish` (`needs: [verify, contract]`; the only job with `contents: write`, `id-token: write` and `attestations: write`): download `dist/`; run `cd dist && sha256sum -c checksums.txt`; run `actions/attest-build-provenance` with `subject-path: dist/nulab-backlog-*.tar.gz`; run `gh release create "$TAG" dist/*.tar.gz dist/checksums.txt --verify-tag --title "$TAG" --generate-notes`, plus `--prerelease` when the tag has a `-` suffix.
  - Every action is pinned to a full commit SHA with its version in a comment. SHAs for `actions/download-artifact` and `actions/attest-build-provenance` are looked up from the actions' release tags at this step and recorded in `code-summary.md`.
- [x] Green: re-run the workflow-policy test, then `make lint` (actionlint plus the policy) until clean. Refactor while green.
- Stories: US7.3 (AC7.3.1, AC7.3.4), US7.4 (AC7.4.1, AC7.4.2), US7.5 (AC7.5.1, AC7.5.2). Rules: team Deployment and Code Style (CI calls Makefile targets); C4 (renaming or removing a target is breaking, adding one is not).

### Step 10 — Local verification of the gates

- [x] From `make clean`, run `make check-format vet lint test coverage check-secrets build package verify-package`. Confirm that Go coverage is still at least 80% and that `go mod tidy` leaves no diff.
- [x] Run `make contract-test KANDEV_MIN_DIR=../kandev` (`../kandev` is at the v0.96.0 tag, the same commit as `min_kandev_version`), and record the output. If the backend cannot start without web assets or helper binaries, record the gap and the extra build target needed. Do not weaken the assertions.
- [x] Negative evidence, recorded without changing any committed threshold:
  - `make coverage COVERAGE_MIN=100` exits non-zero, which shows that the gate fails closed (AC7.3.1).
  - `go run ./cmd/ci secrets -root <scratchpad dir with a planted run-time-built key>` exits 1 and does not print the key (AC7.3.4).
  - `go run ./cmd/ci preflight` with `-on-main=false`, `-released=true`, a mismatched version, and a first release without a record: each exits 1 with its reason (AC7.5.2, AC7.5.3).
  - `go run ./cmd/ci marketplace -registry <copy with a wrong id>` exits 1 (AC7.6.2).
- [x] Do not push any tag. The release workflow and attestation are proven on GitHub only after merge (see Open Questions).
- Stories: US7.3, US7.4, US7.5, US7.6.

### Step 11 — Documentation and traceability

- [x] `README.md`:
  - a "CI checks" section: the two required checks `checks` and `packaged-host-contract`, which repository settings to set (X4), and that `main` protection must require both;
  - a "Releasing" section: (1) bump `manifest.yaml` `version` in a pull request; (2) for the first release, commit the second manual-check record `docs/manual-checks/<date>-first-release.md`; (3) run `make verify-package` locally, then tag `vX.Y.Z` on `main`; (4) verify with `gh attestation verify dist/nulab-backlog-X.Y.Z.tar.gz --repo khuongdo/kandev-plugin-nulab-backlog`; (5) a broken release is marked and fixed forward; a released tag is never deleted or overwritten;
  - a "Marketplace" section: run `make marketplace-entry`, then open a pull request on a fork of `kdlbs/kandev` adding the entry to `plugin-registry/plugins.yaml`; the README states that Kandev shows the package as unsigned.
- [x] Write `code-summary.md`, `source-manifest.json` and `traceability.json` under `construction/ci-release/code-generation/`.
- Stories: US7.5 (AC7.5.3 procedure), US7.6 (AC7.6.1 procedure).

## Story-to-Step Map

| Story | ACs | Steps |
|-------|-----|-------|
| US7.3 Quality gates in CI | AC7.3.1 (all CI steps, merge blocked; through Makefile targets and branch protection) | 9, 10, 11 |
| | AC7.3.2, AC7.3.3 (no bait key or secret URL in errors or logs) | Covered by U1's existing leak tests, which `make test` runs in CI (Step 9). The list, refresh and Git flows add their own tests in U2 to U4 |
| | AC7.3.4 (no real-looking secret in test data or artifacts) | 5, 6, 9, 10 |
| US7.4 Test on the minimum Kandev version | AC7.4.1, AC7.4.2 | 3, 4, 7, 8, 9, 10 |
| US7.5 Release a new version | AC7.5.1 (Release with package, `checksums.txt` and attestation) | 9, 11; the proof on GitHub happens after merge |
| | AC7.5.2 (refuse an existing tag or a commit not on `main`) | 3–6, 9, 10 |
| | AC7.5.3 (second manual-check record before the first tag) | 5, 6, 11 |
| US7.6 Marketplace submission | AC7.6.1 (entry fields) | 5, 6, 9, 11; the pull request is sent after `v0.1.0` |
| | AC7.6.2 (catalogue id differs from manifest id) | 5, 6, 10 |

**Shared files touched**: `Makefile` (new variables and targets, `lint` extended), `.github/workflows/ci.yml` (one target added, one job added), `README.md` (three sections). `manifest.yaml` is read but not changed. U2 is built after U5 and rebases onto these changes.

**Files created**: `internal/ci/{doc,manifest,release,secrets,marketplace,workflows,contract,run}.go` and their `_test.go` files; `cmd/ci/main.go`; `.github/workflows/release.yml`.

## Sources

- [desc] Initial description: Kandev plugin for Nulab Backlog.
- [scope] Workflow-selected scope: `feature`.
- `unit-of-work.md` (U5), `unit-of-work-story-map.md`, `unit-of-work-dependency.md`; `stories.md` (US7.3–US7.6, bait-secret convention); `requirements.md` (FR7.1–FR7.3, NFR3, NFR4, NFR6, NFR8); `contract-summary.md` (C4, C8).
- `bolt-plan.md` (B2, Release Milestone), `external-dependency-map.md` (X4, X8), `risk-and-sequencing-rationale.md` (R-E).
- `team-practices.md`; memory `team.md` (Way of Working, Testing Posture, Deployment, Code Style) and `project.md` (Forbidden, Mandated).
- U1 `cicd-pipeline.md`, `infrastructure-specification.md`, `code-generation-plan.md`, `code-summary.md`.
- Kandev v0.96.0 (`../kandev`): `apps/backend/cmd/kandev/main.go`, `apps/backend/Makefile`; `internal/plugins/handlers.go`, `action_handlers.go`, `store/store.go`; `internal/plugins/manifest/semver.go`, `validate.go`; `internal/auth/authn/identity.go`; `internal/task/handlers/workspace_handlers.go`, `task/dto/dto.go`, `task/repository/sqlite/defaults.go`; `internal/backendapp/health_test.go`; `docs/public/{plugins-authoring,configuration,authentication}.md`; `docs/decisions/2026-08-01-bitbucket-initial-release-remains-unsigned.md`; `plugin-registry/{README.md,plugins.yaml,schema.json}`.
- `github.com/kdlbs/kandev-plugin-bitbucket` `.github/workflows/{ci,release,build}.yml` (main, fetched 2026-10-06).

## Assumptions & Open Questions

- [assumption] The packaged-host contract test runs at the API level, not in a browser. It covers liveness at the exact host version, multipart install, `active` status, `connection.get` and one admin action that fails validation. The Bitbucket template's Playwright spec lives in the Kandev repo and cannot be reused, and the team rule only asks to "install the packaged plugin and run it".
- [assumption] The throwaway server is `kandev __backend`, built with `make build-kandev VERSION=v<min>`, with authentication left at its default (disabled), a temporary `KANDEV_HOME_DIR`, `127.0.0.1` and a fixed port (`CONTRACT_PORT ?= 38529`). Whether the backend starts without the web build or the agentctl helpers is confirmed in Step 10. If it does not, the extra Kandev build target is added; the assertions are not weakened.
- [assumption] `/ready` returns JSON with `status` and `version`. The field name is taken from `health_test.go` (`versionFieldKey`), and the code that defines it was not read.
- [assumption] For a tag-triggered workflow, "refuses if the tag already exists" means that a GitHub Release already exists for that tag. Repository tag-protection rules (set by the user) stop a tag from being moved or deleted.
- [assumption] "Runs only from main" means that the tagged commit is an ancestor of `origin/main` (`fetch-depth: 0`). This matches AC7.5.2 ("tag on a commit not on `main`").
- [assumption] `gh release view` cannot tell "not found" apart from a network error, so `released=false` may be wrong if GitHub is down. `gh release create` still refuses to overwrite an existing Release.
- [assumption] Pre-release tags (`vX.Y.Z-suffix`) are allowed, are published as pre-releases, and skip the first-release manual-check rule. The bolt plan proves the workflow on a pre-release tag.
- [assumption] Build metadata (`+…`) is rejected in tags.
- [assumption] AC7.5.3 is enforced by preflight. The first non-pre-release tag needs `docs/manual-checks/*-first-release.md` with `| Result | pass |`, which follows U1's `TEMPLATE.md` naming.
- [assumption] Secret-scan heuristics: a credential-shaped token is 32 or more alphanumeric characters with at least one upper-case letter, one lower-case letter and one digit; a quoted `password`/`passwd` literal is a finding; an `apiKey=` value of 8 or more characters is a finding; in all three cases, a `TESTSECRET-` prefix is allowed (AC7.3.4). The real Backlog API key format is not stated upstream.
- [assumption] The secret-scan scope is committed test files, `testdata/`, `docs/manual-checks/` and the coverage profiles. `dist/` is not scanned: it is a build artifact, not a test artifact, and holds only binaries and the bundle.
- [assumption] `internal/testutil` keeps its run-time `test-api-key-` prefix. Those keys never reach committed files. Renaming it to `TESTSECRET-` would touch U1 and U2 test code and is left out.
- [assumption] The marketplace entry uses `categories: [integrations]`, as the Bitbucket entry does. The manifest's `categories: ["connector"]` stays as it is.
- [assumption] The marketplace pull request to `kdlbs/kandev` is a documented manual step after `v0.1.0`. Automating it would need a write token for a fork, and CI uses no secrets (U1 `cicd-pipeline`). AC7.6.1's catalogue rules are checked again against the registry's current `main` when the pull request is sent.
- [assumption] actionlint runs through `go run …@<pinned tag>`, like golangci-lint, so it adds no repository dependency. The exact tag is chosen at Step 9 and recorded.
- [assumption] The workflow-policy check (SHA pins, default permissions, no `pull_request_target`, attestation permissions only in `publish`) runs inside `make lint`, so the team's Deployment rules are enforced by machine.
- [assumption] The release uses the `gh` CLI (pre-installed on `ubuntu-latest`) instead of `softprops/action-gh-release`, so there is one fewer third-party action to pin. `dist/checksums.txt` (the tarball's SHA-256) is attached as it is; the archive keeps its own internal `checksums.txt`.
- [assumption] The `publish` job also needs `contents: write` to create the Release. It is the only job with any write permission.
- [assumption] The contract test installs the full five-platform package that CI built, not a host-only package, so the bytes tested are the bytes released.
- Open: the B2 demo ("a pre-release tag produces a GitHub Release whose attestation verifies") needs `manifest.yaml` `version` set to the pre-release (for example `0.0.1-rc.1`) on `main`, and the resulting tag and Release are permanent. Should the user run this after merge, or defer the first attestation proof to `v0.1.0`?
- Open: branch protection (X4) must add `packaged-host-contract` as a required check next to `checks`. The user sets this once in GitHub.
- Open: running `make contract-test KANDEV_MIN_DIR=../kandev` locally writes `bin/kandev` inside the sibling SDK checkout (a build output). A separate `../kandev-min` worktree avoids this.
