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

Current focus:
- minimal working features first
- clean package boundaries
- small incremental tasks
- current documentation should describe implemented behavior, not planned behavior
- v0.2.5 Auth Foundation is complete at implementation commit `41a8ddb`.
- v0.2.6 Registration/Login/Logout is complete: registration, login, idempotent logout, protected
  `/me`, auth navigation, and one-request success flashes.

Current catalog behavior:
- `/books` lists books.
- `/books?author={authorSlug}` filters by author slug.
- `/books?genre={genreSlug}` filters by genre slug.
- `/books?author={authorSlug}&genre={genreSlug}` applies both filters.
- `/books/{bookSlug}` shows book details.
- `/authors/{authorSlug}` shows an author page and books by that author.
- Catalog cards and book details show all related authors and genres.
- Book details use `front` cover metadata when available and a CSS placeholder when it is absent.
- The home page loads a deterministic featured selection through the shared catalog-card read model.

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
  v0.3.1 lifecycle, and v0.3.2 administrator/audit migrations exist under `db/*/migrations`.
- Migration commands use the installed `golang-migrate` CLI through `make db/migrate/up` and `make db/migrate/down`.
- Reset and Docker/Compose bootstrap apply all migrations first and then seed SQL.
- `make db/migrate/smoke` checks clean setup, seed counts, the v0.1 relationship migration, the
  v0.2.5 user/session constraints, v0.3 library constraints, and rollback behavior.
- CI currently keeps migration smoke and a PostgreSQL service job local/manual; PostgreSQL tests
  use a disposable DSN and should run with `go test -p 1` when sharing one database.
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
- v0.3.2 Curated Catalog Operations is active. Deterministic development/demo data and provenance
  (#23), the administrator boundary, catalog MPA forms, immutable slugs, and transactional audit
  (#24) are implemented and accepted. Automated HTTP acceptance passes on SQLite and PostgreSQL;
  the user confirmed successful operator browser review on 2026-10-12. Non-persistent missing-book
  feedback (#25) is implemented at `GET/POST /books/request`, with bounded field validation,
  metadata-only logging and a one-request flash. Automated checks pass; the user confirmed
  successful manual browser review and accepted #25 on 2026-10-15. The subsequent agreed v0.3.2
  follow-up will persist title, optional author and submitting user; it is not implemented yet.
  Operator listing and processed/rejected decisions are planned for v0.4.3. v0.3.2 remains active.
