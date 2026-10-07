# Integration Test Instructions — github-parity-actions

## Scope

The test strategy is Minimal, so no separate integration suite was generated. The integration-level check in this project is the team's packaged-host contract test. It installs the real `0.2.0` package into a throwaway Kandev `v0.96.0` server and runs it, which proves the manifest, including the 7 new actions and the `max_body_bytes` of `issues.quick_actions.save`, is accepted by the minimum host.

## How to Run

```bash
export PATH=$HOME/.local/go/bin:$PATH
make package
for i in $(seq 10); do make contract-test KANDEV_MIN_DIR=../kandev || break; done   # project rule: 10 runs
```

Expected: every run ends with `ci contract: OK nulab-backlog on Kandev v0.96.0`.

## Test Data and Environment

Each run uses a temporary Kandev home directory, removed when the run ends. No Backlog space or credentials are used.
