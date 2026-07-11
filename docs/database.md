# Database

Database docs are split by project stage:

- [Database v0.1](database_v0_1.md): active SQLite development schema.
- [Database v0.2 target](database_v0_2.md): planned normalized target schema.

Current state:

- `APP_ENV=dev` uses SQLite and is the active local development path.
- `APP_ENV=stage` and `APP_ENV=prod` open PostgreSQL using `APP_DB_DSN`.
- There is no migration system yet.
- PostgreSQL has a connection package and v0.1 book repository implementation.
- Docker/Compose has local workflows for SQLite dev and PostgreSQL stage/prod.

## Reset And Seed

`reset` means recreating a local database from scratch. It is destructive and should only be
used for local development or disposable test data.

`seed` means loading deterministic sample/reference data after the schema exists. The current
seed SQL is development data, not production data. It is expected to run after a fresh schema
or reset; it is not treated as a repeatable data migration.

There is no migration runner yet. The current reset scripts apply the v0.1 schema SQL and then
the matching seed SQL directly.

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
PostgreSQL repository tests are opt-in. Set `BOOK_SOCIAL_POSTGRES_TEST_DSN` to a disposable
PostgreSQL database DSN to run them. Those tests drop and recreate the `public` schema before
loading minimal fixture data.
