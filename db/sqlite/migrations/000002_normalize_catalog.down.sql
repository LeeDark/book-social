-- 000002_normalize_catalog.down.sql
-- SQLite version

-- A v0.1 books row can represent at most one author and one genre, and cannot
-- represent covers. Refuse a lossy downgrade before changing any tables.
CREATE TABLE catalog_downgrade_check (
    is_valid INTEGER NOT NULL CHECK (is_valid = 1)
);

INSERT INTO catalog_downgrade_check (is_valid)
SELECT CASE
    WHEN NOT EXISTS (
        SELECT 1
        FROM book_authors
        GROUP BY book_id
        HAVING COUNT(*) > 1
    )
    AND NOT EXISTS (
        SELECT 1
        FROM book_genres
        GROUP BY book_id
        HAVING COUNT(*) > 1
    )
    AND NOT EXISTS (
        SELECT 1
        FROM covers
    )
    THEN 1
    ELSE 0
END;

DROP TABLE catalog_downgrade_check;

ALTER TABLE library RENAME TO library_v0_2;
ALTER TABLE books RENAME TO books_v0_2;

CREATE TABLE books (
       id INTEGER PRIMARY KEY AUTOINCREMENT,
       title TEXT NOT NULL,
       slug TEXT NOT NULL UNIQUE,
       description TEXT NULL,

       book_author_id INTEGER NULL,
       book_genre_id  INTEGER NULL,

       CONSTRAINT fk_books_author
           FOREIGN KEY (book_author_id) REFERENCES authors(id)
               ON UPDATE CASCADE
               ON DELETE SET NULL,

       CONSTRAINT fk_books_genre
           FOREIGN KEY (book_genre_id) REFERENCES genres(id)
               ON UPDATE CASCADE
               ON DELETE SET NULL
);

CREATE INDEX idx_books_author ON books(book_author_id);
CREATE INDEX idx_books_genre ON books(book_genre_id);

INSERT INTO books (id, title, slug, description, book_author_id, book_genre_id)
SELECT
    b.id,
    b.title,
    b.slug,
    b.description,
    (SELECT author_id FROM book_authors WHERE book_id = b.id),
    (SELECT genre_id FROM book_genres WHERE book_id = b.id)
FROM books_v0_2 AS b;

CREATE TABLE library_v0_1 (
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

INSERT INTO library_v0_1 (id, library_shelf_id, library_book_id, library_tag_id)
SELECT id, library_shelf_id, library_book_id, library_tag_id
FROM library_v0_2;

DROP TABLE library_v0_2;
DROP TABLE covers;
DROP TABLE book_genres;
DROP TABLE book_authors;
DROP TABLE books_v0_2;

ALTER TABLE library_v0_1 RENAME TO library;

CREATE INDEX idx_library_shelf ON library(library_shelf_id);
CREATE INDEX idx_library_book ON library(library_book_id);
CREATE INDEX idx_library_tag ON library(library_tag_id);
