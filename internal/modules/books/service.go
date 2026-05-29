package books

import (
	"context"

	"github.com/LeeDark/book-social/internal/http/view"
)

type CatalogPageProvider interface {
	CatalogPage(ctx context.Context) (CatalogPageData, error)
}

type CatalogService struct {
	repo BookRepository
}

func NewCatalogService(repo BookRepository) *CatalogService {
	return &CatalogService{
		repo: repo,
	}
}

func (s *CatalogService) CatalogPage(ctx context.Context) (CatalogPageData, error) {
	books, err := s.repo.ListBooks(ctx)
	if err != nil {
		return CatalogPageData{}, err
	}

	return CatalogPageData{
		Page: view.Page{
			Title:       "Books",
			Description: "Catalog page",
			ActiveNav:   "catalog",
		},
		Books: mapBooksToCards(books),
	}, nil
}
