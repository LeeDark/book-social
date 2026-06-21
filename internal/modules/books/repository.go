package books

import "context"

type BookFilter struct {
	AuthorSlug string
	GenreSlug  string
}

type BookRepository interface {
	ListBooks(ctx context.Context) ([]Book, error)
	ListBooksFiltered(ctx context.Context, filter BookFilter) ([]Book, error)
	GetBookBySlug(ctx context.Context, slug string) (Book, error)
}
