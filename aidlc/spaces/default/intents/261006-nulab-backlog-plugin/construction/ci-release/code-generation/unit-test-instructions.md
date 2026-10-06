# Unit Test Instructions — ci-release (U5)

## Frameworks and Setup

- **Go**: the standard `testing` package with `github.com/stretchr/testify/require`. Tests are table-driven with `t.Run`, and every test name describes a behaviour. Tests always run with `-race`, which needs CGO and a C toolchain. Go 1.26 must be on `PATH` (the U1 runs used the Go 1.26.8 toolchain under `~/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.26.8.linux-amd64/bin` with `GOTOOLCHAIN=local`).
- **Kandev SDK**: `../kandev` must exist at the commit in `.kandev-sdk-ref`, because the module's `replace` points there (`make check-sdk`).
- **Fake Kandev**: a `net/http/httptest` server inside `internal/ci/contract_test.go`. It serves `/ready`; `POST /api/plugins/install` (it reads the multipart field `package`); `GET /api/plugins/{id}`; `GET /api/v1/workspaces`; `POST /api/plugins/{id}/actions/{key}`. Each test scripts its responses: success, 400 or 409, a `warning`, an `error` status, a wrong version, no workspace.
- **Waiting**: `RunContract` takes an injected `Wait` function and its time limits in `Config`. Tests use millisecond limits and a no-op wait, and never call real `time.Sleep`.
- **Files**: manifests, workflow files, manual-check records and scan trees are written under `t.TempDir()`.
- **Repository regression tests** read the real `manifest.yaml`, `.github/workflows/` and the repository tree through `../..` paths. They need no network.
- No new dependency is added.

## Commands for This Unit

Run these from the repository root.

- Go, all U5 code:

  ```
  go test -race ./internal/ci/...
  ```

- Go, one layer while working (examples):

  ```
  go test -race ./internal/ci/... -run 'TestReadManifest|TestParseTag'
  go test -race ./internal/ci/... -run 'TestScanSecrets|TestCheckRelease|TestMarketplace|TestCheckRegistry|TestCheckWorkflows'
  go test -race ./internal/ci/... -run 'TestRunContract|TestRun'
  go test -race ./internal/ci/... -run 'TestRepositoryWorkflowsFollowThePolicy'
  ```

- Go, the existing package verifier, which the release re-runs through `verify-package`:

  ```
  go test -race ./internal/pkgverify/...
  ```

- Coverage for the U5 package:

  ```
  go test -race -coverprofile=coverage.out ./internal/ci/...
  ```

  Then run `go tool cover -func=coverage.out`. The project floor is still enforced by `make coverage` over `./internal/... ./server/...`.

- Workflow files (after Step 9):

  ```
  go run github.com/rhysd/actionlint/cmd/actionlint@$ACTIONLINT_VERSION .github/workflows/ci.yml .github/workflows/release.yml
  go run ./cmd/ci workflows -dir .github/workflows
  ```

- Secret scan of test data and test artifacts (after Step 8):

  ```
  go run ./cmd/ci secrets -root .
  ```

- Packaged-host contract test (integration; builds Kandev at the minimum version, needs gcc and a free port 38529):

  ```
  make package contract-test KANDEV_MIN_DIR=../kandev
  ```

Before Step 2, `./internal/ci/...` matches no packages. Step 2 adds a smoke test so the first command runs before the first Red step. The `cmd/ci` commands work only after Step 8.

## Coverage Target

At least 80% line coverage over `./internal/...` and `./server/...`, excluding only `server/main.go` (team Testing Posture, NFR8). `internal/ci` is inside that scope and aims at the same 80% or more on its own. `cmd/ci/main.go` is a one-line wrapper outside the measured scope, like `cmd/verifypkg`. The floor is never lowered, and no exclusion is added. Using `COVERAGE_MIN=100` on the command line in Step 10 only proves that the gate fails; it is not a change.

## Expected Test Volume (Standard strategy)

| Component | Test file | Approximate tests |
|-----------|-----------|-------------------|
| Manifest facts | `internal/ci/manifest_test.go` | 5 |
| Tag parsing | `internal/ci/release_test.go` (tag part) | 1 table, about 9 rows |
| Release preflight | `internal/ci/release_test.go` (preflight part) | 8 |
| Secret scan | `internal/ci/secrets_test.go` | 8, plus a table of non-findings |
| Marketplace entry and registry check | `internal/ci/marketplace_test.go` | 6 |
| Workflow policy | `internal/ci/workflows_test.go` | 6, including the real-repository test |
| Contract driver | `internal/ci/contract_test.go` | 8 |
| CLI dispatch | `internal/ci/run_test.go` | 5 |

The contract-driver tests with the fake Kandev are the integration tests for the key boundary between the plugin package and the Kandev HTTP API. `make contract-test` is the real end-to-end check against Kandev v0.96.0.

## Mocking and Stubbing

- Never call real Backlog or real GitHub in automated tests. `git` and `gh` facts reach `CheckRelease` as plain values (`OnMain`, `ReleaseExists`, `ExistingTags`), so the preflight logic needs no process calls.
- Use fakes, not mocks with call expectations. The fake Kandev records the request order and the multipart field name, so the tests can check them.
- Bait secrets: every fake key, token or password starts with `TESTSECRET-` (stories bait-secret convention). Credential-shaped strings that the scanner must find are built at run time (for example `strings.Join` of fragments, or random characters from `crypto/rand`). Committed files therefore never contain one, and `make check-secrets` stays clean on this repository.
- Scanner findings and CLI output never print the matched text. Tests check this with `testutil.AssertNoLeak`.

## Test Data

- Keep each test's data local, in `t.TempDir()`, and use `t.Setenv()` if the environment is touched.
- Registry fixtures are tiny `plugins.yaml` strings built inside the test, shaped like `kdlbs/kandev` `plugin-registry/plugins.yaml` (`id`, `repo`, `categories`).
- Manual-check records in tests follow `docs/manual-checks/TEMPLATE.md` (`| Result | pass |`), contain no space URL with a query, and contain no key.
- Package bytes for the contract driver are a few dummy bytes. The fake Kandev only checks that they arrive in the `package` field, and never unpacks them.
