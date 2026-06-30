# Development

## Requirements

- Go 1.25
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

Generate Templ experiment files:

```bash
make templ-generate
```

Reset local SQLite database:

```bash
make db-dev-reset
```

Open local SQLite database:

```bash
make db-dev-shell
```

## Configuration

Current environment variables:

- `APP_ENV`, default `dev`
- `APP_HTTP_ADDR`, default `:8080`
- `APP_DB_DSN`, default `./data/book_social_dev.db`
- `APP_LOG_LEVEL`, default `debug`
- `APP_LOG_FORMAT`, default `text`

## Docker And Compose

Docker and Compose are supported as a basic local development setup for v0.1.

They are not production-ready infrastructure. Do not treat this setup as deployment guidance.

Build the image:

```bash
docker build --progress=plain -t book-social:dev .
```

Start the app:

```bash
docker compose up --build
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
db/sqlite/schema_v0_1_sqlite.sql
db/sqlite/seed_sqlite.sql
```

Reset and seed the Docker database from scratch:

```bash
docker compose down -v
docker compose up --build
```

This removes the Compose volume, then lets the entrypoint initialize a fresh seeded SQLite database.

## Out Of Scope

- Production deployment setup.
- Health checks.
- Reverse proxy configuration.
- Orchestration.
- Kubernetes cleanup.
