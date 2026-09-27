-- 000005_add_library_item_lifecycle.up.sql
-- SQLite version

ALTER TABLE library_items
    ADD COLUMN status TEXT NOT NULL DEFAULT 'want_to_read'
        CONSTRAINT ck_library_items_status
        CHECK (status IN ('want_to_read', 'reading', 'read'));

ALTER TABLE library_items
    ADD COLUMN started_at TEXT;

ALTER TABLE library_items
    ADD COLUMN finished_at TEXT;

ALTER TABLE library_items
    ADD COLUMN version INTEGER NOT NULL DEFAULT 1
        CONSTRAINT ck_library_items_version
        CHECK (version > 0);
