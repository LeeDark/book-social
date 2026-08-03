# Database v0.2

Migration `000002_normalize_catalog` creates this schema after the v0.1 baseline migration.
The database/bootstrap path is v0.2, while application read repositories remain v0.1-shaped until
the dependent v0.2.3 branch.

Main changes from v0.1:

- books/authors become many-to-many
- books/genres become many-to-many
- covers store URL metadata
- legacy `library`, `shelves`, and `tags` remain preserved as demo/legacy structures
- the final user `library_items` model is deferred to v0.3

Slugs should remain part of catalog entities because current routes use book, author, and genre slugs.

```mermaid
erDiagram
    ROLES ||--o{ USERS : has

    BOOKS ||--o{ BOOK_AUTHORS : links
    AUTHORS ||--o{ BOOK_AUTHORS : links

    BOOKS ||--o{ BOOK_GENRES : links
    GENRES ||--o{ BOOK_GENRES : links

    BOOKS ||--o{ COVERS : has

    SHELVES ||--o{ LIBRARY : legacy_contains
    BOOKS   ||--o{ LIBRARY : legacy_lists
    TAGS    ||--o{ LIBRARY : legacy_tags

    ROLES {
        int id PK
        string role_name
        boolean is_admin
    }

    USERS {
        int id PK
        string first_name
        string second_name
        string sur_name
        string login
        string password_hash
        string email
        int role_id FK
    }

    AUTHORS {
        int id PK
        string first_name
        string second_name
        string sur_name
        string slug
        string description
    }

    GENRES {
        int id PK
        string name
        string slug
        string description
    }

    BOOKS {
        int id PK
        string slug
        string title
        string description
    }

    BOOK_AUTHORS {
        int book_id FK
        int author_id FK
    }

    BOOK_GENRES {
        int book_id FK
        int genre_id FK
    }

    COVERS {
        int id PK
        int book_id FK
        string variant
        string url
        string mime_type
        int byte_size
        int width
        int height
        string checksum_sha256
    }

    SHELVES {
        int id PK
        string name
        string description
    }

    TAGS {
        int id PK
        string name
        string description
    }

    LIBRARY {
        int id PK
        int library_shelf_id FK
        int library_book_id FK
        int library_tag_id FK
    }
```

The catalog migration does not create `library_items` or `library_item_tags`. Those tables require
user identity, reading status, privacy, and lifecycle rules that belong to the later personal
library work.

## Slugs and covers

`books.slug`, `authors.slug`, and `genres.slug` remain unique within their entity type and are
treated as stable public identifiers. A slug change requires a later redirect/compatibility
decision. Covers store only external URL and technical metadata; file upload and media storage are
out of scope.
