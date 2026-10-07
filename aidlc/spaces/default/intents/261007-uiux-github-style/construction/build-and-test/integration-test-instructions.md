# Integration Test Instructions — 261007-uiux-github-style

## Scope

The active test strategy is Minimal, which needs no separate integration suite. The one integration check this project always runs is the team's packaged-host contract test (Testing Posture): the packaged plugin is installed into a real Kandev at `min_kandev_version` (0.96.0) and exercised end to end, which covers the new `issues.watches.*` and `git.prs.list` actions being registered and callable.

## Setup

- `../kandev` checked out at tag `v0.96.0` (same as `.kandev-sdk-ref`), Go with gcc.
- `make package` has produced the package under `build/`.

## How to Run

```bash
make contract-test KANDEV_MIN_DIR=../kandev
```

Per the project correction on asynchronous Host injection, run it repeatedly to catch startup races:

```bash
for i in 1 2 3 4 5 6 7 8 9 10; do make contract-test KANDEV_MIN_DIR=../kandev || break; done
```

## Expected Result

Every run passes. CI's `packaged-host-contract` job runs the same check on each pull request.

## Test Data

The test uses a temporary Kandev home and no Backlog connection; no credentials are involved.
