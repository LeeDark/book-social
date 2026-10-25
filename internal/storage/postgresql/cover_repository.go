package postgresql

import (
	"context"
	"database/sql"
	"errors"

	"github.com/LeeDark/book-social/internal/modules/books"
)

const coverColumns = `
    c.id, c.variant, c.url, c.mime_type, c.byte_size, c.width, c.height, c.checksum_sha256,
    c.storage_key, c.thumbnail_storage_key, c.thumbnail_mime_type, c.thumbnail_byte_size,
    c.thumbnail_width, c.thumbnail_height, c.thumbnail_checksum_sha256
`

// scanCover also accepts leading scan destinations for list queries such as book_id.
func scanCover(row interface{ Scan(...any) error }, leading ...any) (books.Cover, error) {
	var cover books.Cover
	var mimeType, checksum, storageKey sql.NullString
	var byteSize, width, height sql.NullInt64
	var thumbnailKey, thumbnailMIME, thumbnailChecksum sql.NullString
	var thumbnailBytes, thumbnailWidth, thumbnailHeight sql.NullInt64
	destinations := append(leading,
		&cover.ID, &cover.Variant, &cover.URL, &mimeType, &byteSize, &width, &height, &checksum,
		&storageKey, &thumbnailKey, &thumbnailMIME, &thumbnailBytes, &thumbnailWidth,
		&thumbnailHeight, &thumbnailChecksum,
	)
	if err := row.Scan(destinations...); err != nil {
		return books.Cover{}, err
	}
	cover.MIMEType = nullStringPointer(mimeType)
	cover.ByteSize = nullInt64Pointer(byteSize)
	cover.Width = nullIntPointer(width)
	cover.Height = nullIntPointer(height)
	cover.ChecksumSHA256 = nullStringPointer(checksum)
	cover.StorageKey = nullStringPointer(storageKey)
	if thumbnailKey.Valid {
		cover.Thumbnail = &books.StoredCoverFile{
			StorageKey:     thumbnailKey.String,
			MIMEType:       thumbnailMIME.String,
			ByteSize:       thumbnailBytes.Int64,
			Width:          int(thumbnailWidth.Int64),
			Height:         int(thumbnailHeight.Int64),
			ChecksumSHA256: thumbnailChecksum.String,
		}
	}
	return cover, nil
}

func (r *CatalogOperatorRepository) getFrontCover(ctx context.Context, bookID int) (books.Cover, error) {
	cover, err := scanCover(r.db.QueryRowContext(ctx, `
        SELECT `+coverColumns+`
        FROM covers c
        WHERE c.book_id = $1 AND c.variant = 'front'
    `, bookID))
	if errors.Is(err, sql.ErrNoRows) {
		return books.Cover{}, sql.ErrNoRows
	}
	if err != nil {
		return books.Cover{}, books.ErrCatalogInternal
	}
	return cover, nil
}

// SetLocalFrontCover stores metadata for files already validated/published by the
// caller. Use WithinCatalogTransaction to commit the reference with its audit.
func (r *CatalogOperatorRepository) SetLocalFrontCover(ctx context.Context, bookID int, cover books.LocalFrontCover) error {
	if err := cover.Validate(); err != nil {
		return err
	}
	var thumbnailKey, thumbnailMIME, thumbnailChecksum *string
	var thumbnailBytes *int64
	var thumbnailWidth, thumbnailHeight *int
	if thumbnail := cover.Thumbnail; thumbnail != nil {
		thumbnailKey = &thumbnail.StorageKey
		thumbnailMIME = &thumbnail.MIMEType
		thumbnailBytes = &thumbnail.ByteSize
		thumbnailWidth = &thumbnail.Width
		thumbnailHeight = &thumbnail.Height
		thumbnailChecksum = &thumbnail.ChecksumSHA256
	}
	_, err := r.db.ExecContext(ctx, `
        INSERT INTO covers (
            book_id, variant, url, storage_key, mime_type, byte_size, width, height,
            checksum_sha256, thumbnail_storage_key, thumbnail_mime_type, thumbnail_byte_size,
            thumbnail_width, thumbnail_height, thumbnail_checksum_sha256
        )
        VALUES ($1, 'front', $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
        ON CONFLICT (book_id, variant) DO UPDATE
        SET url = excluded.url,
            storage_key = excluded.storage_key,
            mime_type = excluded.mime_type,
            byte_size = excluded.byte_size,
            width = excluded.width,
            height = excluded.height,
            checksum_sha256 = excluded.checksum_sha256,
            thumbnail_storage_key = excluded.thumbnail_storage_key,
            thumbnail_mime_type = excluded.thumbnail_mime_type,
            thumbnail_byte_size = excluded.thumbnail_byte_size,
            thumbnail_width = excluded.thumbnail_width,
            thumbnail_height = excluded.thumbnail_height,
            thumbnail_checksum_sha256 = excluded.thumbnail_checksum_sha256
    `,
		bookID, cover.Main.URL(), cover.Main.StorageKey, cover.Main.MIMEType,
		cover.Main.ByteSize, cover.Main.Width, cover.Main.Height, cover.Main.ChecksumSHA256,
		thumbnailKey, thumbnailMIME, thumbnailBytes, thumbnailWidth, thumbnailHeight, thumbnailChecksum,
	)
	if err != nil {
		return books.ErrCatalogInternal
	}
	return nil
}
