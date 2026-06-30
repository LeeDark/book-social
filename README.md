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

Not current production direction:
- Templ and gomponents routes are experiments only.
- Docker and Docker Compose are early/experimental.
- PostgreSQL, migrations, authentication, user libraries, search, pagination, and social features are planned later.

## Tech Stack

- Go 1.25
- chi router
- `html/template`
- SQLite via `modernc.org/sqlite`
- Pico CSS plus project CSS
- HTMX vendored locally for a small catalog filter spike
- Templ and gomponents as rendering experiments

## Run Locally

Reset the local development database:

```bash
make db-dev-reset
```

Run the web app:

```bash
make run
```

Default address:

```text
http://localhost:8080
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
- Treat Docker/Compose as experimental until runtime asset/database behavior is cleaned up.

v0.2 direction:
- Quality baseline: format/test/lint/CI.
- Database strategy: migrations and SQLite/PostgreSQL decision.
- Catalog read model updates for the v0.2 schema.
- Authentication and user flows after the data foundation is clearer.
