# Curated Catalog Operations v0.3.2 Contract

This document defines the accepted v0.3.2 contract for issues #23–#25. Demo data/provenance (#23)
and the operator workflow (#24) are implemented; automated operator HTTP acceptance passes on
SQLite and PostgreSQL, and the user confirmed successful manual browser review. Missing-book
feedback (#25) remains planned and is not available in the runtime.

## Purpose and Boundaries

v0.3.2 maintains a small, curated demo catalog for a closed cohort. It does not introduce an
external catalog provider, public contribution flow, import job, media-storage system, or general
catalog administration platform.

The three data categories have separate responsibilities:

- **Test fixtures** are minimal, scenario-specific data owned by tests. Ordinary unit, handler, and
  repository tests must not depend on the development seed.
- **Development seed data** is deterministic runtime bootstrap data applied after migrations to a
  disposable SQLite or PostgreSQL development database.
- **Curated demo catalog** is the reviewed catalog subset represented by development seed data and
  accompanied by provenance documentation. It is not a test fixture and is not a production import
  source.

Every demo record with a cover must have documented metadata source, cover URL, permitted-use
basis, and known limitations. Covers remain external URL metadata only. An unverified, unavailable,
or omitted cover is represented by the existing CSS placeholder; v0.3.2 does not download, proxy,
cache, upload, or store cover files.

## Administrator Authorization

The existing `roles.is_admin` capability is used by one new `admin` role. The paired v0.3.2
migration `000006` adds that role; ordinary registration continues to assign only `user`.

The environment owner appoints the first administrator through one documented, deliberate database
operation. There is no role-management page, self-service promotion, or seeded administrator
account.

Current-user loading exposes an internal `IsAdmin` capability to the authorization guard. It is
not copied to page models, templates, forms, flashes, logs, or public responses. The guard
has these outcomes:

| Requester                   | Admin route outcome         |
|-----------------------------|-----------------------------|
| Anonymous                   | `303 See Other` to `/login` |
| Authenticated ordinary user | `403 Forbidden`             |
| Authenticated administrator | Handler executes            |

The existing `no-store` policy and `http.CrossOriginProtection` apply unchanged.

## Operator Routes

```text
GET  /admin/catalog                    optional query field: q
GET  /admin/catalog/books/new          creation form
POST /admin/catalog/books/new          title, slug, description, author_ids, genre_ids, front_cover_url
GET  /admin/catalog/books/{id}/edit    correction form
POST /admin/catalog/books/{id}/edit    title, description, author_ids, genre_ids, front_cover_url
```

`GET /admin/catalog` is an administrator-only find page. A non-blank `q` performs a
case-insensitive substring match against title and slug, ordered by title then ID. A blank query
shows the form without a full catalog listing. This bounded search needs no pagination for the
curated demo catalog.

The route ID is a positive book ID. A malformed ID or form field is `422`; an unknown book is `404`.
Successful creation redirects with `303` to that book's edit page. Successful correction redirects
with `303` back to the same edit page. Both set a one-request success flash.

## Book Create and Correct Rules

A book title is trimmed, required, and limited to the existing 255-character catalog column limit.
The form initially generates a lower-case ASCII slug from the title. The administrator may edit the
proposal before creation. Before writing, the slug must be non-empty, unique, no more than 255
characters, and match:

```text
^[a-z0-9]+(?:-[a-z0-9]+)*$
```

The generator may not transliterate Unicode. If a title produces an empty or unsuitable proposal,
the administrator supplies a valid slug. A duplicate slug is `409 Conflict`. Once created, a slug
is immutable in v0.3.2 so existing public URLs stay valid; redirect history is deferred.

Creation and correction may set a description, zero or more existing authors, zero or more existing
genres, and one optional `front` cover URL. Duplicate selected IDs are normalized. A missing or
invalid selected ID is `422`. The optional front-cover value is either empty, which clears the
front cover on correction, or a valid HTTPS URL. Other cover variants, author/genre creation or
correction, book deletion, and last-write-wins or optimistic concurrent-edit policy are outside this
release.

Handlers parse HTTP and render accessible form errors. A catalog service validates and orchestrates
use cases; repository implementations own SQL. Each create or correction, its author/genre link
replacement, its front-cover change, and its audit entry use one service-owned transaction.

## Audit Contract

The `catalog_audit_log` table records every successful book creation or correction. The application
boundary exposes only append operations; it has no audit-update/delete API. A database owner can
still mutate rows through direct SQL, so this is not a database-level tamper-proof log:

```text
id, actor_user_id, entity_type, entity_id, action, summary, occurred_at
```

`entity_type` is `book`; `action` is `create` or `correct`; `occurred_at` is UTC. `summary` is
server-generated: `created` for a creation, or a stable sorted list of changed groups from `title`,
`description`, `authors`, `genres`, and `front_cover`. It does not contain submitted values, a full
before/after snapshot, credentials, session data, or arbitrary form text.
A successful no-op correction also appends exactly one record, with an empty summary. Audit writes
commit in the same transaction as catalog changes; an audit failure rolls all catalog writes back.

## Planned Missing-Book Feedback (#25)

```text
GET  /books/request
POST /books/request    title, author
```

The form is available only to an authenticated user and is linked from the public catalog. `title`
is trimmed, required, and limited to 255 characters; `author` is optional and limited to 255
characters. No comments, contact details, or other free-form fields are accepted. Invalid input is
rendered as an accessible `422` form error. Anonymous users receive `303 /login`.

A valid request emits one structured `catalog_missing_book_requested` event with request ID, actor
ID, normalized bounded title, and optional normalized bounded author. Structured logging prevents
log-format injection; credentials, session tokens, and unrelated request data are never logged.
The POST performs no database write, external lookup, import, email, queue, or administrator inbox
operation. It redirects with `303` to `/books` and a one-request confirmation flash.

## Deferred Work

v0.3.2 does not include external imports or APIs, public submissions, persistent requests, catalog
deletion, author/genre administration, role-management UI, slug redirects, Unicode slugs or
transliteration, cover uploads/proxying, source adapters, audit browsing, pagination, full-text
search, or concurrent editor conflict handling.
