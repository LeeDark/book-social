# Domain Model

This document describes domain concepts. Database column details live in:

- [Database v0.1](database_v0_1.md)
- [Database v0.2](database_v0_2.md)

## Current v0.2.5 Model

Book Social currently models a small book catalog.

### Book

A book has:

- title
- slug
- description
- zero or more authors
- zero or more genres
- zero or more cover metadata records

Catalog reads preserve deterministic ordering: books by title and ID; related authors, genres,
and covers by their documented fields and ID. A catalog filter selects matching books without
hiding the other authors or genres of a selected book.

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
available to the read-side. Book details use the `front` variant when present; otherwise the UI
uses a CSS placeholder. Uploading, proxying, or storing cover files is not part of the current
model.

The normalized schema is described in [database_v0_2.md](database_v0_2.md).

### User and Authentication Foundation

The current `users` module provides internal foundations only:

- registration input normalization and validation;
- server-owned assignment of the ordinary `user` role;
- bcrypt hashing with a minimum of 12 Unicode characters and a maximum of 72 UTF-8 bytes;
- credential verification with one neutral invalid-credentials outcome;
- a dummy bcrypt verification path for an unknown account;
- DB-backed opaque-session creation, lookup, expiry, and invalidation;
- minimal current-user identity without password hash or role internals.

Sessions have a seven-day absolute lifetime without sliding renewal. The raw high-entropy token is
browser-facing; persistence receives only its 32-byte SHA-256 hash. Missing, invalid, or expired
sessions represent an anonymous request, not an internal server failure.

The HTTP layer has a typed current-user context and a testable authentication guard. Registration,
login, logout, production `/me`, auth navigation, and flashes are not current routes; they belong to
v0.2.6.

### Library, Shelves, Tags

The schema still contains legacy/demo `library`, `shelves`, and `tags` structures. They are not
the user-facing personal-library model. The user-facing authentication workflow and final
`library_items` model are deferred to v0.2.6 and v0.3 respectively.

## Current Design Rules

- Keep database details out of templates.
- Use page/view models for rendering.
- Keep SQL in repository implementations.
- Keep handler, service, and repository responsibilities separate.
- Keep password hashes and raw session tokens out of user/page models, errors, logs, and test
  diagnostics.
- Keep user-facing auth and library behavior within their v0.2.6 and v0.3 roadmap scopes.
