package books

import (
	"fmt"

	"github.com/LeeDark/book-social/internal/http/view"
)

type CatalogPageData struct {
	view.Page
	Books []BookCardView
}

type BookCardView struct {
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

type BookDetailsPageData struct {
	view.Page
	Book BookDetailsView
	//Book BookCardView
}

type BookDetailsView struct {
	ID          int
	Title       string
	Slug        string
	Description string
	CoverClass  string

	Authors []AuthorLinkView
	Genres  []GenreLinkView
}

type AuthorLinkView struct {
	Name string
	URL  string
}

type GenreLinkView struct {
	Name string
	URL  string
}

func coverClassForBook(id int) string {
	return fmt.Sprintf("cover-%d", id%5)
}

func mapBooksToCards(books []Book) []BookCardView {
	cards := make([]BookCardView, 0, len(books))
	for _, book := range books {
		card := BookCardView{
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

func mapBookToDetailsView(book Book) BookDetailsView {
	details := BookDetailsView{
		ID:          book.ID,
		Title:       book.Title,
		Slug:        book.Slug,
		Description: book.Description,
		CoverClass:  coverClassForBook(book.ID),
	}

	return details
}
