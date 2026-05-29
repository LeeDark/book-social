package app

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/LeeDark/book-social/internal/http/render"
	"github.com/LeeDark/book-social/internal/http/response"
)

type HomeHandler struct {
	renderer *render.Renderer
	logger   *slog.Logger
}

type HomePageData struct {
	Title string
}

func NewHomeHandler(renderer *render.Renderer, logger *slog.Logger) *HomeHandler {
	return &HomeHandler{
		renderer: renderer,
		logger:   logger,
	}
}

func (h *HomeHandler) Index(w http.ResponseWriter, r *http.Request) {
	data := HomePageData{
		Title: "Book Social",
	}

	if err := h.renderer.Render(w, http.StatusOK, "home.tmpl", data); err != nil {
		response.ServerError(w, r, h.logger, fmt.Errorf("render home page: %w", err))
		return
	}
}

func (h *HomeHandler) About(w http.ResponseWriter, r *http.Request) {
	data := HomePageData{
		Title: "About Book Social",
	}

	if err := h.renderer.Render(w, http.StatusOK, "about.tmpl", data); err != nil {
		response.ServerError(w, r, h.logger, fmt.Errorf("render about page: %w", err))
		return
	}
}
