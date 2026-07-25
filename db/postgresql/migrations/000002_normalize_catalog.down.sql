-- 000002_normalize_catalog.down.sql
-- PostgreSQL version

BEGIN;

-- A CHECK violation aborts this transaction before a lossy downgrade.
CREATE TEMPORARY TABLE catalog_downgrade_check (
    is_valid BOOLEAN NOT NULL CHECK (is_valid)
) ON COMMIT DROP;

INSERT INTO catalog_downgrade_check (is_valid)
SELECT
    NOT EXISTS (
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
    );

DROP TABLE catalog_downgrade_check;

ALTER TABLE books
    ADD COLUMN book_author_id INTEGER NULL,
    ADD COLUMN book_genre_id INTEGER NULL;

UPDATE books AS b
SET book_author_id = ba.author_id
FROM book_authors AS ba
WHERE ba.book_id = b.id;

UPDATE books AS b
SET book_genre_id = bg.genre_id
FROM book_genres AS bg
WHERE bg.book_id = b.id;

ALTER TABLE books
    ADD CONSTRAINT fk_books_author
        FOREIGN KEY (book_author_id) REFERENCES authors(id)
            ON UPDATE CASCADE
            ON DELETE SET NULL,
    ADD CONSTRAINT fk_books_genre
        FOREIGN KEY (book_genre_id) REFERENCES genres(id)
            ON UPDATE CASCADE
            ON DELETE SET NULL;

CREATE INDEX idx_books_author ON books(book_author_id);
CREATE INDEX idx_books_genre ON books(book_genre_id);

DROP TABLE covers;
DROP TABLE book_genres;
DROP TABLE book_authors;

COMMIT;
