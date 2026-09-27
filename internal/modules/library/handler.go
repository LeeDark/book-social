package library

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	httpauth "github.com/LeeDark/book-social/internal/http/auth"
	"github.com/LeeDark/book-social/internal/http/flash"
	"github.com/LeeDark/book-social/internal/http/render"
	"github.com/LeeDark/book-social/internal/http/response"
	"github.com/LeeDark/book-social/internal/http/view"
	"github.com/go-chi/chi/v5"
)

const maxLibraryFormBytes = 16 << 10

type libraryService interface {
	Add(ctx context.Context, userID int, bookSlug string) error
	List(ctx context.Context, userID int) ([]Item, error)
	Get(ctx context.Context, userID, itemID int) (Item, error)
	UpdateStatus(
		ctx context.Context,
		userID, itemID int,
		status ReadingStatus,
		expectedVersion int,
	) error
	Remove(ctx context.Context, userID, itemID, expectedVersion int) error
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

func (h *Handler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	identity, ok := httpauth.CurrentUserFromRequest(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	itemID, version, err := lifecycleFormValues(w, r)
	if err != nil {
		http.Error(w, "A valid status and version are required.", http.StatusUnprocessableEntity)
		return
	}
	status := ReadingStatus(r.PostForm.Get("status"))
	err = h.service.UpdateStatus(r.Context(), identity.ID, itemID, status, version)
	if err == nil {
		h.flashes.Set(w, flash.LibraryStatusChanged)
		http.Redirect(w, r, "/me/library", http.StatusSeeOther)
		return
	}
	h.lifecycleError(w, r, err)
}

func (h *Handler) RemoveConfirmation(w http.ResponseWriter, r *http.Request) {
	identity, ok := httpauth.CurrentUserFromRequest(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	itemID, err := itemIDFromRequest(r)
	if err != nil {
		http.Error(w, "A valid library item is required.", http.StatusUnprocessableEntity)
		return
	}
	item, err := h.service.Get(r.Context(), identity.ID, itemID)
	if err != nil {
		h.lifecycleError(w, r, err)
		return
	}
	data := RemovalPageData{
		Page: view.Page{
			Title:     "Remove from library",
			ActiveNav: "library",
			Nav:       view.MainNavigation(),
		},
		Item: mapItemsToViews([]Item{item})[0],
	}
	view.ApplyRequestState(&data.Page, r)
	if err := h.renderer.Render(w, http.StatusOK, "library_remove.tmpl", data); err != nil {
		response.ServerError(w, r, h.logger, fmt.Errorf("render library removal: %w", err))
	}
}

func (h *Handler) Remove(w http.ResponseWriter, r *http.Request) {
	identity, ok := httpauth.CurrentUserFromRequest(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	itemID, version, err := lifecycleFormValues(w, r)
	if err != nil {
		http.Error(w, "A valid version is required.", http.StatusUnprocessableEntity)
		return
	}
	if err := h.service.Remove(r.Context(), identity.ID, itemID, version); err != nil {
		h.lifecycleError(w, r, err)
		return
	}
	h.flashes.Set(w, flash.LibraryItemRemoved)
	http.Redirect(w, r, "/me/library", http.StatusSeeOther)
}

func lifecycleFormValues(w http.ResponseWriter, r *http.Request) (int, int, error) {
	itemID, err := itemIDFromRequest(r)
	if err != nil {
		return 0, 0, err
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxLibraryFormBytes)
	if err := r.ParseForm(); err != nil {
		return 0, 0, err
	}
	version, err := strconv.Atoi(r.PostForm.Get("version"))
	if err != nil || version <= 0 {
		return 0, 0, errors.New("invalid version")
	}
	return itemID, version, nil
}

func itemIDFromRequest(r *http.Request) (int, error) {
	itemID, err := strconv.Atoi(chi.URLParam(r, "itemID"))
	if err != nil || itemID <= 0 {
		return 0, errors.New("invalid item ID")
	}
	return itemID, nil
}

func (h *Handler) lifecycleError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, ErrInvalidInput):
		http.Error(w, "Invalid library request.", http.StatusUnprocessableEntity)
	case errors.Is(err, ErrItemNotFound):
		http.Error(w, "Library item not found.", http.StatusNotFound)
	case errors.Is(err, ErrItemVersionConflict):
		http.Error(w, "This item changed. Reload your library and try again.", http.StatusConflict)
	default:
		response.ServerError(w, r, h.logger, fmt.Errorf("library lifecycle: %w", err))
	}
}
