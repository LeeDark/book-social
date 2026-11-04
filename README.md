# Book Social

Book Social is a learning Go web project for building a small server-rendered book catalog.

The project is intentionally simple: modular monolith, layered architecture, SQLite for local development, and MPA pages rendered on the server.

## Current Status

v0.3.2 Curated Catalog Operations is closed: PR #27 was merged and its merge commit tagged
`v0.3.2` on 2026-11-03. It builds on v0.3.0 Private Library Foundation and v0.3.1 Reading-State
Lifecycle. The active release is **v0.3.3 Reliability and Release Closure**; see the
[roadmap](docs/roadmap.md#v033--reliability-and-release-closure). No deployed release is claimed.

v0.3.2 added deterministic demo data with documented provenance, administrator-only catalog
find/create/correct pages with transactional audit, authenticated missing-book requests saved before
confirmation, local cover-file storage and HTTP delivery, administrator cover uploads, and reviewed
demo front covers for all 109 books. Apply migrations through `000008` before using the current
application.

Cover artwork is kept outside Git and installed from a reviewed bundle with `coverctl`; see the
[cover contract](docs/cover_storage_v0_3_2.md) and the
[delivery/recovery procedure](docs/cover_operations.md). Before recreating an older container,
preserve its cover files: the volumes do not migrate an old writable layer automatically. Demo
metadata has documented but unverified external provenance, and the SQL seed contains no covers.
Operator listing and processed/rejected decisions for saved requests remain planned for v0.4.3.
Verification and acceptance evidence: [testing](docs/testing.md#v032-acceptance-coverage-26),
[cover acceptance](docs/cover_acceptance_v0_3_2.md) and
[full-catalog cover verification](docs/cover_catalog_verification_v0_3_2.md).

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
- `golang-migrate` CLI wiring and paired migrations `000001`–`000008` for SQLite and PostgreSQL.
- Normalized catalog tables for many-to-many authors/genres and cover references: external HTTPS
  URLs or local files with technical metadata.
- Catalog cards and book details render all related authors and genres.
- Book details use the main `front` image; home/catalog/author/library cards prefer its thumbnail.
  Missing covers use a CSS placeholder; a small external script reveals it on image-load failure,
  including after HTMX updates. Images preserve proportions without cropping or stretching.
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
- Registration, login, idempotent POST logout, protected `/me`, auth navigation, and one-request
  success flashes use the DB-backed opaque-session foundation.
- Authenticated users can add a known catalog book from catalog and detail pages, then view only
  their own deterministic private library at `/me/library`.
- SQLite and PostgreSQL migration `000004` provide owner-scoped `library_items` persistence with
  duplicate prevention and deterministic list ordering; migration `000005` adds lifecycle status,
  nullable timestamps, and an optimistic-lock version.
- Users can change an owned item's state between Want to read, Reading, and Read; stale mutations
  are conflicts, repeating the current state is a no-op, and removal requires a confirmation page.
- Administrators find, create, and correct books at `/admin/catalog`. Slugs are immutable, every
  write appends a transactional audit record, and the administrator role is appointed through a
  documented database operation.
- Authenticated users can request a missing book at `/books/request/new`. The title, optional
  author, and user are saved before confirmation; the submitted text stays out of logs.
- Local cover files are named by their SHA-256, served from `/media/covers/` with immutable caching,
  uploaded by administrators, and installed or verified offline with `coverctl install|verify`.

Not current production direction:
- Docker and Docker Compose are supported as local environment workflows, not production infrastructure.
- Notes, ratings, search, pagination, and social features remain planned later.

## Tech Stack

- Go 1.26
- chi router
- `html/template`
- SQLite via `modernc.org/sqlite`
- PostgreSQL driver via `github.com/lib/pq`
- bcrypt via `golang.org/x/crypto/bcrypt`
- WebP decoding via `golang.org/x/image/webp` for cover validation
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
On every start, the container applies pending SQLite migrations to
`/app/data/book_social_dev.db`; it seeds the database only if the file was missing or empty. The
stage and prod Compose setups run a local PostgreSQL container initialized by the app entrypoint
with all PostgreSQL migrations and `db/postgresql/seed.sql`.

Stop and restart the Docker environment while retaining its database and cover volumes:

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
/me/library
/books/request/new
/admin/catalog          administrators only
```

## Test

```bash
make test
```

In constrained environments, this command is preferred over starting a real web server.

## Project Structure

```text
cmd/web/                      application entrypoint
cmd/coverctl/                 offline cover bundle installation and verification
internal/app/                 app wiring, routes, home and auth handlers, module adapters
internal/modules/books/       catalog, administrator operations, missing-book requests
internal/modules/users/       auth/user/session service and repository contracts
internal/modules/library/     private-library lifecycle service, handler, and repository contracts
internal/coverops/            cover bundle loading, installation, and verification
internal/storage/sqlite/      SQLite repository implementation
internal/storage/postgresql/  PostgreSQL connection and repository implementation
internal/storage/covers/      local cover-file store and image validation
internal/http/auth/           session cookie, current-user context, and route guards
internal/http/cover/          `/media/covers/` file delivery
internal/http/                rendering, response helpers, middleware, view models
internal/web/                 server templates, static assets, and JavaScript tests
db/sqlite/                    local SQLite schema, migrations, seed, reset script
db/postgresql/                PostgreSQL schema, migrations, seed, reset script
assets/demo-covers/           cover manifests, prompts, review records (images live outside Git)
k8s/                          local kind learning manifests
docs/                         project documentation
docs/ai/                      AI-agent context, task history, spike notes
```

## Documentation

- [Architecture](docs/architecture.md)
- [Development](docs/development.md)
- [Routes](docs/routes.md)
- [Domain model](docs/domain.md)
- [Go and SQL readability](docs/code-readability.md)
- [Private Library v0.3 contract](docs/library_v0_3.md)
- [Curated Catalog Operations v0.3.2 contract](docs/catalog_operations_v0_3_2.md)
- [Demo catalog provenance](docs/demo_catalog_provenance.md)
- [Cover file storage contract](docs/cover_storage_v0_3_2.md)
- [Cover delivery and recovery](docs/cover_operations.md)
- [Database v0.1](docs/database_v0_1.md)
- [Database v0.2](docs/database_v0_2.md)
- [MPA auth contract](docs/drafts/auth-contract.ru.md)
- [Testing](docs/testing.md)
- [Roadmap (English)](docs/roadmap.md)
- [Дорожня карта (українською)](docs/roadmap.uk.md)
- [Technical backlog](docs/backlog.md)
- [AI project context](docs/ai/project-context.md)

## Roadmap Summary

Near-term work:
- v0.3.3 Reliability and Release Closure: critical-flow end-to-end HTTP tests, the documented
  SQLite/PostgreSQL verification, safe library-operation logs, accessibility smoke checks, a release
  checklist, and the defects found in the 2026-11-04 code-history review.
- v0.3.4 Public Demo Polish: a reviewed book-addition workflow, growth to 500 books, and interface
  polish.
- v0.3.5 Hetzner Cloud Hosted Stage: one reproducible PostgreSQL stage with HTTPS, backups, and a
  verified restore.
- Keep Docker/Compose as local environment workflows; do not add production deployment claims yet.

v0.2 direction:
- Quality baseline: format/test/lint/CI.
- Database strategy: migrations and schema evolution (v0.2.2 complete).
- Catalog read model updates for the v0.2 schema (v0.2.3 complete).
- HTTP foundation: lifecycle, middleware, security, caching, recovery, and timeout policy (v0.2.4 complete).
- Auth foundation: password policy, DB sessions, current-user/guard boundary, and cross-origin
  browser protection (v0.2.5 complete).
- Registration/login/logout forms, production `/me`, navigation, and flashes are complete (v0.2.6).
