-- Existing external covers retain nullable technical metadata and no storage key.
ALTER TABLE covers ADD COLUMN storage_key TEXT NULL
    CONSTRAINT ck_covers_local_file CHECK (
        storage_key IS NULL OR (
            variant = 'front'
            AND checksum_sha256 IS NOT NULL
            AND LENGTH(checksum_sha256) = 64
            AND checksum_sha256 NOT GLOB '*[^0-9a-f]*'
            AND mime_type IS NOT NULL
            AND mime_type IN ('image/jpeg', 'image/png', 'image/webp')
            AND storage_key = checksum_sha256 || CASE mime_type
                WHEN 'image/jpeg' THEN '.jpg'
                WHEN 'image/png' THEN '.png'
                WHEN 'image/webp' THEN '.webp'
            END
            AND url = '/media/covers/' || storage_key
            AND byte_size IS NOT NULL AND byte_size BETWEEN 1 AND 5242880
            AND width IS NOT NULL AND width BETWEEN 1 AND 4096
            AND height IS NOT NULL AND height BETWEEN 1 AND 4096
            AND height > width AND width * height <= 12000000
        )
    );

ALTER TABLE covers ADD COLUMN thumbnail_storage_key TEXT NULL;
ALTER TABLE covers ADD COLUMN thumbnail_mime_type TEXT NULL;
ALTER TABLE covers ADD COLUMN thumbnail_byte_size INTEGER NULL;
ALTER TABLE covers ADD COLUMN thumbnail_width INTEGER NULL;
ALTER TABLE covers ADD COLUMN thumbnail_height INTEGER NULL;
-- Put the group constraint on the last column so down can drop it first.
ALTER TABLE covers ADD COLUMN thumbnail_checksum_sha256 TEXT NULL
    CONSTRAINT ck_covers_thumbnail_file CHECK (
        (
            thumbnail_storage_key IS NULL AND thumbnail_mime_type IS NULL
            AND thumbnail_byte_size IS NULL AND thumbnail_width IS NULL
            AND thumbnail_height IS NULL AND thumbnail_checksum_sha256 IS NULL
        ) OR (
            storage_key IS NOT NULL
            AND thumbnail_storage_key IS NOT NULL
            AND thumbnail_storage_key <> storage_key
            AND thumbnail_checksum_sha256 IS NOT NULL
            AND LENGTH(thumbnail_checksum_sha256) = 64
            AND thumbnail_checksum_sha256 NOT GLOB '*[^0-9a-f]*'
            AND thumbnail_mime_type IS NOT NULL AND thumbnail_mime_type = mime_type
            AND thumbnail_storage_key = thumbnail_checksum_sha256 || CASE thumbnail_mime_type
                WHEN 'image/jpeg' THEN '.jpg'
                WHEN 'image/png' THEN '.png'
                WHEN 'image/webp' THEN '.webp'
            END
            AND thumbnail_byte_size IS NOT NULL AND thumbnail_byte_size BETWEEN 1 AND 5242880
            AND thumbnail_width IS NOT NULL AND thumbnail_width BETWEEN 1 AND 4096
            AND thumbnail_height IS NOT NULL AND thumbnail_height BETWEEN 1 AND 4096
            AND thumbnail_width < width AND thumbnail_height < height
            AND thumbnail_width * height = thumbnail_height * width
        )
    );
