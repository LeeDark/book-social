package books

import (
	"context"

	"github.com/LeeDark/book-social/internal/http/view"
)

type CatalogPageProvider interface {
	CatalogPage(ctx context.Context, filter BookFilter) (CatalogPageData, error)
	BookDetailsPage(ctx context.Context, slug string) (BookDetailsPageData, error)
	AuthorPage(ctx context.Context, slug string) (AuthorPageData, error)
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

func (s *CatalogService) AuthorPage(ctx context.Context, slug string) (AuthorPageData, error) {
	author, err := s.repo.GetAuthorBySlug(ctx, slug)
	if err != nil {
		return AuthorPageData{}, err
	}

	authorBooks, err := s.repo.ListBooksFiltered(ctx, BookFilter{AuthorSlug: slug})
	if err != nil {
		return AuthorPageData{}, err
	}

	authorName := author.FirstName + " " + author.SecondName + " " + author.SurName

	return AuthorPageData{
		Page: view.Page{
			Title:       authorName,
			Description: "Author page",
			ActiveNav:   "authors",
			Nav:         view.MainNavigation(),
			Breadcrumbs: []view.Breadcrumb{
				{Label: "Home", Href: "/"},
				{Label: "Authors", Href: "/authors"},
				{Label: authorName},
			},
		},
		Author: mapAuthorToView(author),
		Books:  mapBooksToCards(authorBooks),
	}, nil
}
