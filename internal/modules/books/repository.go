package books

import "context"

type BookRepository interface {
	ListBooks(ctx context.Context) ([]Book, error)
}
