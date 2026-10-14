# Routes

Current MPA routes are registered in `internal/app/routes.go`.

## HTTP Middleware

The router applies middleware in this order:

```text
SecurityHeaders -> RequestID -> TrustedRealIP -> request logger -> Recoverer -> dynamic timeout -> no-store -> current-user -> flash -> CrossOriginProtection -> route guard/handler
```

`TrustedRealIP` is a no-op unless `APP_TRUSTED_PROXY_CIDRS` is configured. When configured, forwarded
client-IP headers are accepted only from an immediate peer inside one of those networks; otherwise
the logger uses the connection `RemoteAddr`.

`http.CrossOriginProtection` is applied to dynamic MPA routes and has no insecure bypass patterns. Unsafe browser requests
identified as cross-origin by Fetch Metadata or `Origin` receive `403` before reaching a handler;
same-origin requests and safe methods continue normally. Requests without either browser signal are
accepted as non-browser/unknown clients, so HTTPS and `SameSite=Lax` remain separate defenses.

The application timeout is added only to dynamic MPA page routes:

```text
GET /, /about, /books, /books/{slug}, /authors/{slug} -> Timeout(30s)
GET/POST /books/request -> Timeout(30s)
```

`/healthz`, `/static/*`, and the fallback 404 handler do not use the application timeout. The 404
handler still loads current-user and flash request state so its navigation matches dynamic pages.
Server transport timeouts remain configured separately in `internal/app/server.go`.

Every response receives the conservative browser policy from `internal/http/middleware/security.go`:
content type sniffing and framing are disabled, referrers are reduced to origin on cross-origin
requests, unused browser capabilities are disabled, and the content policy allows self-hosted
assets plus HTTPS/data cover images. HSTS is intentionally not enabled for the local HTTP workflow.

Static assets use `Cache-Control: public, max-age=3600` for successful responses. Missing or failed
static responses use `Cache-Control: no-store`. Dynamic MPA pages, HTML 404 pages, and HTMX partial
responses use `Cache-Control: no-store`, because navigation and flash state can be session-specific.

## Error Responses

The recovery middleware converts handler panics into a generic 500 response and writes structured
panic diagnostics to the server logger; panic values are not sent to clients. Service and repository
failures use the same generic `internal server error` body, while the detailed error is written only
to the server logger. Domain not-found and invalid input errors retain their 404 and 400 status
contracts. Template rendering is buffered before the status is committed, so a rendering failure
can still become a correct 500 response.

## Pages

```text
GET /                    Home page
GET /about               About page
GET /books               Catalog page
GET /books/{slug}        Book details page
GET /authors/{slug}      Author page
GET /static/*            Static files
```

Authentication routes use the v0.2.5 opaque DB-backed session foundation:

```text
GET /register       registration form; signed-in users redirect to /me
POST /register      create user, role, and first session atomically; 303 /me or 422 field errors
GET /login          login form; signed-in users redirect to /me
POST /login         authenticate and create a new session; 303 /me or neutral 422 refusal
POST /logout        invalidate the current session, clear cookies, and 303 /
GET /me             protected minimal identity page; anonymous users receive 303 /login
GET /me/library     protected owner-scoped library page; anonymous users receive 303 /login
POST /me/library    protected add form with book_slug; 303 /me/library, 422/404/409 safe errors
```

Unsafe browser POST requests remain protected by `http.CrossOriginProtection`. A cross-origin
request receives `403` before mutation. Session and flash cookies are `HttpOnly`, `SameSite=Lax`,
and use `Secure` outside development. Dynamic pages respond with `Cache-Control: no-store`.

`POST /me/library` accepts only the form body's `book_slug`; the owner is always the typed
current-user identity. A repeated item is `409`, an unknown book is `404`, and an empty slug is
`422`. Successful additions set a one-request success flash. There is no route for another user's
library.

### v0.3.1 lifecycle routes

The following protected owner-scoped routes implement the reading-state and removal flow:

```text
POST /me/library/{itemID}/status  form fields: status, version
GET  /me/library/{itemID}/remove  explicit removal confirmation page
POST /me/library/{itemID}/remove  form field: version
```

The status action and confirmed removal use `303 /me/library` plus a one-request flash on success.
Malformed item IDs, status values, and versions return `422`; owner-scoped missing items return
`404`; stale forms return `409` without overwriting a newer state or deleting the item. The existing
authentication, cross-origin, and `no-store` policies apply unchanged.

### v0.3.2 catalog-operation routes

All five administrator routes are wired into the runtime with server-rendered search, creation,
and correction pages. DB-backed HTTP acceptance verifies both database backends; the user confirmed
successful local browser review on 2026-10-12.
The guard reads the current role through session loading on every request; promotion and demotion
take effect on the next request without signing in again. Page models and request logs do not
receive the internal `IsAdmin` capability. Operator services also check the actor's database
capability, inside the transaction for catalog writes.

```text
GET  /admin/catalog                    optional query field: q
GET  /admin/catalog/books/new          creation form
POST /admin/catalog/books/new          book form
GET  /admin/catalog/books/{id}/edit    correction form
POST /admin/catalog/books/{id}/edit    book form
```

Admin routes redirect anonymous users to `/login`, return `403` for ordinary authenticated users,
and retain existing `no-store` and cross-origin policies. Book forms use Post/Redirect/Get; malformed
input is `422`, unknown book IDs are `404`, and duplicate create slugs are `409`. Successful writes
redirect to the book's edit page with a fixed one-request flash. Blank search does not list the
catalog. Validation/conflict pages retain submitted text and existing-author/genre selections,
with linked field errors. URL-encoded POST bodies are limited to 128 KiB; malformed/oversized forms,
duplicate scalar fields, and malformed selected IDs return `422` without reaching a catalog write.
Only POST-body fields are accepted; correction ignores any submitted slug, actor, or audit fields.
The creation form proposes an ASCII slug through optional JavaScript until the operator edits it.
Without JavaScript, the operator enters a slug manually; server validation is always required.

The missing-book feedback routes for #25 are wired into the runtime:

```text
GET  /books/request                    missing-book form
POST /books/request                    title, author
```

The form uses `RequireAuthentication`; anonymous or expired sessions receive `303 /login`.
Only body fields `title` and `author` are used. Invalid fields, malformed encoding, and bodies over
64 KiB return accessible `422` form errors. Successful submission emits one metadata-only event
and redirects with `303 /books` and a one-request confirmation flash. Unsafe cross-origin POST
requests receive `403` before event creation; `no-store` applies to these routes.
The literal `/books/request` route takes precedence over `/books/{slug}`. The flow does not persist
a request or call an external service. Manual browser review is pending. Details are in
[Curated Catalog Operations v0.3.2 Contract](catalog_operations_v0_3_2.md).

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
