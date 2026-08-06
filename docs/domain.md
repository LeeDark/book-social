# Domain Model

This document describes domain concepts. Database column details live in:

- [Database v0.1](database_v0_1.md)
- [Database v0.2](database_v0_2.md)

## Current v0.2 Catalog Model

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

### User, Library, Shelves, Tags

The schema still contains legacy/demo `library`, `shelves`, and `tags` structures. They are not
the user-facing personal-library model. Authentication and the final `library_items` model are
deferred to later v0.2 and v0.3 work.

## Current Design Rules

- Keep database details out of templates.
- Use page/view models for rendering.
- Keep SQL in repository implementations.
- Keep handler, service, and repository responsibilities separate.
- Do not add auth/library behavior until the data and roadmap are clearer.
