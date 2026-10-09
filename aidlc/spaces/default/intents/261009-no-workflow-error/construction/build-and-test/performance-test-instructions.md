# Performance Test Instructions — 261009-no-workflow-error

## Scope

The Minimal test strategy generates no load or benchmark suite, and no performance-validation stage is scheduled for this bugfix scope.

## Applicable Target

- NFR2: when the Kandev context is missing, the dialog opens (or the no-workflow toast appears) within 1 s on a normal connection, with at most one host call.

## How It Is Checked

- "At most one host call": asserted by unit tests (`HasWorkflow` uses `Workflows().List` with `Limit: 1`; `start-task.test.tsx` checks the action is called only when the context is null).
- "Within 1 s": observed during the real-Kandev manual check (integration-test-instructions.md step 2).
