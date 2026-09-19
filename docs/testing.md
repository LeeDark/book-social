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
- logging middleware, including recovered panic status and structured panic diagnostics
- trusted-proxy client IP handling, including trusted, untrusted, and disabled configurations
- graceful shutdown lifecycle and `http.ErrServerClosed` handling with fake servers
- dynamic timeout cancellation and `504 Gateway Timeout` behavior
- security headers and static cache policy
- renderer failure before response status commit
- bcrypt hashing/verification, byte limits, invalid stored hashes, and a dummy verification path
  for unknown accounts
- registration normalization, validation, duplicate identity mapping, transaction ownership, and
  safe internal errors
- DB-backed session create/load/delete, absolute expiry, 32-byte token-hash constraints, and
  invalidation
- two-phase token/cookie policy, current-user typed context, anonymous behavior, and test-only route
  guard behavior
- unsafe cross-origin refusal and same-origin success through `http.CrossOriginProtection`
- registration/login/logout handlers, safe `422` form outcomes, session cookies, and password
  non-repopulation in rendered forms

There are also small integration-style HTTP tests that use `httptest` and a temporary SQLite database.

## Test Databases

Tests should not read or write the local development database at `./data/book_social_dev.db`.

Current repository and HTTP integration tests create disposable SQLite databases:

- repository tests use in-memory SQLite databases
- HTTP integration tests use temporary SQLite files from `t.TempDir()`
- tests insert minimal deterministic data needed by the behavior under test

Shared helpers in `internal/testutil` build SQLite and PostgreSQL test schemas by applying the
checked-in migrations through private-library migration `000004`, then seed a small deterministic
catalog fixture and ordinary user role. Tests can still keep scenario-specific fixture rows locally
when they need more data than the default helper provides.

PostgreSQL repository tests are opt-in because they require a real PostgreSQL database. By default
they skip unless `BOOK_SOCIAL_POSTGRES_TEST_DSN` is set. The reproducible project check is
`make test/integration`: it starts an isolated Docker Compose database, applies and rolls back the
PostgreSQL migrations with a pinned `golang-migrate` binary, runs PostgreSQL repository and HTTP
parity tests (including the private-library flow) sequentially, then removes its own Compose resources.

The PostgreSQL test DSN must connect to the exact disposable database `book_social_test`. Before
every schema reset, the helper queries `current_database()` and refuses to run unless it equals
`book_social_test` and ends in `_test`. It never prints the DSN or credentials. The repository
tests then drop and recreate only that database's `public` schema before loading fixture data. The
opt-in suite includes user and session repository parity, including the 32-byte token-hash
constraint. When packages share this database, run PostgreSQL tests sequentially with `-p 1`.

Example:

```bash
BOOK_SOCIAL_POSTGRES_TEST_DSN='postgres://book_social:book_social@localhost:5432/book_social_test?sslmode=disable' \
  go test -p 1 ./internal/storage/postgresql
```

For normal local and CI verification, prefer:

```bash
make test/integration
```

The target does not read a user-provided DSN and uses only its own disposable `book_social_test`
database. Docker Compose is its only external prerequisite.

Do not use the full development seed dataset in ordinary unit or handler tests. Use full seed data
only for an explicit seed smoke test or database setup check.

## Testing Guidance

- Prefer table-driven tests.
- Use `httptest` for HTTP handlers.
- Use fake repositories/services for unit tests.
- Avoid browser/e2e tests for now.
- Avoid large HTML snapshot tests.
- Avoid database integration tests unless the task explicitly needs them.
- Test HTTP middleware with `httptest` and exact header/status assertions.
- Use fake servers for shutdown lifecycle tests; do not open a real listener for unit tests.
- Cover both successful and failed static assets, and verify HTML/HTMX responses do not receive a
  public long-lived cache policy.
- Verify panic responses do not expose panic values and that access logs record status 500.
- Do not format credentials, password hashes, raw session tokens, token hashes, or issued cookie
  values into test failure messages.

## Manual Browser Smoke: v0.2.6 Auth Flow

Status: complete as part of v0.2.6 release closure.

Run this smoke test on a local machine, outside the Codex sandbox. It is a release check for the
server-rendered registration, login, logout, navigation, and accessibility flow; it does not replace
the `httptest` coverage above.

Start from a disposable local development database. The reset command deletes the configured SQLite
database, so do not point it at data that must be kept:

```bash
make db/reset
make run
```

Open `http://localhost:8080` in a private browser window or clear site data first. Use a unique login
and email address for the test account if the database was not reset.

1. Open `/me` while anonymous. It must redirect to `/login`; navigation must show Login and Register.
2. Open `/register`. Use only the keyboard: `Tab` must reach every navigation link, form input, and
   Register button in a sensible order; each focused control must have the visible focus outline.
   Labels must describe the focused inputs.
3. Submit the registration form with a required field empty, then with non-matching passwords. The
   page must return a clear inline error next to the relevant field. Tabbing to that field must keep
   its error associated through the accessible description; the error text is explanatory, not a
   separate interactive control. First name, login, and email may remain populated, but neither
   password field may be repopulated.
4. Register with a valid first name, login, email, and matching password of at least 12 characters.
   Expect `/me`, the "Registration complete." flash, the display name, and Logout navigation. Reload
   `/me`: the flash must be gone.
5. Open an unknown route such as `/missing-page`. Expect a 404 page that still shows the authenticated
   navigation and no Login/Register links.
6. Use Logout with the keyboard and confirm the redirect to `/`. Reopen `/me`; it must redirect to
   `/login`. Navigation must again show Login and Register.
7. On `/login`, submit an unknown login and then the registered login with a wrong password. Both
   outcomes must show the same "Invalid login or password." message, leave the password blank, and
   not expose whether the account exists. Submit the valid credentials and confirm `/me`, the
   "You are signed in." flash, and authenticated navigation. Logout once more.
8. In browser DevTools, inspect responses for `/register`, `/login`, `/me`, and `/missing-page`.
   Each dynamic HTML response must include `Cache-Control: no-store`, `Content-Security-Policy`,
   `Permissions-Policy`, `Referrer-Policy`, `X-Content-Type-Options: nosniff`, and
   `X-Frame-Options: DENY`. In local `dev`, HSTS is not expected.
9. Inspect cookies after successful registration or login. `book_social_session` must have `HttpOnly`,
   `Path=/`, and `SameSite=Lax`; `Secure` is intentionally absent for local HTTP development. After
   logout, the session cookie must be cleared. Do not copy cookie values into issue reports or logs.

Record the date, browser/version, commands used, and any failed step in the release evidence. Mark
plan item 66 complete only after this smoke test passes in a real local browser.

## Codex Sandbox Note

Do not start the web server inside the Codex sandbox for verification.

For HTTP behavior, add or update Go tests using `httptest`.
