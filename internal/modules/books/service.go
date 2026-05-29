package books

import (
	"context"
	"fmt"

	"github.com/LeeDark/book-social/internal/http/view"
)

type CatalogPageProvider interface {
	CatalogPage(ctx context.Context) (CatalogPageData, error)
	//	BookDetailsPage(ctx context.Context, slug string) (BookDetailsPageData, error)
	//	AuthorPage(ctx context.Context, id int64) (AuthorPageData, error)
	//	BooksByAuthorPage(ctx context.Context, authorID int64) (CatalogPageData, error)
	//	BooksByGenrePage(ctx context.Context, genreSlug string) (CatalogPageData, error)
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

func coverClassForBook(id int) string {
	return fmt.Sprintf("cover-%d", id%5)
}

func mapBooksToCards(books []Book) []BookCardViewModel {
	cards := make([]BookCardViewModel, 0, len(books))
	for _, book := range books {
		card := BookCardViewModel{
			Title:       book.Title,
			Slug:        book.Slug,
			Description: book.Description,
			BookURL:     fmt.Sprintf("/books/%s", book.Slug),
			CoverClass:  coverClassForBook(book.ID),
			AuthorName:  book.Author.SurName,
			AuthorURL:   fmt.Sprintf("/authors/%d", book.Author.ID),
			GenreName:   book.Genre.Name,
			GenreURL:    fmt.Sprintf("/books?genre=%s", book.Genre.Slug),
		}

		//if book.Author != nil {
		//	card.AuthorName = book.Author.SurName
		//	card.AuthorURL = fmt.Sprintf("/authors/%d", book.Author.ID)
		//}

		//if book.Genre != nil {
		//	card.GenreName = book.Genre.Name
		//	card.GenreURL = fmt.Sprintf("/books?genre=%s", book.Genre.Slug)
		//}

		cards = append(cards, card)
	}

	return cards
}
