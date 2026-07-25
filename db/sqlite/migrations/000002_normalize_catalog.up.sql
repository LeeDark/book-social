-- 000002_normalize_catalog.up.sql
-- SQLite version

-- SQLite rewrites foreign-key references when a table is renamed. Rebuild the
-- unused legacy library table alongside books so its data continues to refer
-- to the replacement books table without disabling foreign-key enforcement.
ALTER TABLE library RENAME TO library_v0_1;
ALTER TABLE books RENAME TO books_v0_1;

CREATE TABLE books (
       id INTEGER PRIMARY KEY AUTOINCREMENT,
       title TEXT NOT NULL,
       slug TEXT NOT NULL UNIQUE,
       description TEXT NULL
);

INSERT INTO books (id, title, slug, description)
SELECT id, title, slug, description
FROM books_v0_1;

CREATE TABLE book_authors (
    book_id INTEGER NOT NULL,
    author_id INTEGER NOT NULL,

    PRIMARY KEY (book_id, author_id),

    CONSTRAINT fk_book_authors_book
        FOREIGN KEY (book_id) REFERENCES books(id)
            ON UPDATE CASCADE
            ON DELETE CASCADE,

    CONSTRAINT fk_book_authors_author
        FOREIGN KEY (author_id) REFERENCES authors(id)
            ON UPDATE CASCADE
            ON DELETE CASCADE
);

CREATE INDEX idx_book_authors_author ON book_authors(author_id);

CREATE TABLE book_genres (
    book_id INTEGER NOT NULL,
    genre_id INTEGER NOT NULL,

    PRIMARY KEY (book_id, genre_id),

    CONSTRAINT fk_book_genres_book
        FOREIGN KEY (book_id) REFERENCES books(id)
            ON UPDATE CASCADE
            ON DELETE CASCADE,

    CONSTRAINT fk_book_genres_genre
        FOREIGN KEY (genre_id) REFERENCES genres(id)
            ON UPDATE CASCADE
            ON DELETE CASCADE
);

CREATE INDEX idx_book_genres_genre ON book_genres(genre_id);

CREATE TABLE covers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    book_id INTEGER NOT NULL,
    variant TEXT NOT NULL,
    url TEXT NOT NULL,
    mime_type TEXT NULL,
    byte_size INTEGER NULL CHECK (byte_size IS NULL OR byte_size >= 0),
    width INTEGER NULL CHECK (width IS NULL OR width >= 0),
    height INTEGER NULL CHECK (height IS NULL OR height >= 0),
    checksum_sha256 TEXT NULL,

    CONSTRAINT uq_covers_book_variant UNIQUE (book_id, variant),
    CONSTRAINT fk_covers_book
        FOREIGN KEY (book_id) REFERENCES books(id)
            ON UPDATE CASCADE
            ON DELETE CASCADE
);

CREATE INDEX idx_covers_book ON covers(book_id);

INSERT INTO book_authors (book_id, author_id)
SELECT id, book_author_id
FROM books_v0_1
WHERE book_author_id IS NOT NULL;

INSERT INTO book_genres (book_id, genre_id)
SELECT id, book_genre_id
FROM books_v0_1
WHERE book_genre_id IS NOT NULL;

-- Abort the transaction if copying v0.1 relationships lost or added a row.
CREATE TABLE catalog_migration_check (
    is_valid INTEGER NOT NULL CHECK (is_valid = 1)
);

INSERT INTO catalog_migration_check (is_valid)
SELECT CASE
    WHEN
        (SELECT COUNT(*) FROM books_v0_1 WHERE book_author_id IS NOT NULL) =
        (SELECT COUNT(*) FROM book_authors)
        AND
        (SELECT COUNT(*) FROM books_v0_1 WHERE book_genre_id IS NOT NULL) =
        (SELECT COUNT(*) FROM book_genres)
    THEN 1
    ELSE 0
END;

DROP TABLE catalog_migration_check;

CREATE TABLE library_v0_2 (
     id INTEGER PRIMARY KEY AUTOINCREMENT,
     library_shelf_id INTEGER NOT NULL,
     library_book_id  INTEGER NOT NULL,
     library_tag_id   INTEGER NULL,

     CONSTRAINT uq_library_shelf_book UNIQUE (library_shelf_id, library_book_id),

     CONSTRAINT fk_library_shelf
         FOREIGN KEY (library_shelf_id) REFERENCES shelves(id)
             ON UPDATE CASCADE
             ON DELETE RESTRICT,

     CONSTRAINT fk_library_book
         FOREIGN KEY (library_book_id) REFERENCES books(id)
             ON UPDATE CASCADE
             ON DELETE RESTRICT,

     CONSTRAINT fk_library_tag
         FOREIGN KEY (library_tag_id) REFERENCES tags(id)
             ON UPDATE CASCADE
             ON DELETE SET NULL
);

CREATE INDEX idx_library_v0_2_shelf ON library_v0_2(library_shelf_id);
CREATE INDEX idx_library_v0_2_book ON library_v0_2(library_book_id);
CREATE INDEX idx_library_v0_2_tag ON library_v0_2(library_tag_id);

INSERT INTO library_v0_2 (id, library_shelf_id, library_book_id, library_tag_id)
SELECT id, library_shelf_id, library_book_id, library_tag_id
FROM library_v0_1;

DROP TABLE library_v0_1;
DROP TABLE books_v0_1;

ALTER TABLE library_v0_2 RENAME TO library;
