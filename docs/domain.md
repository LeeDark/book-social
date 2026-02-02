# Domain Model v0.1

Goal: a simple schema (no many-to-many yet) to quickly create tables and start building features.

## Conventions
- Naming: snake_case for fields.
- Passwords are never stored in plain form: `password_hash` only.
- `required/optional` marked explicitly.
- Foreign keys are marked as `FK -> table.column`.
- Table naming: plural (roles, users, books, ...).

---

## Roles / Users

### roles
- id (PK)
- role_name (required, unique)
- is_admin (required)

### users
- id (PK)
- first_name (required)
- second_name (optional)
- sur_name (optional)
- login (required, unique)
- password_hash (required)
- email (required, unique)
- role_id (required, FK -> roles.id)

Relations:
- roles 1 -> N users

---

## Books / Authors / Genres (v0.1)

### authors
- id (PK)
- first_name (required)
- second_name (optional)
- sur_name (optional)
- description (optional)

### genres
- id (PK)
- name (required, unique)
- description (optional)

### books
v0.1 assumes a single author and a single genre per book.
- id (PK)
- title (required)
- description (optional)
- author_id (optional, FK -> authors.id)
- genre_id  (optional, FK -> genres.id)

Relations (v0.1):
- authors 1 -> N books (books.author_id)
- genres  1 -> N books (books.genre_id)

---

## Covers (v0.1)
TBD in v0.1.

Direction decision:
- store cover as URL + metadata (not BLOB)

---

## Library / Shelves / Tags (v0.1)

### shelves
- id (PK)
- name (required, unique)
- description (optional)

### tags
- id (PK)
- name (required, unique)
- description (optional)

### library
A library record is: "book on shelf" + optionally a single tag.
- id (PK)
- shelf_id (required, FK -> shelves.id)
- book_id  (required, FK -> books.id)
- tag_id   (optional, FK -> tags.id)

Relations (v0.1):
- shelves 1 -> N library
- books   1 -> N library
- tags    1 -> N library (optional)

---

## Mermaid ERD (v0.1)

```mermaid
erDiagram
  ROLES ||--o{ USERS : has

  AUTHORS ||--o{ BOOKS : writes
  GENRES  ||--o{ BOOKS : classifies

  SHELVES ||--o{ LIBRARY : contains
  BOOKS   ||--o{ LIBRARY : listed
  TAGS    ||--o{ LIBRARY : tagged

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
    string description
  }

  GENRES {
    int id PK
    string name
    string description
  }

  BOOKS {
    int id PK
    string title
    string description
    int author_id FK
    int genre_id FK
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
    int shelf_id FK
    int book_id FK
    int tag_id FK
  }
```