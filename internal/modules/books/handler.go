package books

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/LeeDark/book-social/internal/http/auth"
	"github.com/LeeDark/book-social/internal/http/render"
	"github.com/LeeDark/book-social/internal/http/response"
	"github.com/LeeDark/book-social/internal/http/view"
	"github.com/go-chi/chi/v5"
)

type LibraryStateProvider interface {
	StatesForBooks(ctx context.Context, userID int, bookIDs []int) (map[int]string, error)
}

type CatalogHandler struct {
	service  CatalogPageProvider
	renderer *render.Renderer
	logger   *slog.Logger
	states   LibraryStateProvider
}

func NewCatalogHandler(
	service CatalogPageProvider,
	renderer *render.Renderer,
	logger *slog.Logger,
	states LibraryStateProvider,
) *CatalogHandler {
	return &CatalogHandler{
		service:  service,
		renderer: renderer,
		logger:   logger,
		states:   states,
	}
}

func (h *CatalogHandler) Catalog(w http.ResponseWriter, r *http.Request) {
	filter := BookFilter{
		AuthorSlug: r.URL.Query().Get("author"),
		GenreSlug:  r.URL.Query().Get("genre"),
	}

	data, err := h.service.CatalogPage(r.Context(), filter)
	if err != nil {
		response.ServerError(w, r, h.logger, fmt.Errorf("get catalog page: %w", err))
		return
	}
	view.ApplyRequestState(&data.Page, r)
	if err := h.applyCatalogLibraryStates(r, &data); err != nil {
		response.ServerError(w, r, h.logger, fmt.Errorf("get current user's catalog library states: %w", err))
		return
	}

	for index := range data.Books {
		data.Books[index].CanAddToLibrary = data.CurrentUser != nil && data.Books[index].LibraryState == nil
	}

	//h.logger.Debug("Catalog page", slog.Any("data", data))

	if r.Header.Get("HX-Request") == "true" {
		if err := h.renderer.RenderPartial(w, http.StatusOK, "catalog.tmpl", "book_list", data); err != nil {
			response.ServerError(w, r, h.logger, fmt.Errorf("render catalog book list partial: %w", err))
			return
		}
		return
	}

	if err := h.renderer.Render(w, http.StatusOK, "catalog.tmpl", data); err != nil {
		response.ServerError(w, r, h.logger, fmt.Errorf("render catalog page: %w", err))
		return
	}
}
func (h *CatalogHandler) BookDetails(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	data, err := h.service.BookDetailsPage(r.Context(), slug)
	if err != nil {
		if errors.Is(err, ErrBookNotFound) {
			response.RenderNotFound(w, r, h.logger, h.renderer)
			return
		}

		response.ServerError(w, r, h.logger, fmt.Errorf("get book details page: %w", err))
		return
	}
	view.ApplyRequestState(&data.Page, r)
	if err := h.applyDetailLibraryState(r, &data); err != nil {
		response.ServerError(w, r, h.logger, fmt.Errorf("get current user's book library state: %w", err))
		return
	}

	data.Book.CanAddToLibrary = data.CurrentUser != nil && data.Book.LibraryState == nil

	//h.logger.Debug("Book Details page", slog.Any("data", data))

	if err := h.renderer.Render(w, http.StatusOK, "book_details.tmpl", data); err != nil {
		response.ServerError(w, r, h.logger, fmt.Errorf("render book details page: %w", err))
		return
	}
}

func (h *CatalogHandler) Author(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	data, err := h.service.AuthorPage(r.Context(), slug)
	if err != nil {
		if errors.Is(err, ErrAuthorNotFound) {
			response.RenderNotFound(w, r, h.logger, h.renderer)
			return
		}

		response.ServerError(w, r, h.logger, fmt.Errorf("get author page: %w", err))
		return
	}
	view.ApplyRequestState(&data.Page, r)

	if err := h.renderer.Render(w, http.StatusOK, "author.tmpl", data); err != nil {
		response.ServerError(w, r, h.logger, fmt.Errorf("render author page: %w", err))
		return
	}
}

func (h *CatalogHandler) applyCatalogLibraryStates(r *http.Request, data *CatalogPageData) error {
	identity, ok := auth.CurrentUserFromRequest(r)
	if !ok || h.states == nil {
		return nil
	}

	bookIDs := make([]int, 0, len(data.Books))
	for _, book := range data.Books {
		bookIDs = append(bookIDs, book.ID)
	}

	states, err := h.states.StatesForBooks(r.Context(), identity.ID, bookIDs)
	if err != nil {
		return err
	}

	for index := range data.Books {
		if label, exists := states[data.Books[index].ID]; exists {
			data.Books[index].LibraryState = &LibraryStateView{Label: label}
		}
	}
	return nil
}

func (h *CatalogHandler) applyDetailLibraryState(r *http.Request, data *BookDetailsPageData) error {
	identity, ok := auth.CurrentUserFromRequest(r)
	if !ok || h.states == nil {
		return nil
	}

	states, err := h.states.StatesForBooks(r.Context(), identity.ID, []int{data.Book.ID})
	if err != nil {
		return err
	}

	if label, exists := states[data.Book.ID]; exists {
		data.Book.LibraryState = &LibraryStateView{Label: label}
	}
	return nil
}
