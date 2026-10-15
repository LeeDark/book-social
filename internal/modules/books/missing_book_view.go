package books

import "github.com/LeeDark/book-social/internal/http/view"

type MissingBookPageData struct {
	view.Page
	Form MissingBookForm
}

type MissingBookForm struct {
	Title  string
	Author string
	Errors map[string]string
}
