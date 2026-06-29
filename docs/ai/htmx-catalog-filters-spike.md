# HTMX Catalog Filters Spike

Date: 2026-06-29

## What changed

- Added local HTMX for progressively enhanced catalog filters.
- Kept `/books` as the source-of-truth MPA route.
- Normal catalog requests render the full page.
- HTMX catalog requests render only the book list partial.
- Kept filter links as normal `href` links so they work without JavaScript.

## Vendored HTMX

- Version: HTMX 2.0.4
- Source: `https://unpkg.com/htmx.org@2.0.4/dist/htmx.min.js`
- Stored at: `internal/web/static/js/vendor/htmx.min.js`
- Included from: `internal/web/templates/base.tmpl`
- Served as: `/static/js/vendor/htmx.min.js`

The layout includes HTMX with:

```html
<script defer src="/static/js/vendor/htmx.min.js"></script>
```

## Templates

- `internal/web/templates/pages/catalog.tmpl`
  - Adds a stable `#book-list` target.
  - Reuses the `book_list` partial for full-page rendering.
- `internal/web/templates/partials/book_list.tmpl`
  - Renders only the catalog result list or empty state.
- `internal/web/templates/partials/book_card.tmpl`
  - Adds HTMX attributes to genre filter links.
  - Adds a small author filter link while preserving the existing author detail link.

Filter links use normal `href` plus HTMX attributes:

```html
hx-target="#book-list"
hx-swap="innerHTML"
hx-push-url="true"
```

## Handler behavior

`internal/modules/books/handler.go` checks:

```go
r.Header.Get("HX-Request") == "true"
```

When true, it renders only `book_list` from `catalog.tmpl`. Otherwise, it renders the normal full catalog page.

## Known limitations

- This is only a small catalog-filter UX spike.
- There is no search, pagination, or sorting.
- The visible author filter control is intentionally minimal.
- No frontend build pipeline was added.

## Recommendation

Keep the spike for now. Expand later only if catalog filtering grows enough to need clearer dedicated filter controls.
