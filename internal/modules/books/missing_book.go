package books

import (
	"context"
	"log/slog"
	"strings"
	"unicode/utf8"

	"github.com/LeeDark/book-social/internal/modules/users"
)

type MissingBookRequest struct {
	Title  string
	Author string
}

type MissingBookService struct {
	logger *slog.Logger
}

func NewMissingBookService(logger *slog.Logger) *MissingBookService {
	return &MissingBookService{logger: logger}
}

func (s *MissingBookService) Request(ctx context.Context, actorID int, requestID string, input MissingBookRequest) error {
	if actorID <= 0 {
		return users.ErrUnauthenticated
	}

	input.Title = strings.TrimSpace(input.Title)
	input.Author = strings.TrimSpace(input.Author)
	if input.Title == "" || !validMissingBookText(input.Title) {
		return CatalogValidationError{Field: "title", Message: "Enter a title of 1 to 255 characters."}
	}
	if !validMissingBookText(input.Author) {
		return CatalogValidationError{Field: "author", Message: "Enter an author of up to 255 characters or leave it empty."}
	}

	// Submitted text is deliberately excluded, even after validation: book fields
	// can contain personal information or credentials supplied by the reader.
	s.logger.InfoContext(ctx, "catalog_missing_book_requested",
		"request_id", requestID,
		"actor_id", actorID,
		"title_length", utf8.RuneCountInString(input.Title),
		"author_length", utf8.RuneCountInString(input.Author),
		"has_author", input.Author != "",
	)
	return nil
}

func validMissingBookText(text string) bool {
	return utf8.ValidString(text) &&
		!strings.ContainsRune(text, 0) &&
		utf8.RuneCountInString(text) <= 255
}
