-- 000005_add_library_item_lifecycle.down.sql
-- PostgreSQL version

BEGIN;

-- Do not remove lifecycle information that the v0.3.0 schema cannot represent.
CREATE TEMPORARY TABLE library_items_lifecycle_downgrade_check (
    is_valid BOOLEAN NOT NULL CHECK (is_valid)
) ON COMMIT DROP;

INSERT INTO library_items_lifecycle_downgrade_check (is_valid)
SELECT NOT EXISTS (
    SELECT 1
    FROM library_items
    WHERE status != 'want_to_read'
        OR started_at IS NOT NULL
        OR finished_at IS NOT NULL
);

DROP TABLE library_items_lifecycle_downgrade_check;

ALTER TABLE library_items
    DROP COLUMN version,
    DROP COLUMN finished_at,
    DROP COLUMN started_at,
    DROP COLUMN status;

COMMIT;
