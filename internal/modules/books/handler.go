package books

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/LeeDark/book-social/internal/http/render"
	"github.com/LeeDark/book-social/internal/http/response"
	"github.com/go-chi/chi/v5"
)

type CatalogHandler struct {
	service  CatalogPageProvider
	renderer *render.Renderer
	logger   *slog.Logger
}

func NewCatalogHandler(service CatalogPageProvider, renderer *render.Renderer, logger *slog.Logger) *CatalogHandler {
	return &CatalogHandler{
		service:  service,
		renderer: renderer,
		logger:   logger,
	}
}

func (h *CatalogHandler) Catalog(w http.ResponseWriter, r *http.Request) {
	data, err := h.service.CatalogPage(r.Context())
	if err != nil {
		response.ServerError(w, r, h.logger, fmt.Errorf("get catalog page: %w", err))
		return
	}

	//h.logger.Debug("Catalog page", slog.Any("data", data))

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
			response.NotFound(w)
			return
		}

		response.ServerError(w, r, h.logger, fmt.Errorf("get book details page: %w", err))
		return
	}

	//h.logger.Debug("Book Details page", slog.Any("data", data))

	if err := h.renderer.Render(w, http.StatusOK, "book_details.tmpl", data); err != nil {
		response.ServerError(w, r, h.logger, fmt.Errorf("render book details page: %w", err))
		return
	}
}
