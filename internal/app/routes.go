package app

import (
	"net/http"

	"github.com/LeeDark/book-social/internal/http/response"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

func (app *App) RegisterRoutes(r chi.Router, deps Deps) {
	dynamic := r.With(chimiddleware.Timeout(applicationTimeout))

	r.Handle("/static/*", http.StripPrefix(
		"/static/",
		http.FileServer(http.Dir("./internal/web/static"))))

	r.Get("/healthz", healthz)
	dynamic.Get("/", app.HomeHandler.Index)
	dynamic.Get("/about", app.HomeHandler.About)

	dynamic.Get("/books", app.CatalogHandler.Catalog)
	dynamic.Get("/books/{slug}", app.CatalogHandler.BookDetails)
	dynamic.Get("/authors/{slug}", app.CatalogHandler.Author)

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		response.RenderNotFound(w, r, deps.Logger, deps.Renderer)
	})
}

func healthz(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}
