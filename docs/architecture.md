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
