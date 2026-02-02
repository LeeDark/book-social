-- schema_v0_2_target.sql
-- Target schema for Domain model v0.2 (no migrations yet)

-- ===== Roles / Users =====

CREATE TABLE roles (
   id INT UNSIGNED NOT NULL AUTO_INCREMENT,
   role_name VARCHAR(64) NOT NULL,
   is_admin BOOLEAN NOT NULL DEFAULT FALSE,
   PRIMARY KEY (id),
   UNIQUE KEY uq_roles_role_name (role_name)
) ENGINE=InnoDB;

CREATE TABLE users (
   id INT UNSIGNED NOT NULL AUTO_INCREMENT,
   first_name  VARCHAR(100) NOT NULL,
   second_name VARCHAR(100) NULL,
   sur_name    VARCHAR(100) NULL,
   login       VARCHAR(64)  NOT NULL,
   password_hash VARCHAR(255) NOT NULL,
   email       VARCHAR(254) NOT NULL,
   user_role_id INT UNSIGNED NOT NULL,

   PRIMARY KEY (id),
   UNIQUE KEY uq_users_login (login),
   UNIQUE KEY uq_users_email (email),
   KEY idx_users_role (user_role_id),

   CONSTRAINT fk_users_role
       FOREIGN KEY (user_role_id) REFERENCES roles(id)
           ON UPDATE CASCADE
           ON DELETE RESTRICT
) ENGINE=InnoDB;

-- ===== Authors / Genres / Books =====

CREATE TABLE authors (
     id INT UNSIGNED NOT NULL AUTO_INCREMENT,
     first_name  VARCHAR(100) NOT NULL,
     second_name VARCHAR(100) NULL,
     sur_name    VARCHAR(100) NULL,
     description TEXT NULL,

     PRIMARY KEY (id),
     KEY idx_authors_name (sur_name, first_name)
) ENGINE=InnoDB;

CREATE TABLE genres (
    id INT UNSIGNED NOT NULL AUTO_INCREMENT,
    name VARCHAR(120) NOT NULL,
    description TEXT NULL,

    PRIMARY KEY (id),
    UNIQUE KEY uq_genres_name (name)
) ENGINE=InnoDB;

CREATE TABLE books (
    id INT UNSIGNED NOT NULL AUTO_INCREMENT,
    title VARCHAR(255) NOT NULL,
    description TEXT NULL,

    PRIMARY KEY (id),
    KEY idx_books_title (title)
) ENGINE=InnoDB;

-- many-to-many: BookAuthor (LATER in v0.1) -> v0.2
CREATE TABLE book_authors (
    book_id INT UNSIGNED NOT NULL,
    author_id INT UNSIGNED NOT NULL,

    PRIMARY KEY (book_id, author_id),
    KEY idx_book_authors_author (author_id),

    CONSTRAINT fk_book_authors_book
      FOREIGN KEY (book_id) REFERENCES books(id)
          ON UPDATE CASCADE
          ON DELETE CASCADE,

    CONSTRAINT fk_book_authors_author
      FOREIGN KEY (author_id) REFERENCES authors(id)
          ON UPDATE CASCADE
          ON DELETE RESTRICT
) ENGINE=InnoDB;

-- many-to-many: BookGenre (LATER in v0.1) -> v0.2
CREATE TABLE book_genres (
    book_id INT UNSIGNED NOT NULL,
    genre_id INT UNSIGNED NOT NULL,

    PRIMARY KEY (book_id, genre_id),
    KEY idx_book_genres_genre (genre_id),

    CONSTRAINT fk_book_genres_book
     FOREIGN KEY (book_id) REFERENCES books(id)
         ON UPDATE CASCADE
         ON DELETE CASCADE,

    CONSTRAINT fk_book_genres_genre
     FOREIGN KEY (genre_id) REFERENCES genres(id)
         ON UPDATE CASCADE
         ON DELETE RESTRICT
) ENGINE=InnoDB;

-- ===== Covers =====
-- Choice: store URL + metadata (images live in object storage / CDN)
-- variant: 'original' | 'small' | 'medium' etc.
CREATE TABLE covers (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    book_id INT UNSIGNED NOT NULL,
    variant VARCHAR(32) NOT NULL DEFAULT 'original',

    url VARCHAR(2048) NOT NULL,
    mime_type VARCHAR(100) NULL,
    byte_size INT UNSIGNED NULL,
    width  INT UNSIGNED NULL,
    height INT UNSIGNED NULL,
    checksum_sha256 CHAR(64) NULL,

    PRIMARY KEY (id),
    UNIQUE KEY uq_covers_book_variant (book_id, variant),
    KEY idx_covers_book (book_id),

    CONSTRAINT fk_covers_book
        FOREIGN KEY (book_id) REFERENCES books(id)
            ON UPDATE CASCADE
            ON DELETE CASCADE
) ENGINE=InnoDB;

-- ===== Library / Shelf / Tag =====

CREATE TABLE shelves (
    id INT UNSIGNED NOT NULL AUTO_INCREMENT,
    name VARCHAR(120) NOT NULL,
    description TEXT NULL,

    PRIMARY KEY (id),
    UNIQUE KEY uq_shelves_name (name)
) ENGINE=InnoDB;

CREATE TABLE tags (
    id INT UNSIGNED NOT NULL AUTO_INCREMENT,
    name VARCHAR(120) NOT NULL,
    description TEXT NULL,

    PRIMARY KEY (id),
    UNIQUE KEY uq_tags_name (name)
) ENGINE=InnoDB;

-- v0.2: Library -> library_items (shelf + book)
CREATE TABLE library_items (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    library_shelf_id INT UNSIGNED NOT NULL,
    library_book_id  INT UNSIGNED NOT NULL,

    -- ⚠️ РЕКОМЕНДУЕМОЕ ПОЛЕ ДЛЯ РЕАЛЬНОЙ СОЦСЕТИ:
    user_id INT UNSIGNED NOT NULL,
    -- Если включишь user_id, то уникальность должна быть (user_id, shelf_id, book_id),
    -- и shelves обычно тоже становятся user-specific.

    PRIMARY KEY (id),
    UNIQUE KEY uq_library_items_pair (library_shelf_id, library_book_id),
    KEY idx_library_items_shelf (library_shelf_id),
    KEY idx_library_items_book  (library_book_id),

    CONSTRAINT fk_library_items_shelf
       FOREIGN KEY (library_shelf_id) REFERENCES shelves(id)
           ON UPDATE CASCADE
           ON DELETE RESTRICT,

    CONSTRAINT fk_library_items_book
       FOREIGN KEY (library_book_id) REFERENCES books(id)
           ON UPDATE CASCADE
           ON DELETE RESTRICT,

    -- Если включишь user_id:
    CONSTRAINT fk_library_items_user
        FOREIGN KEY (user_id) REFERENCES users(id)
        ON UPDATE CASCADE
        ON DELETE CASCADE
) ENGINE=InnoDB;

-- v0.2: LibraryTag (LATER in v0.1) -> library_item_tags
CREATE TABLE library_item_tags (
    library_item_id BIGINT UNSIGNED NOT NULL,
    tag_id INT UNSIGNED NOT NULL,

    PRIMARY KEY (library_item_id, tag_id),
    KEY idx_library_item_tags_tag (tag_id),

    CONSTRAINT fk_library_item_tags_item
       FOREIGN KEY (library_item_id) REFERENCES library_items(id)
           ON UPDATE CASCADE
           ON DELETE CASCADE,

    CONSTRAINT fk_library_item_tags_tag
       FOREIGN KEY (tag_id) REFERENCES tags(id)
           ON UPDATE CASCADE
           ON DELETE RESTRICT
) ENGINE=InnoDB;
