package pages

import (
	"fmt"

	bookcomponents "github.com/LeeDark/book-social/internal/web/gomponents/components"
	g "maragu.dev/gomponents"
	gc "maragu.dev/gomponents/components"
	h "maragu.dev/gomponents/html"
)

func BooksPage(data BooksPageData) g.Node {
	return gc.HTML5(gc.HTML5Props{
		Title:    "Books gomponents Spike",
		Language: "en",
		Head: g.Group{
			h.Link(h.Rel("stylesheet"), h.Href("/static/css/pico.min.css")),
			h.Link(h.Rel("stylesheet"), h.Href("/static/css/app.css")),
			h.Link(h.Rel("stylesheet"), h.Href("/static/css/layout.css")),
			h.Link(h.Rel("stylesheet"), h.Href("/static/css/components/book-card.css")),
			h.Link(h.Rel("stylesheet"), h.Href("/static/css/components/cover.css")),
			h.Link(h.Rel("stylesheet"), h.Href("/static/css/components/empty-state.css")),
			h.Link(h.Rel("stylesheet"), h.Href("/static/css/pages/catalog.css")),
		},
		Body: g.Group{
			h.Main(
				h.Class("container"),
				h.Section(
					h.Class("page-header catalog-header"),
					h.P(h.Class("eyebrow"), g.Text("gomponents spike")),
					h.H1(g.Text("Books rendered with gomponents cards")),
					h.P(g.Text("This experimental route reuses the catalog service and renders only the book cards with gomponents.")),
				),
				booksContent(data.Books),
			),
		},
	})
}

func booksContent(books []bookcomponents.BookCardView) g.Node {
	if len(books) == 0 {
		return h.Section(
			h.Class("empty-state"),
			h.Aria("labelledby", "gomponents-catalog-empty-heading"),
			h.Div(h.Class("empty-state__icon"), h.Aria("hidden", "true"), g.Text("BS")),
			h.P(h.Class("eyebrow"), g.Text("No results")),
			h.H2(h.ID("gomponents-catalog-empty-heading"), g.Text("No books found")),
			h.P(g.Text("The catalog does not have books for this view yet.")),
		)
	}

	return h.Section(
		h.Class("catalog-results"),
		h.Aria("labelledby", "gomponents-catalog-results-heading"),
		h.Div(
			h.Class("catalog-results__bar"),
			h.Div(
				h.P(h.Class("eyebrow"), g.Text("Results")),
				h.H2(h.ID("gomponents-catalog-results-heading"), g.Text("Catalog books")),
			),
			h.P(g.Text(bookCountText(len(books)))),
		),
		h.Div(
			h.Class("book-grid catalog-grid"),
			g.Map(books, func(book bookcomponents.BookCardView) g.Node {
				return bookcomponents.BookCard(book)
			}),
		),
	)
}

func bookCountText(count int) string {
	if count == 1 {
		return "1 book"
	}
	return fmt.Sprintf("%d books", count)
}
