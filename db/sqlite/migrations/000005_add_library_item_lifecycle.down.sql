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

ALTER TABLE library_items DROP COLUMN version;
ALTER TABLE library_items DROP COLUMN finished_at;
ALTER TABLE library_items DROP COLUMN started_at;
ALTER TABLE library_items DROP COLUMN status;
