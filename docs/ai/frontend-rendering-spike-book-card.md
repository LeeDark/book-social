# Frontend Rendering Spike: BookCard

Date: 2026-06-29

## What Was Added

- Kept the existing `html/template` catalog route at `/books` as the source of truth.
- Kept the Templ `BookCard` experiment at `/books-templ`.
- Added gomponents using the current module path: `maragu.dev/gomponents`.
- Added a gomponents `BookCard` component and catalog-like page at `/books-gomponents`.
- Added small HTTP render helpers for Templ and gomponents.
- Added integration-test coverage for both experimental routes.

## Changed Files

- `go.mod`
- `go.sum`
- `Makefile`
- `internal/app/routes.go`
- `internal/app/app_integration_test.go`
- `internal/http/render/templ.go`
- `internal/http/render/gomponents.go`
- `internal/modules/books/handler.go`
- `internal/web/templ/components/*`
- `internal/web/templ/pages/*`
- `internal/web/gomponents/components/*`
- `internal/web/gomponents/pages/*`
- `docs/ai/task-history.md`

## How To Run

- Generate Templ code: `make templ-generate`
- Check generated Templ code: `go tool templ generate -check`
- Run tests: `GOCACHE=/tmp/book-social-go-cache go test ./...`
- Manual routes:
  - `/books`
  - `/books-templ`
  - `/books-gomponents`

## What Worked Well

Current Go templates:
- Still fit the app well for full server-rendered pages.
- Keep page structure readable as HTML.
- Require no extra build step.

Templ:
- The `BookCard` partial maps naturally to a Templ component.
- Generated code is CI-friendly with `go tool templ generate -check`.
- Good fit for typed reusable components while preserving existing CSS.

gomponents:
- No code generation or extra Makefile target is needed.
- Components are normal Go functions, so refactoring and debugging are straightforward.
- The component can render directly to `http.ResponseWriter`.

## What Was Awkward

Templ:
- Adds generated files and a required generation step.
- Import direction matters. To avoid a cycle, the spike uses Templ-local view types and adapters.
- Larger pages still need layout/navigation decisions before migration.

gomponents:
- HTML structure is less visually HTML-like because it is written as nested Go calls.
- Larger pages may become harder to scan than `html/template` or Templ.
- Like Templ, using local component view types avoids module import cycles but adds small adapters.

Current Go templates:
- Less compile-time checking for template fields.
- Reusable partials work, but refactoring component contracts is weaker than typed Go code.

## Comparison

Integration complexity:
- `html/template`: already integrated.
- Templ: moderate; needs dependency, tool entry, generated files, render helper.
- gomponents: low; needs dependency, render helper, plain Go component functions.

Build/tooling complexity:
- `html/template`: none.
- Templ: requires generation, but the project-local `go tool templ` setup is acceptable.
- gomponents: no generation; normal Go build/test path.

Readability:
- `html/template`: best for full-page HTML readability.
- Templ: good balance for component HTML with Go data.
- gomponents: readable for small components, noisier for larger pages.

Type safety and refactoring:
- `html/template`: weakest.
- Templ: strong once generated.
- gomponents: strong because it is plain Go.

Fit for book-social:
- Keep `html/template` for pages/layout for now.
- Templ is a good candidate for reusable visual components later.
- gomponents is useful for small typed HTML experiments, but less compelling for full pages in this MPA.

Migration risk:
- Both alternatives can coexist with the current templates.
- Templ migration requires generation discipline.
- gomponents migration requires accepting HTML-as-Go readability tradeoffs.

## Recommendation

Keep `html/template` for now.

Use Templ later for selected reusable components if typed component contracts become valuable. Keep gomponents as an acceptable small-component experiment, but do not migrate pages/layout to gomponents now.
