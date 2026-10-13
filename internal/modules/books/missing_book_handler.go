package books

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/LeeDark/book-social/internal/http/auth"
	"github.com/LeeDark/book-social/internal/http/response"
	"github.com/LeeDark/book-social/internal/http/view"
	"github.com/LeeDark/book-social/internal/modules/users"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

const maxMissingBookFormBytes = 64 << 10

type missingBookRequester interface {
	Request(ctx context.Context, actorID int, requestID string, input MissingBookRequest) error
}

type missingBookRenderer interface {
	Render(w http.ResponseWriter, status int, page string, data any) error
}

type MissingBookHandler struct {
	service  missingBookRequester
	renderer missingBookRenderer
	logger   *slog.Logger
}

func NewMissingBookHandler(service missingBookRequester, renderer missingBookRenderer, logger *slog.Logger) *MissingBookHandler {
	return &MissingBookHandler{service: service, renderer: renderer, logger: logger}
}

func (h *MissingBookHandler) Request(w http.ResponseWriter, r *http.Request) {
	identity, ok := auth.CurrentUserFromRequest(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.renderForm(w, r, http.StatusOK, MissingBookForm{})
		return
	case http.MethodPost:
	default:
		w.Header().Set("Allow", "GET, POST")
		response.ClientError(w, http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxMissingBookFormBytes)
	if err := r.ParseForm(); err != nil {
		h.renderForm(w, r, http.StatusUnprocessableEntity, MissingBookForm{
			Errors: map[string]string{"form": "Enter a valid request with a title and optional author."},
		})
		return
	}
	form := MissingBookForm{Title: r.PostForm.Get("title"), Author: r.PostForm.Get("author")}
	err := h.service.Request(r.Context(), identity.ID, chimiddleware.GetReqID(r.Context()), MissingBookRequest{
		Title:  form.Title,
		Author: form.Author,
	})
	if err != nil {
		if validation, ok := errors.AsType[CatalogValidationError](err); ok {
			form.Errors = map[string]string{validation.Field: validation.Message}
			h.renderForm(w, r, http.StatusUnprocessableEntity, form)
			return
		}
		if errors.Is(err, users.ErrUnauthenticated) {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		response.ServerError(w, r, h.logger, fmt.Errorf("request missing book: %w", err))
		return
	}

	http.Redirect(w, r, "/books", http.StatusSeeOther)
}

func (h *MissingBookHandler) renderForm(w http.ResponseWriter, r *http.Request, status int, form MissingBookForm) {
	data := MissingBookPageData{
		Page: view.Page{Title: "Request a missing book", ActiveNav: "catalog", Nav: view.MainNavigation()},
		Form: form,
	}
	view.ApplyRequestState(&data.Page, r)
	if err := h.renderer.Render(w, status, "missing_book.tmpl", data); err != nil {
		response.ServerError(w, r, h.logger, fmt.Errorf("render missing-book form: %w", err))
	}
}
