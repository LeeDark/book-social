package library

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/LeeDark/book-social/internal/modules/books"
)

type Service struct {
	repo  Repository
	books BookFinder
	now   func() time.Time
}

func NewService(repo Repository, books BookFinder) *Service {
	return &Service{
		repo:  repo,
		books: books,
		now:   time.Now,
	}
}

func (s *Service) Add(ctx context.Context, userID int, bookSlug string) error {
	if s == nil || s.repo == nil || s.books == nil {
		return ErrInternal
	}
	if userID <= 0 {
		return ErrInvalidInput
	}

	bookSlug = strings.TrimSpace(bookSlug)
	if bookSlug == "" {
		return ErrInvalidInput
	}

	book, err := s.books.GetBookBySlug(ctx, bookSlug)
	if err != nil {
		if errors.Is(err, books.ErrBookNotFound) {
			return ErrBookNotFound
		}
		return ErrInternal
	}
	if book.ID <= 0 {
		return ErrInternal
	}

	err = s.repo.Add(ctx, AddItemParams{
		UserID:  userID,
		BookID:  book.ID,
		AddedAt: s.now().UTC(),
	})
	return mapRepositoryError(err)
}

func (s *Service) List(ctx context.Context, userID int) ([]Item, error) {
	if s == nil || s.repo == nil {
		return nil, ErrInternal
	}
	if userID <= 0 {
		return nil, ErrInvalidInput
	}

	items, err := s.repo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, mapRepositoryError(err)
	}
	return items, nil
}

func mapRepositoryError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, ErrItemAlreadyExists):
		return ErrItemAlreadyExists
	case errors.Is(err, ErrItemNotFound):
		return ErrItemNotFound
	case errors.Is(err, ErrForbidden):
		return ErrForbidden
	default:
		return ErrInternal
	}
}
