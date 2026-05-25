package app

import (
	"log/slog"
	"net/http"

	"github.com/LeeDark/book-social/internal/http/render"
)

type HomePageData struct {
	Title string
}

type HomeHandler struct {
	renderer *render.Renderer
	logger   *slog.Logger
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
		h.logger.Error("render home page", "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (h *HomeHandler) About(w http.ResponseWriter, r *http.Request) {
	data := HomePageData{
		Title: "About Book Social",
	}

	if err := h.renderer.Render(w, http.StatusOK, "about.tmpl", data); err != nil {
		h.logger.Error("render home page", "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}
