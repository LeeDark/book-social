# Routes

Current MPA routes are registered in `internal/app/routes.go`.

## HTTP Middleware

The router applies middleware in this order:

```text
SecurityHeaders -> RequestID -> RealIP -> request logger -> Recoverer -> route handler
```

The application timeout is added only to dynamic MPA page routes:

```text
GET /, /about, /books, /books/{slug}, /authors/{slug} -> Timeout(30s)
```

`/healthz`, `/static/*`, and the fallback 404 handler do not use the application timeout. Server
transport timeouts remain configured separately in `internal/app/server.go`.

Every response receives the conservative browser policy from `internal/http/middleware/security.go`:
content type sniffing and framing are disabled, referrers are reduced to origin on cross-origin
requests, unused browser capabilities are disabled, and the content policy allows self-hosted
assets plus HTTPS/data cover images. HSTS is intentionally not enabled for the local HTTP workflow.

Static assets use `Cache-Control: public, max-age=3600` for successful responses. Missing or failed
static responses use `Cache-Control: no-store`. HTML pages, 404 pages, and HTMX partial responses
do not receive a public long-lived cache policy.

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
- Book details render all related authors and genres. When no `front` cover metadata exists, the
  page renders its CSS cover placeholder.
- Author pages show the selected author's details and catalog cards with each book's complete
  author and genre lists.

## Home Page

`GET /` renders a small deterministic featured selection through the same catalog card read model
as the other pages. It does not claim a chronological "Recently added" order because books do not
currently have `created_at`.

## HTMX Catalog Filter Spike

The catalog handler checks:

```text
HX-Request: true
```

When present, it renders only the book list partial.

Normal links still use `href`, so catalog filters remain usable without JavaScript.

## Historical Rendering Experiments

The former `/books-templ` and `/books-gomponents` spike routes have been removed. Their findings
remain in `docs/ai/`, and possible reconsideration is tracked in `docs/backlog.md`.
