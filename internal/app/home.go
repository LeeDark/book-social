package app

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/LeeDark/book-social/internal/http/render"
	"github.com/LeeDark/book-social/internal/http/response"
	"github.com/LeeDark/book-social/internal/http/view"
)

type HomeHandler struct {
	renderer *render.Renderer
	logger   *slog.Logger
}

type HomePageData struct {
	view.Page
}

func NewHomeHandler(renderer *render.Renderer, logger *slog.Logger) *HomeHandler {
	return &HomeHandler{
		renderer: renderer,
		logger:   logger,
	}
}

func (h *HomeHandler) Index(w http.ResponseWriter, r *http.Request) {
	data := HomePageData{
		Page: view.Page{
			Title: "Book Social",
			Breadcrumbs: []view.Breadcrumb{
				{Label: "Home"},
			},
		},
	}

	if err := h.renderer.Render(w, http.StatusOK, "home.tmpl", data); err != nil {
		response.ServerError(w, r, h.logger, fmt.Errorf("render home page: %w", err))
		return
	}
}

func (h *HomeHandler) About(w http.ResponseWriter, r *http.Request) {
	data := HomePageData{
		Page: view.Page{
			Title: "About Book Social",
		},
	}

	if err := h.renderer.Render(w, http.StatusOK, "about.tmpl", data); err != nil {
		response.ServerError(w, r, h.logger, fmt.Errorf("render about page: %w", err))
		return
	}
}
