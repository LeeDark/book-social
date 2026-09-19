package library

import (
	"fmt"
	"strings"

	"github.com/LeeDark/book-social/internal/http/view"
	"github.com/LeeDark/book-social/internal/modules/books"
)

type LibraryPageData struct {
	view.Page
	Items []ItemView
}

type ItemView struct {
	Title      string
	BookURL    string
	Authors    []AuthorLinkView
	StateLabel string
}

type AuthorLinkView struct {
	Name string
	URL  string
}

func mapItemsToViews(items []Item) []ItemView {
	views := make([]ItemView, 0, len(items))
	for _, item := range items {
		views = append(views, ItemView{
			Title:      item.Book.Title,
			BookURL:    fmt.Sprintf("/books/%s", item.Book.Slug),
			Authors:    mapAuthorsToLinks(item.Book.Authors),
			StateLabel: "Want to read",
		})
	}

	return views
}

func mapAuthorsToLinks(authors []books.Author) []AuthorLinkView {
	links := make([]AuthorLinkView, 0, len(authors))
	for _, author := range authors {
		links = append(links, AuthorLinkView{
			Name: authorName(author),
			URL:  fmt.Sprintf("/authors/%s", author.Slug),
		})
	}

	return links
}

func authorName(author books.Author) string {
	parts := make([]string, 0, 3)
	for _, part := range []string{author.FirstName, author.SecondName, author.SurName} {
		if part != "" {
			parts = append(parts, part)
		}
	}

	return strings.Join(parts, " ")
}
