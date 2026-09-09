package view

import (
	"net/http"

	"github.com/LeeDark/book-social/internal/http/auth"
	"github.com/LeeDark/book-social/internal/http/flash"
)

type Page struct {
	Title       string
	Description string
	ActiveNav   string
	Nav         []NavItem
	Breadcrumbs []Breadcrumb
	CurrentUser *CurrentUser
	Flash       []FlashMessage
}

type Breadcrumb struct {
	Label string
	Href  string
}

type CurrentUser struct {
	ID          int
	DisplayName string
}

type FlashMessage struct {
	Type string
	Text string
}

// ApplyRequestState copies only render-safe identity and fixed flash data into
// a page model. It keeps password, role and session details out of templates.
func ApplyRequestState(page *Page, r *http.Request) {
	if page == nil {
		return
	}
	if identity, ok := auth.CurrentUserFromRequest(r); ok {
		page.CurrentUser = &CurrentUser{ID: identity.ID, DisplayName: identity.FirstName}
	}
	switch flash.FromRequest(r) {
	case flash.Registered:
		page.Flash = []FlashMessage{{Type: "success", Text: "Registration complete."}}
	case flash.SignedIn:
		page.Flash = []FlashMessage{{Type: "success", Text: "You are signed in."}}
	case flash.SignedOut:
		page.Flash = []FlashMessage{{Type: "success", Text: "You are signed out."}}
	}
}
