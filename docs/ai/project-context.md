# Project Context

Book Social is a learning Go web project.

Main goals:
- practice Go web development
- learn layered architecture
- build a modular monolith
- practice MPA / SSR with templates
- start with SQLite
- compare SQLite and PostgreSQL incrementally
- gradually add tests, auth, better catalog behavior, and selected frontend experiments

Current modules:
- home
- books/catalog
- book details
- author details
- static assets
- rendering/templates
- app skeleton: config, logging, errors
- users/auth foundation: password policy, registration/authentication services, DB sessions
- HTTP auth foundation: cookie/token boundary, current-user context, and route guard
- private library: owner-scoped add/list, lifecycle transitions, and confirmed removal workflow
- catalog maintenance: administrator-only find/create/correct MPA pages and transactional audit
- missing-book requests: authenticated form, saved before confirmation, metadata-only logging
- cover files: local content-addressed storage, `/media/covers/` delivery, administrator uploads,
  and offline bundle installation and verification with `coverctl`

Current focus:
- minimal working features first
- clean package boundaries
- small incremental tasks
- current documentation should describe implemented behavior, not planned behavior
- v0.3.3 Reliability and Release Closure is active. Its roadmap section also lists the defects
  found in the 2026-11-04 review of the full code history; code-hygiene items tied to v0.3.3–v0.3.5
  are in `docs/backlog.md`.

Current catalog behavior:
- `/books` lists books.
- `/books?author={authorSlug}` filters by author slug.
- `/books?genre={genreSlug}` filters by genre slug.
- `/books?author={authorSlug}&genre={genreSlug}` applies both filters.
- `/books/{bookSlug}` shows book details.
- `/authors/{authorSlug}` shows an author page and books by that author.
- Catalog cards and book details show all related authors and genres.
- Book details show the main `front` image and cards prefer its thumbnail. A CSS placeholder
  appears when a book has no cover or its image fails to load, including after HTMX updates.
- The home page loads a deterministic featured selection through the shared catalog-card read model.
- `/books/request/new` lets an authenticated user request a missing book.
- `/admin/catalog` lets administrators find, create, and correct books; slugs are immutable.

Current rendering direction:
- `html/template` is the primary rendering path.
- HTMX is present as a small progressive-enhancement spike for catalog filters.
- Completed Templ and gomponents experiments remain documented in spike notes, but their routes,
  dependencies, and executable code have been removed. Revisit conditions live in
  `docs/backlog.md`.

Current infrastructure caveat:
- `APP_ENV=dev` uses SQLite and is the active local database path.
- `APP_ENV=stage` and `APP_ENV=prod` open PostgreSQL with `APP_DB_DSN`.
- SQLite and PostgreSQL catalog repositories implement the same normalized v0.2 read contract.
- SQLite and PostgreSQL v0.1 baseline, v0.2 normalization, v0.2.5 auth, v0.3 private-library,
  v0.3.1 lifecycle, and v0.3.2 administrator/audit, missing-book request, and local cover-file
  migrations through `000008` exist under `db/*/migrations`.
- Cover files live in `APP_COVERS_DIR` (the `/app/covers` volume in Compose). Reviewed artwork is
  kept outside Git and installed from a bundle with a pinned manifest checksum.
- Migration commands use the installed `golang-migrate` CLI through `make db/migrate/up` and `make db/migrate/down`.
- Reset and Docker/Compose bootstrap apply all migrations first and then seed SQL.
- `make db/migrate/smoke` checks clean setup, seed counts, the v0.1 relationship migration, the
  v0.2.5 user/session constraints, v0.3 library constraints, and rollback behavior.
- CI runs quality checks and `make test/integration` in a separate isolated PostgreSQL job.
  Standalone SQLite migration smoke and Node slug-suggestion checks remain local/manual.
  PostgreSQL tests use a disposable DSN and run with `go test -p 1` when sharing one database.
- Legacy `library`, `shelves`, and `tags` remain demo structures. `library_items` is the separate,
  implemented private-library model.
- Docker/Compose are supported as local environment workflows for SQLite dev and PostgreSQL stage/prod.
- Docker/Compose are not production-ready infrastructure.
- HTTP lifecycle uses signal-aware graceful shutdown with a five-second deadline.
- Dynamic MPA routes use a 30-second application timeout; static, health, and fallback routes keep
  their existing semantics.
- Security headers, static cache policy, buffered rendering failures, and structured panic recovery
  are implemented and covered by focused HTTP tests.
- Global `http.CrossOriginProtection` rejects unsafe cross-origin browser requests. DB-backed
  session, cookie, current-user, and guard foundations support the production auth routes.
- `APP_TRUSTED_PROXY_CIDRS` conditionally enables forwarded client-IP handling; it is disabled when
  unset.

Current scope:
- v0.3.0 Private Library Foundation is closed: authenticated users add known catalog books from
  catalog/detail pages and list only their own items at `/me/library`.
- v0.3.1 Reading-State Lifecycle is closed: owner-scoped status changes, timestamp rules,
  optimistic conflicts, catalog state display, and explicit-confirmation removal are implemented.
- v0.3.2 Curated Catalog Operations is closed: PR #27 was merged on 2026-11-03, its merge commit
  is tagged `v0.3.2`, and issues #22–#26, #28 and #29 are closed. It delivered deterministic demo
  data with documented provenance (#23), the administrator boundary with catalog MPA forms,
  immutable slugs and transactional audit (#24), missing-book requests saved before confirmation
  (#25), local cover-file storage and delivery (#28), and reviewed demo front covers for all 109
  books (#29). Verification and manual acceptance evidence is in `docs/testing.md` and the
  `docs/cover_*_v0_3_2.md` records. No deployment is claimed. Operator listing/processing of
  saved requests remains v0.4.3.
- The missing-book form lives at `/books/request/new` so that `/books/request` stays the public
  URL of a book whose slug is `request`.
- v0.3.3 Reliability and Release Closure is the active release.
