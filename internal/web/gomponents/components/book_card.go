package components

import (
	g "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func BookCard(card BookCardView) g.Node {
	return Article(
		Class("book-card"),
		A(
			Class("book-card__cover "+card.CoverClass),
			Href(card.BookURL),
			Aria("label", "Open "+card.Title),
			Span(
				Class("book-card__cover-book"),
				Aria("hidden", "true"),
				Span(Class("book-cover__brand"), g.Text("Book Social")),
				Span(Class("book-cover__title"), g.Text(card.Title)),
				Span(Class("book-cover__rule")),
			),
		),
		Div(
			Class("book-card__body"),
			H2(
				Class("book-card__title"),
				A(Href(card.BookURL), g.Text(card.Title)),
			),
			P(
				Class("book-card__meta"),
				g.Text("by "),
				A(Href(card.AuthorURL), g.Text(card.AuthorName)),
			),
			g.If(card.GenreName != "", P(
				Class("book-card__genre"),
				A(Href(card.GenreURL), g.Text(card.GenreName)),
			)),
			g.If(card.Description != "", P(
				Class("book-card__description"),
				g.Text(card.Description),
			)),
			g.If(card.ShowDetailsLink, P(
				Class("book-card__footer"),
				A(
					Href(card.BookURL),
					Aria("label", "View details for "+card.Title),
					g.Text("View details"),
				),
			)),
		),
	)
}
