# Page Contract

This file is kept as the short page/navigation contract.

For the current route list, see [routes.md](routes.md).

## Rules

- Routes are registered in `internal/app/routes.go`.
- Handlers prepare page/view data.
- Templates render page/view data.
- Navigation state comes from shared page data.
- Catalog filter links should remain normal links.
- HTMX behavior may enhance links, but must not be required for basic navigation.

## Current Navigation Areas

- Home: `/`
- About: `/about`
- Catalog: `/books`
- Book details: `/books/{slug}`
- Author details: `/authors/{slug}`

## Breadcrumbs

- Home should not show breadcrumbs.
- Catalog and detail pages may show subtle breadcrumbs.
- Breadcrumb data should be prepared by handlers/services, not hard-coded per template.
