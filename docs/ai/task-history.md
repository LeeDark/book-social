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

## 2026-06-29 — Frontend rendering spike: Templ vs gomponents for BookCard

Result:
- Kept the existing `/books` Go-template page unchanged.
- Kept the isolated `/books-templ` Templ route.
- Added gomponents using the current module path `maragu.dev/gomponents`.
- Added a gomponents `BookCard` component and catalog-like page.
- Added an isolated `/books-gomponents` route.
- Added integration-test coverage for `/books-gomponents`.
- Added comparison note at `docs/ai/frontend-rendering-spike-book-card.md`.

Changed files:
- `go.mod`
- `go.sum`
- `internal/app/routes.go`
- `internal/app/app_integration_test.go`
- `internal/http/render/gomponents.go`
- `internal/modules/books/handler.go`
- `internal/web/gomponents/components/book_card.go`
- `internal/web/gomponents/components/view.go`
- `internal/web/gomponents/pages/books.go`
- `internal/web/gomponents/pages/view.go`
- `docs/ai/frontend-rendering-spike-book-card.md`
- `docs/ai/templ-spike-book-card.md`
- `docs/ai/task-history.md`

Commands run:
- `GOCACHE=/tmp/book-social-go-cache go list -m -versions maragu.dev/gomponents`
- `GOCACHE=/tmp/book-social-go-cache go get maragu.dev/gomponents@v1.3.0`
- `GOCACHE=/tmp/book-social-go-cache go test ./internal/app ./internal/modules/books ./internal/http/render ./internal/web/gomponents/components ./internal/web/gomponents/pages`
- `GOCACHE=/tmp/book-social-go-cache go mod tidy`
- `make templ-generate`
- `go tool templ generate -check`
- `GOCACHE=/tmp/book-social-go-cache go test ./...`
- `make test`

Validation:
- Focused package tests passed.
- Templ generation/check passed with no generated updates.
- Full `go test ./...` passed.
- `make test` passed.

Decision:
- Keep `html/template` for now.
- Use Templ later for selected reusable components if typed component contracts become valuable.
- Keep gomponents as an acceptable small-component experiment, but do not migrate pages/layout to gomponents now.

## 2026-06-29 — HTMX spike for catalog filters

Result:
- Added local vendored HTMX 2.0.4 at `internal/web/static/js/vendor/htmx.min.js`.
- Included HTMX from the base layout with `defer`.
- Added a `book_list` template partial and stable `#book-list` catalog target.
- Added HTMX attributes to catalog genre filter links.
- Added a small author catalog filter link while preserving existing author detail links.
- Updated the catalog handler to return the full page for normal requests and only the book list partial for `HX-Request: true`.
- Kept `/books`, `/books?author=...`, and `/books?genre=...` working as normal MPA routes.
- Added spike note at `docs/ai/htmx-catalog-filters-spike.md`.

Changed files:
- `internal/http/render/renderer.go`
- `internal/modules/books/handler.go`
- `internal/modules/books/service.go`
- `internal/modules/books/view.go`
- `internal/web/templates/base.tmpl`
- `internal/web/templates/pages/catalog.tmpl`
- `internal/web/templates/partials/book_card.tmpl`
- `internal/web/templates/partials/book_list.tmpl`
- `internal/web/static/js/vendor/htmx.min.js`
- `internal/modules/books/handler_test.go`
- `internal/modules/books/service_test.go`
- `internal/modules/books/view_test.go`
- `internal/app/app_integration_test.go`
- `docs/ai/htmx-catalog-filters-spike.md`
- `docs/ai/task-history.md`

Commands run:
- `curl -L https://unpkg.com/htmx.org@2.0.4/dist/htmx.min.js -o internal/web/static/js/vendor/htmx.min.js`
- `gofmt -w internal/http/render/renderer.go internal/modules/books/handler.go internal/modules/books/view.go internal/modules/books/handler_test.go internal/modules/books/view_test.go internal/app/app_integration_test.go`
- `gofmt -w internal/modules/books/view.go internal/modules/books/service.go internal/modules/books/handler_test.go internal/modules/books/service_test.go`
- `GOCACHE=/tmp/book-social-go-cache go test ./...`
- `make test`

Validation:
- Full Go test suite passed.
- `make test` passed.
- Browser/server verification was not run inside Codex because the project instructions avoid starting the web server in the sandbox.

Decision:
- Keep this as a progressive enhancement spike.
- Postpone broader filter UI, search, pagination, and sorting until a later catalog task.

## 2026-06-30 — Documentation audit and cleanup

Result:
- Added a root `README.md`.
- Added focused project docs for architecture, development, routes, database overview, and testing.
- Updated roadmap to separate the current v0.1 baseline from planned v0.2 work.
- Updated domain and database docs so v0.1 matches the active SQLite schema.
- Marked Docker/Compose as experimental.
- Kept Templ and gomponents as spike-only context, not main README routes.
- Moved raw local AI/task notes into `docs/archive/` for later review without staging or committing them.

Validation:
- Documentation-only change.
- `make test` should still pass before finishing.
