-- 000005_add_library_item_lifecycle.down.sql
-- SQLite version

-- Do not remove lifecycle information that the v0.3.0 schema cannot represent.
CREATE TABLE library_items_lifecycle_downgrade_check (
    is_valid INTEGER NOT NULL CHECK (is_valid = 1)
);

INSERT INTO library_items_lifecycle_downgrade_check (is_valid)
SELECT CASE
    WHEN NOT EXISTS (
        SELECT 1
        FROM library_items
        WHERE status != 'want_to_read'
            OR started_at IS NOT NULL
            OR finished_at IS NOT NULL
    ) THEN 1
    ELSE 0
END;

DROP TABLE library_items_lifecycle_downgrade_check;

DROP INDEX idx_library_items_user_added_at_id;

CREATE TABLE library_items_v0_3_0 (
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

INSERT INTO library_items_v0_3_0(id, user_id, book_id, added_at)
SELECT id, user_id, book_id, added_at
FROM library_items;

DROP TABLE library_items;
ALTER TABLE library_items_v0_3_0 RENAME TO library_items;

CREATE INDEX idx_library_items_user_added_at_id
    ON library_items(user_id, added_at DESC, id DESC);
