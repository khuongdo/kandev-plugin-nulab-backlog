# Performance Test Instructions — 261007-backlog-panel-retouch

## Applicability

Not applicable. The test strategy is Minimal, and the requirements (NFR1–NFR5) define no performance target. The refactor changes UI rendering only; it adds no Backlog API call and removes the 400 ms auto-search (the issue list now reloads only on Enter, blur with a changed query, filter change or saved query), so request volume can only go down.

## Regression Signal

The unit tests in `ui/src/issues/issues-page.test.tsx` assert call counts for `issues.list` (no reload while typing; one reload per commit), which guards against accidental extra requests.
