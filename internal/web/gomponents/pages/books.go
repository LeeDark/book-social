package pages

import (
	"fmt"

	bookcomponents "github.com/LeeDark/book-social/internal/web/gomponents/components"
	g "maragu.dev/gomponents"
	gc "maragu.dev/gomponents/components"
	. "maragu.dev/gomponents/html"
)

func BooksPage(data BooksPageData) g.Node {
	return gc.HTML5(gc.HTML5Props{
		Title:    "Books gomponents Spike",
		Language: "en",
		Head: g.Group{
			Link(Rel("stylesheet"), Href("/static/css/pico.min.css")),
			Link(Rel("stylesheet"), Href("/static/css/app.css")),
			Link(Rel("stylesheet"), Href("/static/css/layout.css")),
			Link(Rel("stylesheet"), Href("/static/css/components/book-card.css")),
			Link(Rel("stylesheet"), Href("/static/css/components/cover.css")),
			Link(Rel("stylesheet"), Href("/static/css/components/empty-state.css")),
			Link(Rel("stylesheet"), Href("/static/css/pages/catalog.css")),
		},
		Body: g.Group{
			Main(
				Class("container"),
				Section(
					Class("page-header catalog-header"),
					P(Class("eyebrow"), g.Text("gomponents spike")),
					H1(g.Text("Books rendered with gomponents cards")),
					P(g.Text("This experimental route reuses the catalog service and renders only the book cards with gomponents.")),
				),
				booksContent(data.Books),
			),
		},
	})
}

func booksContent(books []bookcomponents.BookCardView) g.Node {
	if len(books) == 0 {
		return Section(
			Class("empty-state"),
			Aria("labelledby", "gomponents-catalog-empty-heading"),
			Div(Class("empty-state__icon"), Aria("hidden", "true"), g.Text("BS")),
			P(Class("eyebrow"), g.Text("No results")),
			H2(ID("gomponents-catalog-empty-heading"), g.Text("No books found")),
			P(g.Text("The catalog does not have books for this view yet.")),
		)
	}

	return Section(
		Class("catalog-results"),
		Aria("labelledby", "gomponents-catalog-results-heading"),
		Div(
			Class("catalog-results__bar"),
			Div(
				P(Class("eyebrow"), g.Text("Results")),
				H2(ID("gomponents-catalog-results-heading"), g.Text("Catalog books")),
			),
			P(g.Text(bookCountText(len(books)))),
		),
		Div(
			Class("book-grid catalog-grid"),
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
