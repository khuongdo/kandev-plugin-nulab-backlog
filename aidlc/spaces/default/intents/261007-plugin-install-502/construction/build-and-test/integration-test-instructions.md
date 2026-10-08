# Integration Test Instructions — Plugin install failed: 502

## Scope at Minimal Strategy

The active Test Strategy is Minimal, so no new integration test suite is generated. The only integration-level check this fix depends on is the existing packaged-host contract test, which installs the real package on Kandev at the manifest's `min_kandev_version` and calls two plugin actions.

## How to Run

```bash
export PATH=$HOME/.local/go/bin:$PATH
make contract-test KANDEV_MIN_DIR=../kandev
```

`../kandev` must be at tag `v0.96.0`. Per the project Testing Posture, run it 10 times to catch host startup races.

## Expected Result

`ci contract: OK nulab-backlog on Kandev v0.96.0` on every run. This proves a Linux/amd64 Kandev accepts the 4-executable package (`pkgtar` picks the host platform binary).

## Test Data

None beyond the built package; the target starts a throwaway Kandev home in a temp directory.
