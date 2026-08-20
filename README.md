# Book Social

Book Social is a learning Go web project for building a small server-rendered book catalog.

The project is intentionally simple: modular monolith, layered architecture, SQLite for local development, and MPA pages rendered on the server.

## Current Status

Current v0.2.5 Auth Foundation on top of the v0.2.4 HTTP foundation and normalized catalog:

- Home and About pages.
- Book catalog page.
- Book details page.
- Author page.
- Catalog filtering by author slug and genre slug.
- Server-rendered templates with simple CSS.
- Local SQLite schema and seed data.
- Unit tests and small HTTP/integration-style tests.
- HTMX catalog filter spike as progressive enhancement.
- Environment-based database selection: SQLite for `dev`, PostgreSQL for `stage` and `prod`.
- SQLite and PostgreSQL implement the same normalized catalog read contract.
- `golang-migrate` CLI wiring and v0.1/v0.2 migrations for SQLite and PostgreSQL.
- Normalized catalog tables for many-to-many authors/genres and URL-only cover metadata.
- Catalog cards and book details render all related authors and genres.
- Book details display the `front` cover URL when available and a CSS placeholder otherwise.
- The home page uses the shared catalog-card read model for a deterministic featured selection.
- Migration-first reset and Docker/Compose bootstrap followed by development seed data.
- Local migration/seed smoke checks and v0.2 SQLite/PostgreSQL test helpers.
- Graceful shutdown on `SIGINT`/`SIGTERM` with bounded active-request draining.
- Request IDs, structured access logs, trusted-proxy-aware client IP handling, panic recovery, and
  conservative security headers.
- Dynamic-route timeout, static asset cache policy, and buffered template error handling.
- SQLite and PostgreSQL migration `000003` add the ordinary `user` role and DB-backed opaque
  sessions with 32-byte token hashes and absolute expiry.
- The `users` module provides registration validation, bcrypt password hashing and verification,
  neutral credential refusal, and SQLite/PostgreSQL user/session repositories.
- Session creation, loading, expiry, and invalidation use a seven-day absolute lifetime; only the
  token hash reaches persistence.
- The HTTP auth foundation provides two-phase token/cookie handling, typed current-user context,
  and a testable protected-route guard.
- Global `http.CrossOriginProtection` rejects unsafe cross-origin browser requests without adding
  CSRF tokens to forms.

Not current production direction:
- Docker and Docker Compose are supported as local environment workflows, not production infrastructure.
- User-facing registration/login/logout, production `/me`, auth navigation, and flashes are planned
  for v0.2.6; the current UI does not claim that authentication is available.
- User libraries, search, pagination, and social features are planned later.

## Tech Stack

- Go 1.26
- chi router
- `html/template`
- SQLite via `modernc.org/sqlite`
- PostgreSQL driver via `github.com/lib/pq`
- bcrypt via `golang.org/x/crypto/bcrypt`
- `golang-migrate` CLI for schema migrations, built with SQLite and PostgreSQL drivers
- Pico CSS plus project CSS
- HTMX vendored locally for a small catalog filter spike

## Run Locally

Reset the local development database:

```bash
make db/reset
```

Run the web app:

```bash
make run
```

Default address:

```text
http://localhost:8080
```

Local development defaults to:

```text
APP_ENV=dev
APP_DB_DSN=./data/book_social_dev.db
```

`APP_TRUSTED_PROXY_CIDRS` is optional. When set to comma-separated trusted proxy networks, forwarded
client-IP headers are accepted only from an immediate peer in one of those networks.

Set environment variables before the command when you need a different configuration:

```bash
APP_ENV=dev APP_DB_DSN='./data/book_social_dev.db' make run
```

## Run With Docker

Docker/Compose provides local environment workflows for the v0.2 database bootstrap.
It is not production deployment infrastructure.

Build the image:

```bash
make docker/build
```

Start the dev app with SQLite:

```bash
make compose/dev/up
```

Start the stage app with PostgreSQL:

```bash
make compose/stage/up
```

Start the prod app with PostgreSQL:

```bash
make compose/prod/up
```

Open:

```text
http://localhost:8080
```

The dev Compose setup stores SQLite data in a named volume mounted at `/app/data`.
On first start, the container applies all SQLite migrations and seeds
`/app/data/book_social_dev.db` if it is missing or empty. The stage and prod Compose setups run a
local PostgreSQL container initialized by the app entrypoint with all PostgreSQL migrations and
`db/postgresql/seed.sql`.

Reset the Docker SQLite database:

```bash
make compose/dev/down
make compose/dev/up
```

Useful routes:

```text
/
/about
/books
/books?author=jane-austen
/books?genre=classic
/books/{book-slug}
/authors/{author-slug}
```

## Test

```bash
make test
```

In constrained environments, this command is preferred over starting a real web server.

## Project Structure

```text
cmd/web/                 application entrypoint
internal/app/            app wiring, routes, home handler
internal/modules/books/  books/catalog module
internal/modules/users/  auth/user/session service and repository contracts
internal/storage/sqlite/ SQLite repository implementation
internal/storage/postgresql/ PostgreSQL connection and repository implementation
internal/http/auth/      session cookie, current-user context, and route-guard foundation
internal/http/           rendering, response helpers, middleware, view models
internal/web/            server templates and static assets
db/sqlite/               local SQLite schema, migrations, seed, reset script
db/postgresql/           PostgreSQL schema, migrations, seed, reset script
docs/                    project documentation
docs/ai/                 AI-agent context, task history, spike notes
```

## Documentation

- [Architecture](docs/architecture.md)
- [Development](docs/development.md)
- [Routes](docs/routes.md)
- [Domain model](docs/domain.md)
- [Database v0.1](docs/database_v0_1.md)
- [Database v0.2](docs/database_v0_2.md)
- [MPA auth contract](docs/drafts/auth-contract.ru.md)
- [Testing](docs/testing.md)
- [Roadmap](docs/roadmap.md)
- [Technical backlog](docs/backlog.md)
- [AI project context](docs/ai/project-context.md)

## Roadmap Summary

Near-term work:
- Start v0.2.6 Registration/Login/Logout on the accepted v0.2.5 foundation.
- Keep Docker/Compose as local environment workflows; do not add production deployment claims yet.

v0.2 direction:
- Quality baseline: format/test/lint/CI.
- Database strategy: migrations and schema evolution (v0.2.2 complete).
- Catalog read model updates for the v0.2 schema (v0.2.3 complete).
- HTTP foundation: lifecycle, middleware, security, caching, recovery, and timeout policy (v0.2.4 complete).
- Auth foundation: password policy, DB sessions, current-user/guard boundary, and cross-origin
  browser protection (v0.2.5 complete).
- Registration/login/logout forms, production `/me`, navigation, and flashes are the next active
  scope (v0.2.6).
