package books

import (
	"context"
	"errors"
	"net/url"
	"regexp"
	"slices"
	"strings"
	"time"
	"unicode/utf8"
)

var (
	ErrOperatorForbidden = errors.New("catalog administrator required")
	ErrSlugTaken         = errors.New("book slug already exists")
	ErrCatalogInternal   = errors.New("catalog operation failed")
)

type CatalogValidationError struct {
	Field   string
	Message string
}

func (e CatalogValidationError) Error() string {
	return e.Field + ": " + e.Message
}

// BookCorrection deliberately has no slug field: public URLs cannot be renamed.
type BookCorrection struct {
	Title         string
	Description   string
	AuthorIDs     []int
	GenreIDs      []int
	FrontCoverURL string
}

type BookCreation struct {
	BookCorrection
	Slug string
}

type OperatorBook struct {
	ID   int
	Slug string
	BookCorrection
}

type CatalogOptions struct {
	Authors []Author
	Genres  []Genre
}

type CatalogAudit struct {
	ActorUserID int
	BookID      int
	Action      string
	Summary     string
	OccurredAt  time.Time
}

type OperatorReader interface {
	IsAdministrator(ctx context.Context, userID int) (bool, error)
	FindBooks(ctx context.Context, query string) ([]Book, error)
	GetOperatorBook(ctx context.Context, bookID int) (OperatorBook, error)
	ListOptions(ctx context.Context) (CatalogOptions, error)
}

// OperatorTransaction exposes only the writes owned by catalog maintenance.
// Audit records can only be appended through this application boundary.
type OperatorTransaction interface {
	IsAdministrator(ctx context.Context, userID int) (bool, error)
	GetOperatorBook(ctx context.Context, bookID int) (OperatorBook, error)
	ValidateRelationships(ctx context.Context, authorIDs, genreIDs []int) error
	CreateBook(ctx context.Context, input BookCreation) (int, error)
	CorrectBook(ctx context.Context, bookID int, input BookCorrection) error
	ReplaceRelationships(ctx context.Context, bookID int, authorIDs, genreIDs []int) error
	SetFrontCover(ctx context.Context, bookID int, coverURL string) error
	AppendAudit(ctx context.Context, audit CatalogAudit) error
}

type OperatorRepository interface {
	OperatorReader
	WithinCatalogTransaction(ctx context.Context, fn func(OperatorTransaction) error) error
}

type OperatorService struct {
	repo OperatorRepository
	now  func() time.Time
}

func NewOperatorService(repo OperatorRepository) *OperatorService {
	return &OperatorService{repo: repo, now: time.Now}
}

func requireOperator(ctx context.Context, repo interface {
	IsAdministrator(context.Context, int) (bool, error)
}, actorID int) error {
	if actorID <= 0 {
		return ErrOperatorForbidden
	}

	isAdmin, err := repo.IsAdministrator(ctx, actorID)
	if err != nil {
		return err
	}
	if !isAdmin {
		return ErrOperatorForbidden
	}
	return nil
}

func (s *OperatorService) Find(ctx context.Context, actorID int, query string) ([]Book, error) {
	if err := requireOperator(ctx, s.repo, actorID); err != nil {
		return nil, err
	}

	query = strings.TrimSpace(query)
	if query == "" {
		return []Book{}, nil
	}
	return s.repo.FindBooks(ctx, query)
}

func (s *OperatorService) Get(ctx context.Context, actorID, bookID int) (OperatorBook, error) {
	if err := requireOperator(ctx, s.repo, actorID); err != nil {
		return OperatorBook{}, err
	}
	if bookID <= 0 {
		return OperatorBook{}, CatalogValidationError{Field: "id", Message: "Choose a valid book."}
	}
	return s.repo.GetOperatorBook(ctx, bookID)
}

func (s *OperatorService) Options(ctx context.Context, actorID int) (CatalogOptions, error) {
	if err := requireOperator(ctx, s.repo, actorID); err != nil {
		return CatalogOptions{}, err
	}
	return s.repo.ListOptions(ctx)
}

func (s *OperatorService) Create(ctx context.Context, actorID int, input BookCreation) (int, error) {
	var bookID int
	err := s.repo.WithinCatalogTransaction(ctx, func(tx OperatorTransaction) error {
		if err := requireOperator(ctx, tx, actorID); err != nil {
			return err
		}

		normalized, err := normalizeBookCorrection(input.BookCorrection)
		if err != nil {
			return err
		}
		input.BookCorrection = normalized
		input.Slug = strings.TrimSpace(input.Slug)
		if len(input.Slug) > 255 || !catalogSlugPattern.MatchString(input.Slug) {
			return CatalogValidationError{
				Field:   "slug",
				Message: "Use lowercase letters, numbers, and single hyphens, up to 255 characters.",
			}
		}
		if err := tx.ValidateRelationships(ctx, input.AuthorIDs, input.GenreIDs); err != nil {
			return err
		}

		bookID, err = tx.CreateBook(ctx, input)
		if err != nil {
			return err
		}
		return s.finishWrite(ctx, tx, actorID, bookID, input.BookCorrection, "create", "created")
	})
	if err != nil {
		return 0, err
	}
	return bookID, nil
}

func (s *OperatorService) Correct(ctx context.Context, actorID, bookID int, input BookCorrection) error {
	return s.repo.WithinCatalogTransaction(ctx, func(tx OperatorTransaction) error {
		if err := requireOperator(ctx, tx, actorID); err != nil {
			return err
		}
		if bookID <= 0 {
			return CatalogValidationError{Field: "id", Message: "Choose a valid book."}
		}

		normalized, err := normalizeBookCorrection(input)
		if err != nil {
			return err
		}
		before, err := tx.GetOperatorBook(ctx, bookID)
		if err != nil {
			return err
		}
		if err := tx.ValidateRelationships(ctx, normalized.AuthorIDs, normalized.GenreIDs); err != nil {
			return err
		}

		if err := tx.CorrectBook(ctx, bookID, normalized); err != nil {
			return err
		}
		summary := changedBookGroups(before.BookCorrection, normalized)
		return s.finishWrite(ctx, tx, actorID, bookID, normalized, "correct", summary)
	})
}

func (s *OperatorService) finishWrite(
	ctx context.Context,
	tx OperatorTransaction,
	actorID, bookID int,
	input BookCorrection,
	action, summary string,
) error {
	if err := tx.ReplaceRelationships(ctx, bookID, input.AuthorIDs, input.GenreIDs); err != nil {
		return err
	}
	if err := tx.SetFrontCover(ctx, bookID, input.FrontCoverURL); err != nil {
		return err
	}

	return tx.AppendAudit(ctx, CatalogAudit{
		ActorUserID: actorID,
		BookID:      bookID,
		Action:      action,
		Summary:     summary,
		OccurredAt:  s.now().UTC(),
	})
}

var catalogSlugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

// SuggestSlug only keeps ASCII letters/digits and treats other characters as
// separators. It does not transliterate and is never a substitute for validation.
func SuggestSlug(title string) string {
	var result strings.Builder
	separator := false
	for _, char := range title {
		if char >= 'A' && char <= 'Z' {
			char += 'a' - 'A'
		}
		if char >= 'a' && char <= 'z' || char >= '0' && char <= '9' {
			if separator && result.Len() > 0 {
				result.WriteByte('-')
			}
			result.WriteRune(char)
			separator = false
		} else {
			separator = true
		}
	}
	return result.String()
}

func normalizeBookCorrection(input BookCorrection) (BookCorrection, error) {
	input.Title = strings.TrimSpace(input.Title)
	if input.Title == "" ||
		utf8.RuneCountInString(input.Title) > 255 ||
		!utf8.ValidString(input.Title) ||
		strings.ContainsRune(input.Title, 0) {
		return BookCorrection{}, CatalogValidationError{
			Field:   "title",
			Message: "Enter a title of 1 to 255 characters.",
		}
	}

	// PostgreSQL text rejects NUL and invalid UTF-8. Reject them in both dialects.
	if !utf8.ValidString(input.Description) || strings.ContainsRune(input.Description, 0) {
		return BookCorrection{}, CatalogValidationError{Field: "description", Message: "Enter valid text."}
	}

	var err error
	input.AuthorIDs, err = normalizeSelectedIDs(input.AuthorIDs, "author_ids")
	if err != nil {
		return BookCorrection{}, err
	}
	input.GenreIDs, err = normalizeSelectedIDs(input.GenreIDs, "genre_ids")
	if err != nil {
		return BookCorrection{}, err
	}

	input.FrontCoverURL = strings.TrimSpace(input.FrontCoverURL)
	if input.FrontCoverURL != "" {
		cover, err := url.ParseRequestURI(input.FrontCoverURL)
		if err != nil ||
			cover.Scheme != "https" ||
			cover.Hostname() == "" ||
			!utf8.ValidString(input.FrontCoverURL) ||
			strings.ContainsRune(input.FrontCoverURL, 0) {
			return BookCorrection{}, CatalogValidationError{
				Field:   "front_cover_url",
				Message: "Enter a valid HTTPS URL or leave it empty.",
			}
		}
	}
	return input, nil
}

func normalizeSelectedIDs(ids []int, field string) ([]int, error) {
	result := slices.Clone(ids)
	for _, id := range result {
		if id <= 0 {
			return nil, CatalogValidationError{Field: field, Message: "Choose existing entries."}
		}
	}
	slices.Sort(result)
	return slices.Compact(result), nil
}

func changedBookGroups(before, after BookCorrection) string {
	groups := make([]string, 0, 5)
	if before.Title != after.Title {
		groups = append(groups, "title")
	}
	if before.Description != after.Description {
		groups = append(groups, "description")
	}
	if !sameSelectedIDs(before.AuthorIDs, after.AuthorIDs) {
		groups = append(groups, "authors")
	}
	if !sameSelectedIDs(before.GenreIDs, after.GenreIDs) {
		groups = append(groups, "genres")
	}
	if before.FrontCoverURL != after.FrontCoverURL {
		groups = append(groups, "front_cover")
	}
	slices.Sort(groups)
	return strings.Join(groups, ",")
}

func sameSelectedIDs(first, second []int) bool {
	first = slices.Clone(first)
	second = slices.Clone(second)
	slices.Sort(first)
	slices.Sort(second)
	return slices.Equal(slices.Compact(first), slices.Compact(second))
}
