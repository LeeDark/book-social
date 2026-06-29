# Templ Spike: BookCard Component

Date: 2026-06-29

## What Was Added

- Added Templ as a project dependency and tool entry.
- Added `make templ-generate`, backed by `go tool templ generate`.
- Added a Templ `BookCard` component under `internal/web/templ/components`.
- Added an isolated experimental `/books-templ` route.
- Kept the existing `/books` Go-template route unchanged.

## What Worked Well

- `BookCard` maps naturally to a component because the existing Go template partial already uses a focused view model.
- Templ generation is straightforward with project-local tooling.
- The existing CSS classes can be reused without redesigning the page.
- The experimental route can reuse the real catalog service and SQLite repository wiring.

## What Was Awkward

- Import direction matters. The books handler cannot import a Templ page that imports `internal/modules/books`, so the spike uses a small Templ-local view type and adapter.
- A full layout migration would require more decisions around shared navigation, breadcrumbs, and base page structure.
- Generated code is readable enough for debugging, but it adds committed code churn compared with plain Go templates.

## Recommendation

Use Templ only for components for now.

Do not migrate the whole frontend yet. Templ is promising for repeated UI pieces like `BookCard`, but this project should postpone a layout-wide migration until there are more repeated components or stronger type-safety needs in templates.
