# Database

Database docs are split by project stage:

- [Database v0.1](database_v0_1.md): active SQLite development schema.
- [Database v0.2 target](database_v0_2.md): planned normalized target schema.

Current state:

- `APP_ENV=dev` uses SQLite and is the active local development path.
- `APP_ENV=stage` and `APP_ENV=prod` open PostgreSQL using `APP_DB_DSN`.
- Baseline migration files exist and can be applied with the `golang-migrate` CLI.
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

The first migration pair is the v0.1 baseline schema. Future v0.2 schema changes should add
new numbered migration pairs instead of editing the baseline migration.

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

Reset and bootstrap scripts do not use migrations yet.

CI runs Go tests, `go vet`, and lint. It does not run database migrations or Docker Compose
workflows yet. Migration smoke checks should be run locally against disposable databases.

## Reset And Seed

`reset` means recreating a local database from scratch. It is destructive and should only be
used for local development or disposable test data.

`seed` means loading deterministic sample/reference data after the schema exists. The current
seed SQL is development data, not production data. It is expected to run after a fresh schema
or reset; it is not treated as a repeatable data migration.

The current reset scripts apply the v0.1 schema SQL and then the matching seed SQL directly.
Later, reset should destroy only disposable local database state, run all migrations up, and then
apply seed data.

For local SQLite reset:

```bash
make db/reset
```

This runs `db/sqlite/reset-dev-db.sh`, which removes the configured SQLite database file,
applies `db/sqlite/schema_v0_1.sql`, and then applies `db/sqlite/seed.sql`.

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

For a PostgreSQL database, apply the v0.1 SQL files manually for now:

```bash
psql "$APP_DB_DSN" -f db/postgresql/schema_v0_1.sql
psql "$APP_DB_DSN" -f db/postgresql/seed.sql
```

## Test Databases

Tests do not use the local development database file.

Current SQLite repository and HTTP integration tests create temporary or in-memory SQLite
databases inside the test process. Shared helpers in `internal/testutil` create SQLite test
databases, apply the minimal catalog schema, and optionally seed a small deterministic catalog
fixture.

This keeps tests fast and isolated without depending on the full development seed dataset.
PostgreSQL repository tests are opt-in. Shared helpers in `internal/testutil` open the configured
test database, drop and recreate the `public` schema, and apply the minimal PostgreSQL catalog test
schema. Set `BOOK_SOCIAL_POSTGRES_TEST_DSN` to a disposable PostgreSQL database DSN to run them.
