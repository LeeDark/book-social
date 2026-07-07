package components

import (
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

func BookCard(card BookCardView) g.Node {
	return h.Article(
		h.Class("book-card"),
		h.A(
			h.Class("book-card__cover "+card.CoverClass),
			h.Href(card.BookURL),
			h.Aria("label", "Open "+card.Title),
			h.Span(
				h.Class("book-card__cover-book"),
				h.Aria("hidden", "true"),
				h.Span(h.Class("book-cover__brand"), g.Text("Book Social")),
				h.Span(h.Class("book-cover__title"), g.Text(card.Title)),
				h.Span(h.Class("book-cover__rule")),
			),
		),
		h.Div(
			h.Class("book-card__body"),
			h.H2(
				h.Class("book-card__title"),
				h.A(h.Href(card.BookURL), g.Text(card.Title)),
			),
			h.P(
				h.Class("book-card__meta"),
				g.Text("by "),
				h.A(h.Href(card.AuthorURL), g.Text(card.AuthorName)),
			),
			g.If(card.GenreName != "", h.P(
				h.Class("book-card__genre"),
				h.A(h.Href(card.GenreURL), g.Text(card.GenreName)),
			)),
			g.If(card.Description != "", h.P(
				h.Class("book-card__description"),
				g.Text(card.Description),
			)),
			g.If(card.ShowDetailsLink, h.P(
				h.Class("book-card__footer"),
				h.A(
					h.Href(card.BookURL),
					h.Aria("label", "View details for "+card.Title),
					g.Text("View details"),
				),
			)),
		),
	)
}
