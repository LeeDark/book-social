package books

import (
	"context"
	"io"
	"strings"
)

type CoverAction string

const (
	CoverKeep     CoverAction = "keep"
	CoverUpload   CoverAction = "upload"
	CoverExternal CoverAction = "external"
	CoverRemove   CoverAction = "remove"
)

type CoverUploadFile struct {
	Reader   io.Reader
	Filename string
	MIMEType string
}

type CoverPreparer interface {
	ValidateAndStore(ctx context.Context, main CoverUploadFile, thumbnail *CoverUploadFile) (LocalFrontCover, error)
}

func normalizeCoverChange(input BookCorrection) (BookCorrection, error) {
	input.FrontCoverURL = strings.TrimSpace(input.FrontCoverURL)
	if input.FrontCoverAction == "" {
		input.FrontCoverAction = CoverKeep
	}

	hasFiles := input.FrontCoverFile != nil || input.FrontCoverThumbnail != nil
	isValid := false
	switch input.FrontCoverAction {
	case CoverKeep, CoverRemove:
		isValid = input.FrontCoverURL == "" && !hasFiles
	case CoverExternal:
		isValid = input.FrontCoverURL != "" && !hasFiles
	case CoverUpload:
		isValid = input.FrontCoverURL == "" && input.FrontCoverFile != nil && input.FrontCoverFile.Reader != nil
	}

	if !isValid {
		return BookCorrection{}, CatalogValidationError{Field: "front_cover", Message: "Choose one cover action and supply only its URL or files."}
	}

	return input, nil
}

func (s *OperatorService) publishCover(ctx context.Context, actorID, bookID int, input BookCorrection) (*LocalFrontCover, error) {
	// Publish outside the SQL transaction, but never before authorization and
	// existence checks. The transaction rechecks both before changing references.
	if err := requireOperator(ctx, s.repo, actorID); err != nil {
		return nil, err
	}

	if bookID != 0 {
		if _, err := s.repo.GetOperatorBook(ctx, bookID); err != nil {
			return nil, err
		}
	}

	if s.coverPreparer == nil {
		return nil, ErrCatalogInternal
	}

	cover, err := s.coverPreparer.ValidateAndStore(ctx, *input.FrontCoverFile, input.FrontCoverThumbnail)
	if err != nil {
		return nil, err
	}

	if err := cover.Validate(); err != nil {
		return nil, err
	}

	return &cover, nil
}

func coverChanged(before OperatorBook, input BookCorrection, local *LocalFrontCover) bool {
	switch input.FrontCoverAction {
	case CoverKeep:
		return false
	case CoverRemove:
		return before.FrontCoverURL != ""
	case CoverExternal:
		return before.FrontCoverURL != input.FrontCoverURL
	case CoverUpload:
		if local == nil {
			return false
		}

		old := before.FrontCover
		if old == nil || old.StorageKey == nil || old.MIMEType == nil || old.ByteSize == nil ||
			old.Width == nil || old.Height == nil || old.ChecksumSHA256 == nil {
			return true
		}

		main := StoredCoverFile{
			StorageKey:     *old.StorageKey,
			MIMEType:       *old.MIMEType,
			ByteSize:       *old.ByteSize,
			Width:          *old.Width,
			Height:         *old.Height,
			ChecksumSHA256: *old.ChecksumSHA256,
		}
		if main != local.Main {
			return true
		}

		if old.Thumbnail == nil || local.Thumbnail == nil {
			return old.Thumbnail != nil || local.Thumbnail != nil
		}

		return *old.Thumbnail != *local.Thumbnail
	}

	return false
}
