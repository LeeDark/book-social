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
- `internal/storage/sqlite`: SQLite implementation of repository interfaces.
- `internal/storage/postgresql`: PostgreSQL connection and repository implementation.
- `internal/http`: renderer, response helpers, middleware, shared page/navigation views.
- `internal/web`: server templates, static assets, and rendering experiments.

## Current Decisions

- Keep the project simple and educational.
- Prefer clear Go code over framework-heavy abstractions.
- Keep `html/template` as the primary rendering path.
- Keep Templ and gomponents as experiments until there is a stronger reason to migrate.
- Use SQLite for `APP_ENV=dev`.
- Use PostgreSQL for `APP_ENV=stage` and `APP_ENV=prod`.
- Keep the v0.1 catalog repository behavior aligned between SQLite and PostgreSQL.
- Introduce migrations later, after the v0.1 baseline is documented.

## Package Boundaries

- Handlers should translate HTTP input/output.
- Services should own use-case behavior and page data assembly.
- Repositories should hide SQL details.
- Templates should receive view/page data, not raw database rows.

## Out of Scope For Now

- Large frontend framework.
- Full API/OpenAPI surface.
- Production Docker/Kubernetes setup.
- Authentication and user library features.
