# Integration Test Instructions

## Applicability

Test strategy is Minimal, so no separate integration suite is generated. The team's one integration-level check is the packaged-host contract test on the minimum Kandev version (team Testing Posture), which installs the real package into a throwaway Kandev v0.96.0 server and runs it.

## How to Run

```bash
make package
make contract-test KANDEV_MIN_DIR=../kandev   # run 10 times to catch host start-up races
```

`../kandev` must be at the `min_kandev_version` in `manifest.yaml` (0.96.0).

## Expected Result

All 10 runs pass. The new action `scm.active.set` is part of the packaged manifest that the host loads.

## Test Data

None beyond the contract test's own fixtures; no real Backlog or GitHub/GitLab/Bitbucket credentials.
