# Performance Test Instructions — CI path filter

## Scope

Not applicable. Test Strategy is Minimal and the requirements define no performance target. The intended effect (non-app-only pull requests skip the Kandev build, packaging and contract test) is checked structurally by `TestRepositoryWorkflowsSkipAppJobsOnlyForNonAppChanges` and observed on GitHub Actions run times after merge.

## Commands

None.
