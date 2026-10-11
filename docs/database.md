# Database

Database docs are split by project stage:

- [Database v0.1](database_v0_1.md): baseline schema created by migration `000001`.
- [Database v0.2](database_v0_2.md): normalized catalog schema created by migration `000002`.
- Auth Foundation v0.2.5 is described below and created by migration `000003`.
- Private Library Foundation v0.3.0 is described below and created by migration `000004`.
- Reading-State Lifecycle v0.3.1 is described below and created by migration `000005`.

Current state:

- `APP_ENV=dev` uses SQLite and is the active local development path.
- `APP_ENV=stage` and `APP_ENV=prod` open PostgreSQL using `APP_DB_DSN`.
- Baseline, catalog-normalization, auth-foundation, private-library, and lifecycle migration files
  exist and can be applied with the `golang-migrate` CLI.
- SQLite and PostgreSQL implement the normalized v0.2 catalog read-side.
- SQLite and PostgreSQL implement equivalent user and opaque-session persistence contracts.
- SQLite and PostgreSQL implement equivalent private-library persistence contracts consumed by the
  implemented v0.3.1 lifecycle flow.
- Docker/Compose has local workflows for SQLite dev and PostgreSQL stage/prod.

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

## Curated Catalog Operations Schema

Migration `000006_add_catalog_operations` adds `admin` with `is_admin` enabled and
`catalog_audit_log`: `id`, `actor_user_id`, `entity_type`, `entity_id`, `action`, `summary`,
and `occurred_at`. The schema accepts only the `book` entity type and `create`/`correct` actions.
Actor and book foreign keys use `RESTRICT` for deletion and ID updates to retain audit references.
SQLite stores UTC time as RFC3339Nano text; PostgreSQL uses `TIMESTAMPTZ`.

The down migration refuses rollback if audit rows exist or any user has the `admin` role.
It never silently deletes audit history or revokes an appointed administrator. Empty rollback
removes the table and the unassigned role while retaining ordinary users and catalog data.

The operator service and both repositories now support administrator-only find/create/correct,
service-owned transactions, append-only audit writes, server-generated summaries, and UTC audit
timestamps. A correction input has no slug field. Relationship replacement, front-cover changes,
and the audit entry commit together or roll back together. A successful correction with no changed
groups records an empty summary; otherwise the summary is a sorted comma-separated group list.
Changing a front-cover URL clears its old technical metadata; unchanged URLs retain it, and
other cover variants are preserved. Operator handlers are not yet wired into the runtime.
The schema alone does not prevent
a database owner from updating or deleting audit rows through direct SQL.
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
At the authorization checkpoint this assignment does not yet enable operator HTTP routes: runtime
operator handlers are still pending. The guard observes a role change on the next session request.

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
persistence. The shared library helper applies migrations through `000005`, creates a deterministic
catalog fixture, and verifies the ordinary role, session, and library constraints.

This keeps tests fast and isolated without depending on the full development seed dataset.
PostgreSQL repository tests are opt-in and exercise the same catalog and user/session repository
contracts.
The helper opens the configured disposable database, resets `public`, and applies the normalized
catalog fixture. Set `BOOK_SOCIAL_POSTGRES_TEST_DSN` to run PostgreSQL tests. Because helpers
reset one shared schema, run packages sequentially with `go test -p 1 ./...` when using one DSN.
