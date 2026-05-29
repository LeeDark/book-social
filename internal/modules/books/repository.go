package books

import "context"

type BookRepository interface {
	ListBooks(ctx context.Context) ([]Book, error)
	//GetBookBySlug(ctx context.Context, slug string) (Book, error)
}

type CatalogRepository struct {
}

func (repo *CatalogRepository) ListBooks(ctx context.Context) ([]Book, error) {
	return nil, nil
}
