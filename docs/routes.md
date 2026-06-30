# Routes

Current MPA routes are registered in `internal/app/routes.go`.

## Pages

```text
GET /                    Home page
GET /about               About page
GET /books               Catalog page
GET /books/{slug}        Book details page
GET /authors/{slug}      Author page
GET /static/*            Static files
```

## Catalog Filters

`GET /books` supports:

```text
?author={authorSlug}
?genre={genreSlug}
?author={authorSlug}&genre={genreSlug}
```

Unknown filter slugs return an empty catalog result, not a 404.

## Detail Pages

- Unknown book slugs return 404.
- Unknown author slugs return 404.

## HTMX Catalog Filter Spike

The catalog handler checks:

```text
HX-Request: true
```

When present, it renders only the book list partial.

Normal links still use `href`, so catalog filters remain usable without JavaScript.

## Experimental Rendering Routes

The codebase currently has experimental routes for Templ and gomponents catalog rendering.

These are documented in `docs/ai/` spike notes and are not part of the main user-facing route contract.
