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
- Recorded Docker/Compose status for follow-up; it was later clarified as a dev-only local setup.
- Kept Templ and gomponents as spike-only context, not main README routes.
- Moved raw local AI/task notes into `docs/archive/` for later review without staging or committing them.

Validation:
- Documentation-only change.
- `make test` should still pass before finishing.

## 2026-06-30 — Docker/Compose v0.1 dev setup cleanup

Result:
- Verified `docker build --progress=plain -t book-social:dev .`.
- Found initial Compose failure: final image did not include server templates.
- Copied runtime templates/static assets into the final image.
- Added a small dev entrypoint that initializes and seeds the SQLite database when the Docker database file is missing or empty.
- Wired `APP_DB_DSN` into app startup.
- Documented Docker/Compose as a basic local development setup, not production infrastructure.

Validation:
- `docker compose up --build -d` started the app.
- `curl http://localhost:8080/` returned 200.
- `curl http://localhost:8080/books` returned 200 with seeded catalog data.

## 2026-07-07 — Makefile refactoring and documentation update

Result:
- Reorganized and standardized the `Makefile` with a `group/action` pattern.
- Added targets: `tidy`, `audit`, `build`, `test`, `lint`, `lint/fix`, `db/reset`, `db/shell`, `docker/build`, `docker/up`, `docker/down`, `templ/generate`.
- Consolidated redundant linting targets into a single robust `lint` target with version-aware installation.
- Updated `docs/development.md` to reflect the new standardized commands.
- Partially addressed v0.2.1 Quality & DB Foundation roadmap goals.

Changed files:
- `Makefile`
- `docs/development.md`

Commands run:
- `make help`
- `make test`
- `make lint` (found pre-existing issues)
- `make build`

Validation:
- All tests pass.
- Help output is organized and clear.
- Build successfully produces binary.

## 2026-07-09 — APP_ENV database selection and documentation update

Result:
- Restricted runtime `APP_ENV` values to `dev`, `stage`, and `prod`.
- Kept `APP_ENV=dev` on SQLite using `APP_DB_DSN`.
- Added PostgreSQL startup support for `APP_ENV=stage` and `APP_ENV=prod`.
- Added a PostgreSQL book repository skeleton that satisfies the catalog repository interface but returns not implemented errors.
- Documented terminal environment variable usage and clarified that `APP_ENV=test` is not a runtime mode.

Decisions:
- Keep database selection in `cmd/web/main.go` for now because startup wiring already lives there.
- Do not copy SQLite catalog queries into PostgreSQL yet.
- Treat PostgreSQL as connection/startup support only until the repository implementation is ported deliberately.

Validation:
- `GOCACHE=/tmp/book-social-go-cache make test` passed after the implementation change.
- Documentation follow-up was docs-only.

## 2026-07-10 — PostgreSQL BookRepository implementation

Result:
- Ported the SQLite catalog repository behavior to `internal/storage/postgresql`.
- Implemented book listing, author/genre filtering, book detail lookup by slug, and author lookup by slug.
- Preserved existing domain-level not-found errors for missing books and authors.
- Added PostgreSQL repository tests covering a list, filter, detail, and not-found behavior.

Changed files:
- `internal/storage/postgresql/books_repository.go`
- `internal/storage/postgresql/books_repository_test.go`
- `docs/ai/task-history.md`

Commands run:
- `GOCACHE=/tmp/book-social-go-cache go test ./internal/storage/postgresql`
- `GOCACHE=/tmp/book-social-go-cache make test`

Validation:
- Focused PostgreSQL storage tests passed.
- Full `make test` passed.

Documentation follow-up suggestions, completed in the next entry:
- Update `docs/ai/project-context.md` to say PostgreSQL catalog repository methods are implemented for the v0.1 schema.
- Update `README.md`, `docs/architecture.md`, `docs/database.md`, `docs/database_v0_1.md`, `docs/development.md`, and `docs/roadmap.md` to remove stale "PostgreSQL repository skeleton/placeholders" wording.
- Consider documenting how to initialize a stage/prod PostgreSQL database with `db/postgresql/schema_v0_1.sql` and `db/postgresql/seed.sql`.

## 2026-07-10 — PostgreSQL documentation sync

Result:
- Updated current documentation to reflect that PostgreSQL now has v0.1 catalog repository behavior.
- Removed stale PostgreSQL skeleton/placeholder wording from project-facing docs.
- Documented manual PostgreSQL initialization with `db/postgresql/schema_v0_1.sql` and `db/postgresql/seed.sql`.
- Kept migration system and production deployment setup documented as future work.

Changed files:
- `README.md`
- `docs/architecture.md`
- `docs/database.md`
- `docs/database_v0_1.md`
- `docs/development.md`
- `docs/roadmap.md`
- `docs/ai/project-context.md`
- `docs/ai/task-history.md`

Validation:
- Documentation wording scan found no remaining stale PostgreSQL placeholder language outside historical task-history entries.

## 2026-07-11 — Docker Compose environment workflows

Result:
- Split Compose configuration into a common app file plus environment-specific files for dev, stage, and prod.
- Kept `APP_ENV=dev` on SQLite with a named `/app/data` volume.
- Added local PostgreSQL services for `APP_ENV=stage` and `APP_ENV=prod`, initialized from the existing v0.1 PostgreSQL schema and seed SQL files.
- Updated the Docker entrypoint so SQLite initialization only runs for `APP_ENV=dev`.
- Replaced generic Compose Make targets with explicit environment commands.
- Updated current docs to describe the new Docker/Compose commands and reset behavior.

Changed files:
- `Makefile`
- `compose.yaml`
- `compose.dev.yaml`
- `compose.stage.yaml`
- `compose.prod.yaml`
- `docker/entrypoint.sh`
- `README.md`
- `docs/development.md`
- `docs/database.md`
- `docs/roadmap.md`
- `docs/ai/project-context.md`
- `docs/ai/task-history.md`

Validation:
- `make help`
- `docker compose -f compose.yaml -f compose.dev.yaml config`
- `docker compose -f compose.yaml -f compose.stage.yaml config`
- `docker compose -f compose.yaml -f compose.prod.yaml config`
- `sh -n docker/entrypoint.sh`
- `GOCACHE=/tmp/book-social-go-cache make test`

## 2026-07-11 — Reset, seed, and test DB workflow clarification

Result:
- Documented that local reset is destructive and recreates the database from schema plus seed SQL.
- Clarified that seed data is deterministic development data, not production data or a repeatable migration.
- Documented the current SQLite and PostgreSQL reset paths.
- Added shared SQLite test DB helpers under `internal/testutil`.
- Reused the shared SQLite catalog schema helper in app integration, SQLite repository, and PostgreSQL repository tests.
- Added shared PostgreSQL catalog test DB helpers under `internal/testutil`.
- Converted PostgreSQL repository tests to opt-in real PostgreSQL tests using `BOOK_SOCIAL_POSTGRES_TEST_DSN`.
- Kept PostgreSQL repository tests skipped by default when the test DSN is not set.
- Updated roadmap status for reset/seed workflow clarification and minimal test DB bootstrap helpers.

Changed files:
- `docs/database.md`
- `docs/development.md`
- `docs/testing.md`
- `docs/roadmap.md`
- `internal/testutil/sqlite_db.go`
- `internal/testutil/postgresql_db.go`
- `internal/app/app_integration_test.go`
- `internal/storage/sqlite/books_repository_test.go`
- `internal/storage/postgresql/books_repository_test.go`
- `docs/ai/task-history.md`

Validation:
- `GOCACHE=/tmp/book-social-go-cache go test ./internal/app ./internal/testutil`
- `GOCACHE=/tmp/book-social-go-cache go test ./internal/storage/sqlite ./internal/testutil`
- `GOCACHE=/tmp/book-social-go-cache go test ./internal/storage/postgresql ./internal/testutil`
- `GOCACHE=/tmp/book-social-go-cache make test`

## 2026-07-12 — v0.2.1 migration foundation closure

Result:
- Added baseline migration layout for SQLite and PostgreSQL under `db/*/migrations`.
- Added v0.1 baseline migration pairs for both database dialects.
- Wired `make db/migrate/up` and `make db/migrate/down` to the installed `golang-migrate` CLI.
- Rebuilt the local `migrate` CLI with `sqlite` and `postgres` build tags for verification.
- Documented that reset/bootstrap still use schema plus seed SQL directly until a later reset migration task.
- Confirmed existing CI covers `go test ./...`, `go vet ./...`, and golangci-lint.
- Updated depguard allowlist for the PostgreSQL driver used by current storage code.
- Marked v0.2.1 Quality & DB Foundation roadmap Definition of Done complete.
- Documented how CI, Docker Compose, and migrations relate today.

Changed files:
- `.golangci.yml`
- `Makefile`
- `README.md`
- `docs/database.md`
- `docs/database_v0_1.md`
- `docs/development.md`
- `docs/roadmap.md`
- `docs/ai/project-context.md`
- `docs/ai/task-history.md`
- `db/sqlite/migrations/000001_create_v0_1_schema.up.sql`
- `db/sqlite/migrations/000001_create_v0_1_schema.down.sql`
- `db/postgresql/migrations/000001_create_v0_1_schema.up.sql`
- `db/postgresql/migrations/000001_create_v0_1_schema.down.sql`

Validation:
- `migrate -help` listed `sqlite`, `postgres`, and `postgresql` database drivers after rebuild.
- `make db/migrate/up MIGRATIONS_DATABASE_URL=sqlite:///tmp/book-social-migrate-cli-smoke-20260712.db`
- `make db/migrate/down MIGRATIONS_DATABASE_URL=sqlite:///tmp/book-social-migrate-cli-smoke-20260712.db`
- `GOCACHE=/tmp/book-social-go-cache make test`

Decision:
- Use `golang-migrate` CLI instead of a custom in-project migration runner.
- Keep Docker Compose dev/stage/prod bootstrap on schema plus seed SQL for now.
- Treat migration smoke tests as local/manual until CI grows database service jobs.

## 2026-08-03 — v0.2.2 catalog domain model closure

Result:
- Recorded decisions for legacy `library`/`shelves`/`tags`, strict down-migration limits, stable
  slugs, and the local-only CI boundary.
- Added SQLite and PostgreSQL `000002_normalize_catalog` migrations for many-to-many authors and
  genres plus URL/metadata-only covers.
- Preserved v0.1 relationships during migration and removed the old direct book FK columns.
- Updated SQLite and PostgreSQL seed data to populate junction tables.
- Switched SQLite reset, PostgreSQL reset, and Docker/Compose bootstrap to migrations followed by
  seed data.
- Added v0.2 SQLite/PostgreSQL test helpers and checks for normalized relationships and cover
  uniqueness.
- Extended `make db/migrate/smoke` to cover clean seed, v0.1 data migration, and down-migration.
- Updated database, development, README, roadmap, project-context, and task-history documentation.

Decisions:
- Legacy library tables remain isolated demo data; `library_items` is deferred to v0.3.
- Down-migration refuses ambiguous multi-author or multi-genre data instead of dropping links.
- Slugs remain stable, unique per entity type; redirect behavior after a change is deferred.
- SQLite migration smoke remains local/manual. A PostgreSQL CI service job is deferred.
- v0.2.3 is a separate dependent branch for read-side catalog adaptation.

Validation:
- `make db/migrate/smoke`
- `GOCACHE=/tmp/book-social-go-cache make test`
- `git diff --check`
- PostgreSQL tests pass with a disposable DSN when run sequentially (`go test -p 1 ./...`); they
  skip when `BOOK_SOCIAL_POSTGRES_TEST_DSN` is unset.

## 2026-08-06 — Retire executable Templ and gomponents spikes

Result:
- Removed the `/books-templ` and `/books-gomponents` routes, handlers, adapters, render helpers,
  component/page implementations, generated Templ files, and their positive route tests.
- Removed Templ and gomponents dependencies and the Templ generation target.
- Kept `html/template` as the application rendering path and preserved the existing HTMX catalog
  partial-response behavior.
- Preserved the original spike notes as historical evidence.
- Added `docs/backlog.md` for unprioritized technical ideas and recorded explicit conditions for
  reconsidering typed server-side rendering.

Decision:
- The roadmap remains the source of truth for current priority and accepted release scope.
- Backlog entries are candidates, not commitments, and move to the roadmap only after review.

## 2026-08-06 — v0.2.3 Catalog v0.2 closure

Result:
- Replaced the v0.1 single-author/single-genre catalog read shape with normalized `Authors`,
  `Genres`, and `Covers` collections across SQLite and PostgreSQL repositories.
- Kept catalog filters and stable MPA routes intact. Filters select matching books while cards and
  author pages retain each selected book's full relationships.
- Updated catalog, book-details, author, and HTMX partial templates to render multiple authors and
  genres from view models.
- Added URL-only cover mapping for details: the `front` variant renders as an image and missing or
  non-front-only data uses the existing CSS placeholder.
- Replaced the home page's hardcoded book-card data with a narrow featured-books provider and the
  shared card view model. The section no longer promises a chronological "Recently added" order.
- Removed the executable Templ/gomponents spikes and retained their historical notes.
- Documented current normalized catalog behavior and marked v0.2.3 closed; v0.2.4 HTTP Foundation
  is the next planned release.

Validation:
- Focused service, handler, SQLite, PostgreSQL, and app integration tests passed.
- PostgreSQL parity tests passed against a disposable PostgreSQL DSN, run sequentially.
- `GOCACHE=/tmp/book-social-go-cache make test` passed.
- `make db/migrate/smoke` passed.
- `git diff --check` passed.

Decision:
- Keep external cover URLs as read metadata only. File upload, proxying, caching, galleries,
  search, sorting, pagination, auth, and user-library behavior remain outside v0.2.3.

## 2026-08-09 — v0.2.4 HTTP Foundation implementation

Result:
- Connected `SIGINT` and `SIGTERM` to the web process root context and implemented graceful HTTP
  shutdown with a five-second deadline.
- Treated `http.ErrServerClosed` as normal completion and preserved unexpected listener errors after
  shutdown.
- Documented middleware order: security headers, request ID, trusted real IP, request logger, recovery,
  then route handler.
- Applied the 30-second application timeout only to dynamic MPA routes; health, static, and 404
  routes keep their existing semantics.
- Added conservative security headers and a CSP compatible with self-hosted assets and external
  HTTPS/data cover images. Disabled unused HTMX inline indicator styles to avoid `unsafe-inline`.
- Added a one-hour public cache for successful static assets and `no-store` for missing/failed
  assets. HTML and HTMX partial responses do not receive a public long-lived cache policy.
- Hardened template rendering by buffering output before committing the response status. Panic and
  internal failures remain generic to clients while detailed errors stay in logs.

Validation:
- `GOCACHE=/tmp/book-social-go-cache make test` passed.
- Focused middleware, renderer, response, and app tests passed.
- `GOCACHE=/tmp/book-social-go-cache go vet ./...` passed.
- `git diff --check` passed.
- PostgreSQL tests skipped when `BOOK_SOCIAL_POSTGRES_TEST_DSN` was not set.
- Real server startup was not used; HTTP behavior was verified with `httptest` in accordance with
  the Codex sandbox constraint.

Decisions:
- Keep the current simple `html/template` and chi-based HTTP foundation.
- Defer HSTS, TLS/reverse-proxy policy, auth middleware, rate limiting, CORS, and production
  deployment hardening.

Status: v0.2.4 closed. The next planned release is v0.2.5 Auth Foundation.

## 2026-08-19 — v0.2.5 Auth Foundation closure

Result:
- Added equivalent SQLite/PostgreSQL migration `000003` for the ordinary `user` role and DB-backed
  sessions with unique 32-byte token hashes, absolute expiry, and reversible rollback rules.
- Added the `internal/modules/users` password, registration/authentication, user repository, and
  session-service boundaries without HTTP imports.
- Kept default-role lookup and user creation transactional; normalized duplicate login/email and
  unexpected persistence failures into typed safe domain outcomes.
- Adopted bcrypt with a 12-character minimum and its 72-byte input boundary. Unknown accounts use
  a dummy bcrypt verification path, while malformed stored hashes become internal errors.
- Added SQLite/PostgreSQL user and session repositories, SQLite migration smoke coverage, and
  opt-in PostgreSQL parity tests.
- Added the two-phase HTTP token/cookie boundary: generate the raw token, persist its SHA-256 hash,
  and only then expose the cookie. The central absolute session lifetime is seven days.
- Added typed current-user request context and a testable authentication guard. Production `/me`
  and auth dependency wiring remain v0.2.6 work.
- Added global `http.CrossOriginProtection` after recovery without bypass patterns; focused tests
  cover unsafe cross-origin refusal and same-origin success without form CSRF tokens.
- Reviewed test diagnostics so password hashes, raw session tokens, token hashes, and issued cookie
  values are not printed on failure.

Accepted implementation evidence:
- `book-social` commit `41a8ddb` (`fix(auth): harden password and session boundaries`) is the accepted
  v0.2.5 foundation revision and includes the complete preceding branch history.
- Stage 7A records this as foundation-only evidence; registration/login/logout, production `/me`,
  navigation, flashes, and the full browser flow remain planned for v0.2.6.

Validation:
- Focused users, auth HTTP, config, SQLite/PostgreSQL repository, and migration-helper tests passed.
- `GOCACHE=/tmp/book-social-go-cache make test` passed with the race detector.
- `GOCACHE=/tmp/book-social-go-cache go vet ./...` passed.
- `GOCACHE=/tmp/book-social-go-cache GOLANGCI_LINT_CACHE=/tmp/book-social-golangci-cache make lint`
  passed with zero issues.
- `make db/migrate/smoke` passed for SQLite migration, seed, legacy data, and rollback paths.
- `git diff --check` passed.
- PostgreSQL opt-in tests were skipped in the final default run because
  `BOOK_SOCIAL_POSTGRES_TEST_DSN` was not set; earlier Stage parity evidence remains recorded in the
  private implementation plan.

Decision:
- Close v0.2.5 without user-facing auth routes. The next release is v0.2.6
  Registration/Login/Logout, which consumes this foundation and completes the applied MPA flow.
