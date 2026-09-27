-- 000005_add_library_item_lifecycle.up.sql
-- PostgreSQL version

BEGIN;

ALTER TABLE library_items
    ADD COLUMN status TEXT NOT NULL DEFAULT 'want_to_read',
    ADD COLUMN started_at TIMESTAMPTZ,
    ADD COLUMN finished_at TIMESTAMPTZ,
    ADD COLUMN version INTEGER NOT NULL DEFAULT 1,
    ADD CONSTRAINT ck_library_items_status
        CHECK (status IN ('want_to_read', 'reading', 'read')),
    ADD CONSTRAINT ck_library_items_version
        CHECK (version > 0);

COMMIT;
