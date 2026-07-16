package app

import (
	"net/http"

	"github.com/LeeDark/book-social/internal/http/response"
	"github.com/go-chi/chi/v5"
)

type templCatalogHandler interface {
	CatalogTempl(w http.ResponseWriter, r *http.Request)
}

type gomponentsCatalogHandler interface {
	CatalogGomponents(w http.ResponseWriter, r *http.Request)
}

func (app *App) RegisterRoutes(r chi.Router, deps Deps) {
	r.Handle("/static/*", http.StripPrefix(
		"/static/",
		http.FileServer(http.Dir("./internal/web/static"))))

	r.Get("/healthz", healthz)
	r.Get("/", app.HomeHandler.Index)
	r.Get("/about", app.HomeHandler.About)

	r.Get("/books", app.CatalogHandler.Catalog)
	if handler, ok := app.CatalogHandler.(templCatalogHandler); ok {
		r.Get("/books-templ", handler.CatalogTempl)
	}
	if handler, ok := app.CatalogHandler.(gomponentsCatalogHandler); ok {
		r.Get("/books-gomponents", handler.CatalogGomponents)
	}
	r.Get("/books/{slug}", app.CatalogHandler.BookDetails)
	r.Get("/authors/{slug}", app.CatalogHandler.Author)

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		response.RenderNotFound(w, r, deps.Logger, deps.Renderer)
	})
}

func healthz(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}
