# Database

Database docs are split by project stage:

- [Database v0.1](database_v0_1.md): baseline schema created by migration `000001`.
- [Database v0.2](database_v0_2.md): normalized catalog schema created by migration `000002`.
- Auth Foundation v0.2.5 is described below and created by migration `000003`.
- Private Library Foundation v0.3.0 is described below and created by migration `000004`.
- Reading-State Lifecycle v0.3.1 is described below and created by migration `000005`.
- Curated Catalog Operations v0.3.2 adds administrator/audit migration `000006` and missing-book
  request migration `000007`. The cover-file extension adds migration `000008`, described below.

Current state:

- `APP_ENV=dev` uses SQLite and is the active local development path.
- `APP_ENV=stage` and `APP_ENV=prod` open PostgreSQL using `APP_DB_DSN`.
- Paired SQLite/PostgreSQL migrations through `000008` exist and can be applied with the
  `golang-migrate` CLI before starting the application; runtime wiring does not apply migrations.
- SQLite and PostgreSQL implement the normalized v0.2 catalog read-side.
- SQLite and PostgreSQL implement equivalent user and opaque-session persistence contracts.
- SQLite and PostgreSQL implement equivalent private-library persistence contracts consumed by the
  implemented v0.3.1 lifecycle flow.
- Both backends implement transactional catalog maintenance/audit and missing-book request creation.
- Docker/Compose has local workflows for SQLite dev and PostgreSQL stage/prod.

## Local Cover References (#28, Step 3)

Apply paired migration `000008_add_local_cover_files` before running this application revision.
It extends `covers` with nullable `storage_key`, `thumbnail_storage_key`, `thumbnail_mime_type`,
`thumbnail_byte_size`, `thumbnail_width`, `thumbnail_height` and `thumbnail_checksum_sha256`.
Existing main-image metadata columns remain in use. Image bytes and absolute filesystem paths
are not stored in SQL. No covers or pilot files are seeded by the migration.

Existing URL-only covers and their metadata/IDs remain unchanged. A local `front` has a canonical
SHA-256 storage key, matching `/media/covers/{key}` URL, complete MIME/size/dimension/checksum
metadata and an optional complete thumbnail group. A thumbnail uses the main image's format and
aspect ratio with strictly smaller dimensions; its URL is derived from its key. Named constraints
enforce this consistency and the prepared-file limits on both backends. `UNIQUE(book_id, variant)`
and existing foreign keys remain unchanged.

`SetLocalFrontCover` accepts typed metadata for already validated/published files and derives the
main URL. It validates reference consistency and upserts only `front`; it does not read files or
authorize the actor. The service/caller must verify the files, authorize the action and use
`WithinCatalogTransaction` to save the reference and audit together. Public detail reads and
operator reads return local keys and optional thumbnail metadata. Replacing with a main file only
clears the old thumbnail. The existing HTTPS URL setter clears all local/derivative and technical
metadata when the URL changes, preserves metadata for the identical URL, and deletes only `front`
when empty. No write physically deletes files.

Step 4 wires file verification and publication into the operator service before starting its SQL
write transaction. The transaction rechecks administrator capability and book existence, then
saves the reference and audit together. Explicit keep skips cover writes; metadata edits therefore
retain a local cover. Audit summaries include main or thumbnail changes, including thumbnail-only
changes at the same main URL. Repository methods remain an internal boundary, not an operator CLI.
See [the operator workflow](cover_storage_v0_3_2.md#operator-workflow-step-4).

The down migration refuses while any local storage reference exists. It preserves external covers
on an allowed downgrade; it never silently strips local keys or thumbnails. Stop cover writers and
back up SQL/files before any downgrade. PostgreSQL takes an exclusive lock on `covers` before
checking; SQLite executes the guard and column removal transactionally via the migration CLI.
A refused down from version `8` targets `7 (dirty)`; inspect actual schema/data before any metadata
repair, following the recovery procedure below. This is not permission to delete references or
run `force` automatically.

## Missing-Book Request Storage

Paired migration `000007_add_missing_book_requests` creates `missing_book_requests` with
`id`, `user_id`, `title`, `author`, `created_at`. Both storage packages now implement the
`books.MissingBookRepository.Create` contract with one parameterized insert. The runtime service
uses the selected backend to save a request before emitting its metadata-only success event.
Apply migration `000007` before using the updated form flow; the service does not create the table.

The user reference is required, with `ON DELETE CASCADE` and `ON UPDATE RESTRICT`. Title is
required normalized text of 1–255 Unicode code points. Author is non-null optional text of 0–255
code points, represented as `''` when absent. Creation time is application-assigned UTC, stored as
RFC3339Nano text in SQLite and `TIMESTAMPTZ` in PostgreSQL. Repeated submissions are separate rows;
there is no title/author/user uniqueness constraint or processing status in v0.3.2.

All fields are non-null; absent author text must be explicitly stored as `''`. Named length checks
apply to both dialects. SQLite also explicitly rejects NUL in title/author, matching PostgreSQL's
text behavior. UTF-8 validation, trimming and assignment of UTC creation time remain application
responsibilities. The migration does not create a default timestamp or seed requests.

The down migration refuses rollback while requests exist. PostgreSQL takes an exclusive table lock
before checking emptiness; SQLite migrations run transactionally through the CLI. An empty table
can be rolled back and recreated without changing users, catalog data or audit records.
Run downgrade only with writers stopped. A guarded refusal may leave migration metadata dirty;
see [recovery after a refused downgrade](#recovery-after-a-refused-downgrade).

The full ownership, normalization and later processing boundary is defined in the
[stored-request contract](catalog_operations_v0_3_2.md#stored-request-model).
Repository writes preserve supplied title/author text, store the supplied user ID and convert the
supplied creation time to UTC. Validation, normalization, session identity and time assignment belong
to the service. Unexpected database failures return only `books.ErrMissingBookInternal`, with no
wrapped driver error or submitted text; finished contexts retain cancellation/deadline errors.
Repositories do not log, emit events or provide operator listing/processing methods. The service
validates and normalizes fields, derives the user from session identity and assigns UTC time. Save
failure returns a generic HTTP `500` without confirmation or a successful-submission event.
Real-database HTTP acceptance passes on SQLite and PostgreSQL. The user confirmed persistence
browser review on 2026-10-25 (Zen Browser, SQLite). See the
[browser checklist](testing.md#manual-browser-smoke-v032-missing-book-persistence-25).

## Migration Layout

SQLite and PostgreSQL migrations live in separate folders because the project keeps
dialect-specific SQL explicit:

```text
db/sqlite/migrations/
db/postgresql/migrations/
```

Migration files use matching sequence numbers where they represent the same domain change:

```text
000001_create_v0_1_schema.up.sql
000001_create_v0_1_schema.down.sql
```

The first migration pair is the v0.1 baseline schema. Migration `000002` normalizes catalog
relationships and adds cover metadata without editing the baseline migration. Migration `000003`
adds the ordinary `user` role and session storage for v0.2.5 without adding demo accounts.
Migration `000004` adds the private `library_items` foundation without repurposing legacy demo
library tables.
Migration `000005` adds the reading-state lifecycle without changing the historical foundation.
Migration `000006` adds the administrator reference role and catalog audit storage.
Migration `000007` adds missing-book request storage used by the connected form/service flow.
Migration `000008` adds local cover references and optional derivative metadata.

## Curated Catalog Operations Schema

Migration `000006_add_catalog_operations` adds `admin` with `is_admin` enabled and
`catalog_audit_log`: `id`, `actor_user_id`, `entity_type`, `entity_id`, `action`, `summary`,
and `occurred_at`. The schema accepts only the `book` entity type and `create`/`correct` actions.
Actor and book foreign keys use `RESTRICT` for deletion and ID updates to retain audit references.
SQLite stores UTC time as RFC3339Nano text; PostgreSQL uses `TIMESTAMPTZ`.

The down migration refuses rollback if audit rows exist or any user has the `admin` role.
It never silently deletes audit history or revokes an appointed administrator. Empty rollback
removes the table and the unassigned role while retaining ordinary users and catalog data.
Stop all application/background writers before downgrade: PostgreSQL migration `000006` does not
take the explicit pre-check table lock used by `000007`. A transaction alone does not make its
emptiness/role check safe against concurrent writes.

The operator service and both repositories now support administrator-only find/create/correct,
service-owned transactions, append-only audit writes, server-generated summaries, and UTC audit
timestamps. A correction input has no slug field. Relationship replacement, front-cover changes,
and the audit entry commit together or roll back together. A successful correction with no changed
groups records an empty summary; otherwise the summary is a sorted comma-separated group list.
Changing a front-cover URL clears its old technical metadata; unchanged URLs retain it, and
other cover variants are preserved. The runtime now exposes the guarded operator search and book
forms through these services for both backends. The schema alone does not prevent a database owner
from updating or deleting audit rows through direct SQL.
The full contract is in [Curated Catalog Operations v0.3.2 Contract](catalog_operations_v0_3_2.md).

### Appoint the First Administrator

Apply migration `000006`, then register the intended account normally. The environment owner uses
the database console to identify that account by its exact login:

```sql
SELECT u.id, u.login, r.role_name
FROM users u
JOIN roles r ON r.id = u.user_role_id
WHERE u.login = 'chosen-login';
```

Verify the returned account and replace `123` below with its positive user ID. Both supported
dialects accept this operation:

```sql
BEGIN;

UPDATE users
SET user_role_id = (SELECT id FROM roles WHERE role_name = 'admin')
WHERE id = 123
    AND user_role_id = (SELECT id FROM roles WHERE role_name = 'user');

SELECT u.id, u.login, r.role_name, r.is_admin
FROM users u
JOIN roles r ON r.id = u.user_role_id
WHERE u.id = 123;
```

Commit with `COMMIT;` only when the result is the intended account with role `admin` and enabled
`is_admin`; otherwise run `ROLLBACK;`. Do not execute an unscoped update or change `user.is_admin`.
This is a deliberate account appointment, not a seed operation or a role-management UI.
Ordinary registration remains bound to `user`, and no administrator account is seeded.
This assignment enables the runtime operator routes at `/admin/catalog`. The guard observes a role
change on the next session request without requiring the user to sign in again. Operator services
also recheck the current capability, inside the transaction for writes.

Run pending SQLite migrations against the default local database:

```bash
make db/migrate/up
```

Roll back the latest SQLite migration:

```bash
make db/migrate/down
```

For PostgreSQL, pass the driver and DSN explicitly:

```bash
make db/migrate/up \
  MIGRATIONS_DIR=./db/postgresql/migrations \
  MIGRATIONS_DATABASE_URL='postgres://user:password@localhost:5432/book_social?sslmode=disable&x-multi-statement=true'

make db/migrate/down \
  MIGRATIONS_DIR=./db/postgresql/migrations \
  MIGRATIONS_DATABASE_URL='postgres://user:password@localhost:5432/book_social?sslmode=disable&x-multi-statement=true'
```

The project uses the installed `migrate` binary from `golang-migrate`. It records applied
versions in `schema_migrations`. The Make targets apply all pending migrations on `up` and roll
back one migration on `down`.

### Safe Downgrade Preconditions

- Identify the exact database/backend, current migration version and compatible application
  revision. Stop all application instances, background jobs and other writers for the entire
  downgrade/recovery window; do not rely only on a maintenance page or database migration lock.
- Take a recoverable backup and record schema/version plus relevant row counts before proceeding.
  Test the operation on a disposable copy first. Do not use reset or seed scripts on retained data.
- Inspect the down migration being executed. `000008` refuses if local cover references exist;
  `000007` refuses if requests exist; `000006` refuses
  if audit rows or appointed administrators exist. These are data-protection guards, not errors to
  bypass by deleting requests/audit or silently demoting accounts.
- If the target schema is incompatible with the retained data, cancel the downgrade and keep the
  compatible application revision. Do not restart writers until the schema and migration metadata
  agree and the compatible application has been verified.

### Recovery After a Refused Downgrade

`golang-migrate` v4.19.1 marks the **target** version dirty before executing a migration. A failed
down from `7` to `6` can therefore show `6 (dirty)` even when the schema/data still match version
`7`. Dirty metadata alone does not prove which schema remains. `force N` only sets the version and
clears the dirty flag; it does not execute SQL, undo partial changes or restore lost data.

1. Keep writers stopped. Preserve the migration error/log and backup before any metadata repair.
   Confirm it was the intended guard refusal, not an unrelated SQL, permission or connection error.
2. Read `schema_migrations` and inspect the actual schema against the checked-in migrations. Verify
   tables, columns, constraints/indexes and relevant data, not just table existence. Check for
   partially applied DDL or leftover downgrade-check tables. Standard CLI transactions are expected
   to preserve data on these guard failures; custom driver/transaction options require inspection.
3. Only if the complete pre-downgrade schema and data are confirmed intact may the environment
   owner use `force` to mark that **verified schema version** clean. Examples:

   | Failed step | Possible dirty marker | Verified schema still present | Conditional repair |
   |-------------|-----------------------|-------------------------------|--------------------|
   | `7 → 6`, request guard refused | `6`, dirty | Full version `7`, retained requests/audit/users/catalog | `force 7` |
   | `6 → 5`, audit/admin guard refused | `5`, dirty | Full version `6`, retained audit/admin/users/catalog | `force 6` |

   For multi-step downgrades, earlier steps may have succeeded: choose the verified schema version,
   not automatically the original starting version. If schema/data are partial or uncertain, stop;
   investigate on a copy and use an approved repair or restore plan. Do not guess a version.
4. Recheck version/dirty state and retained data after an authorized repair. Do not immediately
   retry the blocked down migration. Cancel the downgrade, verify the compatible application and
   only then resume service; resolve any intended data transition separately.

Read-only checks for the default SQLite database include:

```bash
sqlite3 -readonly ./data/book_social_dev.db 'SELECT version, dirty FROM schema_migrations;'
sqlite3 -readonly ./data/book_social_dev.db '.schema'
```

For PostgreSQL, use `SELECT version, dirty FROM schema_migrations`, schema inspection in `psql`
(`\d+ missing_book_requests`, `\d+ catalog_audit_log`, `\d+ roles`) and compare with the checked-in
migrations. Inspect relevant request/audit counts and user-role assignments on both backends without
copying private request text into reports.

The following SQLite example is **not a generic fix**: run it only after step 3 has established
that a refused `7 → 6` left the full version-7 schema/data intact, and the owner authorizes repair:

```bash
migrate -path ./db/sqlite/migrations -database 'sqlite://./data/book_social_dev.db' force 7
migrate -path ./db/sqlite/migrations -database 'sqlite://./data/book_social_dev.db' version
```

For PostgreSQL the same `force N` semantics apply, using its migration directory and the exact
verified database URL; never reuse the SQLite example or switch targets during recovery.

### Migration CLI Drivers

The installed `migrate` binary must include the database driver being used. Local SQLite
migrations require a binary with the SQLite driver; PostgreSQL migrations require the PostgreSQL
driver. Check the installed binary with `migrate -help`.

If the SQLite driver is missing, rebuild the CLI with the project drivers:

```bash
go install -tags 'sqlite postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.19.1
```

The PostgreSQL baseline migration contains multiple SQL statements, so the PostgreSQL migration
URL should include `x-multi-statement=true`.

## Auth Foundation Schema

Migration `000003_add_user_role_and_sessions` has matching SQLite and PostgreSQL variants. It:

- ensures the non-administrator `user` role exists;
- creates `sessions` with `user_id`, `token_hash`, `created_at`, and `expires_at`;
- stores only a 32-byte SHA-256 token hash, never the raw browser token;
- enforces unique token hashes and `expires_at > created_at`;
- cascades session deletion when its user is deleted;
- indexes `expires_at` for lazy expiry cleanup.

The down migration removes sessions and the reference role, but refuses rollback when user rows
still depend on that role. SQLite stores UTC timestamps in the repository's RFC3339Nano format;
PostgreSQL uses `TIMESTAMPTZ`. Both repositories treat missing or expired sessions as an
unauthenticated state and map unexpected database failures to the domain's internal error.

Migration `000003` creates the ordinary role as reference data; the seed repeats it with
`ON CONFLICT DO NOTHING` for a migrated database. The development seed does not create a user or
session.

In the v0.2.6 registration flow, user creation, ordinary-role lookup, and initial hashed-token
session creation use one database transaction. A failed session write rolls back the user creation.

## Private Library Foundation Schema

Migration `000004_add_library_items` has matching SQLite and PostgreSQL variants. It creates
`library_items` with `user_id`, `book_id`, and `added_at` plus a generated integer ID. It:

- enforces one item per `user_id + book_id` with `uq_library_items_user_book`;
- cascades item deletion when its owner or catalog book is deleted;
- indexes `(user_id, added_at DESC, id DESC)` for the owner-scoped, deterministic list query;
- stores UTC timestamps as RFC3339Nano text in SQLite and `TIMESTAMPTZ` in PostgreSQL.

The down migration refuses to remove a non-empty `library_items` table. This protects private user
data during rollback; an empty migration can be rolled back normally. The legacy `library`,
`shelves`, and `tags` tables remain demo data and are unrelated to this model.

## Reading-State Lifecycle Schema

Migration `000005_add_library_item_lifecycle` extends `library_items` with a constrained reading
status, nullable `started_at` and `finished_at`, and a positive integer `version` for optimistic
locking. Existing rows become `want_to_read` with unset lifecycle timestamps and `version = 1`; new
rows use the same initial values.

Status changes and removal include the owner ID and expected version in their write condition.
An update increments the version only when it changes state; a repeat of the current state is a
no-op. A stale update or removal is reported as a conflict rather than silently overwriting or
deleting private data.

CI runs Go tests, `go vet`, and lint. It does not run database migrations or Docker Compose.
The local migration and seed smoke check is:

```bash
make db/migrate/smoke
```

It verifies a clean migration plus seed, creation and empty rollback of `library_items`, migration
of one v0.1 catalog row, and the documented down-migration path. CI does not run this target or
Docker Compose workflows yet.

## Reset And Seed

`reset` means recreating a local database from scratch. It is destructive and should only be
used for local development or disposable test data.

`seed` means loading deterministic sample/reference data after the schema exists. The current
seed SQL is development data, not production data. It is expected to run after a fresh schema
or reset; it is not treated as a repeatable data migration.

The seed's catalog provenance, its separation from test fixtures, and the cover-metadata record
are documented in [Curated Demo Catalog Provenance](demo_catalog_provenance.md).
The dataset contains 109 books, 24 authors, 23 genres, and no covers. Fresh bootstrap assigns book
IDs in slug order in both dialects. This does not renumber existing databases: the seed is only for
fresh disposable databases. Stable public identifiers remain the book slugs.

Reset scripts destroy only disposable local database state, run all migrations up, and then apply
the matching seed SQL.

For local SQLite reset:

```bash
make db/reset
```

This runs `db/sqlite/reset-dev-db.sh`, which removes the configured SQLite database file,
applies all SQLite migrations, and then applies `db/sqlite/seed.sql`.

For manual PostgreSQL reset, use `db/postgresql/reset-dev-db.sh` with PostgreSQL environment
variables such as `PGHOST`, `PGPORT`, `PGDATABASE`, `PGUSER`, and `PGPASSWORD`.

For Docker database reset, choose the environment you want to recreate.

Dev SQLite:

```bash
make compose/dev/down
make compose/dev/up
```

Stage PostgreSQL:

```bash
make compose/stage/down
make compose/stage/up
```

Prod PostgreSQL:

```bash
make compose/prod/down
make compose/prod/up
```

For a disposable PostgreSQL database, `db/postgresql/reset-dev-db.sh` drops and recreates the
`public` schema, applies all PostgreSQL migrations, and then applies `db/postgresql/seed.sql`.
The PostgreSQL migration URL used by the script includes `x-multi-statement=true`.

For a PostgreSQL database, apply migrations manually:

```bash
MIGRATIONS_DIR=./db/postgresql/migrations \
MIGRATIONS_DATABASE_URL='postgres://user:password@localhost:5432/book_social?sslmode=disable&x-multi-statement=true' \
make db/migrate/up
psql "$APP_DB_DSN" -f db/postgresql/seed.sql
```

## Test Databases

Tests do not use the local development database file.

Current SQLite repository and HTTP integration tests create temporary or in-memory SQLite
databases inside the test process and exercise the normalized catalog, auth, and private-library
persistence, catalog operations/audit and missing-book requests. The shared library helpers apply
all checked-in migrations (currently through `000008`) and create a minimal deterministic catalog
fixture; ordinary role, session, library, audit and request constraints have focused tests.

This keeps tests fast and isolated without depending on the full development seed dataset.
PostgreSQL repository tests are opt-in and exercise the same catalog and user/session repository
contracts.
The helper opens the configured disposable database, resets `public`, and applies the normalized
catalog fixture. Set `BOOK_SOCIAL_POSTGRES_TEST_DSN` to run PostgreSQL tests. Because helpers
reset one shared schema, run packages sequentially with `go test -p 1 ./...` when using one DSN.
