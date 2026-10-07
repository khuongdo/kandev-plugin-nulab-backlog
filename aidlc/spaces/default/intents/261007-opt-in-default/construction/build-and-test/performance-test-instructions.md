# Performance Test Instructions — Opt-in by default

## Applicability

Not applicable. The requirements for this bugfix define no performance target (requirements.md: NFR1–NFR3 cover regression testing, suite health and security only). The change flips one default; the switch read (`LoadSwitch`) is the same single state read as before, so request cost is unchanged.

## How to Run

No performance tests are run for this change.

## Coverage Expectations

None beyond the existing suite.
