package books

import "github.com/LeeDark/book-social/internal/http/view"

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

type CatalogPageData struct {
	view.Page
	Books []BookCardViewModel
}
