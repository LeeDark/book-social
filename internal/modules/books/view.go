package books

import (
	"fmt"

	"github.com/LeeDark/book-social/internal/http/view"
)

type CatalogPageData struct {
	view.Page
	Books []BookCardViewModel
}

type BookCardViewModel struct {
	Title       string
	Slug        string
	Description string
	AuthorName  string
	AuthorURL   string
	GenreName   string
	GenreURL    string
	BookURL     string
	CoverClass  string
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

		cards = append(cards, card)
	}

	return cards
}
