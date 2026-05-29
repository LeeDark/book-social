package books

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/LeeDark/book-social/internal/http/render"
	"github.com/LeeDark/book-social/internal/http/response"
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
	}

	h.logger.Debug("Catalog page", slog.Any("data", data))

	if err := h.renderer.Render(w, http.StatusOK, "catalog.tmpl", data); err != nil {
		response.ServerError(w, r, h.logger, fmt.Errorf("render catalog page: %w", err))
		return
	}
}
