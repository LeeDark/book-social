package library

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	httpauth "github.com/LeeDark/book-social/internal/http/auth"
	"github.com/LeeDark/book-social/internal/http/flash"
	"github.com/LeeDark/book-social/internal/http/render"
	"github.com/LeeDark/book-social/internal/http/response"
	"github.com/LeeDark/book-social/internal/http/view"
)

const maxLibraryFormBytes = 16 << 10

type libraryService interface {
	Add(ctx context.Context, userID int, bookSlug string) error
	List(ctx context.Context, userID int) ([]Item, error)
}

type Handler struct {
	service  libraryService
	flashes  *flash.Manager
	renderer *render.Renderer
	logger   *slog.Logger
}

func NewHandler(service libraryService, flashes *flash.Manager, renderer *render.Renderer, logger *slog.Logger) *Handler {
	return &Handler{service: service, flashes: flashes, renderer: renderer, logger: logger}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	identity, ok := httpauth.CurrentUserFromRequest(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	items, err := h.service.List(r.Context(), identity.ID)
	if err != nil {
		response.ServerError(w, r, h.logger, fmt.Errorf("list private library: %w", err))
		return
	}

	data := LibraryPageData{
		Page:  view.Page{Title: "Your library", ActiveNav: "library", Nav: view.MainNavigation()},
		Items: mapItemsToViews(items),
	}
	view.ApplyRequestState(&data.Page, r)

	if err := h.renderer.Render(w, http.StatusOK, "library.tmpl", data); err != nil {
		response.ServerError(w, r, h.logger, fmt.Errorf("render private library: %w", err))
	}
}

func (h *Handler) Add(w http.ResponseWriter, r *http.Request) {
	identity, ok := httpauth.CurrentUserFromRequest(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxLibraryFormBytes)
	if err := r.ParseForm(); err != nil {
		response.BadRequest(w)
		return
	}

	err := h.service.Add(r.Context(), identity.ID, r.PostForm.Get("book_slug"))
	switch {
	case err == nil:
		h.flashes.Set(w, flash.LibraryItemAdded)
		http.Redirect(w, r, "/me/library", http.StatusSeeOther)
	case errors.Is(err, ErrInvalidInput):
		http.Error(w, "Book selection is required.", http.StatusUnprocessableEntity)
	case errors.Is(err, ErrBookNotFound):
		http.Error(w, "Book not found.", http.StatusNotFound)
	case errors.Is(err, ErrItemAlreadyExists):
		http.Error(w, "This book is already in your library.", http.StatusConflict)
	default:
		response.ServerError(w, r, h.logger, fmt.Errorf("add private library item: %w", err))
	}
}
