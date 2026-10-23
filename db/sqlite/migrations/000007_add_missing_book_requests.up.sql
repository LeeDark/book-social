-- SQLite version

CREATE TABLE missing_book_requests (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    title TEXT NOT NULL,
    author TEXT NOT NULL,
    created_at TEXT NOT NULL,

    CONSTRAINT ck_missing_book_requests_title_length
        CHECK (LENGTH(title) BETWEEN 1 AND 255),
    CONSTRAINT ck_missing_book_requests_author_length
        CHECK (LENGTH(author) <= 255),
    -- PostgreSQL text rejects NUL; enforce the same constraint in SQLite.
    CONSTRAINT ck_missing_book_requests_text_nul
        CHECK (INSTR(title, CHAR(0)) = 0 AND INSTR(author, CHAR(0)) = 0),
    CONSTRAINT fk_missing_book_requests_user
        FOREIGN KEY (user_id) REFERENCES users(id)
            ON UPDATE RESTRICT
            ON DELETE CASCADE
);
