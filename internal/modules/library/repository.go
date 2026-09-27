package library

import (
	"context"

	"github.com/LeeDark/book-social/internal/modules/books"
)

type BookFinder interface {
	GetBookBySlug(ctx context.Context, slug string) (books.Book, error)
}

type Repository interface {
	Add(ctx context.Context, params AddItemParams) error
	ListByUserID(ctx context.Context, userID int) ([]Item, error)
	ListBookStates(ctx context.Context, userID int, bookIDs []int) (map[int]BookState, error)
	// GetByIDAndUserID returns lifecycle fields only; Book and AddedAt are not populated.
	GetByIDAndUserID(ctx context.Context, userID, itemID int) (Item, error)
	// GetDetailByIDAndUserID also returns the item's book ID, title, and slug for confirmation pages.
	GetDetailByIDAndUserID(ctx context.Context, userID, itemID int) (Item, error)
	UpdateStatus(ctx context.Context, params UpdateStatusParams) (bool, error)
	Remove(ctx context.Context, userID, itemID, expectedVersion int) (bool, error)
}
