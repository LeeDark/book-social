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

## HTTP Lifecycle And Response Policy

The web process listens for `SIGINT` and `SIGTERM` through a root `signal.NotifyContext`. On
shutdown it stops accepting new connections and waits up to five seconds for the HTTP server to
finish active requests. `http.ErrServerClosed` is treated as a normal result; listener and
shutdown errors are returned to the process entrypoint.

The HTTP server keeps these transport timeouts:

- read timeout: 10 seconds;
- write timeout: 35 seconds (application timeout plus a five-second response margin);
- idle timeout: 60 seconds.

Router middleware is applied in this order:

```text
SecurityHeaders -> RequestID -> TrustedRealIP -> request logger -> Recoverer -> CrossOriginProtection -> route handler
```

`TrustedRealIP` is disabled unless `APP_TRUSTED_PROXY_CIDRS` is configured. When configured, it
accepts forwarded client-IP headers only from an immediate peer in those networks.

`CrossOriginProtection` rejects unsafe cross-origin browser requests before route handlers and does
not use insecure bypass patterns. The v0.2.5 route tree still contains no auth form handlers;
current-user middleware and the route guard are foundations for v0.2.6 rather than global session
work on static and health requests.

The 30-second application timeout is applied only to dynamic MPA pages. Health checks, static
files, and the fallback 404 route do not use it. Static assets have a one-hour public cache on
successful responses; missing or failed assets use `no-store`. HTML and HTMX partial responses do
not receive a public long-lived cache policy.

Responses include conservative browser security headers. The CSP allows self-hosted assets and
HTTPS/data cover images, while HSTS is intentionally omitted because local development uses HTTP.
Recovered panics produce a generic 500 response and structured diagnostics in the application log.
Panic and internal failures return generic 500 responses; details are written to server logs only.
Template output is buffered before its status is committed so rendering failures can become 500
responses safely.

Reset local SQLite database:

```bash
make db/reset
```

This is destructive. It removes the configured disposable SQLite database, applies all migrations,
and loads `db/sqlite/seed.sql`.

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

Run the disposable SQLite migration, seed, legacy-data, and down checks with:

```bash
make db/migrate/smoke
```

## Configuration

Current environment variables:

- `APP_ENV`, allowed values `dev`, `stage`, `prod`; default `dev`
- `APP_HTTP_ADDR`, default `:8080`
- `APP_TRUSTED_PROXY_CIDRS`, optional comma-separated trusted proxy networks (for example
  `10.0.0.0/8,192.168.0.0/16`); forwarded client-IP headers are ignored when unset
- `APP_DB_DSN`, default `./data/book_social_dev.db`; use a SQLite DSN for `dev` and a PostgreSQL DSN for `stage` or `prod`
- `APP_LOG_LEVEL`, default `debug`
- `APP_LOG_FORMAT`, default `text`

Auth foundation configuration currently has one central non-environment setting:

- `Config.Auth.SessionLifetime`, fixed to the seven-day `DefaultSessionLifetime`.

The cookie manager uses the same seven-day default when no explicit lifetime is supplied. Its
policy defaults to `book_social_session`, `Path=/`, `HttpOnly`, and `SameSite=Lax`. `Secure` remains
an explicit environment/wiring decision: it must be enabled for HTTPS stage/prod, while local HTTP
development is the documented exception. Production cookie wiring begins with v0.2.6.

`APP_ENV=test` is not a supported runtime environment. Tests build their own configuration
and temporary SQLite databases where needed.

## Auth Foundation Lifecycle

The v0.2.5 token and session sequence is deliberately split:

```text
GenerateToken
  -> SHA-256 hash
  -> persist DB session with seven-day absolute expiry
  -> Set browser cookie
```

If token generation or persistence fails, no cookie should be exposed. Current-user middleware
hashes the cookie value before lookup and puts only minimal identity into typed request context.
Invalid or expired sessions clear browser state and continue anonymously; unexpected store failures
return a generic `500`. Logout/session invalidation and the real register/login handlers remain
v0.2.6 work.

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
- SQLite and PostgreSQL implement the normalized v0.2 catalog read-side, including multi-author,
  multi-genre, and book-cover metadata reads.
- PostgreSQL databases are initialized through migrations plus seed in the reset and Compose
  bootstrap paths.

Initialize a disposable PostgreSQL database with migrations and seed:

```bash
PGHOST=localhost PGPORT=5432 PGDATABASE=book_social PGUSER=book_social \
  db/postgresql/reset-dev-db.sh
```

For a disposable local PostgreSQL database, `db/postgresql/reset-dev-db.sh` drops and recreates
the `public` schema, then applies all PostgreSQL migrations and the seed file. The PostgreSQL
migration URL must include `x-multi-statement=true`.

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

Run the opt-in PostgreSQL repository suite against a disposable database:

```bash
BOOK_SOCIAL_POSTGRES_TEST_DSN='postgres://user:password@localhost:5432/book_social_test?sslmode=disable' \
  GOCACHE=/tmp/book-social-go-cache go test -p 1 ./internal/storage/postgresql
```

The test helper resets the database's `public` schema only after confirming that the connected
database is exactly `book_social_test` (and has the required `_test` suffix); it refuses any other
database without printing the DSN or credentials. Never point this variable at non-disposable data.
When one test database is shared across packages, include `-p 1` in the test command.

## Docker And Compose

Docker and Compose are supported as local environment workflows for the v0.2 bootstrap.

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

CI does not start Docker Compose services and does not run migration smoke tests yet. The current
boundary is intentional: migration smoke remains a local/manual check until the `migrate` CLI and
database setup are made reproducible in CI. A PostgreSQL service job is deferred.

For local Docker Compose work, use reset/bootstrap for disposable seeded environments:

- `make compose/dev/up` starts SQLite dev and initializes the database from all SQLite migrations plus `db/sqlite/seed.sql` when the volume is empty.
- `make compose/stage/up` starts PostgreSQL stage and initializes from all PostgreSQL migrations plus `db/postgresql/seed.sql` when the database is empty.
- `make compose/prod/up` does the same for a local prod-like PostgreSQL environment.

The Docker reset/bootstrap workflow and the migration workflow now use the same migration-first
ordering. Use `make db/migrate/smoke` for a disposable SQLite verification.

## SQLite In Docker

Dev Compose mounts a named volume at:

```text
/app/data
```

On container startup, `docker/entrypoint.sh` checks the configured SQLite database path. If the database file is missing or empty, it runs all SQLite migrations and then:

```text
db/sqlite/seed.sql
```

Reset and seed the Docker database from scratch:

```bash
make compose/dev/down
make compose/dev/up
```

This removes the Compose volume, then lets the entrypoint initialize a fresh seeded SQLite database.

## PostgreSQL In Docker

Stage and prod Compose mount named PostgreSQL data volumes. On first start, the app entrypoint
applies all PostgreSQL migrations and, when the catalog is empty, loads:

```text
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
