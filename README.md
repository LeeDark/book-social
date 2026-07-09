# Book Social

Book Social is a learning Go web project for building a small server-rendered book catalog.

The project is intentionally simple: modular monolith, layered architecture, SQLite for local development, and MPA pages rendered on the server.

## Current Status

Current v0.1 baseline:
- Home and About pages.
- Book catalog page.
- Book details page.
- Author page.
- Catalog filtering by author slug and genre slug.
- Server-rendered templates with simple CSS.
- Local SQLite schema and seed data.
- Unit tests and small HTTP/integration-style tests.
- HTMX catalog filter spike as progressive enhancement.
- Environment-based database selection: SQLite for `dev`, PostgreSQL connection setup for `stage` and `prod`.

Not current production direction:
- Templ and gomponents routes are experiments only.
- Docker and Docker Compose are supported as a basic local development setup, not production infrastructure.
- PostgreSQL catalog repositories, migrations, authentication, user libraries, search, pagination, and social features are planned later.

## Tech Stack

- Go 1.26
- chi router
- `html/template`
- SQLite via `modernc.org/sqlite`
- PostgreSQL driver via `github.com/lib/pq`
- Pico CSS plus project CSS
- HTMX vendored locally for a small catalog filter spike
- Templ and gomponents as rendering experiments

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

Set environment variables before the command when you need a different configuration:

```bash
APP_ENV=dev APP_DB_DSN='./data/book_social_dev.db' make run
```

## Run With Docker

Docker/Compose is a dev-only local setup for v0.1.

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

The Compose setup stores SQLite data in a named volume mounted at `/app/data`.
On first start, the container initializes and seeds `/app/data/book_social_dev.db` if it is missing or empty.

Reset the Docker SQLite database:

```bash
docker compose down -v
docker compose up --build
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
internal/storage/sqlite/ SQLite repository implementation
internal/storage/postgresql/ PostgreSQL connection and repository skeleton
internal/http/           rendering, response helpers, middleware, view models
internal/web/            templates, static assets, rendering experiments
db/sqlite/               local SQLite schema, seed, reset script
docs/                    project documentation
docs/ai/                 AI-agent context, task history, spike notes
```

## Documentation

- [Architecture](docs/architecture.md)
- [Development](docs/development.md)
- [Routes](docs/routes.md)
- [Domain model](docs/domain.md)
- [Database v0.1](docs/database_v0_1.md)
- [Database v0.2 target](docs/database_v0_2.md)
- [Testing](docs/testing.md)
- [Roadmap](docs/roadmap.md)
- [AI project context](docs/ai/project-context.md)

## Roadmap Summary

Near-term cleanup:
- Finish documentation inventory and cleanup.
- Keep v0.1 as the stable learning baseline.
- Keep Docker/Compose as a dev-only local setup; do not add production deployment claims yet.

v0.2 direction:
- Quality baseline: format/test/lint/CI.
- Database strategy: migrations and completing PostgreSQL repository support.
- Catalog read model updates for the v0.2 schema.
- Authentication and user flows after the data foundation is clearer.
