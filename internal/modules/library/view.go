package library

import (
	"fmt"
	"strings"

	"github.com/LeeDark/book-social/internal/http/view"
	"github.com/LeeDark/book-social/internal/modules/books"
)

type LibraryPageData struct {
	view.Page
	Items     []ItemView
	FormError string
}

type ItemView struct {
	ID            int
	Version       int
	RemoveURL     string
	Title         string
	BookURL       string
	Authors       []AuthorLinkView
	StateLabel    string
	StatusOptions []StatusOption
}

type StatusOption struct {
	Value, Label string
	Selected     bool
}

type RemovalPageData struct {
	view.Page
	Item ItemView
}

type AuthorLinkView struct {
	Name string
	URL  string
}

func mapItemsToViews(items []Item) []ItemView {
	views := make([]ItemView, 0, len(items))
	for _, item := range items {
		views = append(views, ItemView{
			ID:            item.ID,
			Version:       item.Version,
			RemoveURL:     fmt.Sprintf("/me/library/%d/remove", item.ID),
			Title:         item.Book.Title,
			BookURL:       fmt.Sprintf("/books/%s", item.Book.Slug),
			Authors:       mapAuthorsToLinks(item.Book.Authors),
			StateLabel:    item.Status.Label(),
			StatusOptions: statusOptions(item.Status),
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

func statusOptions(current ReadingStatus) []StatusOption {
	values := []ReadingStatus{ReadingStatusWantToRead, ReadingStatusReading, ReadingStatusRead}
	options := make([]StatusOption, 0, len(values))
	for _, value := range values {
		options = append(options, StatusOption{
			Value:    string(value),
			Label:    value.Label(),
			Selected: current == value,
		})
	}
	return options
}
