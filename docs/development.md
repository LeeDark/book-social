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
- `APP_LOG_LEVEL`, default `debug`
- `APP_LOG_FORMAT`, default `text`

Important current limitation:

- `cmd/web/main.go` currently opens `./data/book_social_dev.db` directly.
- `APP_DB_DSN` exists in Compose, but is not currently used by runtime wiring.

## Docker And Compose

Docker and Compose are experimental for now.

Known limitations:

- Runtime database path/configuration needs cleanup.
- Template/static asset behavior should be verified before treating the image as a normal run target.
- Kubernetes files are early experiments and intentionally out of scope for current documentation cleanup.

For local development, prefer:

```bash
make db-dev-reset
make run
```
