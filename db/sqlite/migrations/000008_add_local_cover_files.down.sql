-- Run transactionally with cover writers stopped. Keep local references intact on refusal.
CREATE TABLE cover_files_downgrade_check (
    is_valid INTEGER NOT NULL CHECK (is_valid = 1)
);

INSERT INTO cover_files_downgrade_check (is_valid)
SELECT CASE
    WHEN NOT EXISTS (SELECT 1 FROM covers WHERE storage_key IS NOT NULL) THEN 1
    ELSE 0
END;

DROP TABLE cover_files_downgrade_check;
ALTER TABLE covers DROP COLUMN thumbnail_checksum_sha256;
ALTER TABLE covers DROP COLUMN thumbnail_height;
ALTER TABLE covers DROP COLUMN thumbnail_width;
ALTER TABLE covers DROP COLUMN thumbnail_byte_size;
ALTER TABLE covers DROP COLUMN thumbnail_mime_type;
ALTER TABLE covers DROP COLUMN thumbnail_storage_key;
ALTER TABLE covers DROP COLUMN storage_key;
