# Integration Test Instructions - Fix UIUX

## Scope at Minimal Strategy

No new integration suite is generated (Minimal strategy). The existing packaged-host contract test is the integration check: it builds Kandev at `min_kandev_version` (0.96.0), installs the packaged plugin, and runs it. The project rule is to run it 10 times to catch host startup races; this matters here because `initialize` now awaits a bounded ON/OFF check before registering the Home > Integrations entry (FR3, NFR3).

## How to Run

```bash
export PATH=$HOME/.local/go/bin:$PATH
make package
for i in $(seq 1 10); do make contract-test KANDEV_MIN_DIR=../kandev || break; done
```

## Expected Result

Every run prints `ci contract: OK nulab-backlog on Kandev v0.96.0`.

## Not Covered Here

The contract test does not render the browser UI. The hover card, the badge on Home > Tasks rows, the sidebar and the task top bar, and the hidden Integrations entry are covered by Vitest against the fake host; a look in a real Kandev 0.96.0 browser session is a manual check (see build-and-test-summary.md, Known Limitations).
