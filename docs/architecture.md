# Architecture

Book Social is a small modular monolith.

## Current Shape

```text
cmd/web/main.go
  -> config, logging, database selection, renderer, handlers
  -> internal/app.New(...)
  -> chi router

HTTP handler
  -> service/use case
  -> repository interface
  -> SQLite repository for dev
  -> PostgreSQL repository for stage/prod

cmd/coverctl/main.go
  -> config, database selection, cover store
  -> internal/coverops: load, install, or verify a reviewed cover bundle offline
```

## Layers

- `cmd/web`: application bootstrap.
- `cmd/coverctl`: offline cover maintenance command (`install` and `verify`).
- `internal/app`: app wiring, middleware, routes, home/about and auth pages, and adapters between
  modules and storage (library state for catalog pages, cover preparation for operator uploads).
- `internal/modules/books`: catalog domain models, catalog service and handler, administrator
  operator service and pages with audit, missing-book requests, and cover actions.
- `internal/modules/users`: user/auth domain models, password policy, registration/authentication
  services, session service, and repository contracts. It has no HTTP imports.
- `internal/modules/library`: private-library model; add, list, status-change, and removal service
  with optimistic version checks; handler, page views, and repository contract. It may depend on
  the catalog boundary but not vice versa.
- `internal/coverops`: cover bundle loading, installation, and verification for `cmd/coverctl`.
- `internal/storage/sqlite`: SQLite implementation of repository interfaces.
- `internal/storage/postgresql`: PostgreSQL connection and repository implementation.
- `internal/storage/covers`: local cover-file store and image validation.
- `internal/http/auth`: cookie/token boundary, typed current-user request context, and the
  authentication and administrator route guards.
- `internal/http/cover`: `/media/covers/` file delivery.
- `internal/http`: renderer, response helpers, middleware, flash messages, shared page/navigation
  views.
- `internal/web`: server templates and static assets.

## Current Decisions

- Keep the project simple and educational.
- Prefer clear Go code over framework-heavy abstractions.
- Keep `html/template` as the primary rendering path.
- Keep Templ and gomponents out of executable code after the completed comparison spikes; their
  revisit conditions live in `docs/backlog.md`.
- Use SQLite for `APP_ENV=dev`.
- Use PostgreSQL for `APP_ENV=stage` and `APP_ENV=prod`.
- Keep the normalized v0.2 catalog read contract aligned between SQLite and PostgreSQL.
- Use migrations for schema evolution; reset and local Compose bootstrap apply migrations before
  deterministic seed data.
- Keep catalog cards and details on explicit view models: authors and genres are collections, and
  only `front` artwork is displayed. Details use its main image; cards prefer a prepared thumbnail.
  List repositories hydrate cover metadata in a batch, including owner-scoped library lists.
- Keep HTTP lifecycle policy in `internal/app` and reusable response middleware in
  `internal/http/middleware`.
- Use `signal.NotifyContext` for process shutdown, a five-second HTTP graceful-shutdown deadline,
  and treat `http.ErrServerClosed` as normal completion.
- Apply request/security middleware globally, route-level timeouts only to dynamic MPA pages, and
  static cache middleware only to `/static/*`. `/media/covers/*` sets its own immutable cache policy
  and does not load sessions or flashes.
- Apply `http.CrossOriginProtection` to dynamic MPA routes after current-user and flash loading.
  It rejects unsafe cross-origin browser requests without form CSRF tokens; bypass patterns are
  not configured.
- Keep password hashing and credential verification inside `internal/modules/users`; store only
  bcrypt hashes and return neutral invalid-credential outcomes without skipping the expensive
  verification path for an unknown account.
- Use DB-backed opaque sessions. The browser receives a random raw token only after its 32-byte
  SHA-256 hash is persisted; services own the seven-day absolute lifetime.
- Services own transaction boundaries for multi-table use cases. Registration creates the default
  role user and first session atomically.
- Keep generic 500 response bodies free of internal error details; log the detailed error on the
  server side. Buffer template output before committing its status.
- Keep private-library ownership in typed current-user context and owner-scoped repository queries;
  catalog and detail forms submit only a book slug.
- Guard library status changes and removal with an item version: a stale form is a conflict, and
  repeating the current status is a no-op.
- Keep `books` independent of `library`: catalog pages receive library state through an adapter in
  `internal/app`.
- Check the administrator capability in the route guard and again in the operator service, inside
  the transaction for writes. Book slugs are immutable, and every catalog write appends an audit
  record in the same transaction.
- Name local cover files by the SHA-256 of their bytes and publish them before the SQL transaction
  that references them. A rolled-back write may leave an unreferenced file for later cleanup.
- From `000002` onward, a down migration refuses to discard data that the previous schema cannot
  represent; the `000001` baseline down migration drops the schema.

## Package Boundaries

- Handlers should translate HTTP input/output.
- Services should own use-case behavior and page data assembly.
- Repositories should hide SQL details.
- Templates should receive view/page data, not raw database rows.

## Out of Scope For Now

- Large frontend framework.
- Full API/OpenAPI surface.
- Production Docker/Kubernetes setup.
- Notes, ratings, and public libraries.
