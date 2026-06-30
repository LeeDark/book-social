# Domain Model

This document describes domain concepts. Database column details live in:

- [Database v0.1](database_v0_1.md)
- [Database v0.2 target](database_v0_2.md)

## Current v0.1 Model

Book Social currently models a small book catalog.

### Book

A book has:

- title
- slug
- description
- one author
- one genre

Current limitation:

- v0.1 supports only one author and one genre per book.

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

### User, Library, Shelves, Tags

The v0.1 schema includes users, roles, shelves, tags, and library tables, but current user-facing behavior is focused on the public catalog.

User accounts, authentication, and library workflows are planned later.

## v0.2 Target

v0.2 should normalize the catalog model:

- books can have multiple authors
- books can have multiple genres
- covers can be stored as URL metadata
- `library` can become `library_items`
- tags can move to a join table

The target schema is described in [database_v0_2.md](database_v0_2.md).

## Current Design Rules

- Keep database details out of templates.
- Use page/view models for rendering.
- Keep SQL in repository implementations.
- Keep handler, service, and repository responsibilities separate.
- Do not add auth/library behavior until the data and roadmap are clearer.
