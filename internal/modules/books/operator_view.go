package books

import (
	"fmt"
	"slices"
	"strconv"

	"github.com/LeeDark/book-social/internal/http/view"
)

type OperatorFindPageData struct {
	view.Page
	Query     string
	Books     []OperatorBookLink
	FormError string
}

type OperatorBookLink struct {
	Title   string
	Slug    string
	EditURL string
	BookURL string
}

type OperatorFormPageData struct {
	view.Page
	Form      OperatorForm
	ActionURL string
	BookURL   string
	IsNew     bool
	Authors   []OperatorChoice
	Genres    []OperatorChoice
}

// OperatorForm retains submitted text and selections when validation fails.
// It contains no authorization capability or audit input.
type OperatorForm struct {
	Title            string
	Slug             string
	Description      string
	FrontCoverURL    string
	FrontCoverAction CoverAction
	CurrentCoverURL  string
	AuthorIDs        []string
	GenreIDs         []string
	Errors           map[string]string
}

type OperatorChoice struct {
	Value    string
	Label    string
	Selected bool
}

func operatorPage(title string) view.Page {
	return view.Page{
		Title: title,
		Nav:   view.MainNavigation(),
		Breadcrumbs: []view.Breadcrumb{
			{Label: "Home", Href: "/"},
			{Label: "Catalog maintenance", Href: "/admin/catalog"},
		},
	}
}

func operatorEditURL(bookID int) string {
	return fmt.Sprintf("/admin/catalog/books/%d/edit", bookID)
}

func operatorFormFromBook(book OperatorBook) OperatorForm {
	return OperatorForm{
		Title:            book.Title,
		Slug:             book.Slug,
		Description:      book.Description,
		FrontCoverAction: CoverKeep,
		CurrentCoverURL:  book.FrontCoverURL,
		AuthorIDs:        operatorIDValues(book.AuthorIDs),
		GenreIDs:         operatorIDValues(book.GenreIDs),
		Errors:           make(map[string]string),
	}
}

func operatorIDValues(ids []int) []string {
	values := make([]string, 0, len(ids))
	for _, id := range ids {
		values = append(values, strconv.Itoa(id))
	}

	return values
}

func operatorChoices(options CatalogOptions, form OperatorForm) ([]OperatorChoice, []OperatorChoice) {
	authors := make([]OperatorChoice, 0, len(options.Authors))
	for _, author := range options.Authors {
		value := strconv.Itoa(author.ID)
		authors = append(authors, OperatorChoice{
			Value:    value,
			Label:    authorFullName(author),
			Selected: slices.Contains(form.AuthorIDs, value),
		})
	}

	genres := make([]OperatorChoice, 0, len(options.Genres))
	for _, genre := range options.Genres {
		value := strconv.Itoa(genre.ID)
		genres = append(genres, OperatorChoice{
			Value:    value,
			Label:    genre.Name,
			Selected: slices.Contains(form.GenreIDs, value),
		})
	}

	return authors, genres
}
