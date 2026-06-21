package books

import (
	"context"

	"github.com/LeeDark/book-social/internal/http/view"
)

type CatalogPageProvider interface {
	CatalogPage(ctx context.Context, filter BookFilter) (CatalogPageData, error)
	BookDetailsPage(ctx context.Context, slug string) (BookDetailsPageData, error)
}

type CatalogService struct {
	repo BookRepository
}

func NewCatalogService(repo BookRepository) *CatalogService {
	return &CatalogService{
		repo: repo,
	}
}

func (s *CatalogService) CatalogPage(ctx context.Context, filter BookFilter) (CatalogPageData, error) {
	books, err := s.repo.ListBooksFiltered(ctx, filter)
	if err != nil {
		return CatalogPageData{}, err
	}

	return CatalogPageData{
		Page: view.Page{
			Title:       "Books",
			Description: "Catalog page",
			ActiveNav:   "catalog",
			Nav:         view.MainNavigation(),
			Breadcrumbs: []view.Breadcrumb{
				{Label: "Home", Href: "/"},
				{Label: "Catalog"},
			},
		},
		Books: mapBooksToCards(books),
	}, nil
}

func (s *CatalogService) BookDetailsPage(ctx context.Context, slug string) (BookDetailsPageData, error) {
	book, err := s.repo.GetBookBySlug(ctx, slug)
	if err != nil {
		return BookDetailsPageData{}, err
	}

	return BookDetailsPageData{
		Page: view.Page{
			Title:       book.Genre.Name + ": " + book.Title,
			Description: "Book Details page",
			ActiveNav:   "catalog",
			Nav:         view.MainNavigation(),
			Breadcrumbs: []view.Breadcrumb{
				{Label: "Home", Href: "/"},
				{Label: "Catalog", Href: "/books"},
				{Label: book.Genre.Name + ": " + book.Title},
			},
		},
		Book: mapBookToDetailsView(book),
	}, nil
}
