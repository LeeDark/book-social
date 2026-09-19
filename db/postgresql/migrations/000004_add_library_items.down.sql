-- 000004_add_library_items.down.sql
-- PostgreSQL version

BEGIN;

-- A CHECK violation aborts this transaction before private-library data is lost.
CREATE TEMPORARY TABLE library_items_downgrade_check (
    is_valid BOOLEAN NOT NULL CHECK (is_valid)
) ON COMMIT DROP;

INSERT INTO library_items_downgrade_check (is_valid)
SELECT NOT EXISTS (SELECT 1 FROM library_items);

DROP TABLE library_items_downgrade_check;

DROP INDEX idx_library_items_user_added_at_id;
DROP TABLE library_items;

COMMIT;
