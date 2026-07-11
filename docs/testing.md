# Testing

## Command

```bash
make test
```

This runs:

```bash
go test -v -race -count=1 ./...
```

In the Codex sandbox, use a writable Go build cache:

```bash
GOCACHE=/tmp/book-social-go-cache make test
```

## Current Coverage Shape

The project uses the standard Go testing package.

Current tests cover:

- app route registration and HTTP behavior
- home handler behavior
- books service behavior
- books handler behavior with fakes
- SQLite books repository behavior
- renderer behavior
- response helpers
- navigation view helpers
- logging middleware

There are also small integration-style HTTP tests that use `httptest` and a temporary SQLite database.

## Test Databases

Tests should not read or write the local development database at `./data/book_social_dev.db`.

Current repository and HTTP integration tests create disposable SQLite databases:

- repository tests use in-memory SQLite databases
- HTTP integration tests use temporary SQLite files from `t.TempDir()`
- tests insert minimal deterministic data needed by the behavior under test

For now, test schema and fixture setup lives inline in the relevant test files. This is acceptable
for v0.1, but v0.2 schema work should add small shared bootstrap helpers before the test setup grows.

Do not use the full development seed dataset in ordinary unit or handler tests. Use full seed data
only for an explicit seed smoke test or database setup check.

## Testing Guidance

- Prefer table-driven tests.
- Use `httptest` for HTTP handlers.
- Use fake repositories/services for unit tests.
- Avoid browser/e2e tests for now.
- Avoid large HTML snapshot tests.
- Avoid database integration tests unless the task explicitly needs them.

## Codex Sandbox Note

Do not start the web server inside the Codex sandbox for verification.

For HTTP behavior, add or update Go tests using `httptest`.
