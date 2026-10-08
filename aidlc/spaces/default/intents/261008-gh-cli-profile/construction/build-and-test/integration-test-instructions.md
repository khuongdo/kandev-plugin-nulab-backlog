# Integration Test Instructions — 261008-gh-cli-profile

## Scope

Minimal test strategy: no separate integration suite is generated. The integration boundary that matters for this change is the packaged plugin running inside a real Kandev host, covered by the existing packaged-host contract test.

## How to Run

```bash
export PATH=$HOME/.local/go/bin:$PATH
make contract-test KANDEV_MIN_DIR=../kandev
```

Run it 10 times in a row (project rule) to catch host startup races. `../kandev` must be at `min_kandev_version` (v0.96.0).

## Coverage Expectation

Contract test passes 10/10; it installs the 0.5.3 package (which declares the new `scm.providers.cli_accounts` action) on a throwaway Kandev server.

## Test Data / Environment

Throwaway server built by the Makefile target; no real GitHub, Backlog or gh.

## Manual Check (outside this stage)

With two gh accounts logged in on the Kandev server: connect workspace W1 as account A and W2 as account B, `gh auth switch` between them, confirm each card still shows its own `@login` and PR lists differ. Not required by the team posture (manual end-to-end is only at skeleton and first release); recommended before tagging.
