# Database

Database docs are split by project stage:

- [Database v0.1](database_v0_1.md): baseline schema created by migration `000001`.
- [Database v0.2](database_v0_2.md): normalized catalog schema created by migration `000002`.
- Auth Foundation v0.2.5 is described below and created by migration `000003`.
- Private Library Foundation v0.3.0 is described below and created by migration `000004`.

Current state:

- `APP_ENV=dev` uses SQLite and is the active local development path.
- `APP_ENV=stage` and `APP_ENV=prod` open PostgreSQL using `APP_DB_DSN`.
- Baseline, catalog-normalization, auth-foundation, and private-library migration files exist and
  can be applied with the `golang-migrate` CLI.
- SQLite and PostgreSQL implement the normalized v0.2 catalog read-side.
- SQLite and PostgreSQL implement equivalent user and opaque-session persistence contracts.
- SQLite and PostgreSQL implement equivalent private-library persistence contracts consumed by the
  implemented v0.3.0 HTTP add/list flow.
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

## Planned Reading-State Lifecycle Schema

The next paired SQLite/PostgreSQL migration for v0.3.1 will extend `library_items` with a constrained
reading status, nullable `started_at` and `finished_at`, and a positive integer `version` for
optimistic locking. Existing rows will become `want_to_read` with unset lifecycle timestamps and
`version = 1`; new rows will use the same initial values. This is an accepted schema contract, not
the current v0.3.0 database shape.

Status changes and removal will include the owner ID and expected version in their write condition.
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
persistence. The shared library helper applies migrations through `000004`, creates a deterministic
catalog fixture, and verifies the ordinary role, session, and library constraints.

This keeps tests fast and isolated without depending on the full development seed dataset.
PostgreSQL repository tests are opt-in and exercise the same catalog and user/session repository
contracts.
The helper opens the configured disposable database, resets `public`, and applies the normalized
catalog fixture. Set `BOOK_SOCIAL_POSTGRES_TEST_DSN` to run PostgreSQL tests. Because helpers
reset one shared schema, run packages sequentially with `go test -p 1 ./...` when using one DSN.
