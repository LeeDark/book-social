package app

import (
	"bytes"
	"context"
	"errors"
	"io"

	"github.com/LeeDark/book-social/internal/modules/books"
	"github.com/LeeDark/book-social/internal/storage/covers"
)

type CoverPreparer struct{ store *covers.Store }

func NewCoverPreparer(store *covers.Store) *CoverPreparer { return &CoverPreparer{store: store} }

func (p *CoverPreparer) ValidateAndStore(ctx context.Context, main books.CoverUploadFile, thumbnail *books.CoverUploadFile) (books.LocalFrontCover, error) {
	mainData, mainFile, err := inspectUpload(ctx, main)
	if err != nil {
		return books.LocalFrontCover{}, err
	}

	local := books.LocalFrontCover{Main: storedCoverFile(mainFile)}
	var thumbnailData []byte
	if thumbnail != nil {
		data, file, err := inspectUpload(ctx, *thumbnail)
		if err != nil {
			return books.LocalFrontCover{}, err
		}

		thumbnailData = data
		metadata := storedCoverFile(file)
		local.Thumbnail = &metadata
	}

	if err := local.Validate(); err != nil {
		return books.LocalFrontCover{}, err
	}

	if _, err := p.store.Put(ctx, bytes.NewReader(mainData), mainFile.Key, mainFile.MIMEType); err != nil {
		return books.LocalFrontCover{}, err
	}
	if local.Thumbnail != nil {
		if _, err := p.store.Put(ctx, bytes.NewReader(thumbnailData), local.Thumbnail.StorageKey, local.Thumbnail.MIMEType); err != nil {
			return books.LocalFrontCover{}, err
		}
	}

	return local, nil
}

func inspectUpload(ctx context.Context, upload books.CoverUploadFile) ([]byte, covers.File, error) {
	if err := ctx.Err(); err != nil {
		return nil, covers.File{}, err
	}
	if upload.Reader == nil {
		return nil, covers.File{}, books.CatalogValidationError{Field: "front_cover", Message: "Choose an image file."}
	}

	data, err := io.ReadAll(io.LimitReader(upload.Reader, covers.MaxFileBytes+1))
	if err != nil {
		return nil, covers.File{}, err
	}

	file, err := covers.Inspect(data, upload.Filename, upload.MIMEType)
	if errors.Is(err, covers.ErrInvalidImage) {
		return nil, covers.File{}, books.CatalogValidationError{Field: "front_cover", Message: "Choose a valid portrait JPEG, PNG or static WebP within the file and dimension limits."}
	}

	return data, file, err
}

func storedCoverFile(file covers.File) books.StoredCoverFile {
	return books.StoredCoverFile{
		StorageKey:     file.Key,
		MIMEType:       file.MIMEType,
		ByteSize:       file.ByteSize,
		Width:          file.Width,
		Height:         file.Height,
		ChecksumSHA256: file.ChecksumSHA256,
	}
}
