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
	GetByIDAndUserID(ctx context.Context, userID, itemID int) (Item, error)
	UpdateStatus(ctx context.Context, params UpdateStatusParams) (bool, error)
	Remove(ctx context.Context, userID, itemID, expectedVersion int) (bool, error)
}
