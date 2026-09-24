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
		Status:  ReadingStatusWantToRead,
		Version: 1,
		AddedAt: s.now().UTC(),
	})
	return mapRepositoryError(err)
}

func (s *Service) UpdateStatus(
	ctx context.Context,
	userID, itemID int,
	status ReadingStatus,
	expectedVersion int,
) error {
	if s == nil || s.repo == nil {
		return ErrInternal
	}
	if userID <= 0 || itemID <= 0 || expectedVersion <= 0 || !isValidReadingStatus(status) {
		return ErrInvalidInput
	}

	item, err := s.repo.GetByIDAndUserID(ctx, userID, itemID)
	if err != nil {
		return mapRepositoryError(err)
	}
	if item.Status == status {
		return nil
	}

	startedAt, finishedAt := transitionTimestamps(item, status, s.now().UTC())
	updated, err := s.repo.UpdateStatus(ctx, UpdateStatusParams{
		UserID:          userID,
		ItemID:          itemID,
		ExpectedVersion: expectedVersion,
		Status:          status,
		StartedAt:       startedAt,
		FinishedAt:      finishedAt,
	})
	if err != nil {
		return mapRepositoryError(err)
	}
	if updated {
		return nil
	}

	item, err = s.repo.GetByIDAndUserID(ctx, userID, itemID)
	if err != nil {
		return mapRepositoryError(err)
	}
	if item.Status == status {
		return nil
	}
	return ErrItemVersionConflict
}

func (s *Service) Remove(ctx context.Context, userID, itemID, expectedVersion int) error {
	if s == nil || s.repo == nil {
		return ErrInternal
	}
	if userID <= 0 || itemID <= 0 || expectedVersion <= 0 {
		return ErrInvalidInput
	}

	removed, err := s.repo.Remove(ctx, userID, itemID, expectedVersion)
	if err != nil {
		return mapRepositoryError(err)
	}
	if removed {
		return nil
	}
	if _, err := s.repo.GetByIDAndUserID(ctx, userID, itemID); err != nil {
		return mapRepositoryError(err)
	}
	return ErrItemVersionConflict
}

func isValidReadingStatus(status ReadingStatus) bool {
	switch status {
	case ReadingStatusWantToRead, ReadingStatusReading, ReadingStatusRead:
		return true
	default:
		return false
	}
}

func transitionTimestamps(item Item, status ReadingStatus, now time.Time) (*time.Time, *time.Time) {
	switch status {
	case ReadingStatusWantToRead:
		return nil, nil
	case ReadingStatusReading:
		if item.StartedAt != nil {
			return copyTime(item.StartedAt), nil
		}
		return timePointer(now), nil
	case ReadingStatusRead:
		return copyTime(item.StartedAt), timePointer(now)
	default:
		return nil, nil
	}
}

func timePointer(value time.Time) *time.Time {
	value = value.UTC()
	return &value
}

func copyTime(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	return timePointer(*value)
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
	case errors.Is(err, ErrItemVersionConflict):
		return ErrItemVersionConflict
	case errors.Is(err, ErrForbidden):
		return ErrForbidden
	default:
		return ErrInternal
	}
}
