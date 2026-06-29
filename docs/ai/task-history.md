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

## 2026-06-23 — Site presentation polish

Result:
- Polished shared layout, navigation, spacing, and page rhythm.
- Improved Home page hero, copy, and real catalog/genre calls to action.
- Polished Catalog page header, results area, book cards, and responsive 3/2/1 column grid.
- Added reusable empty-state styling for catalog and author pages.
- Added branded rendered 404 page for missing books, authors, and unknown routes.
- Improved placeholder book covers across catalog, home, and book details.
- Aligned Author page header and book grid with Catalog presentation.

Decisions:
- Kept Pico CSS as the base.
- Kept server-rendered templates and existing routing style.
- Did not add frontend frameworks or new heavy dependencies.
- Removed non-functional Home keyword search until real catalog search exists.
- Used generated placeholder covers because the data model has no cover image URL yet.

Tests:
- `make test` passes.
- Added tests for empty catalog rendering and branded 404 behavior.

Known follow-ups:
- Real cover image support can add `<img loading="lazy">` later.
- Real catalog search can reintroduce the Home search form.
- Manual visual checks should be done locally outside Codex.

## 2026-06-29 — Templ spike for BookCard component

Result:
- Added Templ as a project dependency and Go tool.
- Added `make templ-generate`.
- Added a Templ `BookCard` component and generated code.
- Added an isolated `/books-templ` route that reuses the catalog service and renders Templ book cards.
- Kept the existing `/books` Go-template route unchanged.
- Added a short spike note at `docs/ai/templ-spike-book-card.md`.

Changed files:
- `go.mod`
- `go.sum`
- `Makefile`
- `internal/app/routes.go`
- `internal/app/app_integration_test.go`
- `internal/http/render/templ.go`
- `internal/modules/books/handler.go`
- `internal/web/templ/components/book_card.templ`
- `internal/web/templ/components/book_card_templ.go`
- `internal/web/templ/components/view.go`
- `internal/web/templ/pages/books_templ.templ`
- `internal/web/templ/pages/books_templ_templ.go`
- `internal/web/templ/pages/view.go`
- `docs/ai/templ-spike-book-card.md`
- `docs/ai/task-history.md`

Commands run:
- `go get github.com/a-h/templ@v0.3.1020`
- `go get -tool github.com/a-h/templ/cmd/templ@v0.3.1020`
- `make templ-generate`
- `GOCACHE=/tmp/book-social-go-cache go mod tidy`
- `GOCACHE=/tmp/book-social-go-cache go test ./internal/app ./internal/modules/books ./internal/http/render ./internal/web/templ/components ./internal/web/templ/pages`
- `go tool templ generate -check`
- `GOCACHE=/tmp/book-social-go-cache go test ./...`
- `make test`

Decision:
- Use Templ only for components for now.
- Postpone a full layout migration until the project has more repeated UI components or stronger type-safety needs in templates.
