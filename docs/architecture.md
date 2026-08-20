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
```

## Layers

- `cmd/web`: application bootstrap.
- `internal/app`: app wiring, middleware, routes, home/about pages.
- `internal/modules/books`: catalog domain models, service, handler, views.
- `internal/modules/users`: user/auth domain models, password policy, registration/authentication
  services, session service, and repository contracts. It has no HTTP imports.
- `internal/storage/sqlite`: SQLite implementation of repository interfaces.
- `internal/storage/postgresql`: PostgreSQL connection and repository implementation.
- `internal/http/auth`: cookie/token boundary, typed current-user request context, and testable
  authentication guard. Production auth-route wiring is deferred to v0.2.6.
- `internal/http`: renderer, response helpers, middleware, shared page/navigation views.
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
  details may select a `front` cover from cover metadata.
- Keep HTTP lifecycle policy in `internal/app` and reusable response middleware in
  `internal/http/middleware`.
- Use `signal.NotifyContext` for process shutdown, a five-second HTTP graceful-shutdown deadline,
  and treat `http.ErrServerClosed` as normal completion.
- Apply request/security middleware globally, route-level timeouts only to dynamic MPA pages, and
  static cache middleware only to `/static/*`.
- Apply `http.CrossOriginProtection` globally after recovery. It rejects unsafe cross-origin
  browser requests without form CSRF tokens; bypass patterns are not configured.
- Keep password hashing and credential verification inside `internal/modules/users`; store only
  bcrypt hashes and return neutral invalid-credential outcomes without skipping the expensive
  verification path for an unknown account.
- Use DB-backed opaque sessions. The browser receives a random raw token only after its 32-byte
  SHA-256 hash is persisted; services own the seven-day absolute lifetime.
- Services own transaction boundaries for multi-table use cases. v0.2.5 keeps default-role lookup
  and user creation in one transaction; atomic registration plus session is a v0.2.6 prerequisite.
- Keep generic 500 response bodies free of internal error details; log the detailed error on the
  server side. Buffer template output before committing its status.

## Package Boundaries

- Handlers should translate HTTP input/output.
- Services should own use-case behavior and page data assembly.
- Repositories should hide SQL details.
- Templates should receive view/page data, not raw database rows.

## Out of Scope For Now

- Large frontend framework.
- Full API/OpenAPI surface.
- Production Docker/Kubernetes setup.
- User-facing registration/login/logout, production `/me`, auth navigation, and flashes.
- User library features.
