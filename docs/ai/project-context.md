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

Current focus:
- minimal working features first
- clean package boundaries
- small incremental tasks
- current documentation should describe implemented behavior, not planned behavior
- v0.2.5 Auth Foundation is complete at implementation commit `41a8ddb`; the next active release is
  v0.2.6 Registration/Login/Logout.

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
- SQLite and PostgreSQL v0.1 baseline, v0.2 normalization, and v0.2.5 auth migrations exist under
  `db/*/migrations`.
- Migration commands use the installed `golang-migrate` CLI through `make db/migrate/up` and `make db/migrate/down`.
- Reset and Docker/Compose bootstrap apply all migrations first and then seed SQL.
- `make db/migrate/smoke` checks clean setup, seed counts, the v0.1 relationship migration, the
  v0.2.5 user/session constraints, and rollback behavior.
- CI currently keeps migration smoke and a PostgreSQL service job local/manual; PostgreSQL tests
  use a disposable DSN and should run with `go test -p 1` when sharing one database.
- Legacy `library`, `shelves`, and `tags` remain demo structures. The final `library_items` model is
  deferred to v0.3.
- Docker/Compose are supported as local environment workflows for SQLite dev and PostgreSQL stage/prod.
- Docker/Compose are not production-ready infrastructure.
- HTTP lifecycle uses signal-aware graceful shutdown with a five-second deadline.
- Dynamic MPA routes use a 30-second application timeout; static, health, and fallback routes keep
  their existing semantics.
- Security headers, static cache policy, buffered rendering failures, and structured panic recovery
  are implemented and covered by focused HTTP tests.
- Global `http.CrossOriginProtection` rejects unsafe cross-origin browser requests. DB-backed
  session, cookie, current-user, and guard foundations are tested but production auth routes are not
  registered yet.
- `APP_TRUSTED_PROXY_CIDRS` conditionally enables forwarded client-IP handling; it is disabled when
  unset.

Next planned release:
- v0.2.6 Registration/Login/Logout: forms and handlers, production auth dependency wiring,
  protected `/me`, navigation, flashes, and the complete browser flow on the accepted foundation.
