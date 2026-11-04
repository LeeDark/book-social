# v0.3.3 Reliability and Release Closure: Plan

Status: planned on 2026-11-04; implementation has not started.

This document is the execution plan for
[v0.3.3 — Reliability and Release Closure](roadmap.md#v033--reliability-and-release-closure).
The roadmap remains the source of truth for scope. The plan adds the code findings, stage order,
planning decisions and verification for each roadmap item, and the code hygiene that
[backlog.md](backlog.md#v033-with-release-closure) links to this release.

## Outcome

The complete private-library flow is releasable and operable as a curated demonstration: it is
verified end to end on SQLite and PostgreSQL, critical operations are visible in structured logs,
and the release is closed through a written checklist.

Not in this release: deployment and a database-aware readiness check (v0.3.5), interface polish
and the registration error summary (v0.3.4), a monitoring stack, and new dependencies.

## Issues

| Stage | Issue | Title |
|---|---|---|
| 2 | [#33](https://github.com/LeeDark/book-social/issues/33) | Reject out-of-range private-library item IDs and versions with 422 |
| 3 | [#34](https://github.com/LeeDark/book-social/issues/34) | Prevent a new login from matching another account's email address |
| 4 | [#35](https://github.com/LeeDark/book-social/issues/35) | Store SQLite timestamps in a fixed-width UTC text format |
| 5 | [#36](https://github.com/LeeDark/book-social/issues/36) | Lock tables in PostgreSQL guarded downgrades 000003–000006 |
| 6 | [#37](https://github.com/LeeDark/book-social/issues/37) | Record safe library-operation logs and product events |
| 7 | [#38](https://github.com/LeeDark/book-social/issues/38) | Add the critical-flow end-to-end HTTP set and complete migration smoke checks |
| 8 | [#39](https://github.com/LeeDark/book-social/issues/39) | Run the v0.3.3 accessibility smoke check |
| 9 | [#40](https://github.com/LeeDark/book-social/issues/40) | Complete v0.3.3 code hygiene, release checklist and documentation |

Stage 1 is a baseline check without code changes and has no issue.

## Starting Point

Findings from the code at the start of the release (`11a7360`):

| Area | Current behavior |
|---|---|
| Library IDs | `internal/modules/library/handler.go:182,190` parses `version` and `itemID` with `strconv.Atoi`. On PostgreSQL, values above `2147483647` become `500`. Operator routes already use `ParseInt(…, 10, 32)` (`internal/modules/books/operator_handler.go:122`). |
| Login and email | `normalizeRegistrationInput` (`internal/modules/users/service.go:155`) checks only the login length. `FindCredentials` uses `WHERE login = ? OR email = ? LIMIT 1` without `ORDER BY` on both dialects, so the returned row is undefined. |
| SQLite time | `formatSQLiteTime` writes `RFC3339Nano`, which drops trailing zeros: as text, `…:05Z` sorts after `…:05.5Z`. This affects `expires_at > ?`, `ORDER BY added_at` and the SQLite `CHECK (expires_at > created_at)` in migration `000003`. |
| PostgreSQL downgrades | `000003`–`000006` check through a temporary table without a lock. `000007` and `000008` already take `LOCK TABLE … IN ACCESS EXCLUSIVE MODE`. |
| Logs | The request logger records `request_id`. One product event exists, `catalog_missing_book_requested` (`internal/modules/books/missing_book.go:53`), and serves as the pattern. Library operations log nothing. |
| Tests | `internal/app/app_integration_test.go` checks library routes on both dialects, but as separate checks rather than one sequential flow. No test uses `httptest.NewServer`. |
| Smoke and CI | `make db/migrate/smoke` rolls back only 5 migrations and does not run in CI. CI runs `go test ./...` without `-race`; `make test/integration` rehearses only `down 1` on PostgreSQL. |

## Planning Decisions

These are recommended defaults. Confirm each one in its issue before the affected stage starts.

1. **PostgreSQL downgrades:** add `LOCK TABLE` to the released down files `000003`–`000006`.
   golang-migrate does not checksum migrations and down files run only during a downgrade, so the
   change is safe. Keep "stop all writers and take a backup before any downgrade" in the release
   checklist as well. Record the deliberate change to released migrations in
   `docs/ai/task-history.md`.
2. **Login with `@`:** reject `@` in new logins and prefer an email match in `FindCredentials`
   (`ORDER BY CASE WHEN email = ? THEN 0 ELSE 1 END`), so existing conflicting accounts no longer
   block the email owner. Do not add a schema `CHECK` constraint.
3. **Existing SQLite time values:** no data migration. Old values stay readable; they can differ
   only within the same second, and old sessions expire on their own. Document the limitation.
4. **First book addition:** detect it in the same transaction as the insert (`NOT EXISTS` before
   `INSERT`). Adding a book to an emptied library counts as a first addition again.
5. **Application error model:** the v0.3 Engineering Scope item is not assigned to a sub-release.
   Close it with documentation: the operation result types `ok`, `invalid`, `not_found`,
   `conflict`, `unauthorized`, `forbidden` and `internal` are the small error model. Module
   sentinel errors stay as they are.
6. **Delivery:** as in v0.3.2, one release branch (`feat/v0.3.3-reliability-release-closure`), one
   commit per stage, and one PR at release closure.

## Stages

### 1. Baseline

- Run `make test`, `make lint`, `make db/migrate/smoke`, `make test/integration` and the
  JavaScript tests on the starting revision.
- Record pre-existing failures so they are not confused with new ones.

### 2. Library ID range (#33)

- Parse `itemID` and `version` with `ParseInt(…, 10, 32)` and a positive check; values outside
  `1..2147483647` return `422` before the service is called.
- Hygiene: `library.Service.Add` resolves a book ID by slug through a narrow repository lookup on
  both dialects instead of loading authors, genres and covers.
- Tests: table-driven handler tests (`0`, `-1`, `abc`, `2147483648`, `9999999999`) for status
  change, removal confirmation and removal; PostgreSQL integration returns `422`, not `500`.
- Done when both dialects return the same statuses and the Engineering Scope item about
  equivalent SQLite/PostgreSQL semantics can be checked.

### 3. Login must not match another account's email (#34)

- First check every retained database (local SQLite development database, any retained PostgreSQL
  database) for logins containing `@`. Record counts only, without personal data.
- Add a `login` validation error for `@`, a hint on the registration form and the email-first
  ordering in `FindCredentials` on both dialects.
- Hygiene: remove `users.Service.Register` and the then-unused `WithinTransaction` from the
  interface and both repositories; require `RegistrationSessionStore` in the `NewService`
  parameter type instead of a type assertion at run time.
- Tests: normalization, the `422` handler outcome, and a repository test on both dialects where one
  account's login equals another account's email.

### 4. Fixed-width SQLite time (#35)

- Write UTC with microsecond precision, matching PostgreSQL:
  `Truncate(time.Microsecond).Format("2006-01-02T15:04:05.000000Z07:00")`. Keep reading with
  `RFC3339Nano` so old values still parse.
- Move the shared helpers out of `users_repository.go` into their own file in the `sqlite`
  package.
- Tests: same-second ordering and session expiry, reading old-format values, and the
  `expires_at > created_at` check.

### 5. PostgreSQL downgrade locks (#36)

- Lock `users` and `sessions` in `000003`, `library_items` in `000004` and `000005`, and
  `catalog_audit_log` and `users` in `000006`.
- Extend `make test/integration` with a full round trip on the empty test database: all up, all
  down, all up again.

### 6. Operation logs and product events (#37)

- Library operation logs for `add`, `update_status` and `remove`: `operation`, `result`,
  `duration`, `request_id`, `actor_id`. Never log a book slug or ID together with a user ID:
  library contents are private.
- Product events: registration, first book addition, status change (`from` and `to`), and the
  existing missing-book event.
- Document the fields and events; the same documentation closes decision 5.
- Tests: an `slog` handler writing to a buffer; assert that required fields are present and
  sensitive fields are absent.

### 7. End-to-end HTTP flow, smoke and PostgreSQL path (#38)

- One sequential flow test: register → log out → log in → catalog → book page → add → `reading` →
  `read` → removal confirmation → remove → log out → `/me/library` redirects to login. Include
  user isolation and `422` for an out-of-range ID.
- Drive the full router with middleware through `ServeHTTP` and a small cookie-keeping helper;
  avoid `httptest.NewServer` so the test also runs where listening sockets are not allowed (Codex
  sandbox).
- The test builds the application separately from `cmd/web/main.go`; document this drift risk
  instead of refactoring the shared composition.
- Extend `make db/migrate/smoke`: every seeded table, a full down/up round trip on an empty
  database, and a rollback of the seeded database only as far as the downgrade guards allow.
- Optional: run the SQLite smoke in CI (needs `migrate` built with the `sqlite` tag and the
  `sqlite3` CLI).

### 8. Accessibility smoke (#39)

- Routes: `/`, `/books`, `/books/{slug}`, `/authors/{slug}`, `/register`, `/login`, `/me`,
  `/me/library`, `/me/library/{id}/remove`, `/books/request/new` and the `/admin/catalog` pages.
- Check keyboard-only use, visible focus, field labels, errors linked to fields, page titles and
  headings, cover `alt` text and 200 % zoom.
- Fix blocking problems in this release; record the rest for v0.3.4. The user performs the browser
  review locally, and the results go to `docs/testing.md`.

### 9. Remaining hygiene, documentation and closure (#40)

- Hygiene: remove `templ` and `gomponents` from `depguard`; re-panic `http.ErrAbortHandler` in
  `Recoverer` (with a test); call `stop()` after `ctx.Done()` in `cmd/web/main.go` so a second
  interrupt ends the process.
- Add `docs/release_checklist.md`: checks (`make audit`, smoke, integration, JavaScript tests,
  `coverctl verify`, manual browser smoke), stopping writers and taking a backup before any
  downgrade, PR, green CI on the published head, tag `v0.3.3` on the merge commit, and closing the
  issues.
- Update `routes.md`, `domain.md`, `database.md`, `testing.md`, `development.md`, `README.md`,
  `docs/ai/project-context.md`, `docs/ai/task-history.md`, both roadmaps, and remove completed items
  from `backlog.md`.

## Order and Dependencies

```text
1 → {2, 3, 4, 5} → 6 → 7 → 8 → 9
```

Stages 2–5 are independent. Stage 7 follows them so the end-to-end flow pins the final behavior.
Stage 9 documentation can grow alongside the other stages; release closure is the last step.

## Definition of Done

- [ ] All seven v0.3.3 roadmap items and the linked hygiene are done. The open v0.3 Definition of
  Done, Engineering Scope and Observability Baseline items are checked or explicitly moved with a
  reason.
- [ ] `make test` (with `-race`), lint, the SQLite smoke, `make test/integration` and CI on the
  published head pass.
- [ ] The user has confirmed the manual browser review (critical flow and accessibility), and the
  evidence is recorded in `docs/testing.md`.
- [ ] The release checklist was followed, the PR is merged, tag `v0.3.3` is set, and the roadmaps
  and backlog are updated.

## Risks

- The test composition of the application can drift from `cmd/web/main.go`.
- Old SQLite time values can still compare differently within the same second.
- Released down migrations change; this deliberate exception to "do not edit historical
  migrations" must be recorded in `docs/ai/task-history.md`.
