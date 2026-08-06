package books

import (
	"fmt"
	"strings"

	"github.com/LeeDark/book-social/internal/http/view"
)

type CatalogPageData struct {
	view.Page
	Books []BookCardView
}

type BookCardView struct {
	Title           string
	Slug            string
	Description     string
	Authors         []AuthorLinkView
	Genres          []GenreLinkView
	BookURL         string
	CoverClass      string
	ShowDetailsLink bool
	UseHTMXFilters  bool
}

type BookDetailsPageData struct {
	view.Page
	Book BookDetailsView
}

type AuthorPageData struct {
	view.Page
	Author AuthorView
	Books  []BookCardView
}

type AuthorView struct {
	Name        string
	Slug        string
	Description string
}

type BookDetailsView struct {
	ID          int
	Title       string
	Slug        string
	Description string
	CoverClass  string

	Authors    []AuthorLinkView
	Genres     []GenreLinkView
	Covers     []CoverView
	FrontCover *CoverView
}

type AuthorLinkView struct {
	Name      string
	URL       string
	FilterURL string
}

type GenreLinkView struct {
	Name string
	URL  string
}

type CoverView struct {
	Variant        string
	URL            string
	MIMEType       *string
	ByteSize       *int64
	Width          *int
	Height         *int
	ChecksumSHA256 *string
}

func coverClassForBook(id int) string {
	return fmt.Sprintf("cover-%d", id%5)
}

func mapBooksToCards(books []Book) []BookCardView {
	cards := make([]BookCardView, 0, len(books))
	for _, book := range books {
		card := BookCardView{
			Title:           book.Title,
			Slug:            book.Slug,
			Description:     book.Description,
			Authors:         mapAuthorsToLinks(book.Authors),
			Genres:          mapGenresToLinks(book.Genres),
			BookURL:         fmt.Sprintf("/books/%s", book.Slug),
			CoverClass:      coverClassForBook(book.ID),
			ShowDetailsLink: true,
		}

		cards = append(cards, card)
	}

	return cards
}

func enableHTMXFilters(cards []BookCardView) []BookCardView {
	for i := range cards {
		cards[i].UseHTMXFilters = true
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
		Authors:     mapAuthorsToLinks(book.Authors),
		Genres:      mapGenresToLinks(book.Genres),
		Covers:      mapCoversToViews(book.Covers),
	}
	details.FrontCover = frontCover(details.Covers)

	return details
}

func mapAuthorToView(author Author) AuthorView {
	return AuthorView{
		Name:        authorFullName(author),
		Slug:        author.Slug,
		Description: author.Description,
	}
}

func mapAuthorsToLinks(authors []Author) []AuthorLinkView {
	links := make([]AuthorLinkView, 0, len(authors))
	for _, author := range authors {
		links = append(links, AuthorLinkView{
			Name:      authorFullName(author),
			URL:       fmt.Sprintf("/authors/%s", author.Slug),
			FilterURL: fmt.Sprintf("/books?author=%s", author.Slug),
		})
	}
	return links
}

func mapGenresToLinks(genres []Genre) []GenreLinkView {
	links := make([]GenreLinkView, 0, len(genres))
	for _, genre := range genres {
		links = append(links, GenreLinkView{
			Name: genre.Name,
			URL:  fmt.Sprintf("/books?genre=%s", genre.Slug),
		})
	}
	return links
}

func mapCoversToViews(covers []Cover) []CoverView {
	views := make([]CoverView, 0, len(covers))
	for _, cover := range covers {
		views = append(views, CoverView{
			Variant:        cover.Variant,
			URL:            cover.URL,
			MIMEType:       cover.MIMEType,
			ByteSize:       cover.ByteSize,
			Width:          cover.Width,
			Height:         cover.Height,
			ChecksumSHA256: cover.ChecksumSHA256,
		})
	}
	return views
}

func frontCover(covers []CoverView) *CoverView {
	for index := range covers {
		if covers[index].Variant == "front" {
			return &covers[index]
		}
	}

	return nil
}

func authorFullName(author Author) string {
	parts := make([]string, 0, 3)
	for _, part := range []string{author.FirstName, author.SecondName, author.SurName} {
		if part != "" {
			parts = append(parts, part)
		}
	}
	return strings.Join(parts, " ")
}
