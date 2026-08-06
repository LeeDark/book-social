# Database

Database docs are split by project stage:

- [Database v0.1](database_v0_1.md): baseline schema created by migration `000001`.
- [Database v0.2](database_v0_2.md): normalized catalog schema created by migration `000002`.

Current state:

- `APP_ENV=dev` uses SQLite and is the active local development path.
- `APP_ENV=stage` and `APP_ENV=prod` open PostgreSQL using `APP_DB_DSN`.
- Baseline and catalog-normalization migration files exist and can be applied with the
  `golang-migrate` CLI.
- PostgreSQL has a connection package and v0.1 book repository implementation.
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
relationships and adds cover metadata without editing the baseline migration.

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

CI runs Go tests, `go vet`, and lint. It does not run database migrations or Docker Compose
The local migration and seed smoke check is:

```bash
make db/migrate/smoke
```

It verifies a clean migration plus seed, migration of one v0.1 catalog row, and the documented
down-migration path. CI does not run this target or Docker Compose workflows yet.

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
databases inside the test process and exercise the normalized v0.2 catalog read-side. The shared
v0.2 helper creates the normalized catalog schema and a deterministic multi-relation fixture.

This keeps tests fast and isolated without depending on the full development seed dataset.
PostgreSQL repository tests are opt-in and exercise the same normalized v0.2 catalog read-side.
The helper opens the configured disposable database, resets `public`, and applies the normalized
catalog fixture. Set `BOOK_SOCIAL_POSTGRES_TEST_DSN` to run PostgreSQL tests. Because helpers
reset one shared schema, run packages sequentially with `go test -p 1 ./...` when using one DSN.
