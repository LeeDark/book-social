# Roadmap

This roadmap is a working guide, not a release promise.

## v0.1 Baseline

Current v0.1 is a small server-rendered book catalog.

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

Still intentionally rough:
- [ ] Docker/Compose runtime setup is experimental.
- [ ] No database migrations.
- [ ] No PostgreSQL support.
- [ ] No auth or user library features.
- [ ] No search, sorting, or pagination.
- [ ] No real cover image storage.

## v0.1 Cleanup

Before starting larger v0.2 work:
- [ ] Finish documentation cleanup.
- [ ] Decide what to do with archived AI prompt logs.
- [ ] Clarify Docker/Compose status or fix it in a focused task.
- [ ] Keep `html/template` as the primary rendering path.
- [ ] Keep Templ/gomponents as historical spikes unless a later task changes direction.

## v0.2 Direction

Recommended order:

1. Quality baseline
   - [ ] `make fmt`
   - [ ] `make vet`
   - [ ] optional lint command
   - [ ] CI running tests

2. Database foundation
   - [ ] migration strategy
   - [ ] SQLite configuration cleanup
   - [ ] PostgreSQL decision and initial support if still wanted
   - [ ] test database bootstrap helpers

3. Schema v0.2
   - [ ] books/authors many-to-many
   - [ ] books/genres many-to-many
   - [ ] covers table with URL metadata
   - [ ] `library_items`
   - [ ] `library_item_tags`
   - [ ] updated seed data

4. Catalog v0.2
   - [ ] read models for multiple authors and genres
   - [ ] update catalog page
   - [ ] update book details page
   - [ ] update author pages
   - [ ] keep routes stable where possible

5. Auth foundation
   - [ ] session strategy
   - [ ] password hashing policy
   - [ ] CSRF decision for forms
   - [ ] registration, login, logout

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

- [ ] Database: Data Layer Maturity
  - [ ] Transaction policy
  - [ ] Test fixtures / DB bootstrap
  - [ ] Repository contracts
  - [ ] Admin audit basics
- [ ] Error model, Validation strategy
- [ ] More unit tests
- [ ] Integration tests, try testcontainers
- [ ] Admin panel: Basics

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
