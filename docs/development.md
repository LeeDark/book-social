# Development

## Requirements

- Go 1.26
- SQLite CLI for database reset and shell commands

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

Open local SQLite database:

```bash
make db/shell
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
- PostgreSQL databases must be initialized manually for now; there is no migration runner yet.

Initialize a PostgreSQL database:

```bash
psql "$APP_DB_DSN" -f db/postgresql/schema_v0_1.sql
psql "$APP_DB_DSN" -f db/postgresql/seed.sql
```

## Docker And Compose

Docker and Compose are supported as a basic local development setup for v0.1.

They are not production-ready infrastructure. Do not treat this setup as deployment guidance.

Build the image:

```bash
make docker/build
```

Start the app:

```bash
make docker/up
```

Open:

```text
http://localhost:8080
```

The Compose environment sets:

```text
APP_ENV=dev
APP_HTTP_ADDR=:8080
APP_DB_DSN=file:/app/data/book_social_dev.db
```

## SQLite In Docker

Compose mounts a named volume at:

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
make docker/down
make docker/up
```

This removes the Compose volume, then lets the entrypoint initialize a fresh seeded SQLite database.

## Out Of Scope

- Production deployment setup.
- Health checks.
- Reverse proxy configuration.
- Orchestration.
- Kubernetes cleanup.
