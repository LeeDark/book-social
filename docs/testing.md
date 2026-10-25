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
- private-library lifecycle transitions, timestamp rules, optimistic-lock conflicts, idempotent
  repeated status submissions, explicit removal confirmation, malformed lifecycle forms, and
  owner-scoped refusal
- administrator-only catalog forms, literal search, immutable slugs, safe audit records, and
  atomic book/link/cover/audit changes, including real SQLite/PostgreSQL HTTP acceptance
- missing-book validation, session-owned persistence, UTC time, safe events/errors, one-request
  confirmation and refusals without writes, including real SQLite/PostgreSQL HTTP acceptance

There are also small integration-style HTTP tests that use `httptest` and a temporary SQLite database.

## Test Databases

Tests should not read or write the local development database at `./data/book_social_dev.db`.

Current repository and HTTP integration tests create disposable SQLite databases:

- repository tests use in-memory SQLite databases
- HTTP integration tests use temporary SQLite files from `t.TempDir()`
- tests insert minimal deterministic data needed by the behavior under test

Shared helpers in `internal/testutil` build SQLite and PostgreSQL test schemas by applying the
checked-in migrations through missing-book request migration `000007`, then seed a small deterministic
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

### Development Seed Acceptance (#23)

Dedicated development-seed tests apply the checked-in migrations and full seed to disposable
databases. Ordinary fixture helpers remain independent. SQLite tests compare two fresh catalogs,
including IDs, descriptions, relationships, and cover metadata, and check every book against the
provenance inventory. HTTP acceptance covers catalog reads, filters, book/author pages, cover
placeholder fallback, and authenticated addition of a seeded book to the private library.

Run the focused SQLite checks:

```bash
GOCACHE=/tmp/book-social-go-cache go test -v ./internal/testutil ./internal/app -run SQLiteDevelopmentSeed
make db/migrate/smoke
```

The opt-in PostgreSQL checks compare the full catalog with SQLite, verify identity sequences after
seed, and run the same HTTP acceptance. Use only the disposable `book_social_test` database; its
`public` schema is reset. Run packages sequentially:

```bash
BOOK_SOCIAL_POSTGRES_TEST_DSN='postgres://book_social:book_social@localhost:5432/book_social_test?sslmode=disable' \
  GOCACHE=/tmp/book-social-go-cache go test -p 1 -v -count=1 ./internal/testutil ./internal/app -run PostgresDevelopmentSeed
```

`make test/integration` includes these tests and provisions its own isolated PostgreSQL database.
The dataset and source limitations are in [Demo Catalog Provenance](demo_catalog_provenance.md).

## v0.3.2 Acceptance Coverage (#26)

Scope review on 2026-10-25 includes the accepted persistence expansion of #25. Missing-book
submission must save before confirming success; only external lookup/import and operator request
listing/processing are excluded. No additional implementation or test gaps were identified in this
criteria-to-coverage review; this does not replace the final release check run.

| Criterion | Existing coverage |
|-----------|-------------------|
| Paired migrations `000006`/`000007`, constraints and guarded rollback | `internal/testutil/catalog_operations_migration_test.go`, `missing_book_requests_migration_test.go` |
| Reproducible demo seed, backend parity, identity sequences and provenance inventory | `internal/testutil/development_seed_test.go`; public catalog/library HTTP checks in `internal/app/development_seed_test.go` |
| Administrator access, ordinary-user refusal, promotion/demotion and cross-origin refusal | `internal/app/admin_authorization_test.go`; service refusals in `internal/modules/books/operator_test.go` |
| Create/correct/no-op, immutable public slugs, normalized relationships/cover metadata and atomic audit | `internal/testutil/catalog_operator_test.go`, `internal/app/catalog_operator_integration_test.go`; transaction-step failures in `internal/modules/books/operator_test.go` |
| Accessible operator forms, retained errors, statuses, flash and slug suggestions | `internal/app/catalog_operator_test.go`, `internal/web/tests/catalog-operator.test.cjs`; user-confirmed operator browser review below |
| Normalized missing-book fields, session-owned identity, UTC, save-before-event/confirmation and safe failures | `internal/modules/books/missing_book_test.go`, `missing_book_handler_test.go`; `internal/testutil/missing_book_repository_test.go`, `internal/app/missing_book_integration_test.go` |
| Missing-book validation, escaping, refused submissions without writes, flash lifetime and refresh without repeat writes | `internal/app/missing_book_routes_test.go`, `missing_book_integration_test.go`; user-confirmed persistence browser review below |

The database acceptance suites have SQLite and opt-in PostgreSQL variants. Use
`make test/integration` for the isolated PostgreSQL run; ordinary focused runs skip PostgreSQL
without its disposable DSN. Provenance checks verify the documented inventory, not external
authorship/reuse rights: legacy metadata sources remain unverified and the seed contains no covers.

### Final Automated Check Run (2026-10-25)

All checks below passed against implementation commit `78fe3ad`, with the local #26 documentation
changes still uncommitted. No implementation or test changes were required.

```bash
GOCACHE=/tmp/book-social-go-cache make test
GOCACHE=/tmp/book-social-go-cache go vet ./...
make db/migrate/smoke
GOCACHE=/tmp/book-social-go-cache make test/integration
node internal/web/tests/catalog-operator.test.cjs
GOTOOLCHAIN=go1.26.4 GOCACHE=/tmp/book-social-go-cache GOLANGCI_LINT_CACHE=/tmp/book-social-lint-cache ./bin/golangci-lint run ./... --config=./.golangci.yml
git ls-files '*.go' | xargs gofmt -l
git diff --check
```

`make test` passed with the race detector; its opt-in PostgreSQL cases skipped without a DSN.
The separate integration target passed with real isolated PostgreSQL, migration up/down/up and
sequential repository/HTTP acceptance. Its container and network were removed; development databases
were not changed. SQLite migration/seed/legacy-data smoke passed, all four JavaScript tests passed,
lint reported `0 issues`, and formatting/diff checks reported no findings.

`gh pr checks 27` confirmed `quality` and `integration-postgres` passed for published head
`78fe3ad` ([CI run](https://github.com/LeeDark/book-social/actions/runs/40187677075)). This is not CI
verification of uncommitted documentation: recheck CI after the final commit/push. Existing manual
acceptance is recorded separately below; no new browser run or release closure is claimed here.

### Consolidated Manual Acceptance (#26)

The accepted user reports below cover the v0.3.2 operator and missing-book flows. Dates identify
the original reviews, not new runs performed during release closure.

| Review date | Accepted scope and checklist | Reported environment |
|-------------|------------------------------|----------------------|
| 2026-10-12 | [Operator #24](#manual-browser-smoke-v032-operator-24): access refusal, appointment/demotion, find/create/correct, immutable public URL, retained accessible errors, slug suggestions and flashes | Browser/version, database backend and exact launch commands not separately reported |
| 2026-10-15 | [Original feedback #25](#manual-browser-smoke-v032-missing-book-feedback-25): authentication, form/keyboard validation, confirmation, empty-state links, no-store and safe events; no request persistence at that time | Browser/version, database backend and exact launch commands not separately reported |
| 2026-10-25 | [Persistence #25](#manual-browser-smoke-v032-missing-book-persistence-25): form/validation, stored normalized title/optional author, session owner/UTC time, no repeat write on refresh, distinct repeated submissions, safe events and survival after restart | Zen Browser, SQLite; browser version and exact launch commands not separately reported |

The October 15 report alone does not establish persistence acceptance; that is supplied by the
October 25 report. All three are user-confirmed evidence, not browser checks run by Codex.
PostgreSQL has automated acceptance evidence, not a reported manual browser review. The reports
do not establish production deployment or a fresh release-wide browser run on `78fe3ad`.

The #26 coverage review and automated checks required no implementation changes, so no additional
manual run is required for this step. If a subsequent fix changes behavior, repeat only the
affected checklist locally outside the sandbox and record its date, environment and outcome.
Documentation closure, final CI confirmation after commit/push and PR readiness remain separate.

## Testing Guidance

Operator service tests cover normalization, permission refusal, immutable slugs, controlled UTC
audit metadata, no-op corrections, and transaction failures. Shared SQLite/PostgreSQL operator
acceptance exercises literal Unicode search, existing-author/genre selection, public read
compatibility, cover variant preservation, and rollback of book/link/cover changes on audit failure.
PostgreSQL fixture identity sequences are advanced after explicit IDs, allowing generated-ID
operator writes. These tests are included in `make test/integration`.

Focused checks:

```bash
GOCACHE=/tmp/book-social-go-cache go test ./internal/modules/books ./internal/testutil -run 'Operator|SuggestSlug'
```

Operator HTTP tests use a fake service and the real router/templates to cover form GETs, retained
POST values, `422`/`409`/`404`/`403`/`500` outcomes, immutable correction slugs, fixed one-request
flashes, anonymous/ordinary-user refusal, and cross-origin POST protection. JavaScript suggestion
tests use Node's built-in test/assert modules with a small DOM double, not a browser or dependency.

```bash
GOCACHE=/tmp/book-social-go-cache go test ./internal/app -run OperatorHTTP -count=1
node internal/web/tests/catalog-operator.test.cjs
```

For manual review, run the application outside the Codex sandbox, appoint a test administrator
using [the documented procedure](database.md#appoint-the-first-administrator), and open
`/admin/catalog`, `/admin/catalog/books/new`, and `/admin/catalog/books/{valid-id}/edit`.
Check title-to-slug suggestions and manual overrides, keyboard navigation/error links, retained
values after a duplicate slug or invalid HTTPS URL, success messages after saving, and the same
public `/books/{slug}` URL after correction. With JavaScript disabled, supply a slug manually.
Unknown IDs should show `404`; malformed IDs should show `422`. The detailed checklist below
records the completed user-confirmed manual #24 browser review; it is not covered by the DOM double.

DB-backed operator HTTP acceptance exercises real sessions, guard, router, templates, services,
repositories and migrations for both backends. It verifies create/correct/no-op audit counts and
UTC timestamps, normalization, conflict/validation refusals with unchanged data, the permanent
public URL, one-request flashes, safe logs, and rollback after the final audit write fails.

```bash
GOCACHE=/tmp/book-social-go-cache go test ./internal/app -run CatalogOperatorHTTPAcceptance -count=1
GOCACHE=/tmp/book-social-go-cache make test/integration
```

The first command runs SQLite and skips PostgreSQL without the opt-in DSN; the second provisions
the isolated PostgreSQL test database and runs the parity suite sequentially with the race detector.

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

## Manual Browser Smoke: v0.3.2 Operator (#24)

Status: complete. On 2026-10-12 the user confirmed that the local browser review passed and accepted
#24 after clarifying malformed-ID `422` behavior and optional author/genre selections. This is
user-reported manual evidence, not a browser run by Codex; browser version and backend were not
separately recorded. Automated HTTP acceptance independently passes on SQLite and PostgreSQL.
Missing-book feedback (#25) was separately reviewed and accepted on 2026-10-15; this operator
checklist records only #24 and does not itself close v0.3.2.

Run the app locally outside the Codex sandbox with migrations through `000006`. Use an ordinary
registered test account and deliberately appoint it through the procedure in `docs/database.md`.
Use a disposable database for this smoke; writes and audit history are intentionally persistent.
Do not run reset or appoint a production account just to complete a test.

- [x] Anonymous access to `/admin/catalog` redirects to `/login`; an ordinary signed-in account
  receives `403`. After appointment, the same session can open the operator pages.
- [x] `/admin/catalog` initially shows only the find form. Search a known title/slug and an absent
  title; check edit/public links and actionable empty-state text.
- [x] `/admin/catalog/books/new` proposes an ASCII slug while typing a title. Manually edit the
  slug and confirm later title edits do not overwrite it. Try a Unicode-only title and enter a
  valid slug yourself. With JavaScript disabled, manually supplied slugs still work.
- [x] Submit an invalid HTTPS URL, then a duplicate slug; check `422`/`409`, retained text and
  selections, linked error messages, labels, and keyboard-only navigation.
- [x] Create a unique book with selected existing authors/genres and an optional permitted cover
  URL. Check `303` to `/admin/catalog/books/{new-id}/edit`, one success message, then reload to
  confirm the message disappears. The public `/books/{new-slug}` page opens.
- [x] Correct title/description/selections and clear the front cover; verify the same public URL
  shows the changes. The edit form has no mutable slug input. Save again without changes and
  confirm a single success message still appears.
- [x] `/admin/catalog/books/not-an-id/edit` returns `422`; a positive ID absent from the test
  database returns `404`. Demoting the test account to `user` makes the next request return `403`.

The operator roadmap item is marked complete following the user's review and acceptance.

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

## Manual Browser Smoke: v0.3.1 Lifecycle

Status: pending external review and corrections.

After the v0.3.1 automated checks and review pass, use a reset local SQLite database and a real
browser to verify `/books`, `/books/{valid-slug}`, and `/me/library`. Register two users and confirm
that the owner can add a book, move it through Want to read → Reading → Read → Want to read, repeat
a displayed state without changing it, and remove it only from the explicit confirmation page. The
second user must neither see nor mutate the first user's item. Record the date, browser/version,
commands, and result before closing the issue or milestone.

## Manual Browser Smoke: v0.3.2 Missing-Book Feedback (#25)

Status: complete. On 2026-10-15 the user confirmed that the manual browser checklist passed and
accepted #25. This is user-reported evidence, not a browser run by Codex. Browser version, database
backend and exact launch command were not separately reported. Automated service and HTTP checks
passed independently; this confirmation does not establish a new PostgreSQL integration run.

Run the app locally outside the Codex sandbox with `make run` and an already migrated development
database. Use an ordinary registered account; administrator privileges are unnecessary. Record
the date, browser/version, database backend, and result when rerunning this checklist.

- [x] While anonymous, open `/books` and follow "Request a missing book". Expect `/login`;
  direct access to `/books/request` must also redirect to login.
- [x] Sign in and follow the link from `/books`. Verify the form has a required title, optional
  author, a submit button and a back-to-catalog link. Use `Tab` and `Shift+Tab` to reach controls
  with visible focus, then submit using the keyboard.
- [x] Submit a blank or whitespace-only title. Expect `422`, an error summary linked to the title,
  an inline error associated with the input, and retained author text. Repeat with a field longer
  than 255 characters; correct the field and resubmit. No confirmation should appear on rejection.
- [x] Submit a valid title, first without an author and then with one. Expect `303 /books` and the
  confirmation that does not promise the book will be added. Reload `/books`: the flash is gone.
- [x] Check the catalog's empty state using `/books?author=missing-book-smoke-no-match`. The request
  link must remain available; also follow an author/genre filter that has no matching books to
  check the HTMX partial when such a combination exists in the local catalog.
- [x] Inspect response headers for the form, validation response, redirect and catalog:
  `Cache-Control: no-store` must be present. Check one structured event per successful submission;
  fields are `request_id`, `actor_id`, `title_length`, `author_length`, `has_author`, with no submitted
  title/author or session-token values. Do not copy credentials or cookie values into reports.

Cross-origin refusal, expired-session behavior, malformed/oversized bodies, Unicode boundaries,
HTML escaping and exclusion of private request data are covered by `httptest`. The accepted browser
checklist above predates persistence. The updated flow now saves through the configured repository
before its success event/flash/redirect; a failed save returns a generic `500` without confirmation.
Apply migration `000007` before renewed browser review. No external catalog service is called.

### Automated Persistence Checks

The accepted checklist above covers the original non-persistent #25 flow. The user subsequently
agreed to save title, optional author and submitting user in v0.3.2. Migration `000007` is implemented
with shared SQLite/PostgreSQL checks for text limits, required fields, user references, cascade and
guarded rollback. Shared repository tests verify both backends' stored fields, optional author,
UTC time, repeated submissions, parameterized text, safe database errors and finished contexts.
Service/runtime wiring is implemented. Fake-based service/HTTP tests verify normalized fields,
session identity, UTC assignment, save-before-event ordering, refusals without writes and generic
save-failure responses without success events/flashes. Shared real-session HTTP acceptance now
passes on SQLite and PostgreSQL: normalized persisted fields, session-owned identity, UTC time,
refusals without new rows/events, one-request confirmation, GET refresh without another write, and
injected database failures without data loss or diagnostic/text leaks. The user confirmed renewed
browser review on 2026-10-25 (Zen Browser, SQLite); the checklist above records the earlier flow.
Operator listing and processed/rejected decisions belong to v0.4.3.

Run focused HTTP acceptance with:

```bash
GOCACHE=/tmp/book-social-go-cache go test ./internal/app -run MissingBookHTTPAcceptance -count=1 -v
```

It runs SQLite and skips PostgreSQL without a disposable DSN. `make test/integration` runs both
backend scenarios with the race detector using its isolated PostgreSQL database. Failure injection
adds rejecting triggers only inside disposable test fixtures; application schema is unchanged.

- [x] Verify paired SQLite/PostgreSQL migrations and repository round trips for title, optional
  author and submitting user, using disposable databases.
- [x] Verify that a valid authenticated POST saves the normalized fields with the session user's
  ID, ignoring any submitted user/actor ID, before returning the success redirect and flash.
- [x] Verify invalid, anonymous and cross-origin submissions save no request; a failed save creates
  neither a success confirmation nor a successful-submission event. Logs continue to exclude text.
- [x] Complete the renewed browser checklist below and record acceptance.

## Manual Browser Smoke: v0.3.2 Missing-Book Persistence (#25)

Status: complete. On 2026-10-25 the user reported that all manual checks passed in Zen Browser with
SQLite, confirming persistence acceptance. Browser version and exact launch commands were not
separately reported. This is user-reported evidence, not a browser run by Codex or a PostgreSQL
browser check. Automated SQLite/PostgreSQL acceptance passed on 2026-10-23. Operator listing and
processing are not part of v0.3.2; saved requests are inspected with read-only SQL.

Run locally outside the Codex sandbox. The commands below use the default development SQLite
database; do not reset it. If using another database, keep migration, application and inspection
paths consistent. Apply pending migrations, then start the app:

```bash
make db/migrate/up
APP_DB_DSN=./data/book_social_dev.db APP_LOG_FORMAT=json make run
```

In another terminal, confirm schema version `7` with `dirty = 0`, note the initial request count
and find the ID of the ordinary account used for the check (replace the example login):

```bash
sqlite3 -readonly ./data/book_social_dev.db 'SELECT version, dirty FROM schema_migrations;'
sqlite3 -readonly ./data/book_social_dev.db 'SELECT COUNT(*) FROM missing_book_requests;'
sqlite3 -readonly ./data/book_social_dev.db "SELECT id, login FROM users WHERE login = 'your-test-login';"
```

Inspect saved rows after each submission:

```bash
sqlite3 -readonly -header -column ./data/book_social_dev.db \
  'SELECT id, user_id, title, author, created_at FROM missing_book_requests ORDER BY id DESC LIMIT 10;'
```

- [x] Open `http://localhost:8080/books/request` anonymously: expect `/login`, no new row.
  Sign in with the ordinary account and open the form from `/books`.
- [x] Repeat the original keyboard, validation and no-store checks above. A whitespace-only title
  must produce `422`, retain the author and leave the row count unchanged, without confirmation.
- [x] Submit a recognizable test title with surrounding spaces and no author. Expect `303 /books`
  and one confirmation. Verify exactly one new row: trimmed title, empty author, the signed-in
  account's ID and a current UTC timestamp. Refresh `/books`: no flash and no further row.
- [x] Submit another title with an author, both surrounded by spaces. Verify another single row
  with both values trimmed. Submit the same valid data again: it intentionally creates a separate
  request, not a duplicate rejection. Restart the application and verify the rows remain.
- [x] Check one `catalog_missing_book_requested` event per successful POST, with only request/actor
  IDs, lengths and author presence, never title/author values. Log envelope fields such as time and
  level are expected. Do not inject failures into the development database: disposable tests cover
  safe `500` responses, no success confirmation and preservation of existing requests.
- [x] Record user-reported acceptance: 2026-10-25, Zen Browser, SQLite, all checks passed.
  Browser version and exact launch commands were not separately reported. Do not include cookies,
  credentials or other users' submitted text.

The requests created by this smoke remain in the development database. Earlier metadata-only events
cannot be restored as requests. A PostgreSQL browser run can use the same behavioral checks with its
own migrated local database and read-only SQL; automated PostgreSQL acceptance is already recorded.

## Codex Sandbox Note

Do not start the web server inside the Codex sandbox for verification.

For HTTP behavior, add or update Go tests using `httptest`.
