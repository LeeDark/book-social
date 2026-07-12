# Development

## Requirements

- Go 1.26
- SQLite CLI for database reset and shell commands
- `golang-migrate` CLI for database migrations, built with SQLite and PostgreSQL drivers
- Docker and Docker Compose for container workflows

## Common Commands

Build:

```bash
make build
```

Run:

```bash
make run
```

Test:

```bash
make test
```

Generate Templ files:

```bash
make templ/generate
```

Reset local SQLite database:

```bash
make db/reset
```

This is destructive. It recreates the configured SQLite database from
`db/sqlite/schema_v0_1.sql` and `db/sqlite/seed.sql`.

Open local SQLite database:

```bash
make db/shell
```

Apply pending SQLite migrations:

```bash
make db/migrate/up
```

Roll back the latest SQLite migration:

```bash
make db/migrate/down
```

The installed `migrate` binary must include the SQLite database driver for local SQLite
migrations. Check supported drivers with:

```bash
migrate -help
```

If `sqlite` is missing from the `Database drivers` list, rebuild the CLI with the project
drivers:

```bash
go install -tags 'sqlite postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.19.1
```

## Configuration

Current environment variables:

- `APP_ENV`, allowed values `dev`, `stage`, `prod`; default `dev`
- `APP_HTTP_ADDR`, default `:8080`
- `APP_DB_DSN`, default `./data/book_social_dev.db`; use a SQLite DSN for `dev` and a PostgreSQL DSN for `stage` or `prod`
- `APP_LOG_LEVEL`, default `debug`
- `APP_LOG_FORMAT`, default `text`

`APP_ENV=test` is not a supported runtime environment. Tests build their own configuration
and temporary SQLite databases where needed.

Set variables for one command:

```bash
APP_ENV=dev APP_DB_DSN='./data/book_social_dev.db' make run
```

Export variables for the current terminal session:

```bash
export APP_ENV=stage
export APP_DB_DSN='postgres://user:password@localhost:5432/book_social?sslmode=disable'
make run
```

Check what the current shell will pass to the app:

```bash
echo "$APP_ENV"
echo "$APP_DB_DSN"
```

Runtime database selection:

- `APP_ENV=dev` opens `APP_DB_DSN` with the SQLite driver.
- `APP_ENV=stage` and `APP_ENV=prod` open `APP_DB_DSN` with the PostgreSQL driver.
- The v0.1 book repository behavior is implemented for both SQLite and PostgreSQL.
- PostgreSQL databases can be initialized from schema SQL or with the `golang-migrate` CLI.

Initialize a PostgreSQL database:

```bash
psql "$APP_DB_DSN" -f db/postgresql/schema_v0_1.sql
psql "$APP_DB_DSN" -f db/postgresql/seed.sql
```

For a disposable local PostgreSQL database, `db/postgresql/reset-dev-db.sh` drops and recreates
the `public` schema, then applies the PostgreSQL schema and seed files.

Apply PostgreSQL migrations manually:

```bash
make db/migrate/up \
  MIGRATIONS_DIR=./db/postgresql/migrations \
  MIGRATIONS_DATABASE_URL='postgres://user:password@localhost:5432/book_social?sslmode=disable&x-multi-statement=true'
```

Rollback the latest PostgreSQL migration manually:

```bash
make db/migrate/down \
  MIGRATIONS_DIR=./db/postgresql/migrations \
  MIGRATIONS_DATABASE_URL='postgres://user:password@localhost:5432/book_social?sslmode=disable&x-multi-statement=true'
```

## Docker And Compose

Docker and Compose are supported as local environment workflows for v0.1.

They are not production-ready infrastructure. Do not treat the `prod` Compose workflow as
deployment guidance; it only runs the app with `APP_ENV=prod` locally.

Build the image:

```bash
make docker/build
```

Start the dev app with SQLite:

```bash
make compose/dev/up
```

Open:

```text
http://localhost:8080
```

The dev Compose environment sets:

```text
APP_ENV=dev
APP_HTTP_ADDR=:8080
APP_DB_DSN=file:/app/data/book_social_dev.db
```

Start the stage app with PostgreSQL:

```bash
make compose/stage/up
```

Start the prod app with PostgreSQL:

```bash
make compose/prod/up
```

Both PostgreSQL Compose workflows run a local `postgres` service and set `APP_DB_DSN`
to the matching container database.

## CI, Compose, And Migrations

CI currently runs code quality checks only:

```text
go test ./...
go vet ./...
golangci-lint
```

CI does not start Docker Compose services and does not run migration smoke tests yet.

For local Docker Compose work, use reset/bootstrap for disposable seeded environments:

- `make compose/dev/up` starts SQLite dev and initializes the database from `db/sqlite/schema_v0_1.sql` plus `db/sqlite/seed.sql` when the volume is empty.
- `make compose/stage/up` starts PostgreSQL stage and initializes from `db/postgresql/schema_v0_1.sql` plus `db/postgresql/seed.sql` when the volume is empty.
- `make compose/prod/up` does the same for a local prod-like PostgreSQL environment.

Use migration commands separately when you want to test migration files against a disposable
database. The migration workflow and the Docker reset/bootstrap workflow are not unified yet.

## SQLite In Docker

Dev Compose mounts a named volume at:

```text
/app/data
```

On container startup, `docker/entrypoint.sh` checks the configured SQLite database path. If the database file is missing or empty, it runs:

```text
db/sqlite/schema_v0_1.sql
db/sqlite/seed.sql
```

Reset and seed the Docker database from scratch:

```bash
make compose/dev/down
make compose/dev/up
```

This removes the Compose volume, then lets the entrypoint initialize a fresh seeded SQLite database.

## PostgreSQL In Docker

Stage and prod Compose mount named PostgreSQL data volumes. On first start, the official
PostgreSQL image initializes the database from:

```text
db/postgresql/schema_v0_1.sql
db/postgresql/seed.sql
```

Reset and seed the stage PostgreSQL database from scratch:

```bash
make compose/stage/down
make compose/stage/up
```

Reset and seed the prod PostgreSQL database from scratch:

```bash
make compose/prod/down
make compose/prod/up
```

## Out Of Scope

- Production deployment setup.
- Health checks.
- Reverse proxy configuration.
- Orchestration.
- Kubernetes cleanup.
