# Domain Model

This document describes domain concepts. Database column details live in:

- [Database v0.1](database_v0_1.md)
- [Database v0.2](database_v0_2.md)

## Current v0.2.6 Model and v0.3 Foundation

Book Social currently models a small book catalog.

### Book

A book has:

- title
- slug
- description
- zero or more authors
- zero or more genres
- zero or more cover metadata records

Catalog reads preserve deterministic ordering: books by title and ID; related authors, genres, and
covers by their documented fields and ID. A catalog filter selects matching books without hiding the
other authors or genres of a selected book.

### Author

An author has:

- name parts
- slug
- description

Author pages are addressed by slug:

```text
/authors/{authorSlug}
```

### Genre

A genre has:

- name
- slug
- description

Genre filtering uses:

```text
/books?genre={genreSlug}
```

### Catalog

The catalog can:

- list books
- filter by author slug
- filter by genre slug
- combine author and genre filters
- open book details by book slug

### Cover

A cover is URL metadata associated with a book. Its variant and optional technical metadata are
available to the read-side. Book details use the `front` variant when present; otherwise the UI uses
a CSS placeholder. Uploading, proxying, or storing cover files is not part of the current model.

The normalized schema is described in [database_v0_2.md](database_v0_2.md).

### Curated Catalog Operations

v0.3.2 lets a deliberately appointed administrator find, create, and correct demo books while
preserving stable public book slugs. Operators select existing authors and genres, may set or clear
one front-cover URL, and produce one audit record for each successful create or correction,
including no-op corrections. Catalog, relationship, cover and audit writes are atomic. Audit records
are append-only through the application API, not protected against a database owner using direct
SQL. The internal authorization capability does not enter page models or templates.

Missing-book feedback provides an authenticated MPA form. Its original non-persistent flow was
accepted; the v0.3.2 service now saves a request before emitting one metadata-only structured event.
The stored row contains a generated ID, submitting user ID, normalized title, optional author text
and UTC creation time. Submitted title/author values are not logged.
The user comes from session identity, an absent author is stored as an empty string, and deleting
the submitting account will cascade to its requests. Repeated submissions are separate requests;
there is no required catalog book/author reference or account-data snapshot.

Real-database HTTP acceptance passes on SQLite/PostgreSQL. The user confirmed persistence browser
review on 2026-10-25 (Zen Browser, SQLite).
Operator listing and processed/rejected decisions are planned for v0.4.3; the v0.3.2 model adds no
processing status.
Catalog deletion, author/genre administration,
external import, and cover asset handling remain deferred. Current and planned rules are in
[Curated Catalog Operations v0.3.2 Contract](catalog_operations_v0_3_2.md).

### User and Authentication Foundation

The current `users` module provides:

- registration input normalization and validation;
- server-owned assignment of the ordinary `user` role;
- bcrypt hashing with a minimum of 12 Unicode characters and a maximum of 72 UTF-8 bytes;
- credential verification with one neutral invalid-credentials outcome;
- a dummy bcrypt verification path for an unknown account;
- DB-backed opaque-session creation, lookup, expiry, and invalidation;
- current-user identity without password hash, with an internal `IsAdmin` capability loaded from
  the current database role on each authenticated request.

Sessions have a seven-day absolute lifetime without sliding renewal. The raw high-entropy token is
browser-facing; persistence receives only its 32-byte SHA-256 hash. Missing, invalid, or expired
sessions represent an anonymous request, not an internal server failure.

The HTTP layer exposes registration, login, logout, and protected `/me` MPA routes. Successful
registration creates the user, ordinary role assignment, and the first hashed-token session
atomically; successful login creates a new session. Navigation receives only the current user's
display name. Passwords, hashes, tokens, role internals, and private-library data remain outside
page models.

### Library, Shelves, Tags

The schema still contains legacy/demo `library`, `shelves`, and `tags` structures. They are not the
user-facing personal-library model.

Migrations `000004` and `000005` plus the `library` module provide the private `library_items`
behavior: one owner, one catalog book, a uniqueness rule, `added_at`, one of `want_to_read`,
`reading`, or `read`, nullable start/finish timestamps, and an optimistic-locking version. Its
service validates owner IDs and book slugs, translates catalog and storage errors to library
application errors, and returns only catalog data plus item metadata. Authenticated users can add a
catalog book, list only their own items at `/me/library`, change an owner-scoped status, and remove
an item only after an explicit confirmation. The detailed rules are in the
[Private Library v0.3 Contract](library_v0_3.md).

## Current Design Rules

- Keep database details out of templates.
- Use page/view models for rendering.
- Keep SQL in repository implementations.
- Keep handler, service, and repository responsibilities separate.
- Keep password hashes and raw session tokens out of user/page models, errors, logs, and test
  diagnostics.
- Keep private-library behavior within the v0.3 roadmap scope.
