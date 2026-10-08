# Integration Test Instructions

## Scope

The Minimal test strategy adds no new integration test file for this change. Two existing integration-level checks still apply because they are team quality targets:

- **Packaged-host contract test** (team Testing Posture): installs the real package into a throwaway Kandev server built from `../kandev` at `v0.96.0` and runs it. It verifies that the new manifest action `scm.providers.use_cli` is accepted by the minimum Kandev version and that the plugin starts.
- **Manifest/runtime parity test** (`internal/plugin/manifest_test.go`): every runtime action key is declared in `manifest.yaml` and vice versa.

The CLI itself (`gh`, `glab`) is never run in automated tests (NFR3); the service uses a fake runner.

## How to run

```bash
export PATH="$HOME/.local/go/bin:$PATH"
# Project rule: run the contract test 10 times to catch host startup races.
for i in $(seq 1 10); do make contract-test KANDEV_MIN_DIR=../kandev || break; done

# Parity test
go test -race ./internal/plugin/ -run 'Manifest'
```

## Expected result

10/10 contract-test runs pass; the parity test passes.

## Manual end-to-end check

The team's two manual checks against a real Backlog space (walking skeleton, first release) are already done; this change does not require one. An optional manual check of the new button on the self-hosted Kandev (server user logged in with `gh auth login`) is recommended before release.
