# Roadmap

This roadmap is a working guide, not a release promise.

## v0.1 Baseline

v0.1 is closed as the baseline for the next development stage.

Done:
- [x] Basic app skeleton: config, logging, errors, build info, Makefile.
- [x] chi router, middleware, routes, 404 rendering.
- [x] Modular monolith shape.
- [x] Layered books/catalog module: handler, service, repository.
- [x] SQLite schema, seed data, reset script.
- [x] Home and About pages.
- [x] Catalog page.
- [x] Book details page.
- [x] Author page.
- [x] Author and genre catalog filters.
- [x] View/page models, navigation, breadcrumbs.
- [x] Basic unit tests and small HTTP/integration-style tests.
- [x] HTMX catalog filter spike.
- [x] Templ and gomponents rendering spikes.
- [x] Initial documentation audit.
- [x] Close v0.1 as the baseline.

Still intentionally rough:
- [x] Docker/Compose local environment workflows.
- [x] No database migrations in the v0.1 runtime/reset workflow.
- [x] PostgreSQL has startup support and v0.1 catalog repository behavior.
- [x] No auth or user library features.
- [x] No search, sorting, or pagination.
- [x] No real cover image storage.

## v0.2 Direction

v0.2 should be implemented by dependency order, not as one large mixed feature batch.

Main waves:
- Data / infrastructure: quality checks, migrations, SQLite/PostgreSQL decision, schema v0.2,
  repositories, catalog changes.
- User / auth: sessions, cookies, registration, login/logout, auth middleware, user-facing auth UI.

## v0.2.1 Quality & DB Foundation

Goal: make schema changes safe before changing the catalog model.

- [x] Add or confirm Make targets: `test`, `fmt`, `vet`, optional `lint`.
- [x] Add CI for `go test ./...`, `go vet ./...`, and lint if configured.
- [x] Decide how the database driver / config is selected.
- [x] Define migration layout for SQLite and possible PostgreSQL support.
- [x] Add migration commands using the `golang-migrate` CLI.
- [x] Clarify `reset-db` and seed workflow.
- [x] Add minimal test DB bootstrap helpers if needed for repository tests.

Definition of Done:
- [x] `make test` passes.
- [x] The project has a clear migration path for v0.2 schema work.
- [x] PostgreSQL is minimally supported at startup and has v0.1 catalog repository behavior.

Related docs:
- `README.md`
- `docs/development.md`
- `docs/database.md`
- `docs/database_v0_1.md`
- `docs/testing.md`
- `docs/ai/project-context.md`

## v0.2.2 Domain Model v0.2

Goal: move the data model from the v0.1 simple catalog shape to the v0.2 target shape.

Preparation:
- [ ] Add a SQLite migration smoke target for a disposable database.
- [ ] Decide whether SQLite migration smoke should run in CI now or remain local/manual.
- [ ] Decide when to add a PostgreSQL service job to CI for repository and migration checks.
- [ ] Redesign disposable reset flow to run migrations up, then apply seed data.
- [ ] Keep seed data as development/sample data, separate from schema migrations.
- [ ] Decide when Docker/Compose bootstrap should switch from schema SQL to migrations plus seed.
- [ ] Plan the v0.2 migration sequence before repository changes.

Schema:
- [ ] Add migrations for `book_authors` and `book_genres`.
- [ ] Add `covers` with URL and metadata fields, including `UNIQUE(book_id, variant)`.
- [ ] Add `library_items`.
- [ ] Add `library_item_tags`.
- [ ] Migrate v0.1 data from `books.book_author_id` and `books.book_genre_id`.
- [ ] Migrate `library` rows into `library_items`.
- [ ] Migrate `library.library_tag_id` into `library_item_tags`.
- [ ] Remove old `books.book_author_id`, `books.book_genre_id`, and `library` after data is migrated.
- [ ] Update seed data for v0.2.
- [ ] Decide and document slug policy for books, authors, and genres.

Definition of Done:
- [ ] Fresh database setup works with v0.2 seed data.
- [ ] Existing v0.1 catalog data has a defined migration story.
- [ ] SQLite migration up/down smoke passes.
- [ ] Seed smoke passes after migrations.
- [ ] SQLite test DB can be created from the v0.2 schema.
- [ ] Covers are stored as metadata/URLs only; upload/storage is deferred.

## v0.2.3 Catalog v0.2

Goal: restore and improve catalog reads after the schema change.

- [ ] Add read models for book cards with multiple authors and genres.
- [ ] Add a book details read model with authors, genres, and covers.
- [ ] Add author details read model with the author's books.
- [ ] Update catalog page.
- [ ] Update book details page.
- [ ] Update author details page.
- [ ] Update the home page to use the new read model.
- [ ] Update MPA endpoint documentation in `docs/routes.md` or a focused endpoint doc.

Definition of Done:
- [ ] Stable routes still work where possible: `/`, `/books`, `/books/{slug}`, `/authors/{slug}`.
- [ ] Repository and handler tests cover the new catalog read shape.
- [ ] Templates still use view/page models rather than raw database structs.

## v0.2.4 HTTP Foundation

Goal: prepare the HTTP layer for auth and user workflows.

- [ ] Add or review a graceful shutdown.
- [ ] Confirm a base middleware chain.
- [ ] Add request ID if useful.
- [ ] Confirm logging middleware behavior.
- [ ] Add panic recovery/error page behavior if missing.
- [ ] Add secure headers' policy.
- [ ] Add a static / cache headers policy.
- [ ] Decide timeout middleware scope.

Definition of Done:
- [ ] HTTP middleware order is documented or obvious in code.
- [ ] Handler tests cover important error/status behavior.
- [ ] No real server startup is required for Codex verification; use `httptest`.

## v0.2.5 Auth Foundation

Goal: create the minimal auth/session core before building forms.

- [ ] Decide session strategy; prefer minimal DB-backed sessions for this project.
- [ ] Add `sessions` table or equivalent session storage.
- [ ] Define password hashing policy: hash only, no plaintext password storage or logging.
- [ ] Add user repository/service foundations.
- [ ] Add a session create/delete/load behavior.
- [ ] Add current-user middleware.
- [ ] Define minimal transaction rule: services own transactions for use cases touching multiple tables.
- [ ] Decide CSRF strategy for MPA forms.
- [ ] Define a minimal validation strategy for auth inputs.

Definition of Done:
- [ ] Auth core can create users, verify passwords, create sessions, delete sessions, and load the current user.
- [ ] Unit tests cover success and failure paths.
- [ ] Password and session behavior are documented enough for future maintenance.

## v0.2.6 Registration/Login/Logout

Goal: expose the minimal user-facing auth workflow.

- [ ] Add a registration form and handler.
- [ ] Add login form and handler.
- [ ] Add a logout handler.
- [ ] Add `/me` or another minimal protected route.
- [ ] Update navigation for anonymous and logged-in states.
- [ ] Add flash messages for login/logout/register outcomes.
- [ ] Add validation errors for duplicate login/email and invalid credentials.
- [ ] Add handler tests with `httptest`.

Definition of Done:
- [ ] Anonymous users can register and log in.
- [ ] Logged-in users can log out.
- [ ] Anonymous users are redirected from the protected route.
- [ ] `make test` passes.
- [ ] v0.2 release notes or task history are updated.

## Later

These notes are intentionally less detailed than v0.1/v0.2. Keep them as direction markers and refine them only when the project gets closer to that stage.

Short summary:
- User library: statuses, shelves, tags, notes.
- Search, sorting, and pagination.
- Importers and cover storage.
- Admin basics.
- Accessibility and i18n pass.
- Observability and deployment hardening.

## v0.3

- [ ] Data layer maturity beyond the minimal v0.2 rules.
- [ ] Repository contracts.
- [ ] Test fixtures beyond minimal v0.2 bootstrap.
- [ ] Admin audit basics.
- [ ] Error model and mature validation strategy.
- [ ] More unit tests.
- [ ] Integration tests; consider testcontainers.
- [ ] Admin panel basics.

## v0.4

- [ ] User's Library features (to add later)
- [ ] Canonical URL strategy
- [ ] Reading statuses, ratings, notes, shelves/tags
- [ ] Pagination, filters, search
- [ ] Search contract
- [ ] Empty states / UX states
- [ ] Accessibility basics
- [ ] i18n support

## v0.5

- [ ] Catalog features (to add later)
- [ ] Import books from third-party databases like Goodreads
- [ ] Importer abstraction, background jobs
- [ ] Pagination, filters, search
- [ ] Full-text search in catalog
- [ ] Cover storage

## v0.6

- [ ] User features more deeply, social features step by step: following, feeds, likes
- [ ] Monitoring and tracing with Grafana/Prometheus
