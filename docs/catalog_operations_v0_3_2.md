# Curated Catalog Operations v0.3.2 Contract

This document defines the accepted v0.3.2 contract for issues #23–#25. Demo data/provenance (#23)
and the operator workflow (#24) are implemented; automated operator HTTP acceptance passes on
SQLite and PostgreSQL, and the user confirmed successful manual browser review. Missing-book
feedback (#25) is implemented with automated service and HTTP coverage; the user confirmed
successful manual browser review and accepted it on 2026-10-15.

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

## Missing-Book Feedback (#25)

```text
GET  /books/request
POST /books/request    title, author
```

The form is available only to an authenticated user and is linked from the public catalog and its
empty state, including HTMX partial responses. Both fields are trimmed: `title` is required and
limited to 255 Unicode code points; `author` is optional with the same limit. Invalid UTF-8 and NUL
are rejected. Only `title` and `author` from the form body are used; query parameters and additional
submitted fields are ignored. The form offers no comments or contact fields.

Invalid input, malformed form encoding, or a body exceeding 64 KiB returns an accessible `422`
form error. Field errors preserve and HTML-escape submitted values, link the error summary to the
affected input, and use `aria-invalid` and `aria-describedby`. Anonymous or expired sessions receive
`303 /login`. The existing `no-store` policy and cross-origin protection apply; unsafe cross-origin
POST requests receive `403` before event creation.

A valid request emits one structured `catalog_missing_book_requested` event with `request_id`,
`actor_id`, `title_length`, `author_length`, and `has_author`. Lengths count Unicode code points after
trimming; a blank author has length zero and `has_author=false`. The actor comes from current-user
identity, not submitted fields. Submitted title/author text, credentials, session tokens, and
unrelated request data are excluded. This metadata-only event replaces the earlier plan to log
bounded title/author text and follows #25's restriction on arbitrary free-text logging.
The POST performs no database write, external lookup, import, email, queue, or administrator inbox
operation. The log records the fact of submission, not enough data to identify and fulfil the
requested book. It redirects with `303` to `/books` and a fixed one-request confirmation flash that
does not promise catalog addition. Following the redirect or refreshing creates no further event.

Manual review was confirmed successful by the user on 2026-10-15; see the
[completed missing-book browser checklist](testing.md#manual-browser-smoke-v032-missing-book-feedback-25).

## Planned Missing-Book Persistence Follow-Up (v0.3.2)

After accepting the non-persistent #25 flow, the user expanded v0.3.2 on 2026-10-15:

- Save the normalized title, optional author and authenticated submitting user's ID as application
  data in the database. User identity must come from the session, not submitted form fields.
- Keep the existing form and validation. Set the success flash and redirect only after a successful
  save; a failed save must not produce a success confirmation or successful-submission event.
- Keep title/author values out of logs; saving bounded application data is separate from logging.
- Paired SQLite/PostgreSQL migration `000007` now creates the request table with constraints and
  guarded rollback. Both storage packages now implement one-insert request creation through
  `MissingBookRepository`. Service wiring, HTTP outcomes and save-failure behavior remain pending
  and are not covered by the earlier accepted browser smoke.

### Stored Request Model (Defined; Persistence Not Implemented)

The persistence model was defined on 2026-10-23. Paired migration `000007` now creates
`missing_book_requests`; runtime persistence remains pending. The model-definition step added no
runtime behavior. Migrations and repository implementations do not yet change the service or HTTP
flow.

| Field | Meaning and ownership | Stored value |
|-------|-----------------------|--------------|
| `id` | Database-generated request identity | Positive integer primary key |
| `user_id` | Submitting account, taken from authenticated session identity | Required foreign key to `users.id` |
| `title` | Reader's requested book title, normalized by the service | Required text, 1–255 Unicode code points after trimming |
| `author` | Reader's optional author text, normalized by the service | Non-null text, 0–255 Unicode code points; absent or whitespace-only input becomes `""` |
| `created_at` | Submission time assigned by the application | Required UTC timestamp; RFC3339Nano text in SQLite and `TIMESTAMPTZ` in PostgreSQL |

Both text fields use the existing trimming and validation rules: invalid UTF-8 and NUL are rejected.
Author text is not an `authors` reference, and a request has no required `books` reference: the
requested book or author may be absent from the catalog. Names, emails and other account details
are not copied into the request. The existing `MissingBookRequest` Go type is form/use-case input
with title and author only; server-owned identity, request ID and creation time are not form fields.

The submitting-user foreign key uses `ON DELETE CASCADE` and `ON UPDATE RESTRICT`. Deleting an
account deletes its requests rather than retaining orphaned rows or copied personal identifiers.
This rule applies only to missing-book requests; it does not change existing audit restrictions or
introduce account-deletion functionality.

`CreateMissingBookRequestParams` carries service-validated title/author and server-owned user ID and
creation time to `MissingBookRepository.Create`. Both implementations use a parameterized insert
without normalization, timestamp assignment, events or logging. Unexpected database failures become
`ErrMissingBookInternal`, without wrapping raw diagnostics that may contain submitted text. Context
cancellation and deadline errors are preserved. The service will own validation and success ordering
when connected in the next step.

Each successful submission creates a separate row. There is no uniqueness rule on title, author or
user, no deduplication and no retry/idempotency guarantee. Refreshing the redirected catalog page is
a GET and creates no request. `id` distinguishes repeated submissions; `created_at` records when
each was made.

v0.3.2 stores submitted requests without a status, decision reason, processing timestamp, reviewer
or resolved-book link. All requests still present when the v0.4.3 processing workflow is introduced
will be initial unprocessed requests; that release will add the required processing model through
a later migration. This step does not add operator queries, indexes for an undecided listing
contract, notification delivery, or a background job model.

The existing metadata-only events cannot reconstruct earlier submissions. Persistence will collect
new requests after implementation. An operator list and the ability to mark requests processed or
rejected are explicitly planned for v0.4.3, alongside controlled import. v0.3.2 does not add that
operator workflow or notifications.

## Deferred Work

v0.3.2 does not include external imports or APIs, public anonymous submissions, an operator request
inbox or request processing, catalog deletion, author/genre administration, role-management UI,
slug redirects, Unicode slugs or
transliteration, cover uploads/proxying, source adapters, audit browsing, pagination, full-text
search, or concurrent editor conflict handling.
