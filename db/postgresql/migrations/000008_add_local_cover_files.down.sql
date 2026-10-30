BEGIN;

-- Prevent attachment racing the check and losing its storage reference.
LOCK TABLE covers IN ACCESS EXCLUSIVE MODE;
CREATE TEMPORARY TABLE cover_files_downgrade_check (
    is_valid BOOLEAN NOT NULL CHECK (is_valid)
) ON COMMIT DROP;

INSERT INTO cover_files_downgrade_check (is_valid)
SELECT NOT EXISTS (SELECT 1 FROM covers WHERE storage_key IS NOT NULL);

DROP TABLE cover_files_downgrade_check;
ALTER TABLE covers
    DROP CONSTRAINT ck_covers_thumbnail_file,
    DROP CONSTRAINT ck_covers_local_file,
    DROP COLUMN thumbnail_checksum_sha256,
    DROP COLUMN thumbnail_height,
    DROP COLUMN thumbnail_width,
    DROP COLUMN thumbnail_byte_size,
    DROP COLUMN thumbnail_mime_type,
    DROP COLUMN thumbnail_storage_key,
    DROP COLUMN storage_key;

COMMIT;
