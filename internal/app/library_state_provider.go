package app

import (
	"context"

	"github.com/LeeDark/book-social/internal/modules/library"
)

type libraryBookStateService interface {
	ListBookStates(ctx context.Context, userID int, bookIDs []int) (map[int]library.BookState, error)
}
type LibraryStateProvider struct {
	service libraryBookStateService
}

func NewLibraryStateProvider(service libraryBookStateService) *LibraryStateProvider {
	return &LibraryStateProvider{service: service}
}

func (p *LibraryStateProvider) StatesForBooks(
	ctx context.Context,
	userID int,
	bookIDs []int,
) (map[int]string, error) {
	states, err := p.service.ListBookStates(ctx, userID, bookIDs)
	if err != nil {
		return nil, err
	}

	labels := make(map[int]string, len(states))
	for bookID, state := range states {
		labels[bookID] = libraryStateLabel(state.Status)
	}
	return labels, nil
}

func libraryStateLabel(status library.ReadingStatus) string {
	switch status {
	case library.ReadingStatusReading:
		return "Reading"
	case library.ReadingStatusRead:
		return "Read"
	default:
		return "Want to read"
	}
}
