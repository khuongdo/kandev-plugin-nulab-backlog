# Integration Test Instructions — Multi-provider source control

## Scope

Test strategy is Minimal, so no new integration suite was generated. The boundaries of this change are covered by tests that already run in the normal suite:

- Plugin action boundary (`internal/plugin` ↔ `internal/scm` ↔ fake host): `internal/plugin/actions_scm_test.go`, including the Backlog on/off guard and error-code mapping.
- v0.3.0 data compatibility through the real `git.*` actions: `internal/plugin/v030_test.go` with fixtures in `internal/plugin/testdata/v030/`.
- Client ↔ provider HTTP boundary: `internal/{github,gitlab,bitbucket}/client_test.go` against `httptest` fake servers with JSON fixtures.
- Packaged plugin ↔ real Kandev host: the packaged-host contract test.

## How to Run

```bash
export PATH=$HOME/.local/go/bin:$PATH
go test -race -count=1 -run 'SCM|Scm|Manifest|V030' ./internal/plugin/...
go test -race -count=1 ./internal/github/... ./internal/gitlab/... ./internal/bitbucket/...
for i in $(seq 1 10); do make contract-test KANDEV_MIN_DIR=../kandev || break; done
```

## Coverage Expectations

Covered by the Go 80% line-coverage floor (`make coverage`); no separate integration coverage target.

## Test Data and Environment

Fake providers and fake tokens only (`internal/<pkg>/testdata/`). No real GitHub, GitLab or Bitbucket account is used; a manual check against real accounts is not part of this stage (see the summary's known limitations).
