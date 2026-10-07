# Integration Test Instructions — Opt-in by default

## Applicability

Test strategy is Minimal, so no separate integration suite is generated. The one integration-level check that matters for this fix already exists: the packaged-host contract test, which installs the real package on Kandev v0.96.0 and was updated to the opt-in default (FR6).

## How to Run

```bash
export PATH=$HOME/.local/go/bin:$HOME/go/bin:$PATH
make package
make contract-test KANDEV_MIN_DIR=../kandev   # run up to 10 times to catch host startup races
```

## Expected Result

`ci contract: OK nulab-backlog on Kandev v0.96.0`. The contract asserts a fresh install reports `enabled: false`, refuses `connection.connect_api_key` with `409 integration_disabled`, then turns the switch on via `connection.set_enabled` and still exercises the real connect validation (`400 validation spaceUrl`).

## Test Data and Environment

Throwaway Kandev home created by the target; no real Backlog space or credentials.
