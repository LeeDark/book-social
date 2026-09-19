-- 000004_add_library_items.down.sql
-- SQLite version

-- Do not silently delete private-library data during a rollback.
CREATE TABLE library_items_downgrade_check (
    is_valid INTEGER NOT NULL CHECK (is_valid = 1)
);

INSERT INTO library_items_downgrade_check (is_valid)
SELECT CASE
    WHEN NOT EXISTS (SELECT 1 FROM library_items) THEN 1
    ELSE 0
END;

DROP TABLE library_items_downgrade_check;

DROP INDEX idx_library_items_user_added_at_id;
DROP TABLE library_items;
