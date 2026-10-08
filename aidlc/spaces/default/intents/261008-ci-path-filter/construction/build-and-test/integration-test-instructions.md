# Integration Test Instructions — CI path filter

## Scope

Test Strategy is Minimal, so no separate integration test suite is generated; unit tests from Code Generation cover the change (`unit-test-instructions.md`).

## Existing Integration Check Run in This Stage

The packaged-host contract test is the project's existing integration check and was run once locally:

```bash
make contract-test KANDEV_MIN_DIR=../kandev
```

## Real-CI Behaviour (observed after the pull request is opened)

The workflow behaviour can only be proven on GitHub Actions. On the pull request for this change (it touches `.github/` and `internal/`, so it is app):

- `changes` prints `app=true`; `checks`, `packaged-host-contract` and `secret-scan` all run and pass.

After merge, a records-only pull request (only `aidlc/` files) should show `checks` and `packaged-host-contract` as skipped (passing) and `secret-scan` as passed, and be mergeable.
