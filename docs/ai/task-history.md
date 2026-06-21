# AI Task History

## 2025-11-20 - App skeleton: config, logging, chi router

## 2026-02-02 - Domain model and database diagram

## 2026-05-04 - SQL scripts and database reset

## 2026-05-19 - App struct, Docker and Docker Compose basics

## 2026-05-25 - Home page, static files

## 2026-05-29 - Catalog page

## 2026-06-01 - Breadcrumbs

## 2026-06-01 - Home hero design

## 2026-06-17 — Add seed data

Result:
- Added larger seed dataset.
- Around 100 books.
- Multiple authors.
- Multiple genres.
- Multiple books per author/genre.

Notes:
- Seed data is for development only.
- Do not treat it as production content.

## 2026-06-17 — Book details page

Result:
- Added book details page.
- Used `BookDetailsView` naming.
- Added/updated styles for book cover/details.
- Split CSS into layout/components/pages structure.

Decisions:
- Prefer `BookDetailsView` over `BookDetailsViewModel`.
- Keep view structs near HTTP/template layer.
- CSS is organized by base/layout/components/pages.

## 2026-06-18 — Unit tests: minimal tests

Planned:
- Add minimal unit tests.
- Focus on service and handler tests.
- Use fake dependencies.
- Avoid DB integration tests for now.

## 2026-06-21 — Catalog improvements: author and genre filters, author page

Result:
- Added author slugs to the SQLite schema and seed data.
- Reused existing genre slugs.
- Added catalog filtering by author slug and genre slug.
- Added support for combined author + genre filters.
- Added author detail page at `/authors/{slug}`.
- Updated author links to use `/authors/{authorSlug}`.
- Kept genre links using `/books?genre={genreSlug}`.
- Updated home sample books and genres to use existing seed data.

Areas changed:
- SQLite schema and seed data.
- Books domain models, view models, service, handler, and repository.
- Catalog and author routing.
- Server-rendered templates for author pages.
- Minimal repository, service, and handler tests.

Decisions:
- Catalog filters use slugs and return empty results for unknown filter slugs.
- Unknown author page slugs return 404.
- SQL for filters remains in the SQLite repository.
- Author pages reuse existing book card view data.

Known follow-ups:
- No pagination, sorting, search, or empty-state redesign was added.
- No database migration system was introduced; local development still uses reset SQL.
