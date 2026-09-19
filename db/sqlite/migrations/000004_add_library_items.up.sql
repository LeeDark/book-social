-- 000004_add_library_items.up.sql
-- SQLite version

CREATE TABLE library_items (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    book_id INTEGER NOT NULL,
    added_at TEXT NOT NULL,

    CONSTRAINT uq_library_items_user_book UNIQUE (user_id, book_id),

    CONSTRAINT fk_library_items_user
        FOREIGN KEY (user_id) REFERENCES users(id)
            ON UPDATE CASCADE
            ON DELETE CASCADE,

    CONSTRAINT fk_library_items_book
        FOREIGN KEY (book_id) REFERENCES books(id)
            ON UPDATE CASCADE
            ON DELETE CASCADE
);

CREATE INDEX idx_library_items_user_added_at_id
    ON library_items(user_id, added_at DESC, id DESC);
